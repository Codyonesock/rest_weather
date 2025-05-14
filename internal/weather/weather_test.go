package weather_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codyonesock/rest_weather/internal/shared"
	"github.com/codyonesock/rest_weather/internal/weather"
	"go.uber.org/zap"
)

type MockStorage struct {
	LoadUserDataFunc func() (shared.UserData, error)
	SaveUserDataFunc func(shared.UserData) error
}

func (m *MockStorage) LoadUserData() (shared.UserData, error) {
	return m.LoadUserDataFunc()
}

func (m *MockStorage) SaveUserData(data shared.UserData) error {
	return m.SaveUserDataFunc(data)
}

func setupMockWeatherService() (*weather.Service, *MockStorage) {
	mockStorage := &MockStorage{
		LoadUserDataFunc: nil,
		SaveUserDataFunc: nil,
	}
	logger, _ := zap.NewDevelopment()

	weatherService := weather.NewWeatherService(
		logger,
		mockStorage,
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true",
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&daily=temperature_2m_max,temperature_2m_min",
		"https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json",
	)

	return weatherService, mockStorage
}

func TestGetCurrentWeatherByCity(t *testing.T) {
	t.Parallel()

	weatherService, _ := setupMockWeatherService()

	rec := httptest.NewRecorder()

	_, err := weatherService.GetCurrentWeatherByCity(rec, "halifax")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetForecastByCity(t *testing.T) {
	t.Parallel()

	weatherService, _ := setupMockWeatherService()

	rec := httptest.NewRecorder()

	_, err := weatherService.GetForecastByCity(rec, "halifax")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := response["daily"]; !ok {
		t.Errorf("expected 'daily' in response, got %v", response)
	}
}

func TestGetUserData(t *testing.T) {
	t.Parallel()

	weatherService, mockStorage := setupMockWeatherService()

	mockStorage.LoadUserDataFunc = func() (shared.UserData, error) {
		return shared.UserData{
			Cities: []string{"Halifax", "Berlin"},
			Units:  "metric",
		}, nil
	}

	rec := httptest.NewRecorder()

	_, err := weatherService.GetUserData(rec)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var response shared.UserData
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Cities) != 2 || response.Cities[0] != "Halifax" || response.Cities[1] != "Berlin" {
		t.Errorf("expected cities to be ['Halifax', 'Berlin'], got %v", response.Cities)
	}

	if response.Units != "metric" {
		t.Errorf("expected units to be 'metric', got %v", response.Units)
	}
}

func TestAddCity(t *testing.T) {
	t.Parallel()

	weatherService, mockStorage := setupMockWeatherService()

	mockStorage.LoadUserDataFunc = func() (shared.UserData, error) {
		return shared.UserData{
			Cities: []string{"Halifax"},
			Units:  "metric",
		}, nil
	}
	mockStorage.SaveUserDataFunc = func(data shared.UserData) error {
		if len(data.Cities) != 2 || !strings.Contains(strings.Join(data.Cities, ","), "Berlin") {
			t.Errorf("expected cities to include 'Berlin', got %v", data.Cities)
		}

		return nil
	}

	rec := httptest.NewRecorder()

	err := weatherService.AddCity(rec, "Berlin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestDeleteCity(t *testing.T) {
	t.Parallel()

	weatherService, mockStorage := setupMockWeatherService()

	mockStorage.LoadUserDataFunc = func() (shared.UserData, error) {
		return shared.UserData{
			Cities: []string{"Halifax", "Berlin"},
			Units:  "metric",
		}, nil
	}
	mockStorage.SaveUserDataFunc = func(data shared.UserData) error {
		if len(data.Cities) != 1 || data.Cities[0] != "Halifax" {
			t.Errorf("expected cities to only include 'Halifax', got %v", data.Cities)
		}

		return nil
	}

	rec := httptest.NewRecorder()

	err := weatherService.DeleteCity(rec, "Berlin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUpdateUserUnits(t *testing.T) {
	t.Parallel()

	weatherService, mockStorage := setupMockWeatherService()

	mockStorage.LoadUserDataFunc = func() (shared.UserData, error) {
		return shared.UserData{
			Cities: []string{},
			Units:  "metric",
		}, nil
	}
	mockStorage.SaveUserDataFunc = func(data shared.UserData) error {
		if data.Units != "imperial" {
			t.Errorf("expected units to be 'imperial', got %v", data.Units)
		}

		return nil
	}

	body := bytes.NewBufferString(`{"units": "imperial"}`)
	req := httptest.NewRequest(http.MethodPut, "/user/units", body)
	rec := httptest.NewRecorder()

	err := weatherService.UpdateUserUnits(rec, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}
