// Sampling profiler for the boot: every few ms each thread of the process is
// suspended, its RIP and a scan of its stack for return addresses into the
// recompiled code are mapped back to guest functions, and counts are kept per
// boot phase. At the end the hottest guest functions (self and inclusive) per
// thread go to the log. No locks are taken while a thread is suspended.
#include "core.h"

#define TH32CS_SNAPTHREAD 0x4
#define THREAD_ALL_NEEDED (0x0002 | 0x0008 | 0x0040)  // suspend/resume, get context, query
#define CONTEXT_CONTROL_INT 0x100003

typedef struct {
  DWORD dwSize, cntUsage, th32ThreadID, th32OwnerProcessID;
  long tpBasePri, tpDeltaPri;
  DWORD dwFlags;
} THREADENTRY32;

DLLIMPORT HANDLE WINAPI CreateToolhelp32Snapshot(DWORD flags, DWORD pid);
DLLIMPORT BOOL WINAPI Thread32First(HANDLE h, THREADENTRY32* te);
DLLIMPORT BOOL WINAPI Thread32Next(HANDLE h, THREADENTRY32* te);
DLLIMPORT HANDLE WINAPI OpenThread(DWORD access, BOOL inherit, DWORD tid);
DLLIMPORT DWORD WINAPI SuspendThread(HANDLE h);
DLLIMPORT DWORD WINAPI ResumeThread(HANDLE h);
DLLIMPORT BOOL WINAPI GetThreadContext(HANDLE h, void* ctx);
DLLIMPORT DWORD WINAPI GetCurrentProcessId(void);

// ---------------------------------------------------------------- func map
// PPCFuncMappings in skate3.exe .data: {u64 guest, u64 host} pairs, sorted by
// guest. It is found by content (first pair is the start of the guest .text)
// so a copy of the build with another layout still maps.
#define TEXT_RVA 0x1000u
#define MAP_FIRST_GUEST 0x82380000ull

static u32 g_text_end;
static u64* g_map;     // raw table
static u32 g_mapn;
static u32* g_hrva;   // sorted host RVAs of guest functions
static u32* g_hguest; // guest address for each
static u32 g_nfn;

static void sort_map(void) {
  // shell sort by host rva (48k entries, once)
  u32 n = g_nfn;
  for (u32 gap = n / 2; gap; gap /= 2)
    for (u32 i = gap; i < n; i++) {
      u32 a = g_hrva[i], b = g_hguest[i];
      u32 j = i;
      for (; j >= gap && g_hrva[j - gap] > a; j -= gap) {
        g_hrva[j] = g_hrva[j - gap];
        g_hguest[j] = g_hguest[j - gap];
      }
      g_hrva[j] = a;
      g_hguest[j] = b;
    }
}

static int host_rva(u64 h, u32* out) {
  u64 exe = (u64)g_exe;
  if (h >= exe + TEXT_RVA && h - exe < g_text_end) { *out = (u32)(h - exe); return 1; }
  if (h >= 0x140000000ull + TEXT_RVA && h - 0x140000000ull < g_text_end) {  // image not relocated
    *out = (u32)(h - 0x140000000ull);
    return 1;
  }
  return 0;
}

static int find_map(void) {
  u8* nt = g_exe + *(u32*)(g_exe + 0x3C);
  u16 nsec = *(u16*)(nt + 6);
  u8* sec = nt + 24 + *(u16*)(nt + 20);
  g_text_end = 0;
  for (u16 i = 0; i < nsec; i++) {
    u8* s = sec + 40 * i;
    if (memcmp(s, ".text", 6) == 0) g_text_end = *(u32*)(s + 12) + *(u32*)(s + 8);
  }
  if (!g_text_end) return 0;
  for (u16 i = 0; i < nsec; i++) {
    u8* s = sec + 40 * i;
    u32 va = *(u32*)(s + 12), vs = *(u32*)(s + 8);
    if (memcmp(s, ".data", 6) != 0 && memcmp(s, ".rdata", 7) != 0) continue;
    u64* p = (u64*)(g_exe + ((va + 7) & ~7u));
    u64* e = (u64*)(g_exe + va + vs) - 4;
    for (; p < e; p++) {
      u32 r;
      if (p[0] != MAP_FIRST_GUEST || p[2] <= p[0] || p[2] >= 0x84000000ull || !host_rva(p[1], &r) ||
          !host_rva(p[3], &r))
        continue;
      u32 n = 0;
      u64* q = p;
      while (q + 1 < (u64*)(g_exe + va + vs) && q[0] >= 0x82000000ull && q[0] < 0x84000000ull &&
             host_rva(q[1], &r)) {
        n++;
        q += 2;
      }
      if (n < 1000) continue;
      g_map = p;
      g_mapn = n;
      return 1;
    }
  }
  return 0;
}

