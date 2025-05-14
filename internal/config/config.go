// Package config is for wiring up config.
package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config is your config.
type Config struct {
	Port                  string `envconfig:"PORT" default:":8080"`
	CurrentWeatherAPIURL  string `envconfig:"CURRENT_WEATHER_API_URL"`
	ForecastWeatherAPIURL string `envconfig:"FORECAST_WEATHER_API_URL"`
	GeocodeAPIURL         string `envconfig:"GEOCODE_API_URL"`
	DatabaseURL           string `envconfig:"DATABASE_URL" default:"userdata.json"`
}

// LoadConfig loads the application config.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	return &cfg, nil
}
