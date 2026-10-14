# skate3-direct

`xinput9_1_0.dll` que el modeador pone junto a `skate3.exe` con "Directo a tu partida". El recomp la carga al arrancar: le pasa el joystick al `xinput9_1_0.dll` de Windows y, solo si `skate3.exe` es la v2.0.2 (sello de compilación; cada parche se ubica por la tabla de funciones del propio exe y se verifica byte a byte), parchea el arranque dentro del juego y escribe `skate3-direct.log`.

- `src/main.c`: entrada de la dll, reenvío de XInputGetState, chequeo de versión.
- `src/core.c`: printf mínimo, log y el motor de ganchos (salto a una función nuestra al principio de una función recompilada; la original sigue disponible).
- `src/game.c`: los ganchos del arranque (BootFlow, pantallas del front end, películas), el corte del logo de EA, la carga sin tope de FPS (saca `skate3_guest_fps_cap(_auto)` del recomp hasta llegar al juego y los devuelve), Press Start al toque (activa la pantalla 22 con su propia función sin esperar su animación; el START lo aprieta el recomp y la pantalla elige mando y perfil como siempre: forzar sus banderas a mano cargaba modo libre sin tu partida) y el monitor que anota cada espera.
- `src/prof.c`: perfilador por muestreo de la carga (qué funciones del juego ocupan cada hilo), al final del log. Solo con `SKATE3_DIRECT_PERFIL=1`: suspende todos los hilos cada 4 ms y alarga la carga.
- `test/`: arnés para Wine que mapea `skate3.exe`, instala los ganchos y prueba el perfilador; `flag.c` prueba la API de ajustes de `rexruntime.dll`.

Compilar (Linux, clang + lld): `sh build.sh` (deja la dll en `../go/payload/`). Sin CRT ni dependencias.

`SKATE3_DIRECT=0` la apaga entera; `SKATE3_DIRECT_LOGO=0`, solo el corte del logo; `SKATE3_DIRECT_FPS=0`, solo la carga sin tope; `SKATE3_DIRECT_PRESS=0`, solo el Press Start al toque. `SKATE3_DIRECT_START=1` da el START por dentro (solo para pruebas: pierde la partida). `SKATE3_DIRECT_FADE=1` prende el turbo del paso post-START (apagado: no ganaba nada).
