//go:build windows

package main

// Arrival sign: when the direct load ends, the name of the place you're at
// shows for a few seconds in a corner of the game's screen, like a skate
// video's spot title.

import (
	"syscall"
	"time"
	"unsafe"
)

const arrivalTimer = 9

var (
	arrivalHwnd       uintptr
	arrivalStart      time.Time
	arrivalName       string
	arrivalArea       string
	arrivalClassOK    bool
	arrivalProc       = syscall.NewCallback(arrivalWndProc)
	arrivalShowFor    = 4 * time.Second
	arrivalFadeFrames = 8
)

func showArrival(root string) {
	if loadConfig(root)["ui_arrival"] == "0" || arrivalHwnd != 0 {
		return
	}
	arrivalName, arrivalArea = "Port Carverton", T("De vuelta en tu partida", "Back in your save")
	inst, _, _ := pGetModuleHandleW.Call(0)
	if !arrivalClassOK {
		wc := wndClassEx{WndProc: arrivalProc, Instance: inst, ClassName: utf16("Skate3ModsArrival")}
		wc.Size = uint32(unsafe.Sizeof(wc))
		pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		arrivalClassOK = true
	}
	m := coverRect(root)
	if gw := gameWindow(); gw != 0 {
		mon, _, _ := pMonitorFromWindow.Call(gw, 2)
		mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
		pGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
		m = mi.Monitor
	}
	ht := m.B - m.T
	w, h := (m.R-m.L)*2/5, ht/7
	x, y := m.L+ht/20, m.B-h-ht/12
	hw, _, _ := pCreateWindowExW.Call(wsExTopmost|wsExToolWindow|wsExLayered|wsExTransparent|wsExNoActivate,
		uintptr(unsafe.Pointer(utf16("Skate3ModsArrival"))), uintptr(unsafe.Pointer(utf16("Skate 3"))), wsPopup,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h), 0, 0, inst, 0)
	if hw == 0 {
		return
	}
	arrivalHwnd, arrivalStart = hw, time.Now()
	pSetLayeredWindowAttributes.Call(hw, 0, 0, lwaAlpha)
	pShowWindow.Call(hw, swShowNoActivate)
	pSetTimer.Call(hw, arrivalTimer, 50, 0)
}

func arrivalWndProc(h, m, wp, lp uintptr) uintptr {
	switch m {
	case wmNcHitTest:
		return htTransparent
	case wmEraseBkgnd:
		return 1
	case wmTimer:
		if wp == arrivalTimer {
			el := time.Since(arrivalStart)
			a := 235
			fade := time.Duration(arrivalFadeFrames) * 50 * time.Millisecond
			switch {
			case el < fade:
				a = int(235 * el / fade)
			case el > arrivalShowFor+fade:
				pKillTimer.Call(h, arrivalTimer)
				pDestroyWindow.Call(h)
				arrivalHwnd = 0
				return 0
			case el > arrivalShowFor:
				a = int(235 * (arrivalShowFor + fade - el) / fade)
			}
			pSetLayeredWindowAttributes.Call(h, 0, uintptr(a), lwaAlpha)
			return 0
		}
	case wmPaint:
		paintArrival(h)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(h, m, wp, lp)
	return r
}

// paintArrival: an orange sticker with the spot name and a dark strip with
// the district under it.
func paintArrival(h uintptr) {
	var ps paintStruct
	dc, _, _ := pBeginPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	defer pEndPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	var cr rect
	pGetClientRect.Call(h, uintptr(unsafe.Pointer(&cr)))
	w, ht := cr.R, cr.B
	fill := func(r rect, color uintptr) {
		b, _, _ := pCreateSolidBrush.Call(color)
		pFillRect.Call(dc, uintptr(unsafe.Pointer(&r)), b)
		pDeleteObject.Call(b)
	}
	draw := func(s string, r rect, size int32, color uintptr) {
		f, _, _ := pCreateFontW.Call(uintptr(-size), 0, 0, 0, 600, 0, 0, 0, 0, 0, 0, 4, 0, uintptr(unsafe.Pointer(utf16("Bahnschrift SemiBold Condensed"))))
		old, _, _ := pSelectObject.Call(dc, f)
		pSetTextColor.Call(dc, color)
		pSetBkMode.Call(dc, 1)
		u, _ := syscall.UTF16FromString(s)
		pDrawTextW.Call(dc, uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1), uintptr(unsafe.Pointer(&r)), dtVCenter|dtSingleLine)
		pSelectObject.Call(dc, old)
		pDeleteObject.Call(f)
	}
	top := ht * 62 / 100
	fill(rect{0, 0, w, ht}, rgb(10, 10, 10))
	fill(rect{0, 0, w, top}, rgb(255, 90, 31))
	draw(arrivalName, rect{ht / 6, 0, w, top}, top*7/10, rgb(17, 17, 17))
	draw(arrivalArea, rect{ht / 6, top, w, ht}, (ht-top)*6/10, rgb(198, 255, 61))
}
