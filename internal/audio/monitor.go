package audio

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/theonlyway/AudioDeviceMonitor/internal/windows"
)

const (
	eRender             = 0 // Audio rendering stream
	eConsole            = 0 // Games, system notifications, and voice commands
	eAllDevices         = 0 // All devices
	DEVICE_STATE_ACTIVE = 1 // Device is active
	S_FALSE             = 0x00000001 // COM already initialized
)

// Device represents an audio device
type Device struct {
	ID   string
	Name string
}

// Monitor monitors audio device changes
type Monitor struct {
	CurrentDeviceID   string
	CurrentDeviceName string
	mu                sync.Mutex // Protect COM operations
}

// New creates a new audio monitor
func New() *Monitor {
	return &Monitor{}
}

// GetDefaultDevice gets the current default audio device
func (m *Monitor) GetDefaultDevice() (deviceID, deviceName string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	err = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil && err.(*ole.OleError).Code() != ole.S_OK && err.(*ole.OleError).Code() != S_FALSE {
		return "", "", fmt.Errorf("failed to initialize COM: %w", err)
	}
	defer ole.CoUninitialize()

	// Create device enumerator
	unknown, err := ole.CreateInstance(windows.CLSID_MMDeviceEnumerator, windows.IID_IMMDeviceEnumerator)
	if err != nil {
		return "", "", fmt.Errorf("failed to create device enumerator: %w", err)
	}
	defer unknown.Release()

	// Get the IMMDeviceEnumerator interface
	enumerator := (*windows.IMMDeviceEnumerator)(unsafe.Pointer(unknown))

	// Get default audio endpoint (Render, Console role)
	var device *windows.IMMDevice
	ret, _, _ := syscall.SyscallN(
		enumerator.Vtbl.GetDefaultAudioEndpoint,
		uintptr(unsafe.Pointer(enumerator)),
		uintptr(eRender),
		uintptr(eConsole),
		uintptr(unsafe.Pointer(&device)),
	)
	if ret != 0 {
		return "", "", fmt.Errorf("failed to get default audio endpoint: HRESULT 0x%X", ret)
	}
	defer windows.ReleaseDevice(device)

	// Get device ID
	var deviceIDPtr *uint16
	ret, _, _ = syscall.SyscallN(
		device.Vtbl.GetId,
		uintptr(unsafe.Pointer(device)),
		uintptr(unsafe.Pointer(&deviceIDPtr)),
	)
	if ret != 0 {
		return "", "", fmt.Errorf("failed to get device ID: HRESULT 0x%X", ret)
	}
	deviceID = ole.UTF16PtrToString(deviceIDPtr)

	// Open property store
	var propStore *windows.IPropertyStore
	ret, _, _ = syscall.SyscallN(
		device.Vtbl.OpenPropertyStore,
		uintptr(unsafe.Pointer(device)),
		uintptr(windows.STGM_READ),
		uintptr(unsafe.Pointer(&propStore)),
	)
	if ret != 0 {
		return deviceID, "", fmt.Errorf("failed to open property store: HRESULT 0x%X", ret)
	}
	defer windows.ReleasePropertyStore(propStore)

	// Get friendly name - PKEY_Device_FriendlyName = {a45c254e-df1c-4efd-8020-67d146a850e0}, 14
	pkey := windows.PROPERTYKEY{
		Fmtid: *ole.NewGUID("{a45c254e-df1c-4efd-8020-67d146a850e0}"),
		Pid:   14,
	}

	var propVariant windows.PROPVARIANT
	ret, _, _ = syscall.SyscallN(
		propStore.Vtbl.GetValue,
		uintptr(unsafe.Pointer(propStore)),
		uintptr(unsafe.Pointer(&pkey)),
		uintptr(unsafe.Pointer(&propVariant)),
	)
	if ret != 0 {
		return deviceID, "", fmt.Errorf("failed to get device name: HRESULT 0x%X", ret)
	}

	deviceName = ""
	if propVariant.Vt == 0x1f { // VT_LPWSTR
		if propVariant.PwszVal != nil {
			deviceName = ole.UTF16PtrToString(propVariant.PwszVal)
		}
	}

	return deviceID, deviceName, nil
}

