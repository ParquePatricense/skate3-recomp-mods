// skate3-direct: loaded by skate3.exe as xinput9_1_0.dll (it sits next to
// the exe). Forwards XInputGetState to the system DLL and, when skate3.exe
// is the exact v2.0.2 build, patches the boot flow for a direct load.
#include "core.h"

#define EXE_TIMESTAMP 0x6A63D45Du
#define EXE_IMAGESIZE 0x060F2000u

int game_install(void);  // game.c

static HMODULE g_self;

// ---------------------------------------------------------------- XInput
typedef DWORD(WINAPI* XInputGetState_t)(DWORD, void*);
static XInputGetState_t g_real_xigs;
static volatile long g_xi_loaded;

__declspec(dllexport) DWORD WINAPI XInputGetState(DWORD idx, void* state) {
  if (!g_xi_loaded) {
    wchar p[600];
    u32 n = GetSystemDirectoryW(p, 560);
    if (n && n < 560) {
      path_join(p + n, 600 - n, L"\\", L"xinput9_1_0.dll");
      HMODULE m = LoadLibraryW(p);
      if (m) g_real_xigs = (XInputGetState_t)GetProcAddress(m, "XInputGetState");
    }
    g_xi_loaded = 1;
  }
  if (!g_real_xigs) return 1167;  // ERROR_DEVICE_NOT_CONNECTED
  return g_real_xigs(idx, state);
}

// The modder finds its own dll by this export (and the version in it).
__declspec(dllexport) u32 skate3_direct_version(void) { return 2; }

// ---------------------------------------------------------------- startup
// The v2.0.2 build is recognised by its timestamp; copies of it can differ in
// size (a section added after linking), so the size is only logged. What
// decides is game_install: every hook checks its exact bytes first and
// nothing is touched if a single one differs.
static int exe_matches(void) {
  IMAGE_DOS_HEADER* d = (IMAGE_DOS_HEADER*)g_exe;
  if (d->e_magic != 0x5A4D) return 0;
  u8* nt = g_exe + d->e_lfanew;
  u32 ts = *(u32*)(nt + 8);
  u32 size = *(u32*)(nt + 24 + 56);
  if (ts != EXE_TIMESTAMP) {
    logf("skate3.exe no es la v2.0.2 conocida (stamp %08X, size %08X): no toco nada", ts, size);
    return 0;
  }
  if (size != EXE_IMAGESIZE) logf("skate3.exe v2.0.2 con otro tamaño (%08X): reviso byte a byte", size);
  return 1;
}

static int disabled_by_env(void) {
  wchar v[8];
  DWORD n = GetEnvironmentVariableW(L"SKATE3_DIRECT", v, 8);
  return n == 1 && v[0] == '0';
}

BOOL WINAPI DllMain(HMODULE self, DWORD reason, void* reserved) {
  (void)reserved;
  if (reason != 1) return TRUE;
  g_self = self;
  DisableThreadLibraryCalls(self);
  g_exe = (u8*)GetModuleHandleW(NULL);
  DWORD n = GetModuleFileNameW(NULL, g_dir, 512);
  while (n && g_dir[n - 1] != '\\' && g_dir[n - 1] != '/') n--;
  g_dir[n] = 0;
  log_open();
  logf("skate3-direct cargado");
  if (disabled_by_env()) {
    logf("desactivado (SKATE3_DIRECT=0)");
    return TRUE;
  }
  if (!exe_matches()) return TRUE;
  int ok = game_install();
  logf("parches %s", ok ? "instalados" : "NO instalados (se juega normal)");
  return TRUE;
}
