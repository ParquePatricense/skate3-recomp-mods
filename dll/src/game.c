// Skate 3 boot flow patches (skate3.exe v2.0.2).
//
// BootFlow (guest object, one handler per state, called with r4 = event:
// 0 enter, 1 update, 3/4 exit; ChangeState = sub_826DFB30(this, next)):
//   826FDD70 language select -> 826FDE48 EA logo (FE 5) -> 826FDEF0 title
//   (FE 24) -> 826FE058 -> 826FE138 (FE 22) -> 826FE1D8 Press Start
//   -> 826DF8B0 fade -> 826FE4A8 -> 826FE4F8 -> ... -> gameplay
// FE flow manager: GetInst() = [0x830CFE1C]; its +8 is the array of FE
// screens by id. The EA logo state waits for screen5+0x40, Press Start for
// screen22+0x3C and then for the START flag at 0x83083973.
#include "core.h"

// ---------------------------------------------------------------- runtime API
typedef u32 (*GameplayContextValue_t)(void);
static GameplayContextValue_t p_gameplay_ctx;

// The recomp's synthetic pad input (xam SetSyntheticAutoTap: START at Press
// Start, A in the boot dialogs; 5 polls pressed, 15 released), read from
// rexruntime.dll's .data only on the build these offsets come from.
#define RT_TIMESTAMP 0x6A63D454u
#define RT_IMAGESIZE 0x01328000u
static u8* g_rt_syn;  // rexruntime base when it matches, else NULL

static void resolve_runtime(void) {
  HMODULE rt = GetModuleHandleW(L"rexruntime.dll");
  if (!rt) return;
  p_gameplay_ctx = (GameplayContextValue_t)GetProcAddress(
      rt, "?GameplayContextValue@guest_presence@kernel@rex@@YAIXZ");
  u8* b = (u8*)rt;
  u8* nt = b + ((IMAGE_DOS_HEADER*)b)->e_lfanew;
  if (*(u32*)(nt + 8) == RT_TIMESTAMP && *(u32*)(nt + 24 + 56) == RT_IMAGESIZE) g_rt_syn = b;
}
static u32 syn_polls(void) { return g_rt_syn ? *(volatile u32*)(g_rt_syn + 0x11A4C8C) : 0; }

// ---------------------------------------------------------------- options
static int g_skip_logo = 1;

static int env_off(const wchar* name) {
  wchar v[8];
  DWORD n = GetEnvironmentVariableW(name, v, 8);
  return n == 1 && v[0] == '0';
}

// ---------------------------------------------------------------- state
enum { PH_BOOT = 0, PH_LOGO, PH_PRESS, PH_LOADING, PH_GAME };
volatile int g_phase = PH_BOOT;
static const char* kPhase[] = {"arranque", "logo", "press-start", "cargando", "juego"};
const char* kPhaseName(int p) { return p >= 0 && p < 5 ? kPhase[p] : "?"; }
int prof_start(void);
void prof_stop(void);

#define ST_LANG 0x826FDD70u
#define ST_LOGO 0x826FDE48u
#define ST_TITLE 0x826FDEF0u
#define ST_D 0x826FE058u
#define ST_E 0x826FE138u
#define ST_PRESS 0x826FE1D8u
#define ST_FADE 0x826DF8B0u
#define ST_POST 0x826FE4A8u
#define ST_F 0x826FE4F8u

static const char* state_name(u32 s) {
  switch (s) {
    case ST_LANG: return "idioma";
    case ST_LOGO: return "logo EA";
    case ST_TITLE: return "título";
    case ST_D: return "D";
    case ST_E: return "E";
    case ST_PRESS: return "Press Start";
    case ST_FADE: return "fundido";
    case ST_POST: return "post-start";
    case ST_F: return "F";
  }
  return "?";
}

static volatile u32 g_state;       // current BootFlow handler
static volatile u32 g_boot;        // BootFlow object (r3 of the handlers)
static double g_state_t;
static volatile u32 g_frames;      // updates of the current state
static u8* g_base;                 // guest memory base (from the first hook call)

// when each step started (0 = not seen), for the summary at gameplay
static double g_t_lang, g_t_logo, g_t_title, g_t_press, g_t_fade;

static void set_phase(int p) {
  if (g_phase >= p) return;
  g_phase = p;
  logf("fase: %s", kPhase[p]);
}

#define MAX_SEEN 64
static u32 g_seen[MAX_SEEN];
static int g_nseen;
static CRITICAL_SECTION g_seen_cs;

