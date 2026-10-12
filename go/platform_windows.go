//go:build windows

package main

// Windows side: the "play" launcher and in-game helper (Alt+F4, quick
// restart, priority, keep awake, cursor hiding, Windows key, window memory,
// do not disturb).

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")

	pFreeConsole             = kernel32.NewProc("FreeConsole")
	pGetModuleHandleW        = kernel32.NewProc("GetModuleHandleW")
	pWaitForSingleObject     = kernel32.NewProc("WaitForSingleObject")
	pSetPriorityClass        = kernel32.NewProc("SetPriorityClass")
	pSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")

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
	pGetAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	pEnumWindows              = user32.NewProc("EnumWindows")
	pIsWindowVisible          = user32.NewProc("IsWindowVisible")
	pIsIconic                 = user32.NewProc("IsIconic")
	pGetWindow                = user32.NewProc("GetWindow")
	pGetWindowRect            = user32.NewProc("GetWindowRect")
	pGetClientRect            = user32.NewProc("GetClientRect")
	pGetWindowLongW           = user32.NewProc("GetWindowLongW")
	pGetDpiForWindow          = user32.NewProc("GetDpiForWindow")
	pMonitorFromWindow        = user32.NewProc("MonitorFromWindow")
	pGetMonitorInfoW          = user32.NewProc("GetMonitorInfoW")
	pEnumDisplayMonitors      = user32.NewProc("EnumDisplayMonitors")
	pGetCursorPos             = user32.NewProc("GetCursorPos")
	pCreateCursor             = user32.NewProc("CreateCursor")
	pSetSystemCursor          = user32.NewProc("SetSystemCursor")
	pSystemParametersInfoW    = user32.NewProc("SystemParametersInfoW")
)

const (
	processTerminate  = 0x0001
	processSetInfo    = 0x0200
	synchronize       = 0x00100000
	whKeyboardLL      = 13
	wmKeyDown         = 0x0100
	wmKeyUp           = 0x0101
	wmSysKeyDown      = 0x0104
	wmSysKeyUp        = 0x0105
	wmClose           = 0x0010
	wmTimer           = 0x0113
	vkF4              = 0x73
	vkR               = 0x52
	vkControl         = 0x11
	vkLWin            = 0x5B
	vkRWin            = 0x5C
	llkhfAltDown      = 0x20
	gameExe           = "skate3.exe"
	highPriorityClass = 0x80
	esContinuous      = 0x80000000
	esSystemRequired  = 0x1
	esDisplayRequired = 0x2
	ocrNormal         = 32512
	spiSetCursors     = 0x57
	wsCaption         = 0x00C00000
	gwlStyle          = ^uintptr(15) // -16
	gwOwner           = 4
)

type point struct{ X, Y int32 }
type rect struct{ L, T, R, B int32 }
type monitorInfo struct {
	Size          uint32
	Monitor, Work rect
	Flags         uint32
}
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

// ---------------------------------------------------------------- helper state

var (
	hc          config // settings for this run
	gamePID     uint32
	closeSent   time.Time
	wantRestart bool
)

func gameInFront() (uintptr, bool) {
	fg, _, _ := pGetForegroundWindow.Call()
	return fg, fg != 0 && windowPID(fg) == gamePID
}

// keyboardProc: Alt+F4 becomes a normal window close (the recomp swallows
// it), Ctrl+Alt+R asks for a restart and the Windows keys are ignored while
// the game is in front.
func keyboardProc(code int, wParam uintptr, lParam uintptr) uintptr {
	if code == 0 {
		k := *(**kbdll)(unsafe.Pointer(&lParam))
		if fg, front := gameInFront(); front {
			switch {
			case hc.feature("winkey") && (k.VkCode == vkLWin || k.VkCode == vkRWin):
				return 1
			case hc.feature("altf4") && wParam == wmSysKeyDown && k.VkCode == vkF4 && k.Flags&llkhfAltDown != 0:
				pPostMessageW.Call(fg, wmClose, 0, 0)
				closeSent = time.Now()
				return 1
			case hc.feature("restart") && (wParam == wmSysKeyDown || wParam == wmKeyDown) && k.VkCode == vkR &&
				k.Flags&llkhfAltDown != 0 && keyDown(vkControl):
				wantRestart = true
				return 1
			}
		}
	}
	r, _, _ := pCallNextHookEx.Call(0, uintptr(code), wParam, lParam)
	return r
}

func keyDown(vk uintptr) bool {
	r, _, _ := pGetAsyncKeyState.Call(vk)
	return r&0x8000 != 0
}

