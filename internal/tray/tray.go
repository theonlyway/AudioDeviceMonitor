package tray

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"time"

	"github.com/getlantern/systray"
	"github.com/tc-hib/winres"
	"github.com/theonlyway/AudioDeviceMonitor/assets"
	"github.com/theonlyway/AudioDeviceMonitor/internal/audio"
	"github.com/theonlyway/AudioDeviceMonitor/internal/config"
	"github.com/theonlyway/AudioDeviceMonitor/internal/notification"
)

// Run starts the system tray application
func Run() {
	systray.Run(onReady, onExit)
}

// convertPNGToICO converts a PNG byte slice to ICO format for system tray
func convertPNGToICO(pngData []byte) ([]byte, error) {
	// Decode PNG
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	// Create ICO
	ico, err := winres.NewIconFromImages([]image.Image{img})
	if err != nil {
		return nil, fmt.Errorf("failed to create icon: %w", err)
	}

	// Encode to ICO format
	var buf bytes.Buffer
	if err := ico.SaveICO(&buf); err != nil {
		return nil, fmt.Errorf("failed to encode ICO: %w", err)
	}

	return buf.Bytes(), nil
}

func onReady() {
	// Convert PNG to ICO for Windows systray
	if iconData, err := convertPNGToICO(assets.SystrayIcon); err == nil {
		systray.SetIcon(iconData)
	} else {
		log.Printf("Warning: failed to convert icon: %v\n", err)
	}

	systray.SetTitle("Audio Monitor")
	systray.SetTooltip("Audio Device Monitor - Running")

	// Pre-create device menu items (max 10 devices)
	maxDevices := 10

	mDevices := systray.AddMenuItem("Devices", "View all audio devices")

	// Pre-create submenu item slots
	deviceMenuItems := make([]*systray.MenuItem, maxDevices)
	clickHandlers := make([]chan struct{}, maxDevices)
	for i := 0; i < maxDevices; i++ {
		item := mDevices.AddSubMenuItem("", "")
		item.Hide()
		deviceMenuItems[i] = item
		clickHandlers[i] = make(chan struct{})
	}

	systray.AddSeparator()

	// Preferred device section
	mPreferred := systray.AddMenuItem("Set Preferred Default", "Set device to auto-select on startup")
	preferredDeviceMenuItems := make([]*systray.MenuItem, maxDevices)
	for i := 0; i < maxDevices; i++ {
		item := mPreferred.AddSubMenuItem("", "")
		item.Hide()
		preferredDeviceMenuItems[i] = item
	}

	systray.AddSeparator()

	// Auto-switch toggle
	mAutoSwitch := systray.AddMenuItem("Auto-Switch on Startup", "Enable/disable automatic switching to preferred device")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Start audio monitoring in background
	monitor := audio.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading config: %v\n", err)
		cfg = &config.Config{}
	}

	// Log auto-switch status at startup
	log.Printf("Auto-switch on startup: %v\n", cfg.IsAutoSwitchEnabled())

	// Set auto-switch checkbox state
	if cfg.IsAutoSwitchEnabled() {
		mAutoSwitch.Check()
	} else {
		mAutoSwitch.Uncheck()
	}

	// Function to update device list
	var updateDeviceList func()
	updateDeviceList = func() {
		devices, err := monitor.GetAllDevices()
		if err != nil {
			log.Printf("Error getting devices: %v\n", err)
			return
		}

		// Hide all submenu items first
		for i := 0; i < maxDevices; i++ {
			deviceMenuItems[i].Hide()
			preferredDeviceMenuItems[i].Hide()
		}

		// Update visible items
		for i, device := range devices {
			if i >= maxDevices {
				break
			}

			var title string
			if device.ID == monitor.CurrentDeviceID {
				title = "● " + device.Name // Filled dot for active device
			} else {
				title = "   " + device.Name // Indent for inactive
			}
			deviceMenuItems[i].SetTitle(title)
			deviceMenuItems[i].SetTooltip(device.Name)
			deviceMenuItems[i].Show()
			deviceMenuItems[i].Enable()

			// Update preferred device menu
			var prefTitle string
			if device.ID == cfg.PreferredDeviceID {
				prefTitle = "✓ " + device.Name // Checkmark for preferred
			} else {
				prefTitle = "   " + device.Name
			}
			preferredDeviceMenuItems[i].SetTitle(prefTitle)
			preferredDeviceMenuItems[i].SetTooltip(device.Name)
			preferredDeviceMenuItems[i].Show()
			preferredDeviceMenuItems[i].Enable()

			// Start click handler for device switching
			deviceID := device.ID
			deviceName := device.Name
			go func(idx int, id, name string) {
				for range deviceMenuItems[idx].ClickedCh {
					log.Printf("Switching to device: %s\n", name)
					if err := monitor.SetDefaultDevice(id); err != nil {
						log.Printf("Error setting default device: %v\n", err)
					} else {
						log.Printf("Successfully switched to: %s\n", name)
					}
				}
			}(i, deviceID, deviceName)

			// Start click handler for setting preferred device
			go func(idx int, id, name string) {
				for range preferredDeviceMenuItems[idx].ClickedCh {
					log.Printf("Setting preferred device: %s\n", name)
					cfg.PreferredDeviceID = id
					cfg.PreferredDeviceName = name
					if err := cfg.Save(); err != nil {
						log.Printf("Error saving config: %v\n", err)
					} else {
						log.Printf("Saved preferred device: %s\n", name)
						updateDeviceList() // Refresh to show checkmark
					}
				}
			}(i, deviceID, deviceName)
		}
	}

	go func() {
		log.Println("Audio Device Monitor started")

		// Get initial device and populate list
		deviceID, deviceName, err := monitor.GetDefaultDevice()
		if err != nil {
			log.Printf("Error getting initial device: %v\n", err)
		} else {
			monitor.CurrentDeviceID = deviceID
			monitor.CurrentDeviceName = deviceName
			log.Printf("Initial audio device: %s (ID: %s)\n", deviceName, deviceID)
			systray.SetTooltip(fmt.Sprintf("Audio Monitor - %s", deviceName))
			updateDeviceList()
		}

		// Apply preferred device if configured and auto-switch is enabled
		if cfg.IsAutoSwitchEnabled() && cfg.PreferredDeviceID != "" {
			log.Printf("Checking preferred device: %s\n", cfg.PreferredDeviceName)
			if cfg.PreferredDeviceID != monitor.CurrentDeviceID {
				// Check if the preferred device is in the active devices list
				devices, err := monitor.GetAllDevices()
				if err != nil {
					log.Printf("Error getting device list: %v\n", err)
				} else {
					deviceFound := false
					for _, device := range devices {
						if device.ID == cfg.PreferredDeviceID {
							deviceFound = true
							break
						}
					}

					if deviceFound {
						log.Printf("Applying preferred device: %s\n", cfg.PreferredDeviceName)
						if err := monitor.SetDefaultDevice(cfg.PreferredDeviceID); err != nil {
							log.Printf("Error applying preferred device: %v\n", err)
						} else {
							log.Printf("Successfully applied preferred device: %s\n", cfg.PreferredDeviceName)
						}
					} else {
						log.Printf("Preferred device is not currently active: %s\n", cfg.PreferredDeviceName)
					}
				}
			}
		} else if cfg.PreferredDeviceID != "" {
			log.Printf("Auto-switch is disabled, skipping preferred device: %s\n", cfg.PreferredDeviceName)
		}

		// Poll for changes every second
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			deviceID, deviceName, err := monitor.GetDefaultDevice()
			if err != nil {
				log.Printf("Error checking device: %v\n", err)
				continue
			}

			// Check if device changed
			if deviceID != monitor.CurrentDeviceID {
				log.Printf("Audio device changed from '%s' to '%s'\n", monitor.CurrentDeviceName, deviceName)

				monitor.CurrentDeviceID = deviceID
				monitor.CurrentDeviceName = deviceName
				systray.SetTooltip(fmt.Sprintf("Audio Monitor - %s", deviceName))
				updateDeviceList()

				// Show toast notification
				if err := notification.ShowToast(deviceName); err != nil {
					log.Printf("Error showing toast: %v\n", err)
				}
			}
		}
	}()

	// Handle auto-switch toggle
	go func() {
		for range mAutoSwitch.ClickedCh {
			// Toggle the value
			newValue := !cfg.IsAutoSwitchEnabled()
			cfg.AutoSwitchEnabled = &newValue
			if newValue {
				mAutoSwitch.Check()
				log.Println("Auto-switch enabled")
			} else {
				mAutoSwitch.Uncheck()
				log.Println("Auto-switch disabled")
			}
			if err := cfg.Save(); err != nil {
				log.Printf("Error saving config: %v\n", err)
			}
		}
	}()

	// Handle menu clicks
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	log.Println("Audio Device Monitor stopped")
}
