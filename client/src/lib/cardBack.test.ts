import { describe, it, expect } from "vitest";

import { showsCardBack } from "./cardBack";
import type { CardView } from "./protocol";

// cardBack.test.ts — #95. The server redacts a face-down card the
// viewer does not know down to its game state, and omits
// `known_by_you` rather than sending `false`. Card.svelte has to draw
// a back from that shape, not a blank front.

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "c1", name: "", owner: "me", controller: "me", ...extras };
}

describe("showsCardBack", () => {
  it("draws a back for a face-down card whose known_by_you is absent (the wire shape)", () => {
    expect(showsCardBack(card({ face_down: true }))).toBe(true);
  });

  it("draws the face for a face-down card the viewer knows", () => {
    expect(showsCardBack(card({ name: "Forest", face_down: true, known_by_you: true }))).toBe(
      false,
    );
  });

  it("draws the face for an ordinary known card", () => {
    expect(showsCardBack(card({ name: "Forest", known_by_you: true }))).toBe(false);
  });

  // #1194 / ADR 0082 decision 9. A morph is the first face-down card
  // that lives on the BATTLEFIELD, and the rule needs nothing new for
  // it: the one predicate is zone-independent, and the CR 708.2 body
  // the wire ships beside the back (a nameless 2/2 Creature) is not
  // evidence of a face, so it must not talk the renderer out of the
  // back.
  it("draws a back for a morph on the battlefield, body and all", () => {
    expect(
      showsCardBack(
        card({
          face_down: true,
          face_down_kind: "morphed",
          type_line: "Creature",
          power: 2,
          toughness: 2,
        }),
      ),
    ).toBe(true);
  });

  it("draws the face of a morph for the controller who may look at it", () => {
    expect(
      showsCardBack(
        card({
          name: "Willbender",
          face_down: true,
          face_down_kind: "morphed",
          known_by_you: true,
          face_visible: true,
        }),
      ),
    ).toBe(false);
  });

  it("keeps the explicit-false and parent-prop arms", () => {
    expect(showsCardBack(card({ known_by_you: false }))).toBe(true);
    expect(showsCardBack(card({ name: "Forest", known_by_you: true }), true)).toBe(true);
  });
});
