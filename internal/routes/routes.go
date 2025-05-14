// Package routes will help mount and handle routes.
package routes

import (
	"net/http"

	"github.com/codyonesock/rest_weather/internal/weather"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// RegisterRoutes sets up all the app routes.
func RegisterRoutes(
	r *chi.Mux,
	weatherService *weather.Service,
) {
	r.Route("/weather", func(r chi.Router) {
		r.Get("/{city}", getCurrentWeatherHandler(weatherService))
	})

	r.Route("/forecast", func(r chi.Router) {
		r.Get("/{city}", getForecastHandler(weatherService))
	})

	r.Route("/user", func(r chi.Router) {
		r.Get("/data", getUserDataHandler(weatherService))
		r.Post("/cities/{city}", addCityHandler(weatherService))
		r.Delete("/cities/{city}", deleteCityHandler(weatherService))
		r.Put("/units", updateUserUnitsHandler(weatherService))
	})
}

func getCurrentWeatherHandler(weatherService *weather.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if _, err := weatherService.GetCurrentWeatherByCity(w, city); err != nil {
			weatherService.Logger.Error("Error getting current weather", zap.String("city", city), zap.Error(err))
			http.Error(w, "Error getting current weather", http.StatusInternalServerError)
		}
	}
}

func getForecastHandler(weatherService *weather.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if _, err := weatherService.GetForecastByCity(w, city); err != nil {
			weatherService.Logger.Error("Error getting forecast data", zap.String("city", city), zap.Error(err))
			http.Error(w, "Error getting forecast data", http.StatusInternalServerError)
		}
	}
}
func getUserDataHandler(weatherService *weather.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if _, err := weatherService.GetUserData(w); err != nil {
			weatherService.Logger.Error("Error getting user data", zap.Error(err))
			http.Error(w, "Error getting user data", http.StatusInternalServerError)
		}
	}
}
func addCityHandler(weatherService *weather.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if err := weatherService.AddCity(w, city); err != nil {
			weatherService.Logger.Error("Error adding city to user data", zap.String("city", city), zap.Error(err))
			http.Error(w, "Error adding city to user data", http.StatusInternalServerError)
		}
	}
}
func deleteCityHandler(weatherService *weather.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if err := weatherService.DeleteCity(w, city); err != nil {
			weatherService.Logger.Error("Error deleting city from user data", zap.String("city", city), zap.Error(err))
			http.Error(w, "Error deleting city from user data", http.StatusInternalServerError)
		}
	}
}
func updateUserUnitsHandler(weatherService *weather.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := weatherService.UpdateUserUnits(w, r); err != nil {
			weatherService.Logger.Error("Error updating units in user data", zap.Error(err))
			http.Error(w, "Error updating units in user data", http.StatusInternalServerError)
		}
	}
}
