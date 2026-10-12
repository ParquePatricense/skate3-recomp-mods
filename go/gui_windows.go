//go:build windows

package main

// The modder window: a WebView2 page (ui/index.html) talking to apiDo.
// WebView2 comes with Windows 10/11; without it we fall back to the
// console menu.

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"time"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"github.com/jchv/go-webview2/pkg/edge"
)

//go:embed ui/index.html
var indexHTML string

// The French, Italian, German, Chinese and Japanese texts of the page.
//
//go:embed ui/i18n.json
var uiI18nJSON string

func runGUI() error {
	if webview2Version() == "" {
		return errors.New("WebView2 runtime not installed")
	}
	guiMode = true
	setLang(config{})
	// Its own taskbar identity: without it Windows groups the modder with any
	// pinned shortcut to Skate3-Mods.exe ("Skate 3 (directo)" and the like)
	// and shows that shortcut's Skate 3 icon.
	if id, err := syscall.UTF16PtrFromString("ParquePatricense.Skate3Mods.Modeador"); err == nil {
		syscall.NewLazyDLL("shell32.dll").NewProc("SetCurrentProcessExplicitAppUserModelID").Call(uintptr(unsafe.Pointer(id)))
	}
	data := filepath.Join(os.Getenv("LOCALAPPDATA"), "Skate3-Mods", "webview")
	// the radio starts by itself when the modder opens
	os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "--autoplay-policy=no-user-gesture-required")
	width, height := windowSize(1200, 780)
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		DataPath:  data,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title: "Skate 3 Recomp Mods", Width: width, Height: height, Center: true, IconId: 1,
		},
	})
	if w == nil {
		guiMode = false
		return errors.New("could not open the WebView2 window")
	}
	defer w.Destroy()
	root := ""
	w.Bind("s3state", func() map[string]any {
		apiMu.Lock()
		defer apiMu.Unlock()
		return apiState(root)
	})
	w.Bind("s3do", func(id int, cmd, arg string) {
		go func() {
			msg, err := apiDo(&root, cmd, arg)
			res := map[string]any{"msg": msg}
			if err != nil {
				res["err"] = err.Error()
			}
			b, _ := json.Marshal(res)
			w.Dispatch(func() { w.Eval(fmt.Sprintf("window.s3done(%d,%s)", id, b)) })
		}()
	})
	hwnd := uintptr(w.Window())
	setWindowIcons(hwnd)
	w.Bind("s3win", func(mode string) string {
		switch mode {
		case "full", "window":
			setFullscreen(hwnd, mode == "full")
			saveWindowMode(mode)
		case "toggle":
			setFullscreen(hwnd, !uiFull)
			saveWindowMode(map[bool]string{true: "full", false: "window"}[uiFull])
		case "min":
			pShowWindow.Call(hwnd, swMinimize)
		default:
			if n, ok := strings.CutPrefix(mode, "mon:"); ok {
				saveUIMonitor(n)
				full := uiFull
				setFullscreen(hwnd, false)
				moveToMonitor(hwnd, uiMonitorSetting())
				setFullscreen(hwnd, full)
			}
		case "close":
			w.Terminate()
		}
		return map[bool]string{true: "full", false: "window"}[uiFull]
	})
	ctrl := controller(w)
	// s3zoom fits the page to the window: the same share of the screen at any
	// resolution and DPI, times the user's own size (1 = automatic).
	w.Bind("s3zoom", func(mult float64) float64 {
		if mult < .5 || mult > 2 {
			mult = 1
		}
		var cr rect
		pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&cr)))
		dpi, _, _ := pGetDpiForWindow.Call(hwnd)
		if dpi == 0 {
			dpi = 96
		}
		cssW, cssH := float64(cr.R-cr.L)*96/float64(dpi), float64(cr.B-cr.T)*96/float64(dpi)
		z := math.Min(cssW/2050, cssH/1150)
		z = math.Max(.5, math.Min(z, 2.5)) * mult
		if ctrl != nil && cssW > 0 {
			w.Dispatch(func() { putZoom(ctrl, z) })
		}
		return z
	})
	gameExitHook = func() {
		w.Dispatch(func() {
			pShowWindow.Call(hwnd, swRestore)
			pSetForegroundWindow.Call(hwnd)
			w.Eval("window.s3back && window.s3back()")
		})
	}
	moveToMonitor(hwnd, uiMonitorSetting())
	setFullscreen(hwnd, windowModeSetting() != "window")
	w.SetHtml(strings.Replace(indexHTML, "/*I18N*/{}", uiI18nJSON, 1))
	w.Run()
	return nil
}

