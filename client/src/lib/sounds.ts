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
  | "loss";

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
};

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

// volumeMultiplier is the master * effects gain pushed by the
// settings panel (S11.5). Stored as a plain number so play() can
// fold it into per-call audio.volume without crossing a Svelte
// store boundary on every fire. Range [0, 1]; clamped on set.
let volumeMultiplier = 1;

// setVolumeMultiplier is called from settings.ts whenever
// audio.masterVolume or audio.effectsVolume changes. Kept exported
// + dependency-free so the module's "no stores" contract holds.
export function setVolumeMultiplier(v: number): void {
  volumeMultiplier = Math.min(1, Math.max(0, v));
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

function pickVariant(name: SoundName): string | null {
  const variants = SOUND_MANIFEST[name];
  if (!variants || variants.length === 0) return null;
  const idx = Math.floor(Math.random() * variants.length);
  return variants[idx] ?? variants[0];
}

function preloadAll(): void {
  if (!hasWindow()) return;
  for (const variants of Object.values(SOUND_MANIFEST)) {
    for (const url of variants) {
      if (pool.has(url)) continue;
      const a = new Audio(url);
      a.preload = "auto";
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
}

// play fires a sound. No-ops if muted, if the name has no variants,
// or if a same-name call landed within RATE_LIMIT_MS.
export function play(name: SoundName, opts?: PlayOptions): void {
  if (!hasWindow()) return;
  if (isMuted()) return;

  const now = performance.now();
  const last = lastPlayedAt.get(name) ?? -Infinity;
  if (now - last < RATE_LIMIT_MS) return;
  lastPlayedAt.set(name, now);

  const url = pickVariant(name);
  if (!url) return;

  // Clone from the preloaded element when available so the decoded
  // buffer is reused; fall back to a fresh Audio if armAudio... hasn't
  // fired yet (e.g. tests, or a sound somehow triggered before the
  // first gesture).
  const base = pool.get(url);
  const audio = base ? (base.cloneNode(true) as HTMLAudioElement) : new Audio(url);

  // Final volume = caller's per-event hint × the user's master×effects
  // multiplier from the settings panel. Both clamped; product stays in
  // [0, 1]. A multiplier of 0 effectively mutes without taking the
  // separate isMuted() path — useful for "0 master volume" UX without
  // stomping the explicit mute checkbox.
  const callerVol = Math.min(1, Math.max(0, opts?.volume ?? 1));
  audio.volume = callerVol * volumeMultiplier;
  if (opts?.rate !== undefined) audio.playbackRate = opts.rate;

  // play() returns a promise that rejects if autoplay is blocked.
  // Swallow — the first-gesture arm should have unblocked us, and
  // a lost sound is not worth surfacing to the user.
  void audio.play().catch(() => {});
}
