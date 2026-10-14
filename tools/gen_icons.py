"""Draws the button icon sets (original art) for Skate 3's buttons.rx2:
PlayStation (PS3/PS4/PS5), Nintendo Switch (Switch 1/2, Pro, Joy-Con), Xbox
Series, GameSir and Steam Deck.

Each texture is 64x64; the game samples the icon from the top-left 48x48 area
(icon centre at 24,24), same as the original Xbox art. `_hud` variants get an
extra dark outline, matching the original set.

Usage: python gen_icons.py <out_dir>
"""
import math, os, sys
from PIL import Image, ImageDraw, ImageFont, ImageFilter

S = 8                     # supersampling factor
N = 64 * S
C = 24 * S                # icon centre
# Inter (SIL OFL). Override with S3PS_FONT=/path/to/Inter-ExtraBold.otf
def _find_font():
    env = os.environ.get('S3PS_FONT')
    cands = [env] if env else []
    cands += [os.path.join(os.path.dirname(__file__), 'fonts', 'Inter-ExtraBold.otf'),
              '/usr/share/fonts/opentype/inter/Inter-ExtraBold.otf']
    for c in cands:
        if c and os.path.isfile(c):
            return c
    raise SystemExit('Inter-ExtraBold.otf not found; set S3PS_FONT')
FONT = FONT_TXT = _find_font()
DARK = (40, 40, 42, 255)
WHITE = (245, 245, 245, 255)
OUTLINE = (24, 24, 26, 255)

NAMES = """A A_hud B B_hud Back Back_hud DPad DPad_Arrows DPad_Arrows_hud DPad_Down DPad_Down_hud
DPad_hud DPad_Left DPad_Left_hud DPad_Right DPad_Right_hud DPad_Up DPad_Up_hud DPad_UpDown DPad_UpDown_hud
L3_Click L3_Click_hud L_Bumper L_Bumper_hud L_Trigger L_Trigger_hud R3_Click R3_Click_hud R_Bumper R_Bumper_hud
R_Trigger R_Trigger_hud Start Start_hud Stick_Left Stick_Left_Down Stick_Left_Down_hud Stick_Left_hud
Stick_Left_L Stick_Left_L_hud Stick_Left_L_R Stick_Left_L_R_hud Stick_Left_R Stick_Left_R_hud Stick_Left_Up
Stick_Left_Up_down Stick_Left_Up_down_hud Stick_Left_Up_hud Stick_Right Stick_Right_Down Stick_Right_Down_hud
Stick_Right_hud Stick_Right_L Stick_Right_L_hud Stick_Right_L_R Stick_Right_L_R_hud Stick_Right_R
Stick_Right_R_hud Stick_Right_Up Stick_Right_Up_down Stick_Right_Up_down_hud Stick_Right_Up_hud X X_hud Y Y_hud""".split()
assert len(NAMES) == 66

