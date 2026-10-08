// combatStrikes — what moves when combat damage lands (ADR 0134).
//
// A beat (combatBeats.ts) says which combat damage entries landed
// together. This module turns one beat into a plan: which creatures
// lunge, at what, which targets shake at contact, and which of them
// died. CombatStrikes.svelte measures and draws the plan; every rule is
// here, as plain functions over plain data, so it is unit-tested
// without mounting anything.
//
// The rules (ADR 0134 §3), each from the Comprehensive Rules:
//
//   - one beat is one moment (CR 510.2): every lunge in it starts
//     together, and every target shakes at the same contact;
//   - an attacker blocked by several creatures lunges ONCE, at the
//     centroid of the blockers it damaged, because CR 510.1c divides
//     its damage among them in no order;
//   - trample's excess to a player, planeswalker or battle (CR 702.19b)
//     is dealt in the same moment, so that target shakes at the same
//     contact and is not aimed at;
//   - a blocker lunges only in a beat where the attacker it blocks dealt
//     it no damage (a first-strike blocker, or a 0-power attacker);
//     otherwise it braces, shaking at the attacker's contact, while the
//     attacker's copy flashes;
//   - a creature that died plays its death after the hit (CR 704.3);
//   - a source is on the attacking side when its entry's `seat` is the
//     seat on its damage step's `step` entry: the active player is the
//     attacking player (CR 506.2). The `block` entries are the fallback
//     when the step entry has fallen out of the 200-entry window.

import { combatKeyer, type BoardSize, type Point } from "./combatBeats";
import type { LogEvent } from "./protocol";
import {
  IMPACT_SHAKE_MS,
  STRIKE_BACK_MS,
  STRIKE_DEATH_FADE_MS,
  STRIKE_HOLD_MS,
  STRIKE_MS,
  STRIKE_OUT_MS,
} from "./animations";

// At most this many copies fly in one beat (ADR 0134 §3, "Many
// attackers"). Past it, the remaining movers stay put, and what they hit
// still shakes.
export const STRIKE_COPY_CAP = 12;

// The copy travels to contact plus this overlap, so the two visibly
// touch (ADR 0134 §2).
export const STRIKE_CONTACT_OVERLAP_PX = 8;

// The travel is clamped to this share of the centre-to-centre distance.
export const LUNGE_MIN_FRACTION = 0.5;
export const LUNGE_MAX_FRACTION = 0.85;

// A strike that would start more than this late is dropped (ADR 0134 §4).
export const STRIKE_LATE_MS = 1000;

export type StrikeTarget = { kind: "card"; cardID: string } | { kind: "seat"; seat: number };

export interface StrikeMover {
  cardID: string;
  side: "attacker" | "blocker";
  // The copy lunges at the centroid of these.
  aim: StrikeTarget[];
  // The creature died in this beat: its copy fades on the way back.
  dies: boolean;
  // Something hit it in this beat: its copy flashes at contact instead
  // of its tile shaking.
  flash: boolean;
}

export interface StrikeImpact {
  target: StrikeTarget;
  // A card target that died in this beat: it shakes, then fades.
  dies: boolean;
}

export interface StrikePlan {
  movers: StrikeMover[];
  impacts: StrikeImpact[];
}

function targetOf(e: LogEvent): StrikeTarget | null {
  if (e.target_seat !== undefined && e.target_seat !== null) {
    return { kind: "seat", seat: e.target_seat };
  }
  return e.target ? { kind: "card", cardID: e.target } : null;
}

function targetKey(t: StrikeTarget): string {
  return t.kind === "card" ? `c:${t.cardID}` : `s:${t.seat}`;
}

// stepSeatOf returns a function giving the seat on the `step` entry an
// entry happened under: the active player (CR 506.2). Null when no step
// entry precedes it in the window.
function stepSeatOf(log: readonly LogEvent[]): (e: LogEvent) => number | null {
  const steps = log.filter((e) => e.kind === "step");
  return (e) => {
    for (let i = steps.length - 1; i >= 0; i--) {
      if (steps[i].seq < e.seq) return steps[i].seat;
    }
    return null;
  };
}

