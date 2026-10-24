import { guardedWritable } from "./guardedStore";
import type { AlternativeCostView, CardView, PendingChoiceView } from "./protocol";

// ADR 0099 §7. Answering a free-cast pick only opens the free cast: the
// server stamps a grant on the card, and the cast itself is an ordinary
// cast_spell. The prompt lives in ChoicePromptModal and the one cast entry
// point (handlePlayCard, #874) lives in Board, so the modal leaves the
// card's instance id here and Board starts the cast chain — face picker,
// targets, modes — once a snapshot shows the card castable.
//
// The window closes on the player's next pass, so a request that never
// becomes castable (no legal target, say) is dropped rather than kept:
// the player passes, and the card goes where the keyword sends it.
export const freeCastRequest = guardedWritable<string | null>(null, "freeCastRequest");

// #4250: no may_cast keyword opens a cast from here any more. Every
// keyword cast (cascade, discover, rebound, suspend, madness, the
// exile-until shape, the paid "if you don't" cast and the free-or-hand
// offers) is a resolution_cast, cast from ResolutionCastSheet inside the
// resolution that offers it. freeCastRequest is set by the library pick
// below (Aetherworks Marvel).

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

// freeCastLocation is freeCastTarget over the two places an offered card
// can sit: exile (discover, cascade) and the viewer's own library, where a
// "look at the top card, you may cast it free" offer (#3801, Ziatora's
// Envoy) leaves it. The library copy of a card is only on the wire for a
// seat that has looked at it, so a request for a card in neither place is
// still "not yet".
//
// `offerOpen` is whether the may_cast prompt that offered the card is still
// on the frame (mayCastOfferOpen). While it is, the answer has not landed
// and the card has no grant yet, so that is "not yet" too. #3490: a rebound
// or suspend card is in exile BEFORE the answer and not castable, so the
// request used to be dropped as stale on the very frame "Cast it free" was
// clicked on. Nothing opened, and the card stayed in exile.
export function freeCastLocation(
  exile: CardView[] | undefined,
  library: CardView[] | undefined,
  id: string | null,
  offerOpen = false,
): { card: CardView; zone: "exile" | "library" } | null | undefined {
  if (!id) return null;
  if (offerOpen) return undefined;
  const inExile = exile?.find((c) => c.instance_id === id);
  if (inExile) return inExile.castable_here ? { card: inExile, zone: "exile" } : null;
  const inLibrary = library?.find((c) => c.instance_id === id);
  if (inLibrary) return inLibrary.castable_here ? { card: inLibrary, zone: "library" } : null;
  return undefined;
}

// mayCastOfferOpen reports whether the viewer still owes the may_cast
// prompt offering card `id`: true on the frame "Cast it free" was clicked
// on, false once the answer's frame (which carries the grant) arrives.
export function mayCastOfferOpen(
  pending: PendingChoiceView[] | undefined,
  id: string | null,
  viewerID: string | null,
): boolean {
  if (!id) return false;
  return (pending ?? []).some(
    (c) => c.kind === "may_cast" && c.may_cast_card === id && (!viewerID || c.chooser === viewerID),
  );
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

// CHOOSE_CAST_FREE_LIBRARY is the same pick over cards the viewer looked at
// on top of their own library ("look at the top six cards of your library.
// You may cast a spell from among them without paying its mana cost",
// Aetherworks Marvel, #2110). The picked card stays in the library, so the
// request goes to freeCastRequest (which Board resolves against the exile
// and the viewer's library), not to the hand request.
export const CHOOSE_CAST_FREE_LIBRARY = "cast_free_library";

// #3955 (ADR 0198's 2026-10-16 amendment): "cast any number of Eldrazi
// cards you own from outside the game without paying their mana costs"
// (Spawnsire of Ulamog) and "you may play a card you own from outside the
// game this turn" (Wish). The pick is a choose_cards over the viewer's
// sideboard, or an outside_collection over the card database, and its
// answer leaves the picked cards in the sideboard with a cast permission
// on each. A database pick names cards that have no instance id until
// the server creates them, so the request is a flag rather than an id:
// once the answer lands, Board opens the sideboard tab, where every
// granted card carries its cast button.
export const outsideCastRequest = guardedWritable<boolean>(false, "outsideCastRequest");

// CHOOSE_CAST_FREE_OUTSIDE and CHOOSE_PLAY_OUTSIDE are the pick prompts'
// `choose_destination`.
export const CHOOSE_CAST_FREE_OUTSIDE = "cast_free_outside";
export const CHOOSE_PLAY_OUTSIDE = "play_outside";

// isOutsideCastPick is the prompt whose answer opens a cast or a play from
// outside the game.
export function isOutsideCastPick(choice: PendingChoiceView | null | undefined): boolean {
  if (!choice) return false;
  if (choice.kind !== "choose_cards" && choice.kind !== "outside_collection") return false;
  return (
    choice.choose_destination === CHOOSE_CAST_FREE_OUTSIDE ||
    choice.choose_destination === CHOOSE_PLAY_OUTSIDE
  );
}

// outsideCastReady is Board's half: the request is ready to act on once
// the viewer owes no such pick any more (the answer has landed, so the
// frame carries the permissions) and the sideboard holds a card the viewer
// may cast. `null` means the request is stale and is dropped (the pick
// was declined, or nothing was granted); `undefined` means "not yet".
export function outsideCastReady(
  pending: PendingChoiceView[] | undefined,
  sideboard: CardView[] | undefined,
  viewerID: string | null,
): boolean | null | undefined {
  const open = (pending ?? []).some(
    (c) => isOutsideCastPick(c) && (!viewerID || c.chooser === viewerID),
  );
  if (open) return undefined;
  return (sideboard ?? []).some((c) => c.castable_here === true) ? true : null;
}

// isLibraryFreeCastPick is the library-looking free-cast pick.
export function isLibraryFreeCastPick(choice: PendingChoiceView | null | undefined): boolean {
  return choice?.kind === "choose_cards" && choice.choose_destination === CHOOSE_CAST_FREE_LIBRARY;
}

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
