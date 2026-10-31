// extort — the Scales of Souls table effect (#4544).
//
// When an extort trigger resolves and drains ("each opponent loses 1 life
// and you gain that much"), the table plays a short scene: a soul leaves
// each opponent that lost life and flies into the dark pan of a golden
// scale, the scale levels (the coin hits), and the light flies on to the
// extort player. ExtortScales.svelte measures, draws and plays; every rule
// is here, pure and unit-tested:
//
//   - WHAT AN EXTORT DRAIN IS. The wire has no "extort" flag and needs
//     none: a resolved extort trigger writes a `resolve` log entry whose
//     `label` ends " — extort" (every extort trigger is built by one
//     constructor, effects.Extort, and labelled "<card> — extort"), and
//     the life it moved is `life` entries in the same `batch` (everything
//     one resolution emits is one batch, CR 603.2c). Declining the {W/B}
//     writes no life, so there is nothing to play. A drain by any other
//     source (Gray Merchant, Blood Artist) is a different batch with no
//     such `resolve` entry and plays nothing. server/internal/cards/effects/
//     extort_log_test.go pins the label, the batch and the entries. A
//     source this viewer may not identify (a face-down permanent) has its
//     label redacted, and then the effect does not play.
//   - HOW THE DRAINS RUN (QUEUE_MAX below): one scale at a time. Drains
//     that resolve in one frame are one scale. A drain that resolves while
//     a scale is playing waits, and everything that arrives while it
//     waits merges into that one waiting scale (its souls add up), so a
//     run of ten extort triggers is at most two scales and about five
//     seconds, never ten scales stacked on the table.
//   - THE CLOCK. The scale's levelling, the coin hitting, lands at
//     EXTORT_HIT_MS; the `extort` sound is cut to put its coin there
//     (client/public/sounds/README.md).
//   - THE LIFE POPUPS stand down (popupIsExtort): the cue is the sound of
//     the whole drain, so the avatars' own `damage` and `heal` do not also
//     play over it.
//
// It never touches game state, and nothing waits on it (ADR 0209 §1).

import type { GameView, LogEvent } from "./protocol";
import {
  EFFECT_SHORT_CAP_MS,
  EFFECT_STILL_HOLD_MS,
  speedFactor,
  type EffectMotion,
} from "./tableEffects";

// ---- Reading the log ----

/** The end of an extort trigger's label on the wire ("Pontiff of Blight — extort"). */
export const EXTORT_LABEL_SUFFIX = " — extort";

export function isExtortLabel(label: string | undefined | null): boolean {
  return typeof label === "string" && label.endsWith(EXTORT_LABEL_SUFFIX);
}

/** One seat's share of a drain: how much life it lost, or gained. */
export interface ExtortShare {
  seat: number;
  amount: number;
}

/** An extort drain, as the log tells it. `count` is how many triggers it stands for. */
export interface ExtortDrain {
  /** The seats that lost life, in log order; `amount` is positive. */
  victims: ExtortShare[];
  /** The seats that gained life; `amount` is positive. */
  gainers: ExtortShare[];
  count: number;
}

interface Pending {
  batch: number;
  /** The seat that controlled the trigger (the resolve entry's `seat`), or -1. */
  seat: number;
  done: boolean;
}

/**
 * What the fold remembers between frames: the extort resolutions seen
 * (their batches), so life entries that arrive in a later frame than the
 * `resolve` line (a payment prompt in between) still find their drain.
 */
export interface ExtortTracker {
  pending: Pending[];
}

export function emptyExtortTracker(): ExtortTracker {
  return { pending: [] };
}

/** How many extort resolutions the tracker remembers. */
const PENDING_MAX = 8;

function addShare(list: ExtortShare[], seat: number, amount: number): void {
  const at = list.find((s) => s.seat === seat);
  if (at) at.amount += amount;
  else list.push({ seat, amount });
}

/**
 * foldExtort reads one frame's new log entries: it remembers each extort
 * `resolve`, and turns the `life` entries of a remembered batch into a
 * drain. A drain needs at least one seat that lost life; a trigger that
 * drained nothing (declined, or every loss replaced away) plays nothing.
 * Each batch drains once. Pure: the input tracker is not mutated.
 */