STYLES = {
    'ps3': dict(disc=(12, 12, 14, 255), rim=(150, 150, 156, 255),
                sym={'cross': (122, 168, 230), 'circle': (240, 84, 88), 'square': (230, 140, 196), 'triangle': (58, 196, 160)}),
    'ps4': dict(disc=(30, 30, 34, 255), rim=None,
                sym={'cross': (140, 180, 245), 'circle': (255, 102, 102), 'square': (245, 160, 210), 'triangle': (64, 226, 160)}),
    'ps5': dict(disc=(36, 36, 40, 255), rim=None,
                sym={k: (232, 232, 236) for k in ('cross', 'circle', 'square', 'triangle')}),
    # black and white: black PlayStation shapes on a white disc, the most
    # contrast for a TV across the room or low vision
    'bn': dict(disc=(246, 246, 246, 255), rim=(20, 20, 22, 255),
               sym={k: (16, 16, 18) for k in ('cross', 'circle', 'square', 'triangle')}),
}
# Letter styles. The game asks for the pad's A/B/X/Y by position (A = bottom),
# so a Nintendo pad shows its own letter for each position: B bottom, A right,
# Y left, X top.
LETTERS = {
    'switch': dict(disc=(34, 34, 38, 255), face={'A': 'B', 'B': 'A', 'X': 'Y', 'Y': 'X'},
                   col={k: (240, 240, 244) for k in 'ABXY'}, bump=('L', 'R'), trig=('ZL', 'ZR'), menu='nintendo', click=('LS', 'RS')),
    'xseries': dict(disc=(30, 30, 34, 255), face={k: k for k in 'ABXY'},
                    col={'A': (110, 205, 90), 'B': (240, 80, 70), 'X': (70, 150, 245), 'Y': (250, 200, 50)},
                    bump=('LB', 'RB'), trig=('LT', 'RT'), menu='xbox', click=('LS', 'RS')),
    'gamesir': dict(disc=(26, 26, 30, 255), face={k: k for k in 'ABXY'}, col={k: (236, 236, 240) for k in 'ABXY'},
                    bump=('LB', 'RB'), trig=('LT', 'RT'), menu='xbox', click=('LS', 'RS'), rim=(255, 90, 31, 255)),
    'steamdeck': dict(disc=(52, 56, 64, 255), face={k: k for k in 'ABXY'}, col={k: (236, 238, 242) for k in 'ABXY'},
                      bump=('L1', 'R1'), trig=('L2', 'R2'), menu='xbox', click=('L3', 'R3')),
    # colour-blind safe: Xbox letters in Okabe-Ito colours (no red against
    # green; they also differ in lightness), so letter and colour both tell
    'daltonico': dict(disc=(24, 24, 28, 255), face={k: k for k in 'ABXY'},
                      col={'A': (86, 180, 233), 'B': (230, 159, 0), 'X': (204, 121, 167), 'Y': (240, 228, 66)},
                      bump=('LB', 'RB'), trig=('LT', 'RT'), menu='xbox', click=('LS', 'RS'), rim=(235, 235, 240, 255)),
}
for _k in LETTERS:
    STYLES[_k] = LETTERS[_k]
FACE = {'A': 'cross', 'B': 'circle', 'X': 'square', 'Y': 'triangle'}

def letter_face(style, btn):
    st = LETTERS[style]
    im, d = canvas()
    disc(d, 12, st['disc'], st.get('rim'), 1.4)
    text_c(d, C, C + s(0.6), st['face'][btn], 15, st['col'][btn] + (255,))
    return im

def canvas():
    im = Image.new('RGBA', (N, N), (0, 0, 0, 0))
    return im, ImageDraw.Draw(im)

def s(v):
    return int(round(v * S))

def font(px, path=FONT):
    return ImageFont.truetype(path, s(px))

def text_c(d, cx, cy, txt, px, fill, path=FONT, spacing=0.0):
    f = font(px, path)
    if spacing:
        # manual letter spacing (negative = tighter)
        widths = [d.textlength(ch, font=f) for ch in txt]
        total = sum(widths) + s(spacing) * (len(txt) - 1)
        x = cx - total / 2
        for ch, w in zip(txt, widths):
            d.text((x, cy), ch, font=f, fill=fill, anchor='lm')
            x += w + s(spacing)
    else:
        d.text((cx, cy), txt, font=f, fill=fill, anchor='mm')

def disc(d, r, fill, rim=None, rim_w=1.2):
    if rim:
        d.ellipse([C - s(r), C - s(r), C + s(r), C + s(r)], fill=rim)
        r -= rim_w
    d.ellipse([C - s(r), C - s(r), C + s(r), C + s(r)], fill=fill)

