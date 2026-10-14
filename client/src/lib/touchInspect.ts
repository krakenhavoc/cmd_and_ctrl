// touchInspect.ts: reading a card without acting on it, where there is
// no hover (#3338).
//
// The zoom panel (HoverZoomOverlay) shows whatever the hoveredCard store
// holds, and a mouse fills it by resting on a card. A finger has no
// rest: on touch, pointerleave fires on lift and clears it, and the tap
// runs the card's click (cast, tap, select). So a tablet player could
// not read oracle text without risking the card.
//
// Here a card's zoom can be PINNED instead: a long-press on touch, or
// the Z key on a focused card. A pinned zoom stays until the next
// pointer-down anywhere or Escape, and the gesture that dismisses it is
// swallowed, so the tap that closes the zoom never acts on what is under
// it. The long-press's own click is swallowed too.

import { get } from "svelte/store";
import { hoveredCard } from "./cardTypes";
import type { CardView } from "./protocol";

/** How long a finger must rest on a card before its zoom pins. */
export const LONG_PRESS_MS = 450;
/** How far a finger may drift before the press counts as a scroll. */
export const LONG_PRESS_SLOP_PX = 10;
/** How long after the finger lifts a swallow still eats the click. */
const SWALLOW_AFTER_UP_MS = 400;
/** A swallow's hard stop, however long the press is held. */
const SWALLOW_CAP_MS = 10_000;

let pinnedID: string | null = null;
let teardown: (() => void) | null = null;

/** Whether a zoom is pinned, for one card when an id is given. */
export function isZoomPinned(instanceID?: string): boolean {
  return pinnedID !== null && (instanceID === undefined || pinnedID === instanceID);
}

/**
 * Eat the next click anywhere: the click the press now under way ends
 * in, however long it is held. It stays armed until shortly after the
 * pointer lifts, since a press with no click (a scroll) ends in none.
 */
export function swallowNextClick(): void {
  if (typeof window === "undefined") return;
  let after: ReturnType<typeof setTimeout> | null = null;
  const eat = (e: Event): void => {
    e.preventDefault();
    e.stopPropagation();
    e.stopImmediatePropagation();
    done();
  };
  const up = (): void => {
    if (after === null) after = setTimeout(done, SWALLOW_AFTER_UP_MS);
  };
  const cap = setTimeout(done, SWALLOW_CAP_MS);
  function done(): void {
    clearTimeout(cap);
    if (after !== null) clearTimeout(after);
    window.removeEventListener("click", eat, true);
    window.removeEventListener("pointerup", up, true);
    window.removeEventListener("pointercancel", up, true);
  }
  window.addEventListener("click", eat, true);
  window.addEventListener("pointerup", up, true);
  window.addEventListener("pointercancel", up, true);
}

/** Pin card's zoom until the next pointer-down anywhere or Escape. */
export function pinZoom(card: CardView): void {
  unpinZoom();
  hoveredCard.set(card);
  pinnedID = card.instance_id;
  if (typeof window === "undefined") return;
  const onDown = (): void => {
    unpinZoom();
    swallowNextClick();
  };
  const onKey = (e: KeyboardEvent): void => {
    if (e.key === "Escape") unpinZoom();
  };
  // Armed after the gesture that pinned it, so that gesture's own
  // events (the long-press's pointerdown is already past; a keydown is
  // still bubbling) cannot dismiss it at once.
  const arm = setTimeout(() => {
    window.addEventListener("pointerdown", onDown, true);
    window.addEventListener("keydown", onKey, true);
  }, 0);
  teardown = () => {
    clearTimeout(arm);
    window.removeEventListener("pointerdown", onDown, true);
    window.removeEventListener("keydown", onKey, true);
  };
}

/** Close a pinned zoom, if there is one. */
export function unpinZoom(): void {
  teardown?.();
  teardown = null;
  if (pinnedID === null) return;
  const id = pinnedID;
  pinnedID = null;
  if (get(hoveredCard)?.instance_id === id) hoveredCard.set(null);
}

/** Pin card's zoom, or close it when it is the one pinned. */
export function togglePinnedZoom(card: CardView): void {
  if (isZoomPinned(card.instance_id)) unpinZoom();
  else pinZoom(card);
}
