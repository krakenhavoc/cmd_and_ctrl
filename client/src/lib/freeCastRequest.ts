import { guardedWritable } from "./guardedStore";
import type { AlternativeCostView, CardView, PendingChoiceView } from "./protocol";

// ADR 0099 §7. Answering "Cast it free" to a discover's or a cascade's
// may_cast prompt only opens the free cast: the server stamps a grant on
// the exiled card, and the cast itself is an ordinary cast_spell. The
// prompt lives in ChoicePromptModal and the one cast entry point
// (handlePlayCard, #874) lives in Board, so the modal leaves the card's
// instance id here and Board starts the cast chain — face picker,
// targets, modes — once a snapshot shows the card in exile and castable.
//
// The window closes on the player's next pass, so a request that never
// becomes castable (no legal target, say) is dropped rather than kept:
// the player passes, and the card goes where the keyword sends it.
export const freeCastRequest = guardedWritable<string | null>(null, "freeCastRequest");

// MayCastKeywordsThatOpenACast are the may_cast keywords whose accept
// branch is a free cast from exile, as opposed to madness's paid cast
// or an offer the card's own text words.
export const mayCastKeywordsThatOpenACast: ReadonlySet<string> = new Set([
  "discover",
  "cascade",
  "exile_until",
  "free_or_hand",
  "suspend",
  "rebound",
]);

// freeCastTarget is Board's half: the exiled card a pending request
// names, when it is there and the viewer may cast it now. `undefined`
// means "not yet" (the snapshot carrying the grant has not arrived);
// `null` means the request is stale and should be dropped (the card is
// in exile but nobody may cast it).
export function freeCastTarget(
  exile: CardView[] | undefined,
  id: string | null,
): CardView | null | undefined {
  if (!id) return null;
  const card = exile?.find((c) => c.instance_id === id);
  if (!card) return undefined;
  return card.castable_here ? card : null;
}

// #3270 (CR 608.2g): "you may cast a spell from your hand without paying
// its mana cost" (Buster Sword, Kari Zev's and Rishkar's Expertise,
// Surtland Elementalist, Great Intelligence's Plan). The pick prompt is a
// choose_cards that declares `choose_destination: "cast_free"`; answering
// it with a card grants a pass-closed, any-timing free cast of that HAND
// card, and the cast is part of the resolution, so the client starts the
// cast chain itself instead of leaving the player to find the card. The
// modal only leaves the picked card's instance id here; Board starts the
// chain once the snapshot carrying the grant shows the free offer on the
// card in hand.
export const handFreeCastRequest = guardedWritable<string | null>(null, "handFreeCastRequest");

// CHOOSE_CAST_FREE is the pick prompt's `choose_destination`.
export const CHOOSE_CAST_FREE = "cast_free";

// GRANTED_COST_KEY is the alternative-cost key a hand grant that names a
// price is claimed under (game.AltCostKeyGrantedCost).
export const GRANTED_COST_KEY = "granted-cost";

// isFreeCastPick is the prompt whose answer opens a cast.
export function isFreeCastPick(choice: PendingChoiceView | null | undefined): boolean {
  return choice?.kind === "choose_cards" && choice.choose_destination === CHOOSE_CAST_FREE;
}

// freeCastOffer is the granted free offer on a hand card, when the
// engine would accept the cast right now. Timing is judged by the server
// per offer (`timing_closed`), so a sorcery in the combat damage step
// still shows an open offer: CR 608.2g.
export function freeCastOffer(card: CardView): AlternativeCostView | undefined {
  return (card.alternative_costs ?? []).find(
    (o) => o.key === GRANTED_COST_KEY && o.timing_closed !== true,
  );
}

// handFreeCastTarget is Board's half for a hand pick: the card a pending
// request names, once the viewer may cast it free. `undefined` means
// "not yet" — the pick is still open, or the snapshot carrying the grant
// has not arrived. `null` means the request is stale and is dropped: the
// card left the hand or never got an offer (a pick that could not be
// granted), and the pick prompt is gone. `pickOpen` is whether the
// viewer still owes the free-cast prompt.
export function handFreeCastTarget(
  hand: CardView[] | undefined,
  id: string | null,
  pickOpen: boolean,
): { card: CardView; key: string } | null | undefined {
  if (!id) return null;
  const card = hand?.find((c) => c.instance_id === id);
  const offer = card ? freeCastOffer(card) : undefined;
  if (card && offer) return { card, key: offer.key };
  return pickOpen ? undefined : null;
}
