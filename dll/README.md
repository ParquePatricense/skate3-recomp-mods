# skate3-direct

`xinput9_1_0.dll` que el modeador pone junto a `skate3.exe` con "Directo a tu partida". El recomp la carga al arrancar: le pasa el joystick al `xinput9_1_0.dll` de Windows y, solo si `skate3.exe` es la v2.0.2 (sello y tamaño exactos), parchea el arranque dentro del juego y escribe `skate3-direct.log`.

- `src/main.c`: entrada de la dll, reenvío de XInputGetState, chequeo de versión.
- `src/core.c`: printf mínimo, log y el motor de ganchos (salto a una función nuestra al principio de una función recompilada; la original sigue disponible).
- `src/game.c`: los ganchos del arranque (BootFlow, pantallas del front end, películas), el corte del logo de EA y el monitor que anota cada espera.
- `src/prof.c`: perfilador por muestreo de la carga (qué funciones del juego ocupan cada hilo), al final del log.
- `test/`: arnés para Wine que mapea `skate3.exe`, instala los ganchos y prueba el perfilador.

Compilar (Linux, clang + lld): `sh build.sh` (deja la dll en `../go/payload/`). Sin CRT ni dependencias.

`SKATE3_DIRECT=0` la apaga entera; `SKATE3_DIRECT_LOGO=0`, solo el corte del logo.
