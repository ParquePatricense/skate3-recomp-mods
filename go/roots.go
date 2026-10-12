package main

// Two (or more) recomp versions: every game folder the modder has used is
// remembered, each with its own mods and settings (skate3-mods.ini lives in
// the folder), and you pick which one to play, e.g. the stable release and
// a test build.

import (
	"os"
	"path/filepath"
	"strings"
)

func rootsFile() string {
	if f := rememberFile(); f != "" {
		return filepath.Join(filepath.Dir(f), "carpetas.txt")
	}
	return ""
}

func knownRoots() []string {
	var out []string
	if b, err := os.ReadFile(rootsFile()); err == nil {
		for _, l := range strings.Split(string(b), "\n") {
			if l = strings.TrimSpace(l); l != "" && isGameRoot(l) && !containsFold(out, l) {
				out = append(out, l)
			}
		}
	}
	return out
}

func containsFold(list []string, s string) bool {
	for _, x := range list {
		if strings.EqualFold(filepath.Clean(x), filepath.Clean(s)) {
			return true
		}
	}
	return false
}

func saveRoots(list []string) {
	if f := rootsFile(); f != "" {
		os.MkdirAll(filepath.Dir(f), 0o755)
		os.WriteFile(f, []byte(strings.Join(list, "\n")+"\n"), 0o644)
	}
}

func addKnownRoot(dir string) {
	l := knownRoots()
	if !containsFold(l, dir) {
		saveRoots(append(l, dir))
	}
}

func forgetRoot(dir string) {
	var keep []string
	for _, x := range knownRoots() {
		if !strings.EqualFold(filepath.Clean(x), filepath.Clean(dir)) {
			keep = append(keep, x)
		}
	}
	saveRoots(keep)
}

// rootList is what the modder shows, the folder in use first.
func rootList(cur string) []map[string]any {
	out := []map[string]any{}
	for _, r := range append([]string{cur}, knownRoots()...) {
		if r == "" || (len(out) > 0 && strings.EqualFold(filepath.Clean(r), filepath.Clean(cur))) {
			continue
		}
		out = append(out, map[string]any{"path": r, "version": recompVersion(r), "on": r == cur})
	}
	return out
}
