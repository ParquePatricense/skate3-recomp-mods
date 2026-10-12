//go:build windows

package main

// Silence during the direct load: the game's own entry in the Windows volume
// mixer (its audio session) is muted while the loading screen is up, so the
// menu music the boot automation skips through isn't heard. It's Windows'
// per-app mute, not the recomp's audio setting.

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	ole32              = syscall.NewLazyDLL("ole32.dll")
	pCoInitializeEx    = ole32.NewProc("CoInitializeEx")
	pCoCreateInstance  = ole32.NewProc("CoCreateInstance")
	clsidMMDeviceEnum  = windows.GUID{Data1: 0xBCDE0395, Data2: 0xE52F, Data3: 0x467C, Data4: [8]byte{0x8E, 0x3D, 0xC4, 0x57, 0x92, 0x91, 0x69, 0x2E}}
	iidMMDeviceEnum    = windows.GUID{Data1: 0xA95664D2, Data2: 0x9614, Data3: 0x4F35, Data4: [8]byte{0xA7, 0x46, 0xDE, 0x8D, 0xB6, 0x36, 0x17, 0xE6}}
	iidSessionManager2 = windows.GUID{Data1: 0x77AA99A0, Data2: 0x1BD6, Data3: 0x484F, Data4: [8]byte{0x8B, 0xC7, 0x2C, 0x65, 0x4C, 0x9A, 0x9B, 0x6F}}
	iidSessionControl2 = windows.GUID{Data1: 0xBFB7FF88, Data2: 0x7239, Data3: 0x4FC9, Data4: [8]byte{0x8F, 0xA2, 0x07, 0xC9, 0x50, 0xBE, 0x9C, 0x6D}}
	iidSimpleAudioVol  = windows.GUID{Data1: 0x87CE5498, Data2: 0x68D6, Data3: 0x44E5, Data4: [8]byte{0x92, 0x15, 0x6D, 0xA4, 0x7E, 0xF8, 0x83, 0xD8}}
	comReady           bool
)

// comCall calls method i of a COM object.
func comCall(obj uintptr, i int, args ...uintptr) uintptr {
	vtbl := ***(***[32]uintptr)(unsafe.Pointer(&obj)) // obj -> vtable pointer -> vtable
	r, _, _ := syscall.SyscallN(vtbl[i], append([]uintptr{obj}, args...)...)
	return r
}

func comRelease(obj uintptr) {
	if obj != 0 {
		comCall(obj, 2)
	}
}

// muteGame mutes or unmutes every audio session of the game process on the
// default output. It reports whether it found one (the game opens its audio
// a few seconds after starting, so the caller retries).
func muteGame(pid uint32, mute bool) bool {
	if pid == 0 {
		return false
	}
	if !comReady {
		pCoInitializeEx.Call(0, 2) // COINIT_APARTMENTTHREADED
		comReady = true
	}
	var enum, dev, mgr, sessions uintptr
	if r, _, _ := pCoCreateInstance.Call(uintptr(unsafe.Pointer(&clsidMMDeviceEnum)), 0, 1|4|16,
		uintptr(unsafe.Pointer(&iidMMDeviceEnum)), uintptr(unsafe.Pointer(&enum))); r != 0 || enum == 0 {
		return false
	}
	defer comRelease(enum)
	if comCall(enum, 4, 0, 1, uintptr(unsafe.Pointer(&dev))) != 0 || dev == 0 { // GetDefaultAudioEndpoint(eRender, eMultimedia)
		return false
	}
	defer comRelease(dev)
	if comCall(dev, 3, uintptr(unsafe.Pointer(&iidSessionManager2)), 23, 0, uintptr(unsafe.Pointer(&mgr))) != 0 || mgr == 0 { // Activate(CLSCTX_ALL)
		return false
	}
	defer comRelease(mgr)
	if comCall(mgr, 5, uintptr(unsafe.Pointer(&sessions))) != 0 || sessions == 0 { // GetSessionEnumerator
		return false
	}
	defer comRelease(sessions)
	var n int32
	comCall(sessions, 3, uintptr(unsafe.Pointer(&n))) // GetCount
	found := false
	for i := int32(0); i < n; i++ {
		var ctl, ctl2, vol uintptr
		if comCall(sessions, 4, uintptr(i), uintptr(unsafe.Pointer(&ctl))) != 0 || ctl == 0 { // GetSession
			continue
		}
		if comCall(ctl, 0, uintptr(unsafe.Pointer(&iidSessionControl2)), uintptr(unsafe.Pointer(&ctl2))) == 0 && ctl2 != 0 {
			var spid uint32
			comCall(ctl2, 14, uintptr(unsafe.Pointer(&spid))) // GetProcessId
			if spid == pid && comCall(ctl2, 0, uintptr(unsafe.Pointer(&iidSimpleAudioVol)), uintptr(unsafe.Pointer(&vol))) == 0 && vol != 0 {
				m := uintptr(0)
				if mute {
					m = 1
				}
				comCall(vol, 5, m, 0) // SetMute
				comRelease(vol)
				found = true
			}
			comRelease(ctl2)
		}
		comRelease(ctl)
	}
	return found
}
