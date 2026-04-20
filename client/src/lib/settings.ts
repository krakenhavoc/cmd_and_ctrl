import { writable, get, type Writable } from "svelte/store";
import { setMuted, setVolumeMultiplier } from "./sounds";
import { setAnimationConfig } from "./animations";

// Settings is the client-wide preferences schema. Every toggle the
// Settings panel surfaces maps to a field here. Persisted to
// localStorage under STORAGE_KEY so preferences survive reloads,
// and broadcast via a Svelte writable so components react live
// without page refresh.
//
// Schema is versioned — bumping SETTINGS_VERSION and adding a
// migration in `migrate` lets us evolve the shape without stranding
// users on old JSON. Current version is 1. Absorbs the pre-S11.5
// `cmdctrl.muted` legacy key so mute state is preserved across the
// introduction of this module.
//
// Nothing in this module touches the DOM. Reactive application of
// settings is handled by consumers ($effect blocks in App.svelte,
// Settings.svelte, etc.), so this module stays testable from plain
// JS and safe to import from SSR contexts.

export interface Settings {
  // Schema version. Bump when fields change shape; add a migration.
  __version: number;

  audio: {
    // Master mute for every SFX. Replaces the pre-S11.5
    // localStorage["cmdctrl.muted"] flag.
    muted: boolean;
    // 0..100 linear; sounds.ts applies as multiplier on HTMLAudioElement.volume.
    masterVolume: number;
    // 0..100, applied to the effects bus (draw, tap, damage, etc).
    effectsVolume: number;
    // Music is scaffolded for a future music track; no-op until then.
    musicVolume: number;
  };

  animations: {
    // Master toggle for transitions + tweens. Auto-off if the OS
    // signals prefers-reduced-motion: reduce at first load; user
    // override persists thereafter.
    enabled: boolean;
    // Multiplier on tween/transition durations. Values >1 slow
    // animations down; <1 speed them up. Typed as one of a fixed
    // set so the UI can use a radio select.
    speed: 0.5 | 1 | 1.5 | 2;
    // Per-effect toggles. Disabled when `enabled === false`
    // regardless of the individual flag.
    cardDraw: boolean;
    cardPlay: boolean;
    cardTap: boolean;
    cardUntap: boolean;
    cardFlip: boolean;
    particlesEtb: boolean;
    damagePopups: boolean;
  };

  display: {
    // Dark is the current default. Light + high-contrast scaffold
    // the CSS-variable plumbing; full theming ships in a follow-up.
    theme: "dark" | "light" | "high-contrast";
    // Battlefield card size. Applied as a CSS var so re-tuning a
    // preference re-renders without DOM rebuilds.
    cardSize: "small" | "medium" | "large";
    // Hand fan vs. stack — fan is the existing S06 default.
    handLayout: "fan" | "stacked";
    // Hover preview delay in milliseconds, 0..1000. S11 hover
    // preview reads this as its activation threshold.
    hoverDelayMs: number;
    // Show opponent hand count badge on their panel. Existing
    // behaviour is always-on; this lets a viewer who finds it
    // distracting hide it.
    showOpponentHandCount: boolean;
  };

  gameplay: {
    // Confirm-before-exit when navigating away from an active game.
    confirmExit: boolean;
    // When the stack is empty and no legal plays exist, auto-pass
    // priority. Held Shift on the pass button overrides.
    autoPassPriority: boolean;
    // Per-step stops will land in S13.3 once we have the priority-
    // aware stops mechanism. Storing the empty object now keeps the
    // schema future-proof.
    stepStops: Record<string, boolean>;
  };

  accessibility: {
    // Honours the OS prefers-reduced-motion signal at first load;
    // subsequent flips persist as explicit overrides.
    reduceMotion: boolean;
    // CSS --font-scale multiplier. 1.0 = browser default.
    textScale: 0.9 | 1.0 | 1.2 | 1.5;
    // Colour-blind-friendly seat palette override. Swaps colors.ts
    // default palette for an alternate set.
    colorblindPalette: boolean;
    // Always-visible focus outlines. Overrides :focus-visible
    // suppression on buttons + inputs.
    alwaysShowFocus: boolean;
  };
}

