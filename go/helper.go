package main

// Things the in-game helper ("play" mode) can do while the game runs. Each
// one is a key in skate3-mods.ini; a missing key means its default.

type helperFeature struct {
	key            string
	def            bool
	nameES, nameEN string
	descES, descEN string
}

var helperFeatures = []helperFeature{
	{"altf4", true, "Alt+F4 para salir", "Alt+F4 to quit",
		"Cierra el juego con Alt+F4, como la X de la ventana.", "Closes the game with Alt+F4, like the window's X."},
	{"restart", true, "Reinicio rápido (Ctrl+Alt+R)", "Quick restart (Ctrl+Alt+R)",
		"Si el juego se traba, Ctrl+Alt+R lo cierra y lo vuelve a abrir.", "If the game hangs, Ctrl+Alt+R closes and reopens it."},
	{"priority", false, "Prioridad alta", "High priority",
		"Windows le da más CPU al juego que a lo que corre de fondo.", "Windows gives the game more CPU than background apps."},
	{"awake", true, "Pantalla siempre encendida", "Keep the screen on",
		"La PC no se suspende ni apaga la pantalla mientras jugás con el mando.", "The PC won't sleep or turn the screen off while you play with a gamepad."},
	{"cursor", true, "Ocultar el cursor", "Hide the cursor",
		"El puntero desaparece sobre el juego si no movés el mouse 3 segundos.", "The pointer disappears over the game after 3 seconds without moving the mouse."},
	{"winkey", true, "Bloquear la tecla Windows", "Block the Windows key",
		"La tecla Windows no te saca del juego mientras está en primer plano.", "The Windows key won't kick you out while the game is in front."},
	{"window", true, "Recordar la ventana", "Remember the window",
		"El juego vuelve a abrir en el mismo monitor, tamaño y modo (ventana o pantalla completa).", "The game reopens on the same monitor, size and mode (window or fullscreen)."},
	{"dnd", false, "No molestar", "Do not disturb",
		"Silencia las notificaciones de Windows mientras jugás.", "Silences Windows notifications while you play."},
}

func (c config) feature(key string) bool {
	if _, ok := c[key]; ok {
		return c.on(key)
	}
	for _, f := range helperFeatures {
		if f.key == key {
			return f.def
		}
	}
	return false
}

func (c config) anyHelper() bool {
	for _, f := range helperFeatures {
		if c.feature(f.key) {
			return true
		}
	}
	return false
}

func setFeature(root, key string, on bool) error {
	c := loadConfig(root)
	c[key] = "0"
	if on {
		c[key] = "1"
	}
	return c.save(root)
}
