// StackLaneHost.hint.ts — the first-use hint for the stack (ADR 0125
// §3.7). Offered while the stack holds an item. Rule 3 of the quiet
// moment (§3.5) keeps it off an opponent's item while the viewer holds
// priority, so it shows over the viewer's own spell or after they pass.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";
import { stackIsEmpty } from "../../hints/tableWhen";

const hint: Hint = {
  id: "table.stack",
  version: 1,
  place: "table",
  order: 30,
  anchor: { label: L.stackPile.any },
  title: "The stack",
  body: "Spells and abilities wait here, newest on top, until every player passes. Rest the pointer on one to read it.",
  when: (c) => !stackIsEmpty(c),
};

export default hint;
