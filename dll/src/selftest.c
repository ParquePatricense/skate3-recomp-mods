// Test-only export: install a hook into a skate3.exe image mapped by a test
// harness (no game running) and report what happened.
#ifdef S3D_TEST
#include "core.h"

static PPCFunc* o_getinst;
static int g_hits;
static void h_getinst(PPCContext* ctx, u8* base) {
  int quiet = ++g_hits > 2;
  if (!quiet) logf("h_getinst: r3=%08X lr=%08X", ctx->r3.u32, (u32)ctx->lr);
  o_getinst(ctx, base);
  if (!quiet) logf("h_getinst: returned r3=%08X", ctx->r3.u32);
}

static const HookSpec kTest[] = {
  {"getinst", 0x8000E0, 5, {0x41, 0x56, 0x56, 0x57, 0x53}, h_getinst, &o_getinst},
};

__declspec(dllexport) int s3d_selftest(u8* exe, wchar* dir) {
  g_exe = exe;
  usize i = 0;
  for (; dir[i]; i++) g_dir[i] = dir[i];
  g_dir[i] = 0;
  log_open();
  logf("selftest exe=%p", exe);
  if (!hook_entry(&kTest[0])) return -1;
  return 0;
}
__declspec(dllexport) int s3d_hits(void) { return g_hits; }
int load_map(void);
int prof_start(void);
void prof_stop(void);
__declspec(dllexport) int s3d_proftest(int stop) { if (stop) { prof_stop(); return 0; } return prof_start(); }
int game_install(void);
__declspec(dllexport) int s3d_install(void) { return game_install(); }
u32 guest_of(u64 rip);
__declspec(dllexport) int s3d_maptest(void) {
  if (!load_map()) return -1;
  u32 a = guest_of((u64)(g_exe + 0x8000E0 + 10));
  u32 b = guest_of((u64)(g_exe + 0xE41C80));
  u32 c = guest_of((u64)(g_exe + 0x14892A0 + 0x300));
  logf("maptest: %08X %08X %08X", a, b, c);
  return (a == 0x824AD240 && b == 0x825D60C8 && c == 0x826FDD70) ? 1 : 0;
}
#endif
