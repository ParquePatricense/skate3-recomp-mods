package main

import (
	"encoding/binary"
	"errors"
	"math"
	"os"
	"sort"
	"testing"
	"time"
)

// fakeMem simulates guest memory: static data, a heap object holding the
// skater position, lots of noise and a moving pedestrian.
type fakeMem struct {
	mem      map[uint32][]byte // region base -> bytes
	regs     []memRegion
	objBase  uint32
	posOff   uint32
	skaterX  float32
	pedX     float32
	moving   bool
	pedSteps int
}

func newFake(objBase, holder uint32) *fakeMem {
	f := &fakeMem{mem: map[uint32][]byte{}, objBase: objBase, posOff: 0x130}
	add := func(a, n uint32) {
		f.mem[a] = make([]byte, n)
		f.regs = append(f.regs, memRegion{a, n})
	}
	add(0x82000000, 0x10000)
	add(0x40000000, 0x40000)
	add(0xA0000000, 0x20000)
	sort.Slice(f.regs, func(i, j int) bool { return f.regs[i].addr < f.regs[j].addr })
	// static -> holder object -> skater object
	binary.BigEndian.PutUint32(f.at(0x82001230), holder)
	binary.BigEndian.PutUint32(f.at(holder+0x44), objBase)
	// static world data: lots of fixed triples
	for a := uint32(0xA0000000); a < 0xA0010000; a += 12 {
		f.putVec(a, float32(a%977)+10, 5, float32(a%331)+20)
	}
	f.skaterX = 100
	f.pedX = 300
	f.sync()
	return f
}

func (f *fakeMem) at(addr uint32) []byte {
	for _, r := range f.regs {
		if addr >= r.addr && addr < r.addr+r.size {
			return f.mem[r.addr][addr-r.addr:]
		}
	}
	panic("bad addr")
}

func (f *fakeMem) putVec(a uint32, x, y, z float32) {
	b := f.at(a)
	binary.BigEndian.PutUint32(b, math.Float32bits(x))
	binary.BigEndian.PutUint32(b[4:], math.Float32bits(y))
	binary.BigEndian.PutUint32(b[8:], math.Float32bits(z))
}

func (f *fakeMem) sync() {
	f.putVec(f.objBase+f.posOff, f.skaterX, 12, 250)
	f.putVec(0x40030000, f.pedX, 3, 80) // pedestrian always walking
}

func (f *fakeMem) regions() []memRegion { return f.regs }
func (f *fakeMem) read(addr uint32, buf []byte) error {
	for _, r := range f.regs {
		if addr >= r.addr && addr+uint32(len(buf)) <= r.addr+r.size {
			copy(buf, f.mem[r.addr][addr-r.addr:])
			return nil
		}
	}
	return errors.New("unmapped")
}

func (f *fakeMem) step(d time.Duration) {
	secs := float32(d.Seconds())
	if f.moving {
		f.skaterX += 6 * secs
	}
	f.pedX += 1.2 * secs
	f.sync()
}

func runFake(t *testing.T, f *fakeMem, root string) (*chain, error) {
	// The script alternates still/moving on each beep pattern: 1 beep = still,
	// 2 beeps = moving.
	ui := calibUI{
		say:   func(s string) { t.Log(s) },
		beep:  func(n int) { f.moving = n == 2 },
		sleep: func(d time.Duration) { f.step(d) },
	}
	return runCalibration(f, root, ui)
}

func TestCalibrationTwoSessions(t *testing.T) {
	root := t.TempDir()
	f := newFake(0x40012000, 0x40020000)
	ch, err := runFake(t, f, root)
	if err != nil || ch != nil {
		t.Fatalf("first run: chain=%v err=%v", ch, err)
	}
	if _, err := os.Stat(root + "/" + calibFile); err != nil {
		t.Fatal("no calibration file")
	}
	// "Restart": the objects live at different heap addresses now.
	f2 := newFake(0x40018000, 0x40024000)
	ch, err = runFake(t, f2, root)
	if err != nil || ch == nil {
		t.Fatalf("second run: chain=%v err=%v", ch, err)
	}
	if ch.String() != "82001230:44:130" {
		t.Fatalf("unexpected chain %s", ch)
	}
	p, ok := ch.resolve(f2)
	if !ok || p != 0x40018130 {
		t.Fatalf("resolve = %x %v", p, ok)
	}
}

func TestChainParse(t *testing.T) {
	c, err := parseChain("82001230:44:130")
	if err != nil || c.base != 0x82001230 || len(c.offs) != 2 || c.offs[1] != 0x130 {
		t.Fatal(c, err)
	}
}
