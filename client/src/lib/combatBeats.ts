// combatBeats turns the public game log into the two visible moments
// of combat damage — the first-strike step and the regular step — so
// the board can cue them one after the other instead of in a single
// jump cut (#187, ADR 0053).
//
// The server already says which step dealt each combat damage entry:
// `LogEvent.combat_step`, set only when that combat had a first-strike
// step (ADR 0053 Decision 1). Nothing here derives it from keywords.
// READ IT, DON'T DERIVE IT.
//
// The rules, all from ADR 0053 "Beats across frames":
//
//   - an entry is handled once, keyed by `seq`. The first frame a
//     client sees PRIMES: everything already on the wire is marked
//     handled and nothing is cued (the S22 reveals.ts pattern). The
//     shell also asks for a re-prime after a reconnect or a replay
//     toggle, when the board stays mounted but the frames in between
//     were never watched live.
//   - a frame whose highest `seq` is lower than the highest seen is a
//     rewind (undo): the handled watermark drops to it and the frame
//     cues nothing, so the replayed entries are cued when they return.
//   - a `combat_damage` step is keyed by the `seq` of its `step`
//     entry. A beat is that step's tagged combat damage entries for one
//     tag, plus the non-damage entries that follow the nearest tagged
//     damage entry among the same frame's new entries (deaths,
//     lifelink). A step that dealt no damage for a tag has no beat
//     (Decision 3).
//   - beat 2 waits BEAT_PAUSE_MS × animations.speed only when beat 1's
//     entries came before it in the SAME frame. Across frames the time
//     between frames is the pause. A beat already cued in an earlier
//     frame continues with no pause and no repeated label.
//   - nothing a later frame carries cancels a scheduled cue: leaving
//     the step mid-sequence finishes the cues (BeatSequencer never
//     clears a timer except on dispose).
//
// Plain functions over plain data, plus one small timer class, so all
// of it is unit-testable without mounting a component (#689).
// CombatArrows.svelte is the thin shell that measures, draws and
// renders the text cue.

import type { LogEvent } from "./protocol";

export type BeatTag = "first_strike" | "regular";

// "full" pulses and ghosts arrows; "still" shows only the text cue.
export type BeatMode = "full" | "still";

// BEAT_PAUSE_MS is the gap between beat 1 and beat 2 when both land in
// one frame, before the animations.speed multiplier (owner, ADR 0053
// Decision 2 "Timing").
export const BEAT_PAUSE_MS = 400;

// BEAT_EFFECT_MS is one arrow pulse or ghost fade, in and out, before
// the speed multiplier. It must stay shorter than BEAT_PAUSE_MS, so a
// beat's arrow effect is over before the next beat's starts; the clamp
// makes raising it past the pause a no-op rather than an overlap (and
// a unit test pins the margin).
export const BEAT_EFFECT_MS = Math.min(360, BEAT_PAUSE_MS - 40);

// BEAT_CUE_HOLD_MS is how long the text cue stays up after the last
// beat of a frame has been cued. Not speed-scaled on purpose: it is
// text to read, not motion, and speed 0.5 must not make it unreadable.
export const BEAT_CUE_HOLD_MS = 1800;

export const BEAT_LABELS: Record<BeatTag, string> = {
  first_strike: "First strike",
  regular: "Regular damage",
};

// The number of combat_damage steps whose cued beats are remembered.
// Only the current step can gain entries in practice; a few are kept
// so an extra combat never forgets the one before it mid-sequence.
const CUED_STEPS_KEPT = 8;

// beatMode is "full" only when every setting allows motion: the
// animations master switch, the damagePopups per-effect toggle that
// beats reuse, and accessibility.reduceMotion, which the settings
// bridge never pushes into animations.ts — so the shell passes all
// three from the settings store (ADR 0053 Decision 2).
export function beatMode(
  animationsEnabled: boolean,
  damagePopups: boolean,
  reduceMotion: boolean,
): BeatMode {
  return animationsEnabled && damagePopups && !reduceMotion ? "full" : "still";
}