func msgBox(text string) {
	t, _ := syscall.UTF16PtrFromString(text)
	c, _ := syscall.UTF16PtrFromString("Skate 3 Recomp Mods")
	pMessageBoxW.Call(0, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), 0x30)
}

// gameWindow returns the game's main (visible, unowned) window. The
// callback is made once: Windows callbacks can't be freed.
var (
	foundWindow   uintptr
	enumWindowsCB = syscall.NewCallback(func(h, _ uintptr) uintptr {
		if windowPID(h) == gamePID {
			if v, _, _ := pIsWindowVisible.Call(h); v != 0 {
				if o, _, _ := pGetWindow.Call(h, gwOwner); o == 0 {
					foundWindow = h
					return 0
				}
			}
		}
		return 1
	})
)

func gameWindow() uintptr {
	foundWindow = 0
	pEnumWindows.Call(enumWindowsCB, 0)
	return foundWindow
}

var (
	monitorList    []uintptr
	enumMonitorsCB = syscall.NewCallback(func(h, _, _, _ uintptr) uintptr {
		monitorList = append(monitorList, h)
		return 1
	})
)

// monitorIndex gives the recomp's "monitor" number (1 = first) of a monitor.
func monitorIndex(m uintptr) int {
	monitorList = monitorList[:0]
	pEnumDisplayMonitors.Call(0, 0, enumMonitorsCB, 0)
	for i, h := range monitorList {
		if h == m {
			return i + 1
		}
	}
	return 0
}

// monitors lists the screens in the recomp's order.
func monitors() []screen {
	monitorList = monitorList[:0]
	pEnumDisplayMonitors.Call(0, 0, enumMonitorsCB, 0)
	out := []screen{}
	for i, h := range monitorList {
		mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
		pGetMonitorInfoW.Call(h, uintptr(unsafe.Pointer(&mi)))
		out = append(out, screen{N: i + 1, W: int(mi.Monitor.R - mi.Monitor.L), H: int(mi.Monitor.B - mi.Monitor.T), Primary: mi.Flags&1 != 0})
	}
	return out
}

// windowState is what "remember the window" saves when the game closes.
type windowState struct {
	ok         bool
	fullscreen bool
	w, h       int
	monitor    int
}

func readWindowState(w uintptr) windowState {
	if w == 0 {
		return windowState{}
	}
	if ic, _, _ := pIsIconic.Call(w); ic != 0 {
		return windowState{}
	}
	var wr, cr rect
	pGetWindowRect.Call(w, uintptr(unsafe.Pointer(&wr)))
	pGetClientRect.Call(w, uintptr(unsafe.Pointer(&cr)))
	mon, _, _ := pMonitorFromWindow.Call(w, 2)
	mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	pGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	style, _, _ := pGetWindowLongW.Call(w, gwlStyle)
	st := windowState{ok: true, monitor: monitorIndex(mon)}
	st.fullscreen = style&wsCaption == 0 && wr == mi.Monitor
	dpi, _, _ := pGetDpiForWindow.Call(w)
	if dpi == 0 {
		dpi = 96
	}
	st.w = int(cr.R-cr.L) * 96 / int(dpi)
	st.h = int(cr.B-cr.T) * 96 / int(dpi)
	return st
}

func saveWindowState(root string, st windowState) {
	if !st.ok || st.w < 320 || st.h < 200 {
		return
	}
	vals := [][2]string{{"fullscreen", strconv.FormatBool(st.fullscreen)}}
	if !st.fullscreen {
		vals = append(vals, [2]string{"window_width", strconv.Itoa(st.w)}, [2]string{"window_height", strconv.Itoa(st.h)})
	}
	if st.monitor > 0 {
		vals = append(vals, [2]string{"monitor", strconv.Itoa(st.monitor)})
	}
	setTomlValues(root, vals)
}

// Cursor hiding swaps the system arrow for a blank one and puts the user's
// cursors back (SPI_SETCURSORS) as soon as the mouse moves or the game loses
// focus, and always when the helper ends.
var (
	cursorHidden bool
	lastMouse    point
	lastMove     time.Time
)

func restoreCursor() {
	pSystemParametersInfoW.Call(spiSetCursors, 0, 0, 0)
	cursorHidden = false
}

func hideCursor() {
	and := make([]byte, 128)
	for i := range and {
		and[i] = 0xFF
	}
	xor := make([]byte, 128)
	inst, _, _ := pGetModuleHandleW.Call(0)
	cur, _, _ := pCreateCursor.Call(inst, 0, 0, 32, 32, uintptr(unsafe.Pointer(&and[0])), uintptr(unsafe.Pointer(&xor[0])))
	if cur != 0 {
		if r, _, _ := pSetSystemCursor.Call(cur, ocrNormal); r != 0 {
			cursorHidden = true
		}
	}
}

