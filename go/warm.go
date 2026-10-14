package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type warmFile struct {
	p string
	n int64
}

// warmList is what the warm-up reads, in order. First what the recomp opens
// before the language screen (skate3.exe, its DLLs and its shader/pipeline
// cache: on a cold PC that start took 4.25 s instead of 1.9 s), then the
// game's data files, smallest first.
func warmList(root, appData string) []warmFile {
	var first, data []warmFile
	add := func(list *[]warmFile, p string) {
		if st, err := os.Stat(p); err == nil && !st.IsDir() && st.Size() > 0 {
			*list = append(*list, warmFile{p, st.Size()})
		}
	}
	walk := func(list *[]warmFile, dir string) {
		filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				add(list, p)
			}
			return nil
		})
	}
	add(&first, filepath.Join(root, "skate3.exe"))
	if es, err := os.ReadDir(root); err == nil {
		for _, e := range es {
			if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".dll") {
				add(&first, filepath.Join(root, e.Name()))
			}
		}
	}
	walk(&first, recompCacheDir(root, appData))
	walk(&data, filepath.Join(root, "game", "data"))
	sort.SliceStable(data, func(i, j int) bool { return data[i].n < data[j].n })
	return append(first, data...)
}

// recompCacheDir follows the recomp (src/skate3_app_common.cpp): cache_path
// in skate3.toml, else <game>/cache when portable.txt is there, else
// %APPDATA%\skate3\cache.
func recompCacheDir(root, appData string) string {
	if c := readToml(root)["cache_path"]; c != "" {
		if !filepath.IsAbs(c) {
			c = filepath.Join(root, c)
		}
		return c
	}
	if _, err := os.Stat(filepath.Join(root, "portable.txt")); err == nil {
		return filepath.Join(root, "cache")
	}
	if appData == "" {
		return ""
	}
	return filepath.Join(appData, "skate3", "cache")
}
