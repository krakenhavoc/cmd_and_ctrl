import { describe, expect, it } from "vitest";

import { manaAbilityNeedsPrompt } from "./manaAbilityCost";
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
});
