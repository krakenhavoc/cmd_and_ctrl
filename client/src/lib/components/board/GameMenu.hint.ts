// GameMenu.hint.ts — the first-use hint for the ⋯ menu (ADR 0125 §3.7),
// offered from the viewer's second turn.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";
import { atViewersSecondTurn } from "../../hints/tableWhen";

const hint: Hint = {
  id: "table.more",
  version: 1,
  place: "table",
  order: 70,
  anchor: { label: L.moreActions },
  title: "More actions",
  body: "Dice and a coin, untap all, life history, table settings, a vote and Help are in the ⋯ menu.",
  when: atViewersSecondTurn,
};

export default hint;
