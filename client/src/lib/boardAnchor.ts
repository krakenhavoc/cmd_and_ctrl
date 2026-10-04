// boardAnchor — the one place a card tile or a seat's avatar is found
// in the DOM by its `data-instance-id` / `data-seat-id` (ADR 0120 §3).
//
// Today every such lookup takes the first match, and a card or a seat is
// drawn once on the table, so the first match is THE element. ADR 0120
// adds a second copy of one seat's board, drawn over the table in an
// expanded overlay (`[data-board-expanded]`). While it is open the
// overlay's copy is the one a player sees and clicks, and the table's copy
// is covered, so the overlay's copy has to be the anchor: combat and
// stack-target arrows, the fan lane's arrows and rings, tutorial
// spotlights and popover anchors all point at it, and fall back to the
// table the moment the overlay closes.
//
// So every reader asks here, and this module answers "inside an open
// overlay first, then anywhere else in the root". With no overlay mounted
// that is exactly the old first match, which is what keeps this change
// invisible until the overlay lands (ADR 0120 PR 3).

/** The attribute the expanded-board overlay carries (ADR 0120 §3). */
export const BOARD_EXPANDED_ATTR = "data-board-expanded";

function cssEscape(s: string): string {
  return typeof CSS !== "undefined" && typeof CSS.escape === "function"
    ? CSS.escape(s)
    : s.replace(/["\\]/g, "\\$&");
}

/** The selector for a card tile (Card.svelte) by instance ID. */
export function cardSelector(instanceID: string): string {
  return `[data-instance-id="${cssEscape(instanceID)}"]`;
}

/** The selector for a seat's avatar (PlayerIdentity.svelte) by seat ID. */
export function seatSelector(seatID: string): string {
  return `[data-seat-id="${cssEscape(seatID)}"]`;
}

export interface AnchorOptions {
  /**
   * Passes over an element the caller cannot use: one with no size, one
   * inside the stack lane. The first accepted element wins, overlay
   * copies first. Unset accepts every match.
   */
  accept?: (el: HTMLElement) => boolean;
}

/** Whether an element takes up any room on screen. */
export function hasSize(el: Element): boolean {
  const r = el.getBoundingClientRect();
  return r.width > 0 || r.height > 0;
}

function overlaysIn(root: ParentNode): HTMLElement[] {
  return [...root.querySelectorAll<HTMLElement>(`[${BOARD_EXPANDED_ATTR}]`)];
}

/**
 * findAnchor is the element `selector` names under `root`: the first
 * accepted match inside an expanded overlay, else the first accepted
 * match anywhere under `root`. Null when nothing is accepted.
 */
export function findAnchor(
  root: ParentNode,
  selector: string,
  opts: AnchorOptions = {},
): HTMLElement | null {
  const ok = opts.accept ?? (() => true);
  for (const overlay of overlaysIn(root)) {
    for (const el of overlay.querySelectorAll<HTMLElement>(selector)) {
      if (ok(el)) return el;
    }
  }
  for (const el of root.querySelectorAll<HTMLElement>(selector)) {
    if (ok(el)) return el;
  }
  return null;
}

/** findCardAnchor is findAnchor for a card tile by instance ID. */
export function findCardAnchor(
  root: ParentNode,
  instanceID: string,
  opts?: AnchorOptions,
): HTMLElement | null {
  return findAnchor(root, cardSelector(instanceID), opts);
}

/** findSeatAnchor is findAnchor for a seat's avatar by seat ID. */
export function findSeatAnchor(
  root: ParentNode,
  seatID: string,
  opts?: AnchorOptions,
): HTMLElement | null {
  return findAnchor(root, seatSelector(seatID), opts);
}

/**
 * cardAnchors is one element per instance ID under `root`, the one
 * findCardAnchor would return for that ID: an overlay copy beats the
 * table's, and otherwise the first in document order.
 */
export function cardAnchors(root: ParentNode): Map<string, HTMLElement> {
  const out = new Map<string, HTMLElement>();
  const take = (scope: ParentNode) => {
    for (const el of scope.querySelectorAll<HTMLElement>("[data-instance-id]")) {
      const id = el.dataset.instanceId;
      if (id && !out.has(id)) out.set(id, el);
    }
  };
  for (const overlay of overlaysIn(root)) take(overlay);
  take(root);
  return out;
}
