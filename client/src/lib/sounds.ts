// S09 sound pass — audio playback for the table.
//
// Single registry of logical sound name → list of variant URLs.
// Most events ship two takes; play() picks one at random so the
// same event fired back-to-back (e.g. two cards drawn) doesn't
// read as an obvious loop. Files live under client/public/sounds
// so Vite serves them at /sounds/<file>.mp3 with no extra config.
//
// Autoplay policy: browsers block HTMLAudioElement.play() until
// the page has received a user gesture. armAudioOnFirstGesture()
// installs a one-shot pointerdown/keydown listener that pre-decodes
// every variant so the first in-game trigger has zero latency.
// Call it once on app mount — subsequent calls no-op.
//
// Mute: persisted to localStorage["cmdctrl.muted"]. Reduced-motion
// does NOT auto-mute — audio cues are useful even for players who
// suppress animations — but the settings panel (S11.5) can wire an
// explicit toggle on top of setMuted().
//
// Wiring to game events is deliberately not in this module; call
// sites import { play } and fire it from the existing GSAP / state
// seams described in docs/s09-sound-pass.md.
//
// Categories (ADR 0209 §2): every SoundName belongs to one SoundCategory
// (interface, cards, combat, effects, voice), and play() multiplies the
// master volume, the category's volume and the caller's hint. Music has
// its own bus in music.ts. A sound in the interface category ducks the
// others by about 6 dB while it plays; one logical sound plays at most
// once per event Batch (PlayOptions.batch).
//
// Missing files (ADR 0209 §6): a manifest entry may name a file that does
// not exist yet. The first load error marks that URL missing; play() then
// uses the entry's other takes, then SOUND_FALLBACK, then stays silent.
// A missing file is never an error.
//
// Deliberately kept dependency-free (no GSAP, no stores) so it's
// safe to import from any surface, including SSR / tests (it
// short-circuits when `window` is undefined).

export type SoundName =
  | "draw"
  | "play"
  | "tap"
  | "untap_all"
  | "damage"
  // #4432, #4543: combat damage hitting a player, by how much lands: light
  // (1), medium (2-3) and heavy (4+). Each falls back to damage.
  | "face_hit_light"
  | "face_hit_medium"
  | "face_hit_heavy"
  | "heal"
  | "attack"
  | "block"
  | "combat_resolve"
  | "turn_change"
  | "shuffle"
  | "win"
  | "loss"
  // #4449: a player leaves the game (the rock-shatter and the tombstone).
  | "player_eliminated"
  // ADR 0209 §2 interface cues. If its files fail to load, your_turn falls
  // back (SOUND_FALLBACK) and the others stay silent.
  | "your_turn"
  | "priority"
  | "prompt"
  | "timer_warning"
  // ADR 0209 §3 generic effect cues (#4396), all in the effects category.
  | "destroy"
  | "exile"
  | "sacrifice"
  | "sacrifice_knife"
  | "bounce"
  | "counter"
  | "board_wipe"
  | "pw_enter"
  | "pw_leave"
  // ADR 0209 §4 signature effects (#4397), in the effects category. If its
  // files fail to load, each falls back to the generic cue it replaces
  // (SOUND_FALLBACK). Loaded lazily: the first time the
  // card appears in a game (fx/signatures.ts), never by preloadAll.
  | "sig_blasphemous_act"
  | "sig_toxic_deluge"
  | "sig_bitter_triumph"
  | "sig_nicol_bolas"
  | "sig_dragons_approach"
  // #4544: an extort drain resolving (a soul sucked out, a coin on the scale).
  // Plays damage until its file lands (SOUND_FALLBACK, ADR 0209 §6).
  | "extort";

// The volume buses a sound can belong to. The music bus is music.ts's.
export type SoundCategory = "interface" | "cards" | "combat" | "effects" | "voice";

export const SOUND_CATEGORIES: readonly SoundCategory[] = [
  "interface",
  "cards",
  "combat",
  "effects",
  "voice",
];