// scaledMs applies animations.speed (2 = twice as slow, 0.5 = twice as
// fast). Deliberately not animations.ts gatedDuration, which snaps to
// 1 ms when animations or the per-effect flag are off and would erase
// the "still" mode hold.
export function scaledMs(baseMs: number, speed: number): number {
  const s = Number.isFinite(speed) && speed > 0 ? speed : 1;
  return baseMs * s;
}

// ---- Tracking: which entries are new ----

export interface BeatTracker {
  primed: boolean;
  // Highest log seq handled. Log entries are appended in seq order, so
  // a watermark is the same as a seen set here, and a rewind is just
  // lowering it.
  maxSeq: number;
  // Per combat_damage step (keyed by its step entry's seq): each beat
  // already cued, and the seq of the damage entry that first cued it.
  cued: Map<number, Map<BeatTag, number>>;
}

export function emptyBeatTracker(): BeatTracker {
  return { primed: false, maxSeq: 0, cued: new Map() };
}

export interface TrackResult {
  tracker: BeatTracker;
  // The frame's entries this client has not handled, in log order.
  fresh: LogEvent[];
  // True when this frame primed (first frame, or a requested re-prime).
  primed: boolean;
  // True when this frame was a rewind (undo).
  rewound: boolean;
}

function highestSeq(log: readonly LogEvent[] | undefined): number {
  let top = 0;
  for (const e of log ?? []) if (e.seq > top) top = e.seq;
  return top;
}

// track folds one frame's log into the tracker. Returns a new tracker;
// the input is not mutated.
export function track(
  prev: BeatTracker,
  log: readonly LogEvent[] | undefined,
  reprime = false,
): TrackResult {
  const top = highestSeq(log);
  if (!prev.primed || reprime) {
    return {
      tracker: { primed: true, maxSeq: top, cued: new Map() },
      fresh: [],
      primed: true,
      rewound: false,
    };
  }
  if (top < prev.maxSeq) {
    // Undo restored the engine's event counter. Forget everything above
    // the frame's highest seq, including a beat whose first cued entry
    // is gone, so the replay is cued as new.
    const cued = new Map<number, Map<BeatTag, number>>();
    for (const [stepSeq, beats] of prev.cued) {
      if (stepSeq > top) continue;
      const kept = new Map([...beats].filter(([, first]) => first <= top));
      if (kept.size > 0) cued.set(stepSeq, kept);
    }
    return {
      tracker: { primed: true, maxSeq: top, cued },
      fresh: [],
      primed: false,
      rewound: true,
    };
  }
  const fresh = (log ?? []).filter((e) => e.seq > prev.maxSeq);
  return {
    tracker: { primed: true, maxSeq: top, cued: prev.cued },
    fresh,
    primed: false,
    rewound: false,
  };
}

// ---- Beats: grouping new entries ----

export interface Beat {
  // seq of the combat_damage step entry this beat belongs to.
  stepSeq: number;
  tag: BeatTag;
  // Every member, in log order: the tagged damage entries and the
  // entries that followed them in this frame.
  entries: LogEvent[];
  // The tagged combat damage entries only.
  damage: LogEvent[];
  // The combat arrows the damage names (arrowRefsFor), and their IDs.
  arrows: ArrowRef[];
  arrowIDs: string[];
  // This beat was already cued in an earlier frame of its step: no
  // label, no pause.
  continuation: boolean;
  // Beat 2 whose beat-1 entries came before it in this same frame.
  pauseBefore: boolean;
}

function combatTagOf(e: LogEvent): BeatTag | null {
  if (e.kind !== "damage" || !e.combat) return null;
  return e.combat_step === "first_strike" || e.combat_step === "regular" ? e.combat_step : null;
}

