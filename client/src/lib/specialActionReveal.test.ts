// specialActionReveal.test.ts — #6249. A turn-face-up row whose morph
// cost is "Reveal a green card in your hand" (Temur Charger, CR 702.37e)
// carries the hand cards that can pay it as `reveal_options`; the Board
// asks which one (or takes the only one) and sends it as `reveal_id`.

import { describe, it, expect } from "vitest";
import type { CardView, GameView } from "./protocol";
import { specialActionItems } from "./contextMenu.logic";
import { revealOptionCards, specialActionNeedsReveal, withReveal } from "./specialActionReveal";

const charger = {
  instance_id: "tc-1",
  name: "",
  owner: "me",
  controller: "me",
  face_down: true,
  special_actions: [
    {
      kind: "turn_face_up",
      label: "Turn face up — reveal a green card",
      available: true,
      route: "morph",
      reveal_options: ["g-1", "g-2"],
    },
  ],
} as unknown as CardView;

const morph = {
  instance_id: "mo-1",
  name: "Willbender",
  owner: "me",
  controller: "me",
  special_actions: [
    {
      kind: "turn_face_up",
      label: "Turn face up {1}{U}",
      cost: "{1}{U}",
      available: true,
      route: "morph",
    },
  ],
} as unknown as CardView;

const hand = [
  { instance_id: "g-1", name: "Grizzly Bears" },
  { instance_id: "r-1", name: "Raging Goblin" },
  { instance_id: "g-2", name: "Llanowar Elves" },
] as unknown as CardView[];

const view = {
  battlefield: { cards: [charger, morph] },
  seats: [{ id: "me", hand: { cards: hand } }],
} as unknown as GameView;

function rowParams(card: CardView): Record<string, unknown> {
  return specialActionItems(card, "me")[0].action!.params as Record<string, unknown>;
}

describe("a turn-face-up whose morph cost is a reveal", () => {
  it("finds the row that needs a reveal, and only that one", () => {
    const found = specialActionNeedsReveal(view, rowParams(charger));
    expect(found?.card.instance_id).toBe("tc-1");
    expect(specialActionNeedsReveal(view, rowParams(morph))).toBeNull();
    // A payload that already names its card is sent as it is.
    expect(specialActionNeedsReveal(view, withReveal(rowParams(charger), "g-1"))).toBeNull();
  });

  it("offers the viewer's matching hand cards in hand order", () => {
    const row = specialActionNeedsReveal(view, rowParams(charger))!.row;
    expect(revealOptionCards(view, "me", row).map((c) => c.instance_id)).toEqual(["g-1", "g-2"]);
    expect(revealOptionCards(view, null, row)).toEqual([]);
  });

  it("sends the card revealed with the rest of the payload", () => {
    expect(withReveal(rowParams(charger), "g-2")).toMatchObject({
      card_id: "tc-1",
      kind: "turn_face_up",
      reveal_id: "g-2",
    });
  });
});
