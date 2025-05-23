// Package weather is a simple REST API
package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/codyonesock/rest_weather/internal/shared"
	"github.com/codyonesock/rest_weather/internal/storage"
)

// GeocodeResponse is a struct based on geocode data returned from open-meteo.
type GeocodeResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

// CurrentWeatherResponse is a struct based on current weather data returned from open-meteo.
type CurrentWeatherResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
	} `json:"current_weather"`
}

// ForecastResponse is a struct based on forecast data returned from open-meteo.
type ForecastResponse struct {
	Daily struct {
		Dates    []string  `json:"time"`
		MaxTemps []float64 `json:"temperature_2m_max"`
		MinTemps []float64 `json:"temperature_2m_min"`
	} `json:"daily"`
}

// Service handles dependencies and config.
type Service struct {
	Logger                *zap.Logger
	Storage               storage.ServiceInterface
	CurrentWeatherAPIURL  string
	ForecastWeatherAPIURL string
	GeocodeAPIURL         string
}

// NewWeatherService create a new instance of Service.
func NewWeatherService(
	l *zap.Logger,
	si storage.ServiceInterface,
	currentWeatherAPIURL,
	forecastWeatherAPIURL,
	geocodeAPIURL string,
) *Service {
	return &Service{
		Logger:                l,
		Storage:               si,
		CurrentWeatherAPIURL:  currentWeatherAPIURL,
		ForecastWeatherAPIURL: forecastWeatherAPIURL,
		GeocodeAPIURL:         geocodeAPIURL,
	}
}

const contextTimeout = 5 * time.Second

// err113 demands no dynamic errors!
var (
	ErrInvalidURL       = errors.New("invalid URL")
	ErrNoResultsForCity = errors.New("no results for city")
	ErrCityRequired     = errors.New("city is required")
	ErrInvalidUnit      = errors.New("invalid unit type")
)