// Which bus each sound plays on. Compile-checked: a new SoundName cannot
// be added without naming its category.
export const SOUND_CATEGORY: Record<SoundName, SoundCategory> = {
  draw: "cards",
  play: "cards",
  tap: "cards",
  untap_all: "cards",
  shuffle: "cards",
  damage: "combat",
  face_hit_light: "combat",
  face_hit_medium: "combat",
  face_hit_heavy: "combat",
  attack: "combat",
  block: "combat",
  combat_resolve: "combat",
  heal: "effects",
  player_eliminated: "effects",
  turn_change: "interface",
  win: "interface",
  loss: "interface",
  your_turn: "interface",
  priority: "interface",
  prompt: "interface",
  timer_warning: "interface",
  destroy: "effects",
  exile: "effects",
  sacrifice: "effects",
  sacrifice_knife: "effects",
  bounce: "effects",
  counter: "effects",
  board_wipe: "effects",
  pw_enter: "effects",
  pw_leave: "effects",
  sig_blasphemous_act: "effects",
  sig_toxic_deluge: "effects",
  sig_bitter_triumph: "effects",
  sig_nicol_bolas: "effects",
  sig_dragons_approach: "effects",
  extort: "effects",
};

// Each event has two takes. pickVariant() chooses uniformly at
// random; if a sprint adds a third take, just append — no other
// code changes needed.
export const SOUND_MANIFEST: Record<SoundName, string[]> = {
  draw: ["/sounds/draw-1.mp3", "/sounds/draw-2.mp3"],
  play: ["/sounds/play-1.mp3", "/sounds/play-2.mp3"],
  tap: ["/sounds/tap-1.mp3", "/sounds/tap-2.mp3"],
  untap_all: ["/sounds/untap_all-1.mp3", "/sounds/untap_all-2.mp3"],
  damage: ["/sounds/damage-1.mp3", "/sounds/damage-2.mp3"],
  // #4543: one take per damage tier.
  face_hit_light: ["/sounds/face_hit_light-1.mp3"],
  face_hit_medium: ["/sounds/face_hit_medium-1.mp3"],
  face_hit_heavy: ["/sounds/face_hit_heavy-1.mp3"],
  // #4449: exactly one take.
  player_eliminated: ["/sounds/player_eliminated-1.mp3"],
  heal: ["/sounds/heal-1.mp3", "/sounds/heal-2.mp3"],
  attack: ["/sounds/attack-1.mp3", "/sounds/attack-2.mp3"],
  block: ["/sounds/block-1.mp3", "/sounds/block-2.mp3"],
  combat_resolve: ["/sounds/combat_resolve-1.mp3", "/sounds/combat_resolve-2.mp3"],
  turn_change: ["/sounds/turn_change-1.mp3", "/sounds/turn_change-2.mp3"],
  shuffle: ["/sounds/shuffle-1.mp3", "/sounds/shuffle-2.mp3"],
  win: ["/sounds/win-1.mp3", "/sounds/win-2.mp3"],
  loss: ["/sounds/loss-1.mp3", "/sounds/loss-2.mp3"],
  your_turn: ["/sounds/your_turn-1.mp3"],
  priority: ["/sounds/priority-1.mp3"],
  prompt: ["/sounds/prompt-1.mp3", "/sounds/prompt-2.mp3"],
  timer_warning: ["/sounds/timer_warning-1.mp3", "/sounds/timer_warning-2.mp3"],
  destroy: ["/sounds/destroy-1.mp3", "/sounds/destroy-2.mp3"],
  exile: ["/sounds/exile-1.mp3"],
  sacrifice: ["/sounds/sacrifice-1.mp3", "/sounds/sacrifice-2.mp3"],
  sacrifice_knife: ["/sounds/sacrifice_knife-1.mp3", "/sounds/sacrifice_knife-2.mp3"],
  bounce: ["/sounds/bounce-1.mp3", "/sounds/bounce-2.mp3"],
  counter: ["/sounds/counter-1.mp3", "/sounds/counter-2.mp3"],
  board_wipe: ["/sounds/board_wipe-1.mp3", "/sounds/board_wipe-2.mp3"],
  pw_enter: ["/sounds/pw_enter-1.mp3", "/sounds/pw_enter-2.mp3"],
  pw_leave: ["/sounds/pw_leave-1.mp3"],
  sig_blasphemous_act: ["/sounds/sig_blasphemous_act-1.mp3", "/sounds/sig_blasphemous_act-2.mp3"],
  sig_toxic_deluge: ["/sounds/sig_toxic_deluge-1.mp3"],
  sig_bitter_triumph: ["/sounds/sig_bitter_triumph-1.mp3", "/sounds/sig_bitter_triumph-2.mp3"],
  sig_nicol_bolas: ["/sounds/sig_nicol_bolas-1.mp3", "/sounds/sig_nicol_bolas-2.mp3"],
  sig_dragons_approach: [
    "/sounds/sig_dragons_approach-1.mp3",
    "/sounds/sig_dragons_approach-2.mp3",
  ],
  // #4544: file not made yet (ADR 0209 §6). No takes are listed, so none is
  // fetched and the cue plays damage through SOUND_FALLBACK. The PR that adds
  // extort-1.mp3 lists it here (client/public/sounds/README.md, "Files not
  // made yet").
  extort: [],
};

