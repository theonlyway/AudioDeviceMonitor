package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Init initializes the logger to write to both console and rotating log file
func Init() error {
	// Get config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(homeDir, ".audio-monitor")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFile := filepath.Join(logDir, "audio-monitor.log")

	// Configure lumberjack for log rotation
	fileLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,    // megabytes
		MaxBackups: 3,     // keep 3 old log files
		MaxAge:     30,    // days
		Compress:   false, // don't compress old logs
	}

	// Write to both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	// Set the default logger to use both outputs
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("Logging initialized - writing to: %s\n", logFile)

	return nil
}