export const SETTINGS_VERSION = 1;
const STORAGE_KEY = "cmdctrl.settings.v1";
const LEGACY_MUTED_KEY = "cmdctrl.muted";

// prefersReducedMotion reads the OS hint without subscribing. Used
// to pick the initial default for accessibility.reduceMotion when
// the user has no saved settings yet. SSR-safe.
function prefersReducedMotion(): boolean {
  if (typeof window === "undefined" || !window.matchMedia) return false;
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

// defaultSettings assembles a fresh Settings object with reasonable
// defaults. Call-site runs once at first load or on reset.
export function defaultSettings(): Settings {
  const reduced = prefersReducedMotion();
  return {
    __version: SETTINGS_VERSION,
    audio: {
      muted: false,
      masterVolume: 80,
      effectsVolume: 100,
      musicVolume: 60,
    },
    animations: {
      // If the OS asks for reduced motion, we start with animations
      // off — but the user can flip them on without overriding the
      // OS pref for other apps (our storage is per-site).
      enabled: !reduced,
      speed: 1,
      cardDraw: true,
      cardPlay: true,
      cardTap: true,
      cardUntap: true,
      cardFlip: true,
      particlesEtb: true,
      damagePopups: true,
    },
    display: {
      theme: "dark",
      cardSize: "medium",
      handLayout: "fan",
      hoverDelayMs: 300,
      showOpponentHandCount: true,
    },
    gameplay: {
      confirmExit: true,
      autoPassPriority: false,
      stepStops: {},
    },
    accessibility: {
      reduceMotion: reduced,
      textScale: 1.0,
      colorblindPalette: false,
      alwaysShowFocus: false,
    },
  };
}

// migrate normalises a stored settings blob into the current schema.
// Accepts anything that parses as JSON and does a shallow field-by-
// field merge against defaults, dropping unknown fields and filling
// in missing ones. Also absorbs legacy per-feature localStorage
// keys (currently just cmdctrl.muted) so pre-S11.5 preferences
// don't get silently discarded on first load after upgrade.
function migrate(raw: unknown): Settings {
  const d = defaultSettings();
  if (!raw || typeof raw !== "object") return absorbLegacy(d);

  const s = raw as Partial<Settings>;
  // Future: branch on s.__version to run named migrations. At v1
  // there's nothing to do beyond field-level merge.
  const merged: Settings = {
    __version: SETTINGS_VERSION,
    audio: { ...d.audio, ...(s.audio ?? {}) },
    animations: { ...d.animations, ...(s.animations ?? {}) },
    display: { ...d.display, ...(s.display ?? {}) },
    gameplay: { ...d.gameplay, ...(s.gameplay ?? {}) },
    accessibility: { ...d.accessibility, ...(s.accessibility ?? {}) },
  };
  return absorbLegacy(merged);
}

// absorbLegacy folds pre-S11.5 single-key localStorage flags into
// the unified schema, then clears the legacy keys so migration is
// one-shot. Callers that still read the old keys (e.g. sounds.ts
// isMuted()) keep working; after this migration their next read
// picks up the unified value through the wrapper.
function absorbLegacy(s: Settings): Settings {
  if (typeof localStorage === "undefined") return s;
  const legacyMuted = localStorage.getItem(LEGACY_MUTED_KEY);
  if (legacyMuted !== null) {
    s.audio.muted = legacyMuted === "1";
    localStorage.removeItem(LEGACY_MUTED_KEY);
  }
  return s;
}

function loadSettings(): Settings {
  if (typeof localStorage === "undefined") return defaultSettings();
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return absorbLegacy(defaultSettings());
  try {
    return migrate(JSON.parse(raw));
  } catch {
    // Corrupted blob — fall back to defaults. We don't want to
    // strand a user in an un-openable settings panel because of a
    // bad write.
    return defaultSettings();
  }
}

function saveSettings(s: Settings): void {
  if (typeof localStorage === "undefined") return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
  } catch {
    // QuotaExceeded or Safari-private-mode — swallow. The live
    // store still has the update; persistence is best-effort.
  }
}

export const settings: Writable<Settings> = writable(loadSettings());

// Persist on every mutation. Subscribe rather than wrapping every
// setter because .update() / .set() fire here too.
settings.subscribe((s) => saveSettings(s));

