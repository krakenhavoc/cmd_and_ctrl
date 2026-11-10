import { describe, expect, it } from "vitest";
import type { CardView } from "./protocol";
import { applyCastChoices, castRevealCost, optionalCostPayOptions } from "./targeting";

// revealCost.test.ts — ADR 0100 amendment 2026-10-07: a reveal or
// behold branch on the client. The chosen branch drives the one-card
// picker, and the pick rides cast_spell as reveal_ids.

function vanquisher(): CardView {
  return {
    instance_id: "wr",
    name: "Wren's Run Vanquisher",
    owner: "p0",
    controller: "p0",
    additional_cost: {
      label: "Reveal an Elf card from your hand or pay {3}",
      branches: [
        {
          key: "reveal",
          label: "Reveal an Elf card from your hand",
          reveal: true,
          reveal_options: { cards: ["elf"], min: 1, max: 1 },
          payable: true,
        },
        { key: "mana", label: "Pay {3}", mana_cost: "{3}", payable: true },
      ],
    },
  } as CardView;
}

describe("reveal and behold branches", () => {
  it("asks for a card only when the reveal branch is chosen", () => {
    expect(castRevealCost(vanquisher(), { costBranch: 0 })).toEqual({
      label: "Reveal an Elf card from your hand",
      behold: false,
      exile: false,
      options: ["elf"],
      count: 1,
    });
    expect(castRevealCost(vanquisher(), { costBranch: 1 })).toBeUndefined();
    expect(castRevealCost(vanquisher(), undefined)).toBeUndefined();
  });

  it("marks a behold branch", () => {
    const card = vanquisher();
    card.additional_cost!.branches![0].behold = true;
    expect(castRevealCost(card, { costBranch: 0 })?.behold).toBe(true);
  });

  it("sends reveal_ids on cast_spell only when a card was named", () => {
    const sent: Record<string, unknown> = {};
    applyCastChoices(sent, { costBranch: 0, revealIDs: ["elf"] });
    expect(sent.cost_branch).toBe(0);
    expect(sent.reveal_ids).toEqual(["elf"]);
    const none: Record<string, unknown> = {};
    applyCastChoices(none, { costBranch: 1 });
    expect(none.reveal_ids).toBeUndefined();
  });
});

// #4642: "you may behold a Hero" is an optional additional cost, not a
// branch. Ticking it asks the same one-card question over its own options,
// and an offer with nothing to behold cannot be ticked.
function hyperion(options: string[]): CardView {
  return {
    instance_id: "hv",
    name: "Hyperion's Atomic Vision",
    owner: "p0",
    controller: "p0",
    optional_costs: [
      {
        index: 0,
        key: "behold",
        label: "Behold a Hero",
        max_times: 1,
        reveal: true,
        behold: true,
        reveal_options: { cards: options, min: 1, max: 1 },
      },
    ],
  } as CardView;
}

describe("optional behold offers", () => {
  it("asks for a card only when the offer is ticked", () => {
    expect(castRevealCost(hyperion(["hero"]), undefined)).toBeUndefined();
    expect(castRevealCost(hyperion(["hero"]), { optionalCosts: [] })).toBeUndefined();
    expect(castRevealCost(hyperion(["hero", "scout"]), { optionalCosts: [0] })).toEqual({
      label: "Behold a Hero",
      behold: true,
      exile: false,
      options: ["hero", "scout"],
      count: 1,
    });
  });

  it("cannot be taken with nothing to behold", () => {
    expect(optionalCostPayOptions(hyperion([]).optional_costs![0])).toEqual([]);
    expect(optionalCostPayOptions(hyperion(["hero"]).optional_costs![0])).toEqual(["hero"]);
  });

  it("sends the optional cost and the pick together", () => {
    const sent: Record<string, unknown> = {};
    applyCastChoices(sent, { optionalCosts: [0], revealIDs: ["hero"] });
    expect(sent.optional_costs).toEqual([0]);
    expect(sent.reveal_ids).toEqual(["hero"]);
  });
});

// #4687: the Champions' mandatory "behold a Kithkin and exile it" is the
// card's own additional cost, with no branches: every cast asks for the
// one card, and the prompt says it is exiled.
function champion(options: string[]): CardView {
  return {
    instance_id: "cc",
    name: "Champion of the Clachan",
    owner: "p0",
    controller: "p0",
    additional_cost: {
      label: "Behold a Kithkin and exile it",
      reveal: true,
      behold: true,
      behold_exile: true,
      reveal_options: { cards: options, min: 1, max: 1 },
    },
  } as CardView;
}

describe("behold-and-exile costs", () => {
  it("asks for a card on every cast, marked as exiled", () => {
    expect(castRevealCost(champion(["kithkin", "token"]), undefined)).toEqual({
      label: "Behold a Kithkin and exile it",
      behold: true,
      exile: true,
      options: ["kithkin", "token"],
      count: 1,
    });
  });

  it("offers nothing when there is nothing to behold", () => {
    expect(castRevealCost(champion([]), undefined)?.options).toEqual([]);
  });
});

// #5406: Celestial Reunion's "you may choose a creature type and behold
// two creatures of that type". Ticking the offer lists the types; once a
// type is chosen the pick is that type's cards, two of them, and the type
// rides cast_spell as behold_type.
function reunion(): CardView {
  return {
    instance_id: "cr",
    name: "Celestial Reunion",
    owner: "p0",
    controller: "p0",
    optional_costs: [
      {
        index: 0,
        key: "behold",
        label: "Choose a creature type and behold two creatures of that type",
        max_times: 1,
        reveal: true,
        behold: true,
        behold_count: 2,
        behold_types: [
          { type: "Elf", cards: ["elf1", "elf2", "elf3"] },
          { type: "Goblin", cards: ["gob1", "gob2"] },
        ],
        reveal_options: { cards: ["elf1", "elf2", "elf3", "gob1", "gob2"], min: 2, max: 2 },
      },
    ],
  } as CardView;
}

describe("behold of a chosen creature type", () => {
  it("lists the types, then that type's cards, two to pick", () => {
    const before = castRevealCost(reunion(), { optionalCosts: [0] });
    expect(before?.count).toBe(2);
    expect(before?.types?.map((t) => t.type)).toEqual(["Elf", "Goblin"]);
    const after = castRevealCost(reunion(), { optionalCosts: [0], beholdType: "Goblin" });
    expect(after?.options).toEqual(["gob1", "gob2"]);
    expect(after?.count).toBe(2);
  });

  it("lists no types for a behold that prints its type", () => {
    expect(castRevealCost(hyperion(["hero"]), { optionalCosts: [0] })?.types).toBeUndefined();
  });

  it("sends the type with the picks", () => {
    const sent: Record<string, unknown> = {};
    applyCastChoices(sent, { optionalCosts: [0], beholdType: "Elf", revealIDs: ["elf1", "elf2"] });
    expect(sent.behold_type).toBe("Elf");
    expect(sent.reveal_ids).toEqual(["elf1", "elf2"]);
    const none: Record<string, unknown> = {};
    applyCastChoices(none, { optionalCosts: [0], revealIDs: ["hero"] });
    expect(none.behold_type).toBeUndefined();
  });
});
