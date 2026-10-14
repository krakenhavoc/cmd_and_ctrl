// stackPosition.ts — #3429, ADR 0176: where the viewer has dragged the
// docked stack card to.
//
// The server keeps a seat's spot for the rest of the game and sends it back
// on that seat's own view (GameView.stack_positions). A spot is the card's
// top-left corner as a fraction of the board's width and height, so it
// survives a window resize; what is drawn is that spot clamped to keep the
// whole card on the board at the size it has now. The saved fractions are
// not rewritten by the clamp, so a board that grows again gives the card
// back its place.

import type { GameView, StackPositionView } from "./protocol";

export interface Size {
  width: number;
  height: number;
}

/** The viewer's saved spot, null when the card is at its default. */
export function savedStackPosition(
  view: GameView | null | undefined,
  viewerID: string | null,
): StackPositionView | null {
  if (!viewerID) return null;
  const p = view?.stack_positions?.[viewerID];
  if (!p || !Number.isFinite(p.x) || !Number.isFinite(p.y)) return null;
  return { x: p.x, y: p.y };
}

function clamp(n: number, lo: number, hi: number): number {
  return Math.min(Math.max(n, lo), hi);
}

/**
 * Where to draw the card, in pixels from the board's top-left corner, with
 * the whole card kept on the board. A card larger than the board sits at
 * the corner.
 */
export function stackPixels(
  pos: StackPositionView,
  board: Size,
  panel: Size,
): { left: number; top: number } {
  return {
    left: clamp(pos.x * board.width, 0, Math.max(0, board.width - panel.width)),
    top: clamp(pos.y * board.height, 0, Math.max(0, board.height - panel.height)),
  };
}

/**
 * The fractions for a card whose top-left corner is at (left, top) pixels
 * from the board's corner, clamped to stay on the board. The inverse of
 * stackPixels for any spot that is on the board.
 */
export function stackFractions(
  left: number,
  top: number,
  board: Size,
  panel: Size,
): StackPositionView {
  if (board.width <= 0 || board.height <= 0) return { x: 0, y: 0 };
  const px = stackPixels({ x: left / board.width, y: top / board.height }, board, panel);
  return { x: px.left / board.width, y: px.top / board.height };
}

/** Whether two spots are the same to the pixel on a board this size. */
export function sameStackPosition(
  a: StackPositionView | null,
  b: StackPositionView | null,
): boolean {
  if (!a || !b) return a === b;
  return a.x === b.x && a.y === b.y;
}
