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
	w.SetHtml(indexHTML)
	w.Run()
	return nil
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
