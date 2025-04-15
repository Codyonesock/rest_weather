// Package models is for shared models.
package models

// UserData is a struct that represents the local userdata.json file used to track preferences.
type UserData struct {
	Cities []string `json:"cities"`
	Units  string   `json:"units"`
}
