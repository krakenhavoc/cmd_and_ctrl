// payCards.ts — ADR 0108 §5: a pay-unless prompt whose payment is not
// mana. "Echo—Discard a card", "Echo—Sacrifice two lands" and the
// cumulative upkeeps that discard or sacrifice ship `pay_cards`; the
// "Pay" answer names exactly `count` of its options as `card_ids`. The
// payer may only answer "Pay" once they have picked that many, and when
// there are fewer options than that they cannot pay at all (CR 118.3).

import type { CardView, PayCardsView, PlayerView } from "./protocol";

// canPayCards reports whether the chooser has enough to pay with.
export function canPayCards(pc: PayCardsView): boolean {
  return pc.options.length >= pc.count;
}

// payCardsReady reports whether `picks` is a whole payment.
export function payCardsReady(pc: PayCardsView, picks: string[]): boolean {
  return picks.length === pc.count;
}

// togglePayCard adds or removes `id`, never past the count.
export function togglePayCard(pc: PayCardsView, picks: string[], id: string): string[] {
  if (picks.includes(id)) return picks.filter((c) => c !== id);
  if (picks.length >= pc.count || !pc.options.includes(id)) return picks;
  return [...picks, id];
}

// payCardsOptions resolves the option IDs to the cards they name: the
// chooser's hand for a discard, the battlefield for a sacrifice, in the
// server's order.
export function payCardsOptions(
  pc: PayCardsView,
  battlefield: CardView[] | undefined,
  chooser: PlayerView | undefined,
): CardView[] {
  const pool = pc.action === "discard" ? (chooser?.hand?.cards ?? []) : (battlefield ?? []);
  const byID = new Map(pool.map((c) => [c.instance_id, c]));
  return pc.options.map((id) => byID.get(id)).filter((c): c is CardView => c !== undefined);
}

// payCardsVerb is the picker's instruction.
export function payCardsVerb(pc: PayCardsView): string {
  const n = pc.count === 1 ? "one card" : `${pc.count} cards`;
  return pc.action === "discard" ? `Choose ${n} to discard.` : `Choose ${n} to sacrifice.`;
}
