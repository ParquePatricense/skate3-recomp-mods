package main

import (
	"os"
	"testing"
)

func TestPortuguese(t *testing.T) {
	if n := len(ptTexts()); n < 13000 {
		t.Fatalf("only %d texts", n)
	}
	es, err := os.ReadFile("/tmp/claude-0/-home-claude/223a11ea-0269-5735-8ab9-01f7541a17a4/scratchpad/dump/miscboot/LANGUAGE_Spanish_Global_skate3ng.BIN")
	if err != nil {
		t.Skip()
	}
	pt, err := portuguese(es)
	if err != nil {
		t.Fatal(err)
	}
	got := map[uint32]string{}
	if _, _, err := rewriteTable(pt, func(h uint32, s string) string { got[h] = s; return s }); err != nil {
		t.Fatal(err)
	}
	if got[langNameEnglish] != "Português" || got[0x35b95b62] != ptTexts()[0x35b95b62] {
		t.Fatal(got[langNameEnglish], got[0x35b95b62])
	}
	t.Log(got[0x35b95b62])
}