// What a sound plays when none of its own files exist. Only the cues that
// replace an old one name a fallback: your_turn plays turn_change if its
// file fails to load, and each face_hit tier plays damage. The others stay silent, because a placeholder ping
// on every priority would be noise.
// A signature (#4397) falls back to the generic cue it replaces, so a card
// with a signature never sounds quieter than one without.
export const SOUND_FALLBACK: Partial<Record<SoundName, SoundName>> = {
  your_turn: "turn_change",
  face_hit_light: "damage",
  face_hit_medium: "damage",
  face_hit_heavy: "damage",
  sig_blasphemous_act: "board_wipe",
  sig_toxic_deluge: "board_wipe",
  sig_bitter_triumph: "destroy",
  sig_nicol_bolas: "pw_enter",
  sig_dragons_approach: "board_wipe",
  extort: "damage",
};

// Sounds preloadAll leaves alone: a signature's files are fetched the
// first time its card appears in a game (preloadSound), not at start.
export function isLazySound(name: SoundName): boolean {
  return name.startsWith("sig_");
}

// A ducking interface cue lowers every other bus by this much: 10^(-6/20),
// about 6 dB.
export const DUCK_FACTOR = 0.5012;
// A ducked table is released after this long even if the cue's `ended`
// never fires (a stalled load, a tab suspended mid-sound). Generous: an
// interface cue is at most 1.5 s (ADR 0209 §6).
const DUCK_MAX_MS = 2000;
// A sound longer than it should be is cut at the length named here (#4567):
// the war cry's takes run 5 and 7 s and would bury whatever follows combat.
// The volume ramps to 0 over the last SOUND_FADE_MS, then the take is
// paused and released. A take shorter than its cap plays untouched.
export const SOUND_MAX_MS: Partial<Record<SoundName, number>> = {
  combat_resolve: 2000,
};
export const SOUND_FADE_MS = 300;
// The ramp moves in steps this far apart.
const FADE_STEP_MS = 25;
// How many event batches the burst cap remembers.
const BURST_MEMORY = 64;

const MUTE_STORAGE_KEY = "cmdctrl.muted";

// Drop repeat plays of the same name fired within this window —
// prevents "opening hand of 7" from stacking seven draw sounds on
// top of each other. Empirically 40ms is tight enough that true
// user-driven double-taps still feel responsive while machine-gun
// snapshot rebuilds get deduped.
const RATE_LIMIT_MS = 40;

// Pool of preloaded, decoded Audio elements keyed by URL. Populated
// by armAudioOnFirstGesture(); play() clones from here so overlapping
// plays don't cut each other off (HTMLAudioElement can only play one
// position at a time per element).
const pool = new Map<string, HTMLAudioElement>();
const lastPlayedAt = new Map<SoundName, number>();
let armed = false;
let mutedCache: boolean | null = null;

