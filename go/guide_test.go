package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestGuide(t *testing.T) {
	root := os.Getenv("SKATE3_GAME")
	if root == "" {
		t.Skip()
	}
	for _, l := range []string{"es", "en", "pt", "fr"} {
		lang = l
		g := portGuide(root)
		miss := 0
		for _, e := range g {
			if e.Img == "" {
				miss++
			}
		}
		if miss > 0 || len(g) != len(guideSpots) {
			t.Fatalf("%s: %d spots, %d without photo", l, len(g), miss)
		}
		t.Logf("%s: %s / %s / %s", l, g[6].Name, g[6].Area, g[6].Desc)
		for _, e := range g {
			if strings.Contains(e.Name, "«") || e.Name == "" {
				t.Fatal(e.Name)
			}
		}
	}
	lang = "es"
}

func TestTricks(t *testing.T) {
	root := os.Getenv("SKATE3_GAME")
	if root == "" {
		t.Skip()
	}
	lang = "es"
	g := trickList(root)
	n := 0
	for _, x := range g {
		n += len(x.Hints)
	}
	if n != 40 {
		t.Fatal(n)
	}
	t.Log(g[1].Hints)
}

func TestDumpGuideJSON(t *testing.T) {
	root, out := os.Getenv("SKATE3_GAME"), os.Getenv("GUIDE_OUT")
	if root == "" || out == "" {
		t.Skip()
	}
	lang = "es"
	g, _ := json.Marshal(portGuide(root))
	tr, _ := json.Marshal(trickList(root))
	os.WriteFile(out+"/guide.json", g, 0o644)
	os.WriteFile(out+"/tricks.json", tr, 0o644)
	os.WriteFile(out+"/bg.txt", []byte(modderBackground(root)), 0o644)
}
