package main

// Extras for the direct load: the tips on the loading screen.

import (
	"math/rand"
	"strings"
)

// spotPace shortens the recomp's waits for a gameplay input sequence: it
// waits 2.5 s after you're skating and 0.6 s between presses by default.
var spotPace = [][2]string{{"skate3_demo_path_input_settle_ms", "1000"}, {"skate3_demo_path_input_delay_ms", "450"}}

func spotOn(root string) bool { return readToml(root)["skate3_demo_path_gameplay_inputs"] != "" }

// dropSpot undoes the starting spot older versions offered (the PCU
// Library): the direct load goes back to where your save left you.
func dropSpot(root string) {
	c := loadConfig(root)
	if _, ok := c["spot"]; !ok {
		return
	}
	delete(c, "spot")
	c.save(root)
	if spotOn(root) {
		setTomlValues(root, [][2]string{{"skate3_demo_path_gameplay_inputs", `""`}})
	}
}

// loadingTip is one tip for the loading screen, with the buttons of the
// icon style that is installed.
func loadingTip(style string) string {
	b := strings.NewReplacer("{A}", "A", "{Y}", "Y", "{LT}", "LT", "{RT}", "RT")
	switch {
	case strings.HasPrefix(style, "ps") || style == "bn":
		b = strings.NewReplacer("{A}", "✕", "{Y}", "△", "{LT}", "L2", "{RT}", "R2")
	case style == "switch": // the bottom button is B and the top one X on a Nintendo pad
		b = strings.NewReplacer("{A}", "B", "{Y}", "X", "{LT}", "ZL", "{RT}", "ZR")
	case style == "steamdeck":
		b = strings.NewReplacer("{A}", "A", "{Y}", "Y", "{LT}", "L2", "{RT}", "R2")
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
