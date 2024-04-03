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

// Config represents the application configuration.
type Config struct {
	ShellyURL            string  `json:"shellyURL"`            // URL of the Shelly device temperature addon.
	ShellyHeatingOnURL   string  `json:"shellyHeatingOnURL"`   // URL to turn Shelly heating on.
	TemperatureThreshold float64 `json:"temperatureThreshold"` // Temperature threshold in Celsius.
}

// HeatingManager is the main application struct.
type HeatingManager struct {
	Config              Config        // Configuration.
	TemperatureExceeded bool          // Indicates if the temperature threshold has been exceeded.
	CheckInterval       time.Duration // Interval between temperature checks.
	LastCheckFile       string        // File to save and read the last check time.
}

// NewHeatingManager creates a new HeatingManager instance.
func NewHeatingManager() (*HeatingManager, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return &HeatingManager{
		Config:        config,
		CheckInterval: 5 * time.Minute,
		LastCheckFile: "lastCheck.txt",
	}, nil
}

// StartTemperatureMonitoring starts the temperature monitoring loop.
func (hm *HeatingManager) StartTemperatureMonitoring() {
	ticker := time.NewTicker(hm.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		hm.checkTemperature(hm.Config.ShellyURL)
	}
}

// StartWeeklyCheck starts the weekly check loop.
func (hm *HeatingManager) StartWeeklyCheck() {
	weeklyCheckTimer := time.NewTimer(hm.nextWeeklyCheckDuration())
	defer weeklyCheckTimer.Stop()

	for {
		select {
		case <-weeklyCheckTimer.C:
			hm.weeklyCheck(hm.Config.ShellyHeatingOnURL)
			weeklyCheckTimer.Reset(hm.nextWeeklyCheckDuration())
		}
	}
}

// loadConfig loads the application configuration from a JSON file.
func loadConfig() (Config, error) {
	var config Config
	configFile, err := os.Open("config.json")
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %v", err)
	}
	defer configFile.Close()

	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

// checkTemperature checks the temperature of a Shelly device.
func (hm *HeatingManager) checkTemperature(shellyURL string) {
	temperature, err := getTemperature(shellyURL)
	if err != nil {
		log.Printf("Failed to get temperature: %v", err)
		return
	}

	if temperature > hm.Config.TemperatureThreshold {
		fmt.Println("Temperature has exceeded %.1f°C! Legionellaheating will be resheduled.", hm.Config.TemperatureThreshold)
		hm.TemperatureExceeded = true
	} else {
		fmt.Println("Temperature is in order.")
		hm.TemperatureExceeded = false
	}
}

// getTemperature gets the temperature of a Shelly device.
func getTemperature(shellyTempURL string) (float64, error) {
	resp, err := http.Get(shellyTempURL)
	if err != nil {
		return 0, fmt.Errorf("failed to get temperature: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get temperature: status code %d", resp.StatusCode)
	}

	var temperatureStr string
	_, err = fmt.Fscan(resp.Body, &temperatureStr)
	if err != nil {
		return 0, fmt.Errorf("failed to read temperature: %v", err)
	}

	temperature, err := strconv.ParseFloat(temperatureStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse temperature: %v", err)
	}

	return temperature, nil
}

// weeklyCheck checks if the temperature threshold has been exceeded and turns on the Shelly heating if necessary.
func (hm *HeatingManager) weeklyCheck(shellyHeatingOnURL string) {
	if !hm.TemperatureExceeded {
		if err := hm.turnShellyOn(shellyHeatingOnURL); err != nil {
			log.Printf("Failed to turn on Shelly: %v", err)
		}
	}
	hm.TemperatureExceeded = false
	hm.saveLastCheckTime()
}

// turnShellyOn turns on the Shelly heating.
func (hm *HeatingManager) turnShellyOn(shellyHeatingOnURL string) error {
	resp, err := http.Get(shellyHeatingOnURL)
	if err != nil {
		return fmt.Errorf("failed to turn on Shelly: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to turn on Shelly: status code %d", resp.StatusCode)
	}

	fmt.Println("Shelly turned on.")
	return nil
}

// saveLastCheckTime saves the last check time to a file.
func (hm *HeatingManager) saveLastCheckTime() {
	now := time.Now()
	err := os.WriteFile(hm.LastCheckFile, []byte(now.Format(time.RFC3339)), 0644)
	if err != nil {
		log.Printf("Failed to save last check time: %v", err)
	}
}

// nextWeeklyCheckDuration calculates the duration until the next weekly check.
func (hm *HeatingManager) nextWeeklyCheckDuration() time.Duration {
	lastCheck, err := hm.readLastCheckTime()
	if err != nil {
		return 0
	}
	nextCheck := lastCheck.Add(7 * 24 * time.Hour)
	if time.Now().After(nextCheck) {
		return 0
	}
	return nextCheck.Sub(time.Now())
}

// readLastCheckTime reads the last check time from a file.
func (hm *HeatingManager) readLastCheckTime() (time.Time, error) {
	data, err := os.ReadFile(hm.LastCheckFile)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read last check time: %w", err)
	}

	lastCheck, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last check time: %w", err)
	}

	return lastCheck, nil
}