// GetAllDevices returns all active audio output devices
func (m *Monitor) GetAllDevices() ([]Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil && err.(*ole.OleError).Code() != ole.S_OK && err.(*ole.OleError).Code() != S_FALSE {
		return nil, fmt.Errorf("failed to initialize COM: %w", err)
	}
	defer ole.CoUninitialize()

	// Create device enumerator
	unknown, err := ole.CreateInstance(windows.CLSID_MMDeviceEnumerator, windows.IID_IMMDeviceEnumerator)
	if err != nil {
		return nil, fmt.Errorf("failed to create device enumerator: %w", err)
	}
	defer unknown.Release()

	enumerator := (*windows.IMMDeviceEnumerator)(unsafe.Pointer(unknown))

	// Enumerate audio endpoints
	var collection *windows.IMMDeviceCollection
	ret, _, _ := syscall.SyscallN(
		enumerator.Vtbl.EnumAudioEndpoints,
		uintptr(unsafe.Pointer(enumerator)),
		uintptr(eRender),
		uintptr(DEVICE_STATE_ACTIVE),
		uintptr(unsafe.Pointer(&collection)),
	)
	if ret != 0 {
		return nil, fmt.Errorf("failed to enumerate devices: HRESULT 0x%X", ret)
	}
	defer windows.ReleaseDeviceCollection(collection)

	// Get count
	var count uint32
	ret, _, _ = syscall.SyscallN(
		collection.Vtbl.GetCount,
		uintptr(unsafe.Pointer(collection)),
		uintptr(unsafe.Pointer(&count)),
	)
	if ret != 0 {
		return nil, fmt.Errorf("failed to get device count: HRESULT 0x%X", ret)
	}

	devices := make([]Device, 0, count)

	// Iterate through devices
	for i := uint32(0); i < count; i++ {
		var device *windows.IMMDevice
		ret, _, _ = syscall.SyscallN(
			collection.Vtbl.Item,
			uintptr(unsafe.Pointer(collection)),
			uintptr(i),
			uintptr(unsafe.Pointer(&device)),
		)
		if ret != 0 {
			continue
		}

		// Get device ID
		var deviceIDPtr *uint16
		ret, _, _ = syscall.SyscallN(
			device.Vtbl.GetId,
			uintptr(unsafe.Pointer(device)),
			uintptr(unsafe.Pointer(&deviceIDPtr)),
		)
		if ret != 0 {
			windows.ReleaseDevice(device)
			continue
		}
		deviceID := ole.UTF16PtrToString(deviceIDPtr)

		// Get friendly name
		var propStore *windows.IPropertyStore
		ret, _, _ = syscall.SyscallN(
			device.Vtbl.OpenPropertyStore,
			uintptr(unsafe.Pointer(device)),
			uintptr(windows.STGM_READ),
			uintptr(unsafe.Pointer(&propStore)),
		)
		if ret != 0 {
			windows.ReleaseDevice(device)
			continue
		}

		pkey := windows.PROPERTYKEY{
			Fmtid: *ole.NewGUID("{a45c254e-df1c-4efd-8020-67d146a850e0}"),
			Pid:   14,
		}

		var propVariant windows.PROPVARIANT
		ret, _, _ = syscall.SyscallN(
			propStore.Vtbl.GetValue,
			uintptr(unsafe.Pointer(propStore)),
			uintptr(unsafe.Pointer(&pkey)),
			uintptr(unsafe.Pointer(&propVariant)),
		)

		deviceName := ""
		if ret == 0 && propVariant.Vt == 0x1f && propVariant.PwszVal != nil {
			deviceName = ole.UTF16PtrToString(propVariant.PwszVal)
		}

		windows.ReleasePropertyStore(propStore)
		windows.ReleaseDevice(device)

		if deviceName != "" {
			devices = append(devices, Device{
				ID:   deviceID,
				Name: deviceName,
			})
		}
	}

	return devices, nil
}

