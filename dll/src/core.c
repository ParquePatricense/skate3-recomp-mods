// Freestanding runtime pieces: memory helpers, a tiny printf, the log file
// and the x86-64 hook engine used to patch recompiled guest functions inside
// skate3.exe.
#include "core.h"

int _fltused = 0;

void* memset(void* d, int c, usize n) {
  u8* p = (u8*)d;
  while (n--) *p++ = (u8)c;
  return d;
}

void* memcpy(void* d, const void* s, usize n) {
  u8* a = (u8*)d;
  const u8* b = (const u8*)s;
  while (n--) *a++ = *b++;
  return d;
}

int memcmp(const void* x, const void* y, usize n) {
  const u8* a = (const u8*)x;
  const u8* b = (const u8*)y;
  for (; n; n--, a++, b++)
    if (*a != *b) return *a < *b ? -1 : 1;
  return 0;
}

usize s_len(const char* s) {
  usize n = 0;
  while (s[n]) n++;
  return n;
}

usize w_len(const wchar* s) {
  usize n = 0;
  while (s[n]) n++;
  return n;
}

// ---------------------------------------------------------------- printf

static char* put_u64(char* o, char* end, u64 v, int base, int width, char pad, int upper) {
  char tmp[24];
  int n = 0;
  const char* dig = upper ? "0123456789ABCDEF" : "0123456789abcdef";
  do {
    tmp[n++] = dig[v % (u64)base];
    v /= (u64)base;
  } while (v && n < 24);
  while (n < width && n < 24) tmp[n++] = pad;
  while (n && o < end) *o++ = tmp[--n];
  return o;
}

int vfmt(char* buf, usize cap, const char* f, va_list ap) {
  char* o = buf;
  char* end = buf + cap - 1;
  while (*f && o < end) {
    if (*f != '%') {
      *o++ = *f++;
      continue;
    }
    f++;
    char pad = ' ';
    int width = 0, ll = 0, prec = -1;
    if (*f == '0') {
      pad = '0';
      f++;
    }
    while (*f >= '0' && *f <= '9') width = width * 10 + (*f++ - '0');
    if (*f == '.') {
      f++;
      prec = 0;
      while (*f >= '0' && *f <= '9') prec = prec * 10 + (*f++ - '0');
    }
    while (*f == 'l') {
      ll++;
      f++;
    }
    char c = *f++;
    switch (c) {
      case 'd': {
        i64 v = ll ? va_arg(ap, i64) : (i64)va_arg(ap, int);
        if (v < 0 && o < end) {
          *o++ = '-';
          v = -v;
        }
        o = put_u64(o, end, (u64)v, 10, width, pad, 0);
      } break;
      case 'u':
        o = put_u64(o, end, ll ? va_arg(ap, u64) : (u64)va_arg(ap, unsigned), 10, width, pad, 0);
        break;
      case 'x':
      case 'X':
        o = put_u64(o, end, ll ? va_arg(ap, u64) : (u64)va_arg(ap, unsigned), 16, width, pad,
                    c == 'X');
        break;
      case 'p':
        o = put_u64(o, end, (u64)va_arg(ap, void*), 16, 16, '0', 1);
        break;
      case 'f': {  // fixed point, default 3 decimals
        double v = va_arg(ap, double);
        if (prec < 0) prec = 3;
        if (v < 0 && o < end) {
          *o++ = '-';
          v = -v;
        }
        u64 ip = (u64)v;
        o = put_u64(o, end, ip, 10, width, pad, 0);
        if (prec > 0 && o < end) {
          *o++ = '.';
          double fr = v - (double)ip;
          for (int i = 0; i < prec && o < end; i++) {
            fr *= 10.0;
            int d = (int)fr;
            *o++ = (char)('0' + d);
            fr -= d;
          }
        }
      } break;
      case 's': {
        const char* s = va_arg(ap, const char*);
        if (!s) s = "(null)";
        while (*s && o < end) *o++ = *s++;
      } break;
      case 'S': {  // wide string, ASCII only
        const wchar* s = va_arg(ap, const wchar*);
        if (!s) s = L"(null)";
        while (*s && o < end) *o++ = (*s < 128) ? (char)*s : '?', s++;
      } break;
      case 'c':
        *o++ = (char)va_arg(ap, int);
        break;
      case '%':
        *o++ = '%';
        break;
      default:
        break;
    }
  }
  *o = 0;
  return (int)(o - buf);
}

int fmt(char* buf, usize cap, const char* f, ...) {
  va_list ap;
  va_start(ap, f);
  int n = vfmt(buf, cap, f, ap);
  va_end(ap);
  return n;
}

// ---------------------------------------------------------------- paths/log

wchar g_dir[520];  // directory of skate3.exe, with trailing backslash
static HANDLE g_log = INVALID_HANDLE_VALUE;
static CRITICAL_SECTION g_log_cs;
static i64 g_qpc0, g_qpf;

void path_join(wchar* out, usize cap, const wchar* dir, const wchar* name) {
  usize n = 0;
  for (; dir[n] && n + 1 < cap; n++) out[n] = dir[n];
  for (usize i = 0; name[i] && n + 1 < cap; i++) out[n++] = name[i];
  out[n] = 0;
}

double now_s(void) {
  LARGE_INTEGER c;
  QueryPerformanceCounter(&c);
  return (double)(c.QuadPart - g_qpc0) / (double)g_qpf;
}

