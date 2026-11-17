// draftPreview.svelte.ts — the large card preview of the draft room
// (#5848 §2): which card is shown, where it sits, and when it appears.
//
// A card shows a preview after a short pause (the "Hover preview delay"
// setting, display.hoverDelayMs, which the table's own zoom uses too).
// While one is showing, moving to another card swaps it at once with no
// second wait; leaving every card hides it after a short grace, so the
// gap between two cards does not count as leaving. Keyboard focus and a
// touch hold show it with no pause. Escape and scrolling hide it, and it
// stays hidden while a card is being dragged.
//
// The state is module-wide so a swap works from the pack to the picks.
// Each card names the `scope` of the host that draws it
// (components/draft/DraftPreview.svelte), so two hosts on one screen do
// not both draw the same preview.

import { get } from "svelte/store";

import type { DraftCard } from "./draft";
import { settings } from "./settings";

/** The part of a DOMRect the placement needs. */
export interface Anchor {
  left: number;
  top: number;
  right: number;
  bottom: number;
}

export const preview = $state<{ card: DraftCard | null; anchor: Anchor | null; scope: string }>({
  card: null,
  anchor: null,
  scope: "",
});

/** A phone: the preview is centred above or below the card instead of beside it. */
export const PHONE_MAX = 720;
const GAP = 14;
const MARGIN = 8;
/** How long a pointer may be between two cards before the preview goes. */
const LEAVE_GRACE_MS = 90;

/**
 * Where the preview's top-left corner goes, for a card at `anchor` and a
 * preview of `w` by `h` in a `vw` by `vh` window. Beside the card, never
 * on it: to the right, or the left when the right has no room; on a
 * phone, centred above the card, or below it when the top has no room.
 */
export function placePreview(
  anchor: Anchor,
  size: { w: number; h: number },
  view: { vw: number; vh: number },
): { x: number; y: number } {
  const { w, h } = size;
  const { vw, vh } = view;
  const mid = (anchor.top + anchor.bottom) / 2;
  if (vw < PHONE_MAX) {
    const x = Math.max(MARGIN, (vw - w) / 2);
    const above = anchor.top - h - GAP;
    const y = above >= MARGIN ? above : Math.min(vh - h - MARGIN, anchor.bottom + GAP);
    return { x: Math.round(x), y: Math.round(Math.max(MARGIN, y)) };
  }
  let x = anchor.right + GAP;
  if (x + w > vw - MARGIN) x = anchor.left - GAP - w;
  x = Math.max(MARGIN, x);
  const y = Math.max(MARGIN, Math.min(vh - h - MARGIN, mid - h / 2));
  return { x: Math.round(x), y: Math.round(y) };
}

let showTimer: ReturnType<typeof setTimeout> | undefined;
let leaveTimer: ReturnType<typeof setTimeout> | undefined;
let blocked = false;
let unwatch: (() => void) | null = null;

function delayMs(): number {
  const d = get(settings).display.hoverDelayMs;
  return Number.isFinite(d) && d > 0 ? d : 0;
}

function rectOf(el: Element): Anchor {
  const r = el.getBoundingClientRect();
  return { left: r.left, top: r.top, right: r.right, bottom: r.bottom };
}

function show(card: DraftCard, el: Element, scope: string): void {
  clearTimeout(showTimer);
  clearTimeout(leaveTimer);
  if (blocked) return;
  preview.card = card;
  preview.anchor = rectOf(el);
  preview.scope = scope;
  if (!unwatch) {
    // Escape and any scroll hide it: the card it was beside has moved.
    const onKey = (ev: KeyboardEvent): void => {
      if (ev.key === "Escape") hidePreview();
    };
    const onScroll = (): void => hidePreview();
    window.addEventListener("keydown", onKey);
    window.addEventListener("scroll", onScroll, { capture: true, passive: true });
    unwatch = () => {
      window.removeEventListener("keydown", onKey);
      window.removeEventListener("scroll", onScroll, { capture: true });
    };
  }
}

/** True while a preview is on screen. */
export function previewShowing(): boolean {
  return preview.card !== null;
}

/**
 * A pointer came onto a card. The preview appears after the setting's
 * delay, or at once when another card's is already showing.
 */
export function previewEnter(card: DraftCard, el: Element, scope: string): void {
  if (blocked) return;
  clearTimeout(showTimer);
  clearTimeout(leaveTimer);
  if (previewShowing()) show(card, el, scope);
  else showTimer = setTimeout(() => show(card, el, scope), delayMs());
}

/** A pointer left a card: the preview goes unless it lands on another card at once. */
export function previewLeave(): void {
  clearTimeout(showTimer);
  clearTimeout(leaveTimer);
  if (!previewShowing()) return;
  leaveTimer = setTimeout(hidePreview, LEAVE_GRACE_MS);
}

/** Keyboard focus, or a touch hold: shown with no pause. */
export function previewNow(card: DraftCard, el: Element, scope: string): void {
  show(card, el, scope);
}

export function hidePreview(): void {
  clearTimeout(showTimer);
  clearTimeout(leaveTimer);
  preview.card = null;
  preview.anchor = null;
  unwatch?.();
  unwatch = null;
}

/** A drag in progress keeps the preview hidden (and any pending one from coming). */
export function blockPreview(on: boolean): void {
  blocked = on;
  if (on) hidePreview();
}
