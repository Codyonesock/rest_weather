# Weather REST API
This project is a simple REST API for retrieving weather data and managing user preferences. It uses the [Open-Meteo API](https://open-meteo.com/en/docs "Open-Meteo API") to fetch weather information and currently saves the user data to a local json file.

###### Features
- `GET/weather/{city}`
- `GET/forecast/{city}`
- `GET/user/data/`
- `POST/user/cities/{city}`
- `DELETE/user/cities/{city}`
- `PUT/user/cities, -d '{"units": "metric"}, -d '{"units": "imperial"}`

###### CMD
- `curl -X GET http://localhost:8080/weather/halifax`
- `curl -X GET http://localhost:8080/forecast/halifax`
- `curl -X GET http://localhost:8080/user/data`
- `curl -X POST http://localhost:8080/user/cities/halifax`
- `curl -X DELETE http://localhost:8080/user/cities/halifax`
- `curl -X PUT "http://localhost:8080/user/units" -H "Content-Type: application/json" -d '{"units": "metric"}'`
- `curl -X PUT "http://localhost:8080/user/units" -H "Content-Type: application/json" -d '{"units": "imperial"}'`

###### Testing
- ``