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
	"time"

	"go.uber.org/zap"
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

// ServiceInterface depicts the interface for the weather package.
type ServiceInterface interface {
	GetCurrentWeatherByCity(w http.ResponseWriter, city string) (*CurrentWeatherResponse, error)
}

// Service handles dependencies and config.
type Service struct {
	Logger        *zap.Logger
	WeatherAPIURL string
	GeocodeAPIURL string
}

// NewWeatherService create a new instance of Service.
func NewWeatherService(l *zap.Logger, weatherAPIURL, geocodeAPIURL string) *Service {
	return &Service{
		Logger:        l,
		WeatherAPIURL: weatherAPIURL,
		GeocodeAPIURL: geocodeAPIURL,
	}
}

const contextTimeout = 5 * time.Second

// err113 demands no dynamic errors!
var (
	errInvalidURL       = errors.New("invalid URL")
	errNoResultsForCity = errors.New("no results for city")
	errCityRequired     = errors.New("city is required")
)

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
		return "", fmt.Errorf("%w: %s", errInvalidURL, rawURL)
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
		return 0, 0, fmt.Errorf("%w: %s", errNoResultsForCity, city)
	}

	return geoData.Results[0].Latitude, geoData.Results[0].Longitude, nil
}

// GetWeatherData returns weather data based on the passed in url and struct.
func (s *Service) getWeatherData(city string, url string, respStruct interface{}) error {
	if city == "" {
		return errCityRequired
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

// GetCurrentWeatherByCity returns the current weather (temperature and weather speed).
// It uses the lat/lon of the city entered.
func (s *Service) GetCurrentWeatherByCity(
	w http.ResponseWriter,
	city string,
) (*CurrentWeatherResponse, error) {
	var weatherData CurrentWeatherResponse
	if err := s.getWeatherData(city, s.WeatherAPIURL, &weatherData); err != nil {
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

// GetForecastByCity returns a 7 day forecast (dates, min/max temps) using the lat/lon of the city entered
// func GetForecastByCity(w http.ResponseWriter, r *http.Request) {
// 	var forecastData models.ForecastResponse
// 	util.GetWeatherData(w, r, openMeteoBaseURL+"&daily=temperature_2m_max,temperature_2m_min", &forecastData)
// }

// // GetUserData returns user data that's read from a local json file.
// func GetUserData(w http.ResponseWriter, r *http.Request) {
// 	userData, err := storage.LoadUserData()
// 	if err != nil {
// 		http.Error(w, "error loading user data", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(userData)
// }

// // AddOrDeleteUserCity handles POST and DELETE requests for /user/cities/.
// func AddOrDeleteUserCity(w http.ResponseWriter, r *http.Request) {
// 	city := strings.TrimPrefix(r.URL.Path, "/user/cities/")
// 	if city == "" {
// 		http.Error(w, "city is required", http.StatusBadRequest)
// 		return
// 	}

// 	switch r.Method {
// 	case "POST":
// 		// util.AddCity(w, city)
// 	case "DELETE":
// 		// util.DeleteCity(w, city)
// 	default:
// 		http.Error(w, "get or post only", http.StatusMethodNotAllowed)
// 	}
// }

// // UpdateUserUnits allows you to update the global unit type. The options are metric and imperial.
// func UpdateUserUnits(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPut {
// 		http.Error(w, "please use a put", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var reqBody struct {
// 		Units string `json:"units"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
// 		http.Error(w, "invalid body", http.StatusBadRequest)
// 		return
// 	}

// 	if reqBody.Units != "metric" && reqBody.Units != "imperial" {
// 		http.Error(w, "invalid unit, choose metric or imperial", http.StatusBadRequest)
// 		return
// 	}

// 	UserData, err := storage.LoadUserData()
// 	if err != nil {
// 		http.Error(w, "error loading user data", http.StatusInternalServerError)
// 		return
// 	}

// 	UserData.Units = reqBody.Units
// 	if err := storage.SaveUserData(UserData); err != nil {
// 		http.Error(w, "error saving user data", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{"units": UserData.Units})
// }

// // AddCity works in conjunction with weather/AddOrDeleteUserCity. It adds a city to the tracked list.
// func AddCity(w http.ResponseWriter, city string) {
// 	userData, err := storage.LoadUserData()
// 	if err != nil {
// 		http.Error(w, "error loading user data", http.StatusInternalServerError)
// 		return
// 	}

// 	for _, existingCity := range userData.Cities {
// 		if strings.EqualFold(existingCity, city) {
// 			http.Error(w, "city already tracked", http.StatusBadRequest)
// 			return
// 		}
// 	}

// 	caser := cases.Title(language.BritishEnglish)
// 	titleCity := caser.String(city)
// 	userData.Cities = append(userData.Cities, titleCity)

// 	if err := storage.SaveUserData(userData); err != nil {
// 		http.Error(w, "error saving city", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(userData.Cities)
// }

// // DeleteCity works in conjunction with weather/AddOrDeleteUserCity. It removes a city in the tracked list.
// func DeleteCity(w http.ResponseWriter, city string) {
// 	userData, err := storage.LoadUserData()
// 	if err != nil {
// 		http.Error(w, "error loading user data", http.StatusInternalServerError)
// 		return
// 	}

// 	var cityFound bool
// 	for i, existingCity := range userData.Cities {
// 		if strings.EqualFold(existingCity, city) {
// 			userData.Cities = append(userData.Cities[:i], userData.Cities[i+1:]...)
// 			cityFound = true
// 			break
// 		}
// 	}

// 	if !cityFound {
// 		http.Error(w, "city not found", http.StatusNotFound)
// 		return
// 	}

// 	if err := storage.SaveUserData(userData); err != nil {
// 		http.Error(w, "error saving user data", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(userData.Cities)
// }
