// stackArrival — the trace an item leaves as it goes on the stack (ADR
// 0119 amendment 2026-10-16 §D, call 22).
//
// When a NEW item goes on the stack, a trace runs from where it came from
// to where the stack draws it: from the caster's avatar for a spell, from
// the source permanent for an ability (the element the stack's source
// glow already marks, `data-stack-source`), or from the caster's avatar
// when that permanent is not on screen. It follows the stack's own
// target-arrow curve (stackArrows.ts: bowed toward the table's centre,
// clear of the stack panel), in the caster's seat colour: a short bright
// head and a tail of the arrows' own dashes, fading, that runs the curve
// in ARRIVAL_MS × animations.speed and drains into the card. Then the
// gold target arrows draw, as they always have.
//
// It is ours, not another client's sweep: no energy arc, no flare. A
// burst (a board of triggers) is staggered ARRIVAL_STAGGER_MS apart, at
// most ARRIVAL_MAX at once, the newest; the rest arrive without one.
//
// WHAT IS NEW: an item ID on the stack that was not on it the frame
// before. PRIMING: the first frame, a reconnect or a replay toggle (the
// board's beatsPrimeKey) and an undo (the log's highest seq falls)
// trace nothing, so a snapshot never replays the stack's history.
//
// GATED three ways (DESIGN.md §4): `accessibility.reduceMotion` (the
// `data-reduce-motion` root switch), `animations.enabled`, and its own
// `animations.stackArrival`. Nothing is lost without it: the card still
// appears and the arrows still draw.
//
// Plain functions over plain data, pinned by stackArrival.test.ts;
// StackArrival.svelte measures and draws.

import type { StackItemView } from "./protocol";
import {
  BOARD_BOW,
  boxCentre,
  controlTowardCentre,
  edgePoint,
  insideBox,
  type Box,
  type Curve,
} from "./stackArrows";

/** The run along the curve, before animations.speed. */
export const ARRIVAL_MS = 320;
/** A burst's traces start this far apart. */
export const ARRIVAL_STAGGER_MS = 80;
/** At most this many traces run from one frame; the rest arrive without one. */
export const ARRIVAL_MAX = 4;
/** The trace stops this far short of the card it lands on, in px. */
export const ARRIVAL_GAP = 3;

/** The wire fields the tracker reads. */
export type ArrivalItem = Pick<StackItemView, "id" | "kind" | "controller" | "source_card_id">;

export interface ArrivalTracker {
  primed: boolean;
  /** The stack's item IDs at the last frame. */
  ids: ReadonlySet<string>;
  /** The log's highest seq at the last frame, to see an undo. */
  logTop: number;
}

export function emptyArrivalTracker(): ArrivalTracker {
  return { primed: false, ids: new Set(), logTop: 0 };
}

/** One trace to run. */
export interface Arrival {
  itemID: string;
  /** The caster's seat ID: the avatar a spell's trace leaves from. */
  casterSeat: string;
  /** An ability's source permanent, or null for a spell. */
  sourceCardID: string | null;
  /** Its start after the frame, in ms at speed 1: the burst's stagger. */
  delayMs: number;
}

export interface ArrivalFrame {
  tracker: ArrivalTracker;
  arrivals: Arrival[];
}

export interface ArrivalInput {
  /** The stack as the wire sends it, bottom to top. */
  stackItems: readonly ArrivalItem[] | undefined;
  log?: readonly { seq: number }[] | undefined;
}

function highestSeq(log: readonly { seq: number }[] | undefined): number {
  let top = 0;
  for (const e of log ?? []) if (e.seq > top) top = e.seq;
  return top;
}

/**
 * trackArrivals folds one frame into the tracker: the items new on the
 * stack since the last frame, in the order they went on it (bottom
 * first), the newest ARRIVAL_MAX of them, each with its stagger. The
 * input tracker is not mutated.
 */
export function trackArrivals(
  prev: ArrivalTracker,
  frame: ArrivalInput,
  opts: { reprime?: boolean } = {},
): ArrivalFrame {
  const items = frame.stackItems ?? [];
  const logTop = highestSeq(frame.log);
  const tracker: ArrivalTracker = { primed: true, ids: new Set(items.map((i) => i.id)), logTop };
  if (!prev.primed || opts.reprime === true || logTop < prev.logTop) {
    return { tracker, arrivals: [] };
  }
  const fresh = items.filter((i) => !prev.ids.has(i.id));
  const kept = fresh.slice(Math.max(0, fresh.length - ARRIVAL_MAX));
  return {
    tracker,
    arrivals: kept.map((i, n) => ({
      itemID: i.id,
      casterSeat: i.controller,
      sourceCardID: i.kind === "spell" ? null : i.source_card_id || null,
      delayMs: n * ARRIVAL_STAGGER_MS,
    })),
  };
}

