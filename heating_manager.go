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
	lastCheckFile       = "lastCheck.txt"
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

func checkTemperature(shellyURL string) {
	temperature, err := getTemperature(shellyURL)
	if err != nil {
		log.Printf("Fehler beim Abrufen der Temperatur: %v", err)
		return
	}

	if temperature > 55 {
		fmt.Println("Die Temperatur hat 55°C überschritten! Legionellenschaltung wird verschoben.")
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
	temperatureExceeded = false
	saveLastCheckTime()
}

func saveLastCheckTime() {
	now := time.Now()
	err := os.WriteFile(lastCheckFile, []byte(now.Format(time.RFC3339)), 0644)
	if err != nil {
		log.Printf("Fehler beim Speichern des letzten Überprüfungszeitpunkts: %v", err)
	}
}

func getLastCheckTime() time.Time {
	if _, err := os.Stat(lastCheckFile); os.IsNotExist(err) {
		// Datei existiert nicht, also führe sofort eine Überprüfung durch und speichere die Zeit
		now := time.Now()
		saveLastCheckTime()
		return now
	}

	data, err := os.ReadFile(lastCheckFile)
	if err != nil {
		log.Printf("Fehler beim Lesen des letzten Überprüfungszeitpunkts: %v", err)
		return time.Now()
	}

	lastCheck, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		log.Printf("Fehler beim Parsen des letzten Überprüfungszeitpunkts: %v", err)
		return time.Now()
	}

	return lastCheck
}

func nextWeeklyCheckDuration() time.Duration {
	lastCheck := getLastCheckTime()
	nextCheck := lastCheck.Add(7 * 24 * time.Hour)
	if time.Now().After(nextCheck) {
		return 0 // Sollte sofort prüfen, wenn der nächste Check bereits fällig ist
	}
	return nextCheck.Sub(time.Now())
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
	defer ticker.Stop()

	weeklyCheckTimer := time.NewTimer(nextWeeklyCheckDuration())
	defer weeklyCheckTimer.Stop()

	for {
		select {
		case <-ticker.C:
			checkTemperature(config.ShellyURL)
		case <-weeklyCheckTimer.C:
			weeklyCheck(config.ShellyHeatingOnURL)
			weeklyTicker := time.NewTicker(7 * 24 * time.Hour)
			defer weeklyTicker.Stop()
			go func() {
				for {
					select {
					case <-weeklyTicker.C:
						weeklyCheck(config.ShellyHeatingOnURL)
					}
				}
			}()
			weeklyCheckTimer.Stop()
		}
	}
}
