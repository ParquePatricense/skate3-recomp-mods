# 🛹 Skate 3 Recomp Mods 🎮

Pack de mods **modular** para [skate3recomp](https://github.com/mchughalex/skate3recomp): se baja todo junto y activás solo lo que quieras. Cada módulo se pone y se saca desde **el modeador**, una ventana con onda Skate 3 (pantalla completa, F11 la pasa a ventana) y radio propia con boom bap, punk, rock, electro y cumbia, sin tocar nada más.

| Módulo | Qué hace |
|---|---|
| 🎮 **Iconos de botones** | Cambia los iconos de Xbox 360 por los de tu joystick: PS3, PS4, PS5, Switch (1 y 2, Pro, Joy-Con), Xbox Series, GameSir o Steam Deck, más blanco y negro (alto contraste) y para daltónicos |
| 🖼️ **Presets gráficos** | Máxima calidad, Equilibrado, Rendimiento o Ultra liviano, con un clic |
| 🖥️ **Pantalla** | VSync, pantalla completa y en qué monitor abre el juego (ideal con dos monitores) |
| ⏩ **Directo a tu partida** | Carga directa: una pantalla de carga negra tapa la intro y los menús, un parche dentro del juego corta el logo de EA y aparecés en tu carrera en el idioma que elijas |
| 🎮 **Modeador con joystick** | Se maneja con el mando (cruz, A, LB/RB, Start) o con atajos de teclado; se adapta solo a tu pantalla (tamaño a gusto y modo sofá); radio con favoritos, Mix y aleatorio; reportar un problema en un toque |
| ⌨️ **Teclado y mouse** | Activa el modo teclado+mouse del recomp, sensibilidad y remapeo de cada botón tocando la tecla |
| 🔭 **FOV y ultrawide** | Campo de visión (40–120) y modo 21:9 |
| 🛹 **Ayudante en sesión** | Alt+F4, reinicio rápido (Ctrl+Alt+R), prioridad alta, pantalla siempre encendida, cursor oculto, tecla Windows bloqueada, ventana recordada y no molestar |
| ⚡ **Carga** | Defender sin revisar el juego, precarga al prender la PC, mover el juego a un SSD con un botón, gráfico de tus últimas 20 cargas contra los 10 s, la última al lado de Drop in y el log de la carga a un clic |
| 🔗 **Accesos (en El taller)** | Con un botón: "Skate 3 Mods" (el modeador), "Skate 3 (directo)" y "Skate 3 (normal)" en el escritorio y el menú Inicio, uno por idioma, Skate 3 en Steam (Big Picture, Steam Input), tiempo de juego por día y semana, y varias versiones del recomp |
| 🌐 **Español ARG** | Agrega el idioma "Español ARG" (voseo y vocabulario rioplatense) en lugar del neerlandés; el español original queda igual |
| 🇧🇷 **Português (Brasil)** | Los textos del juego en portugués para la carga directa (unos 13 600 textos) |

🔧 **El taller:** aviso cuando sale una versión nueva de skate3recomp, reparar archivos, reaplicar los mods solo si actualizás el recomp, diagnóstico, carpetas a mano, actualización en un toque y desinstalar todo.

🌍 En English, Español, Português (Brasil), Français, Italiano, Deutsch, 中文 y 日本語: elige solo el idioma de Windows y se cambia en **Idioma**.

![modeador](preview/modeador.png)

![preview](preview/preview.png)

| PS4 | PS5 |
|---|---|
| ![PS4 en el juego](preview/ingame-ps4.jpg) | ![PS5 en el juego](preview/ingame-ps5.jpg) |

## 📦 Instalación (Windows)

1. ⬇️ Descargá `Skate3-Mods.zip` desde [Releases](../../releases) y descomprimilo donde quieras (Descargas sirve).
2. 🖱️ Abrí `Skate3-Mods.exe`: se abre el modeador, busca solo la carpeta de skate3recomp, se copia ahí y la recuerda. Si no la encuentra, te deja elegirla.
3. ✅ Activá lo que quieras.
4. 🛹 Jugá con **DROP IN**, los accesos directos "Skate 3 (directo)" / "Skate 3 (normal)" o `Jugar Skate 3 (mods).bat`. El ayudante en sesión solo corre si abrís el juego así; iconos, gráficos e idioma funcionan igual abriendo `skate3.exe`.

> 🪟 El modeador usa WebView2, que ya viene con Windows 10 y 11. Si faltara, se abre el menú clásico de consola (también con `Skate3-Mods.exe menu`).

> ⚠️ El juego tiene que haberse instalado al menos una vez (botón *Select ISO*). Cerralo antes de cambiar iconos o presets.
> 🛡️ Si aparece SmartScreen: "Más información" ➡️ "Ejecutar de todas formas" (el .exe no está firmado).

## 🎮 Iconos de botones

| 🟢 Xbox | 🔵 PlayStation |
|---|---|
| Ⓐ Ⓑ Ⓧ Ⓨ | ✖️ ⭕ 🟪 🔺 |
| LB / RB / LT / RT | L1 / R1 / L2 / R2 |
| LS / RS (click) | L3 / R3 |
| Back / Start | PS3: Select / Start · PS4: Share / Options · PS5: Create / Options |

También: **Switch** (Switch 1 y 2, Pro Controller, Joy-Con: B abajo, A a la derecha, L / R, ZL / ZR, − / +), **Xbox Series** (LB / RB, LT / RT, Vista / Menú), **GameSir** (Nova Lite, T4, G7 y la mayoría de los joysticks de PC) y **Steam Deck** (L1 / R1, L2 / R2, L3 / R3). Para leer mejor: **Blanco y negro** (figuras de PlayStation negras sobre blanco) y **Para daltónicos** (letras de Xbox en colores Okabe-Ito, sin rojo contra verde). Los de Switch van por posición: si usás Steam con "Usar disposición de botones de Nintendo" activado, A y B quedan cruzados.

Los iconos están en `game/data/big/fedynamic.big` ➡️ `buttons.rx2`. El mod guarda el original (`fedynamic.big.xbox-original`) y al elegir "Xbox (original)" lo restaura byte a byte. ➕ El D-Pad queda igual y el dibujo del stick del Trick Guide sigue diciendo "RS" (es otra textura).

## 🖼️ Presets, FOV y ultrawide

Escriben valores en `skate3.toml` (el mismo archivo que usa el menú *Settings* del juego, así que después podés seguir ajustando ahí). La primera vez se guarda tu `skate3.toml` como `skate3.toml.mods-original`; "Volver a mi configuración de antes" lo restaura.

**Ultra liviano** apaga sombras, SSAO, bloom, rayos de sol, niebla y MSAA, y acorta la distancia de dibujado: para PCs muy justas. **Imagen** escribe `present_letterbox` (bandas negras) y `skate3_native_render_scene_ssr` (reflejos, experimental). **Pantalla** escribe `vsync`, `fullscreen` y `monitor` (0 = automática, 1 = primera pantalla, 2 = segunda...).

## ⌨️ Teclado y mouse

Activa `mnk_mode` del recomp: el mouse mueve la cámara (stick derecho) y cada botón del joystick va a una tecla. En **Controles** tocás un botón y apretás la tecla o el click que quieras (si ya la usaba otro botón, se intercambian). Se guardan como `keybind_*` en `skate3.toml`. "Volver a las teclas de fábrica" deja las del recomp.

## ⏩ Directo a tu partida

Usa la automatización de arranque del propio recomp (`skate3_demo_path` con `skate3_demo_path_signed_in`, así conserva tu perfil y tu partida): saltea la intro, acepta la pantalla de idioma, el Press Start y los avisos de arranque, y deja de tocar los controles en cuanto empezás a patinar. Mientras tanto, el ayudante tapa el monitor del juego con una pantalla de carga y la saca cuando el registro del juego (`logs/`, que con esta opción se escribe cada segundo) dice que ya estás patinando; si algo sale raro, se va sola a los 2 minutos. Mientras está puesta, el juego queda muteado en el mezclador de volumen de Windows, el fondo es negro, con un consejo con los botones de tu estilo y una barra que aprende cuánto tarda tu carga.

Para que tarde lo menos posible: el juego carga con prioridad alta, el modeador deja los archivos del juego en la memoria de Windows mientras está abierto (lectura de fondo con prioridad baja, hasta 1,5 GB y nunca más de un cuarto de la memoria libre), la pantalla de carga se va apenas el registro dice que ya podés jugar y en **Carga → Historial de cargas** ves cuánto tardó cada parte: abrir el juego, logos y menús, y tu partida y el mapa. Esa automatización siempre acepta inglés, así que el idioma que elegís en el modeador (Español, Español ARG, Português, Français, Deutsch, Italiano) se pone en el lugar del inglés dentro de `miscboot.big` mientras la opción está activa; al apagarla vuelve el original.

**Parche de carga directa:** con la opción activa, el modeador pone `xinput9_1_0.dll` junto a `skate3.exe`. El recomp la carga al arrancar (le pasa el joystick al xinput de Windows, así que los controles andan igual) y, solo si `skate3.exe` es la v2.0.2 para la que está hecha (cada parche se busca por la tabla de funciones del propio exe y se verifica byte a byte), parchea el arranque dentro del propio juego: el logo de EA termina en el acto en vez de esperar sus ~4,7 s, el tope de FPS del recomp se levanta mientras carga (y vuelve a tu valor al llegar al juego), y en `skate3-direct.log` (junto a `skate3.exe`) queda cuánto tardó cada paso del arranque y en qué se fue la carga del mundo. Con otra versión del recomp no toca nada. Al apagar la opción o desinstalar se borra; si había otra `xinput9_1_0.dll` que no es nuestra, no se toca. Sin el modeador: `SKATE3_DIRECT=0` la apaga entera; `SKATE3_DIRECT_LOGO=0`, y `SKATE3_DIRECT_FPS=0` apagan cada parte. Código en [`dll/`](dll/).

## ⚡ Carga

- 🛡️ **Windows Defender:** agrega la carpeta del juego y el proceso a las exclusiones (pide permiso de administrador una vez), para que no revise cada archivo al cargar.
- 🌅 **Precarga al prender la PC:** 45 s después de iniciar Windows lee en segundo plano, sin ventanas, primero el exe, las DLL y la caché de shaders del recomp y después los archivos del juego (lo mismo hace el modeador mientras está abierto).
- 💽 **SSD:** si el juego está en un disco rígido y tenés un SSD, lo copia a `X:\Skate3Recomp`, verifica el tamaño y reacomoda accesos, Steam y precarga; la carpeta vieja se borra solo si lo pedís.
- 🐢 **Carga lenta:** si una carga tarda bastante más que tu promedio, te dice qué parte se hizo lenta.

La carga directa rondaba los 16,6 s: ~1,8 s el arranque del recomp, ~4,7 s el logo de EA, ~2,4 s la pantalla de Press Start y ~7 s el propio juego armando el mundo. El parche corta el logo; el Press Start y el mundo se recortan con lo que mide `skate3-direct.log`.

## 🌐 Español ARG

En la pantalla de idiomas del juego, "Nederlands" pasa a ser **"Español ARG"**. Ese idioma son los textos en español del juego reescritos con reglas: "¿Quieres jugar? Puedes..." ➡️ "¿Querés jugar? Podés...", "Haz un kickflip y presiona Ⓐ" ➡️ "Hacé un kickflip y presioná Ⓐ", "colega" ➡️ "loco", "coches" ➡️ "autos", "piscina" ➡️ "pileta", "mola" ➡️ "garpa", y así (unos 2700 textos). El español original queda tal cual. Las reglas están en [`go/rioplatense.txt`](go/rioplatense.txt): si ves algo raro, se corrige ahí.

Toca solo `miscboot.big` (lo guarda como `miscboot.big.original`); al desactivarlo vuelve el neerlandés, byte a byte.

## 🛹 Ayudante en sesión

Corre en segundo plano mientras jugás y se cierra con el juego. Cada cosa se prende o apaga en **En sesión**:

- ⌨️ **Alt+F4** cierra el juego como la X de la ventana (el recomp se come la tecla).
- 🔁 **Ctrl+Alt+R** cierra y vuelve a abrir el juego si se traba.
- ⚡ **Prioridad alta** para el proceso del juego.
- ☀️ **Pantalla siempre encendida**, aunque juegues solo con el mando.
- 🖱️ **Cursor oculto** sobre el juego a los 3 segundos sin mover el mouse.
- 🪟 **Tecla Windows bloqueada** mientras el juego está en primer plano.
- 🗔 **Ventana recordada**: monitor, tamaño y modo se guardan en `skate3.toml` al cerrar.
- 🔕 **No molestar**: apaga las notificaciones de Windows y las vuelve a prender al salir.

ℹ️ La intro ya se saltea con Enter o A (lo trae el recomp). No la salteamos sola porque el mismo toque podría pasar de largo la pantalla de idiomas.

## 🔧 El taller

- 🩹 **Reparar:** revisa `fedynamic.big` y `miscboot.big`; si alguno está roto, vuelve desde la copia y se reaplica el mod.
- 🔄 **Después de actualizar el recomp:** al jugar o abrir el modeador, si un archivo cambió, se toma como nuevo original y se vuelven a poner tus mods.
- 📜 **Registro del juego:** abre el log de la última sesión; el reporte de problemas incluye su final.
- 🧭 **Asistente:** tres preguntas (mando, PC y arranque) y queda todo configurado; aparece solo la primera vez.
- 💥 **Cierre inesperado:** si el juego se cierra con un error, el modeador ofrece armar el reporte.
- 🩺 **Diagnóstico:** deja `Skate3-Mods-diagnostico.txt` en el escritorio para pedir ayuda.
- 📁 **Carpetas:** juego, partidas y datos, a un toque.
- ⬇️ **Actualizar:** baja la versión nueva de Releases y se reemplaza solo.
- 💀 **Desinstalar todo:** deja el juego exactamente como estaba antes de los mods.

## 📻 Radio Carverton

El modeador tiene radio: boom bap, punk, rock, electro y cumbia generados en vivo con Web Audio, distintos en cada tema (elegís el género o Mix, y aleatorio). Arranca sola al abrir el modeador, cada tema dura cerca de dos minutos y la tecla M la pausa. No usa ningún archivo de música ni temas con copyright.

## ⚙️ Cómo funciona

- **Iconos y textos:** reemplazan solo su archivo dentro del `.big` (agregan la versión nueva al final); el resto queda intacto.
- **Ayudante:** `Skate3-Mods.exe play` (lo usan DROP IN, los accesos directos, Steam y el `.bat`; `play direct` y `play normal` eligen cómo arranca) abre el juego y deja corriendo un proceso chico sin ventana que se cierra con el juego.
- **Modeador:** una página HTML embebida en el `.exe` (`go/ui/index.html`) dentro de WebView2.
- 🚫 No incluye ni distribuye archivos de EA: los iconos son arte original, los textos se generan con reglas sobre tu propia copia.

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

Al publicar un tag `v*`, GitHub Actions compila el `.exe` y sube el zip a Releases.

## 🖌️ Regenerar los iconos

```sh
pip install pillow numpy
python tools/build_payload.py preview/png   # redibuja y regenera src/payload.py y go/payload/
```

Fuente: Inter (SIL Open Font License, ver `tools/fonts/Inter-LICENSE.txt`).

---

### 🇬🇧 English

Modular mod pack for skate3recomp with a Skate 3 styled modder window (and its own generated-beats radio): button icons (PS3/PS4/PS5, Switch, Xbox Series, GameSir, Steam Deck, black and white, colour-blind safe), graphics presets (including an ultra light one), FOV/ultrawide, VSync/fullscreen/which monitor the game opens on, keyboard and mouse play with click-to-remap keys, an in-game helper (Alt+F4, Ctrl+Alt+R quick restart, high priority, keep awake, hidden cursor, Windows key block, window memory, do not disturb), desktop shortcuts, a Steam (non-Steam game) entry, a "straight to your save" direct load behind a loading screen (with an in-game patch that cuts the EA logo), an "Español ARG" (Rioplatense Spanish) language in Dutch's slot, Brazilian Portuguese game texts, a Loading tab (Defender exclusion, preload at sign-in, move to SSD, last 20 loads against 10 s, your last load next to Drop in, the load log in one click), play time stats, and a shop with a skate3recomp new-version notice, repair, auto re-apply after recomp updates, diagnostics, one-click update and uninstall-all. Unzip anywhere and run `Skate3-Mods.exe` (it finds the skate3recomp folder by itself and copies itself there) to toggle modules, then hit DROP IN (or use the "Skate 3 (directo)" / "Skate 3 (normal)" shortcuts) so the helper runs. Every module is reversible. Not affiliated with EA, Sony or the skate3recomp project.
