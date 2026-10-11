# Skate 3 Recomp – PlayStation Button Icons

Mod para [skate3recomp](https://github.com/mchughalex/skate3recomp) que cambia los iconos de botones de Xbox 360 por iconos de **PS3, PS4 o PS5**. Se instala y se desinstala con un clic.

![preview](preview/preview.png)

| PS4 | PS5 |
|---|---|
| ![PS4 en el juego](preview/ingame-ps4.jpg) | ![PS5 en el juego](preview/ingame-ps5.jpg) |

| Xbox | PlayStation |
|---|---|
| A / B / X / Y | ✕ / ○ / □ / △ |
| LB / RB / LT / RT | L1 / R1 / L2 / R2 |
| LS / RS (click) | L3 / R3 |
| Back / Start | PS3: Select / Start · PS4: Share / Options · PS5: Create / Options |

## Instalación (Windows)

1. Descargá `Skate3-PS-Buttons.zip` desde [Releases](../../releases).
2. Descomprimilo **en la carpeta de skate3recomp** (donde está `skate3.exe`).
3. Doble clic en `Instalar PS3.bat`, `Instalar PS4.bat` o `Instalar PS5.bat`.
4. Abrí el juego.

Para volver a los iconos de Xbox: `Desinstalar (Xbox original).bat`.
Para cambiar de PS3 a PS5 (o al revés), ejecutá directamente el otro `Instalar`.

> El juego tiene que haberse instalado al menos una vez (botón *Select ISO* del recomp).
> Cerrá el juego antes de instalar o desinstalar.

### Linux / macOS / desde código

Requiere Python 3.8+ (sin dependencias):

```sh
python3 src/skate3_ps_buttons.py install ps5 --game /ruta/a/skate3recomp
python3 src/skate3_ps_buttons.py uninstall    --game /ruta/a/skate3recomp
python3 src/skate3_ps_buttons.py status       --game /ruta/a/skate3recomp
```

## Cómo funciona

Los iconos están en `game/data/big/fedynamic.big` → `buttons.rx2` (66 texturas DXT5 de 64×64 en formato tileado de Xbox 360).
El instalador:

- guarda una copia del archivo original como `fedynamic.big.xbox-original` (solo la primera vez),
- reemplaza únicamente las texturas de `buttons.rx2` y agrega la versión nueva al final del `.big` (el resto del archivo queda intacto),
- al desinstalar, restaura el original byte a byte.

El D-Pad queda igual (la cruz es la misma en PlayStation). El dibujo del stick en el Trick Guide ("RS") es otra textura y por ahora no se cambia.

El mod no incluye ni distribuye archivos de EA: los iconos son arte original generado con `tools/gen_icons.py` y se aplican sobre tu propia copia del juego.

## Compilar el .exe

```sh
cd go && GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ../dist/Skate3-PS-Buttons.exe .
```

Al publicar un tag `v*`, GitHub Actions compila el `.exe` y sube el zip a Releases.

## Regenerar los iconos

```sh
pip install pillow numpy
python tools/build_payload.py preview/png   # redibuja y regenera src/payload.py y go/payload/
```

Fuente: Inter (SIL Open Font License, ver `tools/fonts/Inter-LICENSE.txt`).

---

### English

Swaps Skate 3 Recomp's Xbox 360 button prompts for PS3, PS4 or PS5 ones. Unzip the release next to `skate3.exe` and run one of the `Instalar PSx.bat` files (`Desinstalar` restores the Xbox icons). A backup of the original `fedynamic.big` is kept and restored byte-for-byte. Not affiliated with EA, Sony or the skate3recomp project.
