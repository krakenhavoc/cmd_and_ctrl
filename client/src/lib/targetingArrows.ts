// targetingArrows — the arrows drawn while the viewer chooses targets
// (ADR 0119 §4, Arena's "Choose any target" view).
//
// While a targeting prompt is open:
//
//   - the SOURCE glows: the hand card being cast, the permanent whose
//     ability is being activated, or the source of a trigger choosing
//     its target (CR 113.7). It is found by `[data-instance-id]`
//     wherever it is drawn (the hand fan, the battlefield, the command
//     zone), through boardAnchor. A spell being retargeted on the stack
//     is found by its `[data-stack-item-id]`. If it is not on screen,
//     the arrows start from the dock's question line instead;
//   - each PICK gets a curved gold arrow from the source. Targets are
//     announced as the item is put on the stack and can't be changed
//     afterwards (CR 115.1, 601.2c), so these are the arrows the item
//     will carry. A graveyard or exile pick gets no arrow, as on the
//     stack: it has no stable place on the table;
//   - with a mouse or pen, while a pick is still open, one more arrow
//     FOLLOWS the pointer. It snaps to a legal target under the pointer
//     and turns green there. Touch and keyboard targeting have no hover
//     position and get none.
//
// Everything that is a decision lives here and is pinned by
// targetingArrows.test.ts. TargetingArrows.svelte only measures the
// DOM, calls `planTargetingArrows`, and draws what comes back.

import { boardCurve, insideBox, type Box, type Curve, type Point } from "./stackArrows";
import type { TargetRef, TargetingState } from "./targeting";

/** What a pick points at on the table. */
export type PickTargetKind = "permanent" | "player" | "stack";

export interface PickArrowPlan {
  /** Stable key, one per target. */
  id: string;
  kind: PickTargetKind;
  targetID: string;
}

/** The ids the snapshot puts somewhere an arrow can point at. */
export interface PickZones {
  battlefield: ReadonlySet<string>;
  stack: ReadonlySet<string>;
  players: ReadonlySet<string>;
}

/**
 * One arrow per pick, in pick order. A player pick points at that
 * seat's header; a card pick at its permanent on the battlefield or
 * its item on the stack. A card anywhere else (a graveyard, exile, a
 * hand) gets no arrow. A target picked twice by two clauses gets one.
 */
export function planPickArrows(picks: readonly TargetRef[], zones: PickZones): PickArrowPlan[] {
  const out: PickArrowPlan[] = [];
  const seen = new Set<string>();
  for (const p of picks) {
    let kind: PickTargetKind | null = null;
    if (p.kind === "player") {
      if (zones.players.has(p.id)) kind = "player";
    } else if (zones.battlefield.has(p.id)) {
      kind = "permanent";
    } else if (zones.stack.has(p.id)) {
      kind = "stack";
    }
    if (kind === null) continue;
    const id = `${kind}:${p.id}`;
    if (seen.has(id)) continue;
    seen.add(id);
    out.push({ id, kind, targetID: p.id });
  }
  return out;
}

/**
 * Whether the current step still takes another pick: an unbounded
 * clause always does, a counted one until it is full, and a divided
 * one until every point of the amount has a target (#1563).
 */
export function pickOpen(t: TargetingState): boolean {
  if (t.max > 0 && t.picked.length >= t.max) return false;
  if (t.divide !== undefined && t.picked.length >= t.divide) return false;
  return true;
}

/**
 * Only a pointer that hovers gets the follow arrow: a mouse or a pen.
 * Touch has no hover position, and nor does the keyboard.
 */
export function followsPointer(pointerType: string | null | undefined): boolean {
  return pointerType === "mouse" || pointerType === "pen";
}

/** The colours. Gold is the stack's arrow colour; green the legal-target ring's. */
export const PICK_ARROW_COLOR = "#ffd07a";
export const FOLLOW_ARROW_COLOR = "#ffd07a";
export const FOLLOW_SNAP_COLOR = "#6fe3a4";

export interface TargetingArrow extends Curve {
  id: string;
  kind: "pick" | "follow";
  /** For the follow arrow: true while it is snapped to a legal target. */
  snapped: boolean;
  color: string;
}

export interface TargetingArrowsInput {
  board: { width: number; height: number };
  /** The source's box in board coordinates, or null when it is not on screen. */
  source: Box | null;
  /** The dock's question line, in board coordinates, for an unseen source. */
  dock: Box | null;
  /** Each planned pick with its target's measured box, null when not on screen. */
  picks: readonly { plan: PickArrowPlan; box: Box | null; isSource?: boolean }[];
  /**
   * The follow arrow's end, or null for none: the pointer in board
   * coordinates, and the box of the legal target under it, if any.
   */
  follow: { pointer: Point; snap: Box | null } | null;
}

export interface TargetingArrowsPlan {
  /** Where the arrows start, or null when they cannot be drawn. */
  from: "source" | "dock" | null;
  arrows: TargetingArrow[];
}

/**
 * The arrows to draw, in board coordinates. They start from the
 * source, or from the dock's question line when the source is not on
 * screen, and from nowhere (no arrows) when neither is. A pick whose
 * target is not on screen, or is the source itself, gets no arrow.
 * The follow arrow is left out while the pointer is over the source.
 */
export function planTargetingArrows(input: TargetingArrowsInput): TargetingArrowsPlan {
  const start = input.source ?? input.dock;
  const from = input.source ? "source" : input.dock ? "dock" : null;
  const arrows: TargetingArrow[] = [];
  if (!start) return { from, arrows };
  for (const p of input.picks) {
    if (!p.box || p.isSource) continue;
    arrows.push({
      id: p.plan.id,
      kind: "pick",
      snapped: false,
      color: PICK_ARROW_COLOR,
      ...boardCurve(start, p.box, input.board),
    });
  }
  const f = input.follow;
  if (f && !(input.source && insideBox(f.pointer, input.source))) {
    const to: Box = f.snap ?? { left: f.pointer.x, top: f.pointer.y, width: 0, height: 0 };
    arrows.push({
      id: "follow",
      kind: "follow",
      snapped: f.snap !== null,
      color: f.snap ? FOLLOW_SNAP_COLOR : FOLLOW_ARROW_COLOR,
      ...boardCurve(start, to, input.board),
    });
  }
  return { from, arrows };
}
