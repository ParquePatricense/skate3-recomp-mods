// Button icons module: swaps Skate 3's Xbox 360 button icons for
// PlayStation (PS3/PS4/PS5) ones inside game/data/big/fedynamic.big and
// restores the originals on request. Same logic as src/skate3_ps_buttons.py.
package main

import (
	"bytes"
	"compress/zlib"
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	archiveName  = "fedynamic.big"
	backupSuffix = ".xbox-original"
	stateSuffix  = ".ps-buttons"
	entryName    = "buttons.rx2"
	rx2Size      = 0x10A8B8
	texBytes     = 0x4000
)

//go:embed payload/*.bin
var payloadFS embed.FS

var styleNames = map[string]string{"ps3": "PS3", "ps4": "PS4", "ps5": "PS5"}

// ---------------------------------------------------------------- RefPack

func refpackDecompress(src []byte) ([]byte, error) {
	if len(src) < 5 || src[1] != 0xFB {
		return nil, errors.New("no es un stream RefPack")
	}
	flags := src[0]
	i := 2
	var size int
	if flags&0x80 != 0 {
		size = int(binary.BigEndian.Uint32(src[i:]))
		i += 4
		if flags&1 != 0 {
			i += 4
		}
	} else {
		size = int(src[i])<<16 | int(src[i+1])<<8 | int(src[i+2])
		i += 3
		if flags&1 != 0 {
			i += 3
		}
	}
	out := make([]byte, 0, size)
	for {
		if i >= len(src) {
			return nil, errors.New("RefPack truncado")
		}
		b0 := int(src[i])
		var lit, ln, off int
		switch {
		case b0 < 0x80:
			b1 := int(src[i+1])
			i += 2
			lit, ln, off = b0&3, ((b0>>2)&7)+3, ((b0&0x60)<<3)+b1+1
		case b0 < 0xC0:
			b1, b2 := int(src[i+1]), int(src[i+2])
			i += 3
			lit, ln, off = b1>>6, (b0&0x3F)+4, ((b1&0x3F)<<8)+b2+1
		case b0 < 0xE0:
			b1, b2, b3 := int(src[i+1]), int(src[i+2]), int(src[i+3])
			i += 4
			lit, ln, off = b0&3, ((b0&0x0C)<<6)+b3+5, ((b0&0x10)<<12)+(b1<<8)+b2+1
		case b0 < 0xFC:
			i++
			lit = ((b0 & 0x1F) << 2) + 4
			out = append(out, src[i:i+lit]...)
			i += lit
			continue
		default:
			i++
			lit = b0 & 3
			out = append(out, src[i:i+lit]...)
			if len(out) != size {
				return nil, errors.New("RefPack: tamaño incorrecto")
			}
			return out, nil
		}
		out = append(out, src[i:i+lit]...)
		i += lit
		start := len(out) - off
		if start < 0 {
			return nil, errors.New("RefPack: offset inválido")
		}
		for k := 0; k < ln; k++ {
			out = append(out, out[start+k])
		}
	}
}

