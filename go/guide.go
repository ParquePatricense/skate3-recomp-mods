package main

// Port Carverton guide: every spot of the game's map with its photo, its
// district and the game's own name and description, in the modder's
// language (from the game's text tables). The spot list, districts and
// text ids come from the game's database (db.big, skatercollections.bin).

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type guideSpot struct {
	key, zone  string
	name, desc uint32 // text hashes
}

var guideSpots = []guideSpot{
	{"campus_entrance", "uni", 0x35b95b7c, 0x87575538},
	{"chancenter", "uni", 0x35b95b60, 0x192d0afa},
	{"clocktower", "uni", 0x35b95b63, 0x9bb124bc},
	{"carvatron", "uni", 0x35b95b7d, 0x3640ea6f},
	{"observatory", "uni", 0x35b95b5c, 0xf2bd26ff},
	{"daly_estates", "uni", 0xbbb8242e, 0x8b0d40e2},
	{"library", "uni", 0x35b95b62, 0xf042e95c},
	{"stadium", "uni", 0x35b95b61, 0x2b74980f},
	{"super_ultramega_park", "uni", 0x35b95b5f, 0xe2e05b9e},
	{"peterson_pavilion", "uni", 0x35b95b64, 0x148f6df1},
	{"memorial", "dt", 0x34890516, 0xc17153ed},
	{"allmart", "dt", 0x34890517, 0xa4532402},
	{"hotel", "dt", 0x34890519, 0x78d020b7},
	{"rippon", "dt", 0x34890531, 0x29bdc6f1},
	{"parkade", "dt", 0x34890533, 0x7c3fcda7},
	{"uptown_promenade", "dt", 0xc5a9a772, 0xf678b75d},
	{"midtown_promenade", "dt", 0xc5a9a773, 0x6e2aee12},
	{"aletown", "dt", 0xc5a9a774, 0xa3d88b4f},
	{"slappys", "dt", 0x34890515, 0xa4128666},
	{"plaza_of_solitude", "dt", 0x34890512, 0xfae93bcb},
	{"kube", "dt", 0x34890513, 0xcdbfb04d},
	{"crystalplaza", "dt", 0x34890532, 0x1a3c9ffb},
	{"rosalita", "dt", 0x34890514, 0x18b2a3fa},
	{"tresoutta", "dt", 0x34890518, 0xc4807818},
	{"tanker", "ind", 0xc6b4e59a, 0xea7d0181},
	{"miracle_bowl", "ind", 0x09ede28b, 0xf1ecdbad},
	{"quarry", "ind", 0xc97071e5, 0xf3b174b4},
	{"old_factory", "ind", 0xee0003b8, 0x23adb793},
	{"factory_roofs", "ind", 0x02ac4ecc, 0x2754f4fd},
	{"spillway_entrance", "ind", 0x512e6116, 0x8a8a0e21},
	{"loading_docks", "ind", 0xb07456c7, 0x5fd0876e},
	{"hastings", "ind", 0x840637a7, 0xb29be144},
	{"drydocks", "ind", 0x380a78c4, 0xef7379df},
	{"2nd_and_navy", "ind", 0x37389a51, 0x34f29c71},
	{"ghetto", "ind", 0x4ad57f92, 0x8083336d},
	{"pipeworks", "ind", 0x43712843, 0x53880160},
	{"maloof", "park", 0x3524344e, 0x9dcf5db2},
	{"monster", "park", 0xac63be06, 0},
	{"skate_school", "park", 0xd84ce903, 0xb1dfff1e},
	{"startpark_progression", "park", 0xe4812491, 0},
	{"blackboxpark", "park", 0xf4e6c017, 0x873449a4},
}

// the game's district names
var guideZones = map[string]uint32{"dt": 0x42b50d70, "ind": 0x673028f2, "uni": 0xedd449d2}

// gameTextTable reads one language table of the game (from the untouched
// copy of miscboot.big when the language mods made one).
func gameTextTable(root, lang string) map[uint32]string {
	p := filepath.Join(bigDir(root), textsArchive)
	d, err := os.ReadFile(p + textsBackup)
	if err != nil {
		if d, err = os.ReadFile(p); err != nil {
			return nil
		}
	}
	b, err := parseBig(d)
	if err != nil {
		return nil
	}
	src := map[string]string{"es": "Spanish", "pt": "Spanish", "fr": "French", "de": "German", "it": "Italian"}[lang]
	if src == "" {
		src = "English"
	}
	i, err := b.find("LANGUAGE_" + src + "_Global_skate3ng.BIN")
	if err != nil {
		return nil
	}
	t, err := b.read(i)
	if err != nil {
		return nil
	}
	switch lang {
	case "pt":
		t, err = portuguese(t)
	case "es": // the modder speaks rioplatense
		t, _, err = rioplatensize(t)
	}
	if err != nil {
		return nil
	}
	out := map[uint32]string{}
	rewriteTable(t, func(h uint32, s string) string { out[h] = s; return s })
	return out
}

// gameText cleans a game string for the modder: "«" is the game's skate
// logo glyph, "^" a line break.
func gameText(s string) string {
	s = strings.ReplaceAll(s, "«", "skate")
	s = strings.ReplaceAll(s, "^", " ")
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

type guideEntry struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Zone string `json:"zone"`
	Area string `json:"area"` // the district's name ("" for parks)
	Desc string `json:"desc"`
	Img  string `json:"img"`
}

