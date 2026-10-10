// menuPlacement — where a card's ability menu is drawn (#2960).
//
// The menu used to sit inside its Card, absolute, below the card. The
// lands row scrolls and clips, so the menu was cut off under the hand,
// and at 1998×716 it could not be seen at all. It is now a fixed box,
// like ManaSourcePicker, placed from the card's on-screen rect by
// `placeMenu` (above the card, flipped below when there is no room,
// clamped to the viewport) and moved out of the Card into the board's
// popover host by `anchoredMenu`. A Card carries a CSS transform (the
// tap rotation, the hover lift), and a transformed ancestor would pin
// a fixed box to the card instead of the viewport, hence the move.

import { placePopover, type AnchorRect, type PopoverPlacement } from "./manaSource";

/** The clear space kept on every edge of the viewport, in px. */
export const MENU_GUTTER = 8;
/** The gap between the card and its menu, in px. */
export const MENU_GAP = 6;

export interface MenuPlacement extends PopoverPlacement {
  /**
   * The tallest the menu may be drawn: the viewport less its gutters.
   * A menu taller than that scrolls inside itself, so its rows are
   * reachable and no ancestor ever scrolls for it.
   */
  maxHeight: number;
}

/**
 * placeMenu puts a `width`×`height` menu next to `anchor` in a `vw`×`vh`
 * viewport: above the card when it fits there, else below, else on
 * whichever side has more room, always inside the gutters.
 */
export function placeMenu(
  anchor: AnchorRect,
  width: number,
  height: number,
  vw: number,
  vh: number,
): MenuPlacement {
  const maxHeight = Math.max(0, vh - 2 * MENU_GUTTER);
  const h = Math.min(height, maxHeight);
  const p = placePopover(anchor, width, h, vw, vh, MENU_GUTTER, MENU_GAP);
  return { ...p, maxHeight };
}

/** The element the board mounts for menus to be moved into. */
export const POPOVER_HOST_ATTR = "data-popover-host";

/**
 * anchoredMenu is the action on a menu's fixed wrapper, which Card
 * renders as its child. It moves the wrapper into the board's popover
 * host (when the card is on a board; a Card mounted on its own keeps it
 * in place) and keeps it placed against the card, through resizes,
 * scrolls of any ancestor and changes of the menu's own size.
 */
export function anchoredMenu(node: HTMLElement): { destroy(): void } {
  const card = node.parentElement;
  const host = document.querySelector<HTMLElement>(`[${POPOVER_HOST_ATTR}]`);
  if (host && card) host.appendChild(node);

  const place = (): void => {
    if (!card) return;
    const r = card.getBoundingClientRect();
    const p = placeMenu(
      { left: r.left, right: r.right, top: r.top, bottom: r.bottom },
      node.offsetWidth,
      node.scrollHeight,
      window.innerWidth,
      window.innerHeight,
    );
    node.style.left = `${p.left}px`;
    node.style.top = `${p.top}px`;
    node.style.maxHeight = `${p.maxHeight}px`;
    node.dataset.side = p.side;
  };
  place();

  const resize = typeof ResizeObserver === "undefined" ? null : new ResizeObserver(place);
  resize?.observe(node);
  window.addEventListener("resize", place);
  // Capture: a scroll does not bubble, and the lands row scrolls.
  window.addEventListener("scroll", place, true);
  return {
    destroy() {
      resize?.disconnect();
      window.removeEventListener("resize", place);
      window.removeEventListener("scroll", place, true);
      node.remove();
    },
  };
}
