// ActionDock.hint.ts — the first-use hint for the dock (ADR 0125 §3.7).

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";

const hint: Hint = {
  id: "table.dock",
  // 2: ADR 0143 §4.2, Pass turn became End turn and autopass Skip to my turn.
  // 3: #3267, the primary reads Pass, Next or No attacks by what it does.
  version: 3,
  place: "table",
  order: 10,
  anchor: { label: L.actions },
  title: "Your controls",
  body: `${L.next} moves on, ${L.pass} lets a spell resolve, ${L.endTurn} plays out your turn, ${L.skipToMyTurn} passes until yours. The game asks here.`,
};

export default hint;
