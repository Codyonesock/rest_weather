package weather

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetCurrentWeatherByCity(t *testing.T) {
	req, err := http.NewRequest("GET", "/weather?city=Halifax", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(GetCurrentWeatherByCity)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HandleFunc returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGetForecastByCity(t *testing.T) {
	req, err := http.NewRequest("GET", "/forecast?city=Halifax", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(GetForecastByCity)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HandleFunc returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGetUserData(t *testing.T) {
	req, err := http.NewRequest("GET", "/user/data", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(GetUserData)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HandleFunc returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	os.Remove("userdata.json")
}

func TestAddOrDeleteUserCity_Add(t *testing.T) {
	req, err := http.NewRequest("POST", "/user/cities/Halifax", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(AddOrDeleteUserCity)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HandleFunc returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	os.Remove("userdata.json")
}

func TestAddOrDeleteUserCity_Delete(t *testing.T) {
	// Add a city to delete later
	addReq, err := http.NewRequest("POST", "/user/cities/Halifax", nil)
	if err != nil {
		t.Fatal(err)
	}
	addRecorder := httptest.NewRecorder()
	addHandler := http.HandlerFunc(AddOrDeleteUserCity)
	addHandler.ServeHTTP(addRecorder, addReq)

	req, err := http.NewRequest("DELETE", "/user/cities/Halifax", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(AddOrDeleteUserCity)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HandleFunc returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	os.Remove("userdata.json")
}

func TestUpdateUserUnits(t *testing.T) {
	var jsonStr = []byte(`{"units": "metric"}`)
	req, err := http.NewRequest("PUT", "/user/units", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateUserUnits)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("HandleFunc returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	os.Remove("userdata.json")
}
