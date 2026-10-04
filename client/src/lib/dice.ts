// dice — the plan behind the dice layer (ADR 0121 §7).
//
// Every die a card rolls and every coin a card flips arrives as one
// `roll` or `flip` log entry, its result already drawn by the server
// (ADR 0054). This module turns those entries into a schedule of
// animations: who rolled, what they rolled, when each one starts to
// tumble, when it settles on the server's number, when it fades. The
// component (DiceLayer.svelte) only measures and draws.
//
// The rules, all pinned by dice.test.ts:
//
//   - One log entry is one animation. "Alice rolled 2d12" tumbles two
//     dice side by side and settles them together. At most six dice are
//     drawn; the rest are a "+N" chip, and the strip cue keeps every
//     result.
//   - Tumble 900 ms × `animations.speed`, hold 1.6 s (reading time, not
//     scaled), fade 200 ms.
//   - Each seat has its own queue, so four seats rolling at once are
//     four animations at four places. A seat with several entries in
//     one frame plays them in log order, each `max(400 ms, 2.5 s / n)`,
//     and keeps at most three queued, the one on screen included. The
//     older waiting ones are dropped: the log and the strip cue keep
//     them. That is ADR 0119 §3's burst rule with this layer's numbers.
//   - The faces a die shows while it tumbles are a fixed sequence seeded
//     from the entry, never `Math.random`, so every viewer sees the same
//     tumble and it always lands on the server's number.
//   - With motion off (the master switch, the `dice` toggle, or reduced
//     motion) nothing tumbles: the result appears settled at once and
//     holds for the same 1.6 s. The result is information; only the
//     motion is gated.
//   - The strip's text cue for an entry is released when its die
//     settles, so the text never gives the number away mid-tumble; with
//     motion off, or for an entry the queue dropped, at once.
//   - The first frame (and a reconnect or replay toggle) PRIMES: what is
//     already in the log happened before this client was watching, and
//     is never animated.
//
// The queue takes `DiceRoll`s, not log entries, so the opening roll
// (ADR 0121 PR 5) and table rolls (PR 6) feed the same schedule: each
// only needs its own `rollsFrom…` and a stable `key`.

import { isOpeningDieLog } from "./openingRoll";
import type { LogEvent } from "./protocol";

/** The tumble at speed 1. Scaled by `animations.speed`. */
export const DICE_TUMBLE_MS = 900;
/** How long a settled result stays up. Reading time: never scaled. */
export const DICE_HOLD_MS = 1600;
/** The fade out after the hold. */
export const DICE_FADE_MS = 200;
/** One entry's share of the screen: tumble plus hold. */
export const DICE_SLOT_MS = DICE_TUMBLE_MS + DICE_HOLD_MS;
/** The shortest slot an entry of a burst gets. */
export const DICE_BURST_MIN_MS = 400;
/** At most this many entries per seat: the one on screen and the ones waiting. */
export const DICE_QUEUE_MAX = 3;
/** At most this many dice (or coins) are drawn for one entry; the rest are "+N". */
export const DICE_SHOWN_MAX = 6;
/** How many faces a die shows on its way to the result, the result included. */
export const DICE_TUMBLE_STEPS = 10;

export type CoinFace = "heads" | "tails";

/** Where a roll came from. `table` arrives with ADR 0121 PR 6. */
export type DiceSource = "card" | "opening" | "table";

/** One animation's worth of randomness: one log entry, already drawn. */
export interface DiceRoll {
  /**
   * Identity across frames. A card roll is its log `seq`; a table roll
   * (PR 6) will be its `roll_id`, so a line re-emitted after an undo is
   * not animated twice.
   */
  key: string;
  /** The log entry this came from, for the strip cue and the announcer. */
  seq: number;
  /** The roller's seat index. */
  seat: number;
  kind: "die" | "coin";
  /** Faces on the die; 2 for a coin. */
  sides: number;
  /** Each die's result, in roll order. Empty for coins. */
  results: number[];
  /** Each coin's face, in flip order. Empty for dice. */
  faces: CoinFace[];
  /** A called flip's call (ADR 0054's `coin_call`). */
  call?: CoinFace;
  /** The log's public line: the announcer reads it. */
  text: string;
  source: DiceSource;
}

/** One scheduled animation. All times are absolute, in ms (Date.now()). */
export interface DicePlay {
  roll: DiceRoll;
  /** When it appears and starts to tumble. */
  startAt: number;
  /** When it lands on the result. Equal to startAt with motion off. */
  settleAt: number;
  /** When the hold ends and the fade begins. */
  fadeAt: number;
  /** When it is gone. */
  endAt: number;
  /** Whether it tumbles at all. */
  motion: boolean;
  /** Its share of the screen before speed: DICE_SLOT_MS, or a burst's slice. */
  slotMs: number;
  /** When the strip cue and the announcer may say the result. */
  releaseAt: number;
}