void log_open(void) {
  LARGE_INTEGER f, c;
  QueryPerformanceFrequency(&f);
  QueryPerformanceCounter(&c);
  g_qpf = f.QuadPart;
  g_qpc0 = c.QuadPart;
  InitializeCriticalSection(&g_log_cs);
  wchar p[600];
  path_join(p, 600, g_dir, L"skate3-direct.log");
  g_log = CreateFileW(p, GENERIC_WRITE, FILE_SHARE_READ | FILE_SHARE_WRITE, NULL, CREATE_ALWAYS,
                      FILE_ATTRIBUTE_NORMAL, NULL);
}

void logf(const char* f, ...) {
  if (g_log == INVALID_HANDLE_VALUE) return;
  char buf[1024];
  SYSTEMTIME t;
  GetLocalTime(&t);
  int n = fmt(buf, sizeof buf, "[%02u:%02u:%02u.%03u +%8.3f] ", t.wHour, t.wMinute, t.wSecond,
              t.wMilliseconds, now_s());
  va_list ap;
  va_start(ap, f);
  n += vfmt(buf + n, sizeof buf - n - 2, f, ap);
  va_end(ap);
  buf[n++] = '\r';
  buf[n++] = '\n';
  DWORD w;
  EnterCriticalSection(&g_log_cs);
  WriteFile(g_log, buf, (DWORD)n, &w, NULL);
  LeaveCriticalSection(&g_log_cs);
}

// ---------------------------------------------------------------- hooks

u8* g_exe;  // skate3.exe image base in memory

static u8* g_pool;
static usize g_pool_used, g_pool_cap;

// Executable memory within +-2 GB of skate3.exe so rel32 branches reach it.
static u8* alloc_near(u8* target, usize size) {
  const usize gran = 0x10000;
  for (i64 delta = 0x1000000; delta < 0x70000000; delta += (i64)gran * 16) {
    for (int dir = -1; dir <= 1; dir += 2) {
      u8* want = (u8*)(((u64)(target + dir * delta)) & ~(u64)(gran - 1));
      MEMORY_BASIC_INFORMATION mbi;
      if (!VirtualQuery(want, &mbi, sizeof mbi)) continue;
      if (mbi.State != 0x10000 /*MEM_FREE*/ || mbi.RegionSize < size) continue;
      u8* p = (u8*)VirtualAlloc(want, size, MEM_COMMIT | MEM_RESERVE, PAGE_EXECUTE_READWRITE);
      if (p) return p;
    }
  }
  return NULL;
}

static u8* pool_get(usize n) {
  if (!g_pool) {
    g_pool_cap = 0x10000;
    g_pool = alloc_near(g_exe, g_pool_cap);
    if (!g_pool) return NULL;
  }
  n = (n + 15) & ~(usize)15;
  if (g_pool_used + n > g_pool_cap) return NULL;
  u8* p = g_pool + g_pool_used;
  g_pool_used += n;
  return p;
}

static void put_absjmp(u8* p, const void* to) {
  p[0] = 0xFF;
  p[1] = 0x25;
  p[2] = p[3] = p[4] = p[5] = 0;
  u64 a = (u64)to;
  memcpy(p + 6, &a, 8);
}

static int write_code(u8* p, const u8* bytes, usize n) {
  DWORD old;
  if (!VirtualProtect(p, n, PAGE_EXECUTE_READWRITE, &old)) return 0;
  memcpy(p, bytes, n);
  VirtualProtect(p, n, old, &old);
  FlushInstructionCache(GetCurrentProcess(), p, n);
  return 1;
}

int check_bytes(u32 rva, const u8* expect, usize n) {
  return memcmp(g_exe + rva, expect, n) == 0;
}

// Entry hook: the first `len` bytes of the host function (whole
// instructions, no RIP-relative operands) are replaced by a jump to `fn`;
// *orig receives a trampoline that runs the original function.
int hook_entry(const HookSpec* h) {
  u8* p = g_exe + h->rva;
  if (memcmp(p, h->expect, h->len) != 0) {
    logf("hook %s: bytes differ at rva %X", h->name, h->rva);
    return 0;
  }
  if (h->orig) {
    u8* t = pool_get(h->len + 14);
    if (!t) return 0;
    memcpy(t, p, h->len);
    put_absjmp(t + h->len, p + h->len);
    *h->orig = (PPCFunc*)t;
  }
  u8* stub = pool_get(14);
  if (!stub) return 0;
  put_absjmp(stub, (const void*)h->fn);
  u8 patch[32];
  i32 rel = (i32)(stub - (p + 5));
  patch[0] = 0xE9;
  memcpy(patch + 1, &rel, 4);
  for (int i = 5; i < h->len; i++) patch[i] = 0xCC;
  return write_code(p, patch, h->len);
}

// Call-site hook: a `call rel32` at rva is redirected to fn; *orig gets the
// original callee.
int hook_call(const HookSpec* h) {
  u8* p = g_exe + h->rva;
  if (h->len && memcmp(p, h->expect, h->len) != 0) {
    logf("call hook %s: bytes differ at rva %X", h->name, h->rva);
    return 0;
  }
  if (p[0] != 0xE8) return 0;
  i32 rel;
  memcpy(&rel, p + 1, 4);
  if (h->orig) *h->orig = (PPCFunc*)(p + 5 + rel);
  u8* stub = pool_get(14);
  if (!stub) return 0;
  put_absjmp(stub, (const void*)h->fn);
  i32 nrel = (i32)(stub - (p + 5));
  u8 patch[5];
  patch[0] = 0xE8;
  memcpy(patch + 1, &nrel, 4);
  return write_code(p, patch, 5);
}

int patch_bytes(u32 rva, const u8* expect, const u8* repl, usize n) {
  if (memcmp(g_exe + rva, expect, n) != 0) return 0;
  return write_code(g_exe + rva, repl, n);
}
