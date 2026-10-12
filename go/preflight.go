package main

// Check before playing: things that make the game fail in confusing ways
// (no room for saves and logs, a read-only game folder, a missing exe) are
// caught before Drop in, with a plain message.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// preflight returns an error that stops the launch, or a warning to show.
func preflight(root string) (warn string, err error) {
	if recompExe(root) == "" {
		return "", errors.New(T("no está skate3.exe en la carpeta del juego: revisá la carpeta en El taller",
			"skate3.exe is missing from the game folder: check the folder in The shop"))
	}
	probe := filepath.Join(root, ".skate3mods-write-test")
	if f, e := os.Create(probe); e != nil {
		return "", errors.New(T("la carpeta del juego es de solo lectura, así que el juego no puede guardar: movela fuera de Archivos de programa o sacale el solo lectura",
			"the game folder is read-only, so the game can't save: move it out of Program Files or clear read-only"))
	} else {
		f.Close()
		os.Remove(probe)
	}
	if free, ok := freeSpace(root); ok && free < 500<<20 {
		return fmt.Sprintf(T("Ojo: quedan %d MB libres en el disco del juego. Con poco espacio la partida puede no guardarse.",
			"Heads up: only %d MB free on the game's drive. With little space your save may fail."), free>>20), nil
	}
	return "", nil
}
