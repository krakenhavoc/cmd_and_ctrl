// OpeningRollBanner.hint.ts — the first-use hint for the opening roll
// (ADR 0125 §3.7). Offered once the viewer has rolled and someone has
// not: until then the roll's own request is open, and no hint shows over
// a decision (§3.5). Once the roll is over the banner is gone and the
// hint is never offered.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";
import { openingRollModel } from "../../openingRoll";
import { viewerSeat } from "../../hints/tableWhen";

const hint: Hint = {
  id: "table.opening-roll",
  version: 1,
  place: "table",
  order: 20,
  anchor: { label: L.openingRoll },
  title: "Rolling for the first turn",
  body: "Everyone rolls at once. The highest roll chooses who goes first, and a tie rolls again.",
  when: (c) => {
    const me = viewerSeat(c);
    const model = c.view ? openingRollModel(c.view) : null;
    if (!me || !model || model.chooser !== null) return false;
    const mine = model.chips.find((chip) => chip.seat === me.seat);
    return mine?.state === "rolled" && model.waiting.length > 0;
  },
};

export default hint;
