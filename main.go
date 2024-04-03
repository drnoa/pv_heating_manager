package main

import (
	"log"
)

func main() {
	// Erstelle eine neue Instanz des HeatingManager
	manager, err := NewHeatingManager()
	if err != nil {
		log.Fatalf("Failed to initialize heating manager: %v", err)
	}

	// Starte die kontinuierliche Temperaturüberwachung in einem neuen Goroutine
	go manager.StartTemperatureMonitoring()

	// Starte die wöchentliche Überprüfung in einem neuen Goroutine
	go manager.StartWeeklyCheck()

	// Verhindere, dass das Programm endet, indem in einer endlosen Schleife auf Ereignisse gewartet wird
	select {}
}
