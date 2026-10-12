package main

import (
	"testing"
	"time"
)

func TestPlayTime(t *testing.T) {
	root := t.TempDir()
	s := time.Date(2026, 10, 12, 23, 30, 0, 0, time.Local) // Monday
	addPlayTime(root, s, s.Add(time.Hour), 12*time.Second)
	st := playStats(root, s.Add(time.Hour))
	if st["today"] != 1800 || st["week"] != 3600 || st["total"] != 3600 {
		t.Fatal(st)
	}
	l := st["last"].(map[string]any)
	if l["secs"] != 3600 || l["load"] != 12000 {
		t.Fatal(l)
	}
	if d := st["days"].([]map[string]any); len(d) != 7 || d[6]["secs"] != 1800 || d[5]["secs"] != 1800 {
		t.Fatal(d)
	}
}