// GetCurrentWeatherByCity returns the current weather (temperature and weather speed).
func (s *Service) GetCurrentWeatherByCity(
	w http.ResponseWriter,
	city string,
) (*CurrentWeatherResponse, error) {
	var weatherData CurrentWeatherResponse
	if err := s.getWeatherData(city, s.CurrentWeatherAPIURL, &weatherData); err != nil {
		s.Logger.Error("Failed to get weather data", zap.Error(err))
		return nil, fmt.Errorf("failed to get weather data for city %s: %w", city, err)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(weatherData); err != nil {
		s.Logger.Error("Error encoding weatherData", zap.Error(err))
		return nil, fmt.Errorf("failed to encode weatherData: %w", err)
	}

	return &weatherData, nil
}

// GetForecastByCity returns a 7 day forecast (dates, min/max temps) using the lat/lon of the city entered.
func (s *Service) GetForecastByCity(
	w http.ResponseWriter,
	city string,
) (*ForecastResponse, error) {
	var forecastData ForecastResponse
	if err := s.getWeatherData(city, s.ForecastWeatherAPIURL, &forecastData); err != nil {
		s.Logger.Error("Failed to get forecast data", zap.Error(err))
		return nil, fmt.Errorf("failed to get forecast data for city %s: %w", city, err)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(forecastData); err != nil {
		s.Logger.Error("Error encoding forecastData", zap.Error(err))
		return nil, fmt.Errorf("failed to encode forecastData: %w", err)
	}

	return &forecastData, nil
}

// GetUserData returns user data that's read from a local json file.
func (s *Service) GetUserData(w http.ResponseWriter) (*shared.UserData, error) {
	userData, err := s.Storage.LoadUserData()
	if err != nil {
		s.Logger.Error("Error loading user data", zap.Error(err))
		return nil, fmt.Errorf("failed to load user data: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(userData); err != nil {
		s.Logger.Error("Error encoding user data", zap.Error(err))
		return nil, fmt.Errorf("failed to encode user data: %w", err)
	}

	return &userData, nil
}

// AddCity will add the passed in cities to your user data.
func (s *Service) AddCity(w http.ResponseWriter, city string) error {
	if city == "" {
		return fmt.Errorf("%w", ErrCityRequired)
	}

	userData, err := s.Storage.LoadUserData()
	if err != nil {
		s.Logger.Error("Error loading user data", zap.Error(err))
		return fmt.Errorf("failed to load user data: %w", err)
	}

	cities := strings.Split(city, ",")
	for _, newCity := range cities {
		newCity = strings.TrimSpace(newCity)
		if newCity == "" {
			continue
		}

		exists := false

		for _, existingCity := range userData.Cities {
			if strings.EqualFold(existingCity, newCity) {
				exists = true
				break
			}
		}

		if !exists {
			userData.Cities = append(userData.Cities, newCity)
		}
	}

	if err := s.Storage.SaveUserData(userData); err != nil {
		s.Logger.Error("Error saving user data", zap.Error(err))
		return fmt.Errorf("failed to save user data: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(userData); err != nil {
		s.Logger.Error("Error encoding response", zap.Error(err))
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}

// DeleteCity will remove the passed in cities from your user data.
func (s *Service) DeleteCity(w http.ResponseWriter, city string) error {
	if city == "" {
		return fmt.Errorf("%w", ErrCityRequired)
	}

	userData, err := s.Storage.LoadUserData()
	if err != nil {
		s.Logger.Error("Error loading user data", zap.Error(err))
		return fmt.Errorf("failed to load user data: %w", err)
	}

	cities := strings.Split(city, ",")
	for _, cityToRemove := range cities {
		cityToRemove = strings.TrimSpace(cityToRemove)
		if cityToRemove == "" {
			continue
		}

		cityFound := false

		for i, existingCity := range userData.Cities {
			if strings.EqualFold(existingCity, cityToRemove) {
				userData.Cities = append(userData.Cities[:i], userData.Cities[i+1:]...)
				cityFound = true

				break
			}
		}

		if !cityFound {
			s.Logger.Warn("City not found", zap.String("city", cityToRemove))
		}
	}

	if err := s.Storage.SaveUserData(userData); err != nil {
		s.Logger.Error("Error saving user data", zap.Error(err))
		return fmt.Errorf("failed to save user data: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(userData.Cities); err != nil {
		s.Logger.Error("Error encoding response", zap.Error(err))
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}

// UpdateUserUnits allows you to update the global unit type. The options are metric and imperial.
func (s *Service) UpdateUserUnits(w http.ResponseWriter, r *http.Request) error {
	var reqBody struct {
		Units string `json:"units"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		s.Logger.Error("Invalid request body", zap.Error(err))
		return fmt.Errorf("invalid request body: %w", err)
	}

	if reqBody.Units != "metric" && reqBody.Units != "imperial" {
		s.Logger.Warn("Invalid unit type", zap.String("units", reqBody.Units))
		return fmt.Errorf("%w: %s", ErrInvalidUnit, reqBody.Units)
	}

	userData, err := s.Storage.LoadUserData()
	if err != nil {
		s.Logger.Error("Error loading user data", zap.Error(err))
		return fmt.Errorf("failed to load user data: %w", err)
	}

	userData.Units = reqBody.Units
	if err := s.Storage.SaveUserData(userData); err != nil {
		s.Logger.Error("Error saving user data", zap.Error(err))
		return fmt.Errorf("failed to save user data: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]string{"units": userData.Units}); err != nil {
		s.Logger.Error("Error encoding response", zap.Error(err))
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}

// doRequest validates a url, sets up a context, and performs an HTTP request.
func (s *Service) doRequest(method, rawURL string, body io.Reader) (*http.Response, error) {
	validatedURL, err := s.validateURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to validate URL: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, validatedURL, body)
	if err != nil {
		s.Logger.Error("Failed to create HTTP request", zap.String("url", rawURL), zap.Error(err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.Logger.Error("Failed to perform HTTP request", zap.String("url", validatedURL), zap.Error(err))
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	return res, nil
}

// validateStreamURL will validate a url.
func (s *Service) validateURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
		return "", fmt.Errorf("%w: %s", ErrInvalidURL, rawURL)
	}

	return parsedURL.String(), nil
}

// GetGeocode returns the geocode for a city. It's primarily used by open-meteo endpoints which only accept lat/lon.
func (s *Service) getGeocode(city string) (float64, float64, error) {
	geoURL := fmt.Sprintf(s.GeocodeAPIURL, url.QueryEscape(city))

	res, err := s.doRequest(http.MethodGet, geoURL, nil)
	if err != nil {
		s.Logger.Error("Failed to fetch geocode", zap.String("city", city), zap.Error(err))
		return 0, 0, fmt.Errorf("failed to get geocode: %w", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			s.Logger.Error("Error closing response body", zap.Error(err))
		}
	}()

	var geoData GeocodeResponse
	if err := json.NewDecoder(res.Body).Decode(&geoData); err != nil || len(geoData.Results) == 0 {
		s.Logger.Error("No geocode results", zap.String("city", city), zap.Error(err))
		return 0, 0, fmt.Errorf("%w: %s", ErrNoResultsForCity, city)
	}

	return geoData.Results[0].Latitude, geoData.Results[0].Longitude, nil
}

// GetWeatherData returns weather data based on the passed in url and struct.
func (s *Service) getWeatherData(city string, url string, respStruct interface{}) error {
	if city == "" {
		return ErrCityRequired
	}

	lat, lon, err := s.getGeocode(city)
	if err != nil {
		return fmt.Errorf("failed to get geocode: %w", err)
	}

	weatherURL := fmt.Sprintf(url, lat, lon)

	res, err := s.doRequest(http.MethodGet, weatherURL, nil)
	if err != nil {
		s.Logger.Error("Failed to get weather data", zap.String("url", weatherURL), zap.Error(err))
		return fmt.Errorf("failed to get weather data: %w", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			s.Logger.Error("Error closing response body", zap.Error(err))
		}
	}()

	if err := json.NewDecoder(res.Body).Decode(respStruct); err != nil {
		s.Logger.Error("Failed to decode weather data", zap.String("url", weatherURL), zap.Error(err))
		return fmt.Errorf("failed to decode weather data: %w", err)
	}

	return nil
}