// true the first time `key` is seen
static int first_time(u32 key) {
  int r = 1;
  EnterCriticalSection(&g_seen_cs);
  for (int i = 0; i < g_nseen; i++)
    if (g_seen[i] == key) {
      r = 0;
      break;
    }
  if (r && g_nseen < MAX_SEEN) g_seen[g_nseen++] = key;
  LeaveCriticalSection(&g_seen_cs);
  return r;
}

// ---------------------------------------------------------------- safe guest reads
// Guest pages that are not committed fault, so every read of a pointer we
// did not get from the game itself is checked first.
static int host_ok(u8* p, u32 n) {
  MEMORY_BASIC_INFORMATION m;
  if (!VirtualQuery(p, &m, sizeof m)) return 0;
  if (m.State != MEM_COMMIT || (m.Protect & 0x101) || !(m.Protect & 0xEE)) return 0;
  return p + n <= (u8*)m.BaseAddress + m.RegionSize;
}

static int gok(u32 a, u32 n) {
  if (!g_base || a < 0x10000 || a > 0xFFFF0000u) return 0;
  return host_ok(gptr(g_base, a), n);
}
static u32 rd32(u32 a) { return gok(a, 4) ? gload32(g_base, a) : 0; }
static int rd8(u32 a) { return gok(a, 1) ? gload8(g_base, a) : -1; }

// ---------------------------------------------------------------- hooks
static PPCFunc *o_movie, *o_chg, *o_fereq, *o_feset, *o_feaef0, *o_fedone, *o_fepop;
static PPCFunc *o_lang, *o_stB, *o_stC, *o_stD, *o_stE, *o_press, *o_after, *o_stA, *o_stF, *o_startcb;

#define G_FEFLOW 0x830CFE1Cu   // FE flow manager (GetInst)
#define G_FEUI 0x830CFE14u     // FE UI system
#define G_MOVIES 0x830CFE0Cu   // movie system
#define G_LOADER 0x83083BCCu   // [[x]+8]+0x44: state D waits for 3
#define G_START 0x83083973u    // START accepted on the title screen

static u32 screen(u32 id) {
  u32 inst = rd32(G_FEFLOW);
  u32 arr = inst ? rd32(inst + 8) : 0;
  return arr ? rd32(arr + id * 4) : 0;
}

// Movie "still playing?" query. While the EA logo state runs, the movie
// widget update gets "finished": the logo ends on its first frame.
static void h_movie(PPCContext* ctx, u8* base) {
  u32 lr = (u32)ctx->lr;
  if (g_skip_logo && g_state == ST_LOGO && lr == 0x82637094) {
    if (first_time(0x20000000 ^ lr)) logf("película del logo: la doy por terminada");
    ctx->r3.u64 = 0;
    return;
  }
  o_movie(ctx, base);
  if (first_time(0x10000000 ^ lr ^ (ctx->r3.u32 << 27)))
    logf("película: consulta desde %08X = %u (estado %s)", lr, ctx->r3.u32, state_name(g_state));
}

// BootFlow::ChangeState(this, handler, ...)
static void h_chg(PPCContext* ctx, u8* base) {
  u32 to = ctx->r4.u32;
  double t = now_s();
  if (!g_base) g_base = base;
  logf("BootFlow %s -> %s %08X (%.3f s, %u cuadros, %u lecturas de mando)", state_name(g_state),
       state_name(to), to, g_state ? t - g_state_t : 0.0, g_frames, syn_polls());
  g_state = to;
  g_state_t = t;
  g_frames = 0;
  if (to == ST_LOGO && !g_t_logo) g_t_logo = t;
  if (to == ST_TITLE && !g_t_title) g_t_title = t;
  if (to == ST_PRESS && !g_t_press) g_t_press = t;
  if (to == ST_FADE && !g_t_fade) g_t_fade = t;
  if (to == ST_LOGO) set_phase(PH_LOGO);
  if (to == ST_PRESS) set_phase(PH_PRESS);
  if (to == ST_FADE) set_phase(PH_LOADING);
  o_chg(ctx, base);
}

static void h_fereq(PPCContext* ctx, u8* base) {
  logf("FE pedir %u (lr=%08X)", ctx->r4.u32, (u32)ctx->lr);
  o_fereq(ctx, base);
}
static void h_feset(PPCContext* ctx, u8* base) {
  logf("FE poner %u modo %d (lr=%08X)", ctx->r4.u32, ctx->r5.s32, (u32)ctx->lr);
  o_feset(ctx, base);
}
static void h_feaef0(PPCContext* ctx, u8* base) {
  logf("FE apilar %u (lr=%08X)", ctx->r4.u32, (u32)ctx->lr);
  o_feaef0(ctx, base);
}
static void h_fepop(PPCContext* ctx, u8* base) {
  logf("FE sacar %u (lr=%08X)", ctx->r4.u32, (u32)ctx->lr);
  o_fepop(ctx, base);
}

