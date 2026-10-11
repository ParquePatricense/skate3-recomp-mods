//go:build windows

package main

// Windows side: reading the game's memory, Alt+F4, the speedometer overlay
// and the "play" launcher. Plain syscalls, no dependencies.

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")

	pVirtualQueryEx      = kernel32.NewProc("VirtualQueryEx")
	pReadProcessMemory   = kernel32.NewProc("ReadProcessMemory")
	pBeep                = kernel32.NewProc("Beep")
	pFreeConsole         = kernel32.NewProc("FreeConsole")
	pGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	pWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")

	pSetWindowsHookExW         = user32.NewProc("SetWindowsHookExW")
	pCallNextHookEx            = user32.NewProc("CallNextHookEx")
	pGetMessageW               = user32.NewProc("GetMessageW")
	pTranslateMessage          = user32.NewProc("TranslateMessage")
	pDispatchMessageW          = user32.NewProc("DispatchMessageW")
	pGetForegroundWindow       = user32.NewProc("GetForegroundWindow")
	pGetWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
	pPostMessageW              = user32.NewProc("PostMessageW")
	pSetTimer                  = user32.NewProc("SetTimer")
	pRegisterClassExW          = user32.NewProc("RegisterClassExW")
	pCreateWindowExW           = user32.NewProc("CreateWindowExW")
	pDefWindowProcW            = user32.NewProc("DefWindowProcW")
	pSetLayeredWindowAttribute = user32.NewProc("SetLayeredWindowAttributes")
	pSetWindowPos              = user32.NewProc("SetWindowPos")
	pShowWindow                = user32.NewProc("ShowWindow")
	pInvalidateRect            = user32.NewProc("InvalidateRect")
	pBeginPaint                = user32.NewProc("BeginPaint")
	pEndPaint                  = user32.NewProc("EndPaint")
	pFillRect                  = user32.NewProc("FillRect")
	pGetClientRect             = user32.NewProc("GetClientRect")
	pClientToScreen            = user32.NewProc("ClientToScreen")
	pEnumWindows               = user32.NewProc("EnumWindows")
	pIsWindowVisible           = user32.NewProc("IsWindowVisible")
	pMessageBoxW               = user32.NewProc("MessageBoxW")

	pCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	pCreateFontW      = gdi32.NewProc("CreateFontW")
	pSelectObject     = gdi32.NewProc("SelectObject")
	pSetBkMode        = gdi32.NewProc("SetBkMode")
	pSetTextColor     = gdi32.NewProc("SetTextColor")
	pTextOutW         = gdi32.NewProc("TextOutW")
)

const (
	processVMRead      = 0x0010
	processQueryInfo   = 0x0400
	processTerminate   = 0x0001
	synchronize        = 0x00100000
	memCommit          = 0x1000
	memMapped          = 0x40000
	pageNoAccess       = 0x01
	pageGuard          = 0x100
	whKeyboardLL       = 13
	wmSysKeyDown       = 0x0104
	wmClose            = 0x0010
	wmTimer            = 0x0113
	wmPaint            = 0x000F
	wmEraseBkgnd       = 0x0014
	vkF4               = 0x73
	llkhfAltDown       = 0x20
	gameExe            = "skate3.exe"
	overlayW, overlayH = 300, 90
	keyColor           = 0x00FF00FF // magenta (COLORREF 0x00BBGGRR)
)

type memoryBasicInformation struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	PartitionID       uint16
	_                 uint16
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
	_                 uint32
}

// ---------------------------------------------------------------- process

func findGamePID() (uint32, error) {
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer syscall.CloseHandle(snap)
	var pe syscall.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	for err = syscall.Process32First(snap, &pe); err == nil; err = syscall.Process32Next(snap, &pe) {
		if strings.EqualFold(syscall.UTF16ToString(pe.ExeFile[:]), gameExe) {
			return pe.ProcessID, nil
		}
	}
	return 0, errors.New("el juego no esta abierto (no encuentro skate3.exe)")
}

type gameMem struct {
	h    syscall.Handle
	pid  uint32
	base uintptr
}

