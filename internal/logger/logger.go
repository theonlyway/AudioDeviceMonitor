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

	// Try to write to both stdout and file if console is available
	// If there's no console (Windows GUI app), just write to file
	var output io.Writer
	if hasConsole() {
		output = io.MultiWriter(os.Stdout, fileLogger)
	} else {
		output = fileLogger
	}

	// Set the default logger to use the appropriate output
	log.SetOutput(output)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("Logging initialized - writing to: %s\n", logFile)

	return nil
}

// hasConsole checks if the application has a console attached
func hasConsole() bool {
	// On Windows, check if stdout is a valid handle
	// If running without a console (e.g., built with -H windowsgui), this will fail
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	// Check if stdout is a character device (console) or a pipe
	// If it's neither, there's likely no console
	return (stat.Mode() & os.ModeCharDevice) != 0
}