export interface DiceState {
  /** Every key handled: animated, dropped or primed. */
  seen: Set<string>;
  /** The schedule, every seat's, in no particular order. */
  plays: DicePlay[];
  /** seq → when its strip cue may show. Primed and dropped entries are released at once. */
  released: Map<number, number>;
  /** seqs primed rather than played: released, but never announced. */
  primed: Set<number>;
}

export interface DiceOptions {
  /** Whether rolls tumble: master switch on, `dice` on, motion not reduced. */
  motion: boolean;
  /** `animations.speed`; scales the tumble only. */
  speed: number;
}

export function emptyDiceState(): DiceState {
  return { seen: new Set(), plays: [], released: new Map(), primed: new Set() };
}

/** Whether a die or coin animation should move, from the animation settings. */
export function diceMotion(s: { enabled: boolean; dice: boolean; reduceMotion: boolean }): boolean {
  return s.enabled && s.dice && !s.reduceMotion;
}

function coinFace(s: string | undefined): CoinFace | null {
  return s === "heads" || s === "tails" ? s : null;
}

/**
 * rollFromLog is the animation a log entry asks for, or null: a card's
 * `roll` or `flip`, or an opening d20 (a `roll` with no card before the
 * first turn, ADR 0121 PR 5), which is drawn here too and marked
 * `opening` so the strip raises no cue for it. An entry with no result
 * to show (an older server, a redacted line) is null: there is nothing
 * to land on.
 */
export function rollFromLog(log: LogEvent): DiceRoll | null {
  const base = {
    key: `seq:${log.seq}`,
    seq: log.seq,
    seat: log.seat,
    text: log.text,
    source: (isOpeningDieLog(log) ? "opening" : "card") as DiceSource,
  };
  if (log.kind === "roll") {
    const results = (log.results ?? []).filter((n) => Number.isInteger(n));
    const sides = log.sides ?? 0;
    if (results.length === 0 || sides < 1) return null;
    return { ...base, kind: "die", sides, results, faces: [] };
  }
  if (log.kind === "flip") {
    const faces = (log.faces ?? []).map(coinFace).filter((f): f is CoinFace => f !== null);
    if (faces.length === 0) return null;
    const call = coinFace(log.call) ?? undefined;
    return { ...base, kind: "coin", sides: 2, results: [], faces, ...(call ? { call } : {}) };
  }
  return null;
}

/** Every animation the log window carries, in log order. */
export function rollsFromLogs(logs: readonly LogEvent[] | undefined): DiceRoll[] {
  const out: DiceRoll[] = [];
  for (const log of logs ?? []) {
    const r = rollFromLog(log);
    if (r) out.push(r);
  }
  return out;
}

/**
 * How long one entry of a burst of `n` (from one seat, in one frame)
 * keeps the screen: 2.5 s shared out, never under 400 ms.
 */
export function diceSlotMs(n: number): number {
  return Math.max(DICE_BURST_MIN_MS, DICE_SLOT_MS / Math.max(1, n));
}

/** The phases of one entry given its slot, starting at `startAt`. */
export function diceTimings(
  startAt: number,
  slotMs: number,
  opts: DiceOptions,
): Pick<DicePlay, "startAt" | "settleAt" | "fadeAt" | "endAt"> {
  // A burst shortens tumble and hold in proportion; a lone entry gets
  // exactly 900 ms × speed and 1.6 s.
  const f = Math.min(1, slotMs / DICE_SLOT_MS);
  const tumble = opts.motion ? Math.round(DICE_TUMBLE_MS * opts.speed * f) : 0;
  const hold = Math.round(DICE_HOLD_MS * f);
  const settleAt = startAt + tumble;
  const fadeAt = settleAt + hold;
  return { startAt, settleAt, fadeAt, endAt: fadeAt + DICE_FADE_MS };
}

/**
 * primeDice marks everything already in the window as handled, without
 * animating or announcing any of it.
 */
export function primeDice(rolls: readonly DiceRoll[]): DiceState {
  const s = emptyDiceState();
  for (const r of rolls) {
    s.seen.add(r.key);
    s.released.set(r.seq, 0);
    s.primed.add(r.seq);
  }
  return s;
}

/**
 * trackDice folds one frame's rolls (the whole window, in log order)
 * into the schedule. `now` is passed so the rules are testable. Returns
 * a new state; the input is not mutated.
 */
