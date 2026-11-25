// payCards.ts — ADR 0108 §5: a pay-unless prompt whose payment is not
// mana. "Echo—Discard a card", "Echo—Sacrifice two lands", the
// cumulative upkeeps that discard or sacrifice and an any-player offer
// that exiles from a graveyard ship `pay_cards`; the
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

// togglePayCard adds or removes `id`, never past the count. #6189: on a
// repeatable payment a click adds another copy of the same option (one
// choice per age counter, CR 702.24a) until the count is full, and a
// click on a full payment's option clears it.
export function togglePayCard(pc: PayCardsView, picks: string[], id: string): string[] {
  if (!pc.options.includes(id)) return picks;
  if (pc.repeats) {
    if (picks.length >= pc.count) return picks.filter((c) => c !== id);
    return [...picks, id];
  }
  if (picks.includes(id)) return picks.filter((c) => c !== id);
  if (picks.length >= pc.count) return picks;
  return [...picks, id];
}

// payCardsFull reports whether no more picks may be made: the count is
// reached. A repeatable payment's picker then clears an option instead
// (#6189, togglePayCard).
export function payCardsFull(pc: PayCardsView, picks: string[]): boolean {
  return picks.length >= pc.count;
}

// payCardPicksOf is how many times `id` pays, for the picker's badge.
export function payCardPicksOf(picks: string[], id: string): number {
  return picks.filter((c) => c === id).length;
}

// isPlayerPicks reports whether a payment's options are seats (#6189).
export function isPlayerPicks(pc: PayCardsView): boolean {
  return pc.picks === "players";
}

// payCardsOptions resolves the option IDs to the cards they name: the
// chooser's hand for a discard, their own graveyard for an exile (ADR
// 0187 §3), the battlefield for a sacrifice, a tap or a -1/-1 counter
// (#5746), in the server's order.
export function payCardsOptions(
  pc: PayCardsView,
  battlefield: CardView[] | undefined,
  chooser: PlayerView | undefined,
): CardView[] {
  // #6189: a player payment names seats, which are not cards at all.
  if (isPlayerPicks(pc)) return [];
  const pool =
    pc.action === "discard"
      ? (chooser?.hand?.cards ?? [])
      : pc.action === "exile_graveyard"
        ? (chooser?.graveyard?.cards ?? [])
        : (battlefield ?? []);
  const byID = new Map(pool.map((c) => [c.instance_id, c]));
  return pc.options.map((id) => byID.get(id)).filter((c): c is CardView => c !== undefined);
}

// payCardsVerb is the picker's instruction.
export function payCardsVerb(pc: PayCardsView): string {
  const n = pc.count === 1 ? "one card" : `${pc.count} cards`;
  if (pc.action === "discard") return `Choose ${n} to discard.`;
  if (pc.action === "exile_graveyard") return `Choose ${n} from your graveyard to exile.`;
  if (pc.action === "tap") return `Choose ${n} to tap.`;
  if (pc.action === "minus_one_counter") return `Choose ${n} to put a -1/-1 counter on.`;
  // #6189, the cumulative-upkeep payments.
  if (pc.action === "counter_opponent_creature")
    return pc.count === 1
      ? "Choose a creature an opponent controls to put a counter on."
      : `Choose ${pc.count} creatures opponents control to put a counter on — the same one more than once if you like.`;
  if (pc.action === "opponent_gains_life")
    return pc.count === 1
      ? "Choose an opponent to gain the life."
      : `Choose an opponent ${pc.count} times — the same one more than once if you like.`;
  if (pc.action === "opponent_creates_token")
    return pc.count === 1
      ? "Choose an opponent to create the token."
      : `Choose an opponent ${pc.count} times — the same one more than once if you like.`;
  if (pc.action === "gain_control_land")
    return pc.count === 1
      ? "Choose a land you don't control to gain control of."
      : `Choose ${pc.count} lands you don't control to gain control of.`;
  return `Choose ${n} to sacrifice.`;
}