// strikesFor is the plan for one beat: `cue.entries` are the beat's
// members (its combat damage entries and the deaths and other entries
// that followed them), `log` is the frame's whole log.
export function strikesFor(
  cue: { entries: readonly LogEvent[] },
  log: readonly LogEvent[] | undefined,
): StrikePlan {
  const window = log ?? [];
  const damage = cue.entries.filter((e) => e.kind === "damage" && e.combat && e.card_id);
  if (damage.length === 0) return { movers: [], impacts: [] };

  // A creature died in this beat when the beat's own entries move it off
  // the battlefield (ADR 0053's beat membership already holds the SBA
  // deaths that followed the damage).
  const died = new Set<string>();
  for (const e of cue.entries) {
    if (e.kind === "zone" && e.old_zone === "battlefield" && e.card_id) died.add(e.card_id);
  }

  // Who blocked whom, from the same combat's block entries (later wins).
  const combatOf = combatKeyer(window);
  const combat = combatOf(damage[0]);
  const blockerOf = new Map<string, string>(); // blocker → attacker
  for (const e of window) {
    if (e.kind === "block" && e.card_id && e.target && combatOf(e) === combat) {
      blockerOf.set(e.card_id, e.target);
    }
  }
  const stepSeat = stepSeatOf(window);
  const sideOf = (e: LogEvent): "attacker" | "blocker" => {
    const active = stepSeat(e);
    if (active !== null && active >= 0 && e.seat >= 0) {
      return e.seat === active ? "attacker" : "blocker";
    }
    return blockerOf.has(e.card_id!) ? "blocker" : "attacker";
  };

  // Each source's targets, in log order, and who hit each card.
  interface Source {
    side: "attacker" | "blocker";
    targets: StrikeTarget[];
  }
  const sources = new Map<string, Source>();
  const hitBy = new Map<string, Set<string>>(); // target card → sources
  const targets: StrikeTarget[] = [];
  const seenTargets = new Set<string>();
  for (const e of damage) {
    const t = targetOf(e);
    if (!t) continue;
    const src = e.card_id!;
    let s = sources.get(src);
    if (!s) sources.set(src, (s = { side: sideOf(e), targets: [] }));
    const k = targetKey(t);
    if (!s.targets.some((x) => targetKey(x) === k)) s.targets.push(t);
    if (!seenTargets.has(k)) {
      seenTargets.add(k);
      targets.push(t);
    }
    if (t.kind === "card") {
      let by = hitBy.get(t.cardID);
      if (!by) hitBy.set(t.cardID, (by = new Set()));
      by.add(src);
    }
  }

  const candidates: StrikeMover[] = [];
  for (const [cardID, s] of sources) {
    const flash = (hitBy.get(cardID)?.size ?? 0) > 0;
    const dies = died.has(cardID);
    if (s.side === "attacker") {
      const cards = s.targets.filter((t) => t.kind === "card");
      // Its own blockers, when the block entries say which those are;
      // anything else it hit was trample's excess and only shakes.
      const blockers = cards.filter((t) => blockerOf.get(t.cardID) === cardID);
      const seats = s.targets.filter((t) => t.kind === "seat");
      const aim = blockers.length > 0 ? blockers : cards.length > 0 ? cards : seats;
      if (aim.length > 0) candidates.push({ cardID, side: "attacker", aim, dies, flash });
      continue;
    }
    // A blocker braces when what it hit hit it in this beat.
    const braces = s.targets.some((t) => t.kind === "card" && hitBy.get(cardID)?.has(t.cardID));
    if (braces || s.targets.length === 0) continue;
    candidates.push({ cardID, side: "blocker", aim: s.targets, dies, flash });
  }

  const movers = candidates.slice(0, STRIKE_COPY_CAP);
  const moving = new Set(movers.map((m) => m.cardID));
  const impacts: StrikeImpact[] = [];
  for (const t of targets) {
    // A moving creature that was hit flashes on its copy instead.
    if (t.kind === "card" && moving.has(t.cardID)) continue;
    impacts.push({ target: t, dies: t.kind === "card" && died.has(t.cardID) });
  }
  return { movers, impacts };
}

// ---- Geometry ----

// A box by its centre and its size on screen (a tapped tile's on-screen
// size is its rotated bounding box).
export interface StrikeBox {
  x: number;
  y: number;
  w: number;
  h: number;
}

// aimBox is where a copy aims when it hits several targets: the centroid
// of their centres, with their average size for the contact distance.
// Null for no boxes.
export function aimBox(boxes: readonly StrikeBox[]): StrikeBox | null {
  if (boxes.length === 0) return null;
  let x = 0;
  let y = 0;
  let w = 0;
  let h = 0;
  for (const b of boxes) {
    x += b.x;
    y += b.y;
    w += b.w;
    h += b.h;
  }
  const n = boxes.length;
  return { x: x / n, y: y / n, w: w / n, h: h / n };
}

