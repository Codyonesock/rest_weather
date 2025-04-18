// Package main initializes the logger and routing.
// It also loads a config to configure the port and stream url.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/codyonesock/rest_weather/internal/storage"
	"github.com/codyonesock/rest_weather/internal/weather"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 10 * time.Second
)

type config struct {
	Port                  string `json:"port"`
	CurrentWeatherAPIURL  string `json:"current_weather_api_url"`
	ForecastWeatherAPIURL string `json:"forecast_weather_api_url"`
	GeocodeAPIURL         string `json:"geocode_api_url"`
	DatabaseURL           string `json:"database_url"`
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Fatal("Error syncing logger", zap.Error(err))
		}
	}()

	config, err := loadConfig("config.json", logger)
	if err != nil {
		logger.Fatal("Error loading config", zap.Error(err))
		return
	}

	weatherService := initializeServices(config, logger)

	r := setupRouter(weatherService, logger)

	logger.Info("Server running", zap.String("port", config.Port))
	server := &http.Server{
		Addr:         config.Port,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}
}

func loadConfig(filename string, l *zap.Logger) (*config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		l.Error("Error reading config file", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		l.Error("Error unmarshalling JSON", zap.Error(err))
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	l.Info("Config loaded",
		zap.String("port", cfg.Port),
		zap.String("current_weather_api_url", cfg.CurrentWeatherAPIURL),
		zap.String("forecast_weather_api_url", cfg.ForecastWeatherAPIURL),
		zap.String("geocode_api_url", cfg.GeocodeAPIURL),
		zap.String("database_url", cfg.DatabaseURL),
	)

	return &cfg, nil
}

func initializeServices(config *config, logger *zap.Logger) *weather.Service {
	storageService := storage.NewStorageService(config.DatabaseURL, logger)
	weatherService := weather.NewWeatherService(
		logger,
		storageService,
		config.CurrentWeatherAPIURL,
		config.ForecastWeatherAPIURL,
		config.GeocodeAPIURL,
	)

	return weatherService
}

func setupRouter(weatherService weather.ServiceInterface, logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/weather/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if _, err := weatherService.GetCurrentWeatherByCity(w, city); err != nil {
			logger.Error("Error getting current weather", zap.String("city", city), zap.Error(err))
			http.Error(w, "Error getting current weather", http.StatusInternalServerError)
		}
	})

	r.Get("/forecast/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if _, err := weatherService.GetForecastByCity(w, city); err != nil {
			logger.Error("Error getting forecast data", zap.String("city", city), zap.Error(err))
			http.Error(w, "Error getting forecast data", http.StatusInternalServerError)
		}
	})

	r.Get("/user/data", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := weatherService.GetUserData(w); err != nil {
			logger.Error("Error getting user data", zap.Error(err))
			http.Error(w, "Error getting user data", http.StatusInternalServerError)
		}
	})

	r.Post("/user/cities/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if err := weatherService.AddCity(w, city); err != nil {
			logger.Error("Error adding city to user data", zap.String("city", city), zap.Error(err))
			http.Error(w, "Error adding city to user data", http.StatusInternalServerError)
		}
	})

	r.Delete("/user/cities/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if err := weatherService.DeleteCity(w, city); err != nil {
			logger.Error("Error deleting city from user data", zap.String("city", city), zap.Error(err))
			http.Error(w, "Error deleting city from user data", http.StatusInternalServerError)
		}
	})

	r.Put("/user/units", func(w http.ResponseWriter, r *http.Request) {
		if err := weatherService.UpdateUserUnits(w, r); err != nil {
			logger.Error("Error updating units in user data", zap.Error(err))
			http.Error(w, "Error updating units in user data", http.StatusInternalServerError)
		}
	})

	return r
}
