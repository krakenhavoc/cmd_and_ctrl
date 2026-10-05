// castAnyway.ts — the "Cast anyway (don't pay)" row (ADR 0118 §2,
// #2188).
//
// Whenever strict payment is on, every card the viewer could cast — a
// hand card with a castable face, an exile entry in the castable-from-
// other-zones strip whose verb is "cast", a commander in that strip or in
// the command-zone panel — offers this row in its right-click popover's
// Sandbox section, and in the admin override menu. Choosing it casts
// nothing yet: it asks, in the action dock, "Cast <card> without paying
// its mana cost?" (owner decision 6), and only Cast starts the ordinary
// cast chain with `forceCast` (targeting.ts CastChoices).
//
// The row lives on four surfaces three components deep under the Board,
// and the confirmation needs the Board's cast chain, so the request goes
// through this store rather than a callback threaded through every
// panel: the same shape as the Room doors' unlockRequest and discover's
// freeCastRequest. Board watches it, mounts the dock request while it is
// set, and clears it on Cast, Cancel or the card leaving its zone.

import { type Writable } from "svelte/store";

import { guardedWritable } from "./guardedStore";
import type { CardView } from "./protocol";
import { L } from "./labels";
import type { CastAnywayZone } from "./timing";

// The row's text and accessible name: a label contract (AGENTS.md §5).
// cast-anyway-2188.spec.ts selects the menu item by it.
export const CAST_ANYWAY_LABEL = L.castAnyway;
// The row's title on a live row.
export const CAST_ANYWAY_TITLE =
  "Cast it without paying its mana cost. The game log shows the table.";

export interface CastAnywayPending {
  card: CardView;
  zone: CastAnywayZone;
  // The one face an exile grant opens, handed to the cast chain with it
  // exactly as the strip's click hands it (exileStrip.ts).
  face?: number;
}

// The confirmation the dock is asking, or null.
export const castAnywayPending: Writable<CastAnywayPending | null> = guardedWritable(
  null,
  "castAnywayPending",
);

// requestCastAnyway opens the confirmation for `card` out of `zone`. It
// sends nothing.
export function requestCastAnyway(card: CardView, zone: CastAnywayZone, face?: number): void {
  castAnywayPending.set(face === undefined ? { card, zone } : { card, zone, face });
}

// clearCastAnyway closes it, also sending nothing.
export function clearCastAnyway(): void {
  castAnywayPending.set(null);
}
