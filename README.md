# 🛹 Skate 3 Recomp Mods 🎮

**🇦🇷 Español** · [🇬🇧 English](README.en.md) · [🇧🇷 Português](README.pt.md) · [🇫🇷 Français](README.fr.md) · [🇮🇹 Italiano](README.it.md) · [🇩🇪 Deutsch](README.de.md) · [🇨🇳 中文](README.zh.md) · [🇯🇵 日本語](README.ja.md)

Pack de mods **modular** para [skate3recomp](https://github.com/mchughalex/skate3recomp): se baja todo junto y activás solo lo que quieras. Cada módulo se pone y se saca desde **el modeador**, una ventana con onda Skate 3 y radio propia, y todo se puede deshacer.

> 🧩 **Requiere [skate3recomp](https://github.com/mchughalex/skate3recomp) v2.0.2 y tu propia copia de Skate 3.** Esto es solo el modeador: no trae el juego ni el recomp.

> ⚡ **De 16,6 s a ~7 s:** con *Directo a tu partida*, un parche dentro del juego te deja patinando en tu partida en unos 7 segundos desde que tocás **DROP IN**.

![modeador](preview/modeador.png)

## ✨ Qué trae

| Módulo | Qué hace |
|---|---|
| 🎮 **Iconos de botones** | Cambia los iconos de Xbox 360 por los de tu joystick: PS3, PS4, PS5, Switch (1 y 2, Pro, Joy-Con), Xbox Series, GameSir o Steam Deck, más **blanco y negro** (alto contraste) y **para daltónicos** |
| 🖼️ **Presets gráficos** | Máxima calidad, Equilibrado, Rendimiento o Ultra liviano, con un clic |
| 🖥️ **Pantalla** | VSync, pantalla completa y en qué monitor abre el juego (ideal con dos monitores) |
| ⏩ **Directo a tu partida** | Una pantalla de carga negra tapa la intro y los menús, un parche dentro del juego acelera todo el arranque y aparecés en tu carrera en el idioma que elijas |
| ⌨️ **Teclado y mouse** | Activa el modo teclado+mouse del recomp, sensibilidad y remapeo de cada botón tocando la tecla |
| 🔭 **FOV y ultrawide** | Campo de visión (40–120) y modo 21:9 |
| 🛹 **Ayudante en sesión** | Alt+F4, reinicio rápido (Ctrl+Alt+R), prioridad alta, pantalla siempre encendida, cursor oculto, tecla Windows bloqueada, ventana recordada y no molestar |
| ⚡ **Carga** | Defender sin revisar el juego, precarga al prender la PC, mover el juego a un SSD, gráfico de tus últimas 20 cargas, tu última carga al lado de DROP IN y el log de la carga a un clic |
| 🌐 **Español ARG** | Agrega "Español ARG" (voseo y vocabulario rioplatense) en lugar del neerlandés; el español original queda igual |
| 🇧🇷 **Português (Brasil)** | Los textos del juego en portugués para la carga directa (unos 13 600 textos) |
| 🔧 **El taller** | Accesos directos, Steam, tiempo de juego, varias versiones del recomp, reparar, diagnóstico, actualizar y desinstalar todo |
| 🎮 **Modeador con joystick** | Se maneja con el mando o con el teclado, se adapta a tu pantalla y trae radio con favoritos, Mix y aleatorio |

🌍 El modeador está en Español, English, Português (Brasil), Français, Italiano, Deutsch, 中文 y 日本語: elige solo el idioma de Windows y se cambia en **Idioma**.

![preview](preview/preview.png)

| PS4 | PS5 |
|---|---|
| ![PS4 en el juego](preview/ingame-ps4.jpg) | ![PS5 en el juego](preview/ingame-ps5.jpg) |

## 📦 Instalación (Windows)

1. ⬇️ Descargá `Skate3-Mods.zip` desde [Releases](../../releases) y descomprimilo donde quieras (Descargas sirve).
2. 🖱️ Abrí `Skate3-Mods.exe`: busca solo la carpeta de skate3recomp, se copia ahí y la recuerda. Si no la encuentra, te deja elegirla.
3. ✅ Activá lo que quieras.
4. 🛹 Jugá con **DROP IN**, los accesos directos "Skate 3 (directo)" / "Skate 3 (normal)" o `Jugar Skate 3 (mods).bat`. El ayudante en sesión solo corre si abrís el juego así; iconos, gráficos e idioma funcionan igual abriendo `skate3.exe`.

> 🪟 El modeador usa WebView2, que ya viene con Windows 10 y 11. Si faltara, se abre un menú de consola (también con `Skate3-Mods.exe menu`).

> ⚠️ El juego tiene que haberse instalado al menos una vez (botón *Select ISO*). Cerralo antes de cambiar iconos o presets.

> 🛡️ Si aparece SmartScreen: "Más información" ➡️ "Ejecutar de todas formas" (el .exe no está firmado).

## 🎮 Iconos de botones

| 🟢 Xbox | 🔵 PlayStation |
|---|---|
| Ⓐ Ⓑ Ⓧ Ⓨ | ✖️ ⭕ 🟪 🔺 |
| LB / RB / LT / RT | L1 / R1 / L2 / R2 |
| LS / RS (click) | L3 / R3 |
| Back / Start | PS3: Select / Start · PS4: Share / Options · PS5: Create / Options |

También: **Switch** (Switch 1 y 2, Pro Controller, Joy-Con), **Xbox Series** (Vista / Menú), **GameSir** (Nova Lite, T4, G7 y la mayoría de los joysticks de PC) y **Steam Deck**. Para leer mejor: **Blanco y negro** (figuras de PlayStation negras sobre blanco) y **Para daltónicos** (letras de Xbox en colores Okabe-Ito, sin rojo contra verde), con una vista de cómo se ven en el juego. Los de Switch van por posición: si usás Steam con "Usar disposición de botones de Nintendo", A y B quedan cruzados.

Los iconos están en `game/data/big/fedynamic.big` ➡️ `buttons.rx2`. El mod guarda el original (`fedynamic.big.xbox-original`) y "Xbox (original)" lo restaura byte a byte. ➕ El D-Pad queda igual y el dibujo del stick del Trick Guide sigue diciendo "RS" (es otra textura).

## 🖼️ Presets, FOV y ultrawide

Escriben valores en `skate3.toml` (el mismo archivo que usa el menú *Settings* del juego, así que después podés seguir ajustando ahí). La primera vez se guarda tu `skate3.toml` como `skate3.toml.mods-original`; "Volver a mi configuración de antes" lo restaura.

**Ultra liviano** apaga sombras, SSAO, bloom, rayos de sol, niebla y MSAA, y acorta la distancia de dibujado: para PCs muy justas. **Pantalla** escribe `vsync`, `fullscreen` y `monitor` (0 = automática, 1 = primera pantalla, 2 = segunda...).

## ⌨️ Teclado y mouse

Activa `mnk_mode` del recomp: el mouse mueve la cámara (stick derecho) y cada botón del joystick va a una tecla. En **Controles** tocás un botón y apretás la tecla o el click que quieras (si ya la usaba otro botón, se intercambian). "Volver a las teclas de fábrica" deja las del recomp.

## ⏩ Directo a tu partida

Usa la automatización de arranque del propio recomp (`skate3_demo_path` con `skate3_demo_path_signed_in`, así conserva tu perfil y tu partida): saltea la intro, acepta la pantalla de idioma, el Press Start y los avisos de arranque, y suelta los controles en cuanto empezás a patinar. Mientras tanto, una pantalla de carga negra (con un consejo con los botones de tu estilo y una barra que aprende cuánto tarda tu carga) tapa el juego y se va apenas podés jugar; el juego queda muteado hasta entonces. El idioma que elegís se pone en el lugar del inglés dentro de `miscboot.big` mientras la opción está activa.

### 🧩 El parche de carga directa

Con la opción activa, el modeador pone `xinput9_1_0.dll` junto a `skate3.exe`. El recomp la carga al arrancar (le pasa el joystick al xinput de Windows, así que los controles andan igual) y, **solo si `skate3.exe` es la v2.0.2** (cada parche se busca por la tabla de funciones del exe y se verifica byte a byte), acelera el arranque dentro del propio juego:

- 🚫 **Logo de EA** terminado a los 0,03 s en vez de ~4,7 s.
- 🏎️ **Sin tope de FPS** mientras carga (vuelve tu valor al llegar al juego).
- ⏱️ **Press Start** sin su espera mínima y con START al instante.
- 😴 **Cargadores sin siesta:** los hilos de carga dejan de dormir 1 ms entre cada pasada.
- 💓 **Tick del juego a 4 ms** en los pasos cortos de la carga (en vez de 16,7 ms).
- 🖤 **Fundidos en un cuadro** y la placa de video con prioridad alta hasta llegar al juego.
- 📜 **`skate3-direct.log`** junto a `skate3.exe`: cuánto tardó cada paso (se abre desde **Carga**).

Con otra versión del recomp no toca nada. Al apagar la opción o desinstalar se borra; si había otra `xinput9_1_0.dll` que no es nuestra, no se toca. Cada parte se puede apagar con una variable de entorno (`SKATE3_DIRECT=0` la apaga entera). Detalle y código en [`dll/`](dll/).

| | Antes | Ahora |
|---|---|---|
| ⏩ Carga directa completa | ~16,6 s | **~7 s** |

## ⚡ Carga

- 🛡️ **Windows Defender:** agrega la carpeta del juego a las exclusiones (pide permiso de administrador una vez), para que no revise cada archivo al cargar.
- 🌅 **Precarga al prender la PC:** al iniciar Windows lee en segundo plano, sin ventanas, el exe, las DLL, la caché de shaders y los archivos del juego, así la primera carga del día también es rápida.
- 💽 **SSD:** si el juego está en un disco rígido y tenés un SSD, lo copia a `X:\Skate3Recomp`, verifica el tamaño y reacomoda accesos, Steam y precarga; la carpeta vieja se borra solo si lo pedís.
- 📊 **Historial:** tus últimas 20 cargas contra los 10 s y en qué se fue el tiempo de cada una; la última aparece al lado de **DROP IN**.
- 📜 **Log de la carga:** un botón abre la carpeta del juego con `skate3-direct.log` marcado.

## 🌐 Español ARG

En la pantalla de idiomas del juego, "Nederlands" pasa a ser **"Español ARG"**: los textos en español del juego reescritos con reglas. "¿Quieres jugar? Puedes..." ➡️ "¿Querés jugar? Podés...", "Haz un kickflip y presiona Ⓐ" ➡️ "Hacé un kickflip y presioná Ⓐ", "colega" ➡️ "loco", "coches" ➡️ "autos", "piscina" ➡️ "pileta", "mola" ➡️ "garpa", y así (unos 2700 textos). El español original queda tal cual. Las reglas están en [`go/rioplatense.txt`](go/rioplatense.txt): si ves algo raro, se corrige ahí.

Toca solo `miscboot.big` (lo guarda como `miscboot.big.original`); al desactivarlo vuelve el neerlandés, byte a byte.

## 🛹 Ayudante en sesión

Corre en segundo plano mientras jugás y se cierra con el juego. Cada cosa se prende o apaga en **En sesión**:

- ⌨️ **Alt+F4** cierra el juego como la X de la ventana.
- 🔁 **Ctrl+Alt+R** cierra y vuelve a abrir el juego si se traba.
- ⚡ **Prioridad alta** para el proceso del juego.
- ☀️ **Pantalla siempre encendida**, aunque juegues solo con el mando.
- 🖱️ **Cursor oculto** sobre el juego a los 3 segundos sin mover el mouse.
- 🪟 **Tecla Windows bloqueada** mientras el juego está en primer plano.
- 🗔 **Ventana recordada**: monitor, tamaño y modo se guardan al cerrar.
- 🔕 **No molestar**: apaga las notificaciones de Windows y las vuelve a prender al salir.

## 🔧 El taller

- 🖱️ **Accesos directos:** "Skate 3 Mods", "Skate 3 (directo)" y "Skate 3 (normal)" en el escritorio y el menú Inicio, también uno por idioma.
- 🚂 **Steam:** Skate 3 en tu biblioteca (Big Picture, Steam Input).
- ⏱️ **Tiempo de juego** por día y por semana.
- 🧪 **Versiones del recomp:** varias carpetas del juego, cada una con sus mods y ajustes.
- 🩹 **Reparar:** revisa `fedynamic.big` y `miscboot.big` y arregla lo roto desde la copia.
- 🔄 **Después de actualizar el recomp:** si un archivo cambió, se toma como nuevo original y se vuelven a poner tus mods.
- 🩺 **Diagnóstico** y 📣 **reporte de problemas**, con el final del log del juego.
- ⬇️ **Actualizar:** baja la versión nueva de Releases y se reemplaza solo.
- 💀 **Desinstalar todo:** deja el juego exactamente como estaba antes de los mods.

## 📻 Radio Carverton

Boom bap, punk, rock, electro y cumbia generados en vivo con Web Audio, distintos en cada tema y con uno nuevo cada vez que abrís el modeador (elegís el género o Mix, y aleatorio). La tecla M la pausa. No usa archivos de música ni temas con copyright.

## ⚙️ Cómo funciona

- **Iconos y textos:** reemplazan solo su archivo dentro del `.big` (agregan la versión nueva al final); el resto queda intacto.
- **Ayudante:** `Skate3-Mods.exe play` (lo usan DROP IN, los accesos, Steam y el `.bat`; `play direct` y `play normal` eligen cómo arranca) abre el juego y deja un proceso chico sin ventana que se cierra con el juego.
- **Modeador:** una página HTML embebida en el `.exe` (`go/ui/index.html`) dentro de WebView2.
- 🚫 No incluye ni distribuye archivos de EA: los iconos son arte original y los mods se aplican sobre tu propia copia del juego.

### 🐧 Linux / 🍎 macOS

Los iconos funcionan con Python 3.8+ (sin dependencias):

```sh
python3 src/skate3_ps_buttons.py install ps5 --game /ruta/a/skate3recomp
python3 src/skate3_ps_buttons.py uninstall    --game /ruta/a/skate3recomp
```

## 🛠️ Compilar

```sh
cd go && GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -H windowsgui" -o ../dist/Skate3-Mods.exe .
```

La DLL del parche: `cd dll && sh build.sh` (Linux, clang + lld). Al publicar un tag `v*`, GitHub Actions compila el `.exe` y sube el zip a Releases.

## 🖌️ Regenerar los iconos

```sh
pip install pillow numpy
python tools/build_payload.py preview/png   # redibuja y regenera src/payload.py y go/payload/
```

Fuente: Inter (SIL Open Font License, ver `tools/fonts/Inter-LICENSE.txt`).

---

📄 Licencia en [`LICENSE`](LICENSE). Sin relación con EA, Sony, Microsoft, Nintendo, Valve ni el proyecto skate3recomp. Skate 3 es marca de Electronic Arts.
