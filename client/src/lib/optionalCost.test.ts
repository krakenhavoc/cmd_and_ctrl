import { describe, expect, it } from "vitest";

import type { CardView } from "./protocol";
import {
  applyCastChoices,
  castIsForbidden,
  castSacrificeClause,
  castSacrificeLabel,
  optionalCostMaxTimes,
  optionalCostPayOptions,
  optionalCostSelection,
  optionalCostsOf,
} from "./targeting";

// optionalCost.test.ts — ADR 0073 (#664, #760), the client half.
//
// The picker collects per-offer COUNTS and the wire wants repeated
// INDICES, so the translation between the two is the thing worth
// pinning: a multikicker paid three times is [0, 0, 0], and it is
// ascending so two clients that picked the same costs in different
// orders send the same announcement.

function card(over: Partial<CardView> = {}): CardView {
  return {
    instance_id: "spell-1",
    name: "Kickable",
    ...over,
  } as CardView;
}

describe("optional additional costs", () => {
  it("reads the offers off the card, and none off a card with none", () => {
    expect(optionalCostsOf(card())).toEqual([]);
    const kickable = card({
      optional_costs: [{ index: 0, key: "kicker", label: "Kicker {4}", mana_cost: "{4}" }],
    });
    expect(optionalCostsOf(kickable)).toHaveLength(1);
  });

  it("defaults max_times to one, and keeps a multikicker's cap", () => {
    expect(optionalCostMaxTimes({ index: 0, key: "kicker" })).toBe(1);
    expect(optionalCostMaxTimes({ index: 0, key: "multikicker", max_times: 7 })).toBe(7);
    // A cap below one is a malformed offer, not a cost that can never
    // be paid — clamp rather than hide the toggle.
    expect(optionalCostMaxTimes({ index: 0, key: "kicker", max_times: 0 })).toBe(1);
  });

  it("turns per-offer counts into the wire's repeated indices, ascending", () => {
    expect(optionalCostSelection(new Map())).toEqual([]);
    expect(optionalCostSelection(new Map([[0, 1]]))).toEqual([0]);
    // Multikicker: the count IS the repetition (CR 702.33d).
    expect(optionalCostSelection(new Map([[0, 3]]))).toEqual([0, 0, 0]);
    // Two offers, claimed out of order, sent in index order.
    expect(
      optionalCostSelection(
        new Map([
          [1, 1],
          [0, 2],
        ]),
      ),
    ).toEqual([0, 0, 1]);
    // A count of zero is a declined offer, not an entry.
    expect(optionalCostSelection(new Map([[0, 0]]))).toEqual([]);
  });

  it("omits the field entirely when nothing is claimed", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { optionalCosts: [] });
    expect(params.optional_costs).toBeUndefined();

    applyCastChoices(params, { optionalCosts: [0, 0] });
    expect(params.optional_costs).toEqual([0, 0]);
  });

  it("reports an unpayable non-mana offer as present-and-empty", () => {
    // A mana kicker charges no cards: undefined, so the picker offers
    // no sacrifice list at all.
    expect(optionalCostPayOptions({ index: 0, key: "kicker", mana_cost: "{4}" })).toBeUndefined();
    // Constant Mists with no land: present, empty, unpayable.
    expect(
      optionalCostPayOptions({ index: 0, key: "buyback", sacrifice_options: { cards: [] } }),
    ).toEqual([]);
  });
});

describe("the sacrifice clause a cast is paying", () => {
  const mandatory = card({
    additional_cost: { label: "Sacrifice a creature", sacrifice_options: { cards: ["bear"] } },
  });
  const optional = card({
    optional_costs: [
      {
        index: 0,
        key: "buyback",
        label: "Buyback—Sacrifice a land",
        sacrifice_options: { cards: ["forest"] },
      },
    ],
  });

  it("is undefined when the cast owes no sacrifice", () => {
    expect(castSacrificeClause(card(), {})).toBeUndefined();
    // Declared but not claimed: the offer exists, the cast is not
    // paying it, and no picker should open.
    expect(castSacrificeClause(optional, {})).toBeUndefined();
  });

  it("is the mandatory clause when the card has one", () => {
    expect(castSacrificeClause(mandatory, {})?.cards).toEqual(["bear"]);
    expect(castSacrificeLabel(mandatory, {})).toBe("Sacrifice a creature");
  });

  it("is the claimed optional clause when the card has no mandatory one", () => {
    expect(castSacrificeClause(optional, { optionalCosts: [0] })?.cards).toEqual(["forest"]);
    expect(castSacrificeLabel(optional, { optionalCosts: [0] })).toBe("Buyback—Sacrifice a land");
  });

  it("ignores a claimed index the card does not offer", () => {
    expect(castSacrificeClause(optional, { optionalCosts: [5] })).toBeUndefined();
  });
});

describe("the cast gate's view stamp", () => {
  it("is false for a card nothing refuses", () => {
    expect(castIsForbidden(card())).toBe(false);
    expect(castIsForbidden(card({ cant_cast: "" }))).toBe(false);
  });

  it("is true, and carries the printed clause, for one the server would refuse", () => {
    const banned = card({
      cant_cast: "Rule of Law — each player can't cast more than one spell each turn.",
    });
    expect(castIsForbidden(banned)).toBe(true);
    expect(banned.cant_cast).toContain("more than one spell");
  });
});
