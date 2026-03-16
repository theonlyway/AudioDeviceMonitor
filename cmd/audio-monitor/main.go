package main

import (
	"log"

	"github.com/theonlyway/AudioDeviceMonitor/internal/logger"
	"github.com/theonlyway/AudioDeviceMonitor/internal/tray"
)

var Version = "dev"

func main() {
	// Initialize logging
	if err := logger.Init(); err != nil {
		log.Printf("Warning: failed to initialize file logging: %v\n", err)
	}

	log.Printf("Audio Device Monitor version: %s\n", Version)

	tray.Run()
}
