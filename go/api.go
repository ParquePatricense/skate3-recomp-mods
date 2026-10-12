// The modder's actions, shared by the window (gui) and usable from tests.
// Every action takes the game folder and returns a short message.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var apiMu sync.Mutex

func iconsKey(root string) string {
	st, err := os.ReadFile(filepath.Join(bigDir(root), archiveName+stateSuffix))
	if err == nil && exists(filepath.Join(bigDir(root), archiveName+backupSuffix)) {
		return strings.TrimSpace(string(st))
	}
	return "xbox"
}

// apiState is everything the window shows.
func apiState(root string) map[string]any {
	st := map[string]any{"version": version, "root": root, "en": langEN, "os": goos()}
	if root == "" {
		return st
	}
	c := loadConfig(root)
	feats := []map[string]any{}
	for _, f := range helperFeatures {
		feats = append(feats, map[string]any{"key": f.key, "on": c.feature(f.key),
			"name": T(f.nameES, f.nameEN), "desc": T(f.descES, f.descEN)})
	}
	ps := []map[string]any{}
	for _, p := range presets {
		ps = append(ps, map[string]any{"key": p.key, "name": p.label()})
	}
	fov := c["fov"]
	if fov == "" {
		fov = "60"
	}
	st["icons"] = iconsKey(root)
	st["preset"] = c["preset"]
	st["presets"] = ps
	st["fov"] = fov
	st["ultrawide"] = c["ultrawide"] == "1"
	st["features"] = feats
	st["shortcut"] = c.on("shortcut")
	st["arg"] = textsOn(root)
	st["saves"] = savesDir(root)
	st["tomlBackup"] = exists(filepath.Join(root, tomlBackup))
	st["radio"] = c["radio"]
	return st
}

func onArg(arg string) bool { return arg == "on" || arg == "1" || arg == "true" }

// apiDo runs one action. root may change ("folder" / "locate").
func apiDo(root *string, cmd, arg string) (string, error) {
	apiMu.Lock()
	defer apiMu.Unlock()
	r := *root
	if r == "" && cmd != "locate" && cmd != "pickfolder" && cmd != "uilang" {
		return "", errors.New(T("primero elegí la carpeta del juego", "pick the game folder first"))
	}
	switch cmd {
	case "locate":
		d, ok := locateGame()
		if !ok {
			return "", errors.New(T("no encontré skate3recomp en la PC", "could not find skate3recomp on this PC"))
		}
		return useRoot(root, d)
	case "pickfolder":
		d, err := pickFolder(T("Elegí la carpeta de skate3recomp (la que tiene skate3.exe)",
			"Pick the skate3recomp folder (the one with skate3.exe)"))
		if err != nil || d == "" {
			return "", err
		}
		if st, e := os.Stat(d); e == nil && !st.IsDir() {
			d = filepath.Dir(d)
		}
		if !isGameRoot(d) {
			return "", errors.New(T("esa carpeta no tiene skate3.exe y game\\data\\big", "that folder has no skate3.exe and game\\data\\big"))
		}
		rememberRoot(d)
		return useRoot(root, d)
	case "uilang":
		if r != "" {
			c := loadConfig(r)
			c["lang"] = arg
			if err := c.save(r); err != nil {
				return "", err
			}
		}
		langEN = arg == "en"
		return "", nil
	case "icons":
		if arg == "xbox" {
			return T("Iconos originales de Xbox", "Original Xbox icons"), uninstallIcons(r)
		}
		return fmt.Sprintf(T("Iconos de %s instalados", "%s icons installed"), styleNames[arg]), installIcons(arg, r)
	case "preset":
		if arg == "restore" {
			return T("Tu configuración de antes volvió", "Your previous settings are back"), restoreToml(r)
		}
		p := findPreset(arg)
		if p == nil {
			return "", errors.New("preset?")
		}
		return p.label(), applyPreset(r, arg)
	case "fov":
		v, err := strconv.ParseFloat(arg, 64)
		if err != nil || v < 40 || v > 120 {
			return "", errors.New(T("FOV inválido", "invalid FOV"))
		}
		s := strconv.FormatFloat(v, 'f', 1, 64)
		if err := setTomlValues(r, [][2]string{{"skate3_field_of_view", s}}); err != nil {
			return "", err
		}
		c := loadConfig(r)
		c["fov"] = s
		return "FOV " + s + "°", c.save(r)
	case "ultrawide":
		on := onArg(arg)
		if err := setTomlValues(r, [][2]string{{"skate3_ultrawide", strconv.FormatBool(on)}}); err != nil {
			return "", err
		}
		c := loadConfig(r)
		c["ultrawide"] = map[bool]string{true: "1", false: "0"}[on]
		return "Ultrawide " + onOff(on), c.save(r)
	case "feature":
		key, val, _ := strings.Cut(arg, "=")
		for _, f := range helperFeatures {
			if f.key == key {
				return T(f.nameES, f.nameEN) + ": " + onOff(onArg(val)), setFeature(r, key, onArg(val))
			}
		}
		return "", errors.New("?")
	case "shortcut":
		on := onArg(arg)
		var err error
		if on {
			err = createShortcut(r)
		} else {
			err = removeShortcut()
		}
		if err != nil {
			return "", err
		}
		c := loadConfig(r)
		c["shortcut"] = map[bool]string{true: "1", false: "0"}[on]
		return T("Acceso directo", "Shortcut") + ": " + onOff(on), c.save(r)
	case "arg":
		if onArg(arg) {
			return T("Español ARG listo: elegilo en la pantalla de idiomas del juego", "Español ARG ready: pick it on the game's language screen"), installTexts(r, r)
		}
		return T("Volvió el neerlandés", "Dutch is back"), uninstallTexts(r, r)
	case "repair":
		return strings.Join(repairMods(r), "\n"), nil
	case "uninstall":
		out := uninstallAll(r)
		return strings.Join(out, "\n"), nil
	case "diagnose":
		p, err := diagnose(r)
		return T("Reporte guardado en ", "Report saved to ") + p, err
	case "folder":
		return "", openFolder(r, arg)
	case "update":
		if err := selfUpdate(); err != nil {
			return "", err
		}
		os.Exit(0)
	case "radio":
		c := loadConfig(r)
		c["radio"] = arg
		return "", c.save(r)
	case "checkupdate":
		return newerVersion(), nil
	case "play":
		return T("¡Drop in!", "Drop in!"), startPlay(r)
	}
	return "", fmt.Errorf("unknown action %q", cmd)
}

func useRoot(root *string, d string) (string, error) {
	r, err := findRoot(d)
	if err != nil {
		return "", err
	}
	*root = r
	setLang(loadConfig(r))
	settleIn(r)
	msgs := checkMods(r)
	return strings.Join(msgs, "\n"), nil
}
