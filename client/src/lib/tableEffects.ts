// tableEffects — the generic effect cues (ADR 0209 §1 and §3, #4396).
//
// A permanent that is destroyed, exiled, sacrificed or bounced, a spell
// that is countered, a planeswalker that arrives or goes, and a board
// wipe each get one cue over the table, so a player sees what happened
// before they read the log. TableEffects.svelte measures, draws and
// plays; every rule is here, pure and unit-tested:
//
//   - which log entries are new (the combat beats' watermark, reused:
//     the first frame primes, a rewind forgets, a re-prime cancels);
//   - which cue a battlefield exit is, read from `new_zone` FIRST and
//     `cause` second (a destruction replaced by exile has cause
//     `destroy` and new_zone `exile`, and must look like an exile);
//   - which exits are a board wipe: three or more permanents leaving in
//     one `batch` from one `source`;
//   - the sound, the length (capped, then scaled by animations.speed)
//     and the three motion gates.
//
// It never touches game state, and nothing waits on it (ADR 0209 §1).

import type { LogEvent } from "./protocol";
import { seeded } from "./combatStrikes";
import type { SoundName } from "./sounds";

// ---- The setting and the gates ----

/** animations.effects (ADR 0209 §1): how much of each cue plays. */
export type EffectsMode = "full" | "short" | "off";

export const DEFAULT_EFFECTS_MODE: EffectsMode = "full";

export function isEffectsMode(v: unknown): v is EffectsMode {
  return v === "full" || v === "short" || v === "off";
}

/**
 * How a cue is shown: "full" plays it, "short" plays it cut to
 * EFFECT_SHORT_CAP_MS, "still" shows the outcome badge and no motion.
 */
export type EffectMotion = "full" | "short" | "still";

/**
 * effectMotion folds the signals. Any of them makes a cue still: the
 * animations master switch off, `animations.effects` off, the setting
 * `accessibility.reduceMotion`, or the OS's `prefers-reduced-motion`
 * (DESIGN.md §4, all three). accessibility.reduceMotion is never pushed
 * into animations.ts, so the caller passes every signal.
 */
export function effectMotion(s: {
  enabled: boolean;
  effects: EffectsMode;
  reduceMotion: boolean;
  osReduceMotion?: boolean;
}): EffectMotion {
  if (!s.enabled || s.effects === "off" || s.reduceMotion || s.osReduceMotion === true) {
    return "still";
  }
  return s.effects === "short" ? "short" : "full";
}

/** The longest a generic cue plays, before animations.speed (ADR 0209 §1). */
export const EFFECT_CAP_MS = 1500;
/** The longest a cue plays in the Short setting, before animations.speed. */
export const EFFECT_SHORT_CAP_MS = 800;
/**
 * How long a still badge stays up. It is reading time, not motion, so
 * animations.speed does not scale it (StackLinger's 1.5 s is the same).
 */
export const EFFECT_STILL_HOLD_MS = 1200;
/** A cue that starts later than this after it was due is dropped. */
export const EFFECT_LATE_MS = 1000;
/** Most cues drawn from one frame; a bigger wipe draws its sweep and the first few. */
export const EFFECT_CUE_CAP = 12;

/** animations.speed as a safe multiplier (2 = twice as slow). */
export function speedFactor(speed: number): number {
  return Number.isFinite(speed) && speed > 0 ? speed : 1;
}

/**
 * cueMs is a cue's length: its base length, capped (1.5 s, or 0.8 s in
 * Short), then scaled by animations.speed. A still cue has no motion:
 * it is the badge's hold, which speed does not scale.
 */
export function cueMs(baseMs: number, motion: EffectMotion, speed: number): number {
  if (motion === "still") return EFFECT_STILL_HOLD_MS;
  const cap = motion === "short" ? EFFECT_SHORT_CAP_MS : EFFECT_CAP_MS;
  return Math.max(1, Math.min(baseMs, cap) * speedFactor(speed));
}