func refpackCompress(data []byte) []byte {
	n := len(data)
	out := []byte{0x10, 0xFB, byte(n >> 16), byte(n >> 8), byte(n)}
	if n >= 1<<24 {
		out = []byte{0x90, 0xFB, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
	}
	table := map[uint32][]int{}
	key := func(p int) uint32 { return uint32(data[p])<<16 | uint32(data[p+1])<<8 | uint32(data[p+2]) }
	add := func(p int) {
		k := key(p)
		l := append(table[k], p)
		if len(l) > 32 {
			l = l[1:]
		}
		table[k] = l
	}
	litStart := 0
	flush := func(upto int) {
		for upto-litStart >= 4 {
			blk := (upto - litStart) &^ 3
			if blk > 112 {
				blk = 112
			}
			out = append(out, byte(0xE0+((blk-4)>>2)))
			out = append(out, data[litStart:litStart+blk]...)
			litStart += blk
		}
	}
	i := 0
	for i+3 <= n {
		bestLen, bestOff := 0, 0
		cands := table[key(i)]
		for c := len(cands) - 1; c >= 0; c-- {
			p := cands[c]
			off := i - p
			if off > 131072 {
				break
			}
			lim := n - i
			if lim > 1028 {
				lim = 1028
			}
			l := 0
			for l < lim && data[p+l] == data[i+l] {
				l++
			}
			if l > bestLen {
				bestLen, bestOff = l, off
				if l == lim {
					break
				}
			}
		}
		ok := (bestLen >= 3 && bestOff <= 1024) || (bestLen >= 4 && bestOff <= 16384) || bestLen >= 5
		if !ok {
			add(i)
			i++
			continue
		}
		flush(i)
		lit := i - litStart
		ln, off := bestLen, bestOff-1
		switch {
		case ln <= 10 && off < 1024:
			out = append(out, byte(((off>>3)&0x60)|((ln-3)<<2)|lit), byte(off))
		case ln <= 67 && off < 16384:
			out = append(out, byte(0x80|(ln-4)), byte((lit<<6)|(off>>8)), byte(off))
		default:
			out = append(out, byte(0xC0|((off>>12)&0x10)|(((ln-5)>>6)&0x0C)|lit), byte(off>>8), byte(off), byte(ln-5))
		}
		out = append(out, data[litStart:i]...)
		end := i + ln
		step := 1
		if ln >= 32 {
			step = 16
		}
		for q := i; q < end && q < n-2; q += step {
			add(q)
		}
		i = end
		litStart = i
	}
	flush(n)
	out = append(out, byte(0xFC|(n-litStart)))
	out = append(out, data[litStart:]...)
	return out
}

// ---------------------------------------------------------------- BIG (EB v3)

type bigArchive struct {
	data              []byte
	count             int
	shift             uint
	namesOff, nameLen int
}

func parseBig(d []byte) (*bigArchive, error) {
	if len(d) < 0x30 || d[0] != 'E' || d[1] != 'B' || binary.BigEndian.Uint16(d[2:]) != 3 {
		return nil, errors.New("no es un archivo BIG de Skate 3 (EB v3)")
	}
	return &bigArchive{data: d, count: int(binary.BigEndian.Uint32(d[4:])), shift: uint(d[10]),
		namesOff: int(binary.BigEndian.Uint32(d[12:])), nameLen: int(d[20])}, nil
}

func (b *bigArchive) find(name string) (int, error) {
	for i := 0; i < b.count; i++ {
		r := b.namesOff + i*b.nameLen
		nm := b.data[r+2 : r+b.nameLen]
		if z := bytes.IndexByte(nm, 0); z >= 0 {
			nm = nm[:z]
		}
		if string(nm) == name {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no se encontró %s en el archivo", name)
}

func (b *bigArchive) entry(i int) (off, csz, size, h uint32) {
	o := 0x30 + i*16
	d := b.data
	return binary.BigEndian.Uint32(d[o:]), binary.BigEndian.Uint32(d[o+4:]),
		binary.BigEndian.Uint32(d[o+8:]), binary.BigEndian.Uint32(d[o+12:])
}

func (b *bigArchive) read(i int) ([]byte, error) {
	off, csz, _, _ := b.entry(i)
	start := int(off) << b.shift
	raw := b.data[start : start+int(csz)]
	if len(raw) > 1 && raw[1] == 0xFB {
		return refpackDecompress(raw)
	}
	return raw, nil
}

func (b *bigArchive) replace(i int, payload []byte) error {
	comp := refpackCompress(payload)
	chk, err := refpackDecompress(comp)
	if err != nil || !bytes.Equal(chk, payload) {
		return errors.New("falló la verificación interna de compresión")
	}
	align := 1 << b.shift
	pos := (len(b.data) + align - 1) &^ (align - 1)
	b.data = append(b.data, make([]byte, pos-len(b.data))...)
	b.data = append(b.data, comp...)
	b.data = append(b.data, make([]byte, (align-len(b.data)%align)%align)...)
	_, _, _, h := b.entry(i)
	o := 0x30 + i*16
	binary.BigEndian.PutUint32(b.data[o:], uint32(pos>>b.shift))
	binary.BigEndian.PutUint32(b.data[o+4:], uint32(len(comp)))
	binary.BigEndian.PutUint32(b.data[o+8:], uint32(len(payload)))
	binary.BigEndian.PutUint32(b.data[o+12:], h)
	binary.BigEndian.PutUint32(b.data[0x1C:], uint32(len(b.data)))
	return nil
}

// ---------------------------------------------------------------- mod

func appDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func findArchive(game string) (string, error) {
	var roots []string
	if game != "" {
		roots = []string{game}
	} else if d, ok := locateGame(); ok {
		roots = []string{d}
	}
	rels := [][]string{{"game", "data", "big"}, {"data", "big"}, {"big"}, {}}
	for _, root := range roots {
		for _, rel := range rels {
			p := filepath.Join(append(append([]string{root}, rel...), archiveName)...)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return filepath.Abs(p)
			}
		}
	}
	return "", fmt.Errorf(T("no encontré game\\data\\big\\%s.\n"+
		"Abrí Skate3-Mods.exe y arrastrá la carpeta de skate3recomp cuando te la pida,\n"+
		"o usá --game \"<carpeta de skate3recomp>\".\n"+
		"El juego tiene que haberse instalado al menos una vez (Select ISO).",
		"could not find game\\data\\big\\%s.\n"+
			"Open Skate3-Mods.exe and drag the skate3recomp folder in when asked,\n"+
			"or use --game \"<skate3recomp folder>\".\n"+
			"The game must have been installed at least once (Select ISO)."), archiveName)
}

func loadStyle(style string) (map[int][]byte, error) {
	raw, err := payloadFS.ReadFile("payload/" + style + ".bin")
	if err != nil {
		return nil, err
	}
	zr, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	blob, err := io.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	rec := 1 + texBytes
	m := map[int][]byte{}
	for k := 0; k+rec <= len(blob); k += rec {
		m[int(blob[k])] = blob[k+1 : k+rec]
	}
	return m, nil
}

func patchRx2(rx2 []byte, style string) ([]byte, error) {
	if len(rx2) != rx2Size || !bytes.Equal(rx2[:8], []byte("\x89RW4xb2\x00")) {
		return nil, errors.New(entryName + T(" tiene un formato inesperado; esta versión del juego no está soportada", " has an unexpected format; this game version is not supported"))
	}
	base := int(binary.BigEndian.Uint32(rx2[0x44:]))
	if base+66*texBytes != rx2Size {
		return nil, errors.New(entryName + T(" tiene un layout de texturas inesperado", " has an unexpected texture layout"))
	}
	tex, err := loadStyle(style)
	if err != nil {
		return nil, err
	}
	out := append([]byte(nil), rx2...)
	for idx, t := range tex {
		copy(out[base+idx*texBytes:], t)
	}
	return out, nil
}

func copyFile(src, dst string) error {
	d, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, d, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func installIcons(style, game string) error {
	if _, ok := styleNames[style]; !ok {
		return fmt.Errorf(T("estilo desconocido %q (usá ps3, ps4 o ps5)", "unknown style %q (use ps3, ps4 or ps5)"), style)
	}
	archive, err := findArchive(game)
	if err != nil {
		return err
	}
	backup, state := archive+backupSuffix, archive+stateSuffix
	if !exists(backup) {
		fmt.Printf(T("Guardando copia del %s original...\n", "Backing up the original %s...\n"), archiveName)
		if err := copyFile(archive, backup); err != nil {
			return err
		}
	}
	orig, err := os.ReadFile(backup)
	if err != nil {
		return err
	}
	big, err := parseBig(orig)
	if err != nil {
		return err
	}
	idx, err := big.find(entryName)
	if err != nil {
		return err
	}
	fmt.Printf(T("Instalando iconos de %s...\n", "Installing %s icons...\n"), styleNames[style])
	rx2, err := big.read(idx)
	if err != nil {
		return err
	}
	patched, err := patchRx2(rx2, style)
	if err != nil {
		return err
	}
	if err := big.replace(idx, patched); err != nil {
		return err
	}
	tmp := archive + ".tmp"
	if err := os.WriteFile(tmp, big.data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, archive); err != nil {
		os.Remove(tmp)
		return fmt.Errorf(T("no pude reemplazar %s (¿el juego está abierto?): %v", "could not replace %s (is the game running?): %v"), archiveName, err)
	}
	if err := os.WriteFile(state, []byte(style), 0o644); err != nil {
		return err
	}
	recordStamp(archive, "icons")
	fmt.Printf(T("Listo: iconos de %s instalados.\n", "Done: %s icons installed.\n"), styleNames[style])
	return nil
}

func uninstallIcons(game string) error {
	archive, err := findArchive(game)
	if err != nil {
		return err
	}
	backup, state := archive+backupSuffix, archive+stateSuffix
	if !exists(backup) {
		fmt.Println(T("No hay nada que desinstalar: ya están los iconos originales de Xbox.", "Nothing to uninstall: the original Xbox icons are already in place."))
		return nil
	}
	os.Remove(archive)
	if err := os.Rename(backup, archive); err != nil {
		return fmt.Errorf(T("no pude restaurar %s (¿el juego está abierto?): %v", "could not restore %s (is the game running?): %v"), archiveName, err)
	}
	os.Remove(state)
	fmt.Println(T("Listo: iconos originales de Xbox restaurados.", "Done: original Xbox icons restored."))
	return nil
}

// iconsStatus returns "Xbox (original)" or the installed PlayStation style.
func iconsStatus(game string) (string, error) {
	archive, err := findArchive(game)
	if err != nil {
		return "", err
	}
	st, err := os.ReadFile(archive + stateSuffix)
	if err == nil && exists(archive+backupSuffix) {
		s := strings.TrimSpace(string(st))
		if n, ok := styleNames[s]; ok {
			s = n
		}
		return s, nil
	}
	return "Xbox (original)", nil
}
