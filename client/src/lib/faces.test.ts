import { describe, it, expect } from "vitest";

import { cardAsFace, castableFaceIndex, castableFaces, needsFacePicker } from "./faces";
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

// bonecrusher is #992's shape: an adventure card whose ADVENTURE half
// carries an announce block of its own and whose creature half does
// not. Named for the card the issue ships as its proof.
function bonecrusher(): CardView {
  return {
    instance_id: "c3",
    name: "Bonecrusher Giant",
    owner: "p0",
    controller: "p0",
    scryfall_id: "bg",
    type_line: "Creature — Giant",
    mana_cost: "{2}{R}",
    layout: "adventure",
    faces: [
      { name: "Bonecrusher Giant", type_line: "Creature — Giant", mana_cost: "{2}{R}" },
      {
        name: "Stomp",
        type_line: "Instant — Adventure",
        mana_cost: "{1}{R}",
        target_mode: "any",
        legal_targets: { cards: ["bear"], players: [], min: 1, max: 1 },
        clauses: [{ cards: ["bear"], players: [], min: 1, max: 1 }],
        alternative_costs: [{ key: "stomp-alt", label: "Stomp alt", mana_cost: "{R/P}{R/P}" }],
        phyrexian_symbols: 2,
      },
    ],
  };
}

// blankFaces is bonecrusher's face list with every announce block
// stripped — the ordinary case, and what a face a grant does not open
// looks like on the wire.
function blankFaces() {
  return [
    { name: "Bonecrusher Giant", type_line: "Creature — Giant", mana_cost: "{2}{R}" },
    { name: "Stomp", type_line: "Instant — Adventure", mana_cost: "{1}{R}" },
  ];
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

  it("is true for an adventure card — CR 715.3 offers the creature or the Adventure", () => {
    expect(needsFacePicker({ ...seaGate(), layout: "adventure" })).toBe(true);
  });

  it("is false for split, whose fusing is deferred", () => {
    expect(needsFacePicker({ ...seaGate(), layout: "split" })).toBe(false);
  });

  it("is false for an adventure card that arrived without faces", () => {
    expect(needsFacePicker({ ...seaGate(), layout: "adventure", faces: undefined })).toBe(false);
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

  // #992. This used to read "drops the announce-prompt fields",
  // because the server published one block — face 0's — and the
  // front's answers attached to the back are actively wrong. It
  // publishes one per castable face now, so the swap is a SWAP: the
  // chosen half's block replaces the card's, and the clear below is
  // what happens when the chosen half has no block to put there.
  it("swaps in the chosen face's announce block", () => {
    const stomp: CardView = {
      ...bonecrusher(),
      // The card's own block is face 0's: a creature that targets
      // nothing.
      target_mode: undefined,
      legal_targets: undefined,
    };
    const back = cardAsFace(stomp, 1);
    expect(back.target_mode).toBe("any");
    expect(back.legal_targets?.cards).toEqual(["bear"]);
    // The clause list, the price list and the X notes travel with it.
    expect(back.clauses?.length).toBe(1);
    expect(back.alternative_costs?.[0]?.key).toBe("stomp-alt");
    expect(back.phyrexian_symbols).toBe(2);
    // …and the printed half moves too, so the cost prompts price the
    // Adventure rather than the creature.
    expect(back.name).toBe("Stomp");
    expect(back.mana_cost).toBe("{1}{R}");
  });

  it("clears the front's block when the chosen face has none of its own", () => {
    const withFront: CardView = {
      ...bonecrusher(),
      target_mode: "creature",
      legal_targets: { cards: ["front-target"], players: [] },
      phyrexian_symbols: 1,
    };
    // Face 0 of this fixture announces nothing (the creature half),
    // so picking it must not leave the CARD's block — which here is a
    // stand-in for a front face that did target — in place.
    const front = cardAsFace({ ...withFront, faces: blankFaces() }, 0);
    expect(front.target_mode).toBeUndefined();
    expect(front.legal_targets).toBeUndefined();
    expect(front.phyrexian_symbols).toBeUndefined();
  });

  it("drops the announce-prompt fields for a face that announces nothing", () => {
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
      zone_abilities: [{ index: 0, label: "Cycling {3}", discard_self: true }],
      // #1055: the cast-surface bit is the viewer's own answer for the
      // face the grant NAMES, so it belongs to face 0's spec as much
      // as the offer list beside it does.
      castable_here: true,
    };
    const back = cardAsFace(withPrompts, 1);
    expect(back.target_mode).toBeUndefined();
    expect(back.alternative_costs).toBeUndefined();
    expect(back.additional_cost).toBeUndefined();
    expect(back.tap_cost).toBeUndefined();
    expect(back.zone_abilities).toBeUndefined();
    expect(back.legal_targets).toBeUndefined();
    expect(back.modes).toBeUndefined();
    // Kept, it would be a cast button over an empty price list —
    // #1015's "nothing behind it", one face over.
    expect(back.castable_here).toBeUndefined();
  });

  it("returns the card untouched for a face it does not have", () => {
    const c = bears();
    expect(cardAsFace(c, 1)).toBe(c);
    expect(cardAsFace(seaGate(), 7).name).toBe("Sea Gate Restoration");
  });
});

// castableFaces / castableFaceIndex — #1173, #1168. The ONE walk both
// bugs share: reading a multi-face card's own top-level block answers
// for face 0 alone, and #1171 / #992 mean that is no longer the whole
// story for a card whose OTHER face is the one that carries the
// permission or the announce data in question.
describe("castableFaces", () => {
  it("is just the card for a single-faced card", () => {
    const c = bears();
    expect(castableFaces(c)).toEqual([c]);
  });

  it("is just the card for a transform card, whose back isn't independently castable", () => {
    const jace = { ...seaGate(), layout: "transform" };
    expect(castableFaces(jace)).toEqual([jace]);
  });

  it("is every face, materialised, for a modal DFC or an adventure card", () => {
    const faces = castableFaces(seaGate());
    expect(faces.map((f) => f.name)).toEqual(["Sea Gate Restoration", "Sea Gate, Reborn"]);
    expect(faces[0].active_face).toBe(0);
    expect(faces[1].active_face).toBe(1);

    const stomp = castableFaces({ ...bonecrusher(), layout: "adventure" });
    expect(stomp.map((f) => f.name)).toEqual(["Bonecrusher Giant", "Stomp"]);
    expect(stomp[1].target_mode).toBe("any");
  });
});

describe("castableFaceIndex", () => {
  it("picks the first face the test accepts, in printed order", () => {
    const faces = seaGate();
    expect(castableFaceIndex(faces, (f) => f.name === "Sea Gate, Reborn")).toBe(1);
  });

  it("falls back to 0 when no face satisfies the test", () => {
    expect(castableFaceIndex(seaGate(), () => false)).toBe(0);
  });

  it("is 0 for a single-faced card regardless of the test", () => {
    // castableFaces collapses to [card], so the only thing the test
    // can be asked about IS the card, and a miss falls back to 0
    // exactly as it does for a multi-face card with no match.
    expect(castableFaceIndex(bears(), () => false)).toBe(0);
    expect(castableFaceIndex(bears(), () => true)).toBe(0);
  });

  it("finds a back face's castable_here (#1173's shape)", () => {
    const card: CardView = {
      ...seaGate(),
      castable_here: false,
      faces: [
        { name: "Sea Gate Restoration", type_line: "Sorcery", castable_here: false },
        { name: "Sea Gate, Reborn", type_line: "Land", castable_here: true },
      ],
    };
    expect(castableFaceIndex(card, (f) => f.castable_here === true)).toBe(1);
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
