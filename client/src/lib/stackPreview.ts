// stackPreview — the one rule for "may the stack overlay zoom this
// card". Lifted out of StackOverlay.svelte so it can be tested without
// mounting the component, exactly as cardBack.ts was (#95, #646).
//
// #697: the overlay used to test `c.known_by_you === false`, and the
// server never sends that value — the field is `omitempty`, so "not a
// knower" arrives as an ABSENT field. The guard therefore never fired,
// and a spell or ability source the viewer cannot read still opened
// the zoom panel, which then showed a blank card because the server
// had stripped its name and art. Card.svelte had the same bug and
// #646 fixed it in cardBack.ts; this is the same fix in the same
// place, so the overlay and the card can never disagree about what
// the viewer is allowed to see.

import { showsCardBack } from "./cardBack";
import type { CardView } from "./protocol";

// previewableCard is the card the overlay may show for an item, or
// null when there is nothing the viewer is allowed to look at.
//
// A card that renders as a BACK has nothing to zoom: the server has
// redacted it down to game state, so the panel would be blank. That
// includes a face-down object the viewer may not look at — a morph or
// disguise cast face down (CR 708.4), a foretold card that is not
// theirs — and any unknown ability source.
//
// A face-down object the viewer MAY look at is not a back and does
// zoom: that is the controller of their own morph, and hiding it from
// them would be hiding their own card.
export function previewableCard(card: CardView | undefined): CardView | null {
  if (!card) return null;
  // Not a knower: the server has already stripped the name and the
  // art, so there is nothing to zoom and the panel would be blank.
  // `known_by_you` is `omitempty` on the wire, so "not a knower"
  // arrives as an ABSENT field and the test has to be `!== true` — the
  // whole of #697.
  if (card.known_by_you !== true) return null;
  // And never zoom anything Card.svelte would draw as a back, so the
  // overlay and the card cannot disagree about what the viewer may
  // see. Redundant today (a knower is never drawn as a back) and kept
  // so it stays true if either rule moves.
  if (showsCardBack(card)) return null;
  return card;
}
