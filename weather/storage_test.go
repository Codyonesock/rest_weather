package weather

import (
	"os"
	"testing"
)

func TestLoadUserData(t *testing.T) {
	os.Remove("userdata.json")

	_, err := loadUserData()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	os.Remove("userdata.json")
}

func TestSaveUserData(t *testing.T) {
	userData := UserData{
		Cities: []string{"Halifax"},
		Units:  "imperial",
	}

	err := saveUserData(userData)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	loadedData, err := loadUserData()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(loadedData.Cities) != 1 || loadedData.Cities[0] != "Halifax" {
		t.Errorf("expected cities to be ['Halifax'], got %v", loadedData.Cities)
	}

	if loadedData.Units != "imperial" {
		t.Errorf("expected units to be 'imperial', got %v", loadedData.Units)
	}

	os.Remove("userdata.json")
}
