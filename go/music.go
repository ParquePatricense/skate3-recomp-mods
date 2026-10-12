package main

// "Tu música": the helper plays a folder of the user's songs while the game
// runs, like the Xbox 360's custom soundtracks. The game's own music files
// can't be replaced yet (they are XMA inside RenderWare Audio .mus archives
// and there is no free XMA encoder), so the in-game music volume goes to 0
// and this plays on top. Settings live in skate3-mods.ini:
//   music_on, music_dir, music_shuffle, music_vol (0-100), music_focus.

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var musicExts = map[string]bool{".mp3": true, ".wav": true, ".wma": true, ".m4a": true, ".aac": true, ".flac": true}

// musicTracks lists the songs in dir and its subfolders, sorted by path.
func musicTracks(dir string) []string {
	var out []string
	if dir == "" {
		return out
	}
	filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && musicExts[strings.ToLower(filepath.Ext(p))] {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func trackName(p string) string { return strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)) }

func musicVolume(c config) int {
	v, err := strconv.Atoi(c["music_vol"])
	if err != nil || v < 0 || v > 100 {
		return 80
	}
	return v
}

func musicOn(c config) bool { return c.on("music_on") && c["music_dir"] != "" }

// musicFocus: pause the songs while the game isn't in front (default on).
func musicFocus(c config) bool { return c["music_focus"] != "0" }

func musicState(root string) map[string]any {
	c := loadConfig(root)
	tracks := musicTracks(c["music_dir"])
	names := []string{}
	for i, t := range tracks {
		if i == 40 {
			break
		}
		names = append(names, trackName(t))
	}
	return map[string]any{"on": c.on("music_on"), "dir": c["music_dir"], "count": len(tracks), "names": names,
		"shuffle": c["music_shuffle"] != "0", "vol": musicVolume(c), "focus": musicFocus(c)}
}

// playOrder is the order the helper plays the songs in.
func playOrder(c config) []string {
	tracks := musicTracks(c["music_dir"])
	if c["music_shuffle"] != "0" {
		rand.Shuffle(len(tracks), func(i, j int) { tracks[i], tracks[j] = tracks[j], tracks[i] })
	}
	return tracks
}

// musicDo handles the "music" actions of the window: on/off, dir=, shuffle=,
// vol=, focus=.
func musicDo(root, arg string) (string, error) {
	c := loadConfig(root)
	key, val, _ := strings.Cut(arg, "=")
	switch key {
	case "on", "off":
		if key == "on" && len(musicTracks(c["music_dir"])) == 0 {
			return "", errors.New(T("primero elegí una carpeta con canciones", "pick a folder with songs first"))
		}
		c["music_on"] = map[bool]string{true: "1", false: "0"}[key == "on"]
		if err := c.save(root); err != nil {
			return "", err
		}
		if key == "on" {
			return T("Tu música suena al jugar. Bajá la música del juego a 0 en sus opciones de audio.",
				"Your music plays while you skate. Turn the game's music down to 0 in its audio options."), nil
		}
		return T("Vuelve la música del juego (subila en sus opciones de audio)", "The game's music is back (turn it up in its audio options)"), nil
	case "dir":
		d, err := pickFolder(T("Elegí la carpeta con tus canciones (MP3, WAV, WMA, M4A, FLAC)",
			"Pick the folder with your songs (MP3, WAV, WMA, M4A, FLAC)"))
		if err != nil || d == "" {
			return "", err
		}
		n := len(musicTracks(d))
		if n == 0 {
			return "", errors.New(T("esa carpeta no tiene canciones", "that folder has no songs"))
		}
		c["music_dir"] = d
		return fmt.Sprintf(T("%d canciones listas", "%d songs ready"), n), c.save(root)
	case "shuffle", "focus":
		c["music_"+key] = map[bool]string{true: "1", false: "0"}[onArg(val)]
		return "", c.save(root)
	case "vol":
		v, err := strconv.Atoi(val)
		if err != nil || v < 0 || v > 100 {
			return "", errors.New("?")
		}
		c["music_vol"] = strconv.Itoa(v)
		return "", c.save(root)
	case "open":
		if c["music_dir"] == "" {
			return "", nil
		}
		return "", openPath(c["music_dir"])
	}
	return "", errors.New("?")
}