func cursorTick() {
	var p point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	if p != lastMouse {
		lastMouse, lastMove = p, time.Now()
		if cursorHidden {
			restoreCursor()
		}
		return
	}
	fg, front := gameInFront()
	if !front {
		if cursorHidden {
			restoreCursor()
		}
		return
	}
	if cursorHidden || time.Since(lastMove) < 3*time.Second {
		return
	}
	var r rect
	pGetWindowRect.Call(fg, uintptr(unsafe.Pointer(&r)))
	if p.X >= r.L && p.X < r.R && p.Y >= r.T && p.Y < r.B {
		hideCursor()
	}
}

func launchGame(root string) (uint32, error) {
	cmd := exec.Command(filepath.Join(root, gameExe))
	cmd.Dir = root
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf(T("no pude abrir %s: %v", "could not start %s: %v"), gameExe, err)
	}
	pid := uint32(cmd.Process.Pid)
	cmd.Process.Release()
	return pid, nil
}

func watchGame(pid uint32) (syscall.Handle, error) {
	proc, err := syscall.OpenProcess(synchronize|processTerminate|processSetInfo, false, pid)
	if err != nil {
		proc, err = syscall.OpenProcess(synchronize|processTerminate, false, pid)
	}
	if err == nil && hc.feature("priority") {
		pSetPriorityClass.Call(uintptr(proc), highPriorityClass)
	}
	return proc, err
}

