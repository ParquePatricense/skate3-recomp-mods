//go:build windows

package main

// Direct load: while the recomp's boot automation clicks through the intro,
// the language screen, Press Start and the boot dialogs, a black loading
// screen covers the game's monitor. It goes away when the game log says
// gameplay was reached (or after two minutes, whatever happens), so all you
// see is "loading" and then your skater.

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	gdi32 = syscall.NewLazyDLL("gdi32.dll")

	pRegisterClassExW           = user32.NewProc("RegisterClassExW")
	pCreateWindowExW            = user32.NewProc("CreateWindowExW")
	pDefWindowProcW             = user32.NewProc("DefWindowProcW")
	pDestroyWindow              = user32.NewProc("DestroyWindow")
	pBeginPaint                 = user32.NewProc("BeginPaint")
	pEndPaint                   = user32.NewProc("EndPaint")
	pFillRect                   = user32.NewProc("FillRect")
	pDrawTextW                  = user32.NewProc("DrawTextW")
	pInvalidateRect             = user32.NewProc("InvalidateRect")
	pSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	pKillTimer                  = user32.NewProc("KillTimer")
	pLoadCursorW                = user32.NewProc("LoadCursorW")

	pCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	pCreateFontW            = gdi32.NewProc("CreateFontW")
	pSelectObject           = gdi32.NewProc("SelectObject")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pSetTextColor           = gdi32.NewProc("SetTextColor")
	pSetBkMode              = gdi32.NewProc("SetBkMode")
	pCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pBitBlt                 = gdi32.NewProc("BitBlt")
	pDeleteDC               = gdi32.NewProc("DeleteDC")
)

const (
	wsPopup          = 0x80000000
	wsExTopmost      = 0x00000008
	wsExToolWindow   = 0x00000080
	wsExLayered      = 0x00080000
	wsExTransparent  = 0x00000020
	wsExNoActivate   = 0x08000000
	wmPaint          = 0x000F
	wmNcHitTest      = 0x0084
	wmEraseBkgnd     = 0x0014
	htTransparent    = ^uintptr(0) // -1
	lwaAlpha         = 0x2
	swShowNoActivate = 4
	swpNoActivate    = 0x0010
	hwndTopmost      = ^uintptr(0) // -1
	dtCenter         = 0x1
	dtVCenter        = 0x4
	dtSingleLine     = 0x20
	srcCopy          = 0x00CC0020
	coverTimer       = 7
)

type wndClassEx struct {
	Size, Style                   uint32
	WndProc                       uintptr
	ClsExtra, WndExtra            int32
	Instance, Icon, Cursor, Bkgnd uintptr
	MenuName, ClassName           *uint16
	IconSm                        uintptr
}

type paintStruct struct {
	Hdc         uintptr
	Erase       int32
	Paint       rect
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}

var (
	coverHwnd    uintptr
	coverStart   time.Time
	coverFade    int // 0 = showing, 1..n = fading out
	coverClassOK bool
	coverText    string
	coverProc    = syscall.NewCallback(coverWndProc)
)

func rgb(r, g, b uint32) uintptr { return uintptr(r | g<<8 | b<<16) }

func utf16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

// showCover puts the loading screen over the monitor the game opens on.
func showCover(root string) {
	if coverHwnd != 0 {
		return
	}
	inst, _, _ := pGetModuleHandleW.Call(0)
	if !coverClassOK {
		cur, _, _ := pLoadCursorW.Call(0, 32512) // IDC_ARROW
		wc := wndClassEx{WndProc: coverProc, Instance: inst, Cursor: cur, ClassName: utf16("Skate3ModsCover")}
		wc.Size = uint32(unsafe.Sizeof(wc))
		pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		coverClassOK = true
	}
	r := coverRect(root)
	coverText = T("CARGANDO TU PARTIDA", "LOADING YOUR SAVE")
	h, _, _ := pCreateWindowExW.Call(wsExTopmost|wsExToolWindow|wsExLayered|wsExTransparent|wsExNoActivate,
		uintptr(unsafe.Pointer(utf16("Skate3ModsCover"))), uintptr(unsafe.Pointer(utf16("Skate 3"))), wsPopup,
		uintptr(r.L), uintptr(r.T), uintptr(r.R-r.L), uintptr(r.B-r.T), 0, 0, inst, 0)
	if h == 0 {
		return
	}
	coverHwnd, coverStart, coverFade = h, time.Now(), 0
	pSetLayeredWindowAttributes.Call(h, 0, 255, lwaAlpha)
	pShowWindow.Call(h, swShowNoActivate)
	pSetTimer.Call(h, coverTimer, 60, 0)
}

