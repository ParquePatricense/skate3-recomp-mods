//go:build windows

package main

// The song player of "Tu música", driven by MCI (winmm), which plays MP3,
// WAV, WMA and, through Windows' codecs, M4A and FLAC. Every call happens on
// the helper's message-loop thread.

import (
	"strconv"
	"syscall"
	"unsafe"
)

var pMciSendStringW = syscall.NewLazyDLL("winmm.dll").NewProc("mciSendStringW")

const (
	vkSpace = 0x20
	vkLeft  = 0x25
	vkUp    = 0x26
	vkRight = 0x27
	vkDown  = 0x28
)

func mci(cmd string) (string, bool) {
	c, _ := syscall.UTF16PtrFromString(cmd)
	buf := make([]uint16, 256)
	r, _, _ := pMciSendStringW.Call(uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0)
	return syscall.UTF16ToString(buf), r == 0
}

type songPlayer struct {
	tracks       []string
	i, vol       int
	open, paused bool
	autoPaused   bool
}

var music *songPlayer

func newSongPlayer(c config) *songPlayer {
	p := &songPlayer{tracks: playOrder(c), vol: musicVolume(c), i: -1}
	if len(p.tracks) == 0 {
		return nil
	}
	p.step(1)
	return p
}

func (p *songPlayer) close() {
	if p.open {
		mci("close s3song")
		p.open = false
	}
}

// step moves dir songs ahead (or back) and plays it, skipping files Windows
// can't open.
func (p *songPlayer) step(dir int) {
	p.close()
	for tries := 0; tries < len(p.tracks); tries++ {
		p.i = (p.i + dir + len(p.tracks)) % len(p.tracks)
		if _, ok := mci(`open "` + p.tracks[p.i] + `" type mpegvideo alias s3song`); !ok {
			continue
		}
		p.open = true
		p.setVolume(p.vol)
		mci("play s3song")
		p.paused, p.autoPaused = false, false
		return
	}
}

func (p *songPlayer) setVolume(v int) {
	p.vol = max(0, min(100, v))
	if p.open {
		mci("setaudio s3song volume to " + strconv.Itoa(p.vol*10))
	}
}

func (p *songPlayer) pause(on bool) {
	if !p.open || p.paused == on {
		return
	}
	if on {
		mci("pause s3song")
	} else {
		mci("resume s3song")
	}
	p.paused = on
}

// tick goes to the next song when one ends, and pauses while the game is in
// the background (Alt+Tab) when music_focus is on.
func (p *songPlayer) tick(front bool) {
	if musicFocus(hc) {
		if !front && !p.paused {
			p.pause(true)
			p.autoPaused = true
		} else if front && p.autoPaused {
			p.autoPaused = false
			p.pause(false)
		}
	}
	if p.open && !p.paused {
		if mode, ok := mci("status s3song mode"); ok && mode == "stopped" {
			p.step(1)
		}
	}
}

// musicKey handles Ctrl+Alt+→ / ← (next / previous), Ctrl+Alt+Space
// (pause) and Ctrl+Alt+↑ / ↓ (volume). It reports whether it used the key.
func musicKey(vk uint32) bool {
	if music == nil || !keyDown(vkControl) || !keyDown(0x12) {
		return false
	}
	switch vk {
	case vkRight:
		music.step(1)
	case vkLeft:
		music.step(-1)
	case vkSpace:
		music.autoPaused = false
		music.pause(!music.paused)
	case vkUp:
		music.setVolume(music.vol + 10)
	case vkDown:
		music.setVolume(music.vol - 10)
	default:
		return false
	}
	return true
}
