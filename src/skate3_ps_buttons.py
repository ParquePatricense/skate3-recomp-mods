"""Skate 3 Recomp - PlayStation button icons.

Swaps the Xbox 360 button prompts in Skate 3 (skate3recomp) for PlayStation
style ones (PS3, PS4 or PS5) and puts the originals back on request.

    skate3_ps_buttons.py                      interactive menu
    skate3_ps_buttons.py install ps5 [--game DIR]
    skate3_ps_buttons.py uninstall   [--game DIR]
    skate3_ps_buttons.py status      [--game DIR]

DIR is the skate3recomp folder (the one with skate3.exe) or its game folder.
"""
import argparse, base64, os, shutil, struct, sys, zlib

from s3big import BigArchive
import payload

VERSION = '1.0.0'
ARCHIVE = 'fedynamic.big'
BACKUP_SUFFIX = '.xbox-original'
STATE_SUFFIX = '.ps-buttons'
ENTRY = 'buttons.rx2'
RX2_SIZE = 0x10A8B8
TEX_BYTES = 0x4000
STYLE_NAMES = {'ps3': 'PS3', 'ps4': 'PS4', 'ps5': 'PS5'}


class ModError(Exception):
    pass


def app_dir():
    if getattr(sys, 'frozen', False):
        return os.path.dirname(sys.executable)
    return os.path.dirname(os.path.abspath(__file__))


def find_archive(game=None):
    roots = [game] if game else [app_dir(), os.getcwd()]
    rels = [('game', 'data', 'big'), ('data', 'big'), ('big',), ()]
    for root in roots:
        if not root:
            continue
        for rel in rels:
            p = os.path.join(root, *rel, ARCHIVE)
            if os.path.isfile(p):
                return os.path.abspath(p)
    where = game or 'this folder'
    raise ModError(f'Could not find game/data/big/{ARCHIVE} in {where}.\n'
                   'Put the mod next to skate3.exe, or pass --game "<skate3recomp folder>".\n'
                   'The game must have been installed (Select ISO) at least once.')


def load_style(style):
    blob = zlib.decompress(base64.b64decode(''.join(payload.STYLES[style])))
    rec = 1 + TEX_BYTES
    return {blob[k]: blob[k + 1:k + rec] for k in range(0, len(blob), rec)}


def patch_rx2(rx2, style):
    if len(rx2) != RX2_SIZE or rx2[:8] != b'\x89RW4xb2\0':
        raise ModError(f'{ENTRY} has an unexpected layout; this game version is not supported.')
    data_base = struct.unpack('>I', rx2[0x44:0x48])[0]
    if data_base + 66 * TEX_BYTES != RX2_SIZE:
        raise ModError(f'{ENTRY} has an unexpected texture layout; this game version is not supported.')
    out = bytearray(rx2)
    for idx, tex in load_style(style).items():
        o = data_base + idx * TEX_BYTES
        out[o:o + TEX_BYTES] = tex
    return bytes(out)


def paths(archive):
    return archive + BACKUP_SUFFIX, archive + STATE_SUFFIX


def current_style(archive):
    backup, state = paths(archive)
    if os.path.isfile(backup) and os.path.isfile(state):
        return open(state, encoding='utf-8').read().strip() or 'unknown'
    return None


def install(style, game=None, log=print):
    if style not in payload.STYLES:
        raise ModError(f'Unknown style "{style}". Use one of: {", ".join(payload.STYLES)}')
    archive = find_archive(game)
    backup, state = paths(archive)
    if not os.path.isfile(backup):
        log(f'Backing up original {ARCHIVE} ...')
        tmp = backup + '.tmp'
        shutil.copyfile(archive, tmp)
        os.replace(tmp, backup)
    # Always patch from the untouched original so switching styles is clean.
    big = BigArchive(open(backup, 'rb').read())
    idx = big.find(ENTRY)
    log(f'Installing {STYLE_NAMES[style]} button icons ...')
    big.replace(idx, patch_rx2(big.read(idx), style))
    tmp = archive + '.tmp'
    with open(tmp, 'wb') as f:
        f.write(big.data)
    os.replace(tmp, archive)
    with open(state, 'w', encoding='utf-8') as f:
        f.write(style)
    log(f'Done: {STYLE_NAMES[style]} icons installed in {archive}')


def uninstall(game=None, log=print):
    archive = find_archive(game)
    backup, state = paths(archive)
    if not os.path.isfile(backup):
        log('Nothing to uninstall: the original Xbox icons are already in place.')
        return
    os.replace(backup, archive)
    if os.path.isfile(state):
        os.remove(state)
    log(f'Done: original Xbox icons restored in {archive}')


def status(game=None, log=print):
    archive = find_archive(game)
    st = current_style(archive)
    log(f'Game archive: {archive}')
    log(f'Installed: {STYLE_NAMES.get(st, st)} icons' if st else 'Installed: original Xbox icons')


def menu():
    print(f'Skate 3 Recomp - PlayStation button icons v{VERSION}\n')
    try:
        status()
    except ModError as e:
        print(e)
        input('\nPress Enter to exit...')
        return 1
    print('\n  1) Install PS3 icons\n  2) Install PS4 icons\n  3) Install PS5 icons\n'
          '  4) Restore original Xbox icons\n  0) Exit\n')
    choice = input('Choose an option: ').strip()
    try:
        if choice in ('1', '2', '3'):
            install(['ps3', 'ps4', 'ps5'][int(choice) - 1])
        elif choice == '4':
            uninstall()
    except ModError as e:
        print(f'Error: {e}')
    input('\nPress Enter to exit...')
    return 0


def main(argv=None):
    argv = sys.argv[1:] if argv is None else argv
    if not argv:
        return menu()
    ap = argparse.ArgumentParser(prog='skate3_ps_buttons', description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    common = argparse.ArgumentParser(add_help=False)
    common.add_argument('--game', help='skate3recomp folder (default: next to this program)')
    sub = ap.add_subparsers(dest='cmd', required=True)
    p = sub.add_parser('install', parents=[common], help='install PlayStation icons')
    p.add_argument('style', choices=sorted(payload.STYLES))
    sub.add_parser('uninstall', parents=[common], help='restore the original Xbox icons')
    sub.add_parser('status', parents=[common], help='show which icons are installed')
    a = ap.parse_args(argv)
    try:
        if a.cmd == 'install':
            install(a.style, a.game)
        elif a.cmd == 'uninstall':
            uninstall(a.game)
        else:
            status(a.game)
    except ModError as e:
        print(f'Error: {e}', file=sys.stderr)
        return 1
    return 0


if __name__ == '__main__':
    sys.exit(main())