// coverRect is the game's monitor: the one in skate3.toml, else the main one.
func coverRect(root string) rect {
	n := 0
	if v := readToml(root)["monitor"]; v != "" {
		for _, c := range v {
			if c < '0' || c > '9' {
				n = 0
				break
			}
			n = n*10 + int(c-'0')
		}
	}
	monitorList = monitorList[:0]
	pEnumDisplayMonitors.Call(0, 0, enumMonitorsCB, 0)
	var fallback rect
	for i, h := range monitorList {
		mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
		pGetMonitorInfoW.Call(h, uintptr(unsafe.Pointer(&mi)))
		if i+1 == n {
			return mi.Monitor
		}
		if mi.Flags&1 != 0 {
			fallback = mi.Monitor
		}
	}
	if fallback.R == 0 {
		fallback = rect{0, 0, 1920, 1080}
	}
	return fallback
}

// hideCover fades the loading screen out.
func hideCover() {
	if coverHwnd != 0 && coverFade == 0 {
		coverFade = 1
	}
}

// coverTick keeps the cover above the game and follows the game window to
// its monitor once it exists.
func coverTick() {
	if coverHwnd == 0 {
		return
	}
	if time.Since(coverStart) > 2*time.Minute {
		hideCover()
	}
	if gw := gameWindow(); gw != 0 && coverFade == 0 {
		mon, _, _ := pMonitorFromWindow.Call(gw, 2)
		mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
		pGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
		r := mi.Monitor
		pSetWindowPos.Call(coverHwnd, hwndTopmost, uintptr(r.L), uintptr(r.T), uintptr(r.R-r.L), uintptr(r.B-r.T), swpNoActivate)
	}
}

