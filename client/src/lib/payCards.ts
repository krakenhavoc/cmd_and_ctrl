// payCards.ts — ADR 0108 §5: a pay-unless prompt whose payment is not
// mana. "Echo—Discard a card", "Echo—Sacrifice two lands", the
// cumulative upkeeps that discard or sacrifice and an any-player offer
// that exiles from a graveyard ship `pay_cards`; the
// "Pay" answer names exactly `count` of its options as `card_ids`. The
// payer may only answer "Pay" once they have picked that many, and when
// there are fewer options than that they cannot pay at all (CR 118.3).

import type { CardView, PayCardsView, PlayerView } from "./protocol";

// isTotalPower reports whether `pc` is a "sacrifice any number with total
// power N or greater" payment (#6187): no fixed count.
export function isTotalPower(pc: PayCardsView): boolean {
  return pc.action === "sacrifice_total_power";
}

// pickedPower is the total power of `ids` for a total-power payment.
export function pickedPower(pc: PayCardsView, ids: string[]): number {
  return ids.reduce((sum, id) => sum + (pc.powers?.[id] ?? 0), 0);
}

// canPayCards reports whether the chooser has enough to pay with.
export function canPayCards(pc: PayCardsView): boolean {
  if (isTotalPower(pc)) {
    const best = pc.options.filter((id) => (pc.powers?.[id] ?? 0) > 0);
    return pickedPower(pc, best) >= (pc.total_power ?? 0);
  }
  return pc.options.length >= pc.count;
}

// payCardsReady reports whether `picks` is a whole payment.
export function payCardsReady(pc: PayCardsView, picks: string[]): boolean {
  if (isTotalPower(pc)) {
    return picks.length > 0 && pickedPower(pc, picks) >= (pc.total_power ?? 0);
  }
  return picks.length === pc.count;
}

// payCardsFull reports whether no more options may be picked: a fixed
// count reached. A total-power payment takes any number.
export function payCardsFull(pc: PayCardsView, picks: string[]): boolean {
  return !isTotalPower(pc) && picks.length >= pc.count;
}

// payCardsProgress is the line under the picker: "1 / 2 chosen", or
// "power 7 / 12" for a total-power payment.
export function payCardsProgress(pc: PayCardsView, picks: string[]): string {
  if (isTotalPower(pc)) return `power ${pickedPower(pc, picks)} / ${pc.total_power ?? 0}`;
  return `${picks.length} / ${pc.count} chosen`;
}

// togglePayCard adds or removes `id`, never past the count.
export function togglePayCard(pc: PayCardsView, picks: string[], id: string): string[] {
  if (picks.includes(id)) return picks.filter((c) => c !== id);
  if (payCardsFull(pc, picks) || !pc.options.includes(id)) return picks;
  return [...picks, id];
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
  const pool =
    pc.action === "discard" || pc.action === "hand_to_library_top"
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
  if (pc.action === "hand_to_library_top") return `Choose ${n} to put on top of your library.`;
  if (isTotalPower(pc))
    return `Choose any number to sacrifice with total power ${pc.total_power ?? 0} or greater.`;
  return `Choose ${n} to sacrifice.`;
}
