# Audio Device Toast Service

A Windows system tray application that monitors audio devices and provides quick switching capabilities with automatic preferred device selection.

## Features

### System Tray Integration
- Runs quietly in the Windows system tray
- Custom icon with real-time device status
- Clean, organized menu interface

### Audio Device Management
- **Real-time Monitoring** - Automatically detects when the active audio device changes
- **Quick Device Switching** - Switch between audio devices with a single click from the system tray
- **Device List View** - See all available audio output devices at a glance with visual indicators for the active device
- **Toast Notifications** - Receive Windows notifications when the audio device changes

### Preferred Device Selection
- **Auto-Select on Startup** - Set a preferred default device that will be automatically selected when the application starts
- **Persistent Configuration** - Your preferred device choice is saved to disk and persists across restarts
- **Easy Configuration** - Set your preferred device through the system tray menu

### Technical Features
- Lightweight background monitoring with minimal CPU usage
- Thread-safe COM operations with mutex synchronization
- Proper error handling for COM interface initialization
- Embedded application icons (no external dependencies)

## Requirements

- Windows OS
- Go 1.25.3 or higher

## Installation

### From Release (Recommended)

1. Download the latest `audio-monitor-windows-amd64.zip` from the [Releases](https://github.com/theonlyway/AudioDeviceMonitor/releases) page
2. Extract the zip file to a location of your choice (e.g., `C:\Program Files\AudioMonitor\`)
3. Run `audio-monitor.exe`

**Note**: No administrator privileges required.

### Auto-Start on Login

To run the application automatically when you log in to Windows, use this PowerShell command:

```powershell
# Run this in PowerShell (no admin required)
$exePath = "C:\Path\To\audio-monitor.exe"  # Update this path
$action = New-ScheduledTaskAction -Execute $exePath
$trigger = New-ScheduledTaskTrigger -AtLogOn -User $env:USERNAME
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit 0
Register-ScheduledTask -TaskName "AudioDeviceMonitor" -Action $action -Trigger $trigger -Settings $settings -Description "Monitors audio device changes"
```

To remove the auto-start task:
```powershell
Unregister-ScheduledTask -TaskName "AudioDeviceMonitor" -Confirm:$false
```

### From Source

1. Clone the repository:
```bash
git clone https://github.com/theonlyway/AudioDeviceMonitor.git
cd AudioDeviceToastService
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o bin/audio-monitor.exe ./cmd/audio-monitor
```

4. (Optional) Build without console window:
```bash
go build -ldflags -H=windowsgui -o bin/audio-monitor.exe ./cmd/audio-monitor
```

## Usage

Run the application:
```bash
.\audio-monitor.exe
```

The application will minimize to the system tray. Look for the icon in your system tray (near the clock).

### System Tray Menu

The application adds an icon to your system tray with the following options:

- **Devices** (submenu)
  - Lists all available audio output devices
  - Active device marked with ●
  - Click any device to switch to it immediately

- **Set Preferred Default** (submenu)
  - Lists all available audio output devices
  - Current preferred device marked with ✓
  - Click any device to set it as your preferred default
  - Automatically applied on next startup

- **Quit** - Exit the application

### Configuration

The application stores your preferred device configuration in:
```
%USERPROFILE%\.audio-monitor\config.json
```

This file is automatically created when you set a preferred device and includes:
- Preferred device ID
- Preferred device name

## Project Structure

```
AudioDeviceToastService/
├── assets/              # Static assets (icons)
│   ├── assets.go        # Embedded assets
│   ├── systray_icon.png # System tray icon (recommended: 16x16 or 32x32 pixels)
│   └── toast_icon.png   # Toast notification icon (recommended: 48x48 to 256x256 pixels)
├── bin/                 # Build output (gitignored)
│   └── audio-monitor.exe
├── cmd/
│   └── audio-monitor/   # Application entrypoint
│       └── main.go
├── internal/
│   ├── audio/           # Audio device monitoring and control
│   │   └── monitor.go
│   ├── config/          # Configuration management
│   │   └── config.go
│   ├── notification/    # Toast notifications
│   │   └── toast.go
│   ├── tray/            # System tray interface
│   │   └── tray.go
│   └── windows/         # Windows COM interfaces
│       └── com.go
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Icon Guidelines

### System Tray Icon (`systray_icon.png`)
- **Recommended size**: 16x16 or 32x32 pixels
- **Format**: PNG with transparency
- **Notes**: The application converts PNG to ICO format at runtime
- **High DPI**: 32x32 provides better quality on high-DPI displays

### Toast Notification Icon (`toast_icon.png`)
- **Recommended size**: 48x48 to 256x256 pixels
- **Format**: PNG with transparency
- **Notes**: Windows automatically scales the icon; larger source images provide better quality
- **Best practice**: Use 256x256 for optimal display across all scenarios

## Dependencies

- `github.com/go-ole/go-ole` - Windows COM/OLE automation
- `github.com/go-toast/toast` - Windows toast notifications
- `github.com/getlantern/systray` - System tray integration
- `github.com/tc-hib/winres` - ICO conversion for system tray icons

## How It Works

The application uses the Windows Multimedia Device (MMDevice) API through COM to:
1. Enumerate audio devices
2. Get the default audio endpoint
3. Retrieve device properties (ID and friendly name)
4. Poll for changes at regular intervals
5. Display toast notifications when changes are detected

### Technical Details

The application uses official Microsoft-defined GUIDs from the Windows SDK to access COM interfaces:
- **CLSID_MMDeviceEnumerator** - COM class ID to create the device enumerator
- **IID_IMMDeviceEnumerator** - Interface for enumerating audio devices
- **IID_IMMDevice** - Interface for individual audio device operations
- **IID_IPropertyStore** - Interface for accessing device properties
- **PKEY_Device_FriendlyName** - Property key for retrieving the device's display name

These GUIDs are defined in Windows SDK headers (`mmdeviceapi.h`, `functiondiscoverykeys_devpkey.h`) and are consistent across all Windows systems.
