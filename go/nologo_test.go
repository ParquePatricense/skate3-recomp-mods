package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestNoLogoLoose(t *testing.T) {
	root := t.TempDir()
	mv := filepath.Join(root, "game", "data", "movies", "EA_Blackbox.vp6")
	os.MkdirAll(filepath.Dir(mv), 0o755)
	os.MkdirAll(bigDir(root), 0o755)
	os.WriteFile(mv, []byte("movie"), 0o644)
	if s := noLogo(root); s != "off" {
		t.Fatal(s)
	}
	if _, err := setNoLogo(root, true); err != nil {
		t.Fatal(err)
	}
	if exists(mv) || !exists(mv+logoOffExt) || noLogo(root) != "on" {
		t.Fatal("not removed")
	}
	os.WriteFile(mv, []byte("movie"), 0o644) // a recomp update brings it back
	keepNoLogo(root)
	if exists(mv) {
		t.Fatal("not kept off")
	}
	noLogoFailed(root)
	if !exists(mv) || noLogo(root) != "failed" {
		t.Fatal("not restored")
	}
}

func TestNoLogoBig(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(bigDir(root), 0o755)
	const nameLen = 40
	d := make([]byte, 0x30+2*16+2*nameLen)
	copy(d, "EB")
	binary.BigEndian.PutUint16(d[2:], 3)
	binary.BigEndian.PutUint32(d[4:], 2)
	binary.BigEndian.PutUint32(d[12:], 0x30+2*16)
	d[20] = nameLen
	for i, n := range []string{"fe/a.bin", "movies/EA_Blackbox.vp6"} {
		binary.BigEndian.PutUint32(d[0x30+i*16:], 7)
		binary.BigEndian.PutUint32(d[0x30+i*16+4:], 100)
		binary.BigEndian.PutUint32(d[0x30+i*16+8:], 200)
		copy(d[0x30+2*16+i*nameLen+2:], n)
	}
	p := filepath.Join(bigDir(root), "movies.big")
	os.WriteFile(p, d, 0o644)
	if _, err := setNoLogo(root, true); err != nil {
		t.Fatal(err)
	}
	g, _ := os.ReadFile(p)
	if binary.BigEndian.Uint32(g[0x30+16+8:]) != 0 || binary.BigEndian.Uint32(g[0x30+8:]) != 200 {
		t.Fatal("entry not zeroed")
	}
	if _, err := setNoLogo(root, false); err != nil {
		t.Fatal(err)
	}
	g, _ = os.ReadFile(p)
	if string(g) != string(d) {
		t.Fatal("not restored")
	}
}
