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
// One question, one field (#1055). Until then the bit was the LIBRARY
// OWNER's answer and it was public, so this function took the viewer's
// and the owner's seat as a second gate: two seats at a table with
// Oracles of Mul Daya both saw the other's revealed top card marked
// playable, and only one of them could play it. Xanathar, Guild
// Kingpin's "you may play the top card of their library" is the case
// that made the gate more than an ownership check, and it is now the
// case the SERVER answers — it computes that holder's own stamps for
// their frame alone, so the viewer who holds the grant gets the bit
// and the rest of the table does not.
//
// Visibility is still a separate question, and visibleLibraryTop's
// note says why: Oracle of Mul Daya reveals the top card to the whole
// table while opening only lands.
export function libraryTopPlayable(zone: ZoneView | undefined): boolean {
  const top = visibleLibraryTop(zone);
  return top !== null && top.castable_here === true;
}

// libraryTopActionLabel is the verb for the affordance on a visible,
// playable library top, or null when there is nothing to offer —
// either the top isn't visible or `libraryTopPlayable` says no.
//
// A land is PLAYED, not cast (CR 305.1, CR 116.2a): the same wording
// nit the graveyard's flashback button has always had to mind
// (ZoneBrowserModal's castLabelFor). `castable_here` on a land already
// answers the land-play rule rather than the spell rules (#1407,
// #1441), so this reads the type line and nothing else — no second
// legality check.
export function libraryTopActionLabel(zone: ZoneView | undefined): "play" | "cast" | null {
  if (!libraryTopPlayable(zone)) return null;
  const top = visibleLibraryTop(zone);
  if (!top) return null;
  return (top.type_line ?? "").toLowerCase().includes("land") ? "play" : "cast";
}
