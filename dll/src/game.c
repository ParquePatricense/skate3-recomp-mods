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

typedef struct {
  u16 buttons;
  u32 polls;
  u8 lt, rt;
} SynStep;  // rex::kernel::xam::SyntheticInputStep
_Static_assert(sizeof(SynStep) == 12 && __builtin_offsetof(SynStep, polls) == 4, "SyntheticInputStep");
typedef void (*QueueSyn_t)(const SynStep* steps, u64 n);
static QueueSyn_t p_queue_syn;

static void resolve_runtime(void) {
  HMODULE rt = GetModuleHandleW(L"rexruntime.dll");
  if (!rt) return;
  if (!p_queue_syn)
    p_queue_syn = (QueueSyn_t)GetProcAddress(
        rt, "?QueueSyntheticInputSequence@xam@kernel@rex@@YAXPEBUSyntheticInputStep@123@_K@Z");
  p_gameplay_ctx = (GameplayContextValue_t)GetProcAddress(
      rt, "?GameplayContextValue@guest_presence@kernel@rex@@YAIXZ");
  u8* b = (u8*)rt;
  u8* nt = b + ((IMAGE_DOS_HEADER*)b)->e_lfanew;
  if (*(u32*)(nt + 8) == RT_TIMESTAMP && *(u32*)(nt + 24 + 56) == RT_IMAGESIZE) g_rt_syn = b;
}
static u32 syn_polls(void) { return g_rt_syn ? *(volatile u32*)(g_rt_syn + 0x11A4C8C) : 0; }

// ---------------------------------------------------------------- options
static int g_skip_logo = 1;
static int g_uncap = 1;      // no frame cap while booting/loading
static int g_fast_fade = 0;  // run the post-START sub-machine several steps per frame (off: it is
                             // save/profile work, not a fade, and gained nothing)
static int g_fast_press = 0; // Press Start screen activated at once (off: v3.9.2-3.9.4 broke the profile or looped)
static int g_press_inside = 0; // START given from inside: off, it skips the screen's own input
                               // handling (which pad/profile pressed it) and the game then
                               // boots into free play with the default skater and no save
static int g_profile = 0;    // sampling profiler (suspends every thread: slows the load, opt-in)
static int g_profile_world = 1;  // profiler on the loading threads during the world step only
static int g_press_clock = 1; // Press Start's 1.5 s minimum on screen counted as already elapsed
static int g_press_tap = 1;   // START tapped the moment Press Start takes it (same synthetic input the recomp uses)
static int g_fast_black = 1;  // fade to black before gameplay (82709338) finished on its second frame
static int g_no_nap = 1;     // loading threads poll with SleepEx(1): yield instead while loading
static int g_no_query = 0;   // no host occlusion queries while loading (v3.9.12: no gain, opt-in)
static int g_boost = 1;      // GPU Commands + render_thread at high priority while loading
static int g_fast_tick = 1;  // game tick timer at 4 ms instead of ~16.7 ms from the language screen to START
static volatile int g_tick_on;
int boost_threads(void);
void boost_restore(void);
static int g_boosted;

// ---------------------------------------------------------------- loading frame cap
// The recomp paces every guest frame to the display (skate3_guest_fps_cap_auto
// / skate3_guest_fps_cap, src/skate3_native_render.cpp PaceGuestFrame): a
// frame that finished early sleeps until the next beat. While booting and
// loading that sleep only delays the work done frame by frame, so the cap is
// lifted at the first boot state and the player's values go back once in
// gameplay. Only the in-memory values change; skate3.toml is not written.
typedef struct {
  const char* p;
  u64 n;
} StrView;  // std::string_view (passed by address in the x64 ABI)
typedef u8 (*SetFlag_t)(StrView* name, StrView* value);
typedef void* (*GetFlag_t)(void* out, StrView* name);
static SetFlag_t p_setflag;
static GetFlag_t p_getflag;
static volatile int g_uncapped;
static char g_cap_auto[24], g_cap[24];

static StrView sv(const char* s) {
  StrView v = {s, s_len(s)};
  return v;
}

// rex::cvar::GetFlagByName returns a std::string (MSVC layout: 16-byte
// inline buffer or heap pointer, size at +16, capacity at +24). A heap copy
// is left alone rather than freed with the wrong allocator.
static int get_flag(const char* name, char* out, usize cap) {
  u8 str[32];
  for (int i = 0; i < 32; i++) str[i] = 0;
  StrView n = sv(name);
  p_getflag(str, &n);
  u64 len = *(u64*)(str + 16), res = *(u64*)(str + 24);
  const char* p = res > 15 ? *(const char**)str : (const char*)str;
  if (!len || len >= cap || !p) return 0;
  for (u64 i = 0; i < len; i++) out[i] = p[i];
  out[len] = 0;
  return 1;
}

static int set_flag(const char* name, const char* value) {
  StrView n = sv(name), v = sv(value);
  return p_setflag(&n, &v) != 0;
}

static int flags_ready(void) {
  if (p_setflag && p_getflag) return 1;
  HMODULE rt = GetModuleHandleW(L"rexruntime.dll");
  if (!rt) return 0;
  p_setflag = (SetFlag_t)GetProcAddress(
      rt, "?SetFlagByName@cvar@rex@@YA_NV?$basic_string_view@DU?$char_traits@D@std@@@std@@0@Z");
  p_getflag = (GetFlag_t)GetProcAddress(
      rt, "?GetFlagByName@cvar@rex@@YA?AV?$basic_string@DU?$char_traits@D@std@@V?$allocator@D@2@@std@@"
          "V?$basic_string_view@DU?$char_traits@D@std@@@4@@Z");
  return p_setflag && p_getflag;
}

