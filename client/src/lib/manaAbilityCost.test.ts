import { describe, expect, it } from "vitest";

import { manaAbilityNeedsPrompt, manaTapNeedsPicker, manaTapPayment } from "./manaAbilityCost";
import type { ManaAbilityView } from "./protocol";

function ability(extra: Partial<ManaAbilityView>): ManaAbilityView {
  return { index: 0, label: "Add {C}", produced: "{C}", ...extra };
}

describe("manaAbilityNeedsPrompt", () => {
  it("sends a plain tap ability straight out", () => {
    expect(manaAbilityNeedsPrompt(ability({}))).toBe(false);
  });

  // #1283: Cadaverous Bloom asks which card leaves the hand.
  it("asks before an exile-a-card cost", () => {
    expect(
      manaAbilityNeedsPrompt(ability({ exile_cost_n: 1, exile_cost_options: ["a", "b"] })),
    ).toBe(true);
  });

  // #1213: the panel click path used to skip this and send no
  // discard_ids, which the server refuses.
  it("asks before a discard cost", () => {
    expect(
      manaAbilityNeedsPrompt(ability({ discard_cost_n: 1, discard_cost_options: ["a"] })),
    ).toBe(true);
  });

  // #758: Springleaf Drum asks which untapped creature pays the
  // second half of its activation cost.
  it("asks before a tap-another cost", () => {
    expect(
      manaAbilityNeedsPrompt(ability({ tap_others_options: { cards: ["bear"], min: 1, max: 1 } })),
    ).toBe(true);
  });
});

describe("manaTapPayment", () => {
  it("sends the chosen permanents as tap_ids", () => {
    expect(manaTapPayment(["bear", "elf"])).toEqual({ tap_ids: ["bear", "elf"] });
  });

  it("keeps ordinary mana-ability payloads unchanged", () => {
    expect(manaTapPayment([])).toEqual({});
  });
});

// #5447: Hazel of the Rootbloom's "Tap X untapped tokens you control".
describe("manaTapNeedsPicker", () => {
  it("sends a fixed count on its own when the options are exactly that many", () => {
    expect(manaTapNeedsPicker({ cards: ["bear"], min: 1, max: 1 })).toBe(false);
    expect(manaTapNeedsPicker({ cards: ["a", "b"], min: 2, max: 2 })).toBe(false);
  });

  it("asks which permanents pay a fixed count with more options than it needs", () => {
    expect(manaTapNeedsPicker({ cards: ["bear", "elf"], min: 1, max: 1 })).toBe(true);
  });

  it("always asks how many to tap when the count is the player's", () => {
    expect(manaTapNeedsPicker({ cards: ["a"], min: 0, max: 1 })).toBe(true);
    expect(manaTapNeedsPicker({ cards: ["a", "b", "c"], min: 0, max: 3 })).toBe(true);
  });

  it("asks nothing with nothing to tap or no such cost", () => {
    expect(manaTapNeedsPicker({ cards: [], min: 0, max: 0 })).toBe(false);
    expect(manaTapNeedsPicker(undefined)).toBe(false);
  });
});
