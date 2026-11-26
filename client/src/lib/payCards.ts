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

// isGraveyardGroups reports whether `pc` is a "cards from a single
// graveyard" payment (#6121, Jötun Grunt): picks from every seat's
// graveyard, in whole groups of `group_size` per graveyard.
export function isGraveyardGroups(pc: PayCardsView): boolean {
  return pc.action === "graveyard_library_bottom";
}

// graveyardOwners maps each card in a seat's graveyard to that seat.
function graveyardOwners(seats: PlayerView[] | undefined): Map<string, string> {
  const out = new Map<string, string>();
  for (const seat of seats ?? []) {
    for (const c of seat.graveyard?.cards ?? []) out.set(c.instance_id, seat.id);
  }
  return out;
}

// graveyardCounts is how many of `ids` each graveyard holds.
function graveyardCounts(ids: string[], seats: PlayerView[] | undefined): Map<string, number> {
  const owners = graveyardOwners(seats);
  const counts = new Map<string, number>();
  for (const id of ids) {
    const owner = owners.get(id) ?? "";
    counts.set(owner, (counts.get(owner) ?? 0) + 1);
  }
  return counts;
}

// canPayCards reports whether the chooser has enough to pay with. A
// grouped graveyard payment counts whole groups only, when the seats are
// known.
export function canPayCards(pc: PayCardsView, seats?: PlayerView[]): boolean {
  if (isTotalPower(pc)) {
    const best = pc.options.filter((id) => (pc.powers?.[id] ?? 0) > 0);
    return pickedPower(pc, best) >= (pc.total_power ?? 0);
  }
  if (isGraveyardGroups(pc) && seats) {
    const g = Math.max(1, pc.group_size ?? 1);
    let whole = 0;
    for (const n of graveyardCounts(pc.options, seats).values()) whole += Math.floor(n / g) * g;
    return whole >= pc.count;
  }
  return pc.options.length >= pc.count;
}

// payCardsReady reports whether `picks` is a whole payment. A grouped
// graveyard payment also needs each graveyard's picks to be whole groups,
// when the seats are known (the server checks it either way).
export function payCardsReady(pc: PayCardsView, picks: string[], seats?: PlayerView[]): boolean {
  if (isTotalPower(pc)) {
    return picks.length > 0 && pickedPower(pc, picks) >= (pc.total_power ?? 0);
  }
  if (isGraveyardGroups(pc) && seats) {
    const g = Math.max(1, pc.group_size ?? 1);
    for (const n of graveyardCounts(picks, seats).values()) if (n % g !== 0) return false;
  }
  return picks.length === pc.count;
}

// payCardsFull reports whether no more options may be picked: a fixed
// count reached. A total-power payment takes any number.
export function payCardsFull(pc: PayCardsView, picks: string[]): boolean {
  return !isTotalPower(pc) && picks.length >= pc.count;
}

// payCardPicksOf is how many times `id` pays, for the picker's badge
// (#6189: a repeatable payment may name one option twice).
export function payCardPicksOf(picks: string[], id: string): number {
  return picks.filter((c) => c === id).length;
}

// isPlayerPicks reports whether a payment's options are seats (#6189).
export function isPlayerPicks(pc: PayCardsView): boolean {
  return pc.picks === "players";
}

// payCardsProgress is the line under the picker: "1 / 2 chosen", or
// "power 7 / 12" for a total-power payment.
export function payCardsProgress(pc: PayCardsView, picks: string[]): string {
  if (isTotalPower(pc)) return `power ${pickedPower(pc, picks)} / ${pc.total_power ?? 0}`;
  return `${picks.length} / ${pc.count} chosen`;
}

// togglePayCard adds or removes `id`, never past the count.
export function togglePayCard(pc: PayCardsView, picks: string[], id: string): string[] {
  if (!pc.options.includes(id)) return picks;
  // #6189, CR 702.24a: on a repeatable payment a click adds another copy
  // of the same option until the count is full, and a click on a full
  // payment's option clears it.
  if (pc.repeats) {
    if (payCardsFull(pc, picks)) return picks.filter((c) => c !== id);
    return [...picks, id];
  }
  if (picks.includes(id)) return picks.filter((c) => c !== id);
  if (payCardsFull(pc, picks)) return picks;
  return [...picks, id];
}

// payCardsOptions resolves the option IDs to the cards they name: the
// chooser's hand for a discard, their own graveyard for an exile (ADR
// 0187 §3), every seat's graveyard for a grouped graveyard payment
// (#6121), the battlefield for a sacrifice, a tap or a -1/-1 counter
// (#5746), in the server's order.
export function payCardsOptions(
  pc: PayCardsView,
  battlefield: CardView[] | undefined,
  chooser: PlayerView | undefined,
  seats?: PlayerView[],
): CardView[] {
  // #6189: a player payment names seats, which are not cards at all.
  if (isPlayerPicks(pc)) return [];
  // #6121: a grouped graveyard payment picks out of every seat's graveyard.
  if (isGraveyardGroups(pc)) {
    const all = (seats ?? []).flatMap((seat) => seat.graveyard?.cards ?? []);
    const byID = new Map(all.map((c) => [c.instance_id, c]));
    return pc.options.map((id) => byID.get(id)).filter((c): c is CardView => c !== undefined);
  }
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
  if (isGraveyardGroups(pc)) {
    const g = pc.group_size ?? 1;
    return g > 1
      ? `Choose ${n} from graveyards, ${g} at a time from a single graveyard, to put on the bottom of their owners' libraries.`
      : `Choose ${n} from graveyards to put on the bottom of their owners' libraries.`;
  }
  if (pc.action === "tap") return `Choose ${n} to tap.`;
  if (pc.action === "minus_one_counter") return `Choose ${n} to put a -1/-1 counter on.`;
  if (pc.action === "hand_to_library_top") return `Choose ${n} to put on top of your library.`;
  // #6189, the cumulative-upkeep payments.
  if (pc.action === "counter_opponent_creature")
    return pc.count === 1
      ? "Choose a creature an opponent controls to put a counter on."
      : `Choose ${pc.count} creatures opponents control to put a counter on — the same one more than once if you like.`;
  if (pc.action === "opponent_gains_life" || pc.action === "opponent_creates_token")
    return pc.count === 1
      ? "Choose the opponent."
      : `Choose an opponent ${pc.count} times — the same one more than once if you like.`;
  if (pc.action === "gain_control_land")
    return pc.count === 1
      ? "Choose a land you don't control to gain control of."
      : `Choose ${pc.count} lands you don't control to gain control of.`;
  if (isTotalPower(pc))
    return `Choose any number to sacrifice with total power ${pc.total_power ?? 0} or greater.`;
  return `Choose ${n} to sacrifice.`;
}
