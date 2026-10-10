"""Skate 3 'EB' v3 BIG archives and EA RefPack compression (stdlib only)."""
import struct

# ---------------------------------------------------------------- RefPack
def refpack_decompress(src):
    flags = src[0]
    if src[1] != 0xFB:
        raise ValueError('not a RefPack stream')
    i = 2
    if flags & 0x80:
        size = int.from_bytes(src[i:i + 4], 'big'); i += 4
        if flags & 1: i += 4
    else:
        size = int.from_bytes(src[i:i + 3], 'big'); i += 3
        if flags & 1: i += 3
    out = bytearray()
    while True:
        b0 = src[i]
        if b0 < 0x80:
            b1 = src[i + 1]; i += 2
            lit = b0 & 3; ln = ((b0 >> 2) & 7) + 3; off = ((b0 & 0x60) << 3) + b1 + 1
        elif b0 < 0xC0:
            b1, b2 = src[i + 1], src[i + 2]; i += 3
            lit = b1 >> 6; ln = (b0 & 0x3F) + 4; off = ((b1 & 0x3F) << 8) + b2 + 1
        elif b0 < 0xE0:
            b1, b2, b3 = src[i + 1], src[i + 2], src[i + 3]; i += 4
            lit = b0 & 3; ln = ((b0 & 0x0C) << 6) + b3 + 5; off = ((b0 & 0x10) << 12) + (b1 << 8) + b2 + 1
        elif b0 < 0xFC:
            i += 1; lit = ((b0 & 0x1F) << 2) + 4
            out += src[i:i + lit]; i += lit
            continue
        else:
            i += 1; lit = b0 & 3
            out += src[i:i + lit]
            break
        out += src[i:i + lit]; i += lit
        start = len(out) - off
        if off >= ln:
            out += out[start:start + ln]
        else:
            for k in range(ln):
                out.append(out[start + k])
    if len(out) != size:
        raise ValueError('RefPack size mismatch')
    return bytes(out)


def refpack_compress(data):
    n = len(data)
    out = bytearray()
    if n >= 1 << 24:
        out += bytes([0x90, 0xFB]) + n.to_bytes(4, 'big')
    else:
        out += bytes([0x10, 0xFB]) + n.to_bytes(3, 'big')
    table = {}
    i = 0
    lit_start = 0

    def flush_literals(upto):
        # emit literal blocks (multiples of 4) leaving 0..3 for the command
        nonlocal lit_start
        while upto - lit_start >= 4:
            blk = min(112, (upto - lit_start) & ~3)
            out.append(0xE0 + ((blk - 4) >> 2))
            out.extend(data[lit_start:lit_start + blk])
            lit_start += blk

    while i + 3 <= n:
        key = data[i:i + 3]
        best_len = best_off = 0
        for p in reversed(table.get(key, ())):
            off = i - p
            if off > 131072:
                break
            ln = 3
            lim = min(1028, n - i)
            while ln < lim and data[p + ln] == data[i + ln]:
                ln += 1
            if ln > best_len:
                best_len, best_off = ln, off
                if ln == lim:
                    break
        ok = (best_len >= 3 and best_off <= 1024) or (best_len >= 4 and best_off <= 16384) or best_len >= 5
        if not ok:
            table.setdefault(key, []).append(i)
            if len(table[key]) > 32:
                del table[key][0]
            i += 1
            continue
        flush_literals(i)
        lit = i - lit_start
        ln, off = best_len, best_off - 1
        if ln <= 10 and off < 1024:
            out += bytes([((off >> 3) & 0x60) | ((ln - 3) << 2) | lit, off & 0xFF])
        elif ln <= 67 and off < 16384:
            out += bytes([0x80 | (ln - 4), (lit << 6) | (off >> 8), off & 0xFF])
        else:
            ln = min(ln, 1028)
            out += bytes([0xC0 | ((off >> 12) & 0x10) | (((ln - 5) >> 6) & 0x0C) | lit,
                          (off >> 8) & 0xFF, off & 0xFF, (ln - 5) & 0xFF])
        out.extend(data[lit_start:i])
        end = i + ln
        # index a few positions inside the match (cheap, keeps runs findable)
        step = 1 if ln < 32 else 16
        for q in range(i, min(end, n - 2), step):
            lst = table.setdefault(data[q:q + 3], [])
            lst.append(q)
            if len(lst) > 32:
                del lst[0]
        i = end
        lit_start = i
    i = n
    flush_literals(i)
    out.append(0xFC | (n - lit_start))
    out.extend(data[lit_start:n])
    return bytes(out)

# ---------------------------------------------------------------- BIG (EB v3)
class BigArchive:
    def __init__(self, blob):
        d = bytearray(blob)
        if d[:2] != b'EB' or struct.unpack('>H', d[2:4])[0] != 3:
            raise ValueError('not a Skate 3 EB v3 BIG archive')
        self.data = d
        self.count = struct.unpack('>I', d[4:8])[0]
        self.shift = d[10]
        self.names_off, _, self.name_len, _, _ = struct.unpack('>IIBBH', d[12:24])

    def find(self, name):
        target = name.encode('latin1')
        for i in range(self.count):
            r = self.names_off + i * self.name_len
            if bytes(self.data[r + 2:r + self.name_len]).split(b'\0')[0] == target:
                return i
        raise KeyError(name)

    def entry(self, i):
        o = 0x30 + i * 16
        return struct.unpack('>IIII', self.data[o:o + 16])

    def read(self, i):
        off, csz, size, _ = self.entry(i)
        raw = bytes(self.data[off << self.shift:(off << self.shift) + csz])
        if raw[1:2] == b'\xfb':
            return refpack_decompress(raw)
        return raw

    def replace(self, i, payload):
        """Append a RefPack-compressed copy of payload and point entry i at it."""
        comp = refpack_compress(payload)
        if refpack_decompress(comp) != payload:
            raise RuntimeError('internal compression check failed')
        align = 1 << self.shift
        pos = (len(self.data) + align - 1) & ~(align - 1)
        self.data += b'\0' * (pos - len(self.data)) + comp
        self.data += b'\0' * ((-len(self.data)) % align)
        o = 0x30 + i * 16
        _, _, _, h = self.entry(i)
        self.data[o:o + 16] = struct.pack('>IIII', pos >> self.shift, len(comp), len(payload), h)
        self.data[0x1C:0x20] = struct.pack('>I', len(self.data))
