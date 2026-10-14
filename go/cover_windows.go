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
	"strconv"
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
	pStretchDIBits          = gdi32.NewProc("StretchDIBits")
	pSetStretchBltMode      = gdi32.NewProc("SetStretchBltMode")
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
	coverTipText string
	coverExpect  time.Duration // how long the last loads took (0 = unknown)
	coverMuted   bool
	coverProc    = syscall.NewCallback(coverWndProc)
)

type bitmapInfoHeader struct {
	Size                         uint32
	Width, Height                int32
	Planes, BitCount             uint16
	Compression, SizeImage       uint32
	XPelsPerMeter, YPelsPerMeter int32
	ClrUsed, ClrImportant        uint32
}

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
	coverTipText = loadingTip(iconsKey(root))
	coverExpect = 0
	if ms, err := strconv.Atoi(loadConfig(root)["boot_ms"]); err == nil && ms > 3000 {
		coverExpect = time.Duration(ms) * time.Millisecond
	}
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

// hideCover fades the loading screen out and gives the game its sound back.
func hideCover() {
	if coverHwnd != 0 && coverFade == 0 {
		coverFade = 1
	}
	if coverMuted {
		muteGame(gamePID, false)
		coverMuted = false
	}
}

// recordBoot keeps a running average of how long the direct load takes, for
// the loading bar.
func recordBoot(root string, d time.Duration) {
	c := loadConfig(root)
	ms := int(d / time.Millisecond)
	if old, err := strconv.Atoi(c["boot_ms"]); err == nil && old > 0 {
		ms = (old + ms) / 2
	}
	c["boot_ms"] = strconv.Itoa(ms)
	c.save(root)
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
	if coverFade == 0 && muteGame(gamePID, true) {
		coverMuted = true
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
				a := 255 - coverFade*64
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
	draw := func(s string, r rect, size int32, face string, italic bool, color uintptr, align uintptr) {
		it := uintptr(0)
		if italic {
			it = 1
		}
		f, _, _ := pCreateFontW.Call(uintptr(-size), 0, 0, 0, 800, it, 0, 0, 0, 0, 0, 4, 0, uintptr(unsafe.Pointer(utf16(face))))
		old, _, _ := pSelectObject.Call(mdc, f)
		pSetTextColor.Call(mdc, color)
		pSetBkMode.Call(mdc, 1)
		u, _ := syscall.UTF16FromString(s)
		pDrawTextW.Call(mdc, uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1), uintptr(unsafe.Pointer(&r)), align|dtVCenter|dtSingleLine)
		pSelectObject.Call(mdc, old)
		pDeleteObject.Call(f)
	}
	text := func(s string, r rect, size int32, face string, italic bool, color uintptr) {
		draw(s, r, size, face, italic, color, dtCenter)
	}

	fill(rect{0, 0, w, ht}, rgb(10, 10, 10))
	cy := ht / 2
	text("SKATE 3", rect{0, cy - ht/5, w, cy}, ht/7, "Bahnschrift SemiBold Condensed", false, rgb(238, 238, 238))
	// orange sticker with the message
	sw, sh := w/4, ht/22
	fill(rect{(w - sw) / 2, cy + ht/40, (w + sw) / 2, cy + ht/40 + sh}, rgb(255, 90, 31))
	dots := strings.Repeat(".", int(time.Since(coverStart)/(400*time.Millisecond))%4)
	text(coverText+dots, rect{(w - sw) / 2, cy + ht/40, (w + sw) / 2, cy + ht/40 + sh}, sh*6/10, "Bahnschrift SemiBold Condensed", false, rgb(17, 17, 17))
	// lime bar: how much is left, from how long the last loads took; the
	// first time it just sweeps like a board rolling by
	bw, by := w/3, cy+ht/40+sh+ht/18
	fill(rect{(w - bw) / 2, by, (w + bw) / 2, by + 4}, rgb(40, 40, 40))
	var x0, x1 int32
	if el := time.Since(coverStart); coverExpect > 0 {
		f := float64(el) / float64(coverExpect)
		if f > .97 {
			f = .97
		}
		x0, x1 = (w-bw)/2, (w-bw)/2+int32(f*float64(bw))
	} else {
		pos := int32(el/(16*time.Millisecond)) % (bw + bw/4)
		x0 = (w-bw)/2 + pos - bw/4
		x1 = x0 + bw/4
	}
	if x0 < (w-bw)/2 {
		x0 = (w - bw) / 2
	}
	if x1 > (w+bw)/2 {
		x1 = (w + bw) / 2
	}
	if x1 > x0 {
		fill(rect{x0, by, x1, by + 4}, rgb(198, 255, 61))
	}
	if coverTipText != "" {
		text(coverTipText, rect{w / 10, by + ht/14, w - w/10, by + ht/14 + ht/30}, ht/48, "Segoe UI", false, rgb(200, 200, 200))
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
	marks [][]byte
}

var (
	gameplayMarks = [][]byte{[]byte("gameplay reached"), []byte("context=0x8001, value=0x1)")}
	spotMarks     = [][]byte{[]byte("gameplay input sequence complete")}
)

// newGameplayWatch waits for gameplay, or with a starting spot for the
// inputs that take you there to finish.
func newGameplayWatch(root string) *gameplayWatch {
	g := &gameplayWatch{dir: filepath.Join(root, "logs"), since: time.Now().Add(-3 * time.Second), marks: gameplayMarks}
	if spotOn(root) {
		g.marks = spotMarks
	}
	return g
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
	for _, m := range g.marks {
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