export function foldExtort(
  prev: ExtortTracker,
  fresh: readonly LogEvent[],
): { tracker: ExtortTracker; drains: ExtortDrain[] } {
  const pending = prev.pending.map((p) => ({ ...p }));
  for (const e of fresh) {
    if (e.kind !== "resolve" || !isExtortLabel(e.label) || e.batch === undefined) continue;
    if (!pending.some((p) => p.batch === e.batch)) {
      pending.push({ batch: e.batch, seat: e.seat, done: false });
    }
  }
  while (pending.length > PENDING_MAX) pending.shift();

  const drains: ExtortDrain[] = [];
  for (const p of pending) {
    if (p.done) continue;
    const victims: ExtortShare[] = [];
    const gainers: ExtortShare[] = [];
    for (const e of fresh) {
      if (e.kind !== "life" || e.batch !== p.batch) continue;
      const amount = e.amount ?? 0;
      if (amount < 0 && e.seat >= 0) addShare(victims, e.seat, -amount);
      else if (amount > 0 && e.seat >= 0) addShare(gainers, e.seat, amount);
    }
    if (victims.length === 0) continue;
    // The gain can be replaced away; the light still goes to the player
    // who extorted, whom the resolve line names.
    if (gainers.length === 0 && p.seat >= 0 && !victims.some((v) => v.seat === p.seat)) {
      gainers.push({ seat: p.seat, amount: 0 });
    }
    p.done = true;
    drains.push({ victims, gainers, count: 1 });
  }
  return { tracker: { pending }, drains };
}

// ---- Merging and queueing ----

/** mergeDrains adds drains up: per-seat amounts sum, `count` is the triggers they stand for. */
export function mergeDrains(drains: readonly ExtortDrain[]): ExtortDrain | null {
  if (drains.length === 0) return null;
  const victims: ExtortShare[] = [];
  const gainers: ExtortShare[] = [];
  let count = 0;
  for (const d of drains) {
    for (const v of d.victims) addShare(victims, v.seat, v.amount);
    for (const g of d.gainers) addShare(gainers, g.seat, g.amount);
    count += d.count;
  }
  return { victims, gainers, count };
}

/**
 * The scale in play and the one waiting. At most two scales for any run
 * of drains: everything that arrives while one waits is merged into it.
 */
export interface ExtortQueue {
  active: ExtortDrain | null;
  waiting: ExtortDrain | null;
}

export const emptyExtortQueue: ExtortQueue = { active: null, waiting: null };

/**
 * arrive adds a drain: it starts at once when no scale is playing
 * (`start`), else it waits, merged with any drain already waiting.
 */
export function arrive(
  q: ExtortQueue,
  drain: ExtortDrain,
): { queue: ExtortQueue; start: ExtortDrain | null } {
  if (!q.active) return { queue: { active: drain, waiting: null }, start: drain };
  const waiting = mergeDrains(q.waiting ? [q.waiting, drain] : [drain]);
  return { queue: { active: q.active, waiting }, start: null };
}

/** finish ends the scale in play and starts the waiting one, if any. */
export function finish(q: ExtortQueue): { queue: ExtortQueue; start: ExtortDrain | null } {
  return { queue: { active: q.waiting, waiting: null }, start: q.waiting };
}

// ---- The clock ----

/** The whole scene at speed 1, ms. */
export const EXTORT_MS = 2600;
/** The beats of the scene at speed 1, ms (the stylesheet's keyframes are these over EXTORT_MS). */
export const EXTORT_BEATS = {
  /** The souls leave their players and begin to fly. */
  soulsFrom: 150,
  /** The souls reach the dark pan; the arms begin to level. */
  soulsTo: 900,
  /** The arms reach level: the coin hits. The light bursts. */
  hit: 1250,
  /** The light's spark leaves the right pan for the extort player. */
  sparkFrom: 1450,
  /** The spark lands on the extort player's panel. */
  sparkTo: 2250,
  /** The scale begins to fade. */
  fadeFrom: 2300,
} as const;
/** When the coin hits, ms after the scene starts, at speed 1: the `extort` sound is cut to this. */
export const EXTORT_HIT_MS = EXTORT_BEATS.hit;

export type ExtortMotion = "full" | "still";

/**
 * How the scene is shown: with motion only in the Full setting and with
 * no reduced-motion signal. Short and every reduced-motion signal show a
 * static scale (levelled, no flight) and a fade; the sound plays either way.
 */
export function extortMotion(m: EffectMotion): ExtortMotion {
  return m === "full" ? "full" : "still";
}

/**
 * The scene's length in ms. Full is EXTORT_MS scaled by animations.speed
 * (it is not held to the generic cue cap: it carries a sound cut to its
 * clock). A static scene is a reading hold, which speed does not scale:
 * EFFECT_STILL_HOLD_MS, or EFFECT_SHORT_CAP_MS in the Short setting.
 */
export function extortMs(motion: EffectMotion, speed: number): number {
  if (motion === "full") return EXTORT_MS * speedFactor(speed);
  return motion === "short" ? EFFECT_SHORT_CAP_MS : EFFECT_STILL_HOLD_MS;
}

/** The stylesheet's time factor: real ms per scene ms at speed 1. */
export function extortK(motion: EffectMotion, speed: number): number {
  return motion === "full" ? speedFactor(speed) : 1;
}

