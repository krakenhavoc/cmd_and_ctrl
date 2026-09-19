// libraryTop — S42, CR 401.5: the top card of a library is sometimes
// visible, and when it is the pile should show it rather than a back.
//
// "You may look at the top card of your library any time" (Bolas's
// Citadel, Realmwalker) and "play with the top card of your library
// revealed" (Oracle of Mul Daya, Courser of Kruphix) are the two
// printed strengths. The client does not evaluate either: the SERVER
// decides who may see what and projects at most one readable card into
// a library zone, so the whole client-side rule is "did a readable
// card arrive". That keeps the two from drifting, and it means a
// library nobody may look at renders exactly as it always has.
//
// Deliberately NOT keyed on `castable_here`: Oracle of Mul Daya
// reveals the top card to the whole table while opening only LANDS,
// so "visible" and "playable by you" are different questions and only
// the first one decides what the pile face shows.

import type { CardView, ZoneView } from "./protocol";

// visibleLibraryTop returns the top card of a library when this viewer
// can read it, or null.
//
// The top is the LAST element — the same orientation the server uses,
// where PushTop appends. A zone whose last card came back redacted
// (empty name, `known_by_you` false) is not visible, and neither is an
// empty one.
export function visibleLibraryTop(zone: ZoneView | undefined): CardView | null {
  if (!zone || zone.cards.length === 0) return null;
  const top = zone.cards[zone.cards.length - 1];
  if (!top.known_by_you || !top.name) return null;
  return top;
}

// libraryTopPlayable reports whether the viewer may play the visible
// top card from where it sits — `castable_here`, which the server
// stamps from the same permission the cast path validates with, so the
// client can never render a button `cast_spell` would refuse.
//
// `ownerID` and `viewerID` are the second gate, and they are about
// WHOSE answer the bit is (#1035). A library is a per-seat pile, so
// `castable_here` on its top card is its owner's answer and it is
// public: two seats at a table with Oracles of Mul Daya both see the
// other's revealed top card marked playable, and only one of them may
// play it. A cast permission over ANOTHER seat's library top —
// Xanathar, Guild Kingpin's "you may play the top card of their
// library" — is the case that makes the gate more than an ownership
// check: the server computes that holder's own stamps for their frame
// alone and names them in the public `exile_play`, exactly as it does
// for a foreign graveyard cast, so the viewer who holds the grant gets
// the button and the rest of the table does not.
export function libraryTopPlayable(
  zone: ZoneView | undefined,
  viewerID?: string | null,
  ownerID?: string,
): boolean {
  const top = visibleLibraryTop(zone);
  if (top === null || top.castable_here !== true) return false;
  // Called with no seats to compare — the pre-#1035 signature, and
  // every caller looking at the viewer's OWN library. The bit is that
  // library owner's answer, which is this viewer's.
  if (viewerID === undefined || ownerID === undefined) return true;
  if (!viewerID) return false;
  return viewerID === ownerID || top.exile_play?.player === viewerID;
}
