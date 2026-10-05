// ActionDock.hint.ts — the first-use hint for the dock (ADR 0125 §3.7).

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";

const hint: Hint = {
  id: "table.dock",
  version: 1,
  place: "table",
  order: 10,
  anchor: { label: L.actions },
  title: "Your controls",
  body: "The next button moves the game on, Pass turn ends your turn, and anything the game needs from you opens here.",
};

export default hint;
