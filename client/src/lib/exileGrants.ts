// exileGrants.ts — the words for an exile grant (#2559): when it ends,
// and who holds it when that is not the viewer.
//
// Memory Vessel and Rocco, Street Chef hand EVERY player a permission
// over the cards they exiled themselves, so a seat's exile pile can hold
// cards its owner may play, and the rest of the table needs to see that
// as plainly as the owner does. Everything here reads the public
// `exile_play` on the viewer's own frame; nothing is a rule.

import type { CardView, ExilePlayView } from "./protocol";

/** Who a player ID is, for a sentence: "you" for the viewer. */
export type NameOf = (playerID: string) => string | undefined;

function possessive(name: string): string {
  return name === "you" ? "your" : `${name}'s`;
}

function who(id: string | undefined, nameOf: NameOf, viewerID: string | null): string {
  if (!id) return "its controller";
  if (id === viewerID) return "you";
  return nameOf(id) ?? "a player";
}

/**
 * grantWindowText is when the grant ends, as a phrase: "until end of
 * turn", "until Bo's next turn", "until your next end step". A grant
 * from a server without the field reads "until end of turn", which is
 * what every grant said before #2559.
 */
export function grantWindowText(
  grant: ExilePlayView | null | undefined,
  nameOf: NameOf,
  viewerID: string | null,
): string {
  const p = () => possessive(who(grant?.until_player, nameOf, viewerID));
  switch (grant?.until) {
    case "end_of_next_turn":
      return `until the end of ${p()} next turn`;
    case "next_turn":
      return `until ${p()} next turn`;
    case "next_end_step":
      return `until ${p()} next end step`;
    case "while_exiled":
      return "while it stays exiled";
    case "until_another":
      return "until another card is exiled with its source";
    default:
      return "until end of turn";
  }
}

/**
 * grantHolderLine is the bystander's label for a card someone ELSE may
 * play: "Ana may play it until Bo's next turn". Null when the card has
 * no grant or the grant is the viewer's own (the viewer has a button
 * for that instead).
 */
export function grantHolderLine(
  card: CardView,
  nameOf: NameOf,
  viewerID: string | null,
): string | null {
  const grant = card.exile_play;
  if (!grant || grant.player === viewerID) return null;
  const holder = nameOf(grant.player) ?? "Another player";
  const verb = grant.cast_only ? "cast" : "play";
  return `${holder} may ${verb} it ${grantWindowText(grant, nameOf, viewerID)}`;
}

/**
 * heldPlayableCount is how many of `cards` the player `holderID` holds a
 * grant over: the count a seat's exile chip shows for the table.
 */
export function heldPlayableCount(cards: CardView[] | undefined, holderID: string): number {
  let n = 0;
  for (const c of cards ?? []) {
    if (c.exile_play?.player === holderID) n++;
  }
  return n;
}
