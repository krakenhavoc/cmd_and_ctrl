import { writable, get, type Writable } from "svelte/store";
import { setMuted, setVolumeMultiplier } from "./sounds";
import { setAnimationConfig } from "./animations";
import { STEP_IDS, NO_PRIORITY_STEPS, type StepID } from "./turn";

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
    // Per-step stops (S13). For each priority-granting step, true
    // means "stop here when priority lands on me" and false means
    // "auto-pass through it". Untap and Cleanup are not stoppable
    // (they don't grant priority) and are absent from this map.
    // Defaults seeded by defaultStepStops().
    stepStops: Record<string, boolean>;
    // S15: opt-in mana-cost enforcement. When true, the client
    // tags every cast_spell action with `strict: true` and the
    // server gates the cast on the caster's ManaPool actually
    // covering the printed cost (plus commander tax for casts
    // from the command zone). Default false (sandbox / paper
    // tracking). When the gate rejects, the client surfaces an
    // "Override strict mode for this cast" toast that re-fires
    // the action with `force_cast: true`.
    strictMana: boolean;
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

export const SETTINGS_VERSION = 4;
const STORAGE_KEY = "cmdctrl.settings.v1";
const LEGACY_MUTED_KEY = "cmdctrl.muted";

// defaultStepStops seeds the per-step stops map. The defaults match
// MTG Online's standard "stops" — the active player gets stopped on
// their main phases and combat declarations; everyone else passes
// through routine begin/end-step priority unless they opt in.
// Untap and Cleanup are excluded because they don't grant priority
// (CR 502.4 / 514.3); the server's NoPriority sentinel makes any
// attempt to pass during them a no-op anyway.
export function defaultStepStops(): Record<string, boolean> {
  const out: Record<string, boolean> = {};
  const opted: ReadonlySet<StepID> = new Set([
    "precombat_main",
    "declare_attackers",
    "declare_blockers",
    "postcombat_main",
    "end",
  ]);
  for (const id of STEP_IDS) {
    if (NO_PRIORITY_STEPS.has(id)) continue;
    out[id] = opted.has(id);
  }
  return out;
}

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
      // S13 default: on. Pre-S13 this was off because the only
      // gating was "not on viewer's own turn", which felt too
      // aggressive. The S13 stops grid (defaultStepStops) gives
      // the user fine control, so auto-pass-on is now the right
      // default — stops are the affordance for "stop here".
      autoPassPriority: true,
      stepStops: defaultStepStops(),
      // S15 default: off. Sandbox / paper-tracking is the
      // existing posture; players who want Arena-style "can't
      // cast yet" enforcement opt in via Settings.
      strictMana: false,
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
  const merged: Settings = {
    __version: SETTINGS_VERSION,
    audio: { ...d.audio, ...(s.audio ?? {}) },
    animations: { ...d.animations, ...(s.animations ?? {}) },
    display: { ...d.display, ...(s.display ?? {}) },
    gameplay: { ...d.gameplay, ...(s.gameplay ?? {}) },
    accessibility: { ...d.accessibility, ...(s.accessibility ?? {}) },
  };
  // v1 → v2 (S13): the gameplay.stepStops map was scaffolded as `{}`
  // pre-S13. Seed defaults for any user whose stored map is empty so
  // the per-step stops UI has something meaningful on first paint.
  // Existing user-configured maps are preserved untouched. Strip any
  // entries for no-priority steps (Untap / Cleanup) to keep the map
  // canonical. While we're here, flip `autoPassPriority` to true if
  // the user is still on the v1 default (false) — at v1 the toggle
  // was a global "auto-pass on opponents' turns" with no per-step
  // control, so off was the only safe default. With stops, the
  // toggle gates the per-step auto-pass and on is the natural
  // default. Pre-existing v1 users who explicitly turned it on stay
  // on; users who left it off get the new behaviour. We can't
  // distinguish "user explicitly left it off" from "user never
  // touched the default" at v1, but the worst case is "auto-pass
  // through opponents' turns the user wasn't expecting" — which
  // their stops grid (also being seeded here) prevents.
  const storedVersion = typeof s.__version === "number" ? s.__version : 0;
  const fromV1 = storedVersion < 2;
  if (Object.keys(merged.gameplay.stepStops).length === 0) {
    merged.gameplay.stepStops = defaultStepStops();
    if (fromV1) {
      merged.gameplay.autoPassPriority = true;
    }
  } else {
    for (const id of NO_PRIORITY_STEPS) {
      delete merged.gameplay.stepStops[id];
    }
  }
  // v2 → v3 (S13 hotfix): early v2 builds (the previous client
  // commit) shipped autoPassPriority=false through the v1→v2
  // migration even after the new "default true" landed. Anyone who
  // already migrated to v2 with stepStops seeded still has the
  // toggle off and hits the "game halts at draw" trap. v3 flips
  // autoPassPriority on for users coming from v2 whose stops grid
  // matches the seeded default — strong signal they haven't tuned
  // either knob, so re-applying the new pairing is safe. Users who
  // customised stops keep their autoPassPriority value untouched.
  if (storedVersion === 2 && stepStopsMatchDefault(merged.gameplay.stepStops)) {
    merged.gameplay.autoPassPriority = true;
  }
  // v3 → v4 (S15): the gameplay.strictMana toggle is new. The
  // shallow merge above already populated it from defaults
  // (false) for any v3 blob that omits the field; nothing else
  // to do here — calling it out so future migrations have a
  // hook to extend.
  return absorbLegacy(merged);
}

