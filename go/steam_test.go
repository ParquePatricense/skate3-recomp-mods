package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSteamShortcuts(t *testing.T) {
	// one existing non-Steam game, as Steam writes it
	other := &vdfNode{kind: 0, key: "0", kids: []*vdfNode{
		{kind: 2, key: "appid", num: 0x81234567}, {kind: 1, key: "AppName", str: "Otro"},
		{kind: 1, key: "Exe", str: `"C:\otro.exe"`}, {kind: 0, key: "tags", kids: []*vdfNode{{kind: 1, key: "0", str: "fav"}}}}}
	orig := vdfBytes(&vdfNode{kids: []*vdfNode{{kind: 0, key: "shortcuts", kids: []*vdfNode{other}}}})
	dir := t.TempDir()
	p := filepath.Join(dir, "config", "shortcuts.vdf")
	os.MkdirAll(filepath.Dir(p), 0o755)
	os.WriteFile(p, orig, 0o644)

	for i := 0; i < 2; i++ { // the second add changes nothing
		changed, err := steamShortcuts(p, `C:\S3\Skate3-Mods.exe`, `C:\S3\skate3.exe`, `C:\S3`, true)
		if err != nil || changed != (i == 0) {
			t.Fatalf("add %d: changed=%v err=%v", i, changed, err)
		}
	}
	b, _ := os.ReadFile(p)
	doc, err := vdfParse(b)
	if err != nil {
		t.Fatal(err)
	}
	list := doc.get("shortcuts").kids
	if len(list) != 2 || list[1].key != "1" || list[1].text("AppName") != "Skate 3" ||
		list[1].text("Exe") != `"C:\S3\Skate3-Mods.exe"` || list[1].text("LaunchOptions") != "play" {
		t.Fatalf("bad entry: %+v", list)
	}
	if id := list[1].get("appid").num; id&0x80000000 == 0 {
		t.Fatalf("appid %x", id)
	}
	if !bytes.Equal(vdfBytes(doc), b) {
		t.Fatal("parse/write roundtrip changed the file")
	}
	if changed, err := steamShortcuts(p, `C:\S3\Skate3-Mods.exe`, "", `C:\S3`, false); !changed || err != nil {
		t.Fatal(changed, err)
	}
	if b, _ := os.ReadFile(p); !bytes.Equal(b, orig) {
		t.Fatal("remove did not give the original file back")
	}
	// no file yet (a profile that never had non-Steam games)
	p2 := filepath.Join(dir, "x", "config", "shortcuts.vdf")
	if changed, err := steamShortcuts(p2, `C:\S3\Skate3-Mods.exe`, "", `C:\S3`, true); !changed || err != nil {
		t.Fatal(changed, err)
	}
}
