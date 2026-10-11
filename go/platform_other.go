//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type gameMem struct{}

func (gameMem) regions() []memRegion      { return nil }
func (gameMem) read(uint32, []byte) error { return errors.New("no soportado") }
func (gameMem) close()                    {}
func openGame() (*gameMem, error) {
	return nil, errors.New("el velocimetro y Alt+F4 solo funcionan en Windows")
}

func consoleUI() calibUI {
	return calibUI{say: func(s string) { fmt.Println(s) }, beep: func(int) { fmt.Print("\a") }, sleep: time.Sleep}
}

// play just launches the game on other platforms.
func play(root string) error {
	cmd := exec.Command(filepath.Join(root, "skate3"))
	cmd.Dir = root
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}
