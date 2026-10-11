# 🛹 Skate 3 Recomp Mods 🎮

Pack de mods **modular** para [skate3recomp](https://github.com/mchughalex/skate3recomp): se baja todo junto y activás solo lo que quieras. Cada módulo se pone y se saca desde un menú, sin tocar nada más.

| Módulo | Qué hace |
|---|---|
| 🎮 **Iconos PlayStation** | Cambia los iconos de Xbox por PS3, PS4 o PS5 |
| 🖼️ **Presets gráficos** | Máxima calidad, Equilibrado o Rendimiento, con un clic |
| 🔭 **FOV y ultrawide** | Campo de visión (40–120) y modo 21:9 |
| ⌨️ **Alt+F4** | Cierra el juego con Alt+F4 (el recomp no lo deja) |
| 🖥️ **Acceso directo** | Crea "Skate 3" en el escritorio, con mods incluidos |
| 🧉 **Textos rioplatenses** | Pasa los textos del juego a español rioplatense: voseo y vocabulario de acá |

➕ El menú sale en español o en inglés según el idioma de Windows (se cambia con `L`), y avisa cuando hay una versión nueva.

![preview](preview/preview.png)

| PS4 | PS5 |
|---|---|
| ![PS4 en el juego](preview/ingame-ps4.jpg) | ![PS5 en el juego](preview/ingame-ps5.jpg) |

## 📦 Instalación (Windows)

1. ⬇️ Descargá `Skate3-Mods.zip` desde [Releases](../../releases).
2. 📂 Descomprimilo **en la carpeta de skate3recomp** (donde está `skate3.exe`).
3. 🖱️ Abrí `Skate3-Mods.exe` y activá lo que quieras desde el menú.
4. 🛹 Jugá con el acceso directo "Skate 3" (opción 5 del menú) o con `Jugar Skate 3 (mods).bat`. Hace falta para Alt+F4; el resto funciona igual abriendo `skate3.exe`.

> ⚠️ El juego tiene que haberse instalado al menos una vez (botón *Select ISO*). Cerralo antes de cambiar iconos o presets.
> 🛡️ Si aparece SmartScreen: "Más información" ➡️ "Ejecutar de todas formas" (el .exe no está firmado).

## 🎮 Iconos PlayStation

| 🟢 Xbox | 🔵 PlayStation |
|---|---|
| Ⓐ Ⓑ Ⓧ Ⓨ | ✖️ ⭕ 🟪 🔺 |
| LB / RB / LT / RT | L1 / R1 / L2 / R2 |
| LS / RS (click) | L3 / R3 |
| Back / Start | PS3: Select / Start · PS4: Share / Options · PS5: Create / Options |

Los iconos están en `game/data/big/fedynamic.big` ➡️ `buttons.rx2`. El mod guarda el original (`fedynamic.big.xbox-original`) y al elegir "Xbox (original)" lo restaura byte a byte. ➕ El D-Pad queda igual y el dibujo del stick del Trick Guide sigue diciendo "RS" (es otra textura).

## 🖼️ Presets, FOV y ultrawide

Escriben valores en `skate3.toml` (el mismo archivo que usa el menú *Settings* del juego, así que después podés seguir ajustando ahí). La primera vez se guarda tu `skate3.toml` como `skate3.toml.mods-original`; "Volver a mi configuración de antes" lo restaura.

## 🧉 Textos rioplatenses

Reescribe la tabla de textos en español del juego (`miscboot.big` ➡️ `LANGUAGE_Spanish_Global_skate3ng.BIN`) con reglas: "¿Quieres jugar? Puedes..." ➡️ "¿Querés jugar? Podés...", "Haz un kickflip y presiona Ⓐ" ➡️ "Hacé un kickflip y presioná Ⓐ", "colega" ➡️ "loco", "coches" ➡️ "autos", "piscina" ➡️ "pileta", "mola" ➡️ "garpa", y así (cambia unos 2700 textos). Las reglas están en [`go/rioplatense.txt`](go/rioplatense.txt): si ves algo raro, se corrige ahí.

Al activarlo pone el juego en español (`user_language = 5` en `skate3.toml`). Guarda el original como `miscboot.big.original` y al desactivarlo lo restaura byte a byte.

## ⚙️ Cómo funciona

- **Iconos y textos:** reemplazan solo su archivo dentro del `.big` (agregan la versión nueva al final); el resto queda intacto.
- **Alt+F4:** `Jugar Skate 3 (mods).bat` abre el juego y deja corriendo un ayudante chico (sin ventana) que se cierra solo cuando cerrás el juego. Alt+F4 le pide al juego que se cierre, igual que la X de la ventana.
- 🚫 No incluye ni distribuye archivos de EA: los iconos son arte original, los textos se generan con reglas sobre tu propia copia.

### 🐧 Linux / 🍎 macOS

Los iconos funcionan con Python 3.8+ (sin dependencias):

```sh
python3 src/skate3_ps_buttons.py install ps5 --game /ruta/a/skate3recomp
python3 src/skate3_ps_buttons.py uninstall    --game /ruta/a/skate3recomp
```

## 🛠️ Compilar

```sh
cd go && GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ../dist/Skate3-Mods.exe .
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

Modular mod pack for skate3recomp: PlayStation button icons (PS3/PS4/PS5), graphics presets, FOV/ultrawide, Alt+F4 to quit, a desktop shortcut and Rioplatense (Argentine) Spanish texts. Unzip next to `skate3.exe`, run `Skate3-Mods.exe` to toggle modules (the menu follows your Windows language, press `L` to switch), and play through the "Skate 3" shortcut or `Jugar Skate 3 (mods).bat` for Alt+F4. Every module is reversible. Not affiliated with EA, Sony or the skate3recomp project.
