// ExileStrip.hint.ts — the first-use hint for the commander (ADR 0125
// §3.7). Since #2349 the commander sits beside the viewer's hand, in
// the strip's command group, not in a tile with the other piles; the
// hint is offered while a commander waits there.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";
import { viewerSeat } from "../../hints/tableWhen";

const hint: Hint = {
  id: "table.commander",
  version: 1,
  place: "table",
  order: 50,
  anchor: { label: L.commandZone.any, within: L.yourBoard },
  title: "Your commander",
  body: "It waits beside your hand. Click it to cast it, as you would a card in hand.",
  when: (c) => (viewerSeat(c)?.command.count ?? 0) > 0,
};

export default hint;
