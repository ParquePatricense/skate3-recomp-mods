package main

// Speedometer calibration: finds where the running game keeps the skater's
// world position, then a pointer chain from the game's static data to it so
// the helper can find it again after every launch.
//
// The game runs in guest (Xbox 360) memory that the recomp maps at a fixed
// host base; values are big-endian. The scan compares full snapshots taken
// while the player stands still and while they skate:
//   still -> still   : position unchanged
//   still -> skating : position moved a plausible distance
//   skating -> still : (ignored, the skater is braking)
//   still -> still   : unchanged again
//   still -> skating : moved again
// What survives is a handful of positions that follow the skater. A second
// calibration (after restarting the game) keeps only the pointer chains that
// still lead to a surviving position.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

// guestMem reads the game's guest memory.
type guestMem interface {
	regions() []memRegion
	read(addr uint32, buf []byte) error
}

type memRegion struct{ addr, size uint32 }

const (
	staticLo   = 0x82000000 // xex image (static data lives here)
	staticHi   = 0x84000000
	maxWindow  = 0x2000 // max struct offset accepted in a pointer chain
	calibFile  = "skate3-speed-calibration.txt"
	maxTargets = 96
)

func scanRanges(addr uint32) bool {
	return (addr >= 0x40000000 && addr < 0x80000000) || (addr >= 0xA0000000 && addr < 0xC0000000)
}

type snapshot struct {
	regs []memRegion
	data [][]byte
}

func takeSnapshot(m guestMem, filter func(uint32) bool) *snapshot {
	s := &snapshot{}
	for _, r := range m.regions() {
		if !filter(r.addr) {
			continue
		}
		b := make([]byte, r.size)
		if m.read(r.addr, b) != nil {
			continue
		}
		s.regs = append(s.regs, r)
		s.data = append(s.data, b)
	}
	return s
}

func (s *snapshot) find(addr uint32) (int, int) {
	i := sort.Search(len(s.regs), func(i int) bool { return s.regs[i].addr+s.regs[i].size > addr })
	if i < len(s.regs) && addr >= s.regs[i].addr {
		return i, int(addr - s.regs[i].addr)
	}
	return -1, 0
}

func f32(b []byte) float32 { return math.Float32frombits(binary.BigEndian.Uint32(b)) }

func vec(b []byte) (x, y, z float64, ok bool) {
	x, y, z = float64(f32(b)), float64(f32(b[4:])), float64(f32(b[8:]))
	for _, v := range []float64{x, y, z} {
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 30000 {
			return 0, 0, 0, false
		}
	}
	big := 0
	for _, v := range []float64{x, y, z} {
		if math.Abs(v) > 0.5 {
			big++
		}
	}
	return x, y, z, big >= 2
}

func dist(a, b []byte) (float64, bool) {
	ax, ay, az, ok1 := vec(a)
	bx, by, bz, ok2 := vec(b)
	if !ok1 || !ok2 {
		return 0, false
	}
	return math.Sqrt((ax-bx)*(ax-bx) + (ay-by)*(ay-by) + (az-bz)*(az-bz)), true
}

// candidates is a set of guest addresses of 12-byte float triples.
type candidates map[uint32]struct{}

// stillBits marks every 4-aligned triple that did not move between two
// full snapshots (1 bit per slot, so it stays small even for ~1 GB of RAM).
func stillBits(a, b *snapshot) [][]uint64 {
	out := make([][]uint64, len(b.regs))
	for ri, r := range b.regs {
		ai, _ := a.find(r.addr)
		if ai < 0 || a.regs[ai].addr != r.addr || a.regs[ai].size != r.size {
			continue
		}
		ad, bd := a.data[ai], b.data[ri]
		bits := make([]uint64, len(bd)/4/64+1)
		for o := 0; o+12 <= len(bd); o += 4 {
			if d, ok := dist(ad[o:o+12], bd[o:o+12]); ok && still(d) {
				bits[o/4/64] |= 1 << (uint(o/4) % 64)
			}
		}
		out[ri] = bits
	}
	return out
}

// movedFromBits keeps the marked triples of b that moved a plausible
// distance by snapshot s.
func movedFromBits(bits [][]uint64, b, s *snapshot) candidates {
	out := candidates{}
	for ri, r := range b.regs {
		if bits[ri] == nil {
			continue
		}
		si, _ := s.find(r.addr)
		if si < 0 || s.regs[si].addr != r.addr || s.regs[si].size != r.size {
			continue
		}
		bd, sd := b.data[ri], s.data[si]
		for w, word := range bits[ri] {
			for word != 0 {
				k := bitsTrailing(word)
				word &= word - 1
				o := (w*64 + k) * 4
				if o+12 > len(bd) {
					continue
				}
				if d, ok := dist(bd[o:o+12], sd[o:o+12]); ok && moved(d) {
					out[r.addr+uint32(o)] = struct{}{}
				}
			}
		}
	}
	return out
}

