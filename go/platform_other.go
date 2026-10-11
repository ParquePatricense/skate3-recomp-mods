//go:build !windows

package main

import (
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
