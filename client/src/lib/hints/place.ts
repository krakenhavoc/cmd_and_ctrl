// place.ts — where a hint's card goes (ADR 0125 §3.6).
//
// Pure geometry over rectangles, so place.test.ts pins it without a
// browser. The card tries below, above, right and left of its anchor,
// and takes the first side where it fits on screen without covering the
// anchor or anything it must keep clear: the action dock, the stack
// pile, an open dialog or sheet, the tutorial coach. If no side fits,
// the hint waits, exactly as for a moment that is not quiet.
//
// On a phone (below PHONE_MAX_WIDTH) there is no choosing: the card is a
// one-line strip on the bottom edge (stacked on the dock bar at the
// table, which the component's CSS does), with the anchor ringed.

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

export type Side = "below" | "above" | "right" | "left";

/** The order sides are tried in. */
export const SIDES: readonly Side[] = ["below", "above", "right", "left"];

/** The widest viewport that gets the phone strip (ADR 0076 §2.3's ≤599px). */
export const PHONE_MAX_WIDTH = 599;
/** The card's width on a desktop. */
export const CARD_WIDTH = 280;
/** Space between the anchor and the card. */
export const GAP = 10;
/** Space the card keeps from the viewport's edges. */
export const MARGIN = 8;

export type Placement =
  | { kind: "card"; side: Side; left: number; top: number }
  | { kind: "strip" }
  | { kind: "wait" };

export interface PlaceInput {
  anchor: Rect;
  card: Size;
  viewport: Size;
  /** What the card must not cover besides its anchor. */
  avoid?: readonly Rect[];
  gap?: number;
  margin?: number;
}

export function isPhone(viewportWidth: number): boolean {
  return viewportWidth <= PHONE_MAX_WIDTH;
}

/** overlaps reports whether two rects share any area (touching edges do not). */
export function overlaps(a: Rect, b: Rect): boolean {
  return (
    a.left < b.left + b.width &&
    b.left < a.left + a.width &&
    a.top < b.top + b.height &&
    b.top < a.top + a.height
  );
}

function clamp(v: number, lo: number, hi: number): number {
  return Math.max(lo, Math.min(hi, v));
}

/** at is where the card sits on one side of the anchor, centred along it and kept on screen. */
function at(side: Side, a: Rect, card: Size, view: Size, gap: number, margin: number): Rect {
  const maxLeft = view.width - margin - card.width;
  const maxTop = view.height - margin - card.height;
  switch (side) {
    case "below":
    case "above": {
      const left = clamp(a.left + a.width / 2 - card.width / 2, margin, Math.max(margin, maxLeft));
      const top = side === "below" ? a.top + a.height + gap : a.top - gap - card.height;
      return { left, top, ...card };
    }
    case "right":
    case "left": {
      const top = clamp(a.top + a.height / 2 - card.height / 2, margin, Math.max(margin, maxTop));
      const left = side === "right" ? a.left + a.width + gap : a.left - gap - card.width;
      return { left, top, ...card };
    }
  }
}

function onScreen(r: Rect, view: Size, margin: number): boolean {
  return (
    r.left >= margin &&
    r.top >= margin &&
    r.left + r.width <= view.width - margin &&
    r.top + r.height <= view.height - margin
  );
}

/**
 * placeCard is the card's placement: a side and a position, the phone
 * strip, or "wait" when no side fits.
 */
export function placeCard(input: PlaceInput): Placement {
  if (isPhone(input.viewport.width)) return { kind: "strip" };
  const gap = input.gap ?? GAP;
  const margin = input.margin ?? MARGIN;
  const avoid = [input.anchor, ...(input.avoid ?? [])];
  for (const side of SIDES) {
    const r = at(side, input.anchor, input.card, input.viewport, gap, margin);
    if (!onScreen(r, input.viewport, margin)) continue;
    if (avoid.some((x) => overlaps(r, x))) continue;
    return { kind: "card", side, left: r.left, top: r.top };
  }
  return { kind: "wait" };
}
