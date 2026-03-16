package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	PreferredDeviceID   string `json:"preferred_device_id"`
	PreferredDeviceName string `json:"preferred_device_name"`
	AutoSwitchEnabled   *bool  `json:"auto_switch_enabled,omitempty"` // nil = not set (default true), otherwise use value
}

// IsAutoSwitchEnabled returns true if auto-switch is enabled (default is true for backwards compat)
func (c *Config) IsAutoSwitchEnabled() bool {
	if c.AutoSwitchEnabled == nil {
		// Not set - default to true for backwards compatibility
		return true
	}
	return *c.AutoSwitchEnabled
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".audio-monitor")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return &Config{}, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil // Return empty config if file doesn't exist
		}
		return &Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}, err
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
