import { describe, it, expect } from "vitest";

import { cardAsFace, needsFacePicker } from "./faces";
import { cardImageURL, scryfallImageURL } from "./cardImage";
import { applyCastChoices } from "./targeting";
import type { CardView } from "./protocol";

// faces.test.ts — the client half of ADR 0034.
//
// What matters here is the ADDITIVE promise: a single-faced card is
// untouched by every one of these helpers, and a multi-face card's
// flat fields still mean "the active face's", so nothing downstream
// had to change to keep working.

function seaGate(active = 0): CardView {
  const faces = [
    {
      name: "Sea Gate Restoration",
      type_line: "Sorcery",
      mana_cost: "{4}{U}{U}{U}",
      oracle_text: "Draw cards equal to the number of cards in your hand.",
      image: "/cards/sg/image?face=0",
    },
    {
      name: "Sea Gate, Reborn",
      type_line: "Land",
      oracle_text: "As this enters, you may pay 3 life. If you don't, it enters tapped.",
      image: "/cards/sg/image?face=1",
    },
  ];
  const face = faces[active];
  return {
    instance_id: "c1",
    name: face.name,
    owner: "p0",
    controller: "p0",
    scryfall_id: "sg",
    type_line: face.type_line,
    mana_cost: face.mana_cost,
    layout: "modal_dfc",
    faces,
    active_face: active === 0 ? undefined : active,
  };
}

function bears(): CardView {
  return {
    instance_id: "c2",
    name: "Grizzly Bears",
    owner: "p0",
    controller: "p0",
    scryfall_id: "gb",
    type_line: "Creature — Bear",
    mana_cost: "{1}{G}",
  };
}

describe("needsFacePicker", () => {
  it("is true for a modal DFC", () => {
    expect(needsFacePicker(seaGate())).toBe(true);
  });

  it("is false for a single-faced card", () => {
    expect(needsFacePicker(bears())).toBe(false);
  });

  it("is false for a transform card — CR 712.4 casts the front face", () => {
    const jace = { ...seaGate(), layout: "transform" };
    expect(needsFacePicker(jace)).toBe(false);
  });

  it("is false for adventure and split, whose second halves are deferred", () => {
    expect(needsFacePicker({ ...seaGate(), layout: "adventure" })).toBe(false);
    expect(needsFacePicker({ ...seaGate(), layout: "split" })).toBe(false);
  });

  it("is false for a modal_dfc that somehow arrived without faces", () => {
    const broken: CardView = { ...seaGate(), faces: undefined };
    expect(needsFacePicker(broken)).toBe(false);
  });
});

describe("cardAsFace", () => {
  it("swaps the printed fields for the chosen face's", () => {
    const back = cardAsFace(seaGate(), 1);
    expect(back.name).toBe("Sea Gate, Reborn");
    expect(back.type_line).toBe("Land");
    expect(back.mana_cost).toBeUndefined();
    expect(back.active_face).toBe(1);
  });

  it("drops the announce-prompt fields, which describe face 0's spec", () => {
    const withPrompts: CardView = {
      ...seaGate(),
      target_mode: "creature",
      alternative_costs: [{ key: "overload", label: "Overload", mana_cost: "{6}{U}" }],
      additional_cost: { discard_cards: 1 },
      tap_cost: { key: "convoke", max: 2, options: { cards: [] } },
      legal_targets: { cards: ["x"], players: [] },
      modes: { prompt: "Choose one", min: 1, max: 1, options: [{ label: "a" }] },
      // #660: a hand ability is face-0's spec too — a back face that
      // still offered "Cycling {3}" would offer the front's ability.
      hand_abilities: [{ index: 0, label: "Cycling {3}", discard_self: true }],
    };
    const back = cardAsFace(withPrompts, 1);
    expect(back.target_mode).toBeUndefined();
    expect(back.alternative_costs).toBeUndefined();
    expect(back.additional_cost).toBeUndefined();
    expect(back.tap_cost).toBeUndefined();
    expect(back.hand_abilities).toBeUndefined();
    expect(back.legal_targets).toBeUndefined();
    expect(back.modes).toBeUndefined();
  });

  it("returns the card untouched for a face it does not have", () => {
    const c = bears();
    expect(cardAsFace(c, 1)).toBe(c);
    expect(cardAsFace(seaGate(), 7).name).toBe("Sea Gate Restoration");
  });
});

describe("cardImageURL", () => {
  it("omits ?face= for the front face, so the existing cache still hits", () => {
    expect(cardImageURL(bears(), "small")).toBe("/cards/gb/image?size=small");
    expect(scryfallImageURL("gb", "art_crop")).toBe("/cards/gb/image?size=art_crop");
  });

  it("defaults to the card's ACTIVE face", () => {
    expect(cardImageURL(seaGate(1), "normal")).toBe("/cards/sg/image?size=normal&face=1");
    expect(cardImageURL(seaGate(0), "normal")).toBe("/cards/sg/image?size=normal");
  });

  it("takes an explicit face for showing a side that is not up", () => {
    expect(cardImageURL(seaGate(0), "normal", 1)).toBe("/cards/sg/image?size=normal&face=1");
    expect(cardImageURL(seaGate(1), "normal", 0)).toBe("/cards/sg/image?size=normal");
  });

  it("is null without a Scryfall ID — the redacted-card case", () => {
    expect(cardImageURL({ ...bears(), scryfall_id: undefined })).toBeNull();
    expect(cardImageURL(null)).toBeNull();
    expect(cardImageURL(undefined)).toBeNull();
  });
});

describe("applyCastChoices carries the face", () => {
  it("sends face only when it is not the front", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { face: 1 });
    expect(params.face).toBe(1);
  });

  it("omits face 0 — the server default, and omitempty on the wire", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { face: 0 });
    expect(params.face).toBeUndefined();
    applyCastChoices(params, {});
    expect(params.face).toBeUndefined();
  });
});