def face(style, which):
    st = STYLES[style]
    im, d = canvas()
    disc(d, 12, st['disc'], st['rim'])
    col = st['sym'][which] + (255,)
    w = s(2.3)
    k = 5.6
    if which == 'cross':
        d.line([C - s(k), C - s(k), C + s(k), C + s(k)], fill=col, width=w)
        d.line([C - s(k), C + s(k), C + s(k), C - s(k)], fill=col, width=w)
    elif which == 'circle':
        r = s(6.2)
        d.ellipse([C - r, C - r, C + r, C + r], outline=col, width=w)
    elif which == 'square':
        r = s(5.4)
        d.rectangle([C - r, C - r, C + r, C + r], outline=col, width=w)
    elif which == 'triangle':
        r = 7.2
        pts = [(C + s(r * math.cos(math.radians(a))), C + s(0.9) + s(r * math.sin(math.radians(a)))) for a in (-90, 30, 150)]
        d.line(pts + [pts[0], pts[1]], fill=col, width=w, joint='curve')
    return im

def pill(d, w, h, fill):
    d.rounded_rectangle([C - s(w / 2), C - s(h / 2), C + s(w / 2), C + s(h / 2)], radius=s(h / 2), fill=fill)

def menu_button(style, which):
    """which: 'back' or 'start'."""
    im, d = canvas()
    if style == 'ps3':
        pill(d, 26, 16, WHITE)
        if which == 'back':   # SELECT: flat bar
            d.rounded_rectangle([C - s(5.5), C - s(2.2), C + s(5.5), C + s(2.2)], radius=s(1), fill=DARK)
        else:                 # START: right triangle
            d.polygon([(C - s(3.5), C - s(5)), (C - s(3.5), C + s(5)), (C + s(5.5), C)], fill=DARK)
    elif style == 'ps4':
        txt = 'SHARE' if which == 'back' else 'OPTIONS'
        pill(d, 44, 16, WHITE)
        text_c(d, C, C + s(0.4), txt, 9.4 if which == 'back' else 8.4, DARK, spacing=-0.25)
    else:  # ps5: create (rays) / options (three lines)
        pill(d, 26, 16, WHITE)
        if which == 'back':
            for ang in (-140, -90, -40):
                a = math.radians(ang)
                d.line([C + s(2.4 * math.cos(a)), C + s(3.8) + s(2.4 * math.sin(a)),
                        C + s(8 * math.cos(a)), C + s(3.8) + s(8 * math.sin(a))], fill=DARK, width=s(2.1))
        else:
            for dy in (-3.4, 0, 3.4):
                d.rounded_rectangle([C - s(5.5), C + s(dy) - s(0.9), C + s(5.5), C + s(dy) + s(0.9)], radius=s(0.9), fill=DARK)
    return im

def letter_menu(style, which):
    """Xbox View / Menu, or Nintendo - / +."""
    im, d = canvas()
    if LETTERS[style]['menu'] == 'nintendo':
        disc(d, 11, WHITE)
        w = s(2.6)
        d.line([C - s(5.5), C, C + s(5.5), C], fill=DARK, width=w)
        if which == 'start':
            d.line([C, C - s(5.5), C, C + s(5.5)], fill=DARK, width=w)
        return im
    disc(d, 11, WHITE)
    if which == 'back':   # View: two overlapping windows
        d.rectangle([C - s(5.5), C - s(4.5), C + s(1.5), C + s(1.5)], outline=DARK, width=s(1.6))
        d.rectangle([C - s(1.5), C - s(1.5), C + s(5.5), C + s(4.5)], fill=WHITE, outline=DARK, width=s(1.6))
    else:                 # Menu: three lines
        for dy in (-3.4, 0, 3.4):
            d.rounded_rectangle([C - s(5.5), C + s(dy) - s(0.9), C + s(5.5), C + s(dy) + s(0.9)], radius=s(0.9), fill=DARK)
    return im

def shoulder(side, kind, label=None):
    """kind 1 = bumper (L1/R1), 2 = trigger (L2/R2)."""
    im, d = canvas()
    label = label or side + str(kind)
    px = 14 if len(label) <= 2 else 11
    if kind == 1:
        d.rounded_rectangle([C - s(15), C - s(9), C + s(15), C + s(9)], radius=s(4), fill=WHITE)
        text_c(d, C, C + s(0.3), label, px, DARK)
    else:
        # trigger silhouette: tall rounded top, flat bottom
        top, bot = C - s(11), C + s(10)
        d.rounded_rectangle([C - s(14), top, C + s(14), bot], radius=s(9), fill=WHITE)
        d.rectangle([C - s(14), C, C + s(14), bot], fill=WHITE)
        text_c(d, C, C + s(1.2), label, px, DARK)
    return im

