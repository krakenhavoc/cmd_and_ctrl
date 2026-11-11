// draftRatings.svelte.ts — the viewer's hide-ratings choice as reactive
// state, shared by the toggle and every badge (ADR 0222 §7, #5703).
//
// The choice itself is kept by draftRatings.ts, in localStorage, guarded.

import { cardAccessibleName, type DraftCard } from "./draft";
import { loadShowRatings, ratingAria, saveShowRatings } from "./draftRatings";

let show = $state(loadShowRatings());

export const ratingsPref = {
  /** Whether ratings are shown beside cards. */
  get show(): boolean {
    return show;
  },
  set(value: boolean): void {
    show = value;
    saveShowRatings(value);
  },
};

/**
 * A card's accessible name with its rating added, when ratings are shown
 * and the card has one: the badge is decoration for a screen reader, so
 * the card's own control says it ("…, 17Lands win rate 58.3 percent,
 * grade B+").
 */
export function ratedCardName(card: DraftCard): string {
  const name = cardAccessibleName(card);
  return show && card.rating ? `${name}, ${ratingAria(card.rating)}` : name;
}
