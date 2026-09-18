import { describe, expect, it } from "vitest";

import { previewableCard } from "./stackPreview";
import type { CardView } from "./protocol";

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "c1", name: "", owner: "me", controller: "me", ...extras };
}

describe("previewableCard", () => {
  // #697: the overlay tested `known_by_you === false`, and the server
  // never sends that value — `known_by_you` is omitempty, so "not a
  // knower" arrives as an ABSENT field. The guard never fired, and the
  // zoom panel opened blank on a card the server had already stripped.
  it("refuses a card whose known_by_you is absent (the wire shape)", () => {
    expect(previewableCard(card({ face_down: true }))).toBeNull();
  });

  it("refuses a redacted spell on the stack — no name, no known_by_you", () => {
    expect(previewableCard(card({ instance_id: "spell", name: "" }))).toBeNull();
  });

  it("refuses an explicit known_by_you: false", () => {
    expect(previewableCard(card({ known_by_you: false }))).toBeNull();
  });

  it("refuses a missing card — an ability whose source has left", () => {
    expect(previewableCard(undefined)).toBeNull();
  });

  it("shows a card the viewer knows", () => {
    const c = card({ name: "Lightning Bolt", known_by_you: true });
    expect(previewableCard(c)).toBe(c);
  });

  // CR 708.5 / CR 702.143d: a face-down object the viewer is allowed
  // to look at is their own card, and hiding it from them would be
  // hiding their own card.
  it("shows a face-down card the viewer may look at", () => {
    const c = card({
      name: "Griselbrand",
      face_down: true,
      face_down_kind: "manifested",
      face_visible: true,
      known_by_you: true,
    });
    expect(previewableCard(c)).toBe(c);
  });

  it("refuses an opponent's face-down permanent", () => {
    expect(
      previewableCard(card({ face_down: true, face_down_kind: "morphed", controller: "them" })),
    ).toBeNull();
  });
});
