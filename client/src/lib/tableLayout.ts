// The table layouts (settings.display.tableLayout): where the seats sit
// on the board, and for "focus", how much of each opponent is drawn.
//
// "quadrant" keeps the around-the-table seating: the next seat beside
// you, the two across-table seats on top. "row" puts every opponent in
// turn order across the top and gives your board the full width.
// "focus" is the row arrangement split evenly: your board is the bottom
// half, every opponent is a summary in the top half, and you
// hover or click an avatar to see that player's whole board in the
// expanded overlay (ADR 0120).

/** The three ways the board can seat the table. */
export type TableLayout = "quadrant" | "row" | "focus";

/** Every layout, in the order the settings panel lists them. */
export const TABLE_LAYOUTS: readonly TableLayout[] = ["quadrant", "row", "focus"];

/** The default, and what an unknown stored value falls back to. */
export const DEFAULT_TABLE_LAYOUT: TableLayout = "row";

export function isTableLayout(v: unknown): v is TableLayout {
  return typeof v === "string" && (TABLE_LAYOUTS as readonly string[]).includes(v);
}

/**
 * Whether every opponent sits in one row across the top, with your
 * board the full width below. True for "row" and "focus"; "quadrant"
 * keeps the next seat beside you at four players.
 */
export function opponentsInARow(layout: TableLayout): boolean {
  return layout !== "quadrant";
}
