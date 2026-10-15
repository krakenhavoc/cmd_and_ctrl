// BlockListSheet.hint.ts — the first-use hint for blocking on the board
// (#3898, ADR 0125 §3.7). Blockers are declared by clicking a creature,
// then the attacker; the list is the fallback. A table hint shows only
// in a quiet moment, never during the declaration itself, so it is
// offered ahead of it, once the viewer has a creature that could block.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";
import { controlsCreature } from "../../hints/tableWhen";

const hint: Hint = {
  id: "table.block",
  version: 1,
  place: "table",
  order: 47,
  anchor: { label: L.actions },
  title: "Blocking on the board",
  body: `When you're attacked, click your creature, then the attacker it blocks. ${L.showAsList} in the dock lists them instead.`,
  when: controlsCreature,
};

export default hint;