export function trackDice(
  prev: DiceState,
  rolls: readonly DiceRoll[],
  now: number,
  opts: DiceOptions,
): DiceState {
  // The window bounds memory, and an undo that rewinds a roll takes its
  // entry out of the log: forgetting it lets the replayed roll (often
  // the same seq, ADR 0054 Decision 4) animate again.
  const keys = new Set(rolls.map((r) => r.key));
  const seqs = new Set(rolls.map((r) => r.seq));
  const seen = new Set([...prev.seen].filter((k) => keys.has(k)));
  const released = new Map([...prev.released].filter(([seq]) => seqs.has(seq)));
  const primed = new Set([...prev.primed].filter((seq) => seqs.has(seq)));
  // A play whose entry was rewound stops; one that has finished goes.
  const live = prev.plays.filter((p) => keys.has(p.roll.key) && p.endAt > now);

  const fresh: DiceRoll[] = [];
  for (const r of rolls) {
    if (seen.has(r.key)) continue;
    seen.add(r.key);
    fresh.push(r);
  }
  if (fresh.length === 0) return { seen, plays: live, released, primed };

  const bySeat = new Map<number, DiceRoll[]>();
  for (const r of fresh) bySeat.set(r.seat, [...(bySeat.get(r.seat) ?? []), r]);

  let plays = live;
  for (const [seat, incoming] of bySeat) {
    const mine = plays.filter((p) => p.roll.seat === seat).sort((a, b) => a.startAt - b.startAt);
    const others = plays.filter((p) => p.roll.seat !== seat);
    const showing = mine.find((p) => p.startAt <= now) ?? null;
    const waiting = mine.filter((p) => p !== showing);
    const slot = diceSlotMs(incoming.length);
    type Pending = { roll: DiceRoll; slotMs: number; motion: boolean };
    const all: Pending[] = [
      ...waiting.map((p) => ({ roll: p.roll, slotMs: p.slotMs, motion: p.motion })),
      ...incoming.map((roll) => ({ roll, slotMs: slot, motion: opts.motion })),
    ];
    // The one on screen is never cut short; the oldest waiting ones go.
    const room = Math.max(0, DICE_QUEUE_MAX - (showing ? 1 : 0));
    const keep = all.slice(Math.max(0, all.length - room));
    for (const d of all.slice(0, all.length - keep.length)) released.set(d.roll.seq, now);

    let cursor = showing ? showing.endAt : now;
    const scheduled: DicePlay[] = [];
    for (const d of keep) {
      const t = diceTimings(cursor, d.slotMs, { motion: d.motion, speed: opts.speed });
      // Motion off: the strip cue is raised at once (ADR 0121 §7).
      const releaseAt = d.motion ? t.settleAt : Math.min(now, t.settleAt);
      scheduled.push({ roll: d.roll, ...t, motion: d.motion, slotMs: d.slotMs, releaseAt });
      released.set(d.roll.seq, releaseAt);
      cursor = t.endAt;
    }
    plays = [...others, ...(showing ? [showing] : []), ...scheduled];
  }
  return { seen, plays, released, primed };
}

/** The plays on screen at `now`: started and not yet gone. */
export function visiblePlays(state: Pick<DiceState, "plays">, now: number): DicePlay[] {
  return state.plays.filter((p) => p.startAt <= now && now < p.endAt);
}

/** The next instant something in the schedule changes, or null when it is idle. */
export function nextDiceChange(state: Pick<DiceState, "plays">, now: number): number | null {
  let next: number | null = null;
  for (const p of state.plays) {
    for (const t of [p.startAt, p.settleAt, p.fadeAt, p.endAt, p.releaseAt]) {
      if (t > now && (next === null || t < next)) next = t;
    }
  }
  return next;
}

/** Whether `seq`'s strip cue may show at `now`. Unknown seqs are not released. */
export function isReleased(state: Pick<DiceState, "released">, seq: number, now: number): boolean {
  const at = state.released.get(seq);
  return at !== undefined && at <= now;
}

export type DicePhase = "tumble" | "hold" | "fade";

export function dicePhase(p: DicePlay, now: number): DicePhase {
  if (now < p.settleAt) return "tumble";
  if (now < p.fadeAt) return "hold";
  return "fade";
}

// ---------------------------------------------------------------- //
// The tumble: deterministic, seeded from the entry.

/** FNV-1a over a string: a stable 32-bit seed for a key. */
export function diceSeed(key: string, index = 0): number {
  let h = 0x811c9dc5;
  const s = `${key}#${index}`;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 0x01000193);
  }
  return h >>> 0;
}

