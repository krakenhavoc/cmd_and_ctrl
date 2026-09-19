// chosenValues — #781. The one place a permanent's chosen colour or
// named creature type becomes something a player can read.
//
// CR 105.4 and CR 614.12 both make the answer public: it is announced
// as the permanent enters and hidden from nobody. CR 607.2d is why it
// has to be ON THE CARD rather than only in the log — the linked
// ability says "creatures you control of the chosen color", and that
// set is uncomputable without knowing the answer. An opponent needs it
// to know which of their creatures a Heraldic Banner is pumping; the
// controller needs it because the prompt closed ten turns ago.
//
// One function, three call sites (the card tile's badge row, the hover
// panel's footer, and the zone browser through the tile). A second
// mapping of "G" to "Green" somewhere in a component is the thing this
// module exists to stop — and the letters it reads come from the same
// COLOR_META the choose_color prompt itself renders, so the badge says
// the word the player clicked.

import { COLOR_META } from "./manaPick";
import type { CardView } from "./protocol";

/** One chip: what it says, and what its tooltip says. */
export interface ChosenValueChip {
  /** Stable key for `{#each}`, and the CSS modifier for the chip. */
  kind: "color" | "tribe";
  /** The rendered text — "Green", "Elf". */
  label: string;
  /**
   * The tooltip. Quotes the printed clause the answer feeds, because
   * that is what makes the chip mean something: a lone "Green" next to
   * a Coldsteel Heart does not say that its {T} ability is what reads
   * it (CR 607.2d).
   */
  title: string;
}

/**
 * chosenValueChips returns the chips for one card, in a fixed order
 * (colour then type) so two permanents that chose both never render
 * them the other way round.
 *
 * Empty for the overwhelming majority of cards. A card the viewer is
 * not a knower of carries neither field — the server strips them with
 * the other type-derived bits — so this needs no visibility test of
 * its own, and must not grow one: what the viewer may see is the
 * server's decision (#429).
 */
export function chosenValueChips(
  card: Pick<CardView, "chosen_color" | "named_tribe">,
): ChosenValueChip[] {
  const chips: ChosenValueChip[] = [];
  const color = card.chosen_color;
  if (color) {
    // An unknown letter renders as the letter rather than as nothing:
    // the answer was given and the player is entitled to see it, even
    // if a future server learns a colour this client does not.
    const label = COLOR_META[color]?.label ?? color;
    chips.push({
      kind: "color",
      label,
      title: `"the chosen color" is ${label} (CR 105.4)`,
    });
  }
  const tribe = card.named_tribe;
  if (tribe) {
    chips.push({
      kind: "tribe",
      label: tribe,
      title: `"the chosen type" is ${tribe} (CR 614.12)`,
    });
  }
  return chips;
}
