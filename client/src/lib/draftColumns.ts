// draftColumns.ts — Main in columns by mana value, for Deck view (#5848
// §5): one column each for mana value 0 to 5 and a last for 6 and more,
// cards in a column sorted by name.

import type { DraftCard } from "./draft";

/** Columns shown: mana value 0 to 5, and 6 and more. */
export const MV_COLUMNS = 7;

/** The column a card is in: its mana value, 0 for a land, 6 for 6 and more. */
export function columnOf(card: Pick<DraftCard, "mana_value">): number {
  const mv = card.mana_value ?? 0;
  return Math.min(Math.max(Number.isFinite(mv) ? Math.floor(mv) : 0, 0), MV_COLUMNS - 1);
}

/** "0" … "5", "6+". */
export function columnLabel(column: number): string {
  return column >= MV_COLUMNS - 1 ? `${MV_COLUMNS - 1}+` : String(column);
}

/** `cards` in `MV_COLUMNS` columns, each sorted by name then id. */
export function byManaValue(cards: readonly DraftCard[]): DraftCard[][] {
  const cols: DraftCard[][] = Array.from({ length: MV_COLUMNS }, () => []);
  for (const c of cards) cols[columnOf(c)].push(c);
  for (const col of cols)
    col.sort((a, b) => a.name.localeCompare(b.name) || a.id.localeCompare(b.id));
  return cols;
}
