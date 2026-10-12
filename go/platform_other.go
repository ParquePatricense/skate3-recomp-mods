//go:build !windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// play just launches the game on other platforms (Alt+F4 is Windows-only).
func play(root string) error {
	checkMods(root)
	cmd := exec.Command(filepath.Join(root, "skate3"))
	cmd.Dir = root
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func systemSpanish() bool { return envSpanish() }

func createShortcut(string) error { return errors.New("only on Windows") }
func removeShortcut() error       { return nil }
func openLog(string) error        { return errors.New("only on Windows") }

func dndOff(string) {}

func documentsDir() string {
	h, _ := os.UserHomeDir()
	return filepath.Join(h, "Documents")
}

func desktopDir() string {
	h, _ := os.UserHomeDir()
	if d := filepath.Join(h, "Desktop"); exists(d) {
		return d
	}
	return h
}

func openPath(p string) error { return exec.Command("xdg-open", p).Start() }

func systemInfo() string { return "" }

func goos() string { return "other" }

func startPlay(root string) error {
	go play(root)
	return nil
}

func pickFolder(string) (string, error) {
	return "", errors.New("folder picker only on Windows")
}

func monitors() []screen { return nil }

func copyToClipboard(string) error { return errors.New("only on Windows") }

func steamDir() string   { return "" }
func steamRunning() bool { return false }
