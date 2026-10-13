package main

// The game's location photos (the map menu's shots of Port Carverton, read
// from fedynamic.big), shrunk to JPEGs for behind the modder.

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	shotMu    sync.Mutex
	shotCache = map[string]map[string]string{}
)

// shotImages: the location photos as small JPEGs, by spot key.
func shotImages(root string) map[string]string {
	if m, ok := shotCache[root]; ok {
		return m
	}
	m := map[string]string{}
	d, err := os.ReadFile(filepath.Join(bigDir(root), "fedynamic.big"))
	if err != nil {
		return m
	}
	b, err := parseBig(d)
	if err != nil {
		return m
	}
	for _, i := range shotEntries(b) {
		raw, err := readBigEntry(b, i)
		if err != nil {
			continue
		}
		s, err := decodeRx2(raw)
		if err != nil {
			continue
		}
		if u := shotJPEG(s, 2); u != "" {
			m[strings.TrimSuffix(bigName(b, i), ".rx2")] = u
		}
	}
	shotCache[root] = m
	return m
}

// shotJPEG makes a data URL of a shot, shrunk by div.
func shotJPEG(s *shot, div int) string {
	w, h := s.w/div, s.h/div
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, b int
			for dy := 0; dy < div; dy++ {
				for dx := 0; dx < div; dx++ {
					o := ((y*div+dy)*s.w + x*div + dx) * 4
					b += int(s.bgra[o])
					g += int(s.bgra[o+1])
					r += int(s.bgra[o+2])
				}
			}
			n := div * div
			o := (y*w + x) * 4
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = byte(r/n), byte(g/n), byte(b/n), 255
		}
	}
	var buf bytes.Buffer
	if jpeg.Encode(&buf, img, &jpeg.Options{Quality: 82}) != nil {
		return ""
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// modderBackground is one of the game's photos for behind the modder.
func modderBackground(root string) string {
	shotMu.Lock()
	m := shotImages(root)
	shotMu.Unlock()
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return ""
	}
	return m[keys[rand.Intn(len(keys))]]
}
