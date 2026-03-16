package main

import (
	"log"

	"github.com/theonlyway/AudioDeviceMonitor/internal/logger"
	"github.com/theonlyway/AudioDeviceMonitor/internal/tray"
)

func main() {
	// Initialize logging
	if err := logger.Init(); err != nil {
		log.Printf("Warning: failed to initialize file logging: %v\n", err)
	}

	tray.Run()
}
