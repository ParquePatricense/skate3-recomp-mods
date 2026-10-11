package main

// skate3-mods.ini: tiny "key = value" settings file next to skate3.exe that
// the menu writes and the in-game helper (play mode) reads.

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const modsIni = "skate3-mods.ini"

type config map[string]string

func findRoot(game string) (string, error) {
	archive, err := findArchive(game)
	if err != nil {
		return "", err
	}
	// <root>/game/data/big/fedynamic.big -> <root>
	d := filepath.Dir(archive) // big
	for i := 0; i < 4; i++ {
		if exists(filepath.Join(d, "skate3.exe")) || exists(filepath.Join(d, "skate3")) {
			return d, nil
		}
		d = filepath.Dir(d)
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(archive)))), nil
}

func loadConfig(root string) config {
	c := config{}
	f, err := os.Open(filepath.Join(root, modsIni))
	if err != nil {
		return c
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			c[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return c
}

func (c config) on(key string) bool { return c[key] == "1" || strings.EqualFold(c[key], "true") }

func (c config) save(root string) error {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("# Skate 3 Recomp Mods - configuracion (la escribe Skate3-Mods.exe)\n")
	for _, k := range keys {
		b.WriteString(k + " = " + c[k] + "\n")
	}
	return os.WriteFile(filepath.Join(root, modsIni), []byte(b.String()), 0o644)
}