// ---- The cues ----

export type EffectKind =
  | "destroy"
  | "exile"
  | "dies"
  | "sacrifice"
  | "bounce"
  | "counter"
  | "pw_enter"
  | "pw_leave";

/** Base length of each cue at speed 1, before the cap. */
export const EFFECT_BASE_MS: Record<EffectKind, number> = {
  destroy: 900,
  exile: 1100,
  // The existing crumble (STRIKE_CRUMBLE_MS in animations.ts).
  dies: 420,
  sacrifice: 800,
  bounce: 900,
  counter: 900,
  pw_enter: 900,
  pw_leave: 900,
};

/** The sound each cue plays. A death by state-based action has none new. */
export const EFFECT_SOUND: Record<EffectKind, SoundName | null> = {
  destroy: "destroy",
  exile: "exile",
  dies: null,
  sacrifice: "sacrifice",
  bounce: "bounce",
  counter: "counter",
  pw_enter: "pw_enter",
  pw_leave: "pw_leave",
};

/** Three or more permanents leaving in one batch from one source. */
export const WIPE_MIN = 3;
/** The sweep's length, and a per-card cue's length inside a wipe. */
export const WIPE_SWEEP_MS = 700;
export const WIPE_CARD_MS = 600;
/** The first per-card cue starts this long into the sweep. */
export const WIPE_FIRST_CARD_MS = 250;
/** Per-card cues are spread over this long at most, one stagger apart. */
export const WIPE_SPREAD_MS = 500;
export const WIPE_STAGGER_MAX_MS = 60;
/** The volume a per-card sound plays at inside a wipe. */
export const WIPE_CARD_VOLUME = 0.45;

/** The whole wipe, sweep and the last per-card cue, fits one cue's cap. */
export const WIPE_TOTAL_MS = WIPE_FIRST_CARD_MS + WIPE_SPREAD_MS + WIPE_CARD_MS;

export type WipeTone = "dark" | "cold" | "cool";

export interface EffectCue {
  /** Unique within the plan: the entry's seq, with a suffix for a second cue of one entry. */
  id: string;
  seq: number;
  kind: EffectKind;
  /** The permanent (or, for a counter, the spell's card), when the entry names one. */
  cardID: string | null;
  /** For a counter: the stack item that was countered. */
  stackItemID: string | null;
  batch: number | null;
  /** "light" inside a board wipe: shorter, no flourish, a quieter sound. */
  weight: "full" | "light";
  /** Start offset at speed 1 (a wipe staggers its per-card cues). */
  delayMs: number;
}

export interface BoardWipe {
  batch: number;
  source: string;
  /** How many permanents left. */
  count: number;
  tone: WipeTone;
  /** The first of its entries, to order it among the frame's cues. */
  seq: number;
}

export interface EffectPlan {
  cues: EffectCue[];
  wipes: BoardWipe[];
}

export interface PlanContext {
  /** Whether this card is (or just was) a planeswalker, from the views the client holds. */
  isPlaneswalker(cardID: string): boolean;
  /**
   * The combat strike layer is playing this frame (ADR 0134): it crumbles
   * the creatures combat damage kills, so a state-based death in a combat
   * damage step is not drawn twice.
   */
  combatMotion: boolean;
}

const DAMAGE_STEPS = new Set(["first_strike_damage", "combat_damage"]);

interface Exit {
  e: LogEvent;
  step: string;
  /** The zone entry for this card in this batch, if there is one. */
  zone: LogEvent | null;
  sacrificed: boolean;
}

/**
 * exitKind is the cue a permanent leaving the battlefield gets, or null
 * for none. `new_zone` is read first: an exile is an exile whatever the
 * instruction said, and a card that went to a hand is a bounce.
 */
