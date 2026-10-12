package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMusicAndDirect(t *testing.T) {
	d := t.TempDir()
	songs := filepath.Join(d, "songs", "sub")
	os.MkdirAll(songs, 0o755)
	for _, n := range []string{"a.mp3", "b.FLAC", "c.txt", "d.m4a"} {
		os.WriteFile(filepath.Join(songs, n), []byte("x"), 0o644)
	}
	if n := len(musicTracks(filepath.Join(d, "songs"))); n != 3 {
		t.Fatalf("tracks %d", n)
	}
	if _, err := musicDo(d, "on"); err == nil {
		t.Fatal("on without folder")
	}
	c := loadConfig(d)
	c["music_dir"] = filepath.Join(d, "songs")
	c.save(d)
	for _, a := range []string{"on", "shuffle=off", "vol=40", "focus=off"} {
		if _, err := musicDo(d, a); err != nil {
			t.Fatal(a, err)
		}
	}
	st := musicState(d)
	if st["on"] != true || st["count"] != 3 || st["shuffle"] != false || st["vol"] != 40 || st["focus"] != false {
		t.Fatalf("%v", st)
	}
	if o := playOrder(loadConfig(d)); trackName(o[0]) != "a" {
		t.Fatalf("%v", o)
	}
	if _, err := setDirect(d, true); err != nil {
		t.Fatal(err)
	}
	m := readToml(d)
	if m["skate3_demo_path"] != "true" || m["skate3_demo_path_signed_in"] != "true" || controlsState(d)["direct"] != true {
		t.Fatalf("%v", m)
	}
	setDirect(d, false)
	if controlsState(d)["direct"] != false {
		t.Fatal("direct still on")
	}
}
