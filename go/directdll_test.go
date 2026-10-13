package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectDLL(t *testing.T) {
	root := t.TempDir()
	toml := filepath.Join(root, tomlName)
	dll := filepath.Join(root, directDLLName)
	set := func(on bool) {
		v := "false"
		if on {
			v = "true"
		}
		os.WriteFile(toml, []byte("skate3_demo_path = "+v+"\nskate3_demo_path_signed_in = "+v+"\n"), 0o644)
	}
	if len(directDLL) < 4096 || !bytes.Contains(directDLL, directDLLMark) {
		t.Fatal("embedded dll missing or unmarked")
	}
	set(true)
	if err := syncDirectDLL(root); err != nil || directDLLState(root) != "ours" {
		t.Fatal("not installed", err)
	}
	os.WriteFile(dll, append([]byte("MZ old "), directDLLMark...), 0o644)
	if syncDirectDLL(root); directDLLState(root) != "ours" {
		t.Fatal("old version not replaced")
	}
	set(false)
	if syncDirectDLL(root); directDLLState(root) != "" {
		t.Fatal("not removed")
	}
	os.WriteFile(dll, []byte("MZ someone else's xinput"), 0o644)
	set(true)
	if syncDirectDLL(root); directDLLState(root) != "foreign" {
		t.Fatal("foreign dll touched")
	}
	set(false)
	if syncDirectDLL(root); directDLLState(root) != "foreign" {
		t.Fatal("foreign dll removed")
	}
	if removeDirectDLL(root); directDLLState(root) != "foreign" {
		t.Fatal("uninstall removed a foreign dll")
	}
}