// splitBeats groups a frame's fresh entries into beats, per
// combat_damage step, and records the beats it cues on the tracker.
// `log` is the whole frame log: a fresh entry's step entry can be
// older than the entry.
export function splitBeats(
  prev: BeatTracker,
  log: readonly LogEvent[] | undefined,
  fresh: readonly LogEvent[],
): { tracker: BeatTracker; beats: Beat[] } {
  if (fresh.length === 0) return { tracker: prev, beats: [] };
  const freshSeqs = new Set(fresh.map((e) => e.seq));
  const beats: Beat[] = [];
  const byKey = new Map<string, Beat>();
  let step: { seq: number; name: string } | null = null;
  let current: Beat | null = null;

  for (const e of log ?? []) {
    if (e.kind === "step") {
      step = { seq: e.seq, name: e.step ?? "" };
      current = null;
      continue;
    }
    if (!freshSeqs.has(e.seq)) continue;
    // Entries whose step entry fell out of the 200-entry window, or
    // that are outside combat damage, belong to no beat.
    if (step === null || step.name !== "combat_damage") continue;
    const tag = combatTagOf(e);
    if (tag === null) {
      current?.entries.push(e);
      continue;
    }
    const key = `${step.seq}:${tag}`;
    let beat = byKey.get(key);
    if (!beat) {
      beat = {
        stepSeq: step.seq,
        tag,
        entries: [],
        damage: [],
        arrows: [],
        arrowIDs: [],
        continuation: prev.cued.get(step.seq)?.has(tag) ?? false,
        pauseBefore: false,
      };
      byKey.set(key, beat);
      beats.push(beat);
    }
    beat.entries.push(e);
    beat.damage.push(e);
    current = beat;
  }

  if (beats.length === 0) return { tracker: prev, beats };

  const cued = new Map<number, Map<BeatTag, number>>();
  for (const [stepSeq, m] of prev.cued) cued.set(stepSeq, new Map(m));
  for (const beat of beats) {
    beat.arrows = arrowRefsFor(beat.damage, log);
    beat.arrowIDs = beat.arrows.map((a) => a.id);
    if (beat.tag === "regular" && !beat.continuation) {
      const first = byKey.get(`${beat.stepSeq}:first_strike`);
      // Only beat 1 THEN beat 2 in this frame earns the pause. A
      // first-strike entry after regular damage started (#702's order
      // today) is shown at once, not reordered.
      beat.pauseBefore = first !== undefined && first.damage[0].seq < beat.damage[0].seq;
    }
    if (!beat.continuation) {
      let m = cued.get(beat.stepSeq);
      if (!m) cued.set(beat.stepSeq, (m = new Map()));
      m.set(beat.tag, beat.damage[0].seq);
    }
  }
  while (cued.size > CUED_STEPS_KEPT) {
    cued.delete(Math.min(...cued.keys()));
  }
  return { tracker: { ...prev, cued }, beats };
}

// ArrowRef is one combat arrow a damage entry travelled along, with
// the endpoints CombatArrows draws it between: an attack runs from the
// attacker to the defending seat's header, a block from the blocker to
// the attacker. Everything here comes from the log, so it is known even
// when the arrow was never drawn.
export interface ArrowRef {
  id: string;
  kind: "attack" | "block";
  fromCardID: string;
  // Card endpoint of a block arrow.
  toCardID?: string;
  // Seat index of an attack arrow's defending player.
  toSeat?: number;
}

// arrowRefsFor names the CombatArrows arrow each damage entry travelled
// along, from the log alone (ADR 0053 Decision 4). Roles come from
// the same turn's `block` entries, the latest per blocker — never from
// live battlefield state, which has already lost the creatures that
// died. An entry that names no arrow (an attack on a planeswalker or
// battle, or a block entry that fell out of the log window) adds
// nothing; the text cue still carries it.
export function arrowRefsFor(
  damage: readonly LogEvent[],
  log: readonly LogEvent[] | undefined,
): ArrowRef[] {
  const out: ArrowRef[] = [];
  const add = (ref: ArrowRef) => {
    if (!out.some((r) => r.id === ref.id)) out.push(ref);
  };
  const blocksByTurn = new Map<number, Map<string, string>>();
  const blocksOn = (turn: number): Map<string, string> => {
    let m = blocksByTurn.get(turn);
    if (m) return m;
    m = new Map();
    for (const e of log ?? []) {
      if (e.kind === "block" && (e.turn ?? 0) === turn && e.card_id && e.target) {
        m.set(e.card_id, e.target); // blocker → attacker; later wins
      }
    }
    blocksByTurn.set(turn, m);
    return m;
  };

  for (const d of damage) {
    const source = d.card_id;
    if (!source) continue;
    if (d.target_seat !== undefined && d.target_seat !== null) {
      add({ id: `atk-${source}`, kind: "attack", fromCardID: source, toSeat: d.target_seat });
      continue;
    }
    const target = d.target;
    if (!target) continue;
    const blocks = blocksOn(d.turn ?? 0);
    if (blocks.get(target) === source) {
      add({ id: `blk-${target}`, kind: "block", fromCardID: target, toCardID: source });
    } else if (blocks.get(source) === target) {
      add({ id: `blk-${source}`, kind: "block", fromCardID: source, toCardID: target });
    }
  }
  return out;
}