func coverWndProc(h, m, wp, lp uintptr) uintptr {
	switch m {
	case wmNcHitTest:
		return htTransparent
	case wmEraseBkgnd:
		return 1
	case wmTimer:
		if wp == coverTimer {
			if coverFade > 0 {
				coverFade++
				a := 255 - coverFade*32
				if a <= 0 {
					pKillTimer.Call(h, coverTimer)
					pDestroyWindow.Call(h)
					coverHwnd = 0
					return 0
				}
				pSetLayeredWindowAttributes.Call(h, 0, uintptr(a), lwaAlpha)
			}
			pInvalidateRect.Call(h, 0, 0)
			return 0
		}
	case wmPaint:
		paintCover(h)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(h, m, wp, lp)
	return r
}

// paintCover draws the Skate 3 loading screen (double buffered).
func paintCover(h uintptr) {
	var ps paintStruct
	dc, _, _ := pBeginPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	defer pEndPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	var cr rect
	pGetClientRect.Call(h, uintptr(unsafe.Pointer(&cr)))
	w, ht := cr.R, cr.B
	mdc, _, _ := pCreateCompatibleDC.Call(dc)
	bmp, _, _ := pCreateCompatibleBitmap.Call(dc, uintptr(w), uintptr(ht))
	oldBmp, _, _ := pSelectObject.Call(mdc, bmp)

	fill := func(r rect, color uintptr) {
		b, _, _ := pCreateSolidBrush.Call(color)
		pFillRect.Call(mdc, uintptr(unsafe.Pointer(&r)), b)
		pDeleteObject.Call(b)
	}
	text := func(s string, r rect, size int32, face string, italic bool, color uintptr) {
		it := uintptr(0)
		if italic {
			it = 1
		}
		f, _, _ := pCreateFontW.Call(uintptr(-size), 0, 0, 0, 800, it, 0, 0, 0, 0, 0, 4, 0, uintptr(unsafe.Pointer(utf16(face))))
		old, _, _ := pSelectObject.Call(mdc, f)
		pSetTextColor.Call(mdc, color)
		pSetBkMode.Call(mdc, 1)
		u, _ := syscall.UTF16FromString(s)
		pDrawTextW.Call(mdc, uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1), uintptr(unsafe.Pointer(&r)), dtCenter|dtVCenter|dtSingleLine)
		pSelectObject.Call(mdc, old)
		pDeleteObject.Call(f)
	}

	fill(rect{0, 0, w, ht}, rgb(10, 10, 10))
	cy := ht / 2
	text("SKATE 3", rect{0, cy - ht/5, w, cy}, ht/7, "Impact", true, rgb(238, 238, 238))
	// orange sticker with the message
	sw, sh := w/4, ht/22
	fill(rect{(w - sw) / 2, cy + ht/40, (w + sw) / 2, cy + ht/40 + sh}, rgb(255, 90, 31))
	dots := strings.Repeat(".", int(time.Since(coverStart)/(400*time.Millisecond))%4)
	text(coverText+dots, rect{(w - sw) / 2, cy + ht/40, (w + sw) / 2, cy + ht/40 + sh}, sh*6/10, "Impact", true, rgb(17, 17, 17))
	// lime bar sweeping like a board rolling by
	bw, by := w/3, cy+ht/40+sh+ht/18
	fill(rect{(w - bw) / 2, by, (w + bw) / 2, by + 4}, rgb(40, 40, 40))
	pos := int32(time.Since(coverStart)/(16*time.Millisecond)) % (bw + bw/4)
	x0 := (w-bw)/2 + pos - bw/4
	x1 := x0 + bw/4
	if x0 < (w-bw)/2 {
		x0 = (w - bw) / 2
	}
	if x1 > (w+bw)/2 {
		x1 = (w + bw) / 2
	}
	if x1 > x0 {
		fill(rect{x0, by, x1, by + 4}, rgb(198, 255, 61))
	}
	text("Skate 3 Recomp Mods", rect{0, ht - ht/12, w, ht - ht/24}, ht/60, "Segoe UI", false, rgb(110, 110, 110))

	pBitBlt.Call(dc, 0, 0, uintptr(w), uintptr(ht), mdc, 0, 0, srcCopy)
	pSelectObject.Call(mdc, oldBmp)
	pDeleteObject.Call(bmp)
	pDeleteDC.Call(mdc)
}

// gameplayWatch reads the newest game log as it grows and reports when the
// boot automation says gameplay was reached.
type gameplayWatch struct {
	dir   string
	since time.Time
	path  string
	off   int64
	tail  []byte
}

var gameplayMarks = [][]byte{
	[]byte("gameplay reached"),
	[]byte("context=0x8001, value=0x1)"),
}

func newGameplayWatch(root string) *gameplayWatch {
	return &gameplayWatch{dir: filepath.Join(root, "logs"), since: time.Now().Add(-3 * time.Second)}
}

func (g *gameplayWatch) reached() bool {
	if g.path == "" {
		g.path = newestLog(g.dir, g.since)
		if g.path == "" {
			return false
		}
	}
	f, err := os.Open(g.path)
	if err != nil {
		return false
	}
	defer f.Close()
	if _, err := f.Seek(g.off, io.SeekStart); err != nil {
		return false
	}
	b, _ := io.ReadAll(io.LimitReader(f, 4<<20))
	g.off += int64(len(b))
	buf := append(g.tail, b...)
	for _, m := range gameplayMarks {
		if bytes.Contains(buf, m) {
			return true
		}
	}
	if len(buf) > 64 {
		buf = buf[len(buf)-64:]
	}
	g.tail = append(g.tail[:0], buf...)
	return false
}
