// Package main initializes the logger and routing.
// It also loads a config to configure the port and stream url.
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/codyonesock/rest_weather/internal/config"
	"github.com/codyonesock/rest_weather/internal/storage"
	"github.com/codyonesock/rest_weather/internal/weather"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 10 * time.Second
)

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

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Error loading config", zap.Error(err))
		return
	}

	weatherService := initializeServices(cfg, logger)
	r := setupRouter(weatherService, logger)

	logger.Info("Server running", zap.String("port", cfg.Port))
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}
}
func initializeServices(cfg *config.Config, logger *zap.Logger) *weather.Service {
	storageService := storage.NewStorageService(cfg.DatabaseURL, logger)
	weatherService := weather.NewWeatherService(
		logger,
		storageService,
		cfg.CurrentWeatherAPIURL,
		cfg.ForecastWeatherAPIURL,
		cfg.GeocodeAPIURL,
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
