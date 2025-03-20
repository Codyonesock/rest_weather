package weather

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/codyonesock/rest_weather/internal/models"
	"github.com/codyonesock/rest_weather/internal/storage"
	"github.com/codyonesock/rest_weather/internal/util"
)

const openMeteoBaseUrl = "https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f"

// GetCurrentWeatherByCity returns the current weather (temperature and weather speed) using the lat/lon of the city entered
func GetCurrentWeatherByCity(w http.ResponseWriter, r *http.Request) {
	var weatherData models.CurrentWeatherResponse
	util.GetWeatherData(w, r, openMeteoBaseUrl+"&current_weather=true", &weatherData)
}

// GetForecastByCity returns a 7 day forecast (dates, min/max temps) using the lat/lon of the city entered
func GetForecastByCity(w http.ResponseWriter, r *http.Request) {
	var forecastData models.ForecastResponse
	util.GetWeatherData(w, r, openMeteoBaseUrl+"&daily=temperature_2m_max,temperature_2m_min", &forecastData)
}

// GetUserData returns user data that's read from a local json file
func GetUserData(w http.ResponseWriter, r *http.Request) {
	userData, err := storage.LoadUserData()
	if err != nil {
		http.Error(w, "error loading user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userData)
}

// AddOrDeleteUserCity handles POST and DELETE requests for /user/cities/
func AddOrDeleteUserCity(w http.ResponseWriter, r *http.Request) {
	city := strings.TrimPrefix(r.URL.Path, "/user/cities/")
	if city == "" {
		http.Error(w, "city is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "POST":
		util.AddCity(w, city)
	case "DELETE":
		util.DeleteCity(w, city)
	default:
		http.Error(w, "get or post only", http.StatusMethodNotAllowed)
	}
}

// UpdateUserUnits allows you to update the global unit type. The options are metric and imperial
func UpdateUserUnits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "please use a put", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Units string `json:"units"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if reqBody.Units != "metric" && reqBody.Units != "imperial" {
		http.Error(w, "invalid unit, choose metric or imperial", http.StatusBadRequest)
		return
	}

	UserData, err := storage.LoadUserData()
	if err != nil {
		http.Error(w, "error loading user data", http.StatusInternalServerError)
		return
	}

	UserData.Units = reqBody.Units
	if err := storage.SaveUserData(UserData); err != nil {
		http.Error(w, "error saving user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"units": UserData.Units})
}
