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
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
)

//go:embed ui/index.html
var indexHTML string

func runGUI() error {
	if webview2Version() == "" {
		return errors.New("WebView2 runtime not installed")
	}
	guiMode = true
	setLang(config{})
	data := filepath.Join(os.Getenv("LOCALAPPDATA"), "Skate3-Mods", "webview")
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
		case "close":
			w.Terminate()
		}
		return map[bool]string{true: "full", false: "window"}[uiFull]
	})
	gameExitHook = func() {
		w.Dispatch(func() {
			pShowWindow.Call(hwnd, swRestore)
			pSetForegroundWindow.Call(hwnd)
			w.Eval("window.s3back && window.s3back()")
		})
	}
	setFullscreen(hwnd, windowModeSetting() != "window")
	w.SetHtml(indexHTML)
	w.Run()
	return nil
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
