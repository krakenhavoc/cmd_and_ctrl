// stackArrows — where the fan stack lane's target arrows go (#1467).
//
// The fan style draws a curved, dashed arrow from each stack item to
// each thing it targets:
//
//   - a PERMANENT: the card on the board, `[data-instance-id]`;
//   - a PLAYER: that seat's header, `[data-seat-id]`;
//   - ANOTHER STACK ITEM: that item's tile in the lane,
//     `[data-stack-item-id]`, drawn as an arc over the lane (the
//     mockup's Counterspell → Lightning Bolt).
//
// A `card` target (a graveyard or exiled card) gets no arrow: it has
// no stable place on the table to point at.
//
// Same approach as CombatArrows.svelte: read positions with
// getBoundingClientRect, translate them into the coordinates of the
// layer the arrow is drawn in, and let the component re-measure on
// every snapshot tick and on resize. Everything that is a DECISION —
// which targets get an arrow, in which layer, in which colour, where
// each end sits on its rect, how the curve bows — is here and pinned
// by stackArrows.test.ts; the component only calls `measureStackArrows`
// and draws what comes back.

import type { StackLaneItem, StackLaneTargetKind } from "./stackLane";

export interface Point {
  x: number;
  y: number;
}

/** A rect in some coordinate space; a DOMRect satisfies it. */
export interface Box {
  left: number;
  top: number;
  width: number;
  height: number;
}

/** A quadratic bezier: start, control point, end. */
export interface Curve {
  x1: number;
  y1: number;
  cx: number;
  cy: number;
  x2: number;
  y2: number;
}

/**
 * Which SVG an arrow is drawn in. "board" arrows leave the lane for
 * the table and are drawn in a board-sized layer under the lane;
 * "lane" arcs join two tiles in the lane and are drawn inside it.
 */
export type ArrowLayer = "board" | "lane";

/** The target kinds that get an arrow. */
export type ArrowTargetKind = Exclude<StackLaneTargetKind, "card">;

export interface ArrowPlan {
  /** Stable key: one per (item, target slot). */
  id: string;
  itemID: string;
  targetKind: ArrowTargetKind;
  targetID: string;
  layer: ArrowLayer;
  /** A CSS colour: the caster's seat colour, or the flag tone for the top item. */
  color: string;
  /** True for an arrow from the item that resolves next. */
  fromTop: boolean;
}

export interface StackArrow extends ArrowPlan, Curve {}

/**
 * The top item's arrows use the flag tone — the gold the lane uses for
 * "next" and the ring around the top card — so the arrow that is about
 * to matter reads as the one to look at.
 */
export const TOP_ARROW_COLOR = "var(--gold)";

/** Gap between an arrowhead and the rect it points at, in px. */
export const TARGET_GAP = 6;

/** How far a board arrow bows, as a fraction of its chord. */
export const BOARD_BOW = 0.2;

/** The lane arc's lift above the higher of its two tiles, in px. */
export const LANE_ARC_MIN_LIFT = 14;
export const LANE_ARC_MAX_LIFT = 26;

/**
 * Every arrow the model asks for, in item order then target order.
 * A `card` target is skipped, and so is a duplicate slot aimed at the
 * same thing twice (one arrow says it).
 */
export function planArrows(items: readonly StackLaneItem[]): ArrowPlan[] {
  const out: ArrowPlan[] = [];
  for (const item of items) {
    const seen = new Set<string>();
    for (const t of item.targets) {
      if (t.kind === "card") continue;
      const key = `${t.kind}:${t.id}`;
      if (seen.has(key)) continue;
      seen.add(key);
      out.push({
        id: `${item.id}->${key}`,
        itemID: item.id,
        targetKind: t.kind,
        targetID: t.id,
        layer: t.kind === "stack" ? "lane" : "board",
        color: item.isTop ? TOP_ARROW_COLOR : item.casterColor,
        fromTop: item.isTop,
      });
    }
  }
  return out;
}

