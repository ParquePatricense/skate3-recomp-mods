package main

// Extras for the direct load: where you start (the recomp injects the pad
// inputs that go there through the pause menu once you're skating) and the
// tips on the loading screen.

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
)

type spot struct{ key, nameES, nameEN, inputs string }

// spots: "here" is where your save left you. The PCU Library sequence is the
// one the recomp documents for skate3_demo_path_gameplay_inputs.
var spots = []spot{
	{"here", "Donde dejaste", "Where you left off", ""},
	{"pcu", "Biblioteca de la PCU (experimental)", "PCU Library (experimental)",
		"start,a,rt:1500,rt:1500,down,down,a,down,down,down,down,down,a,a"},
}

// spotPace shortens the recomp's waits for the trip to a spot: it waits
// 2.5 s after you're skating and 0.6 s between presses by default. The map
// tabs keep their own slower timing inside the sequence.
var spotPace = [][2]string{{"skate3_demo_path_input_settle_ms", "1000"}, {"skate3_demo_path_input_delay_ms", "450"}}

func spotInputs(key string) string {
	for _, s := range spots {
		if s.key == key {
			return s.inputs
		}
	}
	return ""
}

func spotOn(root string) bool { return readToml(root)["skate3_demo_path_gameplay_inputs"] != "" }

func spotList() []map[string]string {
	out := []map[string]string{}
	for _, s := range spots {
		out = append(out, map[string]string{"key": s.key, "name": T(s.nameES, s.nameEN)})
	}
	return out
}

// setSpot picks where the direct load drops you.
func setSpot(root, key string) (string, error) {
	for _, s := range spots {
		if s.key != key {
			continue
		}
		c := loadConfig(root)
		c["spot"] = key
		if err := c.save(root); err != nil {
			return "", err
		}
		if directOn(root) {
			if err := setTomlValues(root, append([][2]string{{"skate3_demo_path_gameplay_inputs", strconv.Quote(s.inputs)}}, spotPace...)); err != nil {
				return "", err
			}
		}
		return T("Arrancás en: ", "You start at: ") + T(s.nameES, s.nameEN), nil
	}
	return "", errors.New("?")
}

// loadingTip is one tip for the loading screen, with the buttons of the
// icon style that is installed.
func loadingTip(style string) string {
	b := strings.NewReplacer("{A}", "A", "{Y}", "Y", "{LT}", "LT", "{RT}", "RT")
	if strings.HasPrefix(style, "ps") {
		b = strings.NewReplacer("{A}", "✕", "{Y}", "△", "{LT}", "L2", "{RT}", "R2")
	}
	t := loadingTips[rand.Intn(len(loadingTips))]
	return b.Replace(T(t[0], t[1]))
}

var loadingTips = [][2]string{
	{"Para un ollie, bajá el stick derecho y soltalo hacia arriba.", "For an ollie, pull the right stick down and flick it up."},
	{"En el aire, agarrá la tabla con {LT} y {RT}.", "In the air, grab the board with {LT} and {RT}."},
	{"Empujá con {A} para agarrar velocidad.", "Push with {A} to pick up speed."},
	{"Con {Y} te bajás de la tabla y caminás.", "Press {Y} to step off the board and walk."},
	{"Si el juego se traba, Ctrl+Alt+R lo cierra y lo vuelve a abrir.", "If the game hangs, Ctrl+Alt+R closes and reopens it."},
	{"Alt+F4 cierra el juego como la X de la ventana.", "Alt+F4 closes the game like the window's X."},
}