func query(h syscall.Handle, addr uintptr) (memoryBasicInformation, bool) {
	var mbi memoryBasicInformation
	r, _, _ := pVirtualQueryEx.Call(uintptr(h), addr, uintptr(unsafe.Pointer(&mbi)), unsafe.Sizeof(mbi))
	return mbi, r != 0
}

func openGamePID(pid uint32) (*gameMem, error) {
	h, err := syscall.OpenProcess(processVMRead|processQueryInfo|synchronize, false, pid)
	if err != nil {
		return nil, fmt.Errorf("no pude abrir el juego (%v). Proba ejecutar como administrador", err)
	}
	m := &gameMem{h: h, pid: pid}
	for n := 32; n < 48; n++ {
		b := uintptr(1) << n
		mbi, ok := query(h, b)
		if !ok || mbi.Type != memMapped || mbi.AllocationBase != b {
			continue
		}
		if x, ok := query(h, b+0x82000000); ok && x.State == memCommit {
			m.base = b
			return m, nil
		}
	}
	syscall.CloseHandle(h)
	return nil, errors.New("no encontre la memoria del juego; espera a estar patinando y proba de nuevo")
}

func openGame() (*gameMem, error) {
	pid, err := findGamePID()
	if err != nil {
		return nil, err
	}
	return openGamePID(pid)
}

func (m *gameMem) close() { syscall.CloseHandle(m.h) }

func (m *gameMem) read(addr uint32, buf []byte) error {
	var n uintptr
	r, _, err := pReadProcessMemory.Call(uintptr(m.h), m.base+uintptr(addr),
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), uintptr(unsafe.Pointer(&n)))
	if r == 0 || int(n) != len(buf) {
		return err
	}
	return nil
}

func (m *gameMem) regions() []memRegion {
	const chunk = 16 << 20
	var out []memRegion
	for a := uint64(0x40000000); a < 0x100000000; {
		mbi, ok := query(m.h, m.base+uintptr(a))
		if !ok || mbi.RegionSize == 0 {
			break
		}
		start := uint64(mbi.BaseAddress - m.base)
		end := start + uint64(mbi.RegionSize)
		if end > 0x100000000 {
			end = 0x100000000
		}
		if mbi.State == memCommit && mbi.Protect&(pageNoAccess|pageGuard) == 0 && mbi.Protect != 0 {
			for s := start; s < end; s += chunk {
				e := s + chunk
				if e > end {
					e = end
				}
				out = append(out, memRegion{uint32(s), uint32(e - s)})
			}
		}
		a = end
	}
	return out
}

func consoleUI() calibUI {
	return calibUI{
		say: func(s string) { fmt.Println(s) },
		beep: func(n int) {
			for i := 0; i < n; i++ {
				pBeep.Call(1200, 180)
				time.Sleep(120 * time.Millisecond)
			}
		},
		sleep: time.Sleep,
	}
}

// ---------------------------------------------------------------- play

type rect struct{ L, T, R, B int32 }
type point struct{ X, Y int32 }
type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
	_       uint32
}
type kbdll struct {
	VkCode, ScanCode, Flags, Time uint32
	Extra                         uintptr
}
type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}
type paintStruct struct {
	Hdc       uintptr
	Erase     int32
	RcPaint   rect
	Restore   int32
	IncUpdate int32
	Reserved  [32]byte
}

type helper struct {
	pid       uint32
	proc      syscall.Handle
	gameWnd   uintptr
	overlay   uintptr
	font      uintptr
	keyBrush  uintptr
	mem       *gameMem
	chain     *chain
	text      string
	samples   []speedSample
	shown     bool
	kmh       bool
	closeSent time.Time
}

type speedSample struct {
	t       time.Time
	x, y, z float64
}

var h *helper

func windowPID(w uintptr) uint32 {
	var pid uint32
	pGetWindowThreadProcessId.Call(w, uintptr(unsafe.Pointer(&pid)))
	return pid
}

func (hp *helper) findGameWindow() uintptr {
	var found uintptr
	cb := syscall.NewCallback(func(w, _ uintptr) uintptr {
		if windowPID(w) == hp.pid {
			if v, _, _ := pIsWindowVisible.Call(w); v != 0 {
				found = w
				return 0
			}
		}
		return 1
	})
	pEnumWindows.Call(cb, 0)
	return found
}

