package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestScanStamp(t *testing.T) {
	d := t.TempDir()
	for _, off := range []int{0, 1<<20 - 20, 1<<20 - 3, 3 << 20} {
		data := bytes.Repeat([]byte{0x5a}, 4<<20)
		copy(data[off:], "build: skate3-v2.10.3-windows-Release@x")
		p := filepath.Join(d, "skate3.exe")
		os.WriteFile(p, data, 0o644)
		if v := scanStamp(p); v != "2.10.3" {
			t.Errorf("offset %d: got %q", off, v)
		}
	}
	os.WriteFile(filepath.Join(d, "skate3.exe"), []byte("nothing"), 0o644)
	if v := scanStamp(filepath.Join(d, "skate3.exe")); v != "" {
		t.Errorf("got %q", v)
	}
}

func TestBindsToml(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, tomlName), []byte("vsync = true\nkeybind_a = \"Space\" # salto\nmonitor = 2\n"), 0o644)
	if _, err := setBind(d, "keybind_a=Q"); err != nil {
		t.Fatal(err)
	}
	if _, err := setBind(d, "keybind_a=Ctrl+Q"); err == nil {
		t.Fatal("chord accepted")
	}
	m := readToml(d)
	if m["keybind_a"] != "Q" || m["vsync"] != "true" || m["monitor"] != "2" {
		t.Fatalf("%v", m)
	}
	st := controlsState(d)
	if st["vsync"] != true || st["monitor"] != 2 || st["fullscreen"] != true {
		t.Fatalf("%v", st)
	}
}