// arrowIDsFor is arrowRefsFor's IDs.
export function arrowIDsFor(
  damage: readonly LogEvent[],
  log: readonly LogEvent[] | undefined,
): string[] {
  return arrowRefsFor(damage, log).map((r) => r.id);
}

// ---- Schedule ----

export interface ScheduledCue {
  // When to cue, in ms after the frame arrived.
  atMs: number;
  stepSeq: number;
  tag: BeatTag;
  // "First strike" / "Regular damage", or null for a continuation of a
  // beat whose label was already shown.
  label: string | null;
  entries: LogEvent[];
  arrows: ArrowRef[];
  arrowIDs: string[];
  // True only in "full" mode: pulse live arrows, draw ghosts. A "still"
  // schedule has no motion at all, only the text cue.
  motion: boolean;
  // Duration of one pulse or ghost fade, speed-scaled.
  effectMs: number;
  // When this frame's text cue for the step comes down, in ms after
  // the frame arrived. The same for every cue of one step in a frame,
  // so beat 1's label is still up when beat 2's joins it.
  hideAtMs: number;
}

// schedule places a frame's beats in time. Beat 2 after beat 1 in the
// same frame waits one scaled pause, in both modes: "still" holds beat
// 1's text cue alone for the pause instead of collapsing the two.
export function schedule(beats: readonly Beat[], mode: BeatMode, speed: number): ScheduledCue[] {
  const pause = scaledMs(BEAT_PAUSE_MS, speed);
  const effectMs = scaledMs(BEAT_EFFECT_MS, speed);
  const cues: ScheduledCue[] = beats.map((b) => ({
    atMs: b.pauseBefore ? pause : 0,
    stepSeq: b.stepSeq,
    tag: b.tag,
    label: b.continuation ? null : BEAT_LABELS[b.tag],
    entries: b.entries,
    arrows: b.arrows,
    arrowIDs: b.arrowIDs,
    motion: mode === "full",
    effectMs,
    hideAtMs: 0,
  }));
  const lastAt = new Map<number, number>();
  for (const c of cues) lastAt.set(c.stepSeq, Math.max(lastAt.get(c.stepSeq) ?? 0, c.atMs));
  for (const c of cues) c.hideAtMs = (lastAt.get(c.stepSeq) ?? 0) + BEAT_CUE_HOLD_MS;
  return cues;
}

export interface FramePlan {
  tracker: BeatTracker;
  cues: ScheduledCue[];
  primed: boolean;
}

// planFrame is the whole per-frame pipeline: track, split, schedule.
export function planFrame(
  prev: BeatTracker,
  log: readonly LogEvent[] | undefined,
  opts: { mode: BeatMode; speed: number; reprime?: boolean },
): FramePlan {
  const t = track(prev, log, opts.reprime ?? false);
  const split = splitBeats(t.tracker, log, t.fresh);
  return {
    tracker: split.tracker,
    cues: schedule(split.beats, opts.mode, opts.speed),
    primed: t.primed,
  };
}

