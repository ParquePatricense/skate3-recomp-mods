package main

// Display and controls: VSync, fullscreen, the screen the game opens on, and
// keyboard + mouse play (the recomp's mnk_* and keybind_* settings in
// skate3.toml). Keybinds only work with mnk_mode on.

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// readToml reads the flat "key = value" lines of skate3.toml (quotes removed).
func readToml(root string) map[string]string {
	m := map[string]string{}
	f, err := os.Open(filepath.Join(root, tomlName))
	if err != nil {
		return m
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' || line[0] == '[' {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, `"`) {
			if end := strings.Index(v[1:], `"`); end >= 0 {
				v = v[1 : end+1]
			}
		} else if i := strings.Index(v, "#"); i >= 0 {
			v = strings.TrimSpace(v[:i])
		}
		m[strings.TrimSpace(k)] = v
	}
	return m
}

func tomlBool(m map[string]string, key string, def bool) bool {
	if v, ok := m[key]; ok {
		return v == "true"
	}
	return def
}

// keybind is one pad button the keyboard can press: the recomp's default key
// and its name on Xbox and PlayStation pads.
type keybind struct{ key, def, xb, ps string }

var keybinds = []keybind{
	{"keybind_a", "Space", "A", "✕"},
	{"keybind_b", "C", "B", "○"},
	{"keybind_x", "E", "X", "□"},
	{"keybind_y", "F", "Y", "△"},
	{"keybind_left_trigger", "RMB", "LT", "L2"},
	{"keybind_right_trigger", "LMB", "RT", "R2"},
	{"keybind_left_shoulder", "Q", "LB", "L1"},
	{"keybind_right_shoulder", "R", "RB", "R1"},
	{"keybind_lstick_up", "W", "L↑", "L↑"},
	{"keybind_lstick_down", "S", "L↓", "L↓"},
	{"keybind_lstick_left", "A", "L←", "L←"},
	{"keybind_lstick_right", "D", "L→", "L→"},
	{"keybind_lstick_press", "Shift", "LS", "L3"},
	{"keybind_rstick_press", "MMB", "RS", "R3"},
	{"keybind_dpad_up", "Up", "D↑", "D↑"},
	{"keybind_dpad_down", "Down", "D↓", "D↓"},
	{"keybind_dpad_left", "Left", "D←", "D←"},
	{"keybind_dpad_right", "Right", "D→", "D→"},
	{"keybind_back", "Tab", "Back", "Select"},
	{"keybind_start", "Return", "Start", "Options"},
}

// keyNames are the names the recomp understands (sdk/src/ui/keybinds.cpp).
var keyNames = func() map[string]bool {
	m := map[string]bool{}
	for _, n := range strings.Fields(`Backtick Minus Plus Comma Period Semicolon Slash Backslash
		LBracket RBracket Quote Escape Return Space Tab Backspace Delete Insert Home End PageUp
		PageDown Left Right Up Down Shift Control Alt NumpadEnter NumpadPlus NumpadMinus NumpadStar
		NumpadSlash PrintScreen Pause CapsLock NumLock ScrollLock LMB RMB MMB`) {
		m[n] = true
	}
	for c := 'A'; c <= 'Z'; c++ {
		m[string(c)] = true
	}
	for i := 0; i <= 9; i++ {
		m[strconv.Itoa(i)] = true
		m["Numpad"+strconv.Itoa(i)] = true
	}
	for i := 1; i <= 24; i++ {
		m["F"+strconv.Itoa(i)] = true
	}
	return m
}()

func findBind(key string) *keybind {
	for i := range keybinds {
		if keybinds[i].key == key {
			return &keybinds[i]
		}
	}
	return nil
}

func controlsState(root string) map[string]any {
	m := readToml(root)
	binds := []map[string]any{}
	for _, b := range keybinds {
		v, ok := m[b.key]
		if !ok {
			v = b.def
		}
		binds = append(binds, map[string]any{"key": b.key, "val": v, "def": b.def, "xb": b.xb, "ps": b.ps})
	}
	sens := 1.0
	if s, err := strconv.ParseFloat(m["mnk_sensitivity"], 64); err == nil {
		sens = s
	}
	mon, _ := strconv.Atoi(m["monitor"])
	return map[string]any{
		"vsync": tomlBool(m, "vsync", false), "fullscreen": tomlBool(m, "fullscreen", true),
		"monitor": mon, "monitors": monitors(),
		"mnk": tomlBool(m, "mnk_mode", false), "capture": tomlBool(m, "mnk_capture_mouse", false),
		"sens": sens, "binds": binds,
	}
}

func setBool(root, key string, on bool) error {
	return setTomlValues(root, [][2]string{{key, strconv.FormatBool(on)}})
}

// setBind writes "keybind_x = "Name"" (an empty name leaves the button free).
func setBind(root, arg string) (string, error) {
	key, name, _ := strings.Cut(arg, "=")
	b := findBind(key)
	if b == nil || (name != "" && !keyNames[name]) {
		return "", errors.New(T("tecla no válida", "invalid key"))
	}
	if err := setTomlValues(root, [][2]string{{key, strconv.Quote(name)}}); err != nil {
		return "", err
	}
	if name == "" {
		name = T("nada", "nothing")
	}
	return fmt.Sprintf("%s / %s → %s", b.xb, b.ps, name), nil
}

func resetBinds(root string) error {
	vals := [][2]string{}
	for _, b := range keybinds {
		vals = append(vals, [2]string{b.key, strconv.Quote(b.def)})
	}
	return setTomlValues(root, vals)
}

func setMonitor(root, arg string) (string, error) {
	n, err := strconv.Atoi(arg)
	if err != nil || n < 0 || n > 16 {
		return "", errors.New(T("pantalla no válida", "invalid screen"))
	}
	if err := setTomlValues(root, [][2]string{{"monitor", strconv.Itoa(n)}}); err != nil {
		return "", err
	}
	if n == 0 {
		return T("El juego abre en la pantalla principal", "The game opens on the main screen"), nil
	}
	return fmt.Sprintf(T("El juego abre en la pantalla %d", "The game opens on screen %d"), n), nil
}

func setSensitivity(root, arg string) (string, error) {
	v, err := strconv.ParseFloat(arg, 64)
	if err != nil || v < 0.05 || v > 10 {
		return "", errors.New(T("sensibilidad no válida", "invalid sensitivity"))
	}
	s := strconv.FormatFloat(v, 'f', 2, 64)
	return T("Sensibilidad ", "Sensitivity ") + s, setTomlValues(root, [][2]string{{"mnk_sensitivity", s}})
}

// screen is one display as the recomp numbers them ("monitor" cvar, 1 = first).
type screen struct {
	N       int  `json:"n"`
	W       int  `json:"w"`
	H       int  `json:"h"`
	Primary bool `json:"primary"`
}
