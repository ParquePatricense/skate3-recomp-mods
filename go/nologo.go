package main

// No logo: the EA Black Box movie that opens the game is gone for real
// instead of hidden. The game's fe_movie "ea_intro" record points at the
// file "EA_Blackbox"; a loose movie file is renamed out of the way, one
// inside a BIG archive gets a zero-length directory entry (the original
// 16 bytes are kept in the ini). If the game then fails to reach the title
// screen, the helper puts the logo back on its own.

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	logoMovie  = "ea_blackbox"
	logoOffExt = ".mods-off"
)

// logoFiles finds the loose logo movie files under game\ (on or off).
func logoFiles(root string) (on, off []string) {
	base := filepath.Join(root, "game")
	filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.Count(strings.TrimPrefix(p, base), string(filepath.Separator)) > 5 {
				return filepath.SkipDir
			}
			return nil
		}
		n := strings.ToLower(d.Name())
		if strings.HasSuffix(n, logoOffExt) {
			if movieBase(strings.TrimSuffix(n, logoOffExt)) == logoMovie {
				off = append(off, p)
			}
		} else if movieBase(n) == logoMovie {
			on = append(on, p)
		}
		return nil
	})
	return
}

func movieBase(n string) string { return strings.TrimSuffix(n, filepath.Ext(n)) }

// bigLogoEntry looks for the logo inside the BIG archives, reading only
// their directories (some are hundreds of MB).
func bigLogoEntry(root string) (file string, idx int, ok bool) {
	ms, _ := filepath.Glob(filepath.Join(bigDir(root), "*.big"))
	for _, p := range ms {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		h := make([]byte, 0x30)
		if _, err := f.ReadAt(h, 0); err != nil || h[0] != 'E' || h[1] != 'B' {
			f.Close()
			continue
		}
		count := int(binary.BigEndian.Uint32(h[4:]))
		namesOff := int(binary.BigEndian.Uint32(h[12:]))
		nameLen := int(h[20])
		names := make([]byte, count*nameLen)
		_, err = f.ReadAt(names, int64(namesOff))
		f.Close()
		if err != nil || nameLen < 3 {
			continue
		}
		for i := 0; i < count; i++ {
			nm := names[i*nameLen+2 : (i+1)*nameLen]
			if z := strings.IndexByte(string(nm), 0); z >= 0 {
				nm = nm[:z]
			}
			if movieBase(strings.ToLower(filepath.Base(filepath.FromSlash(strings.ReplaceAll(string(nm), "\\", "/"))))) == logoMovie {
				return p, i, true
			}
		}
	}
	return "", 0, false
}

// noLogo is what the modder shows: "on", "off", "failed" (it came back
// because the game didn't boot without it) or "missing" (no logo movie).
func noLogo(root string) string {
	c := loadConfig(root)
	switch {
	case c.on("nologo"):
		return "on"
	case c.on("nologo_failed"):
		return "failed"
	}
	logoMu.Lock()
	defer logoMu.Unlock()
	if logoFound[root] == "" {
		logoFound[root] = noLogoState(root)
	}
	if logoFound[root] == "missing" {
		return "missing"
	}
	return "off"
}

var (
	logoMu    sync.Mutex
	logoFound = map[string]string{}
)

// noLogoState looks at the files: "on", "off" or "missing".
func noLogoState(root string) string {
	if loadConfig(root)["nologo_big"] != "" {
		return "on"
	}
	on, off := logoFiles(root)
	switch {
	case len(off) > 0 && len(on) == 0:
		return "on"
	case len(on) > 0:
		return "off"
	}
	if _, _, ok := bigLogoEntry(root); ok {
		return "off"
	}
	return "missing"
}

func setNoLogo(root string, on bool) (string, error) {
	c := loadConfig(root)
	if on {
		c["nologo"] = "1"
		delete(c, "nologo_failed")
	} else {
		delete(c, "nologo")
	}
	if err := c.save(root); err != nil {
		return "", err
	}
	if err := applyNoLogo(root, on); err != nil {
		return "", err
	}
	if on {
		return T("Logo de EA Black Box sacado: el juego arranca sin él", "EA Black Box logo removed: the game starts without it"), nil
	}
	return T("Volvió el logo de EA Black Box", "The EA Black Box logo is back"), nil
}