// cueSummary is the short count a screen reader hears after a cue's
// label ("First strike, 2 hits"). Deliberately not the log lines: the
// cue region announces on every beat of every combat, and the log panel
// already carries the full sentences.
export function cueSummary(cue: Pick<ScheduledCue, "entries" | "tag">): string {
  const n = cue.entries.filter((e) => combatTagOf(e) === cue.tag).length;
  return n === 1 ? "1 hit" : `${n} hits`;
}

// ---- Arrows: live, ghost or nothing ----

export type ArrowRender = "live" | "ghost" | "none";

// arrowRender picks how a beat shows one arrow: pulse it if both its
// endpoints are mounted, draw a ghost from cached geometry if not,
// and nothing if there is no geometry either.
export function arrowRender(
  arrowID: string,
  liveIDs: { has(id: string): boolean },
  cachedIDs: { has(id: string): boolean },
): ArrowRender {
  if (liveIDs.has(arrowID)) return "live";
  if (cachedIDs.has(arrowID)) return "ghost";
  return "none";
}

export interface Point {
  x: number;
  y: number;
}

export interface BoardSize {
  w: number;
  h: number;
}

// ArrowGeometry is a quadratic bezier: from (x1,y1) to (x2,y2) with
// control point (cx,cy). Board-relative pixels.
export interface ArrowGeometry {
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  cx: number;
  cy: number;
}

// CachedArrow is an attack or block arrow's last measured geometry,
// with the endpoints it was measured from and the board size at the
// time.
export interface CachedArrow {
  kind: "attack" | "block";
  fromCardID: string;
  toSeatID?: string;
  toCardID?: string;
  geo: ArrowGeometry;
  board: BoardSize;
}

// midpointOffset is the arc's control point: bowed perpendicular to
// the chord by ~18% of its length, always toward the top of the board
// so attack arcs don't dive through the hand fan.
export function midpointOffset(
  x1: number,
  y1: number,
  x2: number,
  y2: number,
): { cx: number; cy: number } {
  const mx = (x1 + x2) / 2;
  const my = (y1 + y2) / 2;
  const dx = x2 - x1;
  const dy = y2 - y1;
  const len = Math.hypot(dx, dy) || 1;
  const nx = -dy / len;
  const ny = dx / len;
  const bow = len * 0.18;
  const candA = { cx: mx + nx * bow, cy: my + ny * bow };
  const candB = { cx: mx - nx * bow, cy: my - ny * bow };
  return candA.cy < candB.cy ? candA : candB;
}

export function arrowGeometry(from: Point, to: Point): ArrowGeometry {
  const ctrl = midpointOffset(from.x, from.y, to.x, to.y);
  return { x1: from.x, y1: from.y, x2: to.x, y2: to.y, cx: ctrl.cx, cy: ctrl.cy };
}

// BOARD_RESIZE_TOLERANCE_PX absorbs sub-pixel layout noise.
const BOARD_RESIZE_TOLERANCE_PX = 1;

// ghostGeometry is where a ghost arrow is drawn: cached coordinates
// only for an endpoint that is gone, a fresh measurement for one still
// on the board. Null when the board has changed size since the cache
// was written — a cached point would then be a guess, so the ghost is
// dropped and the text cue carries the beat.
export function ghostGeometry(
  cached: CachedArrow,
  boardNow: BoardSize,
  fromNow: Point | null,
  toNow: Point | null,
): ArrowGeometry | null {
  if (!sameBoard(boardNow, cached.board)) return null;
  const from = fromNow ?? { x: cached.geo.x1, y: cached.geo.y1 };
  const to = toNow ?? { x: cached.geo.x2, y: cached.geo.y2 };
  return arrowGeometry(from, to);
}

// CachedPoint is one card tile's last measured board-relative centre,
// with the board size at the time.
export interface CachedPoint {
  at: Point;
  board: BoardSize;
}

function sameBoard(a: BoardSize, b: BoardSize): boolean {
  return (
    Math.abs(a.w - b.w) <= BOARD_RESIZE_TOLERANCE_PX &&
    Math.abs(a.h - b.h) <= BOARD_RESIZE_TOLERANCE_PX
  );
}