// Master and per-category gains pushed by the settings bridge
// (settings.ts). Plain numbers so play() can fold them into per-call
// audio.volume without crossing a Svelte store boundary on every fire.
// Each is in [0, 1] and clamped on set.
let masterGain = 1;
const categoryGain: Record<SoundCategory, number> = {
  interface: 1,
  cards: 1,
  combat: 1,
  effects: 1,
  voice: 1,
};

function clamp01(v: number): number {
  return Math.min(1, Math.max(0, Number.isFinite(v) ? v : 0));
}

// setMasterVolume is called from settings.ts whenever audio.masterVolume
// changes. Kept exported + dependency-free so the module's "no stores"
// contract holds.
export function setMasterVolume(v: number): void {
  masterGain = clamp01(v);
}

// setCategoryVolume sets one bus's gain, from audio.<category>Volume.
export function setCategoryVolume(category: SoundCategory, v: number): void {
  categoryGain[category] = clamp01(v);
}

// busGain is master x one bus's volume, for a caller that plays its own
// files on a bus (fx/voices.ts, ADR 0209 §5).
export function busGain(category: SoundCategory): number {
  return masterGain * categoryGain[category];
}

// soundGain is the product play() applies before the caller's hint:
// master x the sound's category.
export function soundGain(name: SoundName): number {
  return masterGain * categoryGain[SOUND_CATEGORY[name]];
}

function hasWindow(): boolean {
  return typeof window !== "undefined";
}

export function isMuted(): boolean {
  if (mutedCache !== null) return mutedCache;
  if (!hasWindow()) return false;
  try {
    mutedCache = window.localStorage.getItem(MUTE_STORAGE_KEY) === "1";
  } catch {
    mutedCache = false;
  }
  return mutedCache;
}

export function setMuted(muted: boolean): void {
  mutedCache = muted;
  if (!hasWindow()) return;
  try {
    if (muted) {
      window.localStorage.setItem(MUTE_STORAGE_KEY, "1");
    } else {
      window.localStorage.removeItem(MUTE_STORAGE_KEY);
    }
  } catch {
    // swallow — private-mode / disabled storage still lets in-memory
    // mute work for the current session.
  }
}

export function toggleMuted(): boolean {
  const next = !isMuted();
  setMuted(next);
  return next;
}

// URLs a load failed for (404, a SPA fallback page that is not audio, an
// undecodable file). Never retried within a session.
const missing = new Set<string>();

function markMissing(url: string): void {
  missing.add(url);
}

// isMissing reports a URL a load has failed for. Exported for tests.
export function isMissing(url: string): boolean {
  return missing.has(url);
}

// pickVariant picks a take for `name` that has not failed to load. When
// none of its own takes are left it follows SOUND_FALLBACK, once per
// name, so a cycle cannot loop. Null means stay silent.
function pickVariant(name: SoundName, seen: SoundName[] = []): string | null {
  const variants = (SOUND_MANIFEST[name] ?? []).filter((u) => !missing.has(u));
  if (variants.length > 0) {
    const idx = Math.floor(Math.random() * variants.length);
    return variants[idx] ?? variants[0];
  }
  const next = SOUND_FALLBACK[name];
  if (!next || seen.includes(next) || next === name) return null;
  return pickVariant(next, [...seen, name]);
}

function preloadURLs(variants: readonly string[]): void {
  for (const url of variants) {
    if (pool.has(url)) continue;
    const a = new Audio(url);
    a.preload = "auto";
    // A file that is not there yet (ADR 0209 §6) fails to load; note
    // it so play() skips it instead of trying on every fire.
    a.addEventListener?.("error", () => markMissing(url));
    // Touching .load() nudges the browser to actually start
    // fetching + decoding immediately instead of lazily on first
    // play.
    a.load();
    pool.set(url, a);
  }
}

function preloadAll(): void {
  if (!hasWindow()) return;
  for (const [name, variants] of Object.entries(SOUND_MANIFEST)) {
    if (isLazySound(name as SoundName)) continue;
    preloadURLs(variants);
  }
}