def chevron(d, direction):
    """Arrow chevrons outside the stick disc (disc spans 10..38)."""
    L, w = 4.5, s(2.6)
    if direction == 'L':
        cx, cy, pts = 5.0, 24.0, [(2.2, -L), (-2.2, 0), (2.2, L)]
    elif direction == 'R':
        cx, cy, pts = 43.0, 24.0, [(-2.2, -L), (2.2, 0), (-2.2, L)]
    elif direction == 'U':
        cx, cy, pts = 24.0, 5.0, [(-L, 2.2), (0, -2.2), (L, 2.2)]
    else:
        cx, cy, pts = 24.0, 43.0, [(-L, -2.2), (0, 2.2), (L, -2.2)]
    p = [(s(cx + x), s(cy + y)) for x, y in pts]
    d.line(p, fill=WHITE, width=w, joint='curve')
    for q in (p[0], p[2]):
        d.ellipse([q[0] - w / 2, q[1] - w / 2, q[0] + w / 2, q[1] + w / 2], fill=WHITE)

def stick(side, dirs, label=None):
    im, d = canvas()
    disc(d, 14, DARK)
    text_c(d, C, C + s(0.6), label or side, 16 if not label else 13, WHITE)
    for dr in dirs:
        chevron(d, dr)
    return im

def finish(im, hud):
    small = im.resize((64, 64), Image.BOX)
    if hud:
        a = small.split()[3].filter(ImageFilter.MaxFilter(3))
        base = Image.new('RGBA', (64, 64), OUTLINE)
        base.putalpha(a.point(lambda v: min(255, int(v * 1.0))))
        base.alpha_composite(small)
        small = base
    # make fully transparent texels neutral to avoid fringe colours
    px = small.load()
    for y in range(64):
        for x in range(64):
            if px[x, y][3] == 0:
                px[x, y] = (0, 0, 0, 0)
    return small

STICK_DIRS = {'': '', 'Down': 'D', 'L': 'L', 'L_R': 'LR', 'R': 'R', 'Up': 'U', 'Up_down': 'UD'}

def build(style):
    """Returns {index: 64x64 RGBA Image} for every texture this mod replaces."""
    out = {}
    for i, name in enumerate(NAMES):
        hud = name.endswith('_hud')
        base = name[:-4] if hud else name
        im = None
        lt = LETTERS.get(style)
        if base in FACE:
            im = letter_face(style, base) if lt else face(style, FACE[base])
        elif base == 'Back':
            im = letter_menu(style, 'back') if lt else menu_button(style, 'back')
        elif base == 'Start':
            im = letter_menu(style, 'start') if lt else menu_button(style, 'start')
        elif base in ('L_Bumper', 'R_Bumper', 'L_Trigger', 'R_Trigger'):
            kind = 1 if 'Bumper' in base else 2
            im = shoulder(base[0], kind, lt and lt['bump' if kind == 1 else 'trig'][base[0] == 'R'])
        elif base in ('L3_Click', 'R3_Click'):
            im = stick(base[0], '', label=lt['click'][base[0] == 'R'] if lt else base[:2])
        elif base.startswith('Stick_'):
            side = 'L' if base.startswith('Stick_Left') else 'R'
            rest = base.split('_', 2)[2] if base.count('_') >= 2 else ''
            im = stick(side, STICK_DIRS[rest])
        if im is not None:
            out[i] = finish(im, hud)
    return out

if __name__ == '__main__':
    od = sys.argv[1]
    for style in STYLES:
        os.makedirs(os.path.join(od, style), exist_ok=True)
        for i, im in build(style).items():
            im.save(os.path.join(od, style, f'{i:02d}_{NAMES[i]}.png'))
