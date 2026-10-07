import type { CardView } from "./protocol";
import { manaValue } from "./handOrder";

// discardedManaValue is the X a "Discard a card with mana value X" cost
// announces (#2190, Kozilek, the Great Distortion): the mana value of
// the card the player picked out of their hand. The engine refuses an
// activation whose x_value is any other number, so the client sends
// exactly this and never asks for X separately. undefined when the pick
// is not in the hand (nothing was picked), which leaves the ordinary X
// prompt to ask — and the engine to refuse.
export function discardedManaValue(
  hand: readonly CardView[] | undefined,
  pickedIDs: readonly string[],
): number | undefined {
  if (pickedIDs.length !== 1) return undefined;
  const card = (hand ?? []).find((c) => c.instance_id === pickedIDs[0]);
  return card ? manaValue(card) : undefined;
}