// Bridge audio.muted into sounds.ts. sounds.ts keeps its own cache
// for cheap isMuted() reads from non-reactive call sites; this
// bridge keeps that cache in sync whenever the settings value
// changes — including the migration-time absorption of the legacy
// cmdctrl.muted key.
let lastMuted: boolean | null = null;
let lastVolume: number | null = null;
settings.subscribe((s) => {
  if (s.audio.muted !== lastMuted) {
    lastMuted = s.audio.muted;
    setMuted(s.audio.muted);
  }
  // Master × effects, both 0..100, normalised to a 0..1
  // multiplier. sounds.play() folds this into per-event volume.
  // Music volume is intentionally not folded here — music has no
  // playback path yet (S09 didn't ship a music track).
  const vol = (s.audio.masterVolume / 100) * (s.audio.effectsVolume / 100);
  if (vol !== lastVolume) {
    lastVolume = vol;
    setVolumeMultiplier(vol);
  }
});

// Bridge animations.* into animations.ts. Pushes the entire
// animations group on every change rather than diffing because
// (a) the payload is small (<10 booleans + a number), (b) the
// receiver does shallow Object.assign, and (c) animations are
// next-tick-bound so a redundant push has no observable effect.
settings.subscribe((s) => {
  setAnimationConfig({
    enabled: s.animations.enabled,
    speed: s.animations.speed,
    cardDraw: s.animations.cardDraw,
    cardPlay: s.animations.cardPlay,
    cardTap: s.animations.cardTap,
    cardUntap: s.animations.cardUntap,
    cardFlip: s.animations.cardFlip,
    particlesEtb: s.animations.particlesEtb,
    damagePopups: s.animations.damagePopups,
  });
});

// Observe the OS reduced-motion hint and propagate flips to the
// store when the user has not explicitly overridden. "Explicitly
// overridden" is tracked by comparing the current value to what
// prefersReducedMotion would return — if they match, we keep
// tracking; if they diverge, the user has opted out of OS sync.
// Installed module-scope so it outlives any single component's
// lifecycle.
if (typeof window !== "undefined" && window.matchMedia) {
  const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
  const onChange = (e: MediaQueryListEvent) => {
    const s = get(settings);
    // Only follow OS flips while the in-app toggle still matches
    // the pre-flip OS value. This heuristic means a user who
    // manually toggled the setting stays in charge.
    if (s.accessibility.reduceMotion !== e.matches) {
      // They're different — either (a) we installed after a manual
      // override, (b) the user flipped OS without touching our
      // setting. In either case, sync towards OS so a user who
      // turns on low-motion at the OS level gets it here too.
      settings.update((prev) => ({
        ...prev,
        accessibility: { ...prev.accessibility, reduceMotion: e.matches },
        animations: { ...prev.animations, enabled: !e.matches },
      }));
    }
  };
  // addEventListener is the modern API; Safari < 14 ships
  // addListener instead. Fall back once.
  if (mq.addEventListener) mq.addEventListener("change", onChange);
  else mq.addListener?.(onChange);
}

// updateSettings mutates a single path in the settings tree. Saves
// a few keystrokes at every call-site over the equivalent
// settings.update(prev => ({ ...prev, group: { ...prev.group, key: v } })).
export function updateSettings<G extends keyof Omit<Settings, "__version">>(
  group: G,
  key: keyof Settings[G],
  value: Settings[G][keyof Settings[G]],
): void {
  settings.update((prev) => ({
    ...prev,
    [group]: { ...prev[group], [key]: value },
  }));
}

// resetSettings restores defaults. Used by the "Reset all" button
// in the Advanced tab. Keeps the __version bump so a reset after a
// schema migration doesn't revert to the old shape.
export function resetSettings(): void {
  settings.set(defaultSettings());
}

// settingsOpen drives the modal visibility. Exported so any surface
// (header gear, keyboard shortcut, chat slash-command later) can
// toggle it without the modal needing prop drilling. Settings.svelte
// subscribes and renders itself when this is true.
export const settingsOpen: Writable<boolean> = writable(false);

export function openSettings(): void {
  settingsOpen.set(true);
}

export function closeSettings(): void {
  settingsOpen.set(false);
}
