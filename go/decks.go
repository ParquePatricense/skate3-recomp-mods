package main

// The game's own board graphics for the modder's side deck: the 160 deck
// thumbnails of the board shop (skateboard_*.rx2 in fedynamic.big), read
// from your game, stood upright and handed to the window as PNGs.

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type gameDeck struct {
	Key  string `json:"key"`  // "japan"
	Name string `json:"name"` // "Japan"
	Img  string `json:"img"`  // data:image/png;base64,...
}

var (
	deckMu    sync.Mutex
	deckCache = map[string][]gameDeck{}
)

func deckKeyName(file string) string {
	n := strings.TrimSuffix(strings.TrimPrefix(file, "skateboard_skateboard_"), ".rx2")
	return strings.TrimSuffix(n, "_material")
}

// gameDecks lists the game's boards, sorted by name (cached per game folder).
func gameDecks(root string) ([]gameDeck, error) {
	deckMu.Lock()
	defer deckMu.Unlock()
	if d, ok := deckCache[root]; ok {
		return d, nil
	}
	data, err := os.ReadFile(filepath.Join(bigDir(root), "fedynamic.big"))
	if err != nil {
		return nil, err
	}
	b, err := parseBig(data)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []gameDeck
	for i := 0; i < b.count; i++ {
		n := bigName(b, i)
		if !strings.HasPrefix(n, "skateboard_skateboard_") || !strings.HasSuffix(n, ".rx2") || seen[n] {
			continue
		}
		seen[n] = true
		raw, err := readBigEntry(b, i)
		if err != nil {
			continue
		}
		s, err := decodeRx2(raw)
		if err != nil {
			continue
		}
		var buf bytes.Buffer
		if png.Encode(&buf, uprightDeck(s)) != nil {
			continue
		}
		k := deckKeyName(n)
		out = append(out, gameDeck{Key: k, Name: shotTitle(k + ".rx2"), Img: "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	deckCache[root] = out
	return out, nil
}

func gameDeckImg(root, key string) string {
	ds, _ := gameDecks(root)
	for _, d := range ds {
		if d.Key == key {
			return d.Img
		}
	}
	return ""
}

// uprightDeck turns the shop thumbnail (a board lying at 45°) upright,
// cropped to the board.
func uprightDeck(s *shot) image.Image {
	const scale = 1
	cs, sn := math.Cos(math.Pi/4), math.Sin(math.Pi/4)
	diag := int(float64(s.w) * math.Sqrt2)
	W, H := diag*scale/2, diag*scale
	dst := image.NewNRGBA(image.Rect(0, 0, W, H))
	cx, cy := float64(s.w)/2, float64(s.h)/2
	at := func(x, y int) [4]float64 {
		if x < 0 || y < 0 || x >= s.w || y >= s.h {
			return [4]float64{}
		}
		o := (y*s.w + x) * 4
		return [4]float64{float64(s.bgra[o+2]), float64(s.bgra[o+1]), float64(s.bgra[o]), float64(s.bgra[o+3])}
	}
	minX, minY, maxX, maxY := W, H, 0, 0
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			// destination -> source: undo the 45° turn
			dx, dy := (float64(x)+.5)/scale-float64(W)/2/scale, (float64(y)+.5)/scale-float64(H)/2/scale
			sx, sy := cx+dx*cs+dy*sn-.5, cy-dx*sn+dy*cs-.5
			x0, y0 := int(math.Floor(sx)), int(math.Floor(sy))
			fx, fy := sx-float64(x0), sy-float64(y0)
			a, b, c, d := at(x0, y0), at(x0+1, y0), at(x0, y0+1), at(x0+1, y0+1)
			var p [4]float64
			for i := range p {
				p[i] = (a[i]*(1-fx)+b[i]*fx)*(1-fy) + (c[i]*(1-fx)+d[i]*fx)*fy
			}
			if p[3] < 8 {
				continue
			}
			dst.SetNRGBA(x, y, color.NRGBA{uint8(p[0]), uint8(p[1]), uint8(p[2]), uint8(p[3])})
			minX, minY, maxX, maxY = min(minX, x), min(minY, y), max(maxX, x), max(maxY, y)
		}
	}
	if maxX <= minX || maxY <= minY {
		return dst
	}
	return dst.SubImage(image.Rect(minX, minY, maxX+1, maxY+1))
}

// Your skater for the modder's side: the five starter skaters of the
// skater creator (demoskater*.rx2) and the pros' portraits (unlock_pro_*).
var proNames = map[string]string{
	"ambassador": "Ambassador", "andrewreynolds": "Andrew Reynolds", "atiba": "Atiba Jefferson", "bennyfairfax": "Benny Fairfax",
	"braydonszafranski": "Braydon Szafranski", "burnette": "Mike Burnette", "chriscole": "Chris Cole", "chrishaslam": "Chris Haslam",
	"coachfrank": "Coach Frank", "colinmckay": "Colin McKay", "cuzparry": "Cuz Parry", "dandrehobl": "Dan Drehobl", "dannyway": "Danny Way",
	"darrennavarette": "Darren Navarrette", "deadspace": "Isaac (Dead Space)", "deerhead": "Deerhead", "dembonez": "Dem Bonez",
	"dennisbusenitz": "Dennis Busenitz", "erickoston": "Eric Koston", "jasondill": "Jason Dill", "jerryhsu": "Jerry Hsu",
	"joeybrezinski": "Joey Brezinski", "johncardiel": "John Cardiel", "johnrattray": "John Rattray", "joshkalis": "Josh Kalis",
	"lizardking": "Lizard King", "lucaspuig": "Lucas Puig", "markappleyard": "Mark Appleyard", "meatman": "Meat Man", "mike": "Mike",
	"mikecarroll": "Mike Carroll", "patduffy": "Pat Duffy", "pjladd": "PJ Ladd", "raybarbee": "Ray Barbee", "reda": "Reda",
	"reynolds": "Reynolds", "robdyrdek": "Rob Dyrdek", "ryangallant": "Ryan Gallant", "ryansmith": "Ryan Smith", "shingo": "Shingo",
	"slappy": "Slappy", "terrykennedy": "Terry Kennedy",
}

var (
	avatarMu    sync.Mutex
	avatarCache = map[string][]gameDeck{}
)

func gameAvatars(root string) ([]gameDeck, error) {
	avatarMu.Lock()
	defer avatarMu.Unlock()
	if a, ok := avatarCache[root]; ok {
		return a, nil
	}
	data, err := os.ReadFile(filepath.Join(bigDir(root), "fedynamic.big"))
	if err != nil {
		return nil, err
	}
	b, err := parseBig(data)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var starters, pros []gameDeck
	for i := 0; i < b.count; i++ {
		n := bigName(b, i)
		starter := strings.HasPrefix(n, "demoskater")
		if (!starter && !strings.HasPrefix(n, "unlock_pro_")) || !strings.HasSuffix(n, ".rx2") || seen[n] {
			continue
		}
		seen[n] = true
		raw, err := readBigEntry(b, i)
		if err != nil {
			continue
		}
		s, err := decodeRx2(raw)
		if err != nil {
			continue
		}
		var buf bytes.Buffer
		if png.Encode(&buf, cropAlpha(s)) != nil {
			continue
		}
		img := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
		k := strings.TrimSuffix(n, ".rx2")
		if starter {
			starters = append(starters, gameDeck{Key: k, Name: "Skater " + strings.TrimLeft(strings.TrimPrefix(k, "demoskater"), "0"), Img: img})
		} else {
			p := strings.TrimPrefix(k, "unlock_pro_")
			name := proNames[p]
			if name == "" {
				name = shotTitle(p + ".rx2")
			}
			pros = append(pros, gameDeck{Key: k, Name: name, Img: img})
		}
	}
	sort.Slice(starters, func(i, j int) bool { return starters[i].Key < starters[j].Key })
	sort.Slice(pros, func(i, j int) bool { return pros[i].Name < pros[j].Name })
	out := append(starters, pros...)
	avatarCache[root] = out
	return out, nil
}

func gameAvatarImg(root, key string) string {
	as, _ := gameAvatars(root)
	for _, a := range as {
		if a.Key == key {
			return a.Img
		}
	}
	return ""
}

// cropAlpha cuts the picture to its visible part.
func cropAlpha(s *shot) image.Image {
	im := image.NewNRGBA(image.Rect(0, 0, s.w, s.h))
	minX, minY, maxX, maxY := s.w, s.h, -1, -1
	for y := 0; y < s.h; y++ {
		for x := 0; x < s.w; x++ {
			o := (y*s.w + x) * 4
			a := s.bgra[o+3]
			im.SetNRGBA(x, y, color.NRGBA{s.bgra[o+2], s.bgra[o+1], s.bgra[o], a})
			if a > 16 {
				minX, minY, maxX, maxY = min(minX, x), min(minY, y), max(maxX, x), max(maxY, y)
			}
		}
	}
	if maxX < minX {
		return im
	}
	return im.SubImage(image.Rect(minX, minY, maxX+1, maxY+1))
}