// halfExtent is how far a box reaches from its centre along the unit
// direction (ux, uy): |ux|·w/2 + |uy|·h/2.
function halfExtent(b: StrikeBox, ux: number, uy: number): number {
  return (Math.abs(ux) * b.w) / 2 + (Math.abs(uy) * b.h) / 2;
}

// lungeVector is how far the copy of `from` travels toward `to`: until
// its leading edge meets the target's edge plus STRIKE_CONTACT_OVERLAP_PX,
// clamped to between ½ and 0.85 of the centre-to-centre distance
// (ADR 0134 §2). Zero when the two share a centre.
export function lungeVector(from: StrikeBox, to: StrikeBox): Point & { travel: number } {
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  const d = Math.hypot(dx, dy);
  if (d < 1) return { x: 0, y: 0, travel: 0 };
  const ux = dx / d;
  const uy = dy / d;
  const contact = d - halfExtent(from, ux, uy) - halfExtent(to, ux, uy) + STRIKE_CONTACT_OVERLAP_PX;
  const travel = Math.min(LUNGE_MAX_FRACTION * d, Math.max(LUNGE_MIN_FRACTION * d, contact));
  return { x: ux * travel, y: uy * travel, travel };
}

// ---- Timing ----

export interface StrikeTimeline {
  outMs: number;
  holdMs: number;
  backMs: number;
  // When the copy touches, after the beat's cue: the impacts start here.
  contactMs: number;
  totalMs: number;
  shakeMs: number;
  deathFadeMs: number;
}

// strikeTimeline is every phase of a strike, scaled by animations.speed
// (ADR 0134 §2 and §5).
export function strikeTimeline(speed: number): StrikeTimeline {
  const s = Number.isFinite(speed) && speed > 0 ? speed : 1;
  return {
    outMs: STRIKE_OUT_MS * s,
    holdMs: STRIKE_HOLD_MS * s,
    backMs: STRIKE_BACK_MS * s,
    contactMs: STRIKE_OUT_MS * s,
    totalMs: STRIKE_MS * s,
    shakeMs: IMPACT_SHAKE_MS * s,
    deathFadeMs: STRIKE_DEATH_FADE_MS * s,
  };
}

// ---- Gates ----

// combatMotion is true only when every setting allows the strikes: the
// animations master switch, the combat toggle, and not reduced motion
// (ADR 0134 §5). accessibility.reduceMotion is never pushed into
// animations.ts, so the caller passes all three from the settings store.
export function combatMotion(s: {
  enabled: boolean;
  combat: boolean;
  reduceMotion: boolean;
}): boolean {
  return s.enabled && s.combat && !s.reduceMotion;
}

// strikesScheduled: a frame that arrives while the page is hidden
// schedules no strikes. GSAP runs on requestAnimationFrame, which a
// hidden tab does not run, so they would all play on return against a
// board that has moved on (ADR 0134 §4).
export function strikesScheduled(motion: boolean, visibility: string | undefined): boolean {
  return motion && (visibility === undefined || visibility === "visible");
}

// strikeLate: a strike whose start comes more than STRIKE_LATE_MS after
// the moment it was scheduled for (its frame's arrival plus its cue's
// offset) is dropped. The offset is the same-frame pause, which at speed
// 2 is itself over a second, so it is lateness that is measured, not
// the time since the frame.
export function strikeLate(frameAt: number, atMs: number, now: number): boolean {
  return now - (frameAt + atMs) > STRIKE_LATE_MS;
}

// ---- The cache ----

// CachedTile is a battlefield tile's last measured box (board pixels),
// its size at rest and its tap rotation, with the board size it was
// measured on. A copy of a creature that died is drawn from it.
export interface CachedTile {
  box: StrikeBox;
  // Unrotated width and height, and the rotation in degrees.
  width: number;
  height: number;
  rot: number;
  board: BoardSize;
}

const BOARD_RESIZE_TOLERANCE_PX = 1;

// usableTile returns the cached tile only when it was measured on a
// board of this size: a cached rect from a board of another size is
// never used (ADR 0053's reflow rule), and that strike is skipped.
export function usableTile(cached: CachedTile | undefined, board: BoardSize): CachedTile | null {
  if (!cached) return null;
  return Math.abs(cached.board.w - board.w) <= BOARD_RESIZE_TOLERANCE_PX &&
    Math.abs(cached.board.h - board.h) <= BOARD_RESIZE_TOLERANCE_PX
    ? cached
    : null;
}
