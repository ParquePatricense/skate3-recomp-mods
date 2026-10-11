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
	cmd := exec.Command(filepath.Join(root, "skate3"))
	cmd.Dir = root
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func systemSpanish() bool { return envSpanish() }

func createShortcut(string) error { return errors.New("only on Windows") }
func removeShortcut() error       { return nil }
