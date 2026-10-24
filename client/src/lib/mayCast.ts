// mayCast.ts — the words on a "you may cast it" offer (ADR 0099 §7, ADR
// 0066's 2026-10-23 amendment).
//
// Two prompt kinds carry one. Every keyword cast is a "resolution_cast"
// asked inside the resolution that offers it (CR 608.2g): cascade and
// discover since #4243, and rebound, suspend, madness, the exile-until
// shape, the paid "if you don't" cast and the free-or-hand offers since
// #4250. What declining does is different for each, so the server names the
// rule in `may_cast_keyword` and the two branches in `accept_label` /
// `decline_label`, and keywordCastWords turns the keyword into the sheet's
// rule tag, the verb of its heading and the hint.
//
// #4250, part 2: an effect's own "you may cast it" is the same prompt, under
// one of three keywords: `hand_cast` (a spell from your hand, free: Buster
// Sword, the Expertises, Aetherflux Conduit), `free_cast` (a card in exile,
// a graveyard or a library, free: Etali, Torrential Gearhulk, Aetherworks
// Marvel) and `paid_cast` (the same at a price: Conduit of Worlds, Amped
// Raptor's energy, The Infamous Cruelclaw's discard). The decline label
// says where an uncast card stays.
//
// What is left of "may_cast" is the land play of "look at the top card,
// you may play it" (`play_or_hand`, #3801) and the offers that are only
// their card's text (hideaway, Malcolm); mayCastCopy words those.

export interface MayCastCopy {
  source: string;
  hint: string;
  accept: string;
  decline: string;
}

export function mayCastCopy(
  keyword: string | undefined,
  acceptLabel?: string,
  declineLabel?: string,
): MayCastCopy {
  let copy: MayCastCopy;
  switch (keyword) {
    case "play_or_hand":
      // #3801: "look at the top card of your library. You may play a land
      // from the top ... If you don't, put that card into your hand"
      // (Ziatora's Envoy). Accepting plays the land at once; there is no
      // cast to open.
      copy = {
        source: "play · CR 608.2g",
        hint: "Play it from the top of your library now, or put it into your hand.",
        accept: "Play it",
        decline: "Put it into your hand",
      };
      break;
    default:
      copy = {
        source: "cast · CR 608.2g",
        hint: "Cast it without paying its mana cost, or don't.",
        accept: "Cast it free",
        decline: "Don't cast it",
      };
  }
  if (acceptLabel) copy.accept = acceptLabel;
  if (declineLabel) copy.decline = declineLabel;
  return copy;
}

// KeywordCastWords is how a keyword's resolution_cast speaks: `src` is the
// rule tag on the sheet, `how` finishes "you may cast <card> ..." in the
// heading, and `ifNot` is what happens to the card if it is not cast.
// `many`, when set, stands for the card in the heading when the offer names
// more than one ("a spell from your hand").
export interface KeywordCastWords {
  src: string;
  how: string;
  ifNot: string;
  many?: string;
}

// whereItStays is what a declined effect cast leaves behind, read off the
// server's decline label (#4250).
function whereItStays(declineLabel: string | undefined): string {
  switch (declineLabel) {
    case "Put them on the bottom":
      return "it goes to the bottom of your library with the rest";
    case "Leave it in the graveyard":
      return "it stays in the graveyard";
    case "Keep it in your hand":
    case "Cast no more":
      return "it stays in your hand";
    default:
      return "it stays in exile";
  }
}

const FREE = "without paying its mana cost";

// keywordCastWords is the wording for a resolution_cast that names a
// keyword, or null for one that does not (a copy of a card, #2947). The
// exile-until shape has two declines (the bottom, or "Leave it in exile"
// for Guff Rewrites History, #4205), told apart by the server's label.
export function keywordCastWords(
  keyword: string | undefined,
  declineLabel?: string,
): KeywordCastWords | null {
  switch (keyword) {
    case "cascade":
      return {
        src: "CR 702.85",
        how: FREE,
        ifNot:
          "it goes to the bottom of your library; the rest of the exiled cards go to the bottom either way",
      };
    case "discover":
      return {
        src: "CR 701.57",
        how: FREE,
        ifNot:
          "it goes into your hand; the rest of the exiled cards go to the bottom of your library either way",
      };
    case "exile_until":
      return {
        src: "CR 608.2g",
        how: FREE,
        ifNot:
          (declineLabel === "Leave it in exile"
            ? "it stays in exile"
            : "it goes to the bottom of your library") +
          "; the rest of the exiled cards go to the bottom either way",
      };
    case "free_or_hand":
      return { src: "CR 608.2g", how: FREE, ifNot: "it goes into your hand" };
    case "paid_or_else":
      // #3510: "you may cast that card. If you don't, ..." (Chandra, Torch
      // of Defiance). The cast is paid at the card's own cost, timing aside;
      // the decline label names what not casting does.
      return {
        src: "CR 608.2g",
        how: "paying its mana cost",
        ifNot: "it stays in exile, and what the decline names happens",
      };
    case "suspend":
      return { src: "CR 702.62", how: FREE, ifNot: "it stays in exile" };
    case "rebound":
      return { src: "CR 702.88", how: FREE, ifNot: "it stays in exile" };
    case "madness":
      return {
        src: "CR 702.35",
        how: "for its madness cost",
        ifNot: "it goes into your graveyard",
      };
    case "hand_cast":
      return {
        src: "CR 608.2g",
        how: FREE,
        ifNot: "it stays in your hand",
        many: "a spell from your hand",
      };
    case "free_cast":
      return {
        src: "CR 608.2g",
        how: FREE,
        ifNot: whereItStays(declineLabel),
        many: "a spell from among them",
      };
    case "paid_cast":
      return {
        src: "CR 608.2g",
        how: "paying its cost",
        ifNot: whereItStays(declineLabel),
        many: "a spell from among them",
      };
    default:
      return null;
  }
}
