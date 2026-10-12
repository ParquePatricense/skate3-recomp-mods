package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// S3_TESTGAME=<folder with skate3.exe, game/data/big/fedynamic.big and
// miscboot.big> runs the full mod lifecycle on a scratch copy of it.
func TestModLifecycle(t *testing.T) {
	src := os.Getenv("S3_TESTGAME")
	if src == "" {
		t.Skip("S3_TESTGAME not set")
	}
	root := filepath.Join(t.TempDir(), "game")
	if out, err := exec.Command("cp", "-r", src, root).CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	big := bigDir(root)
	icons, texts := filepath.Join(big, archiveName), filepath.Join(big, textsArchive)
	origIcons, _ := os.ReadFile(icons)
	origTexts, _ := os.ReadFile(texts)

	if err := installIcons("ps5", root); err != nil {
		t.Fatal(err)
	}
	if err := installTexts(root, root); err != nil {
		t.Fatal(err)
	}
	if msgs := checkMods(root); len(msgs) != 0 {
		t.Fatalf("untouched files reported: %v", msgs)
	}

	// touching the patched files (new mtime, same bytes) must not make them "originals"
	time.Sleep(20 * time.Millisecond)
	now := time.Now()
	os.Chtimes(icons, now, now)
	os.Chtimes(texts, now, now)
	if msgs := checkMods(root); len(msgs) != 0 {
		t.Fatalf("touched files reported: %v", msgs)
	}
	if b, _ := os.ReadFile(icons + backupSuffix); !bytes.Equal(b, origIcons) {
		t.Fatal("backup overwritten by the patched file")
	}

	// a recomp update puts fresh originals back
	time.Sleep(20 * time.Millisecond)
	os.WriteFile(icons, origIcons, 0o644)
	os.WriteFile(texts, origTexts, 0o644)
	msgs := checkMods(root)
	t.Log(msgs)
	if iconsKey(root) != "ps5" || !textsOn(root) {
		t.Fatal("mods not on after update")
	}
	if cur, _ := os.ReadFile(icons); bytes.Equal(cur, origIcons) {
		t.Fatal("icons not re-applied")
	}

	// a broken file comes back from the backup
	os.WriteFile(texts, []byte("garbage"), 0o644)
	msgs = checkMods(root)
	t.Log(msgs)
	if !archiveHealthy(texts, textsEntry) || !textsOn(root) {
		t.Fatal("texts not repaired")
	}
	t.Log(repairMods(root))

	t.Log(uninstallAll(root))
	if cur, _ := os.ReadFile(icons); !bytes.Equal(cur, origIcons) {
		t.Fatal("icons not restored")
	}
	if cur, _ := os.ReadFile(texts); !bytes.Equal(cur, origTexts) {
		t.Fatal("texts not restored")
	}
	if exists(filepath.Join(root, modsIni)) {
		t.Fatal("ini left behind")
	}
}

func TestAPI(t *testing.T) {
	src := os.Getenv("S3_TESTGAME")
	if src == "" {
		t.Skip("S3_TESTGAME not set")
	}
	dir := filepath.Join(t.TempDir(), "game")
	if out, err := exec.Command("cp", "-r", src, dir).CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := ""
	if _, err := apiDo(&root, "icons", "ps4"); err == nil {
		t.Fatal("no root should fail")
	}
	root = dir
	for _, c := range [][2]string{{"icons", "ps4"}, {"preset", "max"}, {"fov", "75"}, {"ultrawide", "on"},
		{"feature", "priority=on"}, {"feature", "altf4=off"}, {"arg", "on"}} {
		if msg, err := apiDo(&root, c[0], c[1]); err != nil {
			t.Fatal(c, err)
		} else {
			t.Log(c, msg)
		}
	}
	st := apiState(root)
	if st["icons"] != "ps4" || st["preset"] != "max" || st["fov"] != "75.0" || st["ultrawide"] != true || st["arg"] != true {
		t.Fatalf("state: %v", st)
	}
	c := loadConfig(root)
	if !c.feature("priority") || c.feature("altf4") || !c.feature("restart") {
		t.Fatal("features", c)
	}
	toml, _ := os.ReadFile(filepath.Join(root, tomlName))
	t.Log(string(toml))
}
