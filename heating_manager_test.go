package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// TestNewHeatingManager prüft, ob ein HeatingManager erfolgreich initialisiert werden kann.
func TestNewHeatingManager(t *testing.T) {
	manager, err := NewHeatingManager()
	if err != nil {
		t.Fatalf("Failed to create HeatingManager: %v", err)
	}
	if manager == nil {
		t.Fatal("Expected non-nil HeatingManager instance")
	}
}

// TestCheckTemperature verwendet einen Mock-Server, um die Reaktion von checkTemperature zu testen.
func TestCheckTemperature(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("25")) // Setzen Sie eine Mock-Temperatur.
	}))
	defer ts.Close()

	manager, _ := NewHeatingManager()
	manager.Config.ShellyURL = ts.URL // Verwenden Sie den Mock-Server als ShellyURL.

	manager.checkTemperature(manager.Config.ShellyURL)
	if manager.TemperatureExceeded {
		t.Error("TemperatureExceeded should be false for temperature 25")
	}
}

// TestWeeklyCheck prüft die Logik der weeklyCheck Methode.
// Sie können diese Methode erweitern, um verschiedene Szenarien zu testen.
func TestWeeklyCheck(t *testing.T) {
	manager, _ := NewHeatingManager()
	manager.weeklyCheck("someURL") // Dieser Test könnte ausgebaut werden, um spezifische Verhaltensweisen zu prüfen.
}

// TestGetTemperature prüft, ob die Temperatur korrekt vom HTTP-Response gelesen wird.
func TestGetTemperature(t *testing.T) {
	expectedTemp := 25.0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strconv.FormatFloat(expectedTemp, 'f', -1, 64)))
	}))
	defer ts.Close()

	temp, err := getTemperature(ts.URL)
	if err != nil {
		t.Errorf("getTemperature returned an error: %v", err)
	}
	if temp != expectedTemp {
		t.Errorf("Expected %v, got %v", expectedTemp, temp)
	}
}

// Weitere Tests für turnShellyOn, saveLastCheckTime, nextWeeklyCheckDuration, und readLastCheckTime können hier hinzugefügt werden.
