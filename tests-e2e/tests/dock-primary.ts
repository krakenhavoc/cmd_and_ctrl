// dock-primary.ts — the action dock's primary by its accessible name.
//
// #3267: the primary says what pressing it does, so its name is Pass
// while the stack holds something, No attacks in declare attackers
// before anything attacks, and Next otherwise; #3266 adds Attack with N
// while attackers are staged. A spec that only wants "the button that
// passes" matches any of them.
//
// #3826: at the last pass of your own turn, and the last one before it
// begins, the server marks the pass and the primary reads "Pass · turn
// ends" or "Pass · your turn next" instead of Next.

import { L } from "../../client/src/lib/labels";

export const DOCK_PRIMARY = new RegExp(
  `^(${[L.pass, L.next, L.nextEndsTurn, L.nextYourTurn, L.noAttacks].join("|")}|${L.attackWith.any.stem}\\d+)$`,
);
