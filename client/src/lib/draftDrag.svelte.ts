// draftDrag.svelte.ts — dragging a draft card with pointer events
// (#5848 §1), so it works with a mouse, a pen and a finger.
//
// The old drag was HTML5 drag and drop, which is mouse-only and cannot
// start from the pack. This one is the pointer-event drag of the
// owner-approved mockup:
//
//   mouse / pen  press and move a few pixels: the card lifts.
//   touch        hold about 380 ms (the card lifts, and shows its
//                preview), then drag. A swipe that moves before the hold
//                is a scroll, and the press is dropped.
//
// A lifted card leaves a faded copy in place and a tilted copy follows
// the pointer. Any element with `data-drop-zone="main"` or `"side"` is a
// drop target and is highlighted while the pointer is over it (read
// `dragState.over`). Dropping on a zone calls the card's `ondrop`; Escape
// or a cancelled pointer puts the card back.
//
// This file is the pointer machinery only. The action that a card uses
// is lib/draftCardPointer.ts; the zones are components/draft/DraftSplit.

import { blockPreview, hidePreview } from "./draftPreview.svelte";
import type { Zone } from "./draftSplit";

/** Where a dragged card started: a zone of the picks, or the pack. */
export type DragFrom = Zone | "pack";

export interface DragSpec {
  from: DragFrom;
  /** Called once with the zone the card was dropped on (never the one it came from). */
  ondrop: (to: Zone) => void;
}

/** What a drop zone reads: whether a card is being dragged, and which zone it is over. */
export const dragState = $state<{ active: boolean; over: HTMLElement | null }>({
  active: false,
  over: null,
});

/** How long a touch holds before the card lifts. */
export const HOLD_MS = 380;
/** How far a pointer moves before it counts as a move, in pixels. */
const SLOP = 6;
const EDGE = 64;
const SCROLL_STEP = 16;

interface Press {
  el: HTMLElement;
  pointerId: number;
  x: number;
  y: number;
  touch: boolean;
  lifted: boolean;
  spec: DragSpec | undefined;
  onhold: () => void;
  onend: () => void;
  timer: ReturnType<typeof setTimeout> | undefined;
}
interface Drag {
  ghost: HTMLElement;
  fade: string;
  zone: HTMLElement | null;
  y: number;
  raf: number;
}

let press: Press | null = null;
let drag: Drag | null = null;
let stop: (() => void) | null = null;

function zoneOf(x: number, y: number): HTMLElement | null {
  const at = document.elementFromPoint?.(x, y);
  return at?.closest<HTMLElement>("[data-drop-zone]") ?? null;
}

function ghostOf(el: HTMLElement): HTMLElement {
  const art = el.querySelector<HTMLElement>("[data-card]");
  const g = (art ?? el).cloneNode(true) as HTMLElement;
  // A copy for the eye only: no second button for a test or a reader to find.
  g.removeAttribute("data-move-id");
  g.removeAttribute("aria-label");
  g.removeAttribute("title");
  for (const n of g.querySelectorAll("[data-move-id], [aria-label], [title]")) {
    n.removeAttribute("data-move-id");
    n.removeAttribute("aria-label");
    n.removeAttribute("title");
  }
  g.setAttribute("aria-hidden", "true");
  g.setAttribute("data-testid", "draft-drag-ghost");
  const rect = el.getBoundingClientRect();
  const width = art ? 120 : Math.min(260, rect.width || 220);
  Object.assign(g.style, {
    position: "fixed",
    left: "0",
    top: "0",
    width: `${width}px`,
    margin: "0",
    pointerEvents: "none",
    zIndex: "400",
    transformOrigin: "30% 20%",
    rotate: art ? "-4deg" : "0deg",
    filter: "drop-shadow(0 10px 18px rgba(0,0,0,0.6))",
    opacity: "1",
  });
  return g;
}

function setOver(zone: HTMLElement | null): void {
  if (!drag || drag.zone === zone) return;
  drag.zone = zone;
  dragState.over = zone;
}

function autoScroll(): void {
  if (!drag) return;
  const vh = window.innerHeight;
  const y = drag.y;
  if (y < EDGE) window.scrollBy?.(0, -Math.ceil(SCROLL_STEP * ((EDGE - y) / EDGE)));
  else if (y > vh - EDGE) window.scrollBy?.(0, Math.ceil(SCROLL_STEP * ((y - (vh - EDGE)) / EDGE)));
  drag.raf = requestAnimationFrame(autoScroll);
}

function startDrag(p: Press, ev: PointerEvent): void {
  hidePreview();
  blockPreview(true);
  const ghost = ghostOf(p.el);
  document.body.appendChild(ghost);
  const fade = p.el.style.opacity;
  p.el.style.opacity = "0.35";
  document.body.style.cursor = "grabbing";
  document.body.style.userSelect = "none";
  drag = { ghost, fade, zone: null, y: ev.clientY, raf: requestAnimationFrame(autoScroll) };
  dragState.active = true;
}

