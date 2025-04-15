package storage_test

import (
	"os"
	"testing"

	"go.uber.org/zap"

	"github.com/codyonesock/rest_weather/internal/models"
	"github.com/codyonesock/rest_weather/internal/storage"
)

func setupTestStorage(t *testing.T) (*storage.Service, func()) {
	t.Helper()

	tempFile := t.TempDir() + "/test_userdata.json"

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	storageService := storage.NewStorageService(tempFile, logger)

	cleanup := func() {
		if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
			t.Errorf("failed to remove temp file: %v", err)
		}
	}

	return storageService, cleanup
}

func TestLoadUserData(t *testing.T) {
	t.Parallel()

	storageService, cleanup := setupTestStorage(t)
	defer cleanup()

	userData, err := storageService.LoadUserData()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(userData.Cities) != 0 {
		t.Errorf("expected no cities, got %v", userData.Cities)
	}

	if userData.Units != "metric" {
		t.Errorf("expected units to be 'metric', got %v", userData.Units)
	}
}

func TestSaveUserData(t *testing.T) {
	t.Parallel()

	storageService, cleanup := setupTestStorage(t)
	defer cleanup()

	userData := models.UserData{
		Cities: []string{"Halifax"},
		Units:  "metric",
	}

	if err := storageService.SaveUserData(userData); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	loadedData, err := storageService.LoadUserData()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(loadedData.Cities) != 1 || loadedData.Cities[0] != "Halifax" {
		t.Errorf("expected cities to be ['Halifax'], got %v", loadedData.Cities)
	}

	if loadedData.Units != "metric" {
		t.Errorf("expected units to be 'metric', got %v", loadedData.Units)
	}
}
