// Package storage is used to setup a DB to save/load user data.
package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/codyonesock/rest_weather/internal/models"
)

// ServiceInterface depicts the interface for the storage package.
type ServiceInterface interface {
	LoadUserData() (models.UserData, error)
	SaveUserData(userData models.UserData) error
}

// Service for dependencies and config.
type Service struct {
	FilePath string
	Logger   *zap.Logger
}

// NewStorageService creates a new instance of Service.
func NewStorageService(filePath string, l *zap.Logger) *Service {
	return &Service{
		FilePath: filePath,
		Logger:   l,
	}
}

// LoadUserData loads the data from a local file. If it doesn't exist, it creates a default one.
func (s *Service) LoadUserData() (models.UserData, error) {
	file, err := os.Open(s.FilePath)

	if err != nil {
		if os.IsNotExist(err) {
			s.Logger.Info("creating default file", zap.String("filePath", s.FilePath))
			return s.createDefaultUserData()
		}

		s.Logger.Error("Failed to open file", zap.Error(err))

		return models.UserData{}, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			s.Logger.Error("Error closing file", zap.Error(err))
		}
	}()

	var userData models.UserData
	if err := json.NewDecoder(file).Decode(&userData); err != nil {
		s.Logger.Error("Failed to decode file", zap.Error(err))
		return models.UserData{}, fmt.Errorf("failed to decode file: %w", err)
	}

	return userData, nil
}

// SaveUserData saves user data to the local file.
func (s *Service) SaveUserData(userData models.UserData) error {
	file, err := os.Create(s.FilePath)
	if err != nil {
		s.Logger.Error("Failed to create file", zap.Error(err))
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			s.Logger.Error("Error closing file", zap.Error(err))
		}
	}()

	if err := json.NewEncoder(file).Encode(userData); err != nil {
		s.Logger.Error("Failed to save user data", zap.Error(err))
		return fmt.Errorf("failed to save user data: %w", err)
	}

	return nil
}

// createDefaultUserData creates a default user data file and returns the default data.
func (s *Service) createDefaultUserData() (models.UserData, error) {
	defaultData := models.UserData{
		Cities: []string{},
		Units:  "metric",
	}

	file, err := os.Create(s.FilePath)
	if err != nil {
		s.Logger.Error("Failed to create file", zap.Error(err))
		return models.UserData{}, fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			s.Logger.Error("Error closing file", zap.Error(err))
		}
	}()

	if err := json.NewEncoder(file).Encode(defaultData); err != nil {
		s.Logger.Error("Failed to write default data", zap.Error(err))
		return models.UserData{}, fmt.Errorf("failed to write default data: %w", err)
	}

	return defaultData, nil
}