// ---- The geometry ----

/** The scale's drawing, in its own units. */
export const SCALE_VIEW = { w: 360, h: 320 } as const;
/** The beam's pivot, and how far it leans before the coin (the left pan is down). */
const PIVOT = { x: 180, y: 102 } as const;
export const SCALE_TILT_DEG = -12;
/** The pans' orbs, with the beam level. */
const DARK_ORB = { x: 65, y: 185 } as const;
const LIGHT_ORB = { x: 295, y: 185 } as const;

export interface Box {
  /** The top left, in board px. */
  x: number;
  y: number;
  width: number;
  height: number;
}

/**
 * scaleBox is where the scale is drawn: the centre of the board, compact
 * (a fifth of its width, 150 to 230 px) so the cards a player needs stay
 * readable around it.
 */
export function scaleBox(board: { w: number; h: number }): Box {
  const width = Math.round(Math.min(230, Math.max(150, board.w * 0.2)));
  const height = Math.round((width * SCALE_VIEW.h) / SCALE_VIEW.w);
  return {
    x: Math.round((board.w - width) / 2),
    y: Math.round((board.h - height) / 2),
    width,
    height,
  };
}

function toBoard(box: Box, p: { x: number; y: number }): { x: number; y: number } {
  return {
    x: box.x + (p.x / SCALE_VIEW.w) * box.width,
    y: box.y + (p.y / SCALE_VIEW.h) * box.height,
  };
}

function rotateAbout(
  p: { x: number; y: number },
  c: { x: number; y: number },
  deg: number,
): { x: number; y: number } {
  const a = (deg * Math.PI) / 180;
  const dx = p.x - c.x;
  const dy = p.y - c.y;
  return {
    x: c.x + dx * Math.cos(a) - dy * Math.sin(a),
    y: c.y + dx * Math.sin(a) + dy * Math.cos(a),
  };
}

/** Where the dark pan's orb is while the beam leans: what the souls fly into, in board px. */
export function darkPanAt(box: Box): { x: number; y: number } {
  return toBoard(box, rotateAbout(DARK_ORB, PIVOT, SCALE_TILT_DEG));
}

/** Where the light orb is once the beam is level: where the spark leaves from, in board px. */
export function lightPanAt(box: Box): { x: number; y: number } {
  return toBoard(box, LIGHT_ORB);
}

/**
 * A soul's flight: from `from` (a player's avatar, board px) to the dark
 * pan, as offsets from the pan so the stylesheet can fly each one with
 * two custom properties. `i` staggers the souls a little.
 */
export interface Flight {
  /** The start, relative to the end. */
  dx: number;
  dy: number;
  /** Extra delay, ms at speed 1. */
  delay: number;
}

export const FLIGHT_STAGGER_MS = 70;

export function flight(
  from: { x: number; y: number },
  to: { x: number; y: number },
  i: number,
): Flight {
  return {
    dx: Math.round(from.x - to.x),
    dy: Math.round(from.y - to.y),
    delay: i * FLIGHT_STAGGER_MS,
  };
}

// ---- The life popups ----

/** The batches of the extort resolutions in a log. */
export function extortBatches(log: readonly LogEvent[] | undefined): Set<number> {
  const out = new Set<number>();
  for (const e of log ?? []) {
    if (e.kind === "resolve" && isExtortLabel(e.label) && e.batch !== undefined) out.add(e.batch);
  }
  return out;
}

/**
 * Whether a seat's life popup shows an extort drain, so it keeps quiet:
 * the `extort` cue is the sound of the whole thing. `deltas` is the
 * popup's, oldest first; they are matched against the seat's newest
 * logged changes one for one (as lifeCause.ts does), and false when they
 * do not line up (then the popup keeps its own sound).
 */
export function popupIsExtort(
  view: GameView,
  seatIndex: number,
  deltas: readonly number[],
): boolean {
  if (seatIndex < 0 || deltas.length === 0) return false;
  const batches = extortBatches(view.log);
  if (batches.size === 0) return false;
  const mine: { delta: number; extort: boolean }[] = [];
  for (const e of view.log ?? []) {
    const amount = e.amount ?? 0;
    if (e.kind === "life" && e.seat === seatIndex && amount !== 0) {
      mine.push({ delta: amount, extort: e.batch !== undefined && batches.has(e.batch) });
    } else if (e.kind === "damage" && e.target_seat === seatIndex && amount > 0) {
      mine.push({ delta: -amount, extort: false });
    }
  }
  if (mine.length < deltas.length) return false;
  const newest = mine.slice(mine.length - deltas.length);
  if (newest.some((c, k) => c.delta !== deltas[k])) return false;
  return newest.some((c) => c.extort);
}
