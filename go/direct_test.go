package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Straight-to-save puts the chosen language in the English slot (the one the
// boot automation accepts) and everything comes back when it is turned off.
func TestDirectLanguage(t *testing.T) {
	src := os.Getenv("S3_TESTGAME")
	if src == "" {
		t.Skip("S3_TESTGAME not set")
	}
	root := filepath.Join(t.TempDir(), "game")
	if out, err := exec.Command("cp", "-r", src, root).CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	t.Setenv("HOME", t.TempDir())
	texts := filepath.Join(bigDir(root), textsArchive)
	orig, _ := os.ReadFile(texts)
	en := func() []byte { b, _ := readEntry(texts, "LANGUAGE_English_Global_skate3ng.BIN"); return b }
	origEN, origES := en(), func() []byte { b, _ := readEntry(texts, textsEntry); return b }()

	if _, err := setDirect(root, true); err != nil {
		t.Fatal(err)
	}
	if loadConfig(root)["boot_lang"] != "en" && loadConfig(root)["boot_lang"] != "es" {
		t.Fatal(loadConfig(root)["boot_lang"])
	}
	if _, err := setBootLang(root, "es"); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(en(), origES) || readToml(root)["user_language"] != "1" {
		t.Fatal("English slot is not Spanish")
	}
	if _, err := setBootLang(root, "arg"); err != nil {
		t.Fatal(err)
	}
	ar, _, _ := rioplatensize(origES)
	if !bytes.Equal(en(), ar) || !textsOn(root) || !textsPatched(root) {
		t.Fatal("English slot is not ARG")
	}
	setTomlValues(root, [][2]string{{"skate3_demo_path_gameplay_inputs", `"start,a"`}}) // a spot set by an older version
	c := loadConfig(root)
	c["spot"] = "pcu"
	c.save(root)
	if dropSpot(root); spotOn(root) || loadConfig(root)["spot"] != "" || readToml(root)["log_flush_interval"] != "1" {
		t.Fatal("old spot not dropped")
	}
	if msgs := checkMods(root); len(msgs) != 0 {
		t.Fatal(msgs)
	}
	if err := uninstallTexts(root, root); err != nil {
		t.Fatal(err)
	}
	if loadConfig(root)["boot_lang"] != "es" || !bytes.Equal(en(), origES) {
		t.Fatal("ARG off should fall back to Spanish")
	}
	if _, err := setDirect(root, false); err != nil {
		t.Fatal(err)
	}
	if now, _ := os.ReadFile(texts); !bytes.Equal(now, orig) || !bytes.Equal(en(), origEN) {
		t.Fatal("miscboot.big not restored")
	}
	if controlsState(root)["direct"] != false {
		t.Fatal("still direct")
	}
}
