package weather

import (
	"encoding/json"
	"os"
)

// LoadUserData loads the data from a local userdata.json file, if it doesn't exist it creates a default one
func loadUserData() (UserData, error) {
	file, err := os.Open("userdata.json")
	if err != nil {
		if os.IsNotExist(err) {
			defaultData := UserData{
				Cities: []string{},
				Units:  "metric",
			}

			file, err = os.Create("userdata.json")
			if err != nil {
				return UserData{}, err
			}

			json.NewEncoder(file).Encode(defaultData)
			return defaultData, nil
		}
		return UserData{}, nil
	}
	defer file.Close()

	var userData UserData
	if err := json.NewDecoder(file).Decode(&userData); err != nil {
		return UserData{}, err
	}
	return userData, nil
}

// SaveUserData saves userData to a local userdata.json file
func saveUserData(userData UserData) error {
	file, err := os.Create("userdata.json")
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(userData)
}
