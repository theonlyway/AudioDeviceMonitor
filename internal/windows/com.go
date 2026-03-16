package windows

import (
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

const (
	CLSCTX_ALL = 23
	STGM_READ  = 0
)

// GUID definitions
var (
	CLSID_MMDeviceEnumerator = ole.NewGUID("{BCDE0395-E52F-467C-8E3D-C4579291692E}")
	IID_IMMDeviceEnumerator  = ole.NewGUID("{A95664D2-9614-4F35-A746-DE8DB63617E6}")
	IID_IMMDevice            = ole.NewGUID("{D666063F-1587-4E43-81F1-B948E807363F}")
	IID_IPropertyStore       = ole.NewGUID("{886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99}")
)

// COM Interface definitions
type IMMDeviceEnumeratorVtbl struct {
	QueryInterface                         uintptr
	AddRef                                 uintptr
	Release                                uintptr
	EnumAudioEndpoints                     uintptr
	GetDefaultAudioEndpoint                uintptr
	GetDevice                              uintptr
	RegisterEndpointNotificationCallback   uintptr
	UnregisterEndpointNotificationCallback uintptr
}

type IMMDeviceEnumerator struct {
	Vtbl *IMMDeviceEnumeratorVtbl
}

type IMMDeviceVtbl struct {
	QueryInterface    uintptr
	AddRef            uintptr
	Release           uintptr
	Activate          uintptr
	OpenPropertyStore uintptr
	GetId             uintptr
	GetState          uintptr
}

type IMMDevice struct {
	Vtbl *IMMDeviceVtbl
}

type IMMDeviceCollectionVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	GetCount       uintptr
	Item           uintptr
}

type IMMDeviceCollection struct {
	Vtbl *IMMDeviceCollectionVtbl
}

type IPropertyStoreVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	GetCount       uintptr
	GetAt          uintptr
	GetValue       uintptr
	SetValue       uintptr
	Commit         uintptr
}

type IPropertyStore struct {
	Vtbl *IPropertyStoreVtbl
}

type IPolicyConfigVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	GetMixFormat        uintptr
	GetDeviceFormat     uintptr
	ResetDeviceFormat   uintptr
	SetDeviceFormat     uintptr
	GetProcessingPeriod uintptr
	SetProcessingPeriod uintptr
	GetShareMode        uintptr
	SetShareMode        uintptr
	GetPropertyValue    uintptr
	SetPropertyValue    uintptr
	SetDefaultEndpoint  uintptr
	SetEndpointVisibility uintptr
}

type IPolicyConfig struct {
	Vtbl *IPolicyConfigVtbl
}

type PROPERTYKEY struct {
	Fmtid ole.GUID
	Pid   uint32
}

type PROPVARIANT struct {
	Vt         uint16
	WReserved1 uint16
	WReserved2 uint16
	WReserved3 uint16
	PwszVal    *uint16 // For VT_LPWSTR
	_          uintptr // Union padding
}

// ReleaseDevice releases a COM device interface
func ReleaseDevice(device *IMMDevice) {
	if device != nil {
		syscall.SyscallN(device.Vtbl.Release, uintptr(unsafe.Pointer(device)))
	}
}

// ReleaseDeviceCollection releases a COM device collection interface
func ReleaseDeviceCollection(collection *IMMDeviceCollection) {
	if collection != nil {
		syscall.SyscallN(collection.Vtbl.Release, uintptr(unsafe.Pointer(collection)))
	}
}

// ReleasePropertyStore releases a COM property store interface
func ReleasePropertyStore(propStore *IPropertyStore) {
	if propStore != nil {
		syscall.SyscallN(propStore.Vtbl.Release, uintptr(unsafe.Pointer(propStore)))
	}
}
