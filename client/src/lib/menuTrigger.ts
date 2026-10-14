// menuTrigger — opens a custom context menu on an element (#3499).
//
// The site's own right-click menu, built to be reused (the table card
// is the first surface; seat tiles, History rows and decks can follow).
// A right-click menu nobody can reach without a mouse is a hidden
// feature, so the action answers all three ways in:
//
//   - the mouse: `contextmenu`, at the cursor;
//   - the keyboard: the Menu key and Shift+F10 fire the same event on
//     the focused element, with no cursor (clientX/Y 0), so the menu
//     opens at that element instead;
//   - touch: a long press. Android fires `contextmenu` for it already;
//     iOS Safari does not, so a held touch that has not moved opens it
//     too, and the click that follows the lift is swallowed.
//
// It never takes the browser's menu away from text you might copy or a
// link you might open in a new tab: those keep the native one.
//
// Whatever opens the menu must also be reachable from the visible UI:
// the menu is a shortcut, never the only way to an action.

export interface MenuPoint {
  x: number;
  y: number;
}

/** How long a touch is held before it counts as a long press. */
export const LONG_PRESS_MS = 550;
/** How far a held touch may drift (px) and still be a press, not a scroll. */
export const LONG_PRESS_SLOP = 10;

const NATIVE_MENU =
  "input, textarea, select, a[href], [contenteditable]:not([contenteditable=false])";

/** True when the target keeps the browser's own menu: text fields and links. */
export function keepsNativeMenu(target: EventTarget | null): boolean {
  const el = target instanceof Element ? target : null;
  return el?.closest(NATIVE_MENU) != null;
}

/**
 * Where the menu opens. A keyboard-fired event has no cursor (both
 * coordinates 0), so it opens just inside the focused element's
 * top-left corner instead.
 */
export function menuPoint(
  e: { clientX: number; clientY: number },
  rect: { left: number; top: number } | null,
): MenuPoint {
  if (e.clientX !== 0 || e.clientY !== 0 || !rect) return { x: e.clientX, y: e.clientY };
  return { x: Math.round(rect.left + 16), y: Math.round(rect.top + 16) };
}

export type MenuOpener = (at: MenuPoint, opener: HTMLElement | null) => void;

/** Svelte action: `use:contextMenuTrigger={open}`. */
export function contextMenuTrigger(
  node: HTMLElement,
  open: MenuOpener,
): { update: (o: MenuOpener) => void; destroy: () => void } {
  let onOpen = open;
  let timer: ReturnType<typeof setTimeout> | null = null;
  let start: MenuPoint | null = null;
  // When a long press opened the menu: Android's own contextmenu for the
  // same press, a moment either side, does not open it a second time.
  let pressedAt = 0;
  // The click a long press's lift produces is not a tap on the card.
  let swallowClick = false;

  function cancel(): void {
    if (timer) clearTimeout(timer);
    timer = null;
    start = null;
  }

  function opener(target: EventTarget | null): HTMLElement | null {
    const active = document.activeElement;
    if (active instanceof HTMLElement && node.contains(active)) return active;
    return target instanceof HTMLElement ? target : null;
  }

  function onContextMenu(e: MouseEvent): void {
    if (keepsNativeMenu(e.target)) return;
    e.preventDefault();
    if (Date.now() - pressedAt < 800) return;
    // Android got there first: this press is the menu, so our own
    // timer stands down and the lift's click is not a tap.
    if (timer) {
      cancel();
      pressedAt = Date.now();
      swallowClick = true;
    }
    const from = opener(e.target);
    const rect = (from ?? node).getBoundingClientRect();
    onOpen(menuPoint(e, rect), from);
  }

  function onPointerDown(e: PointerEvent): void {
    swallowClick = false;
    if (e.pointerType !== "touch" || keepsNativeMenu(e.target)) return;
    cancel();
    start = { x: e.clientX, y: e.clientY };
    const target = e.target;
    timer = setTimeout(() => {
      timer = null;
      if (!start) return;
      pressedAt = Date.now();
      swallowClick = true;
      onOpen(start, opener(target));
      start = null;
    }, LONG_PRESS_MS);
  }

  function onPointerMove(e: PointerEvent): void {
    if (!start) return;
    if (Math.hypot(e.clientX - start.x, e.clientY - start.y) > LONG_PRESS_SLOP) cancel();
  }

  function onClick(e: MouseEvent): void {
    if (!swallowClick) return;
    swallowClick = false;
    e.preventDefault();
    e.stopPropagation();
  }

  node.addEventListener("contextmenu", onContextMenu);
  node.addEventListener("pointerdown", onPointerDown);
  node.addEventListener("pointermove", onPointerMove);
  node.addEventListener("pointerup", cancel);
  node.addEventListener("pointercancel", cancel);
  node.addEventListener("click", onClick, true);

  return {
    update(o: MenuOpener) {
      onOpen = o;
    },
    destroy() {
      cancel();
      node.removeEventListener("contextmenu", onContextMenu);
      node.removeEventListener("pointerdown", onPointerDown);
      node.removeEventListener("pointermove", onPointerMove);
      node.removeEventListener("pointerup", cancel);
      node.removeEventListener("pointercancel", cancel);
      node.removeEventListener("click", onClick, true);
    },
  };
}