func bitsTrailing(x uint64) int {
	n := 0
	for x&1 == 0 {
		x >>= 1
		n++
	}
	return n
}

func filterSet(c candidates, a, b *snapshot, keep func(d float64) bool) candidates {
	out := candidates{}
	for addr := range c {
		ai, ao := a.find(addr)
		bi, bo := b.find(addr)
		if ai < 0 || bi < 0 || ao+12 > len(a.data[ai]) || bo+12 > len(b.data[bi]) {
			continue
		}
		if d, ok := dist(a.data[ai][ao:ao+12], b.data[bi][bo:bo+12]); ok && keep(d) {
			out[addr] = struct{}{}
		}
	}
	return out
}

func still(d float64) bool { return d < 0.05 }
func moved(d float64) bool { return d > 3 && d < 300 }
func (c candidates) sorted() []uint32 {
	out := make([]uint32, 0, len(c))
	for a := range c {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// chain: pos = read32(base) [+off -> read32]... + last offset
type chain struct {
	base uint32
	offs []uint32
}

func (c chain) String() string {
	parts := []string{fmt.Sprintf("%08X", c.base)}
	for _, o := range c.offs {
		parts = append(parts, fmt.Sprintf("%X", o))
	}
	return strings.Join(parts, ":")
}

func parseChain(s string) (chain, error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) < 2 {
		return chain{}, errors.New("cadena invalida")
	}
	var c chain
	for i, p := range parts {
		v, err := strconv.ParseUint(p, 16, 32)
		if err != nil {
			return chain{}, err
		}
		if i == 0 {
			c.base = uint32(v)
		} else {
			c.offs = append(c.offs, uint32(v))
		}
	}
	return c, nil
}

func read32(m guestMem, addr uint32) (uint32, bool) {
	var b [4]byte
	if m.read(addr, b[:]) != nil {
		return 0, false
	}
	return binary.BigEndian.Uint32(b[:]), true
}

func (c chain) resolve(m guestMem) (uint32, bool) {
	p, ok := read32(m, c.base)
	if !ok || p == 0 {
		return 0, false
	}
	for _, o := range c.offs[:len(c.offs)-1] {
		if p, ok = read32(m, p+o); !ok || p == 0 {
			return 0, false
		}
	}
	return p + c.offs[len(c.offs)-1], true
}

type window struct{ lo, hi, target uint32 }

// holders returns, for every u32 in snap pointing into [target-maxWindow,
// target], the pair (holder address, target). Optionally restricted to
// holders in [lo, hi).
func holders(snap *snapshot, targets []uint32, lo, hi uint32) map[uint32][]uint32 {
	ws := make([]window, 0, len(targets))
	for _, t := range targets {
		l := uint32(0)
		if t > maxWindow {
			l = t - maxWindow
		}
		ws = append(ws, window{l, t, t})
	}
	sort.Slice(ws, func(i, j int) bool { return ws[i].lo < ws[j].lo })
	out := map[uint32][]uint32{}
	for ri, r := range snap.regs {
		if r.addr+r.size <= lo || r.addr >= hi {
			continue
		}
		d := snap.data[ri]
		for o := 0; o+4 <= len(d); o += 4 {
			h := r.addr + uint32(o)
			if h < lo || h >= hi {
				continue
			}
			v := binary.BigEndian.Uint32(d[o:])
			if v < 0x40000000 {
				continue
			}
			k := sort.Search(len(ws), func(i int) bool { return ws[i].lo > v })
			for j := k - 1; j >= 0 && j >= k-8; j-- {
				if v >= ws[j].lo && v <= ws[j].hi {
					out[h] = append(out[h], ws[j].target)
				}
			}
		}
	}
	return out
}

func findChains(snap *snapshot, targets []uint32) []chain {
	var out []chain
	seen := map[string]bool{}
	add := func(c chain) {
		if k := c.String(); !seen[k] {
			seen[k] = true
			out = append(out, c)
		}
	}
	lvl1 := holders(snap, targets, 0x40000000, 0xC0000000)
	var mids []uint32
	for h, ts := range lvl1 {
		for _, t := range ts {
			v, _ := snapRead32(snap, h)
			if h >= staticLo && h < staticHi {
				add(chain{h, []uint32{t - v}})
			}
		}
		if !(h >= staticLo && h < staticHi) {
			mids = append(mids, h)
		}
	}
	if len(mids) > 200000 {
		mids = mids[:200000]
	}
	lvl2 := holders(snap, mids, staticLo, staticHi)
	for s, hs := range lvl2 {
		sv, _ := snapRead32(snap, s)
		for _, h := range hs {
			hv, _ := snapRead32(snap, h)
			for _, t := range lvl1[h] {
				add(chain{s, []uint32{h - sv, t - hv}})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i].offs) != len(out[j].offs) {
			return len(out[i].offs) < len(out[j].offs)
		}
		return out[i].String() < out[j].String()
	})
	return out
}

