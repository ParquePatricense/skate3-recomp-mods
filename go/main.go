// Skate3-Mods: modular mod pack for Skate 3 Recomp (skate3recomp).
//
// Modules (each one can be turned on or off on its own):
//   - PlayStation button icons (PS3/PS4/PS5)       -> game files
//   - Graphics presets, FOV, ultrawide              -> skate3.toml
//   - Alt+F4 to quit                                 -> helper ("play" mode)
//   - Desktop shortcuts                              -> Skate 3 Mods / (directo) / (normal).lnk
//   - "Español ARG" language (replaces Dutch)      -> game files
//   - In-game helper: quick restart, priority, keep awake, hidden cursor,
//     Windows key, window memory, do not disturb (play mode)
//
// With no arguments it opens the modder window (gui_windows.go); "menu"
// opens the classic console menu.
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

const version = "3.9.2"

var stdin = newStdin()

func newStdin() *bufio.Reader { return bufio.NewReader(os.Stdin) }

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
	fmt.Printf(T("  Alt+F4 para salir : %s\n", "  Alt+F4 to quit    : %s\n"), onOff(c.feature("altf4")))
	fmt.Printf(T("  Acceso directo    : %s\n", "  Desktop shortcut  : %s\n"), onOff(c.on("shortcut")))
	fmt.Printf(T("  Español ARG       : %s\n", "  Español ARG       : %s\n"), onOff(textsOn(root)))
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

func menuHelper(root string) {
	for {
		c := loadConfig(root)
		fmt.Println(T("\nAyudante en el juego (abrí el juego con el acceso directo o el .bat):", "\nIn-game helper (start the game with the shortcut or the .bat):"))
		for i, f := range helperFeatures {
			fmt.Printf("  %d) %-32s %s\n", i+1, T(f.nameES, f.nameEN), onOff(c.feature(f.key)))
		}
		fmt.Print(T("  0) Volver\n", "  0) Back\n"))
		n, _ := strconv.Atoi(ask(T("Elegí: ", "Choose: ")))
		if n < 1 || n > len(helperFeatures) {
			return
		}
		f := helperFeatures[n-1]
		if err := setFeature(root, f.key, !c.feature(f.key)); err != nil {
			report(err)
		}
	}
}

func menuShop(root string) {
	fmt.Print(T("\nTaller:\n  1) Reparar archivos del juego\n  2) Diagnóstico (reporte en el escritorio)\n  3) Actualizar a la última versión\n  4) Desinstalar todo\n  0) Volver\n",
		"\nShop:\n  1) Repair game files\n  2) Diagnostics (report on the desktop)\n  3) Update to the latest version\n  4) Uninstall everything\n  0) Back\n"))
	switch ask(T("Elegí: ", "Choose: ")) {
	case "1":
		fmt.Println(strings.Join(repairMods(root), "\n"))
	case "2":
		p, err := diagnose(root)
		if err == nil {
			fmt.Println(p)
		}
		report(err)
	case "3":
		if err := selfUpdate(); err != nil {
			report(err)
			return
		}
		os.Exit(0)
	case "4":
		if a := strings.ToLower(ask(T("¿Seguro? Escribí SI para confirmar: ", "Sure? Type YES to confirm: "))); a == T("si", "yes") || a == "yes" {
			fmt.Println(strings.Join(uninstallAll(root), "\n"))
		}
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
	next := langs[0].code // the next of the seven languages
	for i, l := range langs {
		if l.code == lang {
			next = langs[(i+1)%len(langs)].code
		}
	}
	c["lang"] = next
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
	setLang(config{})
	root, err := findRoot("")
	if err != nil {
		if d, ok := askGameFolder(true); ok {
			root, err = findRoot(d)
		}
	}
	if err == nil {
		setLang(loadConfig(root))
		if settleIn(root) {
			fmt.Printf(T("Me copié a %s: de ahora en más podés abrirme desde ahí o desde el acceso directo.\n",
				"Copied myself to %s: from now on open me from there or from the shortcut.\n"), root)
		}
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
			"  4) Ayudante en el juego (Alt+F4, reinicio rápido, etc.)\n"+
			"  5) Acceso directo \"Skate 3\" en el escritorio (crear/quitar)\n"+
			"  6) Idioma \"Español ARG\" (rioplatense) en lugar del neerlandés (activar/desactivar)\n"+
			"  7) Taller: reparar, diagnóstico, actualizar, desinstalar todo\n"+
			"  C) Cambiar la carpeta del juego\n"+
			"  L) Language: English\n"+
			"  0) Salir\n\n",
			"\n  1) Button icons (PS3 / PS4 / PS5 / Xbox)\n"+
				"  2) Graphics preset\n"+
				"  3) Field of view / ultrawide\n"+
				"  4) In-game helper (Alt+F4, quick restart, etc.)\n"+
				"  5) \"Skate 3\" desktop shortcut (create/remove)\n"+
				"  6) \"Español ARG\" language (Rioplatense) instead of Dutch (on/off)\n"+
				"  7) Shop: repair, diagnostics, update, uninstall everything\n"+
				"  C) Change the game folder\n"+
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
			menuHelper(root)
		case "5":
			toggleShortcut(root)
		case "6":
			toggleTexts(root)
		case "7":
			menuShop(root)
		case "l":
			toggleLang(root)
			continue
		case "c":
			if d, ok := askGameFolder(false); ok {
				if r, e := findRoot(d); e == nil {
					root = r
					setLang(loadConfig(root))
					settleIn(root)
				}
			}
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
		case "preload": // Windows sign-in: warm the game files, quietly
			preload(root)
		case "play":
			// the "directo" and "normal" shortcuts pick the startup, and
			// "directo, English" and the like the language too
			mode, lang, _ := strings.Cut(arg, ":")
			if want := mode == "direct"; (mode == "direct" || mode == "normal") && want != directOn(root) {
				if _, err = setDirect(root, want); err != nil {
					break
				}
			}
			if mode == "direct" && lang != "" && lang != loadConfig(root)["boot_lang"] {
				if _, err = setBootLang(root, lang); err != nil {
					break
				}
			}
			err = play(root)
		default:
			fmt.Println(T("Uso", "Usage") + ": Skate3-Mods [install ps3|ps4|ps5|switch|xseries|gamesir|steamdeck | uninstall | status | preset max|balanced|perf|ultra|restore | texts on|off | play] [--game DIR]")
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

func main() {
	cleanOldExe()
	args := os.Args[1:]
	switch {
	case len(args) == 0:
		if err := runGUI(); err == nil {
			return
		}
		consoleSetup(false) // no WebView2: the classic menu
	case args[0] == "menu":
		consoleSetup(false)
		args = nil
	case args[0] != "play" && args[0] != "preload":
		consoleSetup(true)
	}
	os.Exit(run(args))
}
