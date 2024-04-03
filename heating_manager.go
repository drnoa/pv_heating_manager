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
	ShellyURL            string  `json:"shellyURL"`
	ShellyHeatingOnURL   string  `json:"shellyHeatingOnURL"`
	TemperatureThreshold float64 `json:"temperatureThreshold"`
}

type HeatingManager struct {
	Config              Config
	TemperatureExceeded bool
	CheckInterval       time.Duration
	LastCheckFile       string
}

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

func (hm *HeatingManager) StartTemperatureMonitoring() {
	ticker := time.NewTicker(hm.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		hm.checkTemperature(hm.Config.ShellyURL)
	}
}

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

func (hm *HeatingManager) checkTemperature(shellyURL string) {
	temperature, err := getTemperature(shellyURL)
	if err != nil {
		log.Printf("Failed to get temperature: %v", err)
		return
	}

	if temperature > hm.Config.TemperatureThreshold {
		fmt.Println("Temperature has exceeded 55°C! Legionellaheating will be resheduled.")
		hm.TemperatureExceeded = true
	} else {
		fmt.Println("Temperature is in order.")
		hm.TemperatureExceeded = false
	}
}

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

func (hm *HeatingManager) weeklyCheck(shellyHeatingOnURL string) {
	if !hm.TemperatureExceeded {
		if err := hm.turnShellyOn(shellyHeatingOnURL); err != nil {
			log.Printf("Failed to turn on Shelly: %v", err)
		}
	}
	hm.TemperatureExceeded = false
	hm.saveLastCheckTime()
}

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

func (hm *HeatingManager) saveLastCheckTime() {
	now := time.Now()
	err := os.WriteFile(hm.LastCheckFile, []byte(now.Format(time.RFC3339)), 0644)
	if err != nil {
		log.Printf("Failed to save last check time: %v", err)
	}
}

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

func (hm *HeatingManager) readLastCheckTime() (time.Time, error) {
	data, err := os.ReadFile(hm.LastCheckFile)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read last check time: %v", err)
	}

	lastCheck, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last check time: %v", err)
	}

	return lastCheck, nil
}
