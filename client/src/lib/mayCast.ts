// mayCast.ts — the words on a "may_cast" prompt (ADR 0099 §7).
//
// One prompt kind serves cascade, discover, suspend and madness, and
// what declining does is different for each: cascade bottoms the card,
// discover puts it into your hand, suspend leaves it in exile and
// madness puts it into your graveyard. The server names the rule in
// `may_cast_keyword` and the two branches in `accept_label` /
// `decline_label`; this module turns that into the tag, the hint and
// the buttons, falling back to generic copy for an offer that is only
// its card's text (hideaway, Malcolm).

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
    case "discover":
      copy = {
        source: "discover · CR 701.57",
        hint:
          "Cast it for nothing now, or put it into your hand. If you choose to cast it and then pass " +
          "without casting, it goes into your hand. The rest of the cards go to the bottom of your library.",
        accept: "Cast it free",
        decline: "Put it into your hand",
      };
      break;
    case "cascade":
      copy = {
        source: "cascade · CR 702.85",
        hint:
          "Cast it for nothing now, or put it on the bottom of your library with the rest. If you choose " +
          "to cast it and then pass without casting, it goes to the bottom.",
        accept: "Cast it free",
        decline: "Put it on the bottom",
      };
      break;
    case "suspend":
      copy = {
        source: "suspend · CR 702.62",
        hint: "The last time counter is gone. Cast it for nothing now, or leave it in exile.",
        accept: "Cast it free",
        decline: "Leave it in exile",
      };
      break;
    case "madness":
      copy = {
        source: "madness · CR 702.35",
        hint: "Cast it for its madness cost, or put it into your graveyard.",
        accept: "Cast it",
        decline: "Put it into your graveyard",
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
