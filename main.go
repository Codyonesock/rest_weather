package main

import (
	"fmt"
	"net/http"

	"github.com/codyonesock/rest_weather/weather"
)

func main() {
	// curl -X GET http://localhost:8080/weather\?city\=halifax
	http.HandleFunc("/weather", weather.GetCurrentWeatherByCity)

	// curl -X GET http://localhost:8080/forecast\?city\=halifax
	http.HandleFunc("/forecast", weather.GetForecastByCity)

	// curl -X GET http://localhost:8080/user/data
	http.HandleFunc("/user/data", weather.GetUserData)

	// curl -X POST http://localhost:8080/user/cities/halifax
	// curl -X DELETE http://localhost:8080/user/cities/halifax
	http.HandleFunc("/user/cities/", weather.AddOrDeleteUserCity)

	// curl -X PUT "http://localhost:8080/user/units" -H "Content-Type: application/json" -d '{"units": "metric"}'
	// curl -X PUT "http://localhost:8080/user/units" -H "Content-Type: application/json" -d '{"units": "imperial"}'
	http.HandleFunc("/user/units", weather.UpdateUserUnits)

	port := ":8080"
	fmt.Printf("Server running on port %s\n", port)
	http.ListenAndServe(port, nil)
}
