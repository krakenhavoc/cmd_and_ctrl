// BoardExpand.hint.ts — the first-use hint for the expanded board (ADR
// 0120, ADR 0125 §3.7). Anchored on the first opponent's board, and
// offered at a table of three or more seats, where the boards are small.

import { L } from "../../labels";
import type { Hint } from "../../hints/hint";
import { firstOpponent } from "../../hints/tableWhen";

const hint: Hint = {
  id: "table.expand",
  version: 1,
  place: "table",
  order: 80,
  anchor: (c) => {
    const opp = firstOpponent(c);
    return opp ? { label: L.seatBoard(opp.name) } : null;
  },
  title: "See a board up close",
  body: "Rest the pointer on a board and its ⤢ opens it full size over the table.",
  when: (c) => (c.view?.seats.length ?? 0) >= 3,
};

export default hint;
