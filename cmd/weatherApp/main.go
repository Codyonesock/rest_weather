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
	"github.com/codyonesock/rest_weather/internal/logger"
	"github.com/codyonesock/rest_weather/internal/routes"
	"github.com/codyonesock/rest_weather/internal/storage"
	"github.com/codyonesock/rest_weather/internal/weather"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 10 * time.Second
)

func main() {
	cfg := loadConfig()
	logger := initializeLogger(cfg)

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush logger: %v\n", err)
		}
	}()

	weatherService := initializeServices(cfg, logger)
	startServer(cfg, logger, weatherService)
}

// loadConfig loads the config.
func loadConfig() *config.Config {
	config, err := config.LoadConfig()
	if err != nil {
		zap.L().Fatal("Error loading config", zap.Error(err))
		os.Exit(1)
	}

	return config
}

// initializeLogger sets up the zap logger.
func initializeLogger(config *config.Config) *zap.Logger {
	logger, err := logger.CreateLogger(config.LogLevel)
	if err != nil {
		zap.L().Fatal("Failed to initialize logger", zap.Error(err))
		os.Exit(1)
	}

	return logger
}

// initializeServices sets up services and returns a weatherService.
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

// startServer sets up the routes and starts the server.
func startServer(cfg *config.Config, logger *zap.Logger, weatherService *weather.Service) {
	r := chi.NewRouter()
	routes.RegisterRoutes(r, weatherService)

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
