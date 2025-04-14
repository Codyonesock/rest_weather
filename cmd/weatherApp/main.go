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
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 10 * time.Second
)

type config struct {
	Port string `json:"port"`
}

func loadConfig(filename string, l *zap.Logger) (*config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		l.Error("Error reading config file", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config config
	if err := json.Unmarshal(data, &config); err != nil {
		l.Error("Error unmarshalling JSON", zap.Error(err))
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	l.Info("Config loaded", zap.String("port", config.Port))

	return &config, nil
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Fatal("Error loading config", zap.Error(err))
		}
	}()

	config, err := loadConfig("config.json", logger)
	if err != nil {
		logger.Fatal("Error loading config", zap.Error(err))
		return
	}

	r := chi.NewRouter()

	// TODO: Fix this mess :D
	// http.HandleFunc("/weather", weather.GetCurrentWeatherByCity)
	// http.HandleFunc("/forecast", weather.GetForecastByCity)
	// http.HandleFunc("/user/data", weather.GetUserData)
	// http.HandleFunc("/user/cities/", weather.AddOrDeleteUserCity)
	// http.HandleFunc("/user/units", weather.UpdateUserUnits)

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
