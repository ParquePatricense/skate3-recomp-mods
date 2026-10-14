package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWarmListOrder(t *testing.T) {
	root := t.TempDir()
	app := t.TempDir()
	put := func(p string, n int) {
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, make([]byte, n), 0o644)
	}
	put(filepath.Join(root, "game", "data", "big.big"), 30)
	put(filepath.Join(root, "game", "data", "small.big"), 10)
	put(filepath.Join(root, "skate3.exe"), 50)
	put(filepath.Join(root, "rexruntime.dll"), 40)
	put(filepath.Join(app, "skate3", "cache", "shaders.bin"), 20)
	got := []string{}
	for _, f := range warmList(root, app) {
		got = append(got, filepath.Base(f.p))
	}
	want := []string{"skate3.exe", "rexruntime.dll", "shaders.bin", "small.big", "big.big"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	// portable install keeps its cache next to the game
	put(filepath.Join(root, "portable.txt"), 1)
	if d := recompCacheDir(root, app); d != filepath.Join(root, "cache") {
		t.Fatalf("portable cache dir %s", d)
	}
}
