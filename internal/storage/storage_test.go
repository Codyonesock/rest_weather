package storage

import (
	"os"
	"testing"

	"github.com/codyonesock/rest_weather/internal/models"
)

func TestLoadUserData(t *testing.T) {
	os.Remove("userdata.json")

	_, err := LoadUserData()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	os.Remove("userdata.json")
}

func TestSaveUserData(t *testing.T) {
	userData := models.UserData{
		Cities: []string{"Halifax"},
		Units:  "metric",
	}

	err := SaveUserData(userData)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	loadedData, err := LoadUserData()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(loadedData.Cities) != 1 || loadedData.Cities[0] != "Halifax" {
		t.Errorf("expected cities to be ['Halifax'], got %v", loadedData.Cities)
	}

	if loadedData.Units != "metric" {
		t.Errorf("expected units to be 'metric', got %v", loadedData.Units)
	}

	os.Remove("userdata.json")
}
