# 🛹 Skate 3 Recomp Mods 🎮

[🇦🇷 Español](README.md) · [🇬🇧 English](README.en.md) · **🇧🇷 Português** · [🇫🇷 Français](README.fr.md) · [🇮🇹 Italiano](README.it.md) · [🇩🇪 Deutsch](README.de.md) · [🇨🇳 中文](README.zh.md) · [🇯🇵 日本語](README.ja.md)

Pack de mods **modular** para o [skate3recomp](https://github.com/mchughalex/skate3recomp): você baixa tudo junto e ativa só o que quiser. Cada módulo é colocado e removido pelo **modder**, uma janela com a cara do Skate 3 e rádio própria, e tudo pode ser desfeito.

> 🧩 **Requer o [skate3recomp](https://github.com/mchughalex/skate3recomp) v2.0.2 e a sua própria cópia de Skate 3.** Isto é só o modder: não traz o jogo nem o recomp.

> ⚡ **De 16,6 s para ~7 s:** com *Direto para o seu save*, um patch dentro do jogo te deixa andando de skate no seu save em uns 7 segundos depois que você aperta **DROP IN**.

![modder](preview/modeador.png)

## ✨ O que vem

| Módulo | O que faz |
|---|---|
| 🎮 **Ícones de botões** | Troca os ícones do Xbox 360 pelos do seu controle: PS3, PS4, PS5, Switch (1 e 2, Pro, Joy-Con), Xbox Series, GameSir ou Steam Deck, além de **preto e branco** (alto contraste) e **para daltônicos** |
| 🖼️ **Presets gráficos** | Qualidade máxima, Equilibrado, Desempenho ou Ultraleve, com um clique |
| 🖥️ **Tela** | VSync, tela cheia e em qual monitor o jogo abre (ideal com dois monitores) |
| ⏩ **Direto para o seu save** | Uma tela de carregamento preta cobre a intro e os menus, um patch dentro do jogo acelera toda a inicialização e você aparece na sua carreira no idioma que escolher |
| ⌨️ **Teclado e mouse** | Ativa o modo teclado+mouse do recomp, sensibilidade e remapeamento de cada botão apertando a tecla |
| 🔭 **FOV e ultrawide** | Campo de visão (40–120) e modo 21:9 |
| 🛹 **Ajudante na sessão** | Alt+F4, reinício rápido (Ctrl+Alt+R), prioridade alta, tela sempre ligada, cursor oculto, tecla Windows bloqueada, janela lembrada e não perturbe |
| ⚡ **Carga** | Defender sem verificar o jogo, pré-carregamento ao ligar o PC, mover o jogo para um SSD, gráfico dos seus últimos 20 carregamentos, seu último carregamento ao lado do DROP IN e o log do carregamento a um clique |
| 🌐 **Español ARG** | Adiciona o "Español ARG" (voseo e vocabulário rioplatense) no lugar do holandês; o espanhol original continua igual |
| 🇧🇷 **Português (Brasil)** | Os textos do jogo em português para o carregamento direto (uns 13.600 textos) |
| 🔧 **A oficina** | Atalhos, Steam, tempo de jogo, várias versões do recomp, reparar, diagnóstico, atualizar e desinstalar tudo |
| 🎮 **Modder com controle** | Dá pra usar com o controle ou com o teclado, se adapta à sua tela e tem rádio com favoritas, Mix e aleatório |

🌍 O modder está em Español, English, Português (Brasil), Français, Italiano, Deutsch, 中文 e 日本語: ele escolhe sozinho o idioma do Windows e dá pra trocar em **Idioma**.

![preview](preview/preview.png)

| PS4 | PS5 |
|---|---|
| ![PS4 no jogo](preview/ingame-ps4.jpg) | ![PS5 no jogo](preview/ingame-ps5.jpg) |

## 📦 Instalação (Windows)

1. ⬇️ Baixe o `Skate3-Mods.zip` em [Releases](../../releases) e descompacte onde quiser (Downloads serve).
2. 🖱️ Abra o `Skate3-Mods.exe`: ele procura sozinho a pasta do skate3recomp, se copia para lá e lembra dela. Se não encontrar, deixa você escolher.
3. ✅ Ative o que quiser.
4. 🛹 Jogue com **DROP IN**, os atalhos "Skate 3 (directo)" / "Skate 3 (normal)" ou `Jugar Skate 3 (mods).bat`. O ajudante na sessão só roda se você abrir o jogo assim; ícones, gráficos e idioma funcionam do mesmo jeito abrindo o `skate3.exe`.

> 🪟 O modder usa o WebView2, que já vem com o Windows 10 e 11. Se faltar, abre um menu de console (também com `Skate3-Mods.exe menu`).

> ⚠️ O jogo precisa ter sido instalado pelo menos uma vez (botão *Select ISO*). Feche o jogo antes de trocar ícones ou presets.

> 🛡️ Se aparecer o SmartScreen: "Mais informações" ➡️ "Executar assim mesmo" (o .exe não é assinado).

## 🎮 Ícones de botões

| 🟢 Xbox | 🔵 PlayStation |
|---|---|
| Ⓐ Ⓑ Ⓧ Ⓨ | ✖️ ⭕ 🟪 🔺 |
| LB / RB / LT / RT | L1 / R1 / L2 / R2 |
| LS / RS (clique) | L3 / R3 |
| Back / Start | PS3: Select / Start · PS4: Share / Options · PS5: Create / Options |

Também: **Switch** (Switch 1 e 2, Pro Controller, Joy-Con), **Xbox Series** (Exibir / Menu), **GameSir** (Nova Lite, T4, G7 e a maioria dos controles de PC) e **Steam Deck**. Para ler melhor: **Preto e branco** (formas do PlayStation pretas no branco) e **Para daltônicos** (letras do Xbox nas cores Okabe-Ito, sem vermelho contra verde), com uma prévia de como ficam no jogo. Os do Switch seguem a posição: se você usa a Steam com "Usar disposição de botões da Nintendo", A e B ficam trocados.

Os ícones ficam em `game/data/big/fedynamic.big` ➡️ `buttons.rx2`. O mod guarda o original (`fedynamic.big.xbox-original`) e "Xbox (original)" o restaura byte a byte. ➕ O D-Pad continua igual e o desenho do analógico no Trick Guide continua dizendo "RS" (é outra textura).

## 🖼️ Presets, FOV e ultrawide

Gravam valores no `skate3.toml` (o mesmo arquivo que o menu *Settings* do jogo usa, então depois você pode continuar ajustando por lá). Na primeira vez, seu `skate3.toml` é salvo como `skate3.toml.mods-original`; "Voltar para minhas configurações anteriores" o restaura.

**Ultraleve** desliga sombras, SSAO, bloom, raios de sol, neblina e MSAA, e encurta a distância de renderização: para PCs bem fracos. **Tela** grava `vsync`, `fullscreen` e `monitor` (0 = automático, 1 = primeira tela, 2 = segunda...).

## ⌨️ Teclado e mouse

Ativa o `mnk_mode` do recomp: o mouse mexe a câmera (analógico direito) e cada botão do controle vai para uma tecla. Em **Controles** você toca num botão e aperta a tecla ou o clique que quiser (se outro botão já usava essa tecla, eles trocam). "Voltar para as teclas padrão" deixa as do recomp.

## ⏩ Direto para o seu save

Usa a automação de inicialização do próprio recomp (`skate3_demo_path` com `skate3_demo_path_signed_in`, assim mantém seu perfil e seu save): pula a intro, aceita a tela de idioma, o Press Start e os avisos de inicialização, e solta os controles assim que você começa a andar. Enquanto isso, uma tela de carregamento preta (com uma dica com os botões do seu estilo e uma barra que aprende quanto tempo seu carregamento leva) cobre o jogo e some assim que você pode jogar; o jogo fica no mudo até lá. O idioma que você escolher entra no lugar do inglês dentro do `miscboot.big` enquanto a opção estiver ativa.

### 🧩 O patch do carregamento direto

Com a opção ativa, o modder coloca a `xinput9_1_0.dll` junto do `skate3.exe`. O recomp a carrega ao iniciar (ela repassa o controle para o xinput do Windows, então os controles funcionam igual) e, **só se o `skate3.exe` for a v2.0.2** (cada patch é localizado pela tabela de funções do exe e verificado byte a byte), acelera a inicialização dentro do próprio jogo:

- 🚫 **Logo da EA** encerrado em 0,03 s em vez de ~4,7 s.
- 🏎️ **Sem limite de FPS** durante o carregamento (seu valor volta quando você chega no jogo).
- ⏱️ **Press Start** sem a espera mínima e com START na hora.
- 😴 **Carregadores sem soneca:** as threads de carregamento param de dormir 1 ms entre cada passada.
- 💓 **Tick do jogo a 4 ms** nas etapas curtas do carregamento (em vez de 16,7 ms).
- 🖤 **Fades em um quadro** e a placa de vídeo com prioridade alta até chegar no jogo.
- 📜 **`skate3-direct.log`** junto do `skate3.exe`: quanto tempo cada etapa levou (abre pela aba **Carga**).

Com outra versão do recomp, não mexe em nada. Ao desligar a opção ou desinstalar, ela é apagada; se havia outra `xinput9_1_0.dll` que não é nossa, ela não é tocada. Cada parte pode ser desligada com uma variável de ambiente (`SKATE3_DIRECT=0` desliga tudo). Detalhes e código em [`dll/`](dll/).

| | Antes | Agora |
|---|---|---|
| ⏩ Carregamento direto completo | ~16,6 s | **~7 s** |

## ⚡ Carga

- 🛡️ **Windows Defender:** adiciona a pasta do jogo às exclusões (pede permissão de administrador uma vez), para ele não verificar cada arquivo no carregamento.
- 🌅 **Pré-carregar ao ligar o PC:** quando o Windows inicia, lê em segundo plano, sem janelas, o exe, as DLLs, o cache de shaders e os arquivos do jogo, assim o primeiro carregamento do dia também é rápido.
- 💽 **SSD:** se o jogo está num HD e você tem um SSD, copia para `X:\Skate3Recomp`, verifica o tamanho e reajusta atalhos, Steam e pré-carregamento; a pasta antiga só é apagada se você pedir.
- 📊 **Histórico:** seus últimos 20 carregamentos contra os 10 s e onde o tempo de cada um foi gasto; o último aparece ao lado do **DROP IN**.
- 📜 **Log do carregamento:** um botão abre a pasta do jogo com o `skate3-direct.log` selecionado.

## 🌐 Español ARG

Na tela de idiomas do jogo, "Nederlands" vira **"Español ARG"**: os textos em espanhol do jogo reescritos com regras para o espanhol rioplatense (voseo e gírias argentinas). "¿Quieres jugar? Puedes..." ➡️ "¿Querés jugar? Podés...", "Haz un kickflip y presiona Ⓐ" ➡️ "Hacé un kickflip y presioná Ⓐ", "colega" ➡️ "loco", "coches" ➡️ "autos", "piscina" ➡️ "pileta", "mola" ➡️ "garpa", e por aí vai (uns 2700 textos). O espanhol original continua do jeito que está. As regras estão em [`go/rioplatense.txt`](go/rioplatense.txt): se você vir algo estranho, é ali que se corrige.

Mexe só no `miscboot.big` (guarda como `miscboot.big.original`); ao desativar, o holandês volta, byte a byte.

## 🛹 Ajudante na sessão

Roda em segundo plano enquanto você joga e fecha junto com o jogo. Cada coisa é ligada ou desligada em **Na sessão**:

- ⌨️ **Alt+F4** fecha o jogo igual ao X da janela.
- 🔁 **Ctrl+Alt+R** fecha e abre o jogo de novo se ele travar.
- ⚡ **Prioridade alta** para o processo do jogo.
- ☀️ **Tela sempre ligada**, mesmo se você jogar só com o controle.
- 🖱️ **Cursor oculto** sobre o jogo depois de 3 segundos sem mexer o mouse.
- 🪟 **Tecla Windows bloqueada** enquanto o jogo está em primeiro plano.
- 🗔 **Janela lembrada**: monitor, tamanho e modo são salvos ao fechar.
- 🔕 **Não perturbe**: desliga as notificações do Windows e liga de novo quando você sai.

## 🔧 A oficina

- 🖱️ **Atalhos:** "Skate 3 Mods", "Skate 3 (directo)" e "Skate 3 (normal)" na área de trabalho e no menu Iniciar, e também um por idioma.
- 🚂 **Steam:** o Skate 3 na sua biblioteca (Big Picture, Steam Input).
- ⏱️ **Tempo de jogo** por dia e por semana.
- 🧪 **Versões do recomp:** várias pastas do jogo, cada uma com seus próprios mods e ajustes.
- 🩹 **Reparar:** verifica o `fedynamic.big` e o `miscboot.big` e conserta o que estiver quebrado usando o backup.
- 🔄 **Depois de atualizar o recomp:** se um arquivo mudou, ele vira o novo original e seus mods são colocados de novo.
- 🩺 **Diagnóstico** e 📣 **relatório de problemas**, com o final do log do jogo.
- ⬇️ **Atualizar:** baixa a versão nova em Releases e se substitui sozinho.
- 💀 **Desinstalar tudo:** deixa o jogo exatamente como era antes dos mods.

## 📻 Rádio Carverton

Boom bap, punk, rock, electro e cumbia gerados ao vivo com Web Audio, diferentes a cada música e com uma nova cada vez que você abre o modder (você escolhe o gênero ou Mix, e aleatório). A tecla M pausa. Não usa arquivos de música nem músicas com copyright.

## ⚙️ Como funciona

- **Ícones e textos:** substituem só o próprio arquivo dentro do `.big` (adicionam a versão nova no final); o resto fica intacto.
- **Ajudante:** `Skate3-Mods.exe play` (usado pelo DROP IN, pelos atalhos, pela Steam e pelo `.bat`; `play direct` e `play normal` escolhem como inicia) abre o jogo e deixa um processo pequeno sem janela que fecha junto com o jogo.
- **Modder:** uma página HTML embutida no `.exe` (`go/ui/index.html`) dentro do WebView2.
- 🚫 Não inclui nem distribui arquivos da EA: os ícones são arte original e os mods são aplicados sobre a sua própria cópia do jogo.

### 🐧 Linux / 🍎 macOS

Os ícones funcionam com Python 3.8+ (sem dependências):

```sh
python3 src/skate3_ps_buttons.py install ps5 --game /caminho/para/skate3recomp
python3 src/skate3_ps_buttons.py uninstall    --game /caminho/para/skate3recomp
```

## 🛠️ Compilar

```sh
cd go && GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -H windowsgui" -o ../dist/Skate3-Mods.exe .
```

A DLL do patch: `cd dll && sh build.sh` (Linux, clang + lld). Ao publicar uma tag `v*`, o GitHub Actions compila o `.exe` e sobe o zip em Releases.

## 🖌️ Regenerar os ícones

```sh
pip install pillow numpy
python tools/build_payload.py preview/png   # redesenha e regenera src/payload.py e go/payload/
```

Fonte: Inter (SIL Open Font License, veja `tools/fonts/Inter-LICENSE.txt`).

---

📄 Licença em [`LICENSE`](LICENSE). Sem relação com a EA, Sony, Microsoft, Nintendo, Valve nem com o projeto skate3recomp. Skate 3 é marca da Electronic Arts.
