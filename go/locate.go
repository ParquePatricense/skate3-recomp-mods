// Finding the game on its own: Skate3-Mods.exe can be opened from anywhere
// (Downloads, the desktop...). It looks next to itself first, then in the
// folder it remembered last time, then searches the disks. Once found, the
// folder is remembered and the mod copies itself into it, so the shortcut
// and the .bat keep working even if the downloaded copy is deleted.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	modsExe   = "Skate3-Mods.exe"
	playBat   = "Jugar Skate 3 (mods).bat"
	batScript = "@echo off\r\ncd /d \"%~dp0\"\r\nstart \"\" \"%~dp0Skate3-Mods.exe\" play\r\n"
)

// isGameRoot reports whether dir is an installed skate3recomp folder.
func isGameRoot(dir string) bool {
	return (exists(filepath.Join(dir, "skate3.exe")) || exists(filepath.Join(dir, "skate3"))) &&
		exists(filepath.Join(dir, "game", "data", "big", archiveName))
}

func rememberFile() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "Skate3-Mods", "carpeta-del-juego.txt")
}

func rememberedRoot() string {
	f := rememberFile()
	if f == "" {
		return ""
	}
	b, err := os.ReadFile(f)
	if err != nil {
		return ""
	}
	dir := strings.TrimSpace(string(b))
	if isGameRoot(dir) {
		return dir
	}
	return ""
}

func rememberRoot(dir string) {
	if f := rememberFile(); f != "" {
		os.MkdirAll(filepath.Dir(f), 0o755)
		os.WriteFile(f, []byte(dir+"\n"), 0o644)
	}
}

// searchStarts are the places worth searching, most likely first.
func searchStarts() []string {
	var starts []string
	if home, err := os.UserHomeDir(); err == nil {
		for _, sub := range []string{"Desktop", "Escritorio", "Downloads", "Descargas", "Documents", "Documentos", "Games", "Juegos", ""} {
			starts = append(starts, filepath.Join(home, sub))
		}
	}
	if runtime.GOOS == "windows" {
		for c := 'C'; c <= 'Z'; c++ {
			if d := string(c) + `:\`; exists(d) {
				starts = append(starts, d)
			}
		}
	}
	return starts
}

var skipDirs = map[string]bool{"windows": true, "$recycle.bin": true, "programdata": true, "appdata": true,
	"system volume information": true, "node_modules": true, ".git": true, "winsxs": true, "microsoft": true}

// searchGame walks the usual places (5 levels deep) and returns every
// skate3recomp folder it finds.
func searchGame() []string {
	var found []string
	seen := map[string]bool{}
	visited := 0
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if visited > 400000 || seen[strings.ToLower(dir)] {
			return
		}
		seen[strings.ToLower(dir)] = true
		visited++
		if isGameRoot(dir) {
			found = append(found, dir)
			return
		}
		if depth == 0 {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			name := e.Name()
			if !e.IsDir() || skipDirs[strings.ToLower(name)] || strings.HasPrefix(name, ".") {
				continue
			}
			walk(filepath.Join(dir, name), depth-1)
		}
	}
	for _, s := range searchStarts() {
		walk(s, 5)
	}
	return found
}

// locateGame finds the game without asking: next to the exe or in the
// current folder, the remembered folder, then a disk search.
func locateGame() (string, bool) {
	for _, d := range []string{appDir(), workDir()} {
		if d != "" && isGameRoot(d) {
			return d, true
		}
	}
	if d := rememberedRoot(); d != "" {
		return d, true
	}
	fmt.Println(T("Buscando la carpeta de skate3recomp en la PC (solo la primera vez)...",
		"Looking for the skate3recomp folder on this PC (first time only)..."))
	found := searchGame()
	switch len(found) {
	case 0:
		return "", false
	case 1:
		rememberRoot(found[0])
		return found[0], true
	}
	fmt.Println(T("Encontré varias copias del juego:", "Found several copies of the game:"))
	for i, d := range found {
		fmt.Printf("  %d) %s\n", i+1, d)
	}
	n := 0
	fmt.Sscan(ask(T("¿Cuál uso? ", "Which one? ")), &n)
	if n < 1 || n > len(found) {
		n = 1
	}
	rememberRoot(found[n-1])
	return found[n-1], true
}

func workDir() string {
	d, _ := os.Getwd()
	return d
}

// askGameFolder asks for the folder until it gets a valid one (drag and
// drop of the folder onto the window pastes its path). Empty answer = give up.
func askGameFolder(missing bool) (string, bool) {
	if !missing {
		fmt.Println(T("\nArrastrá la carpeta del juego (la que tiene skate3.exe) a esta ventana y apretá Enter (Enter vacío para cancelar).",
			"\nDrag the game folder (the one with skate3.exe) onto this window and press Enter (empty Enter to cancel)."))
	} else {
		fmt.Println(T("\nNo encontré skate3recomp. Arrastrá la carpeta del juego (la que tiene skate3.exe) a esta ventana y apretá Enter.",
			"\nCould not find skate3recomp. Drag the game folder (the one with skate3.exe) onto this window and press Enter."))
	}
	for {
		p := strings.Trim(ask("> "), "\"' ")
		if p == "" {
			return "", false
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			p = filepath.Dir(p) // they dropped skate3.exe itself
		}
		if isGameRoot(p) {
			rememberRoot(p)
			return p, true
		}
		fmt.Println(T("Esa carpeta no tiene skate3.exe y game\\data\\big. Probá de nuevo (Enter vacío para salir).",
			"That folder has no skate3.exe and game\\data\\big. Try again (empty Enter to quit)."))
	}
}

// settleIn copies the running exe (and the play .bat) into the game folder
// when it was opened from somewhere else. Returns true if it copied.
func settleIn(root string) bool {
	if runtime.GOOS != "windows" || strings.EqualFold(filepath.Clean(appDir()), filepath.Clean(root)) {
		return false
	}
	self, err := os.Executable()
	if err != nil {
		return false
	}
	me, err := os.ReadFile(self)
	if err != nil {
		return false
	}
	dst := filepath.Join(root, modsExe)
	if cur, err := os.ReadFile(dst); err == nil && bytes.Equal(cur, me) {
		return false
	}
	if os.WriteFile(dst+".tmp", me, 0o755) != nil || os.Rename(dst+".tmp", dst) != nil {
		os.Remove(dst + ".tmp")
		return false
	}
	os.WriteFile(filepath.Join(root, playBat), []byte(batScript), 0o644)
	return true
}