// SetDefaultDevice sets the default audio playback device
func (m *Monitor) SetDefaultDevice(deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil && err.(*ole.OleError).Code() != ole.S_OK && err.(*ole.OleError).Code() != S_FALSE {
		return fmt.Errorf("failed to initialize COM: %w", err)
	}
	defer ole.CoUninitialize()

	// First, verify that the device exists and is active
	deviceExists := false

	// Create device enumerator to check if device is active
	unknown, err := ole.CreateInstance(windows.CLSID_MMDeviceEnumerator, windows.IID_IMMDeviceEnumerator)
	if err != nil {
		return fmt.Errorf("failed to create device enumerator: %w", err)
	}
	defer unknown.Release()

	enumerator := (*windows.IMMDeviceEnumerator)(unsafe.Pointer(unknown))

	// Enumerate active audio endpoints
	var collection *windows.IMMDeviceCollection
	ret, _, _ := syscall.SyscallN(
		enumerator.Vtbl.EnumAudioEndpoints,
		uintptr(unsafe.Pointer(enumerator)),
		uintptr(eRender),
		uintptr(DEVICE_STATE_ACTIVE),
		uintptr(unsafe.Pointer(&collection)),
	)
	if ret != 0 {
		return fmt.Errorf("failed to enumerate devices: HRESULT 0x%X", ret)
	}
	defer windows.ReleaseDeviceCollection(collection)

	// Get count
	var count uint32
	ret, _, _ = syscall.SyscallN(
		collection.Vtbl.GetCount,
		uintptr(unsafe.Pointer(collection)),
		uintptr(unsafe.Pointer(&count)),
	)
	if ret != 0 {
		return fmt.Errorf("failed to get device count: HRESULT 0x%X", ret)
	}

	// Check if the device is in the active devices list
	for i := uint32(0); i < count; i++ {
		var device *windows.IMMDevice
		ret, _, _ = syscall.SyscallN(
			collection.Vtbl.Item,
			uintptr(unsafe.Pointer(collection)),
			uintptr(i),
			uintptr(unsafe.Pointer(&device)),
		)
		if ret != 0 {
			continue
		}

		var deviceIDPtr *uint16
		ret, _, _ = syscall.SyscallN(
			device.Vtbl.GetId,
			uintptr(unsafe.Pointer(device)),
			uintptr(unsafe.Pointer(&deviceIDPtr)),
		)
		windows.ReleaseDevice(device)

		if ret == 0 {
			currentDeviceID := ole.UTF16PtrToString(deviceIDPtr)
			if currentDeviceID == deviceID {
				deviceExists = true
				break
			}
		}
	}

	if !deviceExists {
		return fmt.Errorf("device is not active or not found")
	}

	// Use IPolicyConfig interface (undocumented Windows API)
	// CLSID_CPolicyConfigClient = {870AF99C-171D-4F9E-AF0D-E63DF40C2BC9}
	clsid := ole.NewGUID("{870AF99C-171D-4F9E-AF0D-E63DF40C2BC9}")
	// IID_IPolicyConfig = {F8679F50-850A-41CF-9C72-430F290290C8}
	iid := ole.NewGUID("{F8679F50-850A-41CF-9C72-430F290290C8}")

	unknown2, err := ole.CreateInstance(clsid, iid)
	if err != nil {
		return fmt.Errorf("failed to create policy config: %w", err)
	}
	defer unknown2.Release()

	policyConfig := (*windows.IPolicyConfig)(unsafe.Pointer(unknown2))

	// Convert device ID to UTF16
	deviceIDPtr, err := syscall.UTF16PtrFromString(deviceID)
	if err != nil {
		return fmt.Errorf("failed to convert device ID: %w", err)
	}

	// SetDefaultEndpoint(deviceID, eConsole)
	ret, _, _ = syscall.SyscallN(
		policyConfig.Vtbl.SetDefaultEndpoint,
		uintptr(unsafe.Pointer(policyConfig)),
		uintptr(unsafe.Pointer(deviceIDPtr)),
		uintptr(eConsole),
	)
	if ret != 0 {
		return fmt.Errorf("failed to set default device: HRESULT 0x%X", ret)
	}

	return nil
}
