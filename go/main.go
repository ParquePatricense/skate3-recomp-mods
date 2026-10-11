// Skate3-Mods: modular mod pack for Skate 3 Recomp (skate3recomp).
//
// Modules (each one can be turned on or off on its own):
//   - PlayStation button icons (PS3/PS4/PS5)       -> game files
//   - Graphics presets, FOV, ultrawide              -> skate3.toml
//   - Alt+F4 to quit                                 -> helper ("play" mode)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const version = "2.1.0"

var stdin = bufio.NewReader(os.Stdin)

func ask(prompt string) string {
	fmt.Print(prompt)
	line, _ := stdin.ReadString('\n')
	return strings.TrimSpace(line)
}

func pause() { ask("\nPresiona Enter para seguir...") }

func onOff(b bool) string {
	if b {
		return "ACTIVADO"
	}
	return "desactivado"
}

func printStatus(root string) {
	c := loadConfig(root)
	icons, _ := iconsStatus(root)
	preset := "el tuyo (sin preset)"
	if p := findPreset(c["preset"]); p != nil {
		preset = p.name
	}
	fmt.Printf("  Carpeta del juego: %s\n\n", root)
	fmt.Printf("  Iconos de botones : %s\n", icons)
	fmt.Printf("  Preset grafico    : %s\n", preset)
	if c["fov"] != "" {
		fmt.Printf("  Campo de vision   : %s\n", c["fov"])
	}
	if c["ultrawide"] != "" {
		fmt.Printf("  Ultrawide         : %s\n", onOff(c["ultrawide"] == "1"))
	}
	fmt.Printf("  Alt+F4 para salir : %s\n", onOff(c.on("altf4")))
}

func menuIcons(root string) {
	fmt.Print("\nIconos de botones:\n  1) PS3\n  2) PS4\n  3) PS5\n  4) Xbox (original)\n  0) Volver\n")
	var err error
	switch ask("Elegi: ") {
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
	fmt.Println("\nPreset grafico (lo podes seguir ajustando en Settings > Video del juego):")
	for i, p := range presets {
		fmt.Printf("  %d) %s\n", i+1, p.name)
	}
	fmt.Printf("  %d) Volver a mi configuracion de antes\n  0) Volver\n", len(presets)+1)
	n, _ := strconv.Atoi(ask("Elegi: "))
	switch {
	case n >= 1 && n <= len(presets):
		report(applyPreset(root, presets[n-1].key))
	case n == len(presets)+1:
		report(restoreToml(root))
	}
}

func menuView(root string) {
	c := loadConfig(root)
	fmt.Print("\nCampo de vision y ultrawide:\n  1) Cambiar campo de vision (FOV)\n  2) Activar/desactivar ultrawide (21:9 o mas)\n  0) Volver\n")
	switch ask("Elegi: ") {
	case "1":
		v, err := strconv.ParseFloat(strings.ReplaceAll(ask("FOV en grados (40-120, el original es 60): "), ",", "."), 64)
		if err != nil || v < 40 || v > 120 {
			fmt.Println("Valor invalido.")
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

func toggle(root, key, what string) {
	c := loadConfig(root)
	if c.on(key) {
		c[key] = "0"
	} else {
		c[key] = "1"
	}
	if err := c.save(root); err != nil {
		report(err)
		return
	}
	fmt.Printf("%s: %s\n", what, onOff(c.on(key)))
	if c.on(key) {
		fmt.Println("Recorda abrir el juego con \"Jugar Skate 3 (mods).bat\" para que funcione.")
	}
}

func report(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Listo.")
	}
}

func menu() int {
	root, err := findRoot("")
	for {
		fmt.Printf("\n===== Skate 3 Recomp Mods v%s =====\n\n", version)
		if err != nil {
			fmt.Println("Error:", err)
			pause()
			return 1
		}
		printStatus(root)
		fmt.Print("\n  1) Iconos de botones (PS3 / PS4 / PS5 / Xbox)\n" +
			"  2) Preset grafico\n" +
			"  3) Campo de vision / ultrawide\n" +
			"  4) Alt+F4 para salir (activar/desactivar)\n" +
			"  0) Salir\n\n")
		switch ask("Elegi una opcion: ") {
		case "1":
			menuIcons(root)
		case "2":
			menuPresets(root)
		case "3":
			menuView(root)
		case "4":
			toggle(root, "altf4", "Alt+F4")
		case "0", "":
			return 0
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
	game := fs.String("game", "", "carpeta de skate3recomp (por defecto: la del .exe)")
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
		case "play":
			err = play(root)
		default:
			fmt.Println("Uso: Skate3-Mods [install ps3|ps4|ps5 | uninstall | status | preset max|balanced|perf|restore | play] [--game CARPETA]")
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
