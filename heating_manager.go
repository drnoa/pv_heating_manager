package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ShellyURL          string `json:"shellyURL"`
	ShellyHeatingOnURL string `json:"shellyHeatingOnURL"`
}

var (
	temperatureExceeded bool
	checkInterval       = 5 * time.Minute
)

func getTemperature(shellyTempURL string) (float64, error) {
	resp, err := http.Get(shellyTempURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var temperatureStr string
	_, err = fmt.Fscan(resp.Body, &temperatureStr)
	if err != nil {
		return 0, err
	}

	temperature, err := strconv.ParseFloat(temperatureStr, 64)
	if err != nil {
		return 0, err
	}

	return temperature, nil
}

func checkTemperature(shellyURL string, shellyHeatingOnURL string) {
	temperature, err := getTemperature(shellyURL)
	if err != nil {
		log.Printf("Fehler beim Abrufen der Temperatur: %v", err)
		return
	}

	if temperature > 55 {
		fmt.Println("Warnung: Die Temperatur hat 55°C überschritten!")
		temperatureExceeded = true
	} else {
		fmt.Println("Temperatur ist in Ordnung.")
	}
}

func turnShellyOn(shellyHeatingOnURL string) error {
	resp, err := http.Get(shellyHeatingOnURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to turn on Shelly, status code: %d", resp.StatusCode)
	}

	fmt.Println("Shelly eingeschaltet.")
	return nil
}

func weeklyCheck(shellyHeatingOnURL string) {
	if !temperatureExceeded {
		if err := turnShellyOn(shellyHeatingOnURL); err != nil {
			log.Printf("Fehler beim Einschalten des Shelly: %v", err)
		}
	}
	// Setze den Zustand für die neue Woche zurück
	temperatureExceeded = false
}

func main() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Fehler beim Öffnen der Konfigurationsdatei: %v", err)
	}
	defer configFile.Close()

	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		log.Fatalf("Fehler beim Parsen der Konfigurationsdatei: %v", err)
	}

	ticker := time.NewTicker(checkInterval)
	weeklyTicker := time.NewTicker(7 * 24 * time.Hour)
	defer ticker.Stop()
	defer weeklyTicker.Stop()

	for {
		select {
		case <-ticker.C:
			checkTemperature(config.ShellyURL, config.ShellyHeatingOnURL)
		case <-weeklyTicker.C:
			weeklyCheck(config.ShellyHeatingOnURL)
		}
	}
}