static void uncap_begin(void) {
  if (!g_uncap || g_uncapped) return;
  if (!flags_ready()) {
    logf("tope de FPS: no encontré los ajustes del recomp, lo dejo como está");
    return;
  }
  if (!get_flag("skate3_guest_fps_cap_auto", g_cap_auto, sizeof g_cap_auto) ||
      !get_flag("skate3_guest_fps_cap", g_cap, sizeof g_cap)) {
    logf("tope de FPS: no pude leer los valores actuales, lo dejo como está");
    return;
  }
  if (set_flag("skate3_guest_fps_cap_auto", "false") && set_flag("skate3_guest_fps_cap", "0")) {
    g_uncapped = 1;
    logf("tope de FPS: fuera mientras carga (tenías auto=%s, tope=%s)", g_cap_auto, g_cap);
  } else {
    set_flag("skate3_guest_fps_cap_auto", g_cap_auto);
    set_flag("skate3_guest_fps_cap", g_cap);
    logf("tope de FPS: el recomp no aceptó el cambio, lo dejo como está");
  }
}

static void uncap_end(void) {
  if (!g_uncapped) return;
  g_uncapped = 0;
  int ok = set_flag("skate3_guest_fps_cap", g_cap) & set_flag("skate3_guest_fps_cap_auto", g_cap_auto);
  logf("tope de FPS: vuelve a auto=%s, tope=%s%s", g_cap_auto, g_cap, ok ? "" : " (falló)");
}

// ---------------------------------------------------------------- loading occlusion queries
// With occlusion_query_enable the recomp's D3D12 backend ends every guest
// query with a submission and then waits on the host GPU fence for it
// (D3D12CommandProcessor::EndGuestOcclusionQuery -> CheckSubmissionFence), so
// each query is a full CPU/GPU round trip. While loading nothing the queries
// gate is on screen; the recomp's own fallback (query_occlusion_fake_sample_count,
// "visible") answers them instead, and the player's value is back in gameplay.
static volatile int g_query_off;
static char g_query_was[8];

static void query_begin(void) {
  if (!g_no_query || g_query_off) return;
  if (!flags_ready() || !get_flag("occlusion_query_enable", g_query_was, sizeof g_query_was)) {
    logf("consultas de oclusión: no pude leer el ajuste del recomp, lo dejo como está");
    return;
  }
  if (g_query_was[0] == 'f' || g_query_was[0] == '0') {
    logf("consultas de oclusión: ya estaban apagadas en el recomp");
    return;
  }
  if (set_flag("occlusion_query_enable", "false")) {
    g_query_off = 1;
    logf("consultas de oclusión: fuera mientras carga (tenías %s)", g_query_was);
  } else {
    logf("consultas de oclusión: el recomp no aceptó el cambio, lo dejo como está");
  }
}

static void query_end(void) {
  if (!g_query_off) return;
  g_query_off = 0;
  int ok = set_flag("occlusion_query_enable", g_query_was);
  logf("consultas de oclusión: vuelven a %s%s", g_query_was, ok ? "" : " (falló)");
}

