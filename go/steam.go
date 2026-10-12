package main

// Skate 3 in Steam: a "non-Steam game" entry in every Steam profile's
// shortcuts.vdf, so Big Picture, Steam Input, the overlay and Steam's
// screenshots work with the game. The entry starts the mods helper ("play"),
// which opens the game with the mods. Steam rewrites that file when it
// closes, so this has to happen with Steam closed.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// vdfNode is one key of Steam's binary KeyValues: a string, an int32 or a
// list of children (kept in order).
type vdfNode struct {
	kind byte // 0 = map, 1 = string, 2 = int32
	key  string
	str  string
	num  uint32
	kids []*vdfNode
}

func vdfParse(b []byte) (*vdfNode, error) {
	root := &vdfNode{kind: 0}
	if len(b) == 0 {
		root.kids = []*vdfNode{{kind: 0, key: "shortcuts"}}
		return root, nil
	}
	pos := 0
	cstr := func() (string, error) {
		i := bytes.IndexByte(b[pos:], 0)
		if i < 0 {
			return "", errors.New("shortcuts.vdf: cut string")
		}
		s := string(b[pos : pos+i])
		pos += i + 1
		return s, nil
	}
	var parse func(n *vdfNode) error
	parse = func(n *vdfNode) error {
		for pos < len(b) {
			t := b[pos]
			pos++
			if t == 8 {
				return nil
			}
			k, err := cstr()
			if err != nil {
				return err
			}
			c := &vdfNode{kind: t, key: k}
			switch t {
			case 0:
				if err := parse(c); err != nil {
					return err
				}
			case 1:
				if c.str, err = cstr(); err != nil {
					return err
				}
			case 2:
				if pos+4 > len(b) {
					return errors.New("shortcuts.vdf: cut number")
				}
				c.num = binary.LittleEndian.Uint32(b[pos:])
				pos += 4
			default:
				return errors.New("shortcuts.vdf: unknown type " + strconv.Itoa(int(t)))
			}
			n.kids = append(n.kids, c)
		}
		return nil
	}
	if err := parse(root); err != nil {
		return nil, err
	}
	return root, nil
}

func (n *vdfNode) write(w *bytes.Buffer) {
	for _, c := range n.kids {
		w.WriteByte(c.kind)
		w.WriteString(c.key)
		w.WriteByte(0)
		switch c.kind {
		case 0:
			c.write(w)
			w.WriteByte(8)
		case 1:
			w.WriteString(c.str)
			w.WriteByte(0)
		case 2:
			var x [4]byte
			binary.LittleEndian.PutUint32(x[:], c.num)
			w.Write(x[:])
		}
	}
}

func vdfBytes(root *vdfNode) []byte {
	var w bytes.Buffer
	root.write(&w)
	w.WriteByte(8)
	return w.Bytes()
}

func (n *vdfNode) get(key string) *vdfNode {
	for _, c := range n.kids {
		if strings.EqualFold(c.key, key) {
			return c
		}
	}
	return nil
}

func (n *vdfNode) text(key string) string {
	if c := n.get(key); c != nil {
		return c.str
	}
	return ""
}

// steamAppID is the id Steam gives a non-Steam game (also used for artwork).
func steamAppID(exe, name string) uint32 {
	return crc32.ChecksumIEEE([]byte(exe+name)) | 0x80000000
}

// steamShortcuts adds (or, with add false, removes) the Skate 3 entry in one
// shortcuts.vdf. It reports whether the file changed.
func steamShortcuts(path, modsExePath, gameExePath, root string, add bool) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	doc, err := vdfParse(b)
	if err != nil {
		return false, err
	}
	list := doc.get("shortcuts")
	if list == nil {
		list = &vdfNode{kind: 0, key: "shortcuts"}
		doc.kids = append(doc.kids, list)
	}
	exe := `"` + modsExePath + `"`
	kept := []*vdfNode{}
	found := false
	for _, s := range list.kids {
		if strings.EqualFold(s.text("Exe"), exe) || strings.EqualFold(s.text("exe"), exe) {
			found = true
			if !add {
				continue
			}
		}
		kept = append(kept, s)
	}
	if add == found {
		return false, nil
	}
	if add {
		str := func(k, v string) *vdfNode { return &vdfNode{kind: 1, key: k, str: v} }
		num := func(k string, v uint32) *vdfNode { return &vdfNode{kind: 2, key: k, num: v} }
		e := &vdfNode{kind: 0, kids: []*vdfNode{
			num("appid", steamAppID(exe, "Skate 3")), str("AppName", "Skate 3"), str("Exe", exe),
			str("StartDir", `"`+root+`"`), str("icon", gameExePath), str("ShortcutPath", ""),
			str("LaunchOptions", "play"), num("IsHidden", 0), num("AllowDesktopConfig", 1), num("AllowOverlay", 1),
			num("OpenVR", 0), num("Devkit", 0), str("DevkitGameID", ""), num("DevkitOverrideAppID", 0),
			num("LastPlayTime", 0), str("FlatpakAppID", ""), {kind: 0, key: "tags"}}}
		kept = append(kept, e)
	}
	for i, s := range kept {
		s.key = strconv.Itoa(i)
	}
	list.kids = kept
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	if len(b) > 0 {
		os.WriteFile(path+".skate3mods-backup", b, 0o644)
	}
	return true, os.WriteFile(path, vdfBytes(doc), 0o644)
}

// steamProfiles lists every userdata/<id>/config/shortcuts.vdf of a Steam
// install (the file may not exist yet).
func steamProfiles(steamDir string) []string {
	ents, _ := os.ReadDir(filepath.Join(steamDir, "userdata"))
	out := []string{}
	for _, e := range ents {
		if _, err := strconv.ParseUint(e.Name(), 10, 64); err != nil || !e.IsDir() || e.Name() == "0" {
			continue
		}
		out = append(out, filepath.Join(steamDir, "userdata", e.Name(), "config", "shortcuts.vdf"))
	}
	return out
}

// setSteam adds or removes Skate 3 in Steam.
func setSteam(root string, add bool) (string, error) {
	dir := steamDir()
	if dir == "" {
		return "", errors.New(T("no encontré Steam en esta PC", "Steam is not installed on this PC"))
	}
	if steamRunning() {
		return "", errors.New(T("cerrá Steam (Steam > Salir) y tocá de nuevo: Steam pisa su lista de juegos al cerrarse",
			"quit Steam (Steam > Exit) and try again: Steam overwrites its game list when it closes"))
	}
	profiles := steamProfiles(dir)
	if len(profiles) == 0 {
		return "", errors.New(T("Steam no tiene ninguna cuenta iniciada en esta PC", "no Steam account has signed in on this PC"))
	}
	exe := filepath.Join(root, modsExe)
	if !exists(exe) {
		var err error
		if exe, err = os.Executable(); err != nil {
			return "", err
		}
	}
	for _, p := range profiles {
		if _, err := steamShortcuts(p, exe, recompExe(root), root, add); err != nil {
			return "", err
		}
	}
	c := loadConfig(root)
	c["steam"] = map[bool]string{true: "1", false: "0"}[add]
	if err := c.save(root); err != nil {
		return "", err
	}
	if add {
		return T("Skate 3 quedó en tu biblioteca de Steam. Abrí Steam y buscalo en la lista", "Skate 3 is in your Steam library. Open Steam and look for it in the list"), nil
	}
	return T("Skate 3 salió de Steam", "Skate 3 was removed from Steam"), nil
}
