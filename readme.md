# Weather REST API

A simple REST API for retrieving weather data and managing user preferences. This project uses the [Open-Meteo API](https://open-meteo.com/en/docs) to fetch weather information and stores user data.

## Features

- **Weather Data**
  - `GET /weather/{city}`: Get the current weather for a city.
  - `GET /forecast/{city}`: Get a 7-day weather forecast for a city.
- **User Preferences**
  - `GET /user/data`: Retrieve user preferences (saved cities and units).
  - `POST /user/cities/{city}`: Add a city to the user's saved list.
  - `DELETE /user/cities/{city}`: Remove a city from the user's saved list.
  - `PUT /user/units`: Update the preferred unit type (`metric` or `imperial`).

## Example Commands

```sh
curl -X GET http://localhost:8080/weather/halifax
curl -X GET http://localhost:8080/forecast/halifax
curl -X GET http://localhost:8080/user/data
curl -X POST http://localhost:8080/user/cities/halifax
curl -X POST http://localhost:8080/user/cities/halifax,berlin
curl -X DELETE http://localhost:8080/user/cities/halifax
curl -X DELETE http://localhost:8080/user/cities/halifax,berlin
curl -X PUT "http://localhost:8080/user/units" -H "Content-Type: application/json" -d '{"units": "metric"}'
curl -X PUT "http://localhost:8080/user/units" -H "Content-Type: application/json" -d '{"units": "imperial"}'
```

## .env example

```env
PORT=:8080
CURRENT_WEATHER_API_URL=https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true
FORECAST_WEATHER_API_URL=https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&daily=temperature_2m_max,temperature_2m_min
GEOCODE_API_URL=https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json
DATABASE_URL=userdata.json
LOG_LEVEL=DEBUG
```

## Testing

```sh
go test ./internal/... -race
```

## Linting
```sh
golangci-lint run ./...
```