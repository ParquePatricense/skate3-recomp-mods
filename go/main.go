// Skate3-Mods: modular mod pack for Skate 3 Recomp (skate3recomp).
//
// Modules (each one can be turned on or off on its own):
//   - PlayStation button icons (PS3/PS4/PS5)       -> game files
//   - Graphics presets, FOV, ultrawide              -> skate3.toml
//   - Alt+F4 to quit                                 -> helper ("play" mode)
//   - Desktop shortcut                               -> Skate 3.lnk
//   - Rioplatense Spanish texts (voseo)              -> game files
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const version = "2.2.0"

var stdin = bufio.NewReader(os.Stdin)

func ask(prompt string) string {
	fmt.Print(prompt)
	line, _ := stdin.ReadString('\n')
	return strings.TrimSpace(line)
}

func pause() { ask(T("\nPresioná Enter para seguir...", "\nPress Enter to continue...")) }

func onOff(b bool) string {
	if b {
		return T("ACTIVADO", "ON")
	}
	return T("desactivado", "off")
}

func printStatus(root string) {
	c := loadConfig(root)
	icons, _ := iconsStatus(root)
	preset := T("el tuyo (sin preset)", "yours (no preset)")
	if p := findPreset(c["preset"]); p != nil {
		preset = p.label()
	}
	fmt.Printf(T("  Carpeta del juego: %s\n\n", "  Game folder: %s\n\n"), root)
	fmt.Printf(T("  Iconos de botones : %s\n", "  Button icons      : %s\n"), icons)
	fmt.Printf(T("  Preset gráfico    : %s\n", "  Graphics preset   : %s\n"), preset)
	if c["fov"] != "" {
		fmt.Printf(T("  Campo de visión   : %s\n", "  Field of view     : %s\n"), c["fov"])
	}
	if c["ultrawide"] != "" {
		fmt.Printf("  Ultrawide         : %s\n", onOff(c["ultrawide"] == "1"))
	}
	fmt.Printf(T("  Alt+F4 para salir : %s\n", "  Alt+F4 to quit    : %s\n"), onOff(c.on("altf4")))
	fmt.Printf(T("  Acceso directo    : %s\n", "  Desktop shortcut  : %s\n"), onOff(c.on("shortcut")))
	fmt.Printf(T("  Textos rioplatense: %s\n", "  Rioplatense texts : %s\n"), onOff(textsOn(root)))
}

func toggleTexts(root string) {
	if textsOn(root) {
		report(uninstallTexts(root, root))
		return
	}
	report(installTexts(root, root))
}

func menuIcons(root string) {
	fmt.Print(T("\nIconos de botones:\n", "\nButton icons:\n") +
		"  1) PS3\n  2) PS4\n  3) PS5\n  4) Xbox (original)\n" + T("  0) Volver\n", "  0) Back\n"))
	var err error
	switch ask(T("Elegí: ", "Choose: ")) {
	case "1":
		err = installIcons("ps3", root)
	case "2":
		err = installIcons("ps4", root)
	case "3":
		err = installIcons("ps5", root)
	case "4":
		err = uninstallIcons(root)
	default:
		return
	}
	report(err)
}

func menuPresets(root string) {
	fmt.Println(T("\nPreset gráfico (lo podés seguir ajustando en Settings > Video del juego):",
		"\nGraphics preset (you can keep tweaking it in the game's Settings > Video):"))
	for i, p := range presets {
		fmt.Printf("  %d) %s\n", i+1, p.label())
	}
	fmt.Printf(T("  %d) Volver a mi configuración de antes\n  0) Volver\n",
		"  %d) Restore my previous settings\n  0) Back\n"), len(presets)+1)
	n, _ := strconv.Atoi(ask(T("Elegí: ", "Choose: ")))
	switch {
	case n >= 1 && n <= len(presets):
		report(applyPreset(root, presets[n-1].key))
	case n == len(presets)+1:
		report(restoreToml(root))
	}
}

func menuView(root string) {
	c := loadConfig(root)
	fmt.Print(T("\nCampo de visión y ultrawide:\n  1) Cambiar campo de visión (FOV)\n  2) Activar/desactivar ultrawide (21:9 o más)\n  0) Volver\n",
		"\nField of view and ultrawide:\n  1) Change field of view (FOV)\n  2) Toggle ultrawide (21:9 or wider)\n  0) Back\n"))
	switch ask(T("Elegí: ", "Choose: ")) {
	case "1":
		v, err := strconv.ParseFloat(strings.ReplaceAll(ask(T("FOV en grados (40-120, el original es 60): ", "FOV in degrees (40-120, default 60): ")), ",", "."), 64)
		if err != nil || v < 40 || v > 120 {
			fmt.Println(T("Valor inválido.", "Invalid value."))
			return
		}
		s := strconv.FormatFloat(v, 'f', 1, 64)
		if err := setTomlValues(root, [][2]string{{"skate3_field_of_view", s}}); err != nil {
			report(err)
			return
		}
		c["fov"] = s
		report(c.save(root))
	case "2":
		on := c["ultrawide"] != "1"
		v := "false"
		c["ultrawide"] = "0"
		if on {
			v, c["ultrawide"] = "true", "1"
		}
		if err := setTomlValues(root, [][2]string{{"skate3_ultrawide", v}}); err != nil {
			report(err)
			return
		}
		report(c.save(root))
	}
}