// Press Start's handler for FE message 23 (START accepted).
static void h_startcb(PPCContext* ctx, u8* base) {
  logf("START aceptado por la pantalla (r3=%08X r4=%08X r5=%08X BootFlow=%08X, %.3f s en %s)", ctx->r3.u32,
       ctx->r4.u32, ctx->r5.u32, g_boot, now_s() - g_state_t, state_name(g_state));
  o_startcb(ctx, base);
}

static u8 g_done_last[128];
static void h_fedone(PPCContext* ctx, u8* base) {
  u32 id = ctx->r4.u32, lr = (u32)ctx->lr;
  o_fedone(ctx, base);
  u8 r = (u8)ctx->r3.u32;
  if (id < 128 && g_done_last[id] != (u8)(r + 1)) {
    g_done_last[id] = (u8)(r + 1);
    logf("FE %u activa=%u (lr=%08X)", id, r, lr);
  }
}

// per-state event logging, frame counting
static void state_ev(u32 st, PPCContext* ctx, u8* base) {
  if (!g_base) g_base = base;
  g_boot = ctx->r3.u32;
  u32 ev = ctx->r4.u32;
  if (ev == 1 && g_state == st) g_frames++;
  if (st == ST_LANG && ev == 0 && !g_t_lang) g_t_lang = now_s();
  if (ev != 1) logf("estado %s evento %u", state_name(st), ev);
}

static int g_logo_forced;
static void h_lang(PPCContext* c, u8* b) { state_ev(ST_LANG, c, b); o_lang(c, b); }
static void h_stB(PPCContext* c, u8* b) {
  state_ev(ST_LOGO, c, b);
  // Fallback when the logo is not the movie the widget asks about: after a
  // few frames the "logo done" flag of FE screen 5 is set by hand.
  if (g_skip_logo && c->r4.u32 == 1 && !g_logo_forced && now_s() - g_state_t > 0.2) {
    u32 s5 = screen(5);
    if (s5 && gok(s5 + 0x40, 1) && gload8(b, s5 + 0x40) == 0) {
      gstore8(b, s5 + 0x40, 1);
      g_logo_forced = 1;
      logf("logo EA: marcado como terminado a mano (pantalla 5 = %08X)", s5);
    }
  }
  o_stB(c, b);
}
static void h_stC(PPCContext* c, u8* b) { state_ev(ST_TITLE, c, b); o_stC(c, b); }
static void h_stD(PPCContext* c, u8* b) { state_ev(ST_D, c, b); o_stD(c, b); }
static void h_stE(PPCContext* c, u8* b) { state_ev(ST_E, c, b); o_stE(c, b); }
static void h_press(PPCContext* c, u8* b) { state_ev(ST_PRESS, c, b); o_press(c, b); }
static void h_after(PPCContext* c, u8* b) { state_ev(ST_FADE, c, b); o_after(c, b); }
static void h_stA(PPCContext* c, u8* b) { state_ev(ST_POST, c, b); o_stA(c, b); }
static void h_stF(PPCContext* c, u8* b) { state_ev(ST_F, c, b); o_stF(c, b); }

static const HookSpec kHooks[] = {
  {"movie", 0xE41C80, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_movie, &o_movie},  // sub_825D60C8
  {"chg", 0x13E27D0, 5, {0x41, 0x56, 0x56, 0x57, 0x53}, h_chg, &o_chg},  // sub_826DFB30
  {"fereq", 0x3712E50, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_fereq, &o_fereq},  // sub_82D0ABE0
  {"feset", 0x3714230, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_feset, &o_feset},  // sub_82D0AFA0
  {"feaef0", 0x3713E60, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_feaef0, &o_feaef0},  // sub_82D0AEF0
  {"fepop", 0x3712AB0, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_fepop, &o_fepop},  // sub_82D0AB30
  {"fedone", 0x370CAB0, 5, {0x8B, 0x01, 0x83, 0xC0, 0x10}, h_fedone, &o_fedone},  // sub_82D09A40
  {"lang", 0x14892A0, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_lang, &o_lang},  // sub_826FDD70
  {"stB", 0x14896C0, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_stB, &o_stB},  // sub_826FDE48
  {"stC", 0x1489940, 5, {0x41, 0x57, 0x41, 0x56, 0x56}, h_stC, &o_stC},  // sub_826FDEF0
  {"stD", 0x148A1D0, 6, {0x41, 0x57, 0x41, 0x56, 0x41, 0x54}, h_stD, &o_stD},  // sub_826FE058
  {"stE", 0x148A660, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_stE, &o_stE},  // sub_826FE138
  {"press", 0x148A950, 6, {0x41, 0x57, 0x41, 0x56, 0x41, 0x55}, h_press, &o_press},  // sub_826FE1D8
  {"after", 0x13E1AE0, 6, {0x56, 0x57, 0x48, 0x83, 0xEC, 0x28}, h_after, &o_after},  // sub_826DF8B0
  {"stA", 0x148B7C0, 5, {0x8B, 0x41, 0x20, 0x85, 0xC0}, h_stA, &o_stA},  // sub_826FE4A8
  {"stF", 0x148B910, 6, {0x41, 0x57, 0x41, 0x56, 0x41, 0x54}, h_stF, &o_stF},  // sub_826FE4F8
  {"startcb", 0x148C1B0, 8, {0x48, 0xC7, 0x41, 0x50, 0x00, 0x00, 0x08, 0x83}, h_startcb, &o_startcb},  // sub_826FE638
};

