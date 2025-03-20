package util

import (
	"testing"
)

func TestGetGeocode(t *testing.T) {
	lat, lon, err := getGeocode("Halifax")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lat == 0 || lon == 0 {
		t.Errorf("expected valid latitude and longitude, got %v, %v", lat, lon)
	}
}