func snapRead32(s *snapshot, addr uint32) (uint32, bool) {
	i, o := s.find(addr)
	if i < 0 || o+4 > len(s.data[i]) {
		return 0, false
	}
	return binary.BigEndian.Uint32(s.data[i][o:]), true
}

type calibUI struct {
	say   func(string)
	beep  func(n int)
	sleep func(time.Duration)
}

// calibrate runs the guided scan and returns the positions that follow the
// skater plus the final snapshot (used for the pointer scan).
func calibrate(m guestMem, ui calibUI) ([]uint32, *snapshot, error) {
	ui.say("En 10 segundos empieza. Pasa al juego y deja al skater QUIETO, parado sobre la tabla.")
	ui.sleep(10 * time.Second)
	ui.beep(1)
	ui.say("[1/5] Quieto...")
	a := takeSnapshot(m, scanRanges)
	if len(a.regs) == 0 {
		return nil, nil, errors.New("no pude leer la memoria del juego")
	}
	ui.sleep(2 * time.Second)
	b := takeSnapshot(m, scanRanges)
	bits := stillBits(a, b)
	a = nil
	debug.FreeOSMemory()
	ui.beep(2)
	ui.say("[2/5] PATINA derecho y rapido durante 8 segundos...")
	ui.sleep(8 * time.Second)
	s := takeSnapshot(m, scanRanges)
	c := movedFromBits(bits, b, s)
	bits = nil
	ui.beep(1)
	ui.say(fmt.Sprintf("[3/5] FRENA y quedate QUIETO... (%d candidatos)", len(c)))
	ui.sleep(7 * time.Second)
	b = takeSnapshot(m, scanRanges)
	ui.sleep(2 * time.Second)
	s = takeSnapshot(m, scanRanges)
	c = filterSet(c, b, s, still)
	ui.beep(2)
	ui.say(fmt.Sprintf("[4/5] PATINA otra vez durante 8 segundos... (%d candidatos)", len(c)))
	ui.sleep(8 * time.Second)
	b = s
	s = takeSnapshot(m, scanRanges)
	c = filterSet(c, b, s, moved)
	ui.beep(3)
	ui.say(fmt.Sprintf("[5/5] Listo, ya podes soltar. Quedaron %d posiciones que siguen al skater.", len(c)))
	if len(c) == 0 {
		return nil, nil, errors.New("no encontre la posicion del skater; repeti la calibracion siguiendo los pitidos")
	}
	targets := c.sorted()
	// Overlapping triples (addr and addr+4) can both survive; the game keeps
	// vectors 16-byte aligned (VMX), so prefer those when there are any.
	var aligned []uint32
	for _, t := range targets {
		if t%16 == 0 {
			aligned = append(aligned, t)
		}
	}
	if len(aligned) > 0 {
		targets = aligned
	}
	if len(targets) > maxTargets {
		step := float64(len(targets)) / maxTargets
		var t2 []uint32
		for i := 0; i < maxTargets; i++ {
			t2 = append(t2, targets[int(float64(i)*step)])
		}
		targets = t2
	}
	b, s = nil, nil
	debug.FreeOSMemory()
	full := takeSnapshot(m, func(a uint32) bool { return scanRanges(a) || (a >= staticLo && a < staticHi) })
	return targets, full, nil
}

// runCalibration: full flow incl. cross-session validation. Returns the chain
// to use (if validated) or nil when another run after restarting is needed.
func runCalibration(m guestMem, root string, ui calibUI) (*chain, error) {
	targets, full, err := calibrate(m, ui)
	if err != nil {
		return nil, err
	}
	set := map[uint32]bool{}
	for _, t := range targets {
		set[t] = true
	}
	prevPath := filepath.Join(root, calibFile)
	if prev, err := os.ReadFile(prevPath); err == nil {
		var good []chain
		for _, line := range strings.Split(string(prev), "\n") {
			ch, err := parseChain(line)
			if err != nil {
				continue
			}
			if p, ok := ch.resolve(m); ok && (set[p] || nearAny(p, targets)) {
				good = append(good, ch)
			}
		}
		if len(good) > 0 {
			sort.SliceStable(good, func(i, j int) bool { return len(good[i].offs) < len(good[j].offs) })
			os.Remove(prevPath)
			return &good[0], nil
		}
	}
	ui.say("Buscando como volver a encontrar la posicion la proxima vez...")
	chains := findChains(full, targets)
	if len(chains) == 0 {
		return nil, errors.New("encontre la posicion pero no una forma estable de volver a encontrarla")
	}
	var b strings.Builder
	for _, ch := range chains {
		b.WriteString(ch.String() + "\n")
	}
	if err := os.WriteFile(prevPath, []byte(b.String()), 0o644); err != nil {
		return nil, err
	}
	return nil, nil
}

func nearAny(p uint32, ts []uint32) bool {
	for _, t := range ts {
		if p == t {
			return true
		}
	}
	return false
}
