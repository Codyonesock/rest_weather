package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/codyonesock/rest_weather/internal/models"
	"github.com/codyonesock/rest_weather/internal/storage"
)

// GetGeocode returns the geocode for a city. It's primarily used by open-meteo endpoints which only accept lat/lon
func getGeocode(city string) (float64, float64, error) {
	geoURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json", url.QueryEscape(city))
	resp, err := http.Get(geoURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get geocode: %v", err)
	}
	defer resp.Body.Close()

	var geoData models.GeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoData); err != nil || len(geoData.Results) == 0 {
		return 0, 0, fmt.Errorf("no results for city: %s", city)
	}

	return geoData.Results[0].Latitude, geoData.Results[0].Longitude, nil
}

// GetWeatherData returns weather data based on the passed in url and struct
func GetWeatherData(w http.ResponseWriter, r *http.Request, url string, respStruct interface{}) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "city is required", http.StatusBadRequest)
		return
	}

	lat, lon, err := getGeocode(city)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	weatherURL := fmt.Sprintf(url, lat, lon)
	resp, err := http.Get(weatherURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get data: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(respStruct); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respStruct)
}

// AddCity works in conjunction with weather/AddOrDeleteUserCity. It adds a city to the tracked list
func AddCity(w http.ResponseWriter, city string) {
	userData, err := storage.LoadUserData()
	if err != nil {
		http.Error(w, "error loading user data", http.StatusInternalServerError)
		return
	}

	for _, existingCity := range userData.Cities {
		if strings.EqualFold(existingCity, city) {
			http.Error(w, "city already tracked", http.StatusBadRequest)
			return
		}
	}

	caser := cases.Title(language.BritishEnglish)
	titleCity := caser.String(city)
	userData.Cities = append(userData.Cities, titleCity)

	if err := storage.SaveUserData(userData); err != nil {
		http.Error(w, "error saving city", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userData.Cities)
}

// DeleteCity works in conjunction with weather/AddOrDeleteUserCity. It removes a city in the tracked list
func DeleteCity(w http.ResponseWriter, city string) {
	userData, err := storage.LoadUserData()
	if err != nil {
		http.Error(w, "error loading user data", http.StatusInternalServerError)
		return
	}

	var cityFound bool
	for i, existingCity := range userData.Cities {
		if strings.EqualFold(existingCity, city) {
			userData.Cities = append(userData.Cities[:i], userData.Cities[i+1:]...)
			cityFound = true
			break
		}
	}

	if !cityFound {
		http.Error(w, "city not found", http.StatusNotFound)
		return
	}

	if err := storage.SaveUserData(userData); err != nil {
		http.Error(w, "error saving user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userData.Cities)
}
