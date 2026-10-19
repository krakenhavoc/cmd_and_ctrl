// deckSideboard.ts — the read-only sideboard list on a saved deck (ADR 0191,
// amended 2026-10-19, #3833 and #4054).
//
// A deck's sideboard is the cards its owner keeps outside the game. It is
// not a pile on the table: the Decks page lists it, and in a game only an
// effect reaches it. The list is read from GET /me/decks/{id}/printings,
// which already names every saved row's board, so no new route carries it.

import type { DeckPrintingRow } from "./api";

/** One line of the list: a card and how many copies the sideboard holds. */
export interface SideboardLine {
  name: string;
  count: number;
}

/** The board name the printings route gives a sideboard row. */
export const SIDEBOARD_BOARD = "Sideboard";

/**
 * sideboardLines is the sideboard's cards, one line per name, in name order.
 * Rows of other boards (the mainboard, the commanders, a companion) are not
 * in it; a name listed on two rows is one line with the copies added.
 */
export function sideboardLines(rows: readonly DeckPrintingRow[]): SideboardLine[] {
  const byName = new Map<string, number>();
  for (const r of rows) {
    if (r.board !== SIDEBOARD_BOARD) continue;
    byName.set(r.name, (byName.get(r.name) ?? 0) + Math.max(1, r.count));
  }
  return [...byName.entries()]
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => a.name.localeCompare(b.name));
}

/** sideboardTotal is how many cards the sideboard holds, copies counted. */
export function sideboardTotal(lines: readonly SideboardLine[]): number {
  return lines.reduce((n, l) => n + l.count, 0);
}