export interface ArrowInputs {
  // Arrows drawn right now, by ID: both endpoints mounted.
  live: { get(id: string): ArrowGeometry | undefined };
  // Last measured arrow geometry, by arrow ID.
  arrowCache: { get(id: string): CachedArrow | undefined };
  // Last measured card tile centres, by instance ID. Tiles are measured
  // on every frame they are on the board, long before combat, so this
  // knows both endpoints of an arrow that was never drawn.
  cardCache: { get(id: string): CachedPoint | undefined };
  board: BoardSize;
  // Fresh measurements of the ref's endpoints, or null for one that is
  // not on the board now.
  fromNow: Point | null;
  toNow: Point | null;
}

// resolveArrow decides how a beat shows one arrow, and where:
//
//   1. live — the arrow is drawn right now: pulse it.
//   2. ghost from the arrow cache — it was drawn earlier (ghostGeometry).
//   3. ghost from card tiles — no usable arrow geometry. Typically the
//      arrow was never drawn, because the frame
//      that declared it and the frame that dealt the damage landed
//      inside one animation frame (a bot, or a fast network), but both
//      creatures' tiles were measured earlier. Each endpoint is a fresh
//      measurement if it is still on the board, else its cached centre.
//   4. none — the text cue carries the beat.
//
// A cached point measured on a board of another size is never used
// (the reflow rule): the ghost is dropped instead.
export function resolveArrow(
  ref: ArrowRef,
  inputs: ArrowInputs,
): { render: ArrowRender; geo: ArrowGeometry | null } {
  const liveGeo = inputs.live.get(ref.id);
  const cachedArrow = inputs.arrowCache.get(ref.id);
  const cachedIDs = { has: () => cachedArrow !== undefined };
  const render = arrowRender(ref.id, { has: () => liveGeo !== undefined }, cachedIDs);
  if (render === "live") return { render, geo: liveGeo! };
  if (render === "ghost") {
    const geo = ghostGeometry(cachedArrow!, inputs.board, inputs.fromNow, inputs.toNow);
    if (geo) return { render, geo };
    // The board changed size since the arrow was measured. A tile
    // re-measured after the resize can still place it.
  }
  const cachedPoint = (id: string | undefined): Point | null => {
    if (!id) return null;
    const c = inputs.cardCache.get(id);
    return c && sameBoard(c.board, inputs.board) ? c.at : null;
  };
  const from = inputs.fromNow ?? cachedPoint(ref.fromCardID);
  // A seat header is never cached: it is always mounted, so toNow is it.
  const to = inputs.toNow ?? (ref.kind === "block" ? cachedPoint(ref.toCardID) : null);
  if (!from || !to) return { render: "none", geo: null };
  return { render: "ghost", geo: arrowGeometry(from, to) };
}

// pruneCardCache drops the tiles of cards no longer on the battlefield,
// unless combat or a pending cue may still need them (keepArrowCache).
// Cards still on the battlefield are kept: they are re-measured anyway.
export function pruneCardCache(
  cache: Map<string, CachedPoint>,
  onBattlefield: { has(id: string): boolean },
  keep: boolean,
): void {
  if (keep) return;
  for (const id of [...cache.keys()]) {
    if (!onBattlefield.has(id)) cache.delete(id);
  }
}

// The text cue keeps this far from the board's edges.
const CUE_EDGE_MARGIN_PX = 48;

// cueAnchor is where a step's text cue sits: the average of its
// arrows' curve midpoints, so the label is next to the damage it
// describes, or the middle of the board when no arrow could be drawn.
export function cueAnchor(geos: readonly ArrowGeometry[], board: BoardSize): Point {
  let x = board.w / 2;
  let y = board.h / 2;
  if (geos.length > 0) {
    x = 0;
    y = 0;
    for (const g of geos) {
      // Quadratic bezier at t = 0.5.
      x += 0.25 * g.x1 + 0.5 * g.cx + 0.25 * g.x2;
      y += 0.25 * g.y1 + 0.5 * g.cy + 0.25 * g.y2;
    }
    x /= geos.length;
    y /= geos.length;
  }
  const clamp = (v: number, max: number) =>
    max <= 2 * CUE_EDGE_MARGIN_PX
      ? max / 2
      : Math.min(max - CUE_EDGE_MARGIN_PX, Math.max(CUE_EDGE_MARGIN_PX, v));
  return { x: clamp(x, board.w), y: clamp(y, board.h) };
}

