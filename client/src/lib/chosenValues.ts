// chosenValues — #781, extended by #1648. The one place a permanent's
// chosen colour, named creature type, chosen card name or chosen
// option becomes something a player can read.
//
// CR 105.4 and CR 614.12 all make the answer public: it is announced
// as the permanent enters and hidden from nobody. CR 607.2d is why it
// has to be ON THE CARD rather than only in the log — the linked
// ability says "creatures you control of the chosen color", and that
// set is uncomputable without knowing the answer. An opponent needs it
// to know which of their creatures a Heraldic Banner is pumping; the
// controller needs it because the prompt closed ten turns ago. The
// same is true of a Siege's chosen option (ADR 0071 — it decides which
// of the permanent's own printed abilities exists at all) and of a
// Pithing Needle's chosen name (it is the whole reason a targeted
// source's activated abilities are greyed out).
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
  kind: "color" | "tribe" | "option" | "name" | "number";
  /**
   * The rendered text — "Green", "Elf" for colour/tribe, which read
   * fine bare in context. "Temur" and "Sol Ring" alone do not (a
   * three-clan word and an ordinary card name look like anything else
   * on the tile), so option/name carry their own short prefix:
   * "Mode: Temur", "Named: Sol Ring".
   */
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
 * (colour, type, option, name) so two permanents that chose more than
 * one of these never render them the other way round. In practice no
 * printed card asks more than one of these questions, but the order
 * is still fixed rather than left to object key iteration.
 *
 * Empty for the overwhelming majority of cards. A card the viewer is
 * not a knower of carries none of the fields — the server strips them
 * with the other type-derived bits — so this needs no visibility test
 * of its own, and must not grow one: what the viewer may see is the
 * server's decision (#429).
 */
export function chosenValueChips(
  card: Pick<
    CardView,
    "chosen_color" | "named_tribe" | "chosen_option" | "chosen_name" | "chosen_number"
  >,
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
  const option = card.chosen_option;
  if (option) {
    chips.push({
      kind: "option",
      label: `Mode: ${option}`,
      title: `"the chosen option" is ${option} — the ability after it is live (ADR 0071)`,
    });
  }
  const name = card.chosen_name;
  if (name) {
    chips.push({
      kind: "name",
      label: `Named: ${name}`,
      title: `"the chosen name" is ${name} (CR 614.12)`,
    });
  }
  // #1941: a number chosen (or life paid) as the permanent entered —
  // Phyrexian Processor's, the size of every token it makes.
  const number = card.chosen_number;
  if (number) {
    chips.push({
      kind: "number",
      label: `Paid: ${number}`,
      title: `${number} was paid as it entered (CR 614.12a)`,
    });
  }
  return chips;
}