export function exitKind(
  x: Pick<Exit, "zone" | "sacrificed" | "step">,
  ctx: PlanContext,
): EffectKind | null {
  const z = x.zone;
  // A destruction replaced by exile (Rest in Peace) is `destroy` with
  // new_zone exile: it must show the exile look, never a death.
  if (z?.new_zone === "exile") return "exile";
  if (z?.new_zone === "hand" || (z && z.cause === "bounce")) return "bounce";
  if (x.sacrificed) return "sacrifice";
  if (!z) return null;
  // Into a library, the command zone or anywhere unseen: no cue.
  if (z.new_zone !== "graveyard") return null;
  const cardID = z.card_id ?? "";
  switch (z.cause) {
    case "destroy":
      return cardID && ctx.isPlaneswalker(cardID) ? "pw_leave" : "destroy";
    case "sba":
    case "other":
    case undefined:
    case "": {
      // Combat damage's deaths are the strike layer's crumble.
      if (z.cause === "sba" && ctx.combatMotion && DAMAGE_STEPS.has(x.step)) return null;
      return cardID && ctx.isPlaneswalker(cardID) ? "pw_leave" : "dies";
    }
    default:
      return "dies";
  }
}

function isExit(e: LogEvent): boolean {
  return (e.kind === "zone" && e.old_zone === "battlefield") || e.kind === "sacrifice";
}

/**
 * planEffects turns a frame's new log entries into cues. `fresh` is the
 * entries this client has not handled, `log` the frame's whole log (an
 * entry's step is on an older entry).
 */
export function planEffects(
  fresh: readonly LogEvent[],
  log: readonly LogEvent[] | undefined,
  ctx: PlanContext,
): EffectPlan {
  if (fresh.length === 0) return { cues: [], wipes: [] };
  const freshSeqs = new Set(fresh.map((e) => e.seq));

  // One pass over the log, in order, so each fresh entry knows its step.
  let step = "";
  const exits = new Map<string, Exit>();
  const order: Exit[] = [];
  const cues: EffectCue[] = [];
  for (const e of log ?? []) {
    if (e.kind === "step") {
      step = e.step ?? "";
      continue;
    }
    if (!freshSeqs.has(e.seq)) continue;

    if (e.kind === "counter") {
      const item = e.stack_item_id ?? e.target ?? null;
      cues.push({
        id: String(e.seq),
        seq: e.seq,
        kind: "counter",
        cardID: e.target ?? null,
        stackItemID: item,
        batch: e.batch ?? null,
        weight: "full",
        delayMs: 0,
      });
      continue;
    }

    if (e.kind === "zone" && e.new_zone === "battlefield" && e.old_zone !== "battlefield") {
      if (e.card_id && ctx.isPlaneswalker(e.card_id)) {
        cues.push({
          id: String(e.seq),
          seq: e.seq,
          kind: "pw_enter",
          cardID: e.card_id,
          stackItemID: null,
          batch: e.batch ?? null,
          weight: "full",
          delayMs: 0,
        });
      }
      continue;
    }

    if (!isExit(e)) continue;
    // A card's exits in one batch are one: a sacrifice entry and the zone
    // move that follows it (when the route was not the graveyard) are one cue.
    const key = `${e.batch ?? "-"}:${e.card_id ?? `q${e.seq}`}`;
    let x = exits.get(key);
    if (!x) {
      x = { e, step, zone: null, sacrificed: false };
      exits.set(key, x);
      order.push(x);
    }
    if (e.kind === "sacrifice") x.sacrificed = true;
    else x.zone = e;
  }

  // Board wipes: three or more exits in one batch from one source.
  const groups = new Map<string, Exit[]>();
  for (const x of order) {
    const batch = x.e.batch;
    const source = x.e.source;
    if (batch === undefined || batch === null || !source) continue;
    const k = `${batch}|${source}`;
    const g = groups.get(k);
    if (g) g.push(x);
    else groups.set(k, [x]);
  }
  const wipes: BoardWipe[] = [];
  const inWipe = new Map<Exit, { index: number; of: number }>();
  for (const [, g] of groups) {
    if (g.length < WIPE_MIN) continue;
    g.forEach((x, index) => inWipe.set(x, { index, of: g.length }));
    wipes.push({
      batch: g[0].e.batch as number,
      source: g[0].e.source as string,
      count: g.length,
      tone: toneOf(g, ctx),
      seq: g[0].e.seq,
    });
  }

  for (const x of order) {
    const kind = exitKind(x, ctx);
    if (!kind) continue;
    const w = inWipe.get(x);
    const stagger = w ? Math.min(WIPE_STAGGER_MAX_MS, WIPE_SPREAD_MS / Math.max(1, w.of - 1)) : 0;
    cues.push({
      id: String(x.e.seq),
      seq: x.e.seq,
      kind,
      cardID: x.e.card_id ?? null,
      stackItemID: null,
      batch: x.e.batch ?? null,
      weight: w ? "light" : "full",
      delayMs: w ? WIPE_FIRST_CARD_MS + w.index * stagger : 0,
    });
  }
  cues.sort((a, b) => a.seq - b.seq);
  return { cues, wipes };
}

