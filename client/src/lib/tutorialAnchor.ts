// tutorialAnchor.ts — resolving a tutorial step's anchor to the rect the
// scrim's hole covers (ADR 0076 §2.4, #1079).
//
// An anchor is an aria-label the e2e suite already asserts on, a card's
// instance id (Card.svelte's data-instance-id), or a seat's portrait
// (PlayerIdentity's data-seat-id, which CombatArrows already reads; the
// portrait's own label changes with the life total, so it cannot be the
// anchor). That is deliberate: those
// labels are a user-facing contract (AGENTS.md §5), so a refactor that
// renames one breaks board-layout.spec.ts before it breaks a new
// player's first session.
//
// "Missing" means no element, or only elements with no area: an empty
// list collapses to height 0, and a hole round nothing is exactly the
// "spotlighting empty space" §2.4 rules out. The caller advances the step.

import type { Anchor } from "./tutorial";

export interface AnchorRect {
  left: number;
  top: number;
  width: number;
  height: number;
}

function esc(s: string): string {
  return typeof CSS !== "undefined" && typeof CSS.escape === "function"
    ? CSS.escape(s)
    : s.replace(/["\\]/g, "\\$&");
}

/** resolveAnchor finds the element an anchor names, or null. */
export function resolveAnchor(a: Anchor, root: ParentNode = document): Element | null {
  if ("cardID" in a) {
    return root.querySelector(`[data-instance-id="${esc(a.cardID)}"]`);
  }
  if ("seatID" in a) {
    return root.querySelector(`[data-seat-id="${esc(a.seatID)}"]`);
  }
  let scope: ParentNode = root;
  if (a.within !== undefined) {
    const within = root.querySelector(`[aria-label="${esc(a.within)}"]`);
    if (!within) return null;
    scope = within;
  }
  return scope.querySelector(`[aria-label="${esc(a.label)}"]`);
}

/**
 * anchorRect is the union of the anchors' viewport rects, or null when
 * none of them is on the page with any area. One missing anchor out of
 * several (step 10's creature row plus the opponent's medallion) still
 * spotlights the others.
 */
export function anchorRect(anchors: Anchor[], root: ParentNode = document): AnchorRect | null {
  let l = Infinity;
  let t = Infinity;
  let r = -Infinity;
  let b = -Infinity;
  for (const a of anchors) {
    const el = resolveAnchor(a, root);
    if (!el) continue;
    const box = visibleBox(el);
    if (!box) continue;
    l = Math.min(l, box.left);
    t = Math.min(t, box.top);
    r = Math.max(r, box.right);
    b = Math.max(b, box.bottom);
  }
  if (l === Infinity) return null;
  return { left: l, top: t, width: r - l, height: b - t };
}

const CLIPS = new Set(["hidden", "clip", "auto", "scroll"]);

/**
 * visibleBox is the part of an element its scrolling or clipping
 * ancestors let show, or null when none of it shows. A creature card
 * is taller than the creature row on a short panel, and the row
 * scrolls; the hole goes round the part of the card a player can see
 * and click, not the part scrolled out under the lands.
 */
function visibleBox(
  el: Element,
): { left: number; top: number; right: number; bottom: number } | null {
  const box = el.getBoundingClientRect();
  let { left, top, right, bottom } = box;
  if (box.width <= 0 || box.height <= 0) return null;
  for (let p = el.parentElement; p; p = p.parentElement) {
    const cs = getComputedStyle(p);
    // The longhands, or the shorthand where an engine leaves them empty.
    const [ox, oy = ox] = (cs.overflow || "").split(/\s+/);
    const cx = CLIPS.has(cs.overflowX || ox);
    const cy = CLIPS.has(cs.overflowY || oy);
    if (!cx && !cy) continue;
    const pb = p.getBoundingClientRect();
    if (cx) {
      left = Math.max(left, pb.left);
      right = Math.min(right, pb.right);
    }
    if (cy) {
      top = Math.max(top, pb.top);
      bottom = Math.min(bottom, pb.bottom);
    }
  }
  if (right <= left || bottom <= top) return null;
  return { left, top, right, bottom };
}

/** sameRect compares two rects to the half pixel, to skip no-op updates. */
export function sameRect(a: AnchorRect | null, b: AnchorRect | null): boolean {
  if (a === null || b === null) return a === b;
  return (
    Math.abs(a.left - b.left) < 0.5 &&
    Math.abs(a.top - b.top) < 0.5 &&
    Math.abs(a.width - b.width) < 0.5 &&
    Math.abs(a.height - b.height) < 0.5
  );
}
