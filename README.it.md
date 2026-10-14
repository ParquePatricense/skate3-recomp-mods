# 🛹 Skate 3 Recomp Mods 🎮

[🇦🇷 Español](README.md) · [🇬🇧 English](README.en.md) · [🇧🇷 Português](README.pt.md) · [🇫🇷 Français](README.fr.md) · **🇮🇹 Italiano** · [🇩🇪 Deutsch](README.de.md) · [🇨🇳 中文](README.zh.md) · [🇯🇵 日本語](README.ja.md)

Pacchetto di mod **modulare** per [skate3recomp](https://github.com/mchughalex/skate3recomp): si scarica tutto insieme e attivi solo quello che vuoi. Ogni modulo si mette e si toglie dal **modder**, una finestra in stile Skate 3 con la sua radio, e tutto si può annullare.

> 🧩 **Richiede [skate3recomp](https://github.com/mchughalex/skate3recomp) v2.0.2 e la tua copia di Skate 3.** Questo è solo il modder: non include né il gioco né il recomp.

> ⚡ **Da 16,6 s a ~7 s:** con *Dritto al salvataggio*, una patch dentro il gioco ti fa skateare nella tua partita in circa 7 secondi da quando premi **DROP IN**.

![modder](preview/modeador.png)

## ✨ Cosa contiene

| Modulo | Cosa fa |
|---|---|
| 🎮 **Icone dei tasti** | Sostituisce le icone dell'Xbox 360 con quelle del tuo controller: PS3, PS4, PS5, Switch (1 e 2, Pro, Joy-Con), Xbox Series, GameSir o Steam Deck, più **bianco e nero** (alto contrasto) e **per daltonici** |
| 🖼️ **Preset grafici** | Qualità massima, Bilanciato, Prestazioni o Ultra leggero, con un clic |
| 🖥️ **Schermo** | VSync, schermo intero e su quale monitor si apre il gioco (ideale con due monitor) |
| ⏩ **Dritto al salvataggio** | Una schermata di caricamento nera copre intro e menu, una patch dentro il gioco accelera tutto l'avvio e ti ritrovi nella tua carriera nella lingua che scegli |
| ⌨️ **Tastiera e mouse** | Attiva la modalità tastiera+mouse del recomp, sensibilità e rimappatura di ogni pulsante premendo il tasto |
| 🔭 **FOV e ultrawide** | Campo visivo (40–120) e modalità 21:9 |
| 🛹 **Aiutante in sessione** | Alt+F4, riavvio rapido (Ctrl+Alt+R), priorità alta, schermo sempre acceso, cursore nascosto, tasto Windows bloccato, finestra ricordata e non disturbare |
| ⚡ **Caricamento** | Defender che non controlla il gioco, precaricamento all'avvio del PC, spostamento del gioco su un SSD, grafico dei tuoi ultimi 20 caricamenti, il tuo ultimo caricamento accanto a DROP IN e il log del caricamento a un clic |
| 🌐 **Español ARG** | Aggiunge "Español ARG" (voseo e lessico rioplatense) al posto dell'olandese; lo spagnolo originale resta com'è |
| 🇧🇷 **Português (Brasil)** | I testi del gioco in portoghese per il caricamento diretto (circa 13 600 testi) |
| 🔧 **Officina** | Collegamenti, Steam, tempo di gioco, più versioni del recomp, riparazione, diagnostica, aggiornamento e disinstallazione completa |
| 🎮 **Modder col controller** | Si usa col controller o con la tastiera, si adatta al tuo schermo e ha una radio con preferiti, Mix e casuale |

🌍 Il modder è disponibile in Español, English, Português (Brasil), Français, Italiano, Deutsch, 中文 e 日本語: sceglie da solo la lingua di Windows e si cambia in **Lingua**.

![preview](preview/preview.png)

| PS4 | PS5 |
|---|---|
| ![PS4 nel gioco](preview/ingame-ps4.jpg) | ![PS5 nel gioco](preview/ingame-ps5.jpg) |

## 📦 Installazione (Windows)

1. ⬇️ Scarica `Skate3-Mods.zip` da [Releases](../../releases) ed estrailo dove vuoi (va bene anche Download).
2. 🖱️ Apri `Skate3-Mods.exe`: trova da solo la cartella di skate3recomp, si copia lì e se la ricorda. Se non la trova, ti fa scegliere.
3. ✅ Attiva quello che vuoi.
4. 🛹 Gioca con **DROP IN**, con i collegamenti "Skate 3 (directo)" / "Skate 3 (normal)" o con `Jugar Skate 3 (mods).bat`. L'aiutante in sessione funziona solo se apri il gioco in questo modo; icone, grafica e lingua funzionano anche aprendo `skate3.exe`.

> 🪟 Il modder usa WebView2, già incluso in Windows 10 e 11. Se mancasse, si apre un menu da console (anche con `Skate3-Mods.exe menu`).

> ⚠️ Il gioco deve essere stato installato almeno una volta (pulsante *Select ISO*). Chiudilo prima di cambiare icone o preset.

> 🛡️ Se compare SmartScreen: "Ulteriori informazioni" ➡️ "Esegui comunque" (il .exe non è firmato).

## 🎮 Icone dei tasti

| 🟢 Xbox | 🔵 PlayStation |
|---|---|
| Ⓐ Ⓑ Ⓧ Ⓨ | ✖️ ⭕ 🟪 🔺 |
| LB / RB / LT / RT | L1 / R1 / L2 / R2 |
| LS / RS (click) | L3 / R3 |
| Back / Start | PS3: Select / Start · PS4: Share / Options · PS5: Create / Options |

E inoltre: **Switch** (Switch 1 e 2, Pro Controller, Joy-Con), **Xbox Series** (Visualizza / Menu), **GameSir** (Nova Lite, T4, G7 e la maggior parte dei controller per PC) e **Steam Deck**. Per una lettura migliore: **Bianco e nero** (simboli PlayStation neri su bianco) e **Per daltonici** (lettere Xbox nei colori Okabe-Ito, senza rosso contro verde), con un'anteprima di come appaiono nel gioco. Quelle di Switch seguono la posizione: se usi Steam con "Usa la disposizione dei pulsanti Nintendo", A e B risultano invertiti.

Le icone si trovano in `game/data/big/fedynamic.big` ➡️ `buttons.rx2`. La mod salva l'originale (`fedynamic.big.xbox-original`) e "Xbox (original)" lo ripristina byte per byte. ➕ Il D-Pad resta com'è e il disegno dello stick nella Trick Guide continua a dire "RS" (è un'altra texture).

## 🖼️ Preset, FOV e ultrawide

Scrivono i valori in `skate3.toml` (lo stesso file usato dal menu *Settings* del gioco, quindi dopo puoi continuare a regolare da lì). La prima volta il tuo `skate3.toml` viene salvato come `skate3.toml.mods-original`; "Ripristina le mie impostazioni" lo rimette com'era.

**Ultra leggero** spegne ombre, SSAO, bloom, raggi di sole, nebbia e MSAA, e accorcia la distanza di disegno: per PC molto al limite. **Schermo** scrive `vsync`, `fullscreen` e `monitor` (0 = automatico, 1 = primo schermo, 2 = secondo...).

## ⌨️ Tastiera e mouse

Attiva `mnk_mode` del recomp: il mouse muove la telecamera (stick destro) e ogni pulsante del controller va su un tasto. In **Controlli** tocchi un pulsante e premi il tasto o il clic che vuoi (se lo usava già un altro pulsante, si scambiano). "Torna ai tasti di base" rimette quelli del recomp.

## ⏩ Dritto al salvataggio

Usa l'automazione di avvio del recomp stesso (`skate3_demo_path` con `skate3_demo_path_signed_in`, così mantiene il tuo profilo e il tuo salvataggio): salta l'intro, accetta la schermata della lingua, il Press Start e gli avvisi iniziali, e lascia i comandi appena inizi a skateare. Nel frattempo una schermata di caricamento nera (con un consiglio che usa i tasti del tuo stile e una barra che impara quanto dura il tuo caricamento) copre il gioco e sparisce appena puoi giocare; fino ad allora il gioco resta muto. La lingua che scegli prende il posto dell'inglese dentro `miscboot.big` finché l'opzione è attiva.

### 🧩 La patch del caricamento diretto

Con l'opzione attiva, il modder mette `xinput9_1_0.dll` accanto a `skate3.exe`. Il recomp la carica all'avvio (passa il controller all'xinput di Windows, quindi i comandi funzionano come prima) e, **solo se `skate3.exe` è la v2.0.2** (ogni patch viene cercata tramite la tabella delle funzioni dell'exe e verificata byte per byte), accelera l'avvio dentro il gioco stesso:

- 🚫 **Logo EA** chiuso a 0,03 s invece di ~4,7 s.
- 🏎️ **Nessun limite di FPS** durante il caricamento (il tuo valore torna quando arrivi in gioco).
- ⏱️ **Press Start** senza l'attesa minima e con START immediato.
- 😴 **Caricatori senza pisolino:** i thread di caricamento smettono di dormire 1 ms tra un giro e l'altro.
- 💓 **Tick del gioco a 4 ms** nei passaggi brevi del caricamento (invece di 16,7 ms).
- 🖤 **Dissolvenze in un solo fotogramma** e scheda video a priorità alta fino all'arrivo in gioco.
- 📜 **`skate3-direct.log`** accanto a `skate3.exe`: quanto è durato ogni passaggio (si apre da **Caricamento**).

Con un'altra versione del recomp non tocca niente. Quando disattivi l'opzione o disinstalli viene eliminata; se c'era un'altra `xinput9_1_0.dll` che non è nostra, non viene toccata. Ogni parte si può disattivare con una variabile d'ambiente (`SKATE3_DIRECT=0` la disattiva del tutto). Dettagli e codice in [`dll/`](dll/).

| | Prima | Ora |
|---|---|---|
| ⏩ Caricamento diretto completo | ~16,6 s | **~7 s** |

## ⚡ Caricamento

- 🛡️ **Windows Defender:** aggiunge la cartella del gioco alle esclusioni (chiede i permessi di amministratore una volta), così non controlla ogni file durante il caricamento.
- 🌅 **Precaricamento all'avvio del PC:** all'avvio di Windows legge in background, senza finestre, l'exe, le DLL, la cache degli shader e i file del gioco, così anche il primo caricamento della giornata è veloce.
- 💽 **SSD:** se il gioco è su un disco rigido e hai un SSD, lo copia in `X:\Skate3Recomp`, verifica la dimensione e sistema collegamenti, Steam e precaricamento; la vecchia cartella viene eliminata solo se lo chiedi.
- 📊 **Storico:** i tuoi ultimi 20 caricamenti rispetto ai 10 s e dove è andato il tempo in ognuno; l'ultimo compare accanto a **DROP IN**.
- 📜 **Log del caricamento:** un pulsante apre la cartella del gioco con `skate3-direct.log` selezionato.

## 🌐 Español ARG

Nella schermata delle lingue del gioco, "Nederlands" diventa **"Español ARG"**: i testi spagnoli del gioco riscritti con delle regole per suonare argentini (voseo e lessico rioplatense). "¿Quieres jugar? Puedes..." ➡️ "¿Querés jugar? Podés...", "Haz un kickflip y presiona Ⓐ" ➡️ "Hacé un kickflip y presioná Ⓐ", "colega" ➡️ "loco", "coches" ➡️ "autos", "piscina" ➡️ "pileta", "mola" ➡️ "garpa", e così via (circa 2700 testi). Lo spagnolo originale resta tale e quale. Le regole sono in [`go/rioplatense.txt`](go/rioplatense.txt): se vedi qualcosa di strano, si corregge lì.

Tocca solo `miscboot.big` (lo salva come `miscboot.big.original`); quando lo disattivi torna l'olandese, byte per byte.

## 🛹 Aiutante in sessione

Gira in background mentre giochi e si chiude insieme al gioco. Ogni funzione si attiva o disattiva in **Sessione**:

- ⌨️ **Alt+F4** chiude il gioco come la X della finestra.
- 🔁 **Ctrl+Alt+R** chiude e riapre il gioco se si blocca.
- ⚡ **Priorità alta** per il processo del gioco.
- ☀️ **Schermo sempre acceso**, anche se giochi solo col controller.
- 🖱️ **Cursore nascosto** sul gioco dopo 3 secondi senza muovere il mouse.
- 🪟 **Tasto Windows bloccato** mentre il gioco è in primo piano.
- 🗔 **Finestra ricordata**: monitor, dimensione e modalità vengono salvati alla chiusura.
- 🔕 **Non disturbare**: spegne le notifiche di Windows e le riaccende quando esci.

## 🔧 Officina

- 🖱️ **Collegamenti:** "Skate 3 Mods", "Skate 3 (directo)" e "Skate 3 (normal)" sul desktop e nel menu Start, più uno per ogni lingua.
- 🚂 **Steam:** Skate 3 nella tua libreria (Big Picture, Steam Input).
- ⏱️ **Tempo di gioco** per giorno e per settimana.
- 🧪 **Versioni del recomp:** più cartelle del gioco, ognuna con le sue mod e impostazioni.
- 🩹 **Ripara:** controlla `fedynamic.big` e `miscboot.big` e sistema quello che è rotto partendo dal backup.
- 🔄 **Dopo un aggiornamento del recomp:** se un file è cambiato, viene preso come nuovo originale e le tue mod vengono riapplicate.
- 🩺 **Diagnostica** e 📣 **segnalazione di problemi**, con la parte finale del log del gioco.
- ⬇️ **Aggiorna:** scarica la nuova versione da Releases e si sostituisce da solo.
- 💀 **Disinstalla tutto:** lascia il gioco esattamente com'era prima delle mod.

## 📻 Radio Carverton

Boom bap, punk, rock, electro e cumbia generati dal vivo con Web Audio, diversi in ogni brano e con uno nuovo ogni volta che apri il modder (scegli il genere o Mix, e casuale). Il tasto M la mette in pausa. Non usa file musicali né brani protetti da copyright.

## ⚙️ Come funziona

- **Icone e testi:** sostituiscono solo il proprio file dentro il `.big` (aggiungono la nuova versione in fondo); il resto rimane intatto.
- **Aiutante:** `Skate3-Mods.exe play` (lo usano DROP IN, i collegamenti, Steam e il `.bat`; `play direct` e `play normal` scelgono come parte) apre il gioco e lascia un piccolo processo senza finestra che si chiude insieme al gioco.
- **Modder:** una pagina HTML incorporata nel `.exe` (`go/ui/index.html`) dentro WebView2.
- 🚫 Non include né distribuisce file di EA: le icone sono grafica originale e le mod si applicano alla tua copia del gioco.

### 🐧 Linux / 🍎 macOS

Le icone funzionano con Python 3.8+ (senza dipendenze):

```sh
python3 src/skate3_ps_buttons.py install ps5 --game /ruta/a/skate3recomp
python3 src/skate3_ps_buttons.py uninstall    --game /ruta/a/skate3recomp
```

## 🛠️ Compilare

```sh
cd go && GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -H windowsgui" -o ../dist/Skate3-Mods.exe .
```

La DLL della patch: `cd dll && sh build.sh` (Linux, clang + lld). Quando si pubblica un tag `v*`, GitHub Actions compila il `.exe` e carica lo zip su Releases.

## 🖌️ Rigenerare le icone

```sh
pip install pillow numpy
python tools/build_payload.py preview/png   # ridisegna e rigenera src/payload.py e go/payload/
```

Font: Inter (SIL Open Font License, vedi `tools/fonts/Inter-LICENSE.txt`).

---

📄 Licenza in [`LICENSE`](LICENSE). Nessun legame con EA, Sony, Microsoft, Nintendo, Valve né con il progetto skate3recomp. Skate 3 è un marchio di Electronic Arts.
