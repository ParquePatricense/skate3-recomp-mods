"""Generates the Steam library artwork for the "Skate 3" non-Steam entry
(own sticker style, no EA images): go/steamart/{p,wide,hero}.jpg and logo.png."""
import random
from PIL import Image, ImageChops, ImageDraw, ImageFont, ImageFilter

FONT = "tools/fonts/Inter-ExtraBold.otf"
OUT = "go/steamart/"
ORANGE, LIME, INK = (255, 90, 31), (198, 255, 61), (238, 238, 238)

def asphalt(w, h):
    random.seed(3)
    im = Image.new("RGB", (w, h), (14, 14, 14))
    px = im.load()
    for _ in range(w * h // 6):
        x, y = random.randrange(w), random.randrange(h)
        v = random.randint(18, 40)
        px[x, y] = (v, v, v)
    g = Image.new("RGB", (w, h), (0, 0, 0))
    d = ImageDraw.Draw(g)
    d.ellipse((w * .45, -h * .5, w * 1.3, h * .6), fill=(120, 40, 10))
    d.ellipse((-w * .4, h * .6, w * .4, h * 1.4), fill=(50, 70, 15))
    g = g.filter(ImageFilter.GaussianBlur(min(w, h) / 5))
    return ImageChops.add(im, g)

def skew_text(text, size, color, shadow=True):
    f = ImageFont.truetype(FONT, size)
    l, t, r, b = f.getbbox(text)
    w, h = r - l + size // 2, b - t + size // 3
    layer = Image.new("RGBA", (w + size // 2, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    if shadow:
        d.text((size // 10 - l + size // 8, size // 10 - t + size // 12), text, font=f, fill=(0, 0, 0, 255))
    d.text((-l + size // 8, -t + size // 12), text, font=f, fill=color + (255,))
    k = .18  # italic slant
    return layer.transform((layer.width, layer.height), Image.AFFINE, (1, k, -k * layer.height, 0, 1, 0), Image.BICUBIC)

def sticker(text, size, bg, fg, angle):
    f = ImageFont.truetype(FONT, size)
    l, t, r, b = f.getbbox(text)
    pad = size // 3
    st = Image.new("RGBA", (r - l + 2 * pad + 8, b - t + 2 * pad + 8), (0, 0, 0, 0))
    d = ImageDraw.Draw(st)
    d.rectangle((8, 8, st.width - 1, st.height - 1), fill=(0, 0, 0, 255))
    d.rectangle((0, 0, st.width - 9, st.height - 9), fill=bg + (255,))
    d.text((pad - l, pad - t), text, font=f, fill=fg + (255,))
    return st.rotate(angle, expand=True, resample=Image.BICUBIC)

def wheel(r):
    w = Image.new("RGBA", (2 * r, 2 * r), (0, 0, 0, 0))
    d = ImageDraw.Draw(w)
    d.ellipse((0, 0, 2 * r - 1, 2 * r - 1), fill=(10, 10, 10, 255))
    d.ellipse((r * .12, r * .12, r * 1.88, r * 1.88), fill=LIME + (255,))
    d.ellipse((r * .55, r * .55, r * 1.45, r * 1.45), fill=(216, 248, 138, 255))
    d.ellipse((r * .78, r * .78, r * 1.22, r * 1.22), fill=(34, 34, 34, 255))
    return w

def paste_center(im, layer, cx, cy):
    im.alpha_composite(layer, (int(cx - layer.width / 2), int(cy - layer.height / 2)))

def logo_layer(scale):
    title = skew_text("SKATE 3", int(150 * scale), INK)
    tag = sticker("RECOMP MODS", int(46 * scale), ORANGE, (17, 17, 17), 4)
    lay = Image.new("RGBA", (max(title.width, tag.width) + 20, title.height + tag.height), (0, 0, 0, 0))
    paste_center(lay, title, lay.width / 2, title.height / 2)
    paste_center(lay, tag, lay.width / 2 + 30 * scale, title.height + tag.height / 2 - 18 * scale)
    return lay

def card(w, h, scale, wheel_r, wx, wy):
    im = asphalt(w, h).convert("RGBA")
    paste_center(im, wheel(wheel_r), w * wx, h * wy)
    paste_center(im, logo_layer(scale), w / 2, h * .56)
    return im.convert("RGB")

card(600, 900, .62, 110, .72, .24).save(OUT + "p.jpg", quality=88)
card(920, 430, .8, 60, .9, .2).save(OUT + "wide.jpg", quality=88)
hero = asphalt(1920, 620).convert("RGBA")
for x, y, r in ((1500, 170, 150), (330, 470, 90)):
    paste_center(hero, wheel(r), x, y)
hero.convert("RGB").save(OUT + "hero.jpg", quality=88)
logo = Image.new("RGBA", (1280, 720), (0, 0, 0, 0))
paste_center(logo, logo_layer(1.6), 640, 360)
logo.save(OUT + "logo.png", optimize=True)