// preloadSound fetches one sound's takes ahead of its first play: a
// signature's, the first time its card appears in a game (#4397).
export function preloadSound(name: SoundName): void {
  if (!hasWindow()) return;
  preloadURLs(SOUND_MANIFEST[name] ?? []);
}

// armAudioOnFirstGesture attaches a one-shot pointerdown + keydown
// listener to window. On the first of either, it preloads every
// variant. Safe to call many times — only the first call installs
// listeners, and they auto-remove after firing.
export function armAudioOnFirstGesture(): void {
  if (armed || !hasWindow()) return;
  armed = true;
  const unlock = () => {
    preloadAll();
    window.removeEventListener("pointerdown", unlock);
    window.removeEventListener("keydown", unlock);
  };
  window.addEventListener("pointerdown", unlock, { once: true });
  window.addEventListener("keydown", unlock, { once: true });
}

export interface PlayOptions {
  // 0..1, clamped. Defaults to 1.
  volume?: number;
  // Playback rate — 1 = normal, 2 = double speed, 0.5 = half.
  // Useful for slight pitch variation on repeat plays; defaults to 1.
  rate?: number;
  // The event batch this sound announces (ADR 0209 §2). A logical sound
  // plays at most once per batch, so twelve creatures dying make one
  // board-wipe sound. Callers scope the key to the game (for example
  // `${gameID}:${batch}`), because batch numbers restart with each game.
  batch?: string | number;
}

// ---- Burst cap -----------------------------------------------------------

// batch key -> the names already played for it. Insertion-ordered, so the
// oldest batch is the first key; capped at BURST_MEMORY.
const batches = new Map<string, Set<SoundName>>();

// claimBatch returns true the first time `name` is played for `batch`,
// and false for every repeat.
function claimBatch(name: SoundName, batch: string | number): boolean {
  const key = String(batch);
  let names = batches.get(key);
  if (!names) {
    names = new Set();
    batches.set(key, names);
    if (batches.size > BURST_MEMORY) {
      const oldest = batches.keys().next().value;
      if (oldest !== undefined) batches.delete(oldest);
    }
  }
  if (names.has(name)) return false;
  names.add(name);
  return true;
}

// resetBurstCap forgets every batch; call it when a game starts or ends.
export function resetBurstCap(): void {
  batches.clear();
}

// ---- Ducking -------------------------------------------------------------

interface Playing {
  audio: HTMLAudioElement;
  // Caller hint x master x category: the volume before any ducking.
  base: number;
  category: SoundCategory;
  // 1 until the length cap starts fading the take out, then falls to 0.
  fade: number;
}

const playing = new Set<Playing>();
let duckCount = 0;
let duckListener: ((factor: number) => void) | null = null;

// onDuck registers the one listener that hears the duck factor change
// (1 when released, DUCK_FACTOR while an interface cue plays). The
// settings bridge points it at the music bus, which this module does not
// import.
export function onDuck(fn: ((factor: number) => void) | null): void {
  duckListener = fn;
  fn?.(duckFactor());
}

function duckFactor(): number {
  return duckCount > 0 ? DUCK_FACTOR : 1;
}

function setVolume(p: Playing, duck: number): void {
  p.audio.volume = clamp01((p.category === "interface" ? p.base : p.base * duck) * p.fade);
}

function applyDuck(): void {
  const f = duckFactor();
  for (const p of playing) setVolume(p, f);
  duckListener?.(f);
}

// duckFor ducks the other buses until `audio` ends, errors, or DUCK_MAX_MS
// passes, whichever is first.
function duckFor(audio: HTMLAudioElement): () => void {
  duckCount++;
  let released = false;
  const release = () => {
    if (released) return;
    released = true;
    duckCount = Math.max(0, duckCount - 1);
    applyDuck();
  };
  audio.addEventListener?.("ended", release);
  audio.addEventListener?.("error", release);
  setTimeout(release, DUCK_MAX_MS);
  applyDuck();
  return release;
}

