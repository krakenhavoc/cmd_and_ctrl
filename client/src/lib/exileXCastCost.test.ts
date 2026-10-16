import { describe, expect, it } from "vitest";
import { castExileCardsCost, exileXChoice, hasXCost, xCeilingFor } from "./targeting";
import type { CardView } from "./protocol";

// #1948: "exile X cards from your graveyard" (Harvest Pyre, Chill
// Haunting) — the X prompt opens for a card whose mana cost has no {X},
// is capped at what the graveyard pays, and the picker that follows asks
// for exactly X cards. Soaring Stoneglider's exile branch reads as a fixed
// count once the branch is chosen.

function pyre(options: string[], max = options.length): CardView {
  return {
    instance_id: "spell",
    name: "Harvest Pyre",
    owner: "me",
    controller: "me",
    mana_cost: "{1}{R}",
    additional_cost: {
      label: "Exile X cards from your graveyard",
      demands_x: true,
      exile_cost_x: true,
      exile_cost_x_max: max,
      exile_cost_label: "X cards",
      exile_cost_options: options,
      exile_cost_zone: "graveyard",
    },
  };
}

describe("the X form", () => {
  it("asks for X although the mana cost has none, capped at the graveyard", () => {
    const card = pyre(["g1", "g2", "g3"]);
    expect(hasXCost(card)).toBe(true);
    expect(xCeilingFor(card, undefined)).toBe(3);
    expect(xCeilingFor({ ...card, x_max: 2 }, undefined)).toBe(2);
  });

  it("sizes the picker to the announced X and names the number", () => {
    const got = castExileCardsCost(pyre(["g1", "g2", "g3"]), { xValue: 2 });
    expect(got?.fromX).toBe(true);
    expect(got?.count).toBe(2);
    expect(got?.label).toBe("2 cards");
    expect(got?.options).toEqual(["g1", "g2", "g3"]);
  });

  it("needs no picker at X = 0 or when the graveyard holds exactly X", () => {
    const zero = castExileCardsCost(pyre(["g1", "g2"]), { xValue: 0 })!;
    expect(exileXChoice(zero, { xValue: 0 })?.exileIDs).toEqual([]);
    const exact = castExileCardsCost(pyre(["g1", "g2"]), { xValue: 2 })!;
    expect(exileXChoice(exact, { xValue: 2 })?.exileIDs).toEqual(["g1", "g2"]);
    const choose = castExileCardsCost(pyre(["g1", "g2", "g3"]), { xValue: 2 })!;
    expect(exileXChoice(choose, { xValue: 2 })).toBeUndefined();
  });
});

describe("Soaring Stoneglider's branch", () => {
  const card: CardView = {
    instance_id: "spell",
    name: "Soaring Stoneglider",
    owner: "me",
    controller: "me",
    mana_cost: "{2}{W}",
    additional_cost: {
      label: "Exile two cards from your graveyard or Pay {1}{W}",
      branches: [
        {
          key: "exile",
          label: "Exile two cards from your graveyard",
          exile_cost_n: 2,
          exile_cost_label: "two cards",
          exile_cost_options: ["g1", "g2", "g3"],
          exile_cost_zone: "graveyard",
          payable: true,
        },
        { key: "mana", label: "Pay {1}{W}", mana_cost: "{1}{W}", payable: true },
      ],
    },
  };

  it("asks for two cards on the exile branch and none on the mana branch", () => {
    const exile = castExileCardsCost(card, { costBranch: 0 });
    expect(exile?.fromX).toBe(false);
    expect(exile?.count).toBe(2);
    expect(castExileCardsCost(card, { costBranch: 1 })).toBeUndefined();
    expect(hasXCost(card)).toBe(false);
  });
});