var (
	guideMu    sync.Mutex
	guideCache = map[string][]guideEntry{}
	guideImgs  = map[string]map[string]string{}
)

// shotImages: the location photos as small JPEGs, by spot key.
func shotImages(root string) map[string]string {
	if m, ok := guideImgs[root]; ok {
		return m
	}
	m := map[string]string{}
	d, err := os.ReadFile(filepath.Join(bigDir(root), "fedynamic.big"))
	if err != nil {
		return m
	}
	b, err := parseBig(d)
	if err != nil {
		return m
	}
	for _, i := range shotEntries(b) {
		raw, err := readBigEntry(b, i)
		if err != nil {
			continue
		}
		s, err := decodeRx2(raw)
		if err != nil {
			continue
		}
		if u := shotJPEG(s, 2); u != "" {
			m[strings.TrimSuffix(bigName(b, i), ".rx2")] = u
		}
	}
	guideImgs[root] = m
	return m
}

// shotJPEG makes a data URL of a shot, shrunk by div.
func shotJPEG(s *shot, div int) string {
	w, h := s.w/div, s.h/div
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, b int
			for dy := 0; dy < div; dy++ {
				for dx := 0; dx < div; dx++ {
					o := ((y*div+dy)*s.w + x*div + dx) * 4
					b += int(s.bgra[o])
					g += int(s.bgra[o+1])
					r += int(s.bgra[o+2])
				}
			}
			n := div * div
			o := (y*w + x) * 4
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = byte(r/n), byte(g/n), byte(b/n), 255
		}
	}
	var buf bytes.Buffer
	if jpeg.Encode(&buf, img, &jpeg.Options{Quality: 82}) != nil {
		return ""
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// portGuide is the guide in the modder's language.
func portGuide(root string) []guideEntry {
	guideMu.Lock()
	defer guideMu.Unlock()
	if g, ok := guideCache[root+"|"+lang]; ok {
		return g
	}
	txt := gameTextTable(root, lang)
	imgs := shotImages(root)
	var out []guideEntry
	for _, s := range guideSpots {
		e := guideEntry{Key: s.key, Zone: s.zone, Img: imgs[s.key], Name: gameText(txt[s.name]), Desc: gameText(txt[s.desc])}
		if e.Name == "" {
			e.Name = shotTitle(s.key)
		}
		if z, ok := guideZones[s.zone]; ok {
			e.Area = gameText(txt[z])
		}
		out = append(out, e)
	}
	guideCache[root+"|"+lang] = out
	return out
}

// spotName is a spot's name and district in the modder's language.
func spotName(root, key string) (name, area string) {
	txt := gameTextTable(root, lang)
	for _, s := range guideSpots {
		if s.key == key {
			return gameText(txt[s.name]), gameText(txt[guideZones[s.zone]])
		}
	}
	return "", ""
}

// modderBackground is one of the game's photos for behind the modder.
func modderBackground(root string) string {
	guideMu.Lock()
	m := shotImages(root)
	guideMu.Unlock()
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return ""
	}
	return m[keys[rand.Intn(len(keys))]]
}

// trickGroups: the game's own control hints (tutorials, challenges, loading
// tips), with their button tokens (<B:Stick_Right_Down_hud> and the like)
// left for the modder to draw in your icon style.
var trickGroups = []struct {
	key   string
	hints []uint32
}{
	{"basic", []uint32{0xa8a951fe, 0x122a395e, 0x673c6fb4, 0x673c6f91, 0xf6ca8faf, 0xf6ca8fb0, 0xf952c352}},
	{"flip", []uint32{0xdbb8e090, 0xdbb8e092, 0xdbb8e093, 0x22271597, 0x22271598}},
	{"grab", []uint32{0x5ed8b248, 0xd6168cba, 0xa5728277, 0x06569bda, 0x905928c4}},
	{"grind", []uint32{0x50355235, 0x591b2a1e}},
	{"walk", []uint32{0x0b1ff797, 0x0b1ff799, 0x0b1ff79a, 0x0b1ff79b, 0x0b1ff79c, 0xdc32373a}},
	{"bail", []uint32{0x5ed61b90, 0x27ba5063, 0x13fc2543, 0x003dfa24, 0xd8c1a3e6, 0xb1454da8, 0x3b787b83}},
	{"more", []uint32{0x6df445f7, 0xd4d75e64, 0x0ee908eb, 0x0ee908ec, 0x7bba386e, 0xdc32371c, 0xb803b207, 0xec7e3225}},
}

type trickGroup struct {
	Key   string   `json:"key"`
	Hints []string `json:"hints"`
}

func trickList(root string) []trickGroup {
	txt := gameTextTable(root, lang)
	var out []trickGroup
	for _, g := range trickGroups {
		tg := trickGroup{Key: g.key}
		for _, h := range g.hints {
			if s := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(txt[h], "«", "skate"), "^^", "^")); s != "" {
				tg.Hints = append(tg.Hints, s)
			}
		}
		out = append(out, tg)
	}
	return out
}