function toneOf(g: readonly Exit[], ctx: PlanContext): WipeTone {
  const n = { dark: 0, cold: 0, cool: 0 };
  for (const x of g) {
    const k = exitKind(x, ctx);
    if (k === "exile") n.cold++;
    else if (k === "bounce") n.cool++;
    else n.dark++;
  }
  if (n.cold > n.dark && n.cold >= n.cool) return "cold";
  if (n.cool > n.dark && n.cool > n.cold) return "cool";
  return "dark";
}

/**
 * effectLate: a cue starting more than EFFECT_LATE_MS after it was due
 * (its frame's arrival plus its stagger) is dropped, the way a late
 * combat strike is (ADR 0134 §4): the board has moved on.
 */
export function effectLate(frameAt: number, atMs: number, now: number): boolean {
  return now - (frameAt + atMs) > EFFECT_LATE_MS;
}

/**
 * effectsScheduled: a frame that arrives while the page is hidden draws
 * no motion (Web Animations follow the frame clock, which a hidden tab
 * does not run), though its sounds still play.
 */
export function effectsScheduled(visibility: string | undefined): boolean {
  return visibility === undefined || visibility === "visible";
}

/** The key one logical sound is capped on: once per batch, per game. */
export function soundBatch(gameID: string, batch: number | null): string | undefined {
  return batch === null ? undefined : `${gameID}:${batch}`;
}

/** The sound a cue plays, and its volume hint (a wipe's per-card sounds are quieter). */
export function cueSound(cue: Pick<EffectCue, "kind" | "weight">): {
  name: SoundName;
  volume: number;
} | null {
  const name = EFFECT_SOUND[cue.kind];
  if (!name) return null;
  return { name, volume: cue.weight === "light" ? WIPE_CARD_VOLUME : 1 };
}

// ---- Shapes ----

/**
 * slashGeometry is the dark slash across a card `w` x `h`: it runs along
 * the card's diagonal, a little longer than it, at the diagonal's angle
 * (degrees, clockwise on screen, from the top-left corner down).
 */
export function slashGeometry(w: number, h: number): { length: number; angle: number } {
  const length = Math.hypot(w, h) * 1.15;
  const angle = (Math.atan2(h, w) * 180) / Math.PI;
  return { length: Math.round(length * 100) / 100, angle: Math.round(angle * 100) / 100 };
}

export interface BurstShard {
  /** Where it starts, in percent of the spot's box. */
  x: number;
  y: number;
  /** How far it flies, in px at the box's own scale, its spin in degrees and its size in px. */
  dx: number;
  dy: number;
  rot: number;
  size: number;
}

export const BURST_CRACKS = 7;
export const BURST_SHARDS = 9;