func play(root string) error {
	hc = loadConfig(root)
	setLang(hc)
	checkMods(root) // put the mods back if a recomp update replaced the files
	pid, running := findGamePID()
	if !running {
		var err error
		if pid, err = launchGame(root); err != nil {
			return err
		}
	}
	if !hc.anyHelper() {
		return nil
	}
	pFreeConsole.Call()

	runtime.LockOSThread()
	gamePID = pid
	proc, err := watchGame(pid)
	if err != nil {
		msgBox(T("No pude vigilar el juego: ", "Could not watch the game: ") + err.Error())
		return nil
	}
	if hc.feature("cursor") {
		restoreCursor() // in case a previous run was killed while hiding it
	}
	dnd := hc.feature("dnd") && dndOn(root)
	if hc.feature("awake") {
		pSetThreadExecutionState.Call(esContinuous | esSystemRequired | esDisplayRequired)
	}
	if hc.feature("altf4") || hc.feature("restart") || hc.feature("winkey") {
		inst, _, _ := pGetModuleHandleW.Call(0)
		if r, _, e := pSetWindowsHookExW.Call(whKeyboardLL, syscall.NewCallback(keyboardProc), inst, 0); r == 0 {
			msgBox(T("No pude activar los atajos de teclado: ", "Could not enable the keyboard shortcuts: ") + e.Error())
		}
	}
	finish := func() {
		if cursorHidden {
			restoreCursor()
		}
		if dnd {
			dndOff(root)
		}
		pSetThreadExecutionState.Call(esContinuous)
	}
	var lastWin windowState
	tick := 0
	pSetTimer.Call(0, 0, 250, 0)
	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		if m.Message == wmTimer && m.Hwnd == 0 {
			tick++
			if hc.feature("window") && tick%4 == 0 {
				if st := readWindowState(gameWindow()); st.ok {
					lastWin = st
				}
			}
			if hc.feature("cursor") {
				cursorTick()
			}
			if wantRestart {
				wantRestart = false
				syscall.TerminateProcess(proc, 0)
				pWaitForSingleObject.Call(uintptr(proc), 10000)
				syscall.CloseHandle(proc)
				if hc.feature("window") {
					saveWindowState(root, lastWin)
				}
				if gamePID, err = launchGame(root); err != nil {
					finish()
					msgBox(err.Error())
					os.Exit(1)
				}
				if proc, err = watchGame(gamePID); err != nil {
					finish()
					os.Exit(0)
				}
				closeSent = time.Time{}
				continue
			}
			if w, _, _ := pWaitForSingleObject.Call(uintptr(proc), 0); w == 0 {
				if hc.feature("window") {
					saveWindowState(root, lastWin)
				}
				finish()
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
	finish()
	return nil
}

// Do not disturb: Windows' own "notifications" switch (ToastEnabled), turned
// off while playing and put back afterwards. The previous value is kept in
// the ini so a crash can't leave it off.
const toastKey = `Software\Microsoft\Windows\CurrentVersion\PushNotifications`

func dndOn(root string) bool {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, toastKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	c := loadConfig(root)
	if _, saved := c["dnd_prev"]; !saved {
		prev, _, err := k.GetIntegerValue("ToastEnabled")
		if err != nil {
			prev = 1
		}
		c["dnd_prev"] = strconv.FormatUint(prev, 10)
		c.save(root)
	}
	return k.SetDWordValue("ToastEnabled", 0) == nil
}

func dndOff(root string) {
	c := loadConfig(root)
	prev, ok := c["dnd_prev"]
	if !ok {
		return
	}
	if k, _, err := registry.CreateKey(registry.CURRENT_USER, toastKey, registry.SET_VALUE); err == nil {
		v, _ := strconv.ParseUint(prev, 10, 32)
		k.SetDWordValue("ToastEnabled", uint32(v))
		k.Close()
	}
	delete(c, "dnd_prev")
	c.save(root)
}

var pGetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")

func systemSpanish() bool {
	l, _, _ := pGetUserDefaultUILanguage.Call()
	return l&0x3FF == 0x0A // LANG_SPANISH
}

const shortcutName = "Skate 3.lnk"

func psQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

func powershell(script string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func desktopShortcut() string {
	return "(Join-Path ([Environment]::GetFolderPath('Desktop')) " + psQuote(shortcutName) + ")"
}

// createShortcut puts a "Skate 3" shortcut on the desktop that starts the
// game through the mods helper (minimized, so no console flashes).
func createShortcut(root string) error {
	exe := filepath.Join(root, modsExe)
	if !exists(exe) {
		var err error
		if exe, err = os.Executable(); err != nil {
			return err
		}
	}
	return powershell("$s=(New-Object -ComObject WScript.Shell).CreateShortcut(" + desktopShortcut() + ");" +
		"$s.TargetPath=" + psQuote(exe) + ";$s.Arguments='play';" +
		"$s.WorkingDirectory=" + psQuote(root) + ";" +
		"$s.IconLocation=" + psQuote(filepath.Join(root, gameExe)+",0") + ";" +
		"$s.WindowStyle=7;$s.Description='Skate 3 Recomp (mods)';$s.Save()")
}

func removeShortcut() error {
	return powershell("$p=" + desktopShortcut() + ";if(Test-Path -LiteralPath $p){Remove-Item -LiteralPath $p}")
}

func documentsDir() string {
	if p, err := windows.KnownFolderPath(windows.FOLDERID_Documents, 0); err == nil {
		return p
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, "Documents")
}

func desktopDir() string {
	if p, err := windows.KnownFolderPath(windows.FOLDERID_Desktop, 0); err == nil {
		return p
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, "Desktop")
}

func openPath(p string) error {
	return exec.Command("explorer.exe", p).Start()
}

func systemInfo() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	name, _, _ := k.GetStringValue("ProductName")
	build, _, _ := k.GetStringValue("CurrentBuild")
	return name + " build " + build + " - WebView2 " + webview2Version()
}

// webview2Version is the installed WebView2 runtime ("" if missing).
func webview2Version() string {
	const id = `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`
	for _, loc := range []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\` + id},
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\EdgeUpdate\Clients\` + id},
		{registry.CURRENT_USER, `Software\Microsoft\EdgeUpdate\Clients\` + id},
	} {
		if k, err := registry.OpenKey(loc.root, loc.path, registry.QUERY_VALUE); err == nil {
			v, _, _ := k.GetStringValue("pv")
			k.Close()
			if v != "" && v != "0.0.0.0" {
				return v
			}
		}
	}
	return ""
}

func goos() string { return "windows" }

// startPlay starts the game through the helper as its own process, so the
// modder window can close while you play.
func startPlay(root string) error {
	exe := filepath.Join(root, modsExe)
	if !exists(exe) {
		var err error
		if exe, err = os.Executable(); err != nil {
			return err
		}
	}
	cmd := exec.Command(exe, "play")
	cmd.Dir = root
	return cmd.Start()
}

// pickFolder shows Windows' folder picker.
func pickFolder(title string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-STA", "-NonInteractive", "-Command",
		"Add-Type -AssemblyName System.Windows.Forms;"+
			"$d=New-Object System.Windows.Forms.FolderBrowserDialog;$d.Description="+psQuote(title)+";"+
			"$d.ShowNewFolderButton=$false;"+
			"if($d.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK){[Console]::Out.Write($d.SelectedPath)}")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