static int env_on(const wchar* name) {
  wchar v[8];
  DWORD n = GetEnvironmentVariableW(name, v, 8);
  return n == 1 && v[0] == '1';
}

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
int prof_start_focus(void);
void prof_window(int on);
int load_map(void);
u32 host_of(u32 guest);
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
#define ST_WORLD 0x827077D8u  // world load step (loader +0x44 runs 8..18)

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
static volatile long g_swaps, g_swaps_world;
static double g_swap_t, g_swap_gap;
static int g_swap_hooked;
static void h_chg(PPCContext* ctx, u8* base) {
  u32 to = ctx->r4.u32;
  double t = now_s();
  if (!g_base) g_base = base;
  if (g_swap_hooked && g_state == ST_WORLD && to != ST_WORLD)
    logf("paso del mundo: %d cuadros del juego (%.0f por segundo), el hueco más largo entre cuadros %.0f ms",
         (int)g_swaps_world, g_swaps_world / (t - g_state_t), g_swap_gap * 1000);
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
  if (g_profile_world && !g_profile) {
    static int win;
    if (to == ST_WORLD && !win) {
      win = 1;
      prof_window(1);
      logf("perfil: empieza el paso del mundo");
    } else if (win == 1 && to != ST_WORLD) {
      win = 2;
      prof_window(0);
      prof_stop();
      logf("perfil: termina el paso del mundo");
    }
  }
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
static double g_logo_wait = 0.03;  // SKATE3_DIRECT_LOGO_LENTO=1: 0.2 s
static void h_lang(PPCContext* c, u8* b) {
  if (c->r4.u32 == 0) {
    g_tick_on = 1;
    uncap_begin();
    query_begin();
  }
  state_ev(ST_LANG, c, b);
  o_lang(c, b);
}
static void h_stB(PPCContext* c, u8* b) {
  state_ev(ST_LOGO, c, b);
  // The logo is not the movie the widget asks about: screen 5's update
  // (sub_82608E38) only adds dt to +0x3C and sets +0x40 at 4000 ms, and its
  // skip (sub_82608E18) sets the same byte, so it is set by hand once the
  // screen has had a few ticks (0.03 s; it was 0.2 s up to v3.9.16).
  if (g_skip_logo && c->r4.u32 == 1 && !g_logo_forced && now_s() - g_state_t > g_logo_wait) {
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
// Press Start. Screen 22's input handler (vtable slot 10, sub_82609238)
// only takes START once the screen is active (+0x44, set through slot 11,
// sub_82608EA0, when its intro ends ~1.5 s in). START there picks the pad and
// profile, sets the "ready" flag (+0x3C) and starts the profile/save load
// (sub_82609188); BootFlow then waits for FE message 23 (sub_826FE638).
// With fast_press the screen is activated by calling slot 11 as soon as it is
// up, so the recomp's real START press is taken right away. Neither the
// ready flag nor the START flag is written by hand: that skipped the pad and
// profile choice and the game booted into free play without the save.
static int g_press_active, g_press_start, g_press_seen;
static PPCFunc* g_setactive;  // host code of sub_82608EA0
static void call_guest(PPCFunc* f, PPCContext* c, u8* b, u32 r3, u32 r4) {
  static PPCContext save;  // only the guest main thread gets here
  save = *c;
  c->r3.u64 = r3;
  c->r4.u64 = r4;
  f(c, b);
  *c = save;
}
static void h_press(PPCContext* c, u8* b) {
  state_ev(ST_PRESS, c, b);
  u32 self = c->r3.u32;
  if (g_fast_press && c->r4.u32 == 1 && g_phase < PH_GAME) {
    u32 s22 = screen(22);
    if (s22 && gok(s22 + 0x44, 4)) {
      u32 act = gload32(b, s22 + 0x44);
      if (act && !g_press_seen) {
        g_press_seen = 1;
        logf("Press Start: pantalla activa (%.3f s en el estado%s)", now_s() - g_state_t,
             g_press_active ? ", la activé yo" : ", sola");
      }
      if (!act && !g_press_active && g_setactive && now_s() - g_state_t > 0.1) {
        g_press_active = 1;
        call_guest(g_setactive, c, b, s22, 1);
        logf("Press Start: activo la pantalla 22 sin esperar su animación (%.3f s en el estado)",
             now_s() - g_state_t);
      }
    }
    if (g_press_inside && !g_press_start && gok(self + 0x31, 2) && gload8(b, self + 0x31) == 1 && rd8(G_START) == 0) {
      gstore8(b, self + 0x32, 1);
      gstore8(b, G_START, 1);
      g_press_start = 1;
      logf("Press Start: START dado por dentro (%.3f s en el estado)", now_s() - g_state_t);
    }
  }
  o_press(c, b);
}
// The state after START runs a sub-machine ([BootFlow+20], started by the
// Press Start handler) until it reports done, then switches to post-start.
// With the turbo on, each frame steps it up to 8 more times, stopping as soon
// as BootFlow has left this state.
static void h_after(PPCContext* c, u8* b) {
  state_ev(ST_FADE, c, b);
  u32 self = c->r3.u32, ev = c->r4.u32;
  o_after(c, b);
  if (!g_fast_fade || ev != 1 || g_phase >= PH_GAME) return;
  int n = 0;
  while (n < 8 && rd32(self + 8) == ST_FADE && rd32(self + 20)) {
    c->r3.u64 = self;
    c->r4.u64 = 1;
    o_after(c, b);
    n++;
  }
  if (rd32(self + 8) != ST_FADE && first_time(0x30000000))
    logf("fundido: terminado en %.3f s (%u cuadros, turbo)", now_s() - g_state_t, g_frames);
}
static void h_stA(PPCContext* c, u8* b) { state_ev(ST_POST, c, b); o_stA(c, b); }
static void h_stF(PPCContext* c, u8* b) { state_ev(ST_F, c, b); o_stF(c, b); }


// ---------------------------------------------------------------- Press Start minimum time
// Screen 22's update (sub_82609050, r4 = dt in ms) adds dt to the global
// 0x8308355C until it passes 1500; only then it sets +0x48, and the input
// handler (sub_82609238) takes START only with +0x48 set. Nothing else reads
// that global (it is zeroed again when START is taken), so it is a plain
// minimum display time: the first update finds it already elapsed. START,
// pad and profile still go through the screen's own input handling.
#define PS_CLOCK 0x8308355Cu
static PPCFunc* o_psupd;
static void h_psupd(PPCContext* ctx, u8* base) {
  if (g_press_clock && g_phase == PH_PRESS && gload32(base, PS_CLOCK) <= 1500) {
    static int said;
    if (!said) {
      said = 1;
      logf("Press Start: reloj de 1,5 s dado por cumplido (iba en %u ms)", gload32(base, PS_CLOCK));
    }
    gstore32(base, PS_CLOCK, 1501);
  }
  u32 self = ctx->r3.u32;
  o_psupd(ctx, base);
  // The recomp taps START on a fixed 0.33 s cycle (5 polls down, 15 up), so
  // the screen can sit ready for most of a cycle. Once +0x48 is set, one tap
  // goes through the recomp's own synthetic input (same pad, same path); its
  // auto-tap carries on by itself afterwards.
  static int tapped;
  if (g_press_tap && !tapped && g_phase == PH_PRESS && gload8(base, self + 0x48)) {
    tapped = 1;
    if (!p_queue_syn) resolve_runtime();
    if (p_queue_syn) {
      static const SynStep tap[] = {{0x0010, 5, 0, 0}, {0, 5, 0, 0}};
      p_queue_syn(tap, 2);
      logf("Press Start: START al instante (la pantalla ya lo acepta)");
    } else {
      logf("Press Start: no encontré QueueSyntheticInputSequence, queda el toque del recomp");
    }
  }
}

// ---------------------------------------------------------------- last step before gameplay
// BootFlow 82709338 waits for the fade object [[0x830CFE14]+304] to reach
// black: its +12 climbs 0.04 per frame from 0 to 1 (25 frames, ~0.42 s; seen
// in Matias's v3.9.9 log), and the state moves on once it is 1. The first
// update puts it one step short of the end, so the game's own fade code
// finishes it on the next frame and everything after runs as always.
static PPCFunc* o_prego;
static float f32_of(u32 v) { return *(float*)&v; }
static void h_prego(PPCContext* ctx, u8* base) {
  static int lines;
  u32 inst = rd32(0x830CFE14);
  u32 fade = inst ? rd32(inst + 304) : 0;
  if (fade && ctx->r4.u32 <= 1 && lines < 4) {
    lines++;
    logf("antes del juego ev%u: fundido %.3f", ctx->r4.u32, (double)f32_of(rd32(fade + 12)));
  }
  if (g_fast_black && fade && ctx->r4.u32 == 1 && g_phase < PH_GAME) {
    float cur = f32_of(rd32(fade + 12));
    if (cur > 0.0f && cur < 0.96f) {
      float to = 0.96f;
      gstore32(base, fade + 12, *(u32*)&to);
      static int said;
      if (!said) {
        said = 1;
        logf("antes del juego: fundido a negro adelantado (iba en %.3f)", (double)cur);
      }
    }
  }
  o_prego(ctx, base);
}

// ---------------------------------------------------------------- loader naps
// load_thread, decompress_thread and the two rwfilesys threads hand work to
// each other and poll for it with SleepEx(1, TRUE) (sub_82EE24A8, e.g. the
// load_thread loop in sub_828A3270). Each hand-off then costs a host sleep of
// 1 ms or more. While booting/loading those polls become SleepEx(0, TRUE):
// same call, same APC delivery, but the thread only yields. Other threads and
// other sleep lengths are left alone, and gameplay is never touched.
static PPCFunc* o_sleep;
typedef struct {
  DWORD tid;
  u8 kind;  // 1 loader thread, 2 other
} NapThread;
#define NAP_MAX 160
static NapThread g_nap[NAP_MAX];
static volatile long g_nap_n;
static volatile long g_nap_cut[4];  // load, decompress, rwfilesys
static volatile long g_nap_world;   // cut during the world step
void thread_name_of(HANDLE h, char* out, int cap);

static int nap_kind(void) {
  DWORD tid = GetCurrentThreadId();
  long n = g_nap_n;
  if (n > NAP_MAX) n = NAP_MAX;
  for (long i = 0; i < n; i++)
    if (g_nap[i].tid == tid) return g_nap[i].kind;
  char name[64];
  thread_name_of(GetCurrentThread(), name, sizeof name);
  int k = 0;
  if (memcmp(name, "load_thread", 11) == 0) k = 1;
  else if (memcmp(name, "decompress_thread", 17) == 0) k = 2;
  else if (memcmp(name, "rwfilesys", 9) == 0) k = 3;
  long i = __sync_fetch_and_add(&g_nap_n, 1);
  if (i < NAP_MAX) {
    g_nap[i].kind = (u8)(k ? k : 9);
    g_nap[i].tid = tid;
  }
  if (k) logf("siesta: hilo %u \"%s\" duerme sin esperar mientras carga", tid, name);
  return k ? k : 9;
}

static void h_sleep(PPCContext* ctx, u8* base) {
  if (g_no_nap && ctx->r3.u32 == 1 && g_phase < PH_GAME) {
    int k = nap_kind();
    if (k <= 3) {
      ctx->r3.u64 = 0;
      __sync_fetch_and_add(&g_nap_cut[k], 1);
      if (g_state == ST_WORLD) __sync_fetch_and_add(&g_nap_world, 1);
    }
  }
  o_sleep(ctx, base);
}

// The same 1 ms naps also come in through Sleep(ms) (sub_82EE24E0, straight
// into sub_82EEE988 without SleepEx): the file system's lock (sub_82A40368)
// spins TryEnterCriticalSection + Sleep(1) for up to 10 s, and the
// decompress_thread spent ~28% of the world step there. Same rule: loader
// threads, 1 ms, before gameplay -> yield.
static PPCFunc* o_delay;
static volatile long g_lock_cut[4];
static void h_delay(PPCContext* ctx, u8* base) {
  if (g_no_nap && ctx->r3.u32 == 1 && g_phase < PH_GAME) {
    int k = nap_kind();
    if (k <= 3) {
      ctx->r3.u64 = 0;
      __sync_fetch_and_add(&g_lock_cut[k], 1);
    }
  }
  o_delay(ctx, base);
}

// ---------------------------------------------------------------- guest frames
// VdSwap (xboxkrnl 0x25B) is reached through an import thunk in skate3.exe,
// `jmp [rip+x]` into rexruntime. Swapping that slot counts the game's frames,
// so the log says how many frames the world step took and the longest gap:
// GPU bound shows as few, slow frames.
static PPCFunc* o_swap;
static void h_swap(PPCContext* ctx, u8* base) {
  __sync_fetch_and_add(&g_swaps, 1);
  if (g_state == ST_WORLD) {
    double t = now_s();
    if (g_swaps_world && t - g_swap_t > g_swap_gap) g_swap_gap = t - g_swap_t;
    g_swap_t = t;
    __sync_fetch_and_add(&g_swaps_world, 1);
  }
  o_swap(ctx, base);
}

static int hook_swap(void) {
  u32 r = host_of(0x82F9CDE4);
  if (!r) return 0;
  u8* p = g_exe + r;
  if (p[0] != 0xFF || p[1] != 0x25) return 0;
  PPCFunc** slot = (PPCFunc**)(p + 6 + *(i32*)(p + 2));
  DWORD old;
  if (!VirtualProtect(slot, 8, PAGE_READWRITE, &old)) return 0;
  o_swap = *slot;
  *slot = h_swap;
  VirtualProtect(slot, 8, old, &old);
  return 1;
}

// ---------------------------------------------------------------- game tick
// The game's update tick runs on its own thread (sub_82966930): it waits on a
// waitable timer, ticks the app (BootFlow and the front end included), then
// re-arms the timer with SetWaitableTimer (sub_82EE2248) for one period
// (this+16 = -period in 100 ns; this+24 = base 60 Hz period, this+120 = the
// period corrected by the average error over 60 ticks: ring at this+136,
// average at this+128). Every boot step therefore costs at least one 16.7 ms
// tick even with nothing to do. From the language screen to START the
// timer is armed for 4 ms instead; the game's own value is left as is. After
// START the error history (which saw the short ticks) is cleared and the
// period goes back to the base. v3.9.15 kept it up to gameplay: the screens
// before START got 0.17 s faster, but after START the save/world steps got
// 0.3 s slower (the extra ticks take CPU from the loader threads).
#define TICK_FAST (-40000LL)
static PPCFunc* o_settimer;
static u32 g_tick_obj;
static volatile long g_tick_cut;
static int g_tick_reset, g_tick_fast_last;
static volatile long g_tick_quick;
// After START the fast tick stole CPU from the heavy load steps (v3.9.15),
// but most post-START BootFlow steps just hand over to the next one on their
// first update and wait a whole 16.7 ms tick to do it (Matias's v3.9.17 log:
// ~15 of them at 0.016 s, 0 frames). The tick after one of those is armed
// at 4 ms; the fade, the save/profile load, 8271AC38, 82704038, 8270C608,
// 82707508 and the world step keep the normal tick.
static int quick_state(u32 st) {
  static const u32 k[] = {0x826FE4F8, 0x826FA7B0, 0x826DF9A8, 0x82703E80, 0x82704630, 0x82715390,
                          0x82712D20, 0x82704350, 0x827023B0, 0x82702428, 0x82702650, 0x827076E0,
                          0x82708B78, 0x82708BE8, 0x82708CE0, 0x82709338, 0x82709068, 0x82708D00,
                          0x8270C890};
  for (u32 i = 0; i < sizeof k / sizeof k[0]; i++)
    if (k[i] == st) return 1;
  return 0;
}
static i64 gl64(u8* b, u32 a) { return (i64)(((u64)gload32(b, a) << 32) | gload32(b, a + 4)); }
static void gs64(u8* b, u32 a, i64 v) { gstore32(b, a, (u32)((u64)v >> 32)); gstore32(b, a + 4, (u32)v); }
static void h_settimer(PPCContext* ctx, u8* base) {
  u32 pdue = ctx->r4.u32;
  if (!g_base) g_base = base;
  if (!g_fast_tick || ctx->r5.u32 != 0 || !pdue || !gok(pdue, 8)) {
    o_settimer(ctx, base);
    return;
  }
  u32 obj = pdue - 16;
  i64 due = gl64(base, pdue);
  int is_tick = gok(obj, 1112) && gload32(base, obj + 8) == ctx->r3.u32 && due <= -50000 && due >= -400000;
  int pre = g_tick_on && g_phase < PH_LOADING;
  int quick = g_tick_obj && obj == g_tick_obj && g_phase == PH_LOADING && quick_state(g_state);
  if (is_tick && (pre || quick)) {
    if (!g_tick_obj) {
      g_tick_obj = obj;
      logf("tick del juego: cada 4 ms hasta el START (era %.2f ms)", -due / 10000.0);
    }
    gs64(base, pdue, TICK_FAST);
    __sync_fetch_and_add(pre ? &g_tick_cut : &g_tick_quick, 1);
    g_tick_fast_last = 1;
    o_settimer(ctx, base);
    gs64(base, pdue, due);
    return;
  }
  // Back to the normal tick: the game's drift correction (error ring at
  // obj+128..615, corrected period at obj+120) saw the short ticks as early
  // wakeups, so it is cleared and the period goes back to the base one.
  if (is_tick && obj == g_tick_obj && g_phase >= PH_LOADING && g_tick_fast_last) {
    g_tick_fast_last = 0;
    i64 period = gl64(base, obj + 24);
    for (u32 o = 128; o < 616; o += 8) gs64(base, obj + o, 0);
    gs64(base, obj + 120, period);
    gs64(base, pdue, -period);
    if (!g_tick_reset) {
      g_tick_reset = 1;
      logf("tick del juego: vuelve a %.2f ms (%d ticks cortos)", period / 10000.0, (int)g_tick_cut);
    }
  }
  o_settimer(ctx, base);
}

static const HookSpec kHooks[] = {
  {"movie", 0xE41C80, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_movie, &o_movie, 0x825D60C8},  // sub_825D60C8
  {"chg", 0x13E27D0, 5, {0x41, 0x56, 0x56, 0x57, 0x53}, h_chg, &o_chg, 0x826DFB30},  // sub_826DFB30
  {"fereq", 0x3712E50, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_fereq, &o_fereq, 0x82D0ABE0},  // sub_82D0ABE0
  {"feset", 0x3714230, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_feset, &o_feset, 0x82D0AFA0},  // sub_82D0AFA0
  {"feaef0", 0x3713E60, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_feaef0, &o_feaef0, 0x82D0AEF0},  // sub_82D0AEF0
  {"fepop", 0x3712AB0, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_fepop, &o_fepop, 0x82D0AB30},  // sub_82D0AB30
  {"fedone", 0x370CAB0, 5, {0x8B, 0x01, 0x83, 0xC0, 0x10}, h_fedone, &o_fedone, 0x82D09A40},  // sub_82D09A40
  {"lang", 0x14892A0, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_lang, &o_lang, 0x826FDD70},  // sub_826FDD70
  {"stB", 0x14896C0, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_stB, &o_stB, 0x826FDE48},  // sub_826FDE48
  {"stC", 0x1489940, 5, {0x41, 0x57, 0x41, 0x56, 0x56}, h_stC, &o_stC, 0x826FDEF0},  // sub_826FDEF0
  {"stD", 0x148A1D0, 6, {0x41, 0x57, 0x41, 0x56, 0x41, 0x54}, h_stD, &o_stD, 0x826FE058},  // sub_826FE058
  {"stE", 0x148A660, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_stE, &o_stE, 0x826FE138},  // sub_826FE138
  {"press", 0x148A950, 6, {0x41, 0x57, 0x41, 0x56, 0x41, 0x55}, h_press, &o_press, 0x826FE1D8},  // sub_826FE1D8
  {"after", 0x13E1AE0, 6, {0x56, 0x57, 0x48, 0x83, 0xEC, 0x28}, h_after, &o_after, 0x826DF8B0},  // sub_826DF8B0
  {"stA", 0x148B7C0, 5, {0x8B, 0x41, 0x20, 0x85, 0xC0}, h_stA, &o_stA, 0x826FE4A8},  // sub_826FE4A8
  {"stF", 0x148B910, 6, {0x41, 0x57, 0x41, 0x56, 0x41, 0x54}, h_stF, &o_stF, 0x826FE4F8},  // sub_826FE4F8
  {"psupd", 0xF4DC10, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_psupd, &o_psupd, 0x82609050},  // sub_82609050 Press Start update
  {"prego", 0x14C1530, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x30}, h_prego, &o_prego, 0x82709338},  // sub_82709338
  {"sleep", 0x425CF30, 6, {0x56, 0x57, 0x48, 0x83, 0xEC, 0x28}, h_sleep, &o_sleep, 0x82EE24A8},  // sub_82EE24A8 SleepEx
  {"settimer", 0x425C270, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_settimer, &o_settimer, 0x82EE2248},  // sub_82EE2248 SetWaitableTimer
  {"delay", 0x4302070, 7, {0x56, 0x57, 0x53, 0x48, 0x83, 0xEC, 0x20}, h_delay, &o_delay, 0x82EEE988},  // sub_82EEE988 (Sleep/SleepEx)
  {"startcb", 0x148C1B0, 8, {0x48, 0xC7, 0x41, 0x50, 0x00, 0x00, 0x08, 0x83}, h_startcb, &o_startcb, 0x826FE638},  // sub_826FE638
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
             w_tap = {"toque automático (botones, activo)", -2}, w_step = {"toque: paso", -2},
             w_s22on = {"pantalla22+44 (activa)", -2}, w_s22rdy = {"pantalla22+48 (acepta START)", -2},
             w_s22btn = {"pantalla22+50 (botón apretado)", -2}, w_psclk = {"reloj Press Start (ms)", -2},
             w_s29a = {"pantalla29+44 (antes del idioma)", -2}, w_s29b = {"pantalla29+48 (antes del idioma)", -2};

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
      uncap_end();
      query_end();
      prof_stop();
      if (g_no_nap)
        logf("siesta: esperas de 1 ms salteadas: load_thread %d, decompress %d, rwfilesys %d (en el paso del mundo %d)",
             (int)g_nap_cut[1], (int)g_nap_cut[2], (int)g_nap_cut[3], (int)g_nap_world);
      if (g_fast_tick)
        logf("tick del juego: %d ticks de 4 ms en los pasos cortos después del START", (int)g_tick_quick);
      if (g_no_nap)
        logf("siesta: esperas de 1 ms del candado de archivos salteadas: load_thread %d, decompress %d, rwfilesys %d",
             (int)g_lock_cut[1], (int)g_lock_cut[2], (int)g_lock_cut[3]);
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
      // State 826FDC50 (before language, ~0.6 s with no frames) waits for
      // both of these to stop being -1.
      if (g_phase == PH_BOOT && !g_t_lang) {
        u32 s29 = screen(29);
        watch(&w_s29a, s29 ? (i32)rd32(s29 + 0x44) : -2);
        watch(&w_s29b, s29 ? (i32)rd32(s29 + 0x48) : -2);
      }
      watch(&w_s5f, s5 ? rd8(s5 + 0x40) : -1);
      watch(&w_s22f, s22 ? rd8(s22 + 0x3C) : -1);
      if (g_phase == PH_PRESS && s22) {
        watch(&w_s22on, (i32)rd32(s22 + 0x44));
        watch(&w_s22rdy, rd8(s22 + 0x48));
        watch(&w_s22btn, rd8(s22 + 0x50));
        u32 clk = rd32(PS_CLOCK);
        watch(&w_psclk, clk > 1500 ? 1501 : (i32)(clk / 250 * 250));
      }
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
    if (g_boost && !g_boosted && g_phase == PH_LOADING) {
      g_boosted = 1;
      if (!boost_threads()) logf("prioridad: no encontré los hilos de GPU, la dejo como está");
    }
    if (g_boosted == 1 && g_phase >= PH_GAME) {
      g_boosted = 2;
      boost_restore();
    }
    if (g_phase == PH_GAME && i % 300 == 0) break;
    Sleep(3);
  }
  uncap_end();  // never left without the player's cap, even if gameplay was not seen
  query_end();
  if (g_boosted == 1) boost_restore();
  return 0;
}

int game_install(void) {
  InitializeCriticalSection(&g_seen_cs);
  if (env_off(L"SKATE3_DIRECT_LOGO")) g_skip_logo = 0;
  if (env_on(L"SKATE3_DIRECT_LOGO_LENTO")) g_logo_wait = 0.2;
  if (env_off(L"SKATE3_DIRECT_FPS")) g_uncap = 0;
  if (env_on(L"SKATE3_DIRECT_FADE")) g_fast_fade = 1;
  if (env_on(L"SKATE3_DIRECT_PRESS")) g_fast_press = 1;
  if (env_on(L"SKATE3_DIRECT_START")) g_press_inside = 1;
  if (env_on(L"SKATE3_DIRECT_PERFIL")) g_profile = 1;
  if (env_off(L"SKATE3_DIRECT_PERFIL_MUNDO")) g_profile_world = 0;
  if (env_off(L"SKATE3_DIRECT_SIESTA")) g_no_nap = 0;
  if (env_off(L"SKATE3_DIRECT_RELOJ")) g_press_clock = 0;
  if (env_off(L"SKATE3_DIRECT_TOQUE")) g_press_tap = 0;
  if (env_off(L"SKATE3_DIRECT_NEGRO")) g_fast_black = 0;
  if (env_on(L"SKATE3_DIRECT_CONSULTAS")) g_no_query = 1;
  if (env_off(L"SKATE3_DIRECT_PRIORIDAD")) g_boost = 0;
  if (env_off(L"SKATE3_DIRECT_TICK")) g_fast_tick = 0;
  logf("corte del logo EA: %s | sin tope de FPS al cargar: %s | Press Start lista al toque: %s | START por dentro: %s | turbo post-START: %s | "
       "perfilador: %s | perfil del mundo: %s | cargadores sin siesta: %s | Press Start sin espera mínima: %s | START al instante: %s | fundido final corto: %s | sin consultas de oclusión al cargar: %s | GPU con prioridad al cargar: %s | tick rápido al cargar: %s",
       g_skip_logo ? "sí" : "no", g_uncap ? "sí" : "no", g_fast_press ? "sí" : "no", g_press_inside ? "sí" : "no", g_fast_fade ? "sí" : "no",
       g_profile ? "sí" : "no", g_profile_world ? "sí" : "no", g_no_nap ? "sí" : "no", g_press_clock ? "sí" : "no", g_press_tap ? "sí" : "no", g_fast_black ? "sí" : "no", g_no_query ? "sí" : "no", g_boost ? "sí" : "no", g_fast_tick ? "sí" : "no");
  // Each hook is located through the exe's own guest->host table, so a copy
  // of the build laid out differently still gets hooked at the right place;
  // the bytes are then checked exactly as before.
  static HookSpec hk[sizeof kHooks / sizeof kHooks[0]];
  int have_map = load_map();
  if (!have_map) logf("no encontré la tabla de funciones: uso las direcciones fijas");
  for (usize i = 0; i < sizeof kHooks / sizeof kHooks[0]; i++) {
    hk[i] = kHooks[i];
    u32 r = have_map && hk[i].guest ? host_of(hk[i].guest) : 0;
    if (r && r != hk[i].rva) logf("%s: %08X está en rva %X (no %X)", hk[i].name, hk[i].guest, r, hk[i].rva);
    if (r) hk[i].rva = r;
    if (!check_bytes(hk[i].rva, hk[i].expect, hk[i].len)) {
      logf("no coincide %s: no toco nada", hk[i].name);
      return 0;
    }
  }
  if (have_map) {
    u32 r = host_of(0x82608EA0);
    if (r) g_setactive = host_fn(r);
  }
  if (!g_setactive) logf("Press Start: no encontré sub_82608EA0, queda como siempre");
  int ok = 1;
  for (usize i = 0; i < sizeof hk / sizeof hk[0]; i++) ok &= hook_entry(&hk[i]);
  if (have_map) g_swap_hooked = hook_swap();
  if (!g_swap_hooked) logf("cuadros: no encontré VdSwap, no los cuento");
  CreateThread(NULL, 0, monitor, NULL, 0, NULL);
  if (g_profile) prof_start();
  else if (g_profile_world) prof_start_focus();
  return ok;
}

#ifdef S3D_TEST
// Test-only: the nap cut on a thread named load_thread and on an unnamed one.
static u32 g_seen_r3;
static void stub_sleep(PPCContext* ctx, u8* base) { (void)base; g_seen_r3 = ctx->r3.u32; }
typedef long(WINAPI* SetThreadDescription_t)(HANDLE h, const wchar* name);
static DWORD WINAPI nap_thr(void* arg) {
  SetThreadDescription_t sd = (SetThreadDescription_t)GetProcAddress(GetModuleHandleW(L"kernel32.dll"), "SetThreadDescription");
  if (arg && sd) sd(GetCurrentThread(), L"load_thread (F8000060)");
  PPCContext c;
  memset(&c, 0, sizeof c);
  c.r3.u64 = 1;
  h_sleep(&c, NULL);
  c.r3.u64 = 1;
  h_sleep(&c, NULL);  // second call: cached kind
  return 0;
}
__declspec(dllexport) int s3d_naptest(void) {
  PPCFunc* keep = o_sleep;
  o_sleep = stub_sleep;
  int r = 0;
  for (int named = 1; named >= 0; named--) {
    g_seen_r3 = 77;
    HANDLE h = CreateThread(NULL, 0, nap_thr, named ? (void*)1 : NULL, 0, NULL);
    WaitForSingleObject(h, 5000);
    r = r * 10 + (int)g_seen_r3;
  }
  o_sleep = keep;
  logf("naptest: %d (cortes load %ld)", r, g_nap_cut[1]);
  return r;  // 01: named -> 0, unnamed -> 1
}

static DWORD WINAPI gpu_thr(void* arg) {
  SetThreadDescription_t sd = (SetThreadDescription_t)GetProcAddress(GetModuleHandleW(L"kernel32.dll"), "SetThreadDescription");
  if (sd) sd(GetCurrentThread(), arg ? L"GPU Commands (F8000018)" : L"otro hilo");
  Sleep(1500);
  return 0;
}
__declspec(dllexport) int s3d_boosttest(void) {
  HANDLE a = CreateThread(NULL, 0, gpu_thr, (void*)1, 0, NULL);
  HANDLE b = CreateThread(NULL, 0, gpu_thr, NULL, 0, NULL);
  Sleep(300);
  int n = boost_threads();
  boost_restore();
  WaitForSingleObject(a, 5000);
  WaitForSingleObject(b, 5000);
  logf("boosttest: %d hilos", n);
  return n;  // only the GPU one
}

static i64 g_seen_due;
static void stub_settimer(PPCContext* ctx, u8* base) { g_seen_due = gl64(base, ctx->r4.u32); }
__declspec(dllexport) int s3d_ticktest(u8* base) {
  u32 obj = 0x830D0000;
  PPCFunc* keep = o_settimer;
  o_settimer = stub_settimer;
  g_base = base;
  gstore32(base, obj + 8, 0x1234);
  gs64(base, obj + 16, -166667);
  gs64(base, obj + 24, 166667);
  gs64(base, obj + 120, 300000);
  gs64(base, obj + 128, -133333);
  gs64(base, obj + 200, -99);
  PPCContext c;
  memset(&c, 0, sizeof c);
  c.r3.u64 = 0x1234;
  c.r4.u64 = obj + 16;
  int keep_phase = g_phase;
  g_tick_on = 1;
  g_phase = PH_PRESS;
  h_settimer(&c, base);
  int a = g_seen_due == TICK_FAST && gl64(base, obj + 16) == -166667;
  g_phase = PH_LOADING;
  h_settimer(&c, base);
  int b = g_seen_due == -166667 && gl64(base, obj + 120) == 166667 && gl64(base, obj + 128) == 0 &&
          gl64(base, obj + 200) == 0 && gl64(base, obj + 24) == 166667;
  c.r3.u64 = 0x9999;  // another timer: untouched
  g_phase = PH_PRESS;
  g_tick_reset = 0;
  h_settimer(&c, base);
  int d = g_seen_due == -166667;
  c.r3.u64 = 0x1234;  // a quick step after START: 4 ms, then back with the ring cleared
  u32 keep_state = g_state;
  g_phase = PH_LOADING;
  g_state = 0x82715390;
  gs64(base, obj + 128, -133333);
  h_settimer(&c, base);
  int e = g_seen_due == TICK_FAST && gl64(base, obj + 16) == -166667;
  g_state = ST_WORLD;
  h_settimer(&c, base);
  e = e && g_seen_due == -166667 && gl64(base, obj + 128) == 0;
  h_settimer(&c, base);
  e = e && g_seen_due == -166667;
  g_state = keep_state;
  g_phase = keep_phase;
  o_settimer = keep;
  logf("ticktest: %d %d %d %d", a, b, d, e);
  return a && b && d && e;
}

// Test-only: round-trip a rexruntime flag through the cvar API.
__declspec(dllexport) int s3d_flagtest(wchar* dir) {
  usize i = 0;
  for (; dir[i]; i++) g_dir[i] = dir[i];
  g_dir[i] = 0;
  log_open();
  HMODULE rt = GetModuleHandleW(L"rexruntime.dll");
  if (!rt) return -1;
  p_setflag = (SetFlag_t)GetProcAddress(
      rt, "?SetFlagByName@cvar@rex@@YA_NV?$basic_string_view@DU?$char_traits@D@std@@@std@@0@Z");
  p_getflag = (GetFlag_t)GetProcAddress(
      rt, "?GetFlagByName@cvar@rex@@YA?AV?$basic_string@DU?$char_traits@D@std@@V?$allocator@D@2@@std@@"
          "V?$basic_string_view@DU?$char_traits@D@std@@@4@@Z");
  if (!p_setflag || !p_getflag) return -2;
  char a[24] = {0}, b[24] = {0}, c[24] = {0}, d[24] = {0};
  int r1 = get_flag("vsync", a, sizeof a);
  int s1 = set_flag("vsync", "true");
  int r2 = get_flag("vsync", b, sizeof b);
  int r3 = get_flag("d3d12_present_frame_limiter_fps", c, sizeof c);
  int s2 = set_flag("d3d12_present_frame_limiter_fps", "0");  // out of range: must fail
  int s3 = set_flag("no_such_flag_x", "1");
  set_flag("vsync", a);
  get_flag("vsync", d, sizeof d);
  resolve_runtime();
  if (p_queue_syn) {
    static const SynStep t[] = {{0x0010, 5, 0, 0}, {0, 5, 0, 0}};
    p_queue_syn(t, 2);
    logf("flagtest: QueueSyntheticInputSequence llamado");
  } else return -3;
  logf("flagtest: vsync %d'%s' set %d -> %d'%s' | fps %d'%s' set0 %d | unknown %d | back '%s'", r1, a, s1, r2, b,
       r3, c, s2, s3, d);
  char q[8] = {0};
  g_no_query = 1;
  query_begin();
  int qoff = g_query_off;
  get_flag("occlusion_query_enable", q, sizeof q);
  query_end();
  char q2[8] = {0};
  get_flag("occlusion_query_enable", q2, sizeof q2);
  logf("flagtest: oclusión apagada %d '%s' -> vuelve '%s'", qoff, q, q2);
  return r1 && s1 && r2 && b[0] == 't' && !s2 && !s3 && d[0] == a[0] && qoff && q[0] == 'f' && q2[0] == 't';
}
#endif