int load_map(void) {
  if (g_nfn) return 1;
  if (!g_map && !find_map()) return 0;
  g_hrva = (u32*)VirtualAlloc(NULL, (usize)g_mapn * 8, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
  if (!g_hrva) return 0;
  g_hguest = g_hrva + g_mapn;
  for (u32 i = 0; i < g_mapn; i++) {
    host_rva(g_map[2 * i + 1], &g_hrva[i]);
    g_hguest[i] = (u32)g_map[2 * i];
  }
  g_nfn = g_mapn;
  sort_map();
  return 1;
}

// Host RVA of a guest function, 0 if the table doesn't list it.
u32 host_of(u32 guest) {
  for (u32 i = 0; i < g_nfn; i++)
    if (g_hguest[i] == guest) return g_hrva[i];
  return 0;
}

u32 guest_of(u64 rip) {
  u64 exe = (u64)g_exe;
  if (rip < exe + TEXT_RVA || rip >= exe + g_text_end || !g_nfn) return 0;
  u32 r = (u32)(rip - exe);
  u32 lo = 0, hi = g_nfn;
  while (hi - lo > 1) {
    u32 mid = (lo + hi) / 2;
    if (g_hrva[mid] <= r) lo = mid; else hi = mid;
  }
  return g_hrva[lo] <= r ? g_hguest[lo] : 0;
}

// ---------------------------------------------------------------- counts
// key: phase(<< 12) | thread slot(<< 2, 7 bits) | kind(2 bits: 0 self, 1 incl, 2 wait-in) + guest addr
#define NSLOT 128
#define HT_SIZE 131072
typedef struct {
  u32 guest;
  u16 tag;  // phase<<12 | slot<<2 | kind
  u16 pad;
  u32 count;
} Entry;
static Entry* g_ht;
static u32 g_ht_used;

static void bump(u32 guest, u16 tag) {
  if (!guest) return;
  u32 h = (guest * 2654435761u) ^ (tag * 40503u);
  for (u32 i = 0; i < HT_SIZE; i++) {
    Entry* e = &g_ht[(h + i) & (HT_SIZE - 1)];
    if (e->count == 0) {
      if (g_ht_used > HT_SIZE * 3 / 4) return;
      e->guest = guest;
      e->tag = tag;
      e->count = 1;
      g_ht_used++;
      return;
    }
    if (e->guest == guest && e->tag == tag) {
      e->count++;
      return;
    }
  }
}

typedef struct {
  DWORD tid;
  HANDLE h;
  u32 samples[8];   // per phase
  u32 in_guest[8];  // RIP inside recompiled code
  u32 total, with_guest;  // all samples / samples with any guest frame
  u8 dropped;       // host-only thread (driver, pool): no longer sampled
} Slot;
static Slot g_slots[NSLOT];
static int g_nslots;

extern volatile int g_phase;  // game.c
extern const char* kPhaseName(int p);

static void refresh_threads(DWORD self) {
  HANDLE snap = CreateToolhelp32Snapshot(TH32CS_SNAPTHREAD, 0);
  if (snap == INVALID_HANDLE_VALUE) return;
  THREADENTRY32 te;
  te.dwSize = sizeof te;
  DWORD pid = GetCurrentProcessId();
  if (Thread32First(snap, &te)) do {
      if (te.th32OwnerProcessID != pid || te.th32ThreadID == self) continue;
      int known = 0;
      for (int i = 0; i < g_nslots; i++)
        if (g_slots[i].tid == te.th32ThreadID) known = 1;
      if (known || g_nslots >= NSLOT) continue;
      HANDLE h = OpenThread(THREAD_ALL_NEEDED, FALSE, te.th32ThreadID);
      if (!h) continue;
      g_slots[g_nslots].tid = te.th32ThreadID;
      g_slots[g_nslots].h = h;
      g_nslots++;
    } while (Thread32Next(snap, &te));
  CloseHandle(snap);
}

static __attribute__((aligned(16))) u8 g_ctx[0x4D0];
#define STACK_SCAN 0x6000

static void sample_one(int si, int phase) {
  Slot* s = &g_slots[si];
  if (SuspendThread(s->h) == (DWORD)-1) return;
  *(DWORD*)(g_ctx + 0x30) = CONTEXT_CONTROL_INT;
  u32 chain[6];
  int n = 0;
  u64 rip = 0;
  if (GetThreadContext(s->h, g_ctx)) {
    rip = *(u64*)(g_ctx + 0xF8);
    u64 rsp = *(u64*)(g_ctx + 0x98);
    MEMORY_BASIC_INFORMATION mbi;
    if (VirtualQuery((void*)rsp, &mbi, sizeof mbi) && mbi.State == MEM_COMMIT) {
      u64 end = (u64)mbi.BaseAddress + mbi.RegionSize;
      if (end > rsp + STACK_SCAN) end = rsp + STACK_SCAN;
      u64 exe = (u64)g_exe;
      for (u64 p = rsp & ~7ull; p + 8 <= end && n < 6; p += 8) {
        u64 v = *(u64*)p;
        if (v < exe + TEXT_RVA + 8 || v >= exe + g_text_end) continue;
        u8* ra = (u8*)v;
        // must follow a call: E8 rel32, or FF /2 forms
        if (ra[-5] != 0xE8 && !(ra[-2] == 0xFF && (ra[-1] & 0x38) == 0x10) &&
            !(ra[-3] == 0xFF && (ra[-2] & 0x38) == 0x10) && !(ra[-6] == 0xFF && ra[-5] == 0x15))
          continue;
        u32 g = guest_of(v);
        if (g && (n == 0 || chain[n - 1] != g)) chain[n++] = g;
      }
    }
  }
  ResumeThread(s->h);
  s->samples[phase]++;
  s->total++;
  if (n || guest_of(rip)) s->with_guest++;
  else if (s->total >= 500 && !s->with_guest) s->dropped = 1;  // 2 s without game code
  u32 leaf = guest_of(rip);
  u16 base = (u16)((phase << 12) | (si << 2));
  if (leaf) {
    s->in_guest[phase]++;
    bump(leaf, base | 0);
    bump(leaf, base | 1);
    for (int i = 0; i < n; i++)
      if (chain[i] != leaf) bump(chain[i], base | 1);
  } else if (n) {
    bump(chain[0], base | 2);  // in runtime / OS, called from this guest function
    for (int i = 0; i < n; i++) bump(chain[i], base | 1);
  }
}

static volatile int g_prof_stop;

// Thread names (SetThreadDescription; Windows 10 1607+), resolved at run time.
typedef long(WINAPI* GetThreadDescription_t)(HANDLE h, wchar** name);
typedef void*(WINAPI* LocalFree_t)(void* p);

static void thread_name(HANDLE h, char* out, int cap) {
  out[0] = 0;
  HMODULE k = GetModuleHandleW(L"kernel32.dll");
  GetThreadDescription_t gtd = k ? (GetThreadDescription_t)GetProcAddress(k, "GetThreadDescription") : NULL;
  LocalFree_t lf = k ? (LocalFree_t)GetProcAddress(k, "LocalFree") : NULL;
  wchar* w = NULL;
  if (!gtd || gtd(h, &w) < 0 || !w) return;
  int i = 0;
  for (; w[i] && i < cap - 1; i++) out[i] = w[i] < 128 ? (char)w[i] : '?';
  out[i] = 0;
  if (lf) lf(w);
}

void prof_stop(void) { g_prof_stop = 1; }

static void report(void) {
  logf("---- perfil (muestras cada 4 ms) ----");
  for (int ph = 0; ph < 5; ph++) {
    for (int si = 0; si < g_nslots; si++) {
      Slot* s = &g_slots[si];
      if (s->samples[ph] < 25 || !s->with_guest) continue;
      char name[64];
      thread_name(s->h, name, sizeof name);
      logf("[%s] hilo %u \"%s\": %u muestras, %u en código del juego", kPhaseName(ph), s->tid, name,
           s->samples[ph], s->in_guest[ph]);
      for (int kind = 0; kind < 3; kind++) {
        // top 12 by count, selection without sorting the table
        static u32 shown[12];
        int nshown = 0;
        for (int k = 0; k < 12; k++) {
          Entry* best = NULL;
          for (u32 i = 0; i < HT_SIZE; i++) {
            Entry* e = &g_ht[i];
            if (!e->count || e->tag != (u16)((ph << 12) | (si << 2) | kind)) continue;
            int dup = 0;
            for (int q = 0; q < nshown; q++)
              if (shown[q] == e->guest) dup = 1;
            if (dup) continue;
            if (!best || e->count > best->count) best = e;
          }
          if (!best || best->count * 50 < s->samples[ph]) break;  // < 2 %
          shown[nshown++] = best->guest;
          logf("   %s %08X  %3u%%  (%u)", kind == 0 ? "propio " : kind == 1 ? "incluye" : "espera ",
               best->guest, best->count * 100 / s->samples[ph], best->count);
        }
      }
    }
  }
  logf("---- fin del perfil ----");
}

static DWORD WINAPI prof_thread(void* arg) {
  (void)arg;
  DWORD self = GetCurrentThreadId();
  int tick = 0;
  double t_end = 0;
  while (1) {
    if (g_prof_stop && t_end == 0) t_end = now_s() + 1.0;
    if (t_end != 0 && now_s() > t_end) break;
    if (now_s() > 90) break;
    if ((tick++ % 50) == 0) refresh_threads(self);
    int ph = g_phase;
    for (int i = 0; i < g_nslots; i++)
      if (!g_slots[i].dropped) sample_one(i, ph);
    Sleep(4);
  }
  int nd = 0;
  for (int i = 0; i < g_nslots; i++) nd += g_slots[i].dropped;
  logf("perfil: %d hilos vistos, %d sin código del juego", g_nslots, nd);
  report();
  for (int i = 0; i < g_nslots; i++) CloseHandle(g_slots[i].h);
  return 0;
}

int prof_start(void) {
  if (!load_map()) {
    logf("perfil: no pude leer la tabla de funciones");
    return 0;
  }
  g_ht = (Entry*)VirtualAlloc(NULL, sizeof(Entry) * HT_SIZE, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
  if (!g_ht) return 0;
  CreateThread(NULL, 0, prof_thread, NULL, 0, NULL);
  logf("perfil: %u funciones del juego mapeadas", g_nfn);
  return 1;
}
