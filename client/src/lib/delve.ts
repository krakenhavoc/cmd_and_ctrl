// delve.ts — ADR 0100 §5: the pure half of the delve picker.
//
// Delve (CR 702.66a) lets a spell exile cards from its caster's
// graveyard, each paying for {1} of the generic mana. The server owns
// both numbers the picker needs: which cards may be exiled
// (`card.delve.options`, in its payment order) and how many (the auto-tap
// preview's `delve_budget`, priced for the announcement as it stands).
// Nothing here works a budget out of a mana string (#916's rule).

import type { AutoTapPreview } from "./api";
import type { CardView } from "./protocol";

// delveOptionIDs is the card's delve candidates, in the server's order.
export function delveOptionIDs(card: CardView | null | undefined): string[] {
  return card?.delve?.options?.cards ?? [];
}

// hasDelveChoice reports whether the cast chain should open the delve
// picker at all: the card has delve and there is something to exile.
export function hasDelveChoice(card: CardView | null | undefined): boolean {
  return delveOptionIDs(card).length > 0;
}

// delveLimit is the picker's cap: the preview's budget when it answered
// (an absent `delve_budget` is the server's omitted zero), otherwise the
// card's default-announcement hint, and never more than there are cards
// to exile.
export function delveLimit(
  preview: AutoTapPreview | null | undefined,
  card: CardView | null | undefined,
): number {
  const budget = preview ? (preview.delve_budget ?? 0) : (card?.delve?.max ?? 0);
  return Math.max(0, Math.min(budget, delveOptionIDs(card).length));
}

// chooseDelveForMe is the "Choose for me" button (ADR 0100 owner
// decision 2): the first `limit` candidates in the server's payment
// order — lands first, then cards with no graveyard cast surface, then
// the rest, oldest first — which is what a bot with no fuel policy
// pays. It fills the picker and never confirms for the player.
export function chooseDelveForMe(optionIDs: string[], limit: number): string[] {
  return optionIDs.slice(0, Math.max(0, limit));
}
