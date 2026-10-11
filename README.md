# 🛹 Skate 3 Recomp Mods 🎮

Pack de mods **modular** para [skate3recomp](https://github.com/mchughalex/skate3recomp): se baja todo junto y activás solo lo que quieras. Cada módulo se pone y se saca desde un menú, sin tocar nada más.

| Módulo | Qué hace |
|---|---|
| 🎮 **Iconos PlayStation** | Cambia los iconos de Xbox por PS3, PS4 o PS5 |
| 🖼️ **Presets gráficos** | Máxima calidad, Equilibrado o Rendimiento, con un clic |
| 🔭 **FOV y ultrawide** | Campo de visión (40–120) y modo 21:9 |
| ⌨️ **Alt+F4** | Cierra el juego con Alt+F4 (el recomp no lo deja) |
| 🏎️ **Velocímetro** | Muestra tu velocidad en km/h mientras patinás |

![preview](preview/preview.png)

| PS4 | PS5 |
|---|---|
| ![PS4 en el juego](preview/ingame-ps4.jpg) | ![PS5 en el juego](preview/ingame-ps5.jpg) |

## 📦 Instalación (Windows)

1. ⬇️ Descargá `Skate3-Mods.zip` desde [Releases](../../releases).
2. 📂 Descomprimilo **en la carpeta de skate3recomp** (donde está `skate3.exe`).
3. 🖱️ Abrí `Skate3-Mods.exe` y activá lo que quieras desde el menú.
4. 🛹 Jugá con `Jugar Skate 3 (mods).bat` (hace falta para Alt+F4 y el velocímetro; los iconos y presets funcionan igual abriendo `skate3.exe`).

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

## 🏎️ Velocímetro

El velocímetro lee la posición del skater de la memoria del juego, así que primero hay que **calibrarlo una vez** (2 pasadas de ~40 segundos):

1. Abrí el juego y entrá a patinar.
2. En `Skate3-Mods.exe` elegí **Calibrar velocímetro** y volvé al juego. Seguí los pitidos:
   - 🔔 1 pitido = quedate quieto
   - 🔔🔔 2 pitidos = patiná derecho y rápido
   - 🔔🔔🔔 3 pitidos = terminó
3. Cerrá el juego, abrilo de nuevo y repetí la calibración. Con eso confirma que la encuentra siempre y queda activado.

Se muestra a la derecha de la pantalla solo cuando el juego está en primer plano. Para mph: agregá `speed_unit = mph` en `skate3-mods.ini`.

## ⚙️ Cómo funciona

- **Iconos:** reemplaza solo las texturas de `buttons.rx2` y agrega la versión nueva al final del `.big`; el resto del archivo queda intacto.
- **Alt+F4 y velocímetro:** `Jugar Skate 3 (mods).bat` abre el juego y deja corriendo un ayudante chico que se cierra solo cuando cerrás el juego. Alt+F4 le pide al juego que se cierre (igual que la X de la ventana). El velocímetro es una ventana transparente encima del juego.
- 🚫 No incluye ni distribuye archivos de EA: los iconos son arte original y todo se aplica sobre tu propia copia.

### 🐧 Linux / 🍎 macOS

Los iconos funcionan con Python 3.8+ (sin dependencias):

```sh
python3 src/skate3_ps_buttons.py install ps5 --game /ruta/a/skate3recomp
python3 src/skate3_ps_buttons.py uninstall    --game /ruta/a/skate3recomp
```

## 🛠️ Compilar

```sh
cd go && go test ./... && GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ../dist/Skate3-Mods.exe .
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

Modular mod pack for skate3recomp: PlayStation button icons (PS3/PS4/PS5), graphics presets, FOV/ultrawide, Alt+F4 to quit and a speedometer. Unzip next to `skate3.exe`, run `Skate3-Mods.exe` to toggle modules, and play through `Jugar Skate 3 (mods).bat` for Alt+F4 and the speedometer (calibrate it once from the menu). Every module is reversible. Not affiliated with EA, Sony or the skate3recomp project.