/** The trace's gate: reduced motion, the master switch and its own toggle. */
export function arrivalAllowed(s: {
  animations: { enabled: boolean; stackArrival: boolean };
  accessibility: { reduceMotion: boolean };
}): boolean {
  return s.animations.enabled && s.animations.stackArrival && !s.accessibility.reduceMotion;
}

/** The run's length in ms for a speed multiplier (a bad one counts as 1). */
export function arrivalDuration(speed: number): number {
  const s = Number.isFinite(speed) && speed > 0 ? speed : 1;
  return ARRIVAL_MS * s;
}

/**
 * The trace's curve, in board coordinates: from the source's edge
 * facing the card to ARRIVAL_GAP short of the card's edge, bowed toward
 * the table's centre like the stack's target arrows, and to the side
 * away from the stack panel (`avoid`) when the near side would cross it.
 */
export function arrivalCurve(
  from: Box,
  to: Box,
  board: { width: number; height: number },
  avoid?: Box | null,
): Curve {
  const start = edgePoint(from, boxCentre(to));
  const end = edgePoint(to, start, ARRIVAL_GAP);
  const centre = { x: board.width / 2, y: board.height / 2 };
  const ctrl = controlTowardCentre(start, end, centre, BOARD_BOW, avoid);
  return { x1: start.x, y1: start.y, cx: ctrl.x, cy: ctrl.y, x2: end.x, y2: end.y };
}

/**
 * Whether a trace may start from `start` (board coordinates): its centre
 * is on the board, and not under the stack panel (`cover`), where it
 * cannot be seen. On a two-seat table the pile docks on the left, over
 * the opponent's avatar: a trace from there would run inside the card
 * it lands on. The same test the target arrows use (stackArrows.ts).
 */
export function startShows(
  start: Box,
  board: { width: number; height: number },
  cover?: Box | null,
): boolean {
  if (start.width <= 0 && start.height <= 0) return false;
  const c = boxCentre(start);
  if (!insideBox(c, { left: 0, top: 0, width: board.width, height: board.height })) return false;
  return !(cover && insideBox(c, cover));
}

/** A quadratic curve's length, by sampling it. */
export function curveLength(c: Curve, samples = 32): number {
  let len = 0;
  let px = c.x1;
  let py = c.y1;
  for (let i = 1; i <= samples; i++) {
    const t = i / samples;
    const u = 1 - t;
    const x = u * u * c.x1 + 2 * u * t * c.cx + t * t * c.x2;
    const y = u * u * c.y1 + 2 * u * t * c.cy + t * t * c.y2;
    len += Math.hypot(x - px, y - py);
    px = x;
    py = y;
  }
  return len;
}

/** The head: a short solid run at the front. */
export const TRACE_HEAD_LEN = 12;
/** The tail: the target arrows' own rhythm, a 6px dash every 11px. */
export const TRACE_DASH = 6;
export const TRACE_DASH_GAP = 5;
/** How many tail dashes: one per this many px of curve, within the bounds. */
export const TRACE_TAIL_PER_PX = 48;
export const TRACE_TAIL_MIN = 2;
export const TRACE_TAIL_MAX = 6;

/** One moving dash of a trace. */
export interface TraceSegment {
  /** Its length along the curve, in px. */
  len: number;
  /** How far its front sits behind the head's front, in px. */
  back: number;
  /** 0..1. */
  opacity: number;
  head: boolean;
}

export interface TracePlan {
  segments: TraceSegment[];
  /** How far the head's front travels: the curve, then the tail draining into the card. */
  travel: number;
}

/**
 * The trace along a curve `length` px long: the head, then the tail's
 * dashes fading behind it. The front travels the curve and on by the
 * tail's length, so the whole trace pours into the card and is gone
 * when the run ends.
 */
export function tracePlan(length: number): TracePlan {
  const len = Math.max(0, length);
  const n = Math.max(TRACE_TAIL_MIN, Math.min(TRACE_TAIL_MAX, Math.round(len / TRACE_TAIL_PER_PX)));
  const segments: TraceSegment[] = [{ len: TRACE_HEAD_LEN, back: 0, opacity: 1, head: true }];
  const step = TRACE_DASH + TRACE_DASH_GAP;
  for (let i = 0; i < n; i++) {
    segments.push({
      len: TRACE_DASH,
      back: TRACE_HEAD_LEN + TRACE_DASH_GAP + i * step,
      // From 0.85 behind the head down to 0.2 at the end.
      opacity: Math.round((0.85 - (0.65 * i) / Math.max(1, n - 1)) * 100) / 100,
      head: false,
    });
  }
  const tail = TRACE_HEAD_LEN + TRACE_DASH_GAP + n * step;
  return { segments, travel: len + tail };
}