/**
 * counterBurst plans a counterspell's shatter over a stack item `w` x `h`:
 * jagged cracks run out from the middle (path data over a 0..100 box,
 * each drawn with pathLength 1), and glass shards fly out of it. Seeded by
 * the entry, so a cue always looks the same.
 */
export function counterBurst(
  seed: string,
  w: number,
  h: number,
): { cracks: string[]; shards: BurstShard[] } {
  const rnd = seeded(seed);
  const cracks: string[] = [];
  for (let i = 0; i < BURST_CRACKS; i++) {
    const base = ((i + rnd() * 0.6) / BURST_CRACKS) * Math.PI * 2;
    let x = 50;
    let y = 50;
    let d = `M50 50`;
    const steps = 3 + Math.floor(rnd() * 2);
    for (let s = 1; s <= steps; s++) {
      const a = base + (rnd() - 0.5) * 0.7;
      const r = (s / steps) * (46 + rnd() * 10);
      x = clamp(50 + Math.cos(a) * r, 0, 100);
      y = clamp(50 + Math.sin(a) * r, 0, 100);
      d += ` L${round1(x)} ${round1(y)}`;
    }
    cracks.push(d);
  }
  const reach = Math.max(w, h);
  const shards: BurstShard[] = [];
  for (let i = 0; i < BURST_SHARDS; i++) {
    const a = ((i + rnd() * 0.8) / BURST_SHARDS) * Math.PI * 2;
    const dist = reach * (0.35 + rnd() * 0.45);
    shards.push({
      x: round1(50 + Math.cos(a) * 12),
      y: round1(50 + Math.sin(a) * 12),
      dx: round1(Math.cos(a) * dist),
      dy: round1(Math.sin(a) * dist + reach * 0.15),
      rot: round1((rnd() * 2 - 1) * 200),
      size: round1(8 + rnd() * 10),
    });
  }
  return { cracks, shards };
}

function clamp(v: number, lo: number, hi: number): number {
  return Math.min(hi, Math.max(lo, v));
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}

// ---- Timing ----

/**
 * How much a board wipe is compressed in Short: the whole wipe (the sweep
 * and the last per-card cue) fits EFFECT_SHORT_CAP_MS instead of
 * WIPE_TOTAL_MS.
 */
function wipeScale(motion: EffectMotion): number {
  return motion === "short" ? Math.min(1, EFFECT_SHORT_CAP_MS / WIPE_TOTAL_MS) : 1;
}

/**
 * cueTiming is when a cue's motion starts (an offset from its frame) and
 * how long it runs, both in real milliseconds, scaled by animations.speed.
 * A cue inside a wipe waits for the sweep and is lighter; the cap holds
 * for the whole wipe, so delay plus length never passes the cap (before
 * speed). A still cue starts at once and is the badge's hold.
 */
export function cueTiming(
  cue: Pick<EffectCue, "kind" | "weight" | "delayMs">,
  motion: EffectMotion,
  speed: number,
): { delayMs: number; ms: number } {
  if (motion === "still") return { delayMs: 0, ms: EFFECT_STILL_HOLD_MS };
  const s = speedFactor(speed);
  if (cue.weight === "light") {
    const k = wipeScale(motion);
    return {
      delayMs: cue.delayMs * k * s,
      ms: Math.max(1, Math.min(EFFECT_BASE_MS[cue.kind], WIPE_CARD_MS) * k * s),
    };
  }
  return { delayMs: 0, ms: cueMs(EFFECT_BASE_MS[cue.kind], motion, speed) };
}

/** The sweep's length for a board wipe, in real milliseconds. */
export function wipeTiming(motion: EffectMotion, speed: number): { ms: number } {
  if (motion === "still") return { ms: EFFECT_STILL_HOLD_MS };
  return { ms: Math.max(1, WIPE_SWEEP_MS * wipeScale(motion) * speedFactor(speed)) };
}
