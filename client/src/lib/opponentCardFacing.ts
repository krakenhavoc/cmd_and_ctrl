// #4555: which way the cards on a top-row (flipped) opponent's panel face
// (settings.display.opponentCardFacing).
//
// "them" turns them 180°, facing their owner across the table as they
// would at a real one: the default, and what the table has always done.
// "me" draws them upright, facing the viewer. Only the card drawing
// differs; PlayerPanel.svelte keeps every row, pile and identity where it
// is.

/** The two ways an opponent's cards can face. */
export type OpponentCardFacing = "them" | "me";

/** Every value, in the order the settings panel lists them. */
export const OPPONENT_CARD_FACINGS: readonly OpponentCardFacing[] = ["them", "me"];

/** The default, and what an unknown stored value falls back to. */
export const DEFAULT_OPPONENT_CARD_FACING: OpponentCardFacing = "them";

export function isOpponentCardFacing(v: unknown): v is OpponentCardFacing {
  return typeof v === "string" && (OPPONENT_CARD_FACINGS as readonly string[]).includes(v);
}