func keyboardProc(code int, wParam uintptr, lParam uintptr) uintptr {
	if code == 0 && wParam == wmSysKeyDown {
		k := *(**kbdll)(unsafe.Pointer(&lParam))
		if k.VkCode == vkF4 && k.Flags&llkhfAltDown != 0 {
			fg, _, _ := pGetForegroundWindow.Call()
			if fg != 0 && windowPID(fg) == h.pid {
				pPostMessageW.Call(fg, wmClose, 0, 0)
				h.closeSent = time.Now()
				return 1
			}
		}
	}
	r, _, _ := pCallNextHookEx.Call(0, uintptr(code), wParam, lParam)
	return r
}

func utf16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func overlayProc(w uintptr, m uint32, wp, lp uintptr) uintptr {
	switch m {
	case wmEraseBkgnd:
		return 1
	case wmPaint:
		var ps paintStruct
		dc, _, _ := pBeginPaint.Call(w, uintptr(unsafe.Pointer(&ps)))
		r := rect{0, 0, overlayW, overlayH}
		pFillRect.Call(dc, uintptr(unsafe.Pointer(&r)), h.keyBrush)
		if h.text != "" {
			pSelectObject.Call(dc, h.font)
			pSetBkMode.Call(dc, 1) // TRANSPARENT
			t := syscall.StringToUTF16(h.text)
			n := uintptr(len(t) - 1)
			pSetTextColor.Call(dc, 0x00000000)
			for _, d := range [][2]int{{-3, 0}, {3, 0}, {0, -3}, {0, 3}, {-2, -2}, {2, 2}, {-2, 2}, {2, -2}} {
				pTextOutW.Call(dc, uintptr(10+d[0]), uintptr(10+d[1]), uintptr(unsafe.Pointer(&t[0])), n)
			}
			pSetTextColor.Call(dc, 0x00FFFFFF)
			pTextOutW.Call(dc, 10, 10, uintptr(unsafe.Pointer(&t[0])), n)
		}
		pEndPaint.Call(w, uintptr(unsafe.Pointer(&ps)))
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(w, uintptr(m), wp, lp)
	return r
}

func (hp *helper) createOverlay() {
	inst, _, _ := pGetModuleHandleW.Call(0)
	hp.keyBrush, _, _ = pCreateSolidBrush.Call(keyColor)
	cls := wndClassEx{WndProc: syscall.NewCallback(overlayProc), Instance: inst, ClassName: utf16("Skate3ModsSpeed")}
	cls.Size = uint32(unsafe.Sizeof(cls))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&cls)))
	const exStyle = 0x00080000 | 0x00000020 | 0x00000008 | 0x00000080 | 0x08000000 // layered|transparent|topmost|toolwindow|noactivate
	hp.overlay, _, _ = pCreateWindowExW.Call(exStyle, uintptr(unsafe.Pointer(cls.ClassName)), uintptr(unsafe.Pointer(utf16("speed"))),
		0x80000000, 0, 0, overlayW, overlayH, 0, 0, inst, 0) // WS_POPUP
	pSetLayeredWindowAttribute.Call(hp.overlay, keyColor, 0, 1) // LWA_COLORKEY
	hp.font, _, _ = pCreateFontW.Call(uintptr(0xFFFFFFFF-52+1), 0, 0, 0, 800, 0, 0, 0, 0, 0, 0, 3, 0,
		uintptr(unsafe.Pointer(utf16("Segoe UI"))))
}

