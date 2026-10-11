//go:build windows

package main

// Windows side: the "play" launcher and Alt+F4. Plain syscalls, no
// dependencies.

import (
	"fmt"
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

	pFreeConsole         = kernel32.NewProc("FreeConsole")
	pGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	pWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")

	pSetWindowsHookExW        = user32.NewProc("SetWindowsHookExW")
	pCallNextHookEx           = user32.NewProc("CallNextHookEx")
	pGetMessageW              = user32.NewProc("GetMessageW")
	pTranslateMessage         = user32.NewProc("TranslateMessage")
	pDispatchMessageW         = user32.NewProc("DispatchMessageW")
	pGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	pGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	pPostMessageW             = user32.NewProc("PostMessageW")
	pSetTimer                 = user32.NewProc("SetTimer")
	pMessageBoxW              = user32.NewProc("MessageBoxW")
)

const (
	processTerminate = 0x0001
	synchronize      = 0x00100000
	whKeyboardLL     = 13
	wmSysKeyDown     = 0x0104
	wmClose          = 0x0010
	wmTimer          = 0x0113
	vkF4             = 0x73
	llkhfAltDown     = 0x20
	gameExe          = "skate3.exe"
)

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

func findGamePID() (uint32, bool) {
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, false
	}
	defer syscall.CloseHandle(snap)
	var pe syscall.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	for err = syscall.Process32First(snap, &pe); err == nil; err = syscall.Process32Next(snap, &pe) {
		if strings.EqualFold(syscall.UTF16ToString(pe.ExeFile[:]), gameExe) {
			return pe.ProcessID, true
		}
	}
	return 0, false
}

func windowPID(w uintptr) uint32 {
	var pid uint32
	pGetWindowThreadProcessId.Call(w, uintptr(unsafe.Pointer(&pid)))
	return pid
}

var (
	gamePID   uint32
	closeSent time.Time
)

// keyboardProc turns Alt+F4 into a normal window close (like the X button),
// which the recomp otherwise swallows.
func keyboardProc(code int, wParam uintptr, lParam uintptr) uintptr {
	if code == 0 && wParam == wmSysKeyDown {
		k := *(**kbdll)(unsafe.Pointer(&lParam))
		if k.VkCode == vkF4 && k.Flags&llkhfAltDown != 0 {
			fg, _, _ := pGetForegroundWindow.Call()
			if fg != 0 && windowPID(fg) == gamePID {
				pPostMessageW.Call(fg, wmClose, 0, 0)
				closeSent = time.Now()
				return 1
			}
		}
	}
	r, _, _ := pCallNextHookEx.Call(0, uintptr(code), wParam, lParam)
	return r
}

func msgBox(text string) {
	t, _ := syscall.UTF16PtrFromString(text)
	c, _ := syscall.UTF16PtrFromString("Skate 3 Recomp Mods")
	pMessageBoxW.Call(0, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), 0x30)
}

func play(root string) error {
	c := loadConfig(root)
	pid, running := findGamePID()
	if !running {
		cmd := exec.Command(filepath.Join(root, gameExe))
		cmd.Dir = root
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("no pude abrir %s: %v", gameExe, err)
		}
		pid = uint32(cmd.Process.Pid)
		cmd.Process.Release()
	}
	if !c.on("altf4") {
		return nil
	}
	pFreeConsole.Call()

	runtime.LockOSThread()
	gamePID = pid
	proc, err := syscall.OpenProcess(synchronize|processTerminate, false, pid)
	if err != nil {
		msgBox("No pude vigilar el juego: " + err.Error())
		return nil
	}
	inst, _, _ := pGetModuleHandleW.Call(0)
	if r, _, e := pSetWindowsHookExW.Call(whKeyboardLL, syscall.NewCallback(keyboardProc), inst, 0); r == 0 {
		msgBox("No pude activar Alt+F4: " + e.Error())
		return nil
	}
	pSetTimer.Call(0, 0, 250, 0)
	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		if m.Message == wmTimer && m.Hwnd == 0 {
			if w, _, _ := pWaitForSingleObject.Call(uintptr(proc), 0); w == 0 {
				os.Exit(0) // game closed: the helper goes away with it
			}
			if !closeSent.IsZero() && time.Since(closeSent) > 8*time.Second {
				syscall.TerminateProcess(proc, 0) // the game ignored the close request
			}
			continue
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}