// ---------------------------------------------------------------- monitor
// Polls the guest state the boot waits on and logs every change, so one run
// shows what each wait is made of.
typedef struct {
  const char* name;
  i64 last;
} Watch;

static void watch(Watch* w, i64 v) {
  if (v == w->last) return;
  w->last = v;
  logf("  %s = %lld", w->name, v);
}

static void hexdump(const char* what, u32 a, u32 n) {
  if (!a || !gok(a, n)) return;
  char line[160];
  for (u32 o = 0; o < n; o += 32) {
    int k = fmt(line, sizeof line, "  %s+%02X:", what, o);
    for (u32 i = 0; i < 32 && o + i < n; i += 4)
      k += fmt(line + k, sizeof line - k, " %08X", gload32(g_base, a + o + i));
    logf("%s", line);
  }
}

static void log_screen(u32 id, u32 s) {
  u32 vt = rd32(s);
  char line[400];
  int k = fmt(line, sizeof line, "pantalla %u = %08X vtabla %08X:", id, s, vt);
  if (vt && gok(vt, 40 * 4))
    for (u32 i = 0; i < 40; i++) k += fmt(line + k, sizeof line - k, " %08X", gload32(g_base, vt + i * 4));
  logf("%s", line);
}

static u32 list_ids(u32 lo, u32 hi, char* out, int cap) {
  int k = 0;
  u32 h = 0;
  out[0] = 0;
  if (hi < lo || hi - lo > 20 * 32) return 0xFFFFFFFF;
  for (u32 p = lo; p < hi; p += 20) {
    u32 id = rd32(p);
    h = h * 131 + id + 1;
    if (k < cap - 12) k += fmt(out + k, cap - k, "%u ", id);
  }
  return h;
}

static Watch w_s5f = {"pantalla5+40 (logo listo)", -2}, w_s22f = {"pantalla22+3C (press start lista)", -2},
             w_start = {"START aceptado", -2}, w_ui = {"transición FE", -2}, w_loader = {"cargador+44", -2},
             w_mv0 = {"película+320", -2}, w_mv1 = {"película+321", -2}, w_mv2 = {"película+322", -2},
             w_mvq = {"película en cola", -2}, w_b33 = {"BootFlow+21", -2}, w_b49 = {"BootFlow+31", -2},
             w_g6c = {"global 8308396C", -2}, w_ctx = {"contexto de juego", -2},
             w_tap = {"toque automático (botones, activo)", -2}, w_step = {"toque: paso", -2};