// The steps during which attack and block arrows exist, so their
// geometry is worth keeping.
const COMBAT_ARROW_STEPS = new Set(["declare_attackers", "declare_blockers", "combat_damage"]);

// keepArrowCache: the geometry cache lives through combat and until
// the last scheduled cue has played, even after the game has moved on
// to end_combat. Outside that it is dropped.
export function keepArrowCache(step: string | undefined, pendingCues: number): boolean {
  return pendingCues > 0 || COMBAT_ARROW_STEPS.has(step ?? "");
}

// ---- Runtime: timers ----

export interface BeatHandlers {
  onCue(cue: ScheduledCue): void;
  // The text cue for this step comes down.
  onHide(stepSeq: number): void;
}

// BeatSequencer plays schedules on real timers. A later frame only
// ever ADDS cues: nothing but dispose() cancels one, so a frame that
// leaves combat_damage mid-sequence does not cut the remaining cues
// (ADR 0053, owner 2026-09-16). A later labelled cue for the same step
// pushes that step's hide time out; an earlier one never pulls it in.
export class BeatSequencer {
  private readonly cueTimers = new Set<ReturnType<typeof setTimeout>>();
  private readonly hideTimers = new Map<
    number,
    { handle: ReturnType<typeof setTimeout>; at: number }
  >();

  constructor(private readonly handlers: BeatHandlers) {}

  play(cues: readonly ScheduledCue[]): void {
    const start = Date.now();
    for (const cue of cues) {
      const handle = setTimeout(() => {
        this.cueTimers.delete(handle);
        this.handlers.onCue(cue);
      }, cue.atMs);
      this.cueTimers.add(handle);

      if (cue.label === null) continue;
      const at = start + cue.hideAtMs;
      const prev = this.hideTimers.get(cue.stepSeq);
      if (prev && prev.at >= at) continue;
      if (prev) clearTimeout(prev.handle);
      const hide = setTimeout(() => {
        this.hideTimers.delete(cue.stepSeq);
        this.handlers.onHide(cue.stepSeq);
      }, cue.hideAtMs);
      this.hideTimers.set(cue.stepSeq, { handle: hide, at });
    }
  }

  // pending is the number of cues scheduled but not yet played.
  get pending(): number {
    return this.cueTimers.size;
  }

  dispose(): void {
    for (const h of this.cueTimers) clearTimeout(h);
    for (const { handle } of this.hideTimers.values()) clearTimeout(handle);
    this.cueTimers.clear();
    this.hideTimers.clear();
  }
}

// BeatDirector is the per-component beat runtime: it folds each frame
// into the tracker and plays the frame's cues. A normal frame only ever
// adds cues. A frame that PRIMES (a re-prime after a reconnect or a
// replay scrubber jump) cancels every pending cue and hide first, by
// swapping in a fresh sequencer, and tells the shell to clear what is on
// screen: those cues belong to frames this client is no longer showing,
// and firing them against the new frame would flash a stale label.
export class BeatDirector {
  private tracker = emptyBeatTracker();
  private sequencer: BeatSequencer;

  constructor(private readonly handlers: BeatHandlers & { onReset(): void }) {
    this.sequencer = new BeatSequencer(handlers);
  }

  frame(
    log: readonly LogEvent[] | undefined,
    opts: { mode: BeatMode; speed: number; reprime?: boolean },
  ): FramePlan {
    const plan = planFrame(this.tracker, log, opts);
    this.tracker = plan.tracker;
    if (plan.primed) {
      this.sequencer.dispose();
      this.sequencer = new BeatSequencer(this.handlers);
      this.handlers.onReset();
    }
    this.sequencer.play(plan.cues);
    return plan;
  }

  get pending(): number {
    return this.sequencer.pending;
  }

  dispose(): void {
    this.sequencer.dispose();
  }
}