// setWindowIcons gives the window the sticker icon at the sizes the
// taskbar and Alt+Tab ask for.
func setWindowIcons(h uintptr) {
	inst, _, _ := pGetModuleHandleW.Call(0)
	load := user32.NewProc("LoadImageW")
	send := user32.NewProc("SendMessageW")
	metrics := user32.NewProc("GetSystemMetrics")
	for _, k := range []struct{ which, cx, cy uintptr }{{1, 11, 12}, {0, 49, 50}} { // ICON_BIG: SM_CXICON, ICON_SMALL: SM_CXSMICON
		cx, _, _ := metrics.Call(k.cx)
		cy, _, _ := metrics.Call(k.cy)
		if ic, _, _ := load.Call(inst, 1, 1, cx, cy, 0); ic != 0 {
			send.Call(h, 0x0080, k.which, ic) // WM_SETICON
		}
	}
}

// controller digs the WebView2 controller out of the webview (the package
// keeps it private) to set the zoom and turn off the browser's own zoom keys,
// so Ctrl+wheel goes through s3zoom and is remembered.
func controller(w webview2.WebView) *edge.ICoreWebView2Controller {
	v := reflect.ValueOf(w)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil
	}
	f := v.Elem().FieldByName("browser")
	if !f.IsValid() {
		return nil
	}
	ch, ok := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Interface().(*edge.Chromium)
	if !ok || ch == nil {
		return nil
	}
	if s, err := ch.GetSettings(); err == nil && s != nil {
		s.PutIsZoomControlEnabled(false)
	}
	return ch.GetController()
}

// putZoom calls ICoreWebView2Controller::put_ZoomFactor (vtable slot 8).
func putZoom(c *edge.ICoreWebView2Controller, z float64) {
	vtbl := **(**[9]uintptr)(unsafe.Pointer(c))
	syscall.SyscallN(vtbl[8], uintptr(unsafe.Pointer(c)), uintptr(math.Float64bits(z)))
}

// The modder opens fullscreen (borderless, on its monitor) unless the user
// switched it to a window; the choice is kept next to the remembered folder.
const (
	wsOverlappedWindow = 0x00CF0000
	swMinimize         = 6
	swRestore          = 9
	swpNoSize          = 0x0001
	swpNoMove          = 0x0002
	swpNoZOrder        = 0x0004
	swpFrameChanged    = 0x0020
	swpNoOwnerZOrder   = 0x0200
)

type windowPlacement struct {
	Length, Flags, ShowCmd uint32
	Min, Max               point
	Normal                 rect
}

var (
	pSetWindowLongW      = user32.NewProc("SetWindowLongW")
	pSetWindowPos        = user32.NewProc("SetWindowPos")
	pGetWindowPlacement  = user32.NewProc("GetWindowPlacement")
	pSetWindowPlacement  = user32.NewProc("SetWindowPlacement")
	pShowWindow          = user32.NewProc("ShowWindow")
	pSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	uiFull               bool
	uiPlacement          windowPlacement
)

func setFullscreen(h uintptr, on bool) {
	if h == 0 || on == uiFull {
		return
	}
	style, _, _ := pGetWindowLongW.Call(h, gwlStyle)
	if on {
		uiPlacement.Length = uint32(unsafe.Sizeof(uiPlacement))
		pGetWindowPlacement.Call(h, uintptr(unsafe.Pointer(&uiPlacement)))
		pSetWindowLongW.Call(h, gwlStyle, style&^wsOverlappedWindow)
		mon, _, _ := pMonitorFromWindow.Call(h, 2)
		mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
		pGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
		r := mi.Monitor
		pSetWindowPos.Call(h, 0, uintptr(r.L), uintptr(r.T), uintptr(r.R-r.L), uintptr(r.B-r.T), swpNoOwnerZOrder|swpFrameChanged)
	} else {
		pSetWindowLongW.Call(h, gwlStyle, style|wsOverlappedWindow)
		pSetWindowPlacement.Call(h, uintptr(unsafe.Pointer(&uiPlacement)))
		pSetWindowPos.Call(h, 0, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoZOrder|swpNoOwnerZOrder|swpFrameChanged)
	}
	uiFull = on
}

