package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestShots reads the location shots from a real game folder, when
// SKATE3_GAME points at one.
func TestShots(t *testing.T) {
	root := os.Getenv("SKATE3_GAME")
	if root == "" {
		t.Skip("SKATE3_GAME not set")
	}
	d, _ := os.ReadFile(filepath.Join(bigDir(root), "fedynamic.big"))
	b, err := parseBig(d)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(shotEntries(b)); n < 40 {
		t.Fatalf("only %d shots", n)
	}
	s, err := pickShot(root, "")
	if err != nil || s.w != 512 || s.h != 256 {
		t.Fatal(s, err)
	}
	if out := os.Getenv("SKATE3_SHOT_PNG"); out != "" {
		im := image.NewRGBA(image.Rect(0, 0, s.w, s.h))
		for i := 0; i < s.w*s.h; i++ {
			im.Set(i%s.w, i/s.w, color.RGBA{s.bgra[i*4+2], s.bgra[i*4+1], s.bgra[i*4], 255})
		}
		f, _ := os.Create(out)
		png.Encode(f, im)
		f.Close()
		t.Log(shotTitle(s.name))
	}
}

func TestShotTitle(t *testing.T) {
	if g := shotTitle("map_skatepark_downtown_full.rx2"); g != "Skatepark Downtown Full" {
		t.Fatal(g)
	}
}

func TestGameDecks(t *testing.T) {
	root := os.Getenv("SKATE3_GAME")
	if root == "" {
		t.Skip("SKATE3_GAME not set")
	}
	ds, err := gameDecks(root)
	if err != nil || len(ds) < 100 {
		t.Fatal(len(ds), err)
	}
	if out := os.Getenv("SKATE3_DECKS_HTML"); out != "" {
		h := "<body style='background:#333'>"
		for _, d := range ds[:40] {
			h += "<img title='" + d.Name + "' src='" + d.Img + "' style='height:200px;margin:4px'>"
		}
		os.WriteFile(out, []byte(h), 0o644)
	}
}

func TestGameAvatars(t *testing.T) {
	root := os.Getenv("SKATE3_GAME")
	if root == "" {
		t.Skip("SKATE3_GAME not set")
	}
	as, err := gameAvatars(root)
	if err != nil || len(as) < 40 {
		t.Fatal(len(as), err)
	}
	if out := os.Getenv("SKATE3_AVATARS_HTML"); out != "" {
		h := "<body style='background:#333'>"
		for _, a := range as {
			h += "<img title='" + a.Name + "' src='" + a.Img + "' style='height:120px;margin:4px'>"
		}
		os.WriteFile(out, []byte(h), 0o644)
	}
}
