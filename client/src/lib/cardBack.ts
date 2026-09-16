// cardBack — the one rule for "does this card render as a card back".
// Lifted out of Card.svelte so it can be tested without mounting the
// component.

import type { CardView } from "./protocol";

// showsCardBack is true when the viewer has nothing to draw but a
// back:
//
//   - the parent says so (`faceDown` prop — Hand.svelte for an
//     opponent's unrevealed cards);
//   - the card is face down (CR 406.3 / 708) and the viewer is not a
//     knower of it (#95). The server strips every identifying field
//     from such a card, so without this arm it rendered as a blank
//     front — Necropotence's face-down exiles in the exile browser.
//     `known_by_you` is omitempty on the wire, so "not a knower"
//     arrives as an ABSENT field, never as `false`, and the check has
//     to be `!== true`;
//   - a locally built CardView says `known_by_you: false` outright.
//
// A face-down card the viewer DOES know still renders its face: that
// is the player who knows what they exiled or foretold, and hiding it
// from them would be hiding their own card.
export function showsCardBack(card: CardView, faceDown = false): boolean {
  if (faceDown) return true;
  if (card.face_down && card.known_by_you !== true) return true;
  return card.known_by_you === false;
}