func (hp *helper) tickSpeed() {
	if hp.overlay == 0 {
		return
	}
	if hp.gameWnd == 0 {
		hp.gameWnd = hp.findGameWindow()
	}
	fg, _, _ := pGetForegroundWindow.Call()
	visible := hp.gameWnd != 0 && fg == hp.gameWnd
	if visible {
		var cr rect
		var p point
		pGetClientRect.Call(hp.gameWnd, uintptr(unsafe.Pointer(&cr)))
		pClientToScreen.Call(hp.gameWnd, uintptr(unsafe.Pointer(&p)))
		x := p.X + cr.R - overlayW - cr.R/40
		y := p.Y + cr.B/2 - overlayH/2
		pSetWindowPos.Call(hp.overlay, ^uintptr(0) /*HWND_TOPMOST*/, uintptr(x), uintptr(y), 0, 0, 0x0001|0x0010|0x0040) // NOSIZE|NOACTIVATE|SHOWWINDOW
	} else if hp.shown {
		pShowWindow.Call(hp.overlay, 0)
	}
	hp.shown = visible

	text := ""
	if hp.mem == nil {
		if m, err := openGamePID(hp.pid); err == nil {
			hp.mem = m
		}
	}
	if hp.mem != nil && hp.chain != nil {
		if addr, ok := hp.chain.resolve(hp.mem); ok {
			var b [12]byte
			if hp.mem.read(addr, b[:]) == nil {
				if x, y, z, ok := vec(b[:]); ok {
					now := time.Now()
					if n := len(hp.samples); n > 0 {
						l := hp.samples[n-1]
						if math.Hypot(math.Hypot(x-l.x, y-l.y), z-l.z) > 30 {
							hp.samples = hp.samples[:0] // teleport / respawn
						}
					}
					hp.samples = append(hp.samples, speedSample{now, x, y, z})
					for len(hp.samples) > 2 && now.Sub(hp.samples[0].t) > 400*time.Millisecond {
						hp.samples = hp.samples[1:]
					}
					if n := len(hp.samples); n >= 2 {
						a, l := hp.samples[0], hp.samples[n-1]
						dt := l.t.Sub(a.t).Seconds()
						if dt > 0.05 {
							ms := math.Hypot(math.Hypot(l.x-a.x, l.y-a.y), l.z-a.z) / dt
							if hp.kmh {
								text = fmt.Sprintf("%3.0f km/h", ms*3.6)
							} else {
								text = fmt.Sprintf("%3.0f mph", ms*2.23694)
							}
						}
					}
				}
			}
		}
	}
	if text != hp.text {
		hp.text = text
		pInvalidateRect.Call(hp.overlay, 0, 1)
	}
}

func msgBox(text string) {
	pMessageBoxW.Call(0, uintptr(unsafe.Pointer(utf16(text))), uintptr(unsafe.Pointer(utf16("Skate 3 Recomp Mods"))), 0x30)
}

func play(root string) error {
	c := loadConfig(root)
	pid, err := findGamePID()
	if err != nil {
		cmd := exec.Command(filepath.Join(root, gameExe))
		cmd.Dir = root
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("no pude abrir %s: %v", gameExe, err)
		}
		pid = uint32(cmd.Process.Pid)
		cmd.Process.Release()
	}
	wantSpeed := c.on("speedometer") && c["speed_chain"] != ""
	if !c.on("altf4") && !wantSpeed {
		return nil
	}
	pFreeConsole.Call()

	runtime.LockOSThread()
	h = &helper{pid: pid, kmh: c["speed_unit"] != "mph"}
	h.proc, err = syscall.OpenProcess(synchronize|processTerminate, false, pid)
	if err != nil {
		msgBox("No pude vigilar el juego: " + err.Error())
		return nil
	}
	if c.on("altf4") {
		inst, _, _ := pGetModuleHandleW.Call(0)
		if r, _, e := pSetWindowsHookExW.Call(whKeyboardLL, syscall.NewCallback(keyboardProc), inst, 0); r == 0 {
			msgBox("No pude activar Alt+F4: " + e.Error())
		}
	}
	if wantSpeed {
		if ch, err := parseChain(c["speed_chain"]); err == nil {
			h.chain = &ch
			h.createOverlay()
		}
	}
	pSetTimer.Call(0, 0, 50, 0)
	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		if m.Message == wmTimer && m.Hwnd == 0 {
			if w, _, _ := pWaitForSingleObject.Call(uintptr(h.proc), 0); w == 0 {
				os.Exit(0) // game closed
			}
			if !h.closeSent.IsZero() && time.Since(h.closeSent) > 8*time.Second {
				syscall.TerminateProcess(h.proc, 0) // the game ignored Alt+F4
			}
			h.tickSpeed()
			continue
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}
