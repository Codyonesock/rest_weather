package weather

// GeoCodeResponse is a struct based on geocode data returned from open-meteo
type GeocodeResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

// CurrentWeatherResponse is a struct based on current weather data returned from open-meteo
type CurrentWeatherResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
	} `json:"current_weather"`
}

// ForecastResponse is a struct based on forecast data returned from open-meteo
type ForecastResponse struct {
	Daily struct {
		Dates    []string  `json:"time"`
		MaxTemps []float64 `json:"temperature_2m_max"`
		MinTemps []float64 `json:"temperature_2m_min"`
	} `json:"daily"`
}

// UserData is a struct that represents the local userdata.json file used to track preferences
type UserData struct {
	Cities []string `json:"cities"`
	Units  string   `json:"units"`
}
