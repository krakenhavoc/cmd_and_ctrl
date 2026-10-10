// ActionDock.hint.ts — the first-use hint for the dock (ADR 0125 §3.7).

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";

const hint: Hint = {
  id: "table.dock",
  // 2: ADR 0143 §4.2, Pass turn became End turn and autopass Skip to my turn.
  version: 2,
  place: "table",
  order: 10,
  anchor: { label: L.actions },
  title: "Your controls",
  body: "next moves the game on, End turn plays out your turn, and Skip to my turn passes until yours. Anything the game needs opens here.",
};

export default hint;