function moveDrag(ev: PointerEvent): void {
  if (!drag) return;
  drag.y = ev.clientY;
  drag.ghost.style.transform = `translate(${Math.round(ev.clientX - 30)}px, ${Math.round(ev.clientY - 20)}px)`;
  setOver(zoneOf(ev.clientX, ev.clientY));
}

/** Puts everything back: the ghost, the faded card, the highlight, the listeners. */
function finish(): void {
  const p = press;
  const d = drag;
  stop?.();
  stop = null;
  press = null;
  drag = null;
  if (p) clearTimeout(p.timer);
  if (d) {
    cancelAnimationFrame(d.raf);
    d.ghost.remove();
    if (p) p.el.style.opacity = d.fade;
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
    dragState.active = false;
    dragState.over = null;
    blockPreview(false);
  }
}

/** The click that follows a drag or a touch hold is not a pick. */
function swallowNextClick(): void {
  const eat = (ev: Event): void => {
    ev.preventDefault();
    ev.stopPropagation();
  };
  window.addEventListener("click", eat, { capture: true, once: true });
  setTimeout(() => window.removeEventListener("click", eat, { capture: true }), 60);
}

/** Ends whatever is in progress with nothing dropped (Escape, a removed card, a lost pointer). */
export function cancelDrag(): void {
  if (!press && !drag) return;
  const p = press;
  const hadDrag = drag !== null;
  finish();
  p?.onend();
  if (hadDrag) swallowNextClick();
}

/** Cancels the drag only if `el` is the card being pressed (a card removed mid-drag). */
export function cancelDragOf(el: HTMLElement): void {
  if (press?.el === el) cancelDrag();
}

/**
 * A pointer went down on a card. For a card with a `spec` it can be
 * dragged; `onhold` runs when a touch lifts it (to show its preview) and
 * `onend` when the press is over without a drop.
 */
export function pressCard(
  ev: PointerEvent,
  el: HTMLElement,
  spec: DragSpec | undefined,
  onhold: () => void,
  onend: () => void,
): void {
  if (press) return; // one finger at a time
  if (ev.button !== 0 && ev.pointerType === "mouse") return;
  const touch = ev.pointerType === "touch";
  if (!touch && !spec) return; // a mouse press on a card that cannot move is a click
  const p: Press = {
    el,
    pointerId: ev.pointerId,
    x: ev.clientX,
    y: ev.clientY,
    touch,
    lifted: !touch,
    spec,
    onhold,
    onend,
    timer: undefined,
  };
  press = p;
  if (touch) {
    p.timer = setTimeout(() => {
      if (press !== p) return;
      p.lifted = true;
      onhold();
      navigator.vibrate?.(15);
    }, HOLD_MS);
  }

  const onMove = (e: PointerEvent): void => {
    if (!press || e.pointerId !== press.pointerId) return;
    if (!press.el.isConnected) {
      cancelDrag();
      return;
    }
    if (!drag) {
      if (Math.hypot(e.clientX - press.x, e.clientY - press.y) <= SLOP) return;
      if (!press.lifted) {
        // A touch that moved before the hold is a scroll, not a drag.
        const q = press;
        finish();
        q.onend();
        return;
      }
      if (!press.spec) return;
      startDrag(press, e);
    }
    moveDrag(e);
  };
  const onUp = (e: PointerEvent): void => {
    if (!press || e.pointerId !== press.pointerId) return;
    const q = press;
    const zone = drag ? (zoneOf(e.clientX, e.clientY) ?? drag.zone) : null;
    const wasDrag = drag !== null;
    const wasHold = q.touch && q.lifted;
    finish();
    q.onend();
    if (wasDrag || wasHold) swallowNextClick();
    const to = zone?.dataset.dropZone;
    if (wasDrag && q.spec && (to === "main" || to === "side") && to !== q.spec.from) {
      q.spec.ondrop(to);
    }
  };
  const onCancel = (e: PointerEvent): void => {
    if (press && e.pointerId === press.pointerId) cancelDrag();
  };
  const onKey = (e: KeyboardEvent): void => {
    if (e.key === "Escape") cancelDrag();
  };
  // Once a touch has lifted a card the page must not scroll under it.
  const onTouchMove = (e: TouchEvent): void => {
    if (drag || press?.lifted) e.preventDefault();
  };
  const onMenu = (e: Event): void => {
    if (press?.touch) e.preventDefault();
  };

  window.addEventListener("pointermove", onMove);
  window.addEventListener("pointerup", onUp);
  window.addEventListener("pointercancel", onCancel);
  window.addEventListener("keydown", onKey);
  window.addEventListener("touchmove", onTouchMove, { passive: false });
  window.addEventListener("contextmenu", onMenu);
  stop = () => {
    window.removeEventListener("pointermove", onMove);
    window.removeEventListener("pointerup", onUp);
    window.removeEventListener("pointercancel", onCancel);
    window.removeEventListener("keydown", onKey);
    window.removeEventListener("touchmove", onTouchMove);
    window.removeEventListener("contextmenu", onMenu);
  };
}