// The screen the modder opens on (0 = wherever Windows puts it), numbered
// like the game's "monitor" setting.
func uiMonitorFile() string {
	return filepath.Join(os.Getenv("APPDATA"), "Skate3-Mods", "monitor.txt")
}

func uiMonitorSetting() int {
	b, _ := os.ReadFile(uiMonitorFile())
	n := 0
	fmt.Sscan(strings.TrimSpace(string(b)), &n)
	return n
}

func saveUIMonitor(n string) {
	os.MkdirAll(filepath.Dir(uiMonitorFile()), 0o755)
	os.WriteFile(uiMonitorFile(), []byte(n), 0o644)
}

// moveToMonitor centers the (windowed) modder on screen n.
func moveToMonitor(h uintptr, n int) {
	if n <= 0 {
		return
	}
	monitorList = monitorList[:0]
	pEnumDisplayMonitors.Call(0, 0, enumMonitorsCB, 0)
	if n > len(monitorList) {
		return
	}
	mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	pGetMonitorInfoW.Call(monitorList[n-1], uintptr(unsafe.Pointer(&mi)))
	var wr rect
	pGetWindowRect.Call(h, uintptr(unsafe.Pointer(&wr)))
	ww, wh := wr.R-wr.L, wr.B-wr.T
	r := mi.Work
	if ww > r.R-r.L {
		ww = r.R - r.L
	}
	if wh > r.B-r.T {
		wh = r.B - r.T
	}
	x, y := r.L+(r.R-r.L-ww)/2, r.T+(r.B-r.T-wh)/2
	pSetWindowPos.Call(h, 0, uintptr(x), uintptr(y), uintptr(ww), uintptr(wh), swpNoZOrder|swpNoOwnerZOrder)
	uiPlacement = windowPlacement{}
}

func windowModeFile() string {
	return filepath.Join(os.Getenv("APPDATA"), "Skate3-Mods", "ventana.txt")
}

func windowModeSetting() string {
	b, _ := os.ReadFile(windowModeFile())
	return strings.TrimSpace(string(b))
}

func saveWindowMode(mode string) {
	os.MkdirAll(filepath.Dir(windowModeFile()), 0o755)
	os.WriteFile(windowModeFile(), []byte(mode), 0o644)
}

// gameExitHook brings the modder back when the game closes.
var gameExitHook func()

// waitGameExit waits for the game to start and then to close.
func waitGameExit() {
	seen := false
	for start := time.Now(); ; time.Sleep(2 * time.Second) {
		_, running := findGamePID()
		if running {
			seen = true
		} else if seen || time.Since(start) > 30*time.Second {
			break
		}
	}
	if gameExitHook != nil {
		gameExitHook()
	}
}

// windowSize scales the window to the screen's DPI (the exe is DPI aware)
// and keeps it inside the screen.
func windowSize(w, h uint) (uint, uint) {
	dpi, _, _ := user32.NewProc("GetDpiForSystem").Call()
	if dpi == 0 {
		dpi = 96
	}
	w, h = w*uint(dpi)/96, h*uint(dpi)/96
	metrics := user32.NewProc("GetSystemMetrics")
	sw, _, _ := metrics.Call(0) // SM_CXSCREEN
	sh, _, _ := metrics.Call(1) // SM_CYSCREEN
	if max := uint(sw) * 94 / 100; sw > 0 && w > max {
		w = max
	}
	if max := uint(sh) * 90 / 100; sh > 0 && h > max {
		h = max
	}
	return w, h
}

// consoleSetup gives the GUI-subsystem exe a console for the text modes.
func consoleSetup(parent bool) {
	const attachParent = ^uintptr(0)
	attach := kernel32.NewProc("AttachConsole")
	alloc := kernel32.NewProc("AllocConsole")
	if r, _, _ := attach.Call(attachParent); r == 0 || !parent {
		if r != 0 {
			pFreeConsole.Call()
		}
		alloc.Call()
	}
	if f, err := os.OpenFile("CONOUT$", os.O_RDWR, 0); err == nil {
		os.Stdout, os.Stderr = f, f
	}
	if f, err := os.OpenFile("CONIN$", os.O_RDWR, 0); err == nil {
		os.Stdin = f
		stdin = newStdin()
	}
}
