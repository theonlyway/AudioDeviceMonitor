package notification

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-toast/toast"
	"github.com/theonlyway/AudioDeviceMonitor/assets"
)

var iconPath string

func init() {
	// Write icon once on initialization without resizing
	tempDir := os.TempDir()
	iconPath = filepath.Join(tempDir, "audio-monitor-icon.png")

	// Write the original icon directly - let Windows handle scaling
	if err := os.WriteFile(iconPath, assets.ToastIcon, 0644); err != nil {
		log.Printf("Warning: failed to write icon file: %v\n", err)
		iconPath = ""
	}
}

// ShowToast displays a toast notification for audio device changes
func ShowToast(deviceName string) error {
	notification := toast.Notification{
		AppID:   "Audio Device Monitor",
		Title:   "Audio Device Changed",
		Message: fmt.Sprintf("Active audio device: %s", deviceName),
		Icon:    iconPath,
		Audio:   toast.Silent,
	}

	err := notification.Push()
	if err != nil {
		log.Printf("Toast error: %v\n", err)
	}

	// Small delay to ensure toast is processed
	time.Sleep(100 * time.Millisecond)

	return err
}
