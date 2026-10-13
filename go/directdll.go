package main

// The direct-load patch (skate3-direct, source in dll/): an
// xinput9_1_0.dll next to skate3.exe. The recomp loads it at start; it hands
// the controller to Windows' own xinput and, only on the skate3.exe it was
// made for, patches the boot inside the game (the EA logo ends at once) and
// writes skate3-direct.log with the time of every step. It goes in with
// "Straight to your save" and comes out with it.

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed payload/xinput9_1_0.dll
var directDLL []byte

const (
	directDLLName = "xinput9_1_0.dll"
	directLogName = "skate3-direct.log"
)

// directDLLMark is an export only our dll has.
var directDLLMark = []byte("skate3_direct_version")

// directDLLState: "" (none), "ours", "old" (ours, another version) or
// "foreign" (someone else's dll with that name, never touched).
func directDLLState(root string) string {
	b, err := os.ReadFile(filepath.Join(root, directDLLName))
	switch {
	case err != nil:
		return ""
	case !bytes.Contains(b, directDLLMark):
		return "foreign"
	case !bytes.Equal(b, directDLL):
		return "old"
	}
	return "ours"
}

// syncDirectDLL puts the patch next to skate3.exe while the direct load is
// on and takes it out when it's off. With the game open the file is in use:
// it fails quietly and the next sync (Drop in, opening the modder) does it.
func syncDirectDLL(root string) error {
	p := filepath.Join(root, directDLLName)
	st := directDLLState(root)
	if st == "foreign" {
		return nil
	}
	if !directOn(root) {
		if st != "" {
			return os.Remove(p)
		}
		return nil
	}
	if st == "ours" {
		return nil
	}
	return os.WriteFile(p, directDLL, 0o644)
}

// removeDirectDLL takes our dll out whatever the setting (uninstall).
func removeDirectDLL(root string) error {
	if st := directDLLState(root); st == "ours" || st == "old" {
		return os.Remove(filepath.Join(root, directDLLName))
	}
	return nil
}