function cssEscape(s: string): string {
  return typeof CSS !== "undefined" && typeof CSS.escape === "function"
    ? CSS.escape(s)
    : s.replace(/["\\]/g, "\\$&");
}

/** The attribute selector that finds a target of this kind. */
export function targetSelector(kind: ArrowTargetKind, id: string): string {
  const v = cssEscape(id);
  switch (kind) {
    case "permanent":
      return `[data-instance-id="${v}"]`;
    case "player":
      return `[data-seat-id="${v}"]`;
    case "stack":
      return `[data-stack-item-id="${v}"]`;
  }
}

/** `box` in the coordinates of `origin` (whose top-left becomes 0,0). */
export function relativeTo(box: Box, origin: Box): Box {
  return {
    left: box.left - origin.left,
    top: box.top - origin.top,
    width: box.width,
    height: box.height,
  };
}

export function boxCentre(b: Box): Point {
  return { x: b.left + b.width / 2, y: b.top + b.height / 2 };
}

/**
 * Where the ray from the box's centre toward `toward` leaves the box,
 * pushed `gap` further along the ray. A point inside the box (or at
 * its centre) answers the centre: there is no edge to aim at.
 */
export function edgePoint(b: Box, toward: Point, gap = 0): Point {
  const c = boxCentre(b);
  const dx = toward.x - c.x;
  const dy = toward.y - c.y;
  const hw = b.width / 2;
  const hh = b.height / 2;
  if (dx === 0 && dy === 0) return c;
  // Scale at which the ray meets the nearer pair of sides.
  const sx = dx === 0 ? Infinity : hw / Math.abs(dx);
  const sy = dy === 0 ? Infinity : hh / Math.abs(dy);
  const s = Math.min(sx, sy);
  if (s >= 1) return c; // `toward` is inside the box
  const len = Math.hypot(dx, dy);
  const g = gap / len;
  return { x: c.x + dx * (s + g), y: c.y + dy * (s + g) };
}

/**
 * The control point of a curve from `from` to `to`, bowed off the
 * chord by `bow` of its length, on whichever side is nearer `centre` —
 * the table's centre, so an arrow reads as an arc across the table
 * rather than one flung out past its edge.
 *
 * Unless that side is inside `avoid`. The lane floats over the table's
 * centre, and an arrow that starts on the lane's border and bows back
 * into it would be drawn across the lane's own header and tiles. A
 * quadratic stays inside the triangle of its three points, so with
 * both ends and the control point outside the lane on the same side,
 * the whole curve is.
 */
export function controlTowardCentre(
  from: Point,
  to: Point,
  centre: Point,
  bow = BOARD_BOW,
  avoid?: Box | null,
): Point {
  const mx = (from.x + to.x) / 2;
  const my = (from.y + to.y) / 2;
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  const len = Math.hypot(dx, dy);
  if (len === 0) return { x: mx, y: my };
  const nx = -dy / len;
  const ny = dx / len;
  const a = { x: mx + nx * len * bow, y: my + ny * len * bow };
  const b = { x: mx - nx * len * bow, y: my - ny * len * bow };
  const da = Math.hypot(a.x - centre.x, a.y - centre.y);
  const db = Math.hypot(b.x - centre.x, b.y - centre.y);
  const [near, far] = da <= db ? [a, b] : [b, a];
  if (avoid && insideBox(near, avoid) && !insideBox(far, avoid)) return far;
  return near;
}

/** True when `p` lies inside `b` (edges included). */
export function insideBox(p: Point, b: Box): boolean {
  return p.x >= b.left && p.x <= b.left + b.width && p.y >= b.top && p.y <= b.top + b.height;
}

/**
 * Where a board arrow leaves the lane panel: on the lane's border
 * straight above (or below, or beside) its tile, on the side facing the
 * target. The arrow starts there rather than at the tile so it never
 * crosses the lane's header, summary or the other tiles, and it still
 * reads as coming from that tile.
 */
export function laneExit(tile: Box, target: Point, lane: Box): Point {
  const t = boxCentre(tile);
  const clampX = Math.min(Math.max(t.x, lane.left), lane.left + lane.width);
  const clampY = Math.min(Math.max(t.y, lane.top), lane.top + lane.height);
  if (target.y < lane.top) return { x: clampX, y: lane.top };
  if (target.y > lane.top + lane.height) return { x: clampX, y: lane.top + lane.height };
  if (target.x < lane.left) return { x: lane.left, y: clampY };
  if (target.x > lane.left + lane.width) return { x: lane.left + lane.width, y: clampY };
  return { x: clampX, y: lane.top };
}

/**
 * A board arrow from a lane tile to a target on the table, both in
 * board coordinates. It starts on the lane's border by its tile
 * (laneExit), or on the tile's own edge when there is no lane rect,
 * and ends TARGET_GAP short of the target's edge facing it, so the
 * arrowhead never sits on the card art. It bows toward the table's
 * centre.
 */
export function boardCurve(
  from: Box,
  to: Box,
  board: { width: number; height: number },
  lane?: Box | null,
): Curve {
  const aim = boxCentre(to);
  const start = lane ? laneExit(from, aim, lane) : edgePoint(from, aim);
  const end = edgePoint(to, start, TARGET_GAP);
  const centre = { x: board.width / 2, y: board.height / 2 };
  const ctrl = controlTowardCentre(start, end, centre, BOARD_BOW, lane);
  return { x1: start.x, y1: start.y, cx: ctrl.x, cy: ctrl.y, x2: end.x, y2: end.y };
}

/**
 * An arc over the lane from one tile to another, both in the lane's
 * coordinates. It leaves the source's top edge on the side facing the
 * target, rises above the higher tile, and lands on the target's top
 * edge a gap above it — the mockup's Counterspell → Lightning Bolt.
 */
export function laneArc(from: Box, to: Box): Curve {
  const dir = boxCentre(to).x >= boxCentre(from).x ? 1 : -1;
  const x1 = boxCentre(from).x + dir * from.width * 0.3;
  const y1 = from.top;
  const x2 = boxCentre(to).x;
  const y2 = to.top - TARGET_GAP;
  const lift = Math.min(LANE_ARC_MAX_LIFT, Math.max(LANE_ARC_MIN_LIFT, Math.abs(x2 - x1) * 0.2));
  // A quadratic peaks halfway to its control point, so the control
  // sits twice the lift above the higher end.
  const cy = Math.min(y1, y2) - lift * 2;
  return { x1, y1, cx: (x1 + x2) / 2, cy, x2, y2 };
}

/** SVG path data for a curve. */
export function curvePath(c: Curve): string {
  const r = (n: number) => Math.round(n * 10) / 10;
  return `M ${r(c.x1)} ${r(c.y1)} Q ${r(c.cx)} ${r(c.cy)} ${r(c.x2)} ${r(c.y2)}`;
}

// ---------------------------------------------------------------- //
// DOM measuring

function hasSize(el: Element): boolean {
  const r = el.getBoundingClientRect();
  return r.width > 0 || r.height > 0;
}

/**
 * The element a target resolves to, or null. A permanent or player is
 * looked for on the board OUTSIDE the lane (the lane never carries
 * those attributes today, but a style that did must not point at
 * itself); a stack target only inside it. An element with no size —
 * a collapsed seat, a hidden duplicate — is passed over: there is
 * nothing on screen to point at.
 */
export function findTarget(
  board: ParentNode,
  lane: Element,
  kind: ArrowTargetKind,
  id: string,
): HTMLElement | null {
  const sel = targetSelector(kind, id);
  if (kind === "stack") {
    const el = lane.querySelector<HTMLElement>(sel);
    return el && hasSize(el) ? el : null;
  }
  for (const el of board.querySelectorAll<HTMLElement>(sel)) {
    if (lane.contains(el)) continue;
    if (hasSize(el)) return el;
  }
  return null;
}

/** A board element the lane is pointing at, for the highlight ring. */
export interface MarkedTarget {
  el: HTMLElement;
  color: string;
  fromTop: boolean;
}

export interface MeasuredArrows {
  /** Board coordinates, for the board-sized layer. */
  board: StackArrow[];
  /** Lane-track content coordinates, for the layer inside the track. */
  lane: StackArrow[];
  /** Every board element an arrow reached, once each. */
  targets: MarkedTarget[];
}

export interface MeasureInput {
  /** The .board element; board arrows are relative to it. */
  board: HTMLElement;
  /**
   * The lane panel. Stack targets are found in it, and board arrows
   * leave from its border.
   */
  lane: Element;
  /**
   * The scrolling track the tiles sit in. Lane arcs are relative to
   * its content (scroll included); a tile scrolled out of its visible
   * area draws no board arrow.
   */
  track: HTMLElement;
  items: readonly StackLaneItem[];
}

/**
 * Measure every planned arrow against the live DOM. An arrow is left
 * out — never drawn to a guessed position — when its tile or target is
 * not on screen, when its tile is scrolled out of the lane, or when
 * its target sits under the lane panel, where there is nothing visible
 * to point at. A target that resolves is still MARKED for the
 * highlight ring in the last two cases: the lane does target it.
 */
export function measureStackArrows(input: MeasureInput): MeasuredArrows {
  const { board, lane, track, items } = input;
  const boardRect = board.getBoundingClientRect();
  const trackRect = track.getBoundingClientRect();
  const trackInBoard = relativeTo(trackRect, boardRect);
  const laneInBoard = relativeTo(lane.getBoundingClientRect(), boardRect);
  // Lane arcs live in the track's scrolled content box.
  const laneOrigin: Box = {
    left: trackRect.left - track.scrollLeft + (track.clientLeft || 0),
    top: trackRect.top - track.scrollTop + (track.clientTop || 0),
    width: trackRect.width,
    height: trackRect.height,
  };

  const out: MeasuredArrows = { board: [], lane: [], targets: [] };
  const marked = new Map<HTMLElement, MarkedTarget>();

  for (const plan of planArrows(items)) {
    const src = lane.querySelector<HTMLElement>(targetSelector("stack", plan.itemID));
    if (!src || !hasSize(src)) continue;
    const dst = findTarget(board, lane, plan.targetKind, plan.targetID);
    if (!dst) continue;
    if (plan.layer === "lane") {
      if (dst === src) continue;
      const curve = laneArc(
        relativeTo(src.getBoundingClientRect(), laneOrigin),
        relativeTo(dst.getBoundingClientRect(), laneOrigin),
      );
      out.lane.push({ ...plan, ...curve });
      continue;
    }
    const from = relativeTo(src.getBoundingClientRect(), boardRect);
    const to = relativeTo(dst.getBoundingClientRect(), boardRect);
    const tileX = boxCentre(from).x;
    const tileVisible =
      tileX >= trackInBoard.left && tileX <= trackInBoard.left + trackInBoard.width;
    const targetCovered = insideBox(boxCentre(to), laneInBoard);
    if (tileVisible && !targetCovered) {
      out.board.push({ ...plan, ...boardCurve(from, to, boardRect, laneInBoard) });
    }
    // The first arrow to reach an element picks its ring colour; the
    // top item's arrow always wins, since it is the one that matters
    // next.
    const prev = marked.get(dst);
    if (!prev || (plan.fromTop && !prev.fromTop)) {
      marked.set(dst, { el: dst, color: plan.color, fromTop: plan.fromTop });
    }
  }
  out.targets = [...marked.values()];
  return out;
}

/** The attribute the lane sets on a board element it points at. */
export const TARGET_ATTR = "data-stack-lane-target";
/** The custom property that carries that element's ring colour. */
export const TARGET_RING_VAR = "--stack-lane-ring";

/**
 * Put the highlight ring on exactly `next`'s elements: mark the new
 * ones, update the colour of the kept ones, and unmark the ones that
 * are no longer targeted. Returns the set now marked, for the next
 * call (and for the final unmark on teardown).
 */
export function syncTargetMarks(
  prev: ReadonlySet<HTMLElement>,
  next: readonly MarkedTarget[],
): Set<HTMLElement> {
  const now = new Set<HTMLElement>();
  for (const t of next) {
    t.el.setAttribute(TARGET_ATTR, t.fromTop ? "next" : "");
    t.el.style.setProperty(TARGET_RING_VAR, t.color);
    now.add(t.el);
  }
  for (const el of prev) {
    if (now.has(el)) continue;
    el.removeAttribute(TARGET_ATTR);
    el.style.removeProperty(TARGET_RING_VAR);
  }
  return now;
}
