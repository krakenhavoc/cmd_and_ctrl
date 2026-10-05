// PlayerPanel.hint.ts — the first-use hint for abilities (ADR 0125
// §3.7). A permanent's abilities live on its right-click menu; a touch
// screen has the ready pip instead (ADR 0105). Offered once the viewer
// controls a permanent that has an ability.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";
import { controlsAbility } from "../../hints/tableWhen";

const hint: Hint = {
  id: "table.right-click",
  version: 1,
  place: "table",
  order: 40,
  anchor: { label: L.yourBoard },
  title: "Abilities live on right-click",
  body: "Right-click a permanent to see its abilities and use them. On a touch screen, tap the small bolt on a ready card.",
  when: controlsAbility,
};

export default hint;
