#include "../src/core.h"
#define DONT_RESOLVE_DLL_REFERENCES 1
DLLIMPORT HMODULE WINAPI LoadLibraryExW(const wchar* name, HANDLE f, DWORD flags);
DLLIMPORT void WINAPI ExitProcess(u32 code);
DLLIMPORT HANDLE WINAPI GetStdHandle(DWORD n);
typedef int (*selftest_t)(u8*, wchar*);
typedef int (*hits_t)(void);

static void out(const char* s) { DWORD w; WriteFile(GetStdHandle((DWORD)-11), s, (DWORD)s_len(s), &w, NULL); }

static u8* g_base; static PPCFunc* g_f; static volatile int g_quit;
static DWORD WINAPI busy(void* a) {
  (void)a;
  static PPCContext c;
  while (!g_quit) { c.r1.u64 = 0x7000E000; c.lr = 0x82001234; g_f(&c, g_base); }
  return 0;
}
void mainCRTStartup(void) {
  char buf[256];
  u8* exe = (u8*)LoadLibraryExW(L"skate3.exe", NULL, DONT_RESOLVE_DLL_REFERENCES);
  fmt(buf, sizeof buf, "exe mapped at %p\n", exe); out(buf);
  if (!exe) ExitProcess(1);
  HMODULE d = LoadLibraryW(L"xinput9_1_0.dll");
  selftest_t st = (selftest_t)GetProcAddress(d, "s3d_selftest");
  hits_t hits = (hits_t)GetProcAddress(d, "s3d_hits");
  if (!st) { out("no selftest export\n"); ExitProcess(2); }
  int r = st(exe, L".\\");
  fmt(buf, sizeof buf, "install -> %d\n", r); out(buf);
  // guest memory: reserve 4 GB + 4 KB, commit the global page and a stack
  u8* base = (u8*)VirtualAlloc(NULL, 0x100001000ull, MEM_RESERVE, PAGE_READWRITE);
  fmt(buf, sizeof buf, "guest base %p\n", base); out(buf);
  VirtualAlloc(base + 0x830C0000, 0x20000, MEM_COMMIT, PAGE_READWRITE);
  VirtualAlloc(base + 0x70000000, 0x10000, MEM_COMMIT, PAGE_READWRITE);
  gstore32(base, 0x830CFE1C, 0x40001234);
  static PPCContext ctx;
  ctx.r1.u64 = 0x7000F000;
  ctx.lr = 0x82001234;
  PPCFunc* f = (PPCFunc*)(exe + 0x8000E0);
  f(&ctx, base);
  fmt(buf, sizeof buf, "after call: r3=%08X hits=%d lr=%llX r1=%llX\n", ctx.r3.u32, hits(), ctx.lr, ctx.r1.u64); out(buf);
  typedef int (*mt_t)(void);
  mt_t mt = (mt_t)GetProcAddress(d, "s3d_maptest");
  int m = mt ? mt() : -2;
  fmt(buf, sizeof buf, "maptest -> %d\n", m); out(buf);
  // profiler: a busy thread runs guest code while it samples
  typedef int (*pt_t)(int);
  pt_t pt = (pt_t)GetProcAddress(d, "s3d_proftest");
  g_base = base; g_f = f;
  CreateThread(NULL, 0, busy, NULL, 0, NULL);
  fmt(buf, sizeof buf, "prof start -> %d\n", pt(0)); out(buf);
  Sleep(1500);
  pt(1);
  Sleep(4000);
  g_quit = 1;
  typedef int (*in_t)(void);
  in_t in = (in_t)GetProcAddress(d, "s3d_install");
  int gi = in ? in() : -2;
  fmt(buf, sizeof buf, "game_install -> %d\n", gi); out(buf);
  Sleep(2000);
  if (gi != 1) ExitProcess(4);
  in_t nt = (in_t)GetProcAddress(d, "s3d_naptest");
  int nr = nt ? nt() : -2;
  fmt(buf, sizeof buf, "naptest -> %d\n", nr); out(buf);
  if (nr != 1) ExitProcess(5);
  ExitProcess(ctx.r3.u32 == 0x40001234 && m == 1 ? 0 : 3);
}