func toggleAltF4(root string) {
	c := loadConfig(root)
	if c.on("altf4") {
		c["altf4"] = "0"
	} else {
		c["altf4"] = "1"
	}
	if err := c.save(root); err != nil {
		report(err)
		return
	}
	fmt.Printf("Alt+F4: %s\n", onOff(c.on("altf4")))
	if c.on("altf4") {
		fmt.Println(T("Abrí el juego con el acceso directo \"Skate 3\" o con \"Jugar Skate 3 (mods).bat\" para que funcione.",
			"Start the game with the \"Skate 3\" shortcut or \"Jugar Skate 3 (mods).bat\" for it to work."))
	}
}

func toggleShortcut(root string) {
	c := loadConfig(root)
	var err error
	if c.on("shortcut") {
		err = removeShortcut()
		c["shortcut"] = "0"
	} else {
		err = createShortcut(root)
		c["shortcut"] = "1"
	}
	if err != nil {
		report(err)
		return
	}
	if err := c.save(root); err != nil {
		report(err)
		return
	}
	fmt.Printf(T("Acceso directo \"Skate 3\" en el escritorio: %s\n", "\"Skate 3\" desktop shortcut: %s\n"), onOff(c.on("shortcut")))
}

func toggleLang(root string) {
	c := loadConfig(root)
	if langEN {
		c["lang"] = "es"
	} else {
		c["lang"] = "en"
	}
	setLang(c)
	report(c.save(root))
}

func report(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(T("Listo.", "Done."))
	}
}

func menu() int {
	root, err := findRoot("")
	if err == nil {
		setLang(loadConfig(root))
	} else {
		setLang(config{})
	}
	update := make(chan string, 1)
	go func() { update <- newerVersion() }()
	notice := ""
	select {
	case notice = <-update:
	case <-time.After(1500 * time.Millisecond):
	}
	for {
		fmt.Printf("\n===== Skate 3 Recomp Mods v%s =====\n\n", version)
		if err != nil {
			fmt.Println("Error:", err)
			pause()
			return 1
		}
		if notice != "" {
			fmt.Printf(T("  >> Hay una versión nueva (%s): %s\n\n", "  >> A new version is available (%s): %s\n\n"), notice, releasesURL())
		}
		printStatus(root)
		fmt.Print(T("\n  1) Iconos de botones (PS3 / PS4 / PS5 / Xbox)\n"+
			"  2) Preset gráfico\n"+
			"  3) Campo de visión / ultrawide\n"+
			"  4) Alt+F4 para salir (activar/desactivar)\n"+
			"  5) Acceso directo \"Skate 3\" en el escritorio (crear/quitar)\n"+
			"  6) Textos en español rioplatense, con voseo (activar/desactivar)\n"+
			"  L) Language: English\n"+
			"  0) Salir\n\n",
			"\n  1) Button icons (PS3 / PS4 / PS5 / Xbox)\n"+
				"  2) Graphics preset\n"+
				"  3) Field of view / ultrawide\n"+
				"  4) Alt+F4 to quit (on/off)\n"+
				"  5) \"Skate 3\" desktop shortcut (create/remove)\n"+
				"  6) Rioplatense (Argentine) Spanish texts (on/off)\n"+
				"  L) Idioma: español\n"+
				"  0) Exit\n\n"))
		switch strings.ToLower(ask(T("Elegí una opción: ", "Choose an option: "))) {
		case "1":
			menuIcons(root)
		case "2":
			menuPresets(root)
		case "3":
			menuView(root)
		case "4":
			toggleAltF4(root)
		case "5":
			toggleShortcut(root)
		case "6":
			toggleTexts(root)
		case "l":
			toggleLang(root)
			continue
		case "0", "":
			return 0
		default:
			continue
		}
		pause()
	}
}

func run(args []string) int {
	if len(args) == 0 {
		return menu()
	}
	cmd := args[0]
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	game := fs.String("game", "", T("carpeta de skate3recomp (por defecto: la del .exe)", "skate3recomp folder (default: the .exe's folder)"))
	rest := args[1:]
	var arg string
	if len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		arg, rest = rest[0], rest[1:]
	}
	if err := fs.Parse(rest); err != nil {
		return 2
	}
	root, err := findRoot(*game)
	if err == nil {
		setLang(loadConfig(root))
		switch cmd {
		case "install":
			err = installIcons(strings.ToLower(arg), *game)
		case "uninstall":
			err = uninstallIcons(*game)
		case "status":
			printStatus(root)
		case "preset":
			if arg == "restore" {
				err = restoreToml(root)
			} else {
				err = applyPreset(root, arg)
			}
		case "texts":
			switch arg {
			case "on":
				err = installTexts(root, *game)
			case "off":
				err = uninstallTexts(root, *game)
			default:
				fmt.Println(onOff(textsOn(root)))
			}
		case "play":
			err = play(root)
		default:
			fmt.Println(T("Uso", "Usage") + ": Skate3-Mods [install ps3|ps4|ps5 | uninstall | status | preset max|balanced|perf|restore | texts on|off | play] [--game DIR]")
			return 2
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if cmd == "play" {
			pause()
		}
		return 1
	}
	return 0
}

func main() { os.Exit(run(os.Args[1:])) }
