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
  | "heal"
  | "attack"
  | "block"
  | "combat_resolve"
  | "turn_change"
  | "shuffle"
  | "win"
  | "loss"
  // ADR 0209 §2 interface cues. Files are Ian's to make; until they land
  // each falls back (SOUND_FALLBACK) or stays silent.
  | "your_turn"
  | "priority"
  | "prompt"
  | "timer_warning";

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
  attack: "combat",
  block: "combat",
  combat_resolve: "combat",
  heal: "effects",
  turn_change: "interface",
  win: "interface",
  loss: "interface",
  your_turn: "interface",
  priority: "interface",
  prompt: "interface",
  timer_warning: "interface",
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
  heal: ["/sounds/heal-1.mp3", "/sounds/heal-2.mp3"],
  attack: ["/sounds/attack-1.mp3", "/sounds/attack-2.mp3"],
  block: ["/sounds/block-1.mp3", "/sounds/block-2.mp3"],
  combat_resolve: ["/sounds/combat_resolve-1.mp3", "/sounds/combat_resolve-2.mp3"],
  turn_change: ["/sounds/turn_change-1.mp3", "/sounds/turn_change-2.mp3"],
  shuffle: ["/sounds/shuffle-1.mp3", "/sounds/shuffle-2.mp3"],
  win: ["/sounds/win-1.mp3", "/sounds/win-2.mp3"],
  loss: ["/sounds/loss-1.mp3", "/sounds/loss-2.mp3"],
  your_turn: ["/sounds/your_turn-1.mp3", "/sounds/your_turn-2.mp3"],
  priority: ["/sounds/priority-1.mp3", "/sounds/priority-2.mp3"],
  prompt: ["/sounds/prompt-1.mp3", "/sounds/prompt-2.mp3"],
  timer_warning: ["/sounds/timer_warning-1.mp3", "/sounds/timer_warning-2.mp3"],
};

// What a sound plays when none of its own files exist. Only the cues that
// replace an old one name a fallback: your_turn is turn_change until its
// files land. The others stay silent, because a placeholder ping on every
// priority would be noise.
export const SOUND_FALLBACK: Partial<Record<SoundName, SoundName>> = {
  your_turn: "turn_change",
};

// A ducking interface cue lowers every other bus by this much: 10^(-6/20),
// about 6 dB.
export const DUCK_FACTOR = 0.5012;
// A ducked table is released after this long even if the cue's `ended`
// never fires (a stalled load, a tab suspended mid-sound). Generous: an
// interface cue is at most 1.5 s (ADR 0209 §6).
const DUCK_MAX_MS = 2000;
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

function preloadAll(): void {
  if (!hasWindow()) return;
  for (const variants of Object.values(SOUND_MANIFEST)) {
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

function applyDuck(): void {
  const f = duckFactor();
  for (const p of playing) {
    p.audio.volume = clamp01(p.category === "interface" ? p.base : p.base * f);
  }
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
  const entry: Playing = { audio, base: callerVol * soundGain(name), category };
  audio.volume = clamp01(category === "interface" ? entry.base : entry.base * duckFactor());
  if (opts?.rate !== undefined) audio.playbackRate = opts.rate;

  playing.add(entry);
  const done = () => playing.delete(entry);
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