// ---- play ----------------------------------------------------------------

// play fires a sound. No-ops if muted, if the name has no playable
// variants (and no fallback), if a same-name call landed within
// RATE_LIMIT_MS, or if the same name already played for `opts.batch`.
export function play(name: SoundName, opts?: PlayOptions): void {
  if (!hasWindow()) return;
  if (isMuted()) return;

  if (opts?.batch !== undefined && !claimBatch(name, opts.batch)) return;

  const now = performance.now();
  const last = lastPlayedAt.get(name) ?? -Infinity;
  if (now - last < RATE_LIMIT_MS) return;
  lastPlayedAt.set(name, now);

  fire(name, opts);
}

// capLength schedules the length cap for one take: a fade-out starting
// SOUND_FADE_MS before the maximum, then a pause. It returns the function
// that clears its timers, which the take's end, error or rejection calls.
function capLength(name: SoundName, entry: Playing): () => void {
  const max = SOUND_MAX_MS[name];
  if (max === undefined) return () => {};
  const fadeMs = Math.min(SOUND_FADE_MS, max);
  let ramp: ReturnType<typeof setInterval> | undefined;
  const start = setTimeout(() => {
    const steps = Math.max(1, Math.round(fadeMs / FADE_STEP_MS));
    let step = 0;
    ramp = setInterval(() => {
      step++;
      entry.fade = Math.max(0, 1 - step / steps);
      setVolume(entry, duckFactor());
      if (step >= steps) clearInterval(ramp);
    }, FADE_STEP_MS);
  }, max - fadeMs);
  const cut = setTimeout(() => {
    clearInterval(ramp);
    entry.fade = 0;
    setVolume(entry, duckFactor());
    entry.audio.pause?.();
    playing.delete(entry);
  }, max);
  return () => {
    clearTimeout(start);
    clearTimeout(cut);
    clearInterval(ramp);
  };
}

function fire(name: SoundName, opts: PlayOptions | undefined): void {
  const url = pickVariant(name);
  if (!url) return;

  // Clone from the preloaded element when available so the decoded
  // buffer is reused; fall back to a fresh Audio if armAudio... hasn't
  // fired yet (e.g. tests, or a sound somehow triggered before the
  // first gesture).
  const base = pool.get(url);
  const audio = base ? (base.cloneNode(true) as HTMLAudioElement) : new Audio(url);

  // Final volume = caller's per-event hint x master x the sound's
  // category volume from the settings panel, then x the duck factor
  // while an interface cue plays. All clamped, so the product stays in
  // [0, 1]. A product of 0 effectively mutes without taking the separate
  // isMuted() path: "0 master volume" UX without stomping the explicit
  // mute checkbox.
  const category = SOUND_CATEGORY[name];
  const callerVol = clamp01(opts?.volume ?? 1);
  const entry: Playing = { audio, base: callerVol * soundGain(name), category, fade: 1 };
  setVolume(entry, duckFactor());
  if (opts?.rate !== undefined) audio.playbackRate = opts.rate;

  playing.add(entry);
  const stopCap = capLength(name, entry);
  const done = () => {
    playing.delete(entry);
    stopCap();
  };
  audio.addEventListener?.("ended", done);
  audio.addEventListener?.("error", () => {
    markMissing(url);
    done();
  });
  const releaseDuck = category === "interface" ? duckFor(audio) : null;

  // play() returns a promise that rejects if autoplay is blocked, or
  // with NotSupportedError when the file is not there. Swallow either:
  // the first-gesture arm should have unblocked autoplay, and a lost
  // sound is not worth surfacing to the user. A missing file is noted and
  // the sound tried again with what is left (another take, then the
  // fallback), so that very fire still sounds.
  void audio.play().catch((err: unknown) => {
    done();
    releaseDuck?.();
    if (err instanceof Error && err.name === "NotSupportedError") {
      markMissing(url);
      // Each retry marks one more URL missing, so this ends.
      fire(name, opts);
    }
  });
}
