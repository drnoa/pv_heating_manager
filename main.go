package main

import (
	"log"
)

// HeatingManager is the main entry point of the program.
// It initializes a new HeatingManager instance and
// starts two goroutines for temperature monitoring and weekly check.
// The program then enters an infinite loop, waiting for events.
func main() {
	// Initialize a new HeatingManager instance
	manager, err := NewHeatingManager()
	if err != nil {
		log.Fatalf("Failed to initialize heating manager: %v", err)
	}

	// Start temperature monitoring in a separate goroutine
	go manager.StartTemperatureMonitoring()

	// Start weekly check in a separate goroutine
	go manager.StartWeeklyCheck()

	// Wait for events in an infinite loop
	select {}
}
