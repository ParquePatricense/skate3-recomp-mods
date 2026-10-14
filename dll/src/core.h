#pragma once
#include "win.h"

typedef __builtin_va_list va_list;
#define va_start(ap, x) __builtin_va_start(ap, x)
#define va_arg(ap, t) __builtin_va_arg(ap, t)
#define va_end(ap) __builtin_va_end(ap)

void* memset(void* d, int c, usize n);
void* memcpy(void* d, const void* s, usize n);
int memcmp(const void* a, const void* b, usize n);
usize s_len(const char* s);
usize w_len(const wchar* s);
int vfmt(char* buf, usize cap, const char* f, va_list ap);
int fmt(char* buf, usize cap, const char* f, ...);

// ---------------------------------------------------------------- PPC context
// Mirrors rexglue's PPCContext (include/rex/ppc/context.h, no REX_CONFIG_*
// locals), as compiled into the official skate3.exe v2.0.2.
typedef union {
  i32 s32;
  u32 u32;
  i64 s64;
  u64 u64;
  float f32;
  double f64;
} Reg;

typedef struct {
  u8 lt, gt, eq, so;
} CR;

typedef struct __attribute__((aligned(64))) {
  Reg r3, r0, r1, r2, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13;
  Reg r14, r15, r16, r17, r18, r19, r20, r21, r22, r23, r24, r25, r26, r27, r28, r29, r30, r31;
  u64 lr;              // 0x100
  Reg ctr;             // 0x108
  u8 xer_so, xer_ov, xer_ca, xer_pad;  // 0x110
  u8 reserved_[0x140 - 0x114];
  CR cr[8];            // 0x140
} PPCContext;

_Static_assert(__builtin_offsetof(PPCContext, r31) == 0xF8, "r31");
_Static_assert(__builtin_offsetof(PPCContext, lr) == 0x100, "lr");
_Static_assert(__builtin_offsetof(PPCContext, cr) == 0x140, "cr");

typedef void PPCFunc(PPCContext* ctx, u8* base);

// Guest memory (big-endian). Addresses >= 0xE0000000 carry the runtime's
// 4 KB physical-view offset, like the generated code does.
static inline u8* gptr(u8* base, u32 a) { return base + a + (a >= 0xE0000000u ? 0x1000 : 0); }
static inline u32 bswap32(u32 v) { return __builtin_bswap32(v); }
static inline u32 gload32(u8* base, u32 a) { return bswap32(*(u32*)gptr(base, a)); }
static inline u16 gload16(u8* base, u32 a) { return __builtin_bswap16(*(u16*)gptr(base, a)); }
static inline u8 gload8(u8* base, u32 a) { return *gptr(base, a); }
static inline void gstore32(u8* base, u32 a, u32 v) { *(u32*)gptr(base, a) = bswap32(v); }
static inline void gstore8(u8* base, u32 a, u8 v) { *gptr(base, a) = v; }
static inline float gloadf(u8* base, u32 a) {
  u32 v = gload32(base, a);
  return *(float*)&v;
}

// ---------------------------------------------------------------- runtime
extern wchar g_dir[520];
extern u8* g_exe;
void path_join(wchar* out, usize cap, const wchar* dir, const wchar* name);
void log_open(void);
void logf(const char* f, ...);
double now_s(void);

typedef struct {
  const char* name;
  u32 rva;          // host RVA in skate3.exe (entry or call site)
  u8 len;           // bytes stolen (entry) / verified (call site)
  u8 expect[24];
  PPCFunc* fn;      // replacement
  PPCFunc** orig;   // out: original (trampoline or callee)
  u32 guest;        // guest function the host code belongs to (0: use rva as is)
} HookSpec;

int check_bytes(u32 rva, const u8* expect, usize n);
int hook_entry(const HookSpec* h);
int hook_call(const HookSpec* h);
int patch_bytes(u32 rva, const u8* expect, const u8* repl, usize n);

static inline PPCFunc* host_fn(u32 rva) { return (PPCFunc*)(g_exe + rva); }
