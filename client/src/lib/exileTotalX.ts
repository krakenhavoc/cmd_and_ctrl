import type { CardView } from "./protocol";
import { manaValue } from "./handOrder";

// exiledTotalManaValue is the X an "Exile one or more other artifacts you
// control with total mana value X" cost announces (#5401, Fabrication
// Foundry): the total mana value of the permanents the player picked. The
// engine refuses an activation whose x_value is any other number, so the
// client sends exactly this and never asks for X separately. A
// permanent's {X} counts as zero (CR 202.3e). undefined when nothing was
// picked or a pick is not on the battlefield, which leaves the ordinary
// X prompt to ask — and the engine to refuse.
export function exiledTotalManaValue(
  battlefield: readonly CardView[] | undefined,
  pickedIDs: readonly string[],
): number | undefined {
  if (pickedIDs.length === 0) return undefined;
  let total = 0;
  for (const id of pickedIDs) {
    const card = (battlefield ?? []).find((c) => c.instance_id === id);
    if (!card) return undefined;
    total += manaValue(card);
  }
  return total;
}
