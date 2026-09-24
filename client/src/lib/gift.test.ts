import { get } from "svelte/store";
import { afterEach, describe, expect, it } from "vitest";

import type { CardView, OptionalCostView } from "./protocol";
import {
  applyCastChoices,
  begin,
  cancel,
  castTargetOverride,
  optionalCostOpponentOptions,
  targeting,
} from "./targeting";

// gift.test.ts — CR 702.174 (#1267), the client half.
//
// A gift is one more optional additional cost (ADR 0073) with two
// facets the kicker toggles never needed: naming an opponent, and a
// target clause that changes when the gift is promised. What is worth
// pinning is the wire (gift_opponent rides only with a promise) and the
// clause precedence (alternative cost, then a claimed gift, then the
// card's own).

function card(over: Partial<CardView> = {}): CardView {
  return { instance_id: "spell", name: "Spell", owner: "p0", controller: "p0", ...over };
}

const giftOffer: OptionalCostView = {
  index: 0,
  key: "gift",
  label: "Gift a card",
  chooses_opponent: true,
  opponent_options: ["p1", "p2"],
};

// Long River's Pull: "Counter target creature spell. If the gift was
// promised, instead counter target spell."
const pull = card({
  instance_id: "pull",
  name: "Long River's Pull",
  target_mode: "stack_spell",
  legal_targets: { cards: ["creature-spell"], min: 1, max: 1 },
  optional_costs: [
    {
      ...giftOffer,
      target_mode: "stack_spell",
      legal_targets: { cards: ["creature-spell", "sorcery-spell"], min: 1, max: 1 },
    },
  ],
});

describe("gift — cast_spell params", () => {
  it("writes gift_opponent only when one is set", () => {
    const bare: Record<string, unknown> = {};
    applyCastChoices(bare, { optionalCosts: [0] });
    expect(bare).not.toHaveProperty("gift_opponent");

    const empty: Record<string, unknown> = {};
    applyCastChoices(empty, { giftOpponent: "" });
    expect(empty).not.toHaveProperty("gift_opponent");

    const promised: Record<string, unknown> = {};
    applyCastChoices(promised, { optionalCosts: [0], giftOpponent: "p2" });
    expect(promised).toEqual({ optional_costs: [0], gift_opponent: "p2" });
  });
});

describe("gift — opponent options", () => {
  it("is undefined for an offer that is not a gift", () => {
    expect(optionalCostOpponentOptions({ index: 0, key: "kicker" })).toBeUndefined();
  });

  it("is the list for a gift offer, and empty when nobody can receive it", () => {
    expect(optionalCostOpponentOptions(giftOffer)).toEqual(["p1", "p2"]);
    expect(optionalCostOpponentOptions({ ...giftOffer, opponent_options: [] })).toEqual([]);
    // Missing is treated like empty: a gift with no named recipients
    // cannot be taken.
    expect(optionalCostOpponentOptions({ index: 0, key: "gift", chooses_opponent: true })).toEqual(
      [],
    );
  });
});

describe("gift — target clause override", () => {
  afterEach(() => cancel());

  it("falls back to the card's clause when the gift is not claimed", () => {
    expect(castTargetOverride(pull, undefined)).toBeUndefined();
    expect(castTargetOverride(pull, {})).toBeUndefined();
  });

  it("is the claimed gift's clause when it carries one", () => {
    const o = castTargetOverride(pull, { optionalCosts: [0], giftOpponent: "p1" });
    expect(o?.legal_targets?.cards).toEqual(["creature-spell", "sorcery-spell"]);
  });

  it("ignores a claimed offer that leaves the clause alone", () => {
    const kicked = card({
      target_mode: "creature",
      legal_targets: { cards: ["bear"], min: 1, max: 1 },
      optional_costs: [{ index: 0, key: "kicker", mana_cost: "{2}" }],
    });
    expect(castTargetOverride(kicked, { optionalCosts: [0] })).toBeUndefined();
  });

  it("adds a clause to a card that prints none", () => {
    const untargeted = card({
      optional_costs: [
        {
          ...giftOffer,
          target_mode: "creature",
          legal_targets: { cards: ["bear"], min: 1, max: 1 },
        },
      ],
    });
    expect(castTargetOverride(untargeted, {})).toBeUndefined();
    expect(castTargetOverride(untargeted, { optionalCosts: [0] })?.target_mode).toBe("creature");
  });

  it("lets an alternative cost win over a claimed gift", () => {
    const both = card({
      ...pull,
      alternative_costs: [{ key: "overload", label: "Overload {5}" }],
    });
    // Overload's offer has no clause at all — that means "no targets",
    // not "fall back to the gift's".
    const o = castTargetOverride(both, { altCost: "overload", optionalCosts: [0] });
    expect(o).toBeDefined();
    expect(o?.target_mode).toBeUndefined();
  });

  it("opens the walk on the gift's legal set", () => {
    const choices = { optionalCosts: [0], giftOpponent: "p1" };
    begin(pull, "stack_spell", choices, castTargetOverride(pull, choices));
    const t = get(targeting);
    expect([...(t?.legal?.cards ?? [])]).toEqual(["creature-spell", "sorcery-spell"]);
    expect(t?.choices?.giftOpponent).toBe("p1");
  });
});
