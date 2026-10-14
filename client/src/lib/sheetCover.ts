// sheetCover.ts — whether an open dock sheet covers something the player
// has to click on the board (#2959, #3095).
//
// A dock sheet grows up out of the action dock in the bottom-right
// corner and floats over the board (ADR 0111 §3). On a short desktop
// window, 1280x720 or 1366x768, it reaches the opponents' creature
// rows, and some sheets ask the player to click exactly what is under
// them: a target, a permanent a choice offers (#2880), a blocker's
// damage stepper (ADR 0147). Minimising the sheet was the only way to
// reach them.
//
// When this says a freshly opened sheet covers one of them, the dock
// folds the sheet down to its question line, once per sheet
// (ActionDock's `sheetCoversBoard`). The title, the count and the
// confirm stay in the dock, the board stays as it is, and one click
// brings the sheet back; a sheet brought back is not folded again.
//
// Making the board narrower beside the sheet was tried first and
// rejected: at 1280x720 it leaves a board about 660px wide, where the
// viewer's own panel stacks its rows, clips its creatures and loses the
// hand fan.

// What the player can act on, on the board, while a sheet is open: a
// card ringed as a legal target or a choice's option (Card.svelte's
// `targetable`), one already picked, and a blocker's damage stepper.
export const BOARD_ACTION_SELECTOR =
  '.card.targetable, .card.picked, [data-testid="damage-stepper"]';

export interface Box {
  left: number;
  top: number;
  right: number;
  bottom: number;
}

// overlaps reports whether two boxes share any area. Touching edges and
// an empty box (an element that is not laid out) do not count.
export function overlaps(a: Box, b: Box): boolean {
  const empty = (r: Box): boolean => r.right <= r.left || r.bottom <= r.top;
  if (empty(a) || empty(b)) return false;
  return a.left < b.right && b.left < a.right && a.top < b.bottom && b.top < a.bottom;
}

// sheetCoversBoardAction reports whether the open sheet lies over any of
// the board elements the player can act on. No sheet covers nothing.
export function sheetCoversBoardAction(sheet: Box | null, actions: readonly Box[]): boolean {
  if (sheet === null) return false;
  return actions.some((r) => overlaps(sheet, r));
}
