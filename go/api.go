// The modder's actions, shared by the window (gui) and usable from tests.
// Every action takes the game folder and returns a short message.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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
	st := map[string]any{"version": version, "root": root, "en": langEN, "lang": lang, "os": goos()}
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
	st["ctl"] = controlsState(root)
	st["afterPlay"] = c["after_play"]
	st["prefs"] = map[string]string{"deck": c["ui_deck"], "sfx": c["ui_sfx"], "seen": c["ui_seen"], "size": c["ui_size"], "sofa": c["ui_sofa"], "ollie": c["ui_ollie"]}
	st["bootHist"] = bootHistory(root)
	st["langPref"] = c["lang"]
	st["steam"] = c.on("steam")
	st["crash"] = c["crash"]
	st["hasSteam"] = steamDir() != ""
	st["uiWindow"] = c["ui_window"]
	st["uiMonitor"] = uiMonitorSetting()
	st["recomp"] = c["recomp_version"]
	st["play"] = playStats(root, time.Now())
	st["disk"] = diskState(root)
	st["defender"] = c["defender"] == "1"
	st["preload"] = preloadOn()
	st["roots"] = rootList(root)
	st["repoPublic"] = repoPublic
	st["prefs"].(map[string]string)["arrival"] = c["ui_arrival"]
	st["prefs"].(map[string]string)["bg"] = c["ui_bg"]
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
		setLang(config{"lang": arg})
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
	case "direct":
		return setDirect(r, onArg(arg))
	case "report":
		return problemReport(r, arg)
	case "letterbox":
		return T("Bandas negras ", "Black bars ") + onOff(onArg(arg)), setBool(r, "present_letterbox", onArg(arg))
	case "ssr":
		return T("Reflejos ", "Reflections ") + onOff(onArg(arg)), setBool(r, "skate3_native_render_scene_ssr", onArg(arg))
	case "pref":
		key, val, _ := strings.Cut(arg, "=")
		if key != "deck" && key != "sfx" && key != "seen" && key != "size" && key != "sofa" && key != "ollie" && key != "arrival" && key != "bg" {
			return "", errors.New("?")
		}
		c := loadConfig(r)
		c["ui_"+key] = val
		return "", c.save(r)
	case "steam":
		return setSteam(r, onArg(arg))
	case "langshortcut":
		if bootLangSrc(arg) == "" {
			return "", errors.New("?")
		}
		return T("Acceso creado: Skate 3 (directo, ", "Shortcut created: Skate 3 (directo, ") + bootLangName(arg) + ")", createLangShortcut(r, arg, bootLangName(arg))
	case "crashseen":
		c := loadConfig(r)
		delete(c, "crash")
		return "", c.save(r)
	case "gamelog":
		return "", openLog(r)
	case "defender":
		return setDefender(r, onArg(arg))
	case "preload":
		return setPreload(r, onArg(arg))
	case "move":
		return startMove(r, arg)
	case "deleteold":
		return deleteOldGame(r)
	case "useroot":
		if !containsFold(knownRoots(), arg) || !isGameRoot(arg) {
			return "", errors.New("?")
		}
		rememberRoot(arg)
		if _, err := useRoot(root, arg); err != nil {
			return "", err
		}
		return T("Jugás con: ", "Playing with: ") + arg, nil
	case "forgetroot":
		forgetRoot(arg)
		return "", nil
	case "drift": // JSON for the window
		b, err := json.Marshal(map[string]any{"pads": padsConnected(), "drift": stickDrift()})
		return string(b), err
	case "bg":
		return modderBackground(r), nil
	case "issue":
		if !repoIsPublic() {
			return "", errors.New("?")
		}
		return openIssue(r, arg)
	case "bootlang":
		return setBootLang(r, arg)
	case "afterplay":
		c := loadConfig(r)
		c["after_play"] = arg
		return "", c.save(r)
	case "vsync":
		return "VSync " + onOff(onArg(arg)), setBool(r, "vsync", onArg(arg))
	case "fullscreen":
		return T("Pantalla completa ", "Fullscreen ") + onOff(onArg(arg)), setBool(r, "fullscreen", onArg(arg))
	case "monitor":
		return setMonitor(r, arg)
	case "mnk":
		return T("Teclado y mouse ", "Keyboard and mouse ") + onOff(onArg(arg)), setBool(r, "mnk_mode", onArg(arg))
	case "capture":
		return T("Mouse atrapado en el juego ", "Mouse captured by the game ") + onOff(onArg(arg)), setBool(r, "mnk_capture_mouse", onArg(arg))
	case "sens":
		return setSensitivity(r, arg)
	case "bind":
		return setBind(r, arg)
	case "bindsreset":
		return T("Teclas de fábrica", "Default keys"), resetBinds(r)
	case "recompcheck":
		return recompNews(r), nil
	case "recompskip":
		c := loadConfig(r)
		c["recomp_skip"] = arg
		return "", c.save(r)
	case "recompopen":
		return "", openPath(recompReleasesURL())
	case "checkupdate":
		return newerVersion(), nil
	case "play":
		warn, err := preflight(r)
		if err != nil {
			return "", err
		}
		if warn != "" {
			return warn, startPlay(r)
		}
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
	if c := loadConfig(r); c.on("shortcut") && c["shortcut_set"] != "3.5" { // older versions made only "Skate 3"
		if createShortcut(r) == nil {
			c["shortcut_set"] = "3.5"
			c.save(r)
		}
	}
	msgs := checkMods(r)
	dropSpot(r)
	addKnownRoot(r)
	go repoIsPublic()
	warmGameFiles(r)
	return strings.Join(msgs, "\n"), nil
}
