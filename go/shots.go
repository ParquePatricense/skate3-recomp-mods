package main

// Loading screen backgrounds: the game's own location shots (the 512x256
// photos of every spot in Port Carverton the map menu shows), read straight
// from fedynamic.big on your PC each time. A different one on every load.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

type shot struct {
	name string // "Loading Docks"
	w, h int
	bgra []byte // top-down 32-bit pixels
}

// shotSkip are the 512x256 textures that aren't photos of a spot.
var shotSkip = []string{"billboard", "map_skatepark_temp", "map_skatepark_spacepark", "_empty"}

func shotEntries(b *bigArchive) []int {
	var out []int
	seen := map[string]bool{}
	for i := 0; i < b.count; i++ {
		r := b.namesOff + i*b.nameLen
		nm := b.data[r+2 : r+b.nameLen]
		if z := bytes.IndexByte(nm, 0); z >= 0 {
			nm = nm[:z]
		}
		n := string(nm)
		_, _, size, _ := b.entry(i)
		if !strings.HasSuffix(n, ".rx2") || size < 131600 || size > 131680 || seen[n] {
			continue
		}
		skip := false
		for _, s := range shotSkip {
			skip = skip || strings.Contains(n, s)
		}
		if !skip {
			seen[n] = true
			out = append(out, i)
		}
	}
	return out
}

func bigName(b *bigArchive, i int) string {
	r := b.namesOff + i*b.nameLen
	nm := b.data[r+2 : r+b.nameLen]
	if z := bytes.IndexByte(nm, 0); z >= 0 {
		nm = nm[:z]
	}
	return string(nm)
}

// shotTitle turns "map_skatepark_downtown_full.rx2" into "Skatepark Downtown Full".
func shotTitle(file string) string {
	n := strings.TrimPrefix(strings.TrimSuffix(file, ".rx2"), "map_")
	w := strings.Fields(strings.ReplaceAll(n, "_", " "))
	for i, s := range w {
		w[i] = strings.ToUpper(s[:1]) + s[1:]
	}
	return strings.Join(w, " ")
}

// pickShot reads one location shot, avoiding the one shown last time.
func pickShot(root, last string) (*shot, error) {
	d, err := os.ReadFile(filepath.Join(bigDir(root), "fedynamic.big"))
	if err != nil {
		return nil, err
	}
	b, err := parseBig(d)
	if err != nil {
		return nil, err
	}
	ents := shotEntries(b)
	rand.Shuffle(len(ents), func(i, j int) { ents[i], ents[j] = ents[j], ents[i] })
	for _, i := range ents {
		n := bigName(b, i)
		if n == last && len(ents) > 1 {
			continue
		}
		raw, err := readBigEntry(b, i)
		if err != nil {
			continue
		}
		if s, err := decodeRx2(raw); err == nil {
			s.name = n
			return s, nil
		}
	}
	return nil, errors.New("no shots")
}

// readBigEntry is bigArchive.read that also takes stored (uncompressed) entries.
func readBigEntry(b *bigArchive, i int) ([]byte, error) {
	off, csz, size, _ := b.entry(i)
	if csz == 0 {
		start := int(off) << b.shift
		if start+int(size) > len(b.data) {
			return nil, errors.New("short")
		}
		return b.data[start : start+int(size)], nil
	}
	return b.read(i)
}

