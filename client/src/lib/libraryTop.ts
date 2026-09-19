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
export function libraryTopPlayable(zone: ZoneView | undefined): boolean {
  const top = visibleLibraryTop(zone);
  return top !== null && top.castable_here === true;
}
