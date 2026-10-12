package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBootPhases(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "skate3.log")
	os.WriteFile(log, []byte(`[2026-10-12 20:00:01.500] [info] [rex] [t1] starting
[2026-10-12 20:00:04.000] [info] [rex] [t1] Skate 3 demo path: queued language-select A pulse
[2026-10-12 20:00:07.250] [info] [rex] [t1] Skate 3 demo path: enabled press-start auto-tap (press-start state)
[2026-10-12 20:00:15.000] [info] [rex] [t1] presence context=0x8001, value=0x1)
`), 0o644)
	start, _ := time.ParseInLocation("2006-01-02 15:04:05", "2026-10-12 20:00:00", time.Local)
	p := readBootPhases(log, start, 16*time.Second)
	if p.Open != 1500*time.Millisecond || p.Menus != 5750*time.Millisecond || p.Save != 7750*time.Millisecond || p.Spot != 0 {
		t.Fatalf("%+v", p)
	}
	root := t.TempDir()
	for i := 0; i < 20; i++ {
		addBootHistory(root, start, p)
	}
	h := bootHistory(root)
	if len(h) != bootHistMax || h[0]["total"] != 16000 || h[0]["save"] != 7750 {
		t.Fatalf("%v", h[0])
	}
}
