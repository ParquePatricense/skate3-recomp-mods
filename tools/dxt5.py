"""Minimal DXT5 (BC3) encoder/decoder with numpy, good enough for flat UI icons."""
import numpy as np

def _565(c):
    c = np.clip(np.round(c), 0, 255).astype(int)
    return int(((c[0] >> 3) << 11) | ((c[1] >> 2) << 5) | (c[2] >> 3))

def _unpack565(v):
    r = (v >> 11) & 31; g = (v >> 5) & 63; b = v & 31
    return np.array([(r << 3) | (r >> 2), (g << 2) | (g >> 4), (b << 3) | (b >> 2)], float)

def _alpha_block(a):
    a = a.astype(int)
    a0, a1 = int(a.max()), int(a.min())
    if a0 == a1:
        return bytes([a0, a1]) + b'\0' * 6
    pal = [a0, a1] + [((7 - i) * a0 + i * a1) // 7 for i in range(1, 7)]
    pal = np.array(pal)
    idx = np.abs(a[:, None] - pal[None, :]).argmin(1)
    bits = 0
    for i, v in enumerate(idx):
        bits |= int(v) << (3 * i)
    return bytes([a0, a1]) + bits.to_bytes(6, 'little')

def _try(rgb, w, c0, c1):
    e0, e1 = _565(c0), _565(c1)
    if e0 < e1:
        e0, e1 = e1, e0
    if e0 == e1:
        pal = np.array([_unpack565(e0)])
        idx = np.zeros(16, int)
    else:
        p0, p1 = _unpack565(e0), _unpack565(e1)
        pal = np.array([p0, p1, (2 * p0 + p1) / 3, (p0 + 2 * p1) / 3])
        idx = ((rgb[:, None, :] - pal[None]) ** 2).sum(2).argmin(1)
    err = (((rgb - pal[idx]) ** 2).sum(1) * (w + 1)).sum()
    bits = 0
    for i, v in enumerate(idx):
        bits |= int(v) << (2 * i)
    return err, e0.to_bytes(2, 'little') + e1.to_bytes(2, 'little') + bits.to_bytes(4, 'little')

def _color_block(rgb, w):
    m = w > 8
    pts = rgb[m] if m.any() else rgb
    mean = pts.mean(0)
    cands = [(pts.max(0), pts.min(0))]
    lum = pts @ np.array([0.3, 0.59, 0.11])
    cands.append((pts[lum.argmax()], pts[lum.argmin()]))
    if len(pts) > 1:
        axis = np.linalg.eigh(np.cov((pts - mean).T))[1][:, -1]
        proj = (pts - mean) @ axis
        cands.append((mean + axis * proj.max(), mean + axis * proj.min()))
    wm = np.where(m, w, 0) if m.any() else w
    best = min((_try(rgb, wm, a, b) + (a, b) for a, b in cands), key=lambda t: t[0])
    # least-squares refinement of the best endpoints (a couple of rounds)
    a, b = np.array(best[2], float), np.array(best[3], float)
    for _ in range(3):
        d = b - a
        dd = d @ d
        if dd < 1e-6:
            break
        t = np.clip(np.round(((rgb - a) @ d) / dd * 3) / 3, 0, 1)
        sw = wm + 1
        A = np.stack([1 - t, t], 1) * np.sqrt(sw)[:, None]
        sol, *_ = np.linalg.lstsq(A, rgb * np.sqrt(sw)[:, None], rcond=None)
        a, b = sol[0], sol[1]
        cand = _try(rgb, wm, a, b) + (a, b)
        if cand[0] < best[0]:
            best = cand
    return best[1]

def encode(img):
    """img: HxWx4 uint8 RGBA (H, W multiples of 4). Returns linear BC3 blocks (little endian, PC layout)."""
    h, w = img.shape[:2]
    out = bytearray()
    f = img.astype(float)
    for by in range(0, h, 4):
        for bx in range(0, w, 4):
            blk = f[by:by + 4, bx:bx + 4].reshape(16, 4)
            out += _alpha_block(blk[:, 3])
            out += _color_block(blk[:, :3], blk[:, 3])
    return bytes(out)