/** mulberry32: a tiny seeded generator. Presentation only; never a game draw. */
function seeded(seed: number): () => number {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/**
 * tumbleFaces is the sequence of faces one die shows on its way to
 * `result`: `steps` long, the last one the result. No face repeats the
 * one before it, and (on a die with more than two faces) no face before
 * the last is the result, so the number is not given away early.
 */
export function tumbleFaces(
  seed: number,
  sides: number,
  result: number,
  steps = DICE_TUMBLE_STEPS,
): number[] {
  const rand = seeded(seed);
  const out: number[] = [];
  let prev = -1;
  for (let i = 0; i < steps - 1; i++) {
    // Choose among the faces allowed here, so every pick is valid in one
    // draw and the sequence depends on the seed alone.
    const allowed: number[] = [];
    for (let f = 1; f <= sides; f++) {
      if (f !== prev && !(sides > 2 && f === result)) allowed.push(f);
    }
    const face = allowed.length > 0 ? allowed[Math.floor(rand() * allowed.length)] : result;
    out.push(face);
    prev = face;
  }
  out.push(result);
  return out;
}

/**
 * tumbleIndex is which of `steps` faces shows `elapsed` ms into a
 * tumble of `tumbleMs`. Faces change fast as the die is thrown and slow
 * as it comes to rest (ease-out), and the last face, the result, only
 * shows once the tumble is over.
 */
export function tumbleIndex(elapsed: number, tumbleMs: number, steps = DICE_TUMBLE_STEPS): number {
  if (tumbleMs <= 0 || elapsed >= tumbleMs) return steps - 1;
  const p = Math.max(0, elapsed / tumbleMs);
  const eased = 1 - (1 - p) * (1 - p);
  return Math.min(steps - 2, Math.floor(eased * (steps - 1)));
}

/** The face a die shows at `now`. */
export function faceAt(p: DicePlay, dieIndex: number, now: number): number {
  const result = p.roll.results[dieIndex];
  if (!p.motion || now >= p.settleAt) return result;
  const faces = tumbleFaces(diceSeed(p.roll.key, dieIndex), p.roll.sides, result);
  return faces[tumbleIndex(now - p.startAt, p.settleAt - p.startAt)];
}

/** A die's throw: which way it spins and where it is thrown from. Seeded. */
export interface DiceThrow {
  /** Total spin, degrees, a whole number of turns; the sign is the direction. */
  spin: number;
  /** Where it starts, relative to where it lands, px. */
  fromX: number;
  fromY: number;
}

export function diceThrow(key: string, index: number): DiceThrow {
  const rand = seeded(diceSeed(key, 1000 + index));
  const dir = rand() < 0.5 ? -1 : 1;
  return {
    // Whole turns, so it lands upright: two or three.
    spin: dir * 360 * (rand() < 0.5 ? 2 : 3),
    fromX: Math.round(-dir * (18 + rand() * 22)),
    fromY: Math.round(-(26 + rand() * 18)),
  };
}

/**
 * coinTurns is how many half turns a coin makes before it lands on
 * `face`: heads shows on an even count, tails on an odd one. Seeded, 7
 * to 11.
 */
export function coinTurns(key: string, index: number, face: CoinFace): number {
  const rand = seeded(diceSeed(key, 2000 + index));
  let n = 7 + Math.floor(rand() * 4);
  if ((n % 2 === 0) !== (face === "heads")) n += 1;
  return n;
}

// ---------------------------------------------------------------- //
// What one entry draws.

export interface DiceShown {
  /** The dice or coins drawn: at most DICE_SHOWN_MAX. */
  count: number;
  /** How many more the entry rolled than are drawn. */
  more: number;
}

export function diceShown(r: DiceRoll): DiceShown {
  const total = r.kind === "die" ? r.results.length : r.faces.length;
  const count = Math.min(DICE_SHOWN_MAX, total);
  return { count, more: total - count };
}

/** The small label under a group of dice: "d20", "2d12", "3 coins". */
export function diceLabel(r: DiceRoll): string {
  if (r.kind === "coin") return r.faces.length === 1 ? "coin" : `${r.faces.length} coins`;
  return r.results.length === 1 ? `d${r.sides}` : `${r.results.length}d${r.sides}`;
}

// The drawn size of one entry, for placing it before it is drawn.
export const DIE_PX = 56;
export const COIN_PX = 50;
export const DICE_ROW_GAP = 6;
export const DICE_MORE_PX = 34;
/** The height of one label pill (the caption under the row, the call over it). */
export const DICE_CAPTION_PX = 18;
/** The won / lost tag under each coin of a called flip. */
export const DICE_VERDICT_PX = 16;
const DICE_PAD = 6;
const DICE_STACK_GAP = 4;

/** A label pill's width: 10px mono capitals, letter-spaced, padded. */
function labelWidth(text: string): number {
  return Math.ceil(text.length * 7.4 + 16);
}

/**
 * diceGroupSize is the box one entry is drawn in. The layer sizes the
 * group to it and places it with placeDice, so it is computed, not
 * measured: the box is known before the first frame is drawn.
 */
export function diceGroupSize(r: DiceRoll): Size {
  const { count, more } = diceShown(r);
  const unit = r.kind === "coin" ? COIN_PX : DIE_PX;
  const row =
    count * unit + (count - 1) * DICE_ROW_GAP + (more > 0 ? DICE_ROW_GAP + DICE_MORE_PX : 0);
  const labels = [diceLabel(r), ...(r.kind === "coin" ? ["heads", "tails"] : [])];
  if (r.call) labels.push(`called ${r.call}`);
  const width = Math.max(row, ...labels.map(labelWidth)) + 2 * DICE_PAD;
  const called = r.kind === "coin" && r.call !== undefined;
  const height =
    2 * DICE_PAD +
    (r.call ? DICE_CAPTION_PX + DICE_STACK_GAP : 0) +
    unit +
    (called ? DICE_VERDICT_PX + 2 : 0) +
    DICE_STACK_GAP +
    DICE_CAPTION_PX;
  return { width, height };
}

/** Whether coin `i` of a called flip won. Undefined for an uncalled flip. */
export function coinWon(r: DiceRoll, i: number): boolean | undefined {
  if (!r.call) return undefined;
  return r.faces[i] === r.call;
}

// ---------------------------------------------------------------- //
// Where it is drawn.

export interface Rect {
  left: number;
  top: number;
  width: number;
  height: number;
}

export interface Size {
  width: number;
  height: number;
}

/** The space between the avatar (or the strip) and the dice. */
export const DICE_GAP = 10;
/** The dice never come closer to the board's edge than this. */
export const DICE_EDGE = 6;

export type DiceSide = "left" | "right" | "above" | "below";

export interface DicePlacement {
  left: number;
  top: number;
  /** Which side of the anchor it sits on; "strip" for the fallback. */
  side: DiceSide | "strip";
}

function clamp(v: number, lo: number, hi: number): number {
  return Math.max(lo, Math.min(hi, v));
}

/**
 * placeDice puts a group of `size` beside `avatar` (board pixels), on
 * the side of the avatar toward the board's centre, so it reads as the
 * roller's and four seats never overlap. With no avatar on screen it
 * goes under the attention strip's content (`strip`) instead. Always
 * clamped inside the board.
 */
export function placeDice(
  board: Size,
  avatar: Rect | null,
  size: Size,
  strip: Rect | null,
): DicePlacement {
  const maxLeft = Math.max(DICE_EDGE, board.width - size.width - DICE_EDGE);
  const maxTop = Math.max(DICE_EDGE, board.height - size.height - DICE_EDGE);
  if (!avatar) {
    const s = strip ?? { left: DICE_EDGE, top: DICE_EDGE, width: size.width, height: 0 };
    return {
      left: clamp(s.left, DICE_EDGE, maxLeft),
      top: clamp(s.top + s.height + DICE_GAP, DICE_EDGE, maxTop),
      side: "strip",
    };
  }
  const cx = avatar.left + avatar.width / 2;
  const cy = avatar.top + avatar.height / 2;
  const dx = board.width / 2 - cx;
  const dy = board.height / 2 - cy;
  // Compare in board proportions, so a wide board does not always win
  // the horizontal side.
  const horizontal =
    Math.abs(dx) / Math.max(1, board.width) >= Math.abs(dy) / Math.max(1, board.height);
  let left: number;
  let top: number;
  let side: DiceSide;
  if (horizontal) {
    side = dx >= 0 ? "right" : "left";
    left =
      side === "right"
        ? avatar.left + avatar.width + DICE_GAP
        : avatar.left - DICE_GAP - size.width;
    top = cy - size.height / 2;
  } else {
    side = dy >= 0 ? "below" : "above";
    left = cx - size.width / 2;
    top =
      side === "below"
        ? avatar.top + avatar.height + DICE_GAP
        : avatar.top - DICE_GAP - size.height;
  }
  return { left: clamp(left, DICE_EDGE, maxLeft), top: clamp(top, DICE_EDGE, maxTop), side };
}