static DWORD WINAPI monitor(void* arg) {
  (void)arg;
  u32 last_scr[64] = {0};
  static const u32 kScr[] = {5, 22, 24, 47};
  u32 h_act = 0, h_pend = 0;
  double next_dump = 0;
  for (int i = 0; i < 40000; i++) {  // 3 ms ticks, up to 2 minutes
    if (!p_gameplay_ctx) resolve_runtime();
    u32 gc = p_gameplay_ctx ? p_gameplay_ctx() : 0xFFFFFFFF;
    watch(&w_ctx, (i32)gc);
    if (gc == 1 && g_phase < PH_GAME) {
      set_phase(PH_GAME);
      double tg = now_s();
      logf("EN EL JUEGO a los %.3f s desde que abrió", tg);
      logf("Resumen: idioma a los %.2f s | idioma %.2f s | logo %.2f s | título hasta Press Start %.2f s | "
           "Press Start %.2f s | carga del mundo %.2f s | total %.2f s",
           g_t_lang, g_t_logo && g_t_lang ? g_t_logo - g_t_lang : 0, g_t_title && g_t_logo ? g_t_title - g_t_logo : 0,
           g_t_press && g_t_title ? g_t_press - g_t_title : 0, g_t_fade && g_t_press ? g_t_fade - g_t_press : 0,
           g_t_fade ? tg - g_t_fade : 0, tg);
      prof_stop();
    }
    if (g_base) {
      for (u32 k = 0; k < sizeof kScr / sizeof kScr[0]; k++) {
        u32 s = screen(kScr[k]);
        if (s != last_scr[kScr[k]]) {
          last_scr[kScr[k]] = s;
          if (s) log_screen(kScr[k], s);
        }
      }
      u32 s5 = last_scr[5], s22 = last_scr[22];
      watch(&w_s5f, s5 ? rd8(s5 + 0x40) : -1);
      watch(&w_s22f, s22 ? rd8(s22 + 0x3C) : -1);
      watch(&w_start, rd8(G_START));
      watch(&w_g6c, rd8(0x8308396C));
      u32 ui = rd32(G_FEUI);
      u32 tr = ui ? rd32(ui + 0x108) : 0;
      watch(&w_ui, tr ? (i32)rd32(tr + 0xC) : -1);
      u32 ld = rd32(G_LOADER);
      u32 lo = ld ? rd32(ld + 8) : 0;
      watch(&w_loader, lo ? (i32)rd32(lo + 0x44) : -1);
      u32 mvs = rd32(G_MOVIES);
      u32 mv = mvs ? rd32(mvs + 0x1C) : 0;
      watch(&w_mv0, mv ? rd8(mv + 0x320) : -1);
      watch(&w_mv1, mv ? rd8(mv + 0x321) : -1);
      watch(&w_mv2, mv ? rd8(mv + 0x322) : -1);
      watch(&w_mvq, mv ? (rd32(mv + 0x22C) != mv + 0x22C) : -1);
      if (g_rt_syn) {
        u32 btn = *(volatile u16*)(g_rt_syn + 0x11A4C70), on = *(volatile u8*)(g_rt_syn + 0x11A4C72);
        watch(&w_tap, (i64)btn << 8 | on);
        // each START press while Press Start waits (step 0 = pressed, 1 = released)
        if (g_state == ST_PRESS)
          watch(&w_step, (*(volatile u8*)(g_rt_syn + 0x11A4D50) & 1) ? (i64)*(volatile u64*)(g_rt_syn + 0x11A4C80) : -1);
      }
      u32 b = g_boot;
      watch(&w_b33, b ? rd8(b + 0x21) : -1);
      watch(&w_b49, b ? rd8(b + 0x31) : -1);
      u32 inst = rd32(G_FEFLOW);
      if (inst) {
        char a[200], p[200];
        u32 ha = list_ids(rd32(inst + 16), rd32(inst + 20), a, sizeof a);
        u32 hp = list_ids(rd32(inst + 32), rd32(inst + 36), p, sizeof p);
        if (ha != h_act || hp != h_pend) {
          h_act = ha;
          h_pend = hp;
          logf("  FE activas [ %s] pendientes [ %s]", a, p);
        }
      }
      // snapshots of the screens the boot waits on, twice a second
      double t = now_s();
      if ((g_state == ST_LOGO || g_state == ST_PRESS) && t >= next_dump) {
        next_dump = t + 0.5;
        u32 s = g_state == ST_LOGO ? s5 : s22;
        logf("foto de la pantalla %u (%s, %.2f s en el estado):", g_state == ST_LOGO ? 5 : 22,
             state_name(g_state), t - g_state_t);
        hexdump("p", s, 0x80);
      }
    }
    if (g_phase == PH_GAME && i % 300 == 0) break;
    Sleep(3);
  }
  return 0;
}

int game_install(void) {
  InitializeCriticalSection(&g_seen_cs);
  if (env_off(L"SKATE3_DIRECT_LOGO")) g_skip_logo = 0;
  logf("corte del logo EA: %s", g_skip_logo ? "sí" : "no");
  for (usize i = 0; i < sizeof kHooks / sizeof kHooks[0]; i++) {
    if (!check_bytes(kHooks[i].rva, kHooks[i].expect, kHooks[i].len)) {
      logf("no coincide %s: no toco nada", kHooks[i].name);
      return 0;
    }
  }
  int ok = 1;
  for (usize i = 0; i < sizeof kHooks / sizeof kHooks[0]; i++) ok &= hook_entry(&kHooks[i]);
  CreateThread(NULL, 0, monitor, NULL, 0, NULL);
  prof_start();
  return ok;
}