// stepStopsMatchDefault reports whether the supplied stepStops map
// is structurally identical to defaultStepStops(). Used by the v2→v3
// migration to detect "user hasn't customised stops" so we can
// safely re-seed autoPassPriority without overwriting an explicit
// off-toggle.
function stepStopsMatchDefault(actual: Record<string, boolean>): boolean {
  const expected = defaultStepStops();
  const actualKeys = Object.keys(actual);
  const expectedKeys = Object.keys(expected);
  if (actualKeys.length !== expectedKeys.length) return false;
  for (const k of expectedKeys) {
    if (actual[k] !== expected[k]) return false;
  }
  return true;
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

// exportSettings serialises the current settings to a JSON string
// suitable for copying to the clipboard. The Advanced tab uses this
// for the "copy my settings" button; a user on a second device can
// then paste the blob into importSettings to carry their prefs
// across without a server-side sync layer.
export function exportSettings(): string {
  return JSON.stringify(get(settings), null, 2);
}

// ImportResult reports what happened. `changed` signals whether the
// store actually updated — a valid no-op import (JSON matches
// current state) returns ok: true, changed: false so callers can
// flash an appropriate confirmation.
export interface ImportResult {
  ok: boolean;
  error?: string;
  changed?: boolean;
}

// importSettings validates the blob, merges via the same migration
// path used at module load, and writes to the store. Rejects on
// parse failure or on blobs that aren't a plain object. Schema
// mismatches are tolerated — missing fields fall back to defaults
// via migrate(), unknown fields are dropped.
export function importSettings(raw: string): ImportResult {
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch (e) {
    return { ok: false, error: (e as Error).message };
  }
  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
    return { ok: false, error: "settings JSON must be an object" };
  }
  const before = JSON.stringify(get(settings));
  const next = migrate(parsed);
  const after = JSON.stringify(next);
  settings.set(next);
  return { ok: true, changed: before !== after };
}

// fingerprintSettings returns a short alphanumeric hash of the
// current settings for bug reports. Not cryptographic — FNV-1a
// 32-bit keeps the implementation small and dependency-free, which
// is enough to distinguish "same prefs" from "different prefs" when
// comparing two players' reports.
export function fingerprintSettings(): string {
  const str = JSON.stringify(get(settings));
  let h = 0x811c9dc5;
  for (let i = 0; i < str.length; i++) {
    h ^= str.charCodeAt(i);
    h = (h + ((h << 1) + (h << 4) + (h << 7) + (h << 8) + (h << 24))) >>> 0;
  }
  return h.toString(36).padStart(7, "0");
}
