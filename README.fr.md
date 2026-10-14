# 🛹 Skate 3 Recomp Mods 🎮

[🇦🇷 Español](README.md) · [🇬🇧 English](README.en.md) · [🇧🇷 Português](README.pt.md) · **🇫🇷 Français** · [🇮🇹 Italiano](README.it.md) · [🇩🇪 Deutsch](README.de.md) · [🇨🇳 中文](README.zh.md) · [🇯🇵 日本語](README.ja.md)

Pack de mods **modulaire** pour [skate3recomp](https://github.com/mchughalex/skate3recomp) : tu télécharges tout d'un coup et tu actives seulement ce que tu veux. Chaque module s'active et se retire depuis **le modder**, une fenêtre façon Skate 3 avec sa propre radio, et tout est réversible.

> 🧩 **Nécessite [skate3recomp](https://github.com/mchughalex/skate3recomp) v2.0.2 et ta propre copie de Skate 3.** Ce n'est que le modder : il ne contient ni le jeu ni le recomp.

> ⚡ **De 16,6 s à ~7 s :** avec *Direct à ta sauvegarde*, un patch dans le jeu te fait rider dans ta sauvegarde en environ 7 secondes après avoir appuyé sur **DROP IN**.

![modder](preview/modeador.png)

## ✨ Ce qu'il y a dedans

| Module | Ce qu'il fait |
|---|---|
| 🎮 **Icônes des boutons** | Remplace les icônes Xbox 360 par celles de ta manette : PS3, PS4, PS5, Switch (1 et 2, Pro, Joy-Con), Xbox Series, GameSir ou Steam Deck, plus **noir et blanc** (contraste élevé) et **pour daltoniens** |
| 🖼️ **Préréglages graphiques** | Qualité maximale, Équilibré, Performances ou Ultra léger, en un clic |
| 🖥️ **Écran** | VSync, plein écran et sur quel écran s'ouvre le jeu (idéal avec deux écrans) |
| ⏩ **Direct à ta sauvegarde** | Un écran de chargement noir cache l'intro et les menus, un patch dans le jeu accélère tout le démarrage et tu arrives dans ta carrière dans la langue de ton choix |
| ⌨️ **Clavier et souris** | Active le mode clavier+souris du recomp, la sensibilité et la réassignation de chaque bouton en appuyant sur la touche |
| 🔭 **FOV et ultrawide** | Champ de vision (40–120) et mode 21:9 |
| 🛹 **Assistant en session** | Alt+F4, relance rapide (Ctrl+Alt+R), priorité haute, écran toujours allumé, curseur masqué, touche Windows bloquée, fenêtre mémorisée et ne pas déranger |
| ⚡ **Chargement** | Defender qui ne scanne plus le jeu, préchargement au démarrage du PC, déplacement du jeu vers un SSD, graphique de tes 20 derniers chargements, ton dernier chargement à côté de DROP IN et le journal du chargement en un clic |
| 🌐 **Español ARG** | Ajoute « Español ARG » (voseo et vocabulaire rioplatense) à la place du néerlandais ; l'espagnol d'origine reste tel quel |
| 🇧🇷 **Português (Brasil)** | Les textes du jeu en portugais pour le chargement direct (environ 13 600 textes) |
| 🔧 **L'atelier** | Raccourcis, Steam, temps de jeu, plusieurs versions du recomp, réparation, diagnostic, mise à jour et désinstallation complète |
| 🎮 **Modder à la manette** | Se pilote à la manette ou au clavier, s'adapte à ton écran et inclut une radio avec favoris, Mix et aléatoire |

🌍 Le modder est disponible en Español, English, Português (Brasil), Français, Italiano, Deutsch, 中文 et 日本語 : il choisit tout seul la langue de Windows et tu peux la changer dans **Langue**.

![aperçu](preview/preview.png)

| PS4 | PS5 |
|---|---|
| ![PS4 dans le jeu](preview/ingame-ps4.jpg) | ![PS5 dans le jeu](preview/ingame-ps5.jpg) |

## 📦 Installation (Windows)

1. ⬇️ Télécharge `Skate3-Mods.zip` depuis [Releases](../../releases) et décompresse-le où tu veux (Téléchargements, ça marche).
2. 🖱️ Ouvre `Skate3-Mods.exe` : il cherche tout seul le dossier de skate3recomp, s'y copie et s'en souvient. S'il ne le trouve pas, il te laisse le choisir.
3. ✅ Active ce que tu veux.
4. 🛹 Joue avec **DROP IN**, les raccourcis « Skate 3 (directo) » / « Skate 3 (normal) » ou `Jugar Skate 3 (mods).bat`. L'assistant en session ne tourne que si tu lances le jeu de cette façon ; les icônes, les graphismes et la langue marchent aussi en ouvrant `skate3.exe`.

> 🪟 Le modder utilise WebView2, déjà inclus dans Windows 10 et 11. S'il manquait, un menu en console s'ouvre (aussi avec `Skate3-Mods.exe menu`).

> ⚠️ Le jeu doit avoir été installé au moins une fois (bouton *Select ISO*). Ferme-le avant de changer les icônes ou les préréglages.

> 🛡️ Si SmartScreen apparaît : « Informations complémentaires » ➡️ « Exécuter quand même » (le .exe n'est pas signé).

## 🎮 Icônes des boutons

| 🟢 Xbox | 🔵 PlayStation |
|---|---|
| Ⓐ Ⓑ Ⓧ Ⓨ | ✖️ ⭕ 🟪 🔺 |
| LB / RB / LT / RT | L1 / R1 / L2 / R2 |
| LS / RS (clic) | L3 / R3 |
| Back / Start | PS3 : Select / Start · PS4 : Share / Options · PS5 : Create / Options |

Aussi : **Switch** (Switch 1 et 2, Pro Controller, Joy-Con), **Xbox Series** (Affichage / Menu), **GameSir** (Nova Lite, T4, G7 et la plupart des manettes PC) et **Steam Deck**. Pour une meilleure lisibilité : **Noir et blanc** (formes PlayStation noires sur blanc) et **Pour daltoniens** (lettres Xbox aux couleurs Okabe-Ito, sans rouge contre vert), avec un aperçu de leur rendu dans le jeu. Ceux de la Switch suivent la position : si tu utilises Steam avec « Utiliser la disposition des boutons Nintendo », A et B sont inversés.

Les icônes se trouvent dans `game/data/big/fedynamic.big` ➡️ `buttons.rx2`. Le mod garde l'original (`fedynamic.big.xbox-original`) et « Xbox (original) » le restaure octet par octet. ➕ Le D-Pad ne change pas et le dessin du stick dans le Trick Guide affiche toujours « RS » (c'est une autre texture).

## 🖼️ Préréglages, FOV et ultrawide

Ils écrivent des valeurs dans `skate3.toml` (le même fichier que celui du menu *Settings* du jeu, donc tu peux continuer à peaufiner là-bas ensuite). La première fois, ton `skate3.toml` est sauvegardé sous `skate3.toml.mods-original` ; « Remettre mes anciens réglages » le restaure.

**Ultra léger** désactive les ombres, le SSAO, le bloom, les rayons de soleil, le brouillard et le MSAA, et réduit la distance d'affichage : pour les PC vraiment modestes. **Écran** écrit `vsync`, `fullscreen` et `monitor` (0 = automatique, 1 = premier écran, 2 = deuxième...).

## ⌨️ Clavier et souris

Active `mnk_mode` du recomp : la souris bouge la caméra (stick droit) et chaque bouton de la manette correspond à une touche. Dans **Contrôles**, tu cliques sur un bouton et tu appuies sur la touche ou le clic que tu veux (si un autre bouton l'utilisait déjà, ils sont échangés). « Touches par défaut » remet celles du recomp.

## ⏩ Direct à ta sauvegarde

Utilise l'automatisation de démarrage du recomp lui-même (`skate3_demo_path` avec `skate3_demo_path_signed_in`, ce qui conserve ton profil et ta sauvegarde) : il passe l'intro, valide l'écran des langues, le Press Start et les avertissements de démarrage, puis rend les commandes dès que tu commences à rider. Pendant ce temps, un écran de chargement noir (avec une astuce affichant les boutons de ton style et une barre qui apprend combien de temps prend ton chargement) cache le jeu et disparaît dès que tu peux jouer ; le jeu reste en sourdine jusque-là. La langue choisie prend la place de l'anglais dans `miscboot.big` tant que l'option est active.

### 🧩 Le patch de chargement direct

Avec l'option active, le modder place `xinput9_1_0.dll` à côté de `skate3.exe`. Le recomp la charge au démarrage (elle transmet la manette au xinput de Windows, donc les commandes marchent pareil) et, **seulement si `skate3.exe` est la v2.0.2** (chaque patch est localisé via la table des fonctions de l'exe et vérifié octet par octet), elle accélère le démarrage dans le jeu lui-même :

- 🚫 **Logo EA** terminé en 0,03 s au lieu de ~4,7 s.
- 🏎️ **Pas de limite de FPS** pendant le chargement (ta valeur revient une fois en jeu).
- ⏱️ **Press Start** sans son attente minimale et avec START instantané.
- 😴 **Chargeurs sans sieste :** les threads de chargement arrêtent de dormir 1 ms entre chaque passe.
- 💓 **Tick du jeu à 4 ms** pendant les étapes courtes du chargement (au lieu de 16,7 ms).
- 🖤 **Fondus en une image** et la carte graphique en priorité haute jusqu'à l'arrivée en jeu.
- 📜 **`skate3-direct.log`** à côté de `skate3.exe` : la durée de chaque étape (s'ouvre depuis **Chargement**).

Avec une autre version du recomp, il ne touche à rien. Quand tu désactives l'option ou désinstalles, elle est supprimée ; s'il y avait une autre `xinput9_1_0.dll` qui n'est pas la nôtre, on n'y touche pas. Chaque partie peut être désactivée avec une variable d'environnement (`SKATE3_DIRECT=0` désactive tout). Détails et code dans [`dll/`](dll/).

| | Avant | Maintenant |
|---|---|---|
| ⏩ Chargement direct complet | ~16,6 s | **~7 s** |

## ⚡ Chargement

- 🛡️ **Windows Defender :** ajoute le dossier du jeu aux exclusions (demande les droits administrateur une fois), pour qu'il ne scanne pas chaque fichier au chargement.
- 🌅 **Préchargement au démarrage du PC :** au lancement de Windows, il lit en arrière-plan, sans fenêtre, l'exe, les DLL, le cache de shaders et les fichiers du jeu, pour que le premier chargement de la journée soit rapide lui aussi.
- 💽 **SSD :** si le jeu est sur un disque dur et que tu as un SSD, il le copie vers `X:\Skate3Recomp`, vérifie la taille et met à jour les raccourcis, Steam et le préchargement ; l'ancien dossier n'est supprimé que si tu le demandes.
- 📊 **Historique :** tes 20 derniers chargements face aux 10 s et où est passé le temps pour chacun ; le dernier s'affiche à côté de **DROP IN**.
- 📜 **Journal du chargement :** un bouton ouvre le dossier du jeu avec `skate3-direct.log` sélectionné.

## 🌐 Español ARG

Sur l'écran des langues du jeu, « Nederlands » devient **« Español ARG »** : les textes espagnols du jeu réécrits à l'aide de règles, à la sauce argentine (rioplatense). « ¿Quieres jugar? Puedes... » ➡️ « ¿Querés jugar? Podés... », « Haz un kickflip y presiona Ⓐ » ➡️ « Hacé un kickflip y presioná Ⓐ », « colega » ➡️ « loco », « coches » ➡️ « autos », « piscina » ➡️ « pileta », « mola » ➡️ « garpa », et ainsi de suite (environ 2700 textes). L'espagnol d'origine reste tel quel. Les règles sont dans [`go/rioplatense.txt`](go/rioplatense.txt) : si tu vois un truc bizarre, c'est là qu'on le corrige.

Seul `miscboot.big` est modifié (sauvegardé sous `miscboot.big.original`) ; en le désactivant, le néerlandais revient, octet par octet.

## 🛹 Assistant en session

Il tourne en arrière-plan pendant que tu joues et se ferme avec le jeu. Chaque fonction s'active ou se désactive dans **Session** :

- ⌨️ **Alt+F4** ferme le jeu comme le X de la fenêtre.
- 🔁 **Ctrl+Alt+R** ferme et relance le jeu s'il freeze.
- ⚡ **Priorité haute** pour le processus du jeu.
- ☀️ **Écran toujours allumé**, même si tu joues uniquement à la manette.
- 🖱️ **Curseur masqué** sur le jeu après 3 secondes sans bouger la souris.
- 🪟 **Touche Windows bloquée** tant que le jeu est au premier plan.
- 🗔 **Fenêtre mémorisée** : écran, taille et mode sont enregistrés à la fermeture.
- 🔕 **Ne pas déranger** : coupe les notifications de Windows et les réactive en quittant.

## 🔧 L'atelier

- 🖱️ **Raccourcis :** « Skate 3 Mods », « Skate 3 (directo) » et « Skate 3 (normal) » sur le bureau et dans le menu Démarrer, plus un par langue.
- 🚂 **Steam :** Skate 3 dans ta bibliothèque (Big Picture, Steam Input).
- ⏱️ **Temps de jeu** par jour et par semaine.
- 🧪 **Versions du recomp :** plusieurs dossiers du jeu, chacun avec ses propres mods et réglages.
- 🩹 **Réparer :** vérifie `fedynamic.big` et `miscboot.big` et répare ce qui est cassé à partir de la copie.
- 🔄 **Après une mise à jour du recomp :** si un fichier a changé, il devient le nouvel original et tes mods sont réappliqués.
- 🩺 **Diagnostic** et 📣 **signalement de problèmes**, avec la fin du journal du jeu.
- ⬇️ **Mise à jour :** télécharge la nouvelle version depuis Releases et se remplace tout seul.
- 💀 **Tout désinstaller :** remet le jeu exactement comme il était avant les mods.

## 📻 Radio Carverton

Boom bap, punk, rock, électro et cumbia générés en direct avec Web Audio, différents à chaque morceau, avec un nouveau morceau à chaque ouverture du modder (tu choisis le genre ou Mix, et aléatoire). La touche M la met en pause. Aucun fichier musical ni morceau sous copyright.

## ⚙️ Comment ça marche

- **Icônes et textes :** ils remplacent seulement leur fichier dans le `.big` (en ajoutant la nouvelle version à la fin) ; le reste n'est pas touché.
- **Assistant :** `Skate3-Mods.exe play` (utilisé par DROP IN, les raccourcis, Steam et le `.bat` ; `play direct` et `play normal` choisissent le mode de démarrage) lance le jeu et laisse un petit processus sans fenêtre qui se ferme avec le jeu.
- **Modder :** une page HTML intégrée au `.exe` (`go/ui/index.html`) dans WebView2.
- 🚫 N'inclut ni ne distribue aucun fichier d'EA : les icônes sont des créations originales et les mods s'appliquent à ta propre copie du jeu.

### 🐧 Linux / 🍎 macOS

Les icônes fonctionnent avec Python 3.8+ (sans dépendances) :

```sh
python3 src/skate3_ps_buttons.py install ps5 --game /chemin/vers/skate3recomp
python3 src/skate3_ps_buttons.py uninstall    --game /chemin/vers/skate3recomp
```

## 🛠️ Compiler

```sh
cd go && GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -H windowsgui" -o ../dist/Skate3-Mods.exe .
```

La DLL du patch : `cd dll && sh build.sh` (Linux, clang + lld). À la publication d'un tag `v*`, GitHub Actions compile le `.exe` et envoie le zip dans Releases.

## 🖌️ Régénérer les icônes

```sh
pip install pillow numpy
python tools/build_payload.py preview/png   # redessine et régénère src/payload.py et go/payload/
```

Police : Inter (SIL Open Font License, voir `tools/fonts/Inter-LICENSE.txt`).

---

📄 Licence dans [`LICENSE`](LICENSE). Aucun lien avec EA, Sony, Microsoft, Nintendo, Valve ni le projet skate3recomp. Skate 3 est une marque d'Electronic Arts.