// decodeRx2 decodes a one-texture RenderWare 4 (Xbox 360) file holding a
// DXT5 image, linear or in the Xbox 360 tiled layout.
func decodeRx2(d []byte) (*shot, error) {
	if len(d) < 0x60 || !bytes.Equal(d[:8], []byte("\x89RW4xb2\x00")) {
		return nil, errors.New("not rx2")
	}
	base := int(binary.BigEndian.Uint32(d[0x44:]))
	k := bytes.Index(d, []byte{0xff, 0xff, 0, 0, 0xff, 0xff, 0, 0})
	if k < 0 || k+20 > len(d) {
		return nil, errors.New("no texture header")
	}
	d0, d1, d2 := binary.BigEndian.Uint32(d[k+8:]), binary.BigEndian.Uint32(d[k+12:]), binary.BigEndian.Uint32(d[k+16:])
	tiled := d0>>31 == 1
	if d1&0x3f != 0x14 { // DXT4/5
		return nil, errors.New("not DXT5")
	}
	w, h := int(d2&0x1fff)+1, int(d2>>13&0x1fff)+1
	bw, bh := w/4, h/4
	pitch := (bw + 31) &^ 31 // tiled rows are padded to 32 blocks
	need := bw * bh * 16
	if tiled {
		need = pitch * ((bh + 31) &^ 31) * 16
	}
	if base+need > len(d) {
		return nil, errors.New("short texture")
	}
	px := make([]byte, w*h*4)
	var blk [16]byte
	for by := 0; by < bh; by++ {
		for bx := 0; bx < bw; bx++ {
			idx := by*bw + bx
			if tiled {
				idx = tiledIndex(bx, by, pitch)
			}
			src := d[base+idx*16:]
			for j := 0; j < 16; j += 2 { // 8-in-16 byte swap
				blk[j], blk[j+1] = src[j+1], src[j]
			}
			dxt5Block(blk[:], px, bx*4, by*4, w)
		}
	}
	return &shot{w: w, h: h, bgra: px}, nil
}

// dim darkens the picture so the loading text reads on top of it.
func (s *shot) dim(f float64) {
	for i := 0; i < len(s.bgra); i += 4 {
		s.bgra[i] = byte(float64(s.bgra[i]) * f)
		s.bgra[i+1] = byte(float64(s.bgra[i+1]) * f)
		s.bgra[i+2] = byte(float64(s.bgra[i+2]) * f)
	}
}

func dxt5Block(b []byte, px []byte, x0, y0, w int) {
	a0, a1 := int(b[0]), int(b[1])
	var alpha [8]int
	alpha[0], alpha[1] = a0, a1
	if a0 > a1 {
		for i := 1; i < 7; i++ {
			alpha[i+1] = ((7-i)*a0 + i*a1) / 7
		}
	} else {
		for i := 1; i < 5; i++ {
			alpha[i+1] = ((5-i)*a0 + i*a1) / 5
		}
		alpha[6], alpha[7] = 0, 255
	}
	abits := uint64(b[2]) | uint64(b[3])<<8 | uint64(b[4])<<16 | uint64(b[5])<<24 | uint64(b[6])<<32 | uint64(b[7])<<40
	c0, c1 := binary.LittleEndian.Uint16(b[8:]), binary.LittleEndian.Uint16(b[10:])
	rgb := func(c uint16) [3]int {
		return [3]int{int(c>>11&31) * 255 / 31, int(c>>5&63) * 255 / 63, int(c&31) * 255 / 31}
	}
	var col [4][3]int
	col[0], col[1] = rgb(c0), rgb(c1)
	for i := 0; i < 3; i++ {
		col[2][i] = (2*col[0][i] + col[1][i]) / 3
		col[3][i] = (col[0][i] + 2*col[1][i]) / 3
	}
	cbits := binary.LittleEndian.Uint32(b[12:])
	for p := 0; p < 16; p++ {
		c := col[cbits>>(2*p)&3]
		o := ((y0+p/4)*w + x0 + p%4) * 4
		px[o], px[o+1], px[o+2], px[o+3] = byte(c[2]), byte(c[1]), byte(c[0]), byte(alpha[abits>>(3*p)&7])
	}
}

// tiledIndex is where block (x, y) sits in an Xbox 360 tiled texture of
// 16-byte blocks, w blocks per row.
func tiledIndex(x, y, w int) int {
	const l = 4
	macro := ((y >> 5) * (w >> 5)) << (l + 7)
	micro := ((y & 6) << 2) << l
	base := macro + ((micro &^ 0xF) << 1) + (micro & 0xF) + ((y & 8) << (3 + l)) + ((y & 1) << 4)
	macro = (x >> 5) << (l + 7)
	micro = (x & 7) << l
	off := base + macro + ((micro &^ 0xF) << 1) + (micro & 0xF)
	return (((off &^ 0x1FF) << 3) + ((off & 0x1C0) << 2) + (off & 0x3F) + ((y & 16) << 7) + ((((y&8)>>2)+(x>>3))&3)<<6) >> l
}
