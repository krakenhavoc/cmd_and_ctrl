/**
 * specialActionReveal.ts — a turn-face-up whose morph cost is a reveal (#6249).
 *
 * CR 702.37e: turning a face-down permanent up means paying its morph
 * cost, and Temur Charger's is "Reveal a green card in your hand". The
 * server lists the hand cards that can pay it on the row as
 * `reveal_options`. A row like that is not sent straight from the menu:
 * the Board asks which card to reveal (or takes the only one) and sends
 * the payload with `reveal_id`.
 */

import type { CardView, GameView, SpecialActionView } from "./protocol";

/** The card and the row a `special_action` payload is about, when it needs a reveal. */
export interface SpecialActionRevealRow {
  card: CardView;
  row: SpecialActionView;
}

/**
 * specialActionNeedsReveal finds the row a `special_action` payload names
 * when that row's cost is a reveal and the payload names no card yet, or
 * null. Matched on kind, route and door, as specialActionNeedsX matches.
 */
export function specialActionNeedsReveal(
  view: GameView,
  params: unknown,
): SpecialActionRevealRow | null {
  if (!params || typeof params !== "object") return null;
  const p = params as Record<string, unknown>;
  if (p.reveal_id !== undefined || typeof p.card_id !== "string") return null;
  const card = (view.battlefield?.cards ?? []).find((c) => c.instance_id === p.card_id);
  const row = card?.special_actions?.find(
    (sa) =>
      sa.kind === p.kind &&
      (sa.route ?? undefined) === (p.route ?? undefined) &&
      (sa.door ?? undefined) === (p.door ?? undefined),
  );
  return card && row?.reveal_options?.length ? { card, row } : null;
}

/** revealOptionCards is the viewer's hand cards the row may reveal, in hand order. */
export function revealOptionCards(
  view: GameView,
  viewerID: string | null | undefined,
  row: SpecialActionView,
): CardView[] {
  if (!viewerID) return [];
  const ids = new Set(row.reveal_options ?? []);
  const me = view.seats.find((s) => s.id === viewerID);
  return (me?.hand.cards ?? []).filter((c) => ids.has(c.instance_id));
}

/** withReveal is the payload naming the card revealed. */
export function withReveal(params: unknown, id: string): Record<string, unknown> {
  return { ...(params as Record<string, unknown>), reveal_id: id };
}