func applyNoLogo(root string, on bool) error {
	lon, loff := logoFiles(root)
	if on {
		for _, p := range lon {
			os.Remove(p + logoOffExt)
			if err := os.Rename(p, p+logoOffExt); err != nil {
				return err
			}
		}
		if len(lon) > 0 || len(loff) > 0 {
			return nil
		}
		if loadConfig(root)["nologo_big"] != "" {
			return nil
		}
		file, idx, ok := bigLogoEntry(root)
		if !ok {
			return errors.New(T("no encontré la película del logo (EA_Blackbox) en la carpeta del juego", "could not find the logo movie (EA_Blackbox) in the game folder"))
		}
		return patchBigEntry(root, file, idx)
	}
	for _, p := range loff {
		orig := strings.TrimSuffix(p, logoOffExt)
		if exists(orig) { // a recomp update brought it back
			os.Remove(p)
			continue
		}
		if err := os.Rename(p, orig); err != nil {
			return err
		}
	}
	return unpatchBigEntry(root)
}

// patchBigEntry zeroes the sizes of one BIG directory entry.
func patchBigEntry(root, file string, idx int) error {
	f, err := os.OpenFile(file, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	e := make([]byte, 16)
	if _, err := f.ReadAt(e, int64(0x30+idx*16)); err != nil {
		return err
	}
	c := loadConfig(root)
	c["nologo_big"] = filepath.Base(file) + "|" + strconv.Itoa(idx) + "|" + hex.EncodeToString(e)
	if err := c.save(root); err != nil {
		return err
	}
	z := append([]byte{}, e...)
	copy(z[4:12], make([]byte, 8))
	_, err = f.WriteAt(z, int64(0x30+idx*16))
	return err
}

func unpatchBigEntry(root string) error {
	c := loadConfig(root)
	v := c["nologo_big"]
	if v == "" {
		return nil
	}
	parts := strings.Split(v, "|")
	if len(parts) != 3 {
		parts = []string{"", "x", ""}
	}
	idx, err := strconv.Atoi(parts[1])
	e, err2 := hex.DecodeString(parts[2])
	if err != nil || err2 != nil || len(e) != 16 {
		delete(c, "nologo_big")
		return c.save(root)
	}
	f, err := os.OpenFile(filepath.Join(bigDir(root), parts[0]), os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	cur := make([]byte, 16)
	if _, err := f.ReadAt(cur, int64(0x30+idx*16)); err == nil && bigEntryZeroed(cur) { // else the archive was replaced
		if _, err := f.WriteAt(e, int64(0x30+idx*16)); err != nil {
			return err
		}
	}
	delete(c, "nologo_big")
	return c.save(root)
}

func bigEntryZeroed(e []byte) bool {
	return binary.BigEndian.Uint64(e[4:12]) == 0
}

// keepNoLogo puts the removal back after a recomp update restored the file
// or replaced the archive.
func keepNoLogo(root string) {
	c := loadConfig(root)
	if !c.on("nologo") {
		return
	}
	if v := c["nologo_big"]; v != "" {
		parts := strings.Split(v, "|")
		idx, _ := strconv.Atoi(parts[min(1, len(parts)-1)])
		cur := make([]byte, 16)
		if f, err := os.Open(filepath.Join(bigDir(root), parts[0])); err == nil {
			_, err = f.ReadAt(cur, int64(0x30+idx*16))
			f.Close()
			if err == nil && !bigEntryZeroed(cur) {
				delete(c, "nologo_big")
				c.save(root)
			}
		}
	}
	applyNoLogo(root, true)
}

// noLogoFailed: the game didn't get past the boot without the logo, so it
// comes back and the switch shows why.
func noLogoFailed(root string) {
	applyNoLogo(root, false)
	c := loadConfig(root)
	delete(c, "nologo")
	c["nologo_failed"] = "1"
	c.save(root)
}
