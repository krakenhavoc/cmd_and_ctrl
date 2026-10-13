// dock-primary.ts — the action dock's primary by its accessible name.
//
// #3267: the primary says what pressing it does, so its name is Pass
// while the stack holds something, No attacks in declare attackers
// before anything attacks, and Next otherwise. A spec that only wants
// "the button that passes" matches any of the three.

import { L } from "../../client/src/lib/labels";

export const DOCK_PRIMARY = new RegExp(`^(${[L.pass, L.next, L.noAttacks].join("|")})$`);
