package storage

import (
	"encoding/json"
	"os"

	"github.com/codyonesock/rest_weather/internal/models"
)

// LoadUserData loads the data from a local userdata.json file, if it doesn't exist it creates a default one
func LoadUserData() (models.UserData, error) {
	file, err := os.Open("userdata.json")
	if err != nil {
		if os.IsNotExist(err) {
			defaultData := models.UserData{
				Cities: []string{},
				Units:  "metric",
			}

			file, err = os.Create("userdata.json")
			if err != nil {
				return models.UserData{}, err
			}

			json.NewEncoder(file).Encode(defaultData)
			return defaultData, nil
		}
		return models.UserData{}, nil
	}
	defer file.Close()

	var userData models.UserData
	if err := json.NewDecoder(file).Decode(&userData); err != nil {
		return models.UserData{}, err
	}
	return userData, nil
}

// SaveUserData saves userData to a local userdata.json file
func SaveUserData(userData models.UserData) error {
	file, err := os.Create("userdata.json")
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(userData)
}
