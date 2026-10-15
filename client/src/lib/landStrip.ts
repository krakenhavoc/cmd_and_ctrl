// The land strip's piles (BattlefieldRow `strip`), the parts that are
// pure rules rather than layout.
//
// #3044: on a big board a pile of 28 Forests drew 28 copies, each 4px
// right and 3px up of the one before: a skewed fan of slivers 120px
// wide that read as nothing and was hard to click. A pile now draws at
// most STRIP_DEPTH copies' worth of stack, whatever its size, and its
// count badge says how many; the copies past that sit exactly under
// the bottom one. Tapped lands pile by name too, so twelve tapped
// basics are one turned card with a count, after the untapped piles,
// and the split between the two stays readable at a glance.

import type { CardView } from "./protocol";

/** The most copies of a pile that show as a step in its stack. */
export const STRIP_DEPTH = 4;

/**
 * pileDepths is each copy's step in a pile of `n`, bottom (first in
 * the DOM) to top: 0 for every copy hidden under the bottom one, then
 * 1, 2, … up to STRIP_DEPTH - 1 for the top copy. A pile of
 * STRIP_DEPTH or fewer steps every copy, as it always did.
 */
export function pileDepths(n: number): number[] {
  const hidden = Math.max(0, n - STRIP_DEPTH);
  return Array.from({ length: n }, (_, i) => Math.max(0, i - hidden));
}

/**
 * namePileKey is the pile a land joins in the strip: one per name and
 * tapped state, or null for a land drawn on its own. A commander, or a
 * land with something attached (an enchanted land), keeps a pile of its
 * own: stacking it under identical basics would hide what makes it
 * different.
 */
export function namePileKey(c: CardView, hasAttachments: boolean): string | null {
  if (c.is_commander || hasAttachments) return null;
  const name = c.name || c.instance_id;
  return `${c.tapped ? "tapped" : "pile"}:${name}`;
}
