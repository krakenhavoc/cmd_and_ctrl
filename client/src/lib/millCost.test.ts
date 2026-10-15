import { describe, expect, it } from "vitest";

import { manaAbilityRider } from "./manaSource";
import type { ManaAbilityView } from "./protocol";

// #3845: "Mill N cards" as a cost component on a mana ability.
describe("mill cost", () => {
  const millikin = {
    index: 0,
    ref: "own:0",
    label: "{T}, Mill a card: Add {C}",
    produced: "{C}",
    tap_cost: true,
    mill_cost: 1,
  } as ManaAbilityView;

  it("names the mill in a mana row's rider", () => {
    expect(manaAbilityRider(millikin)).toContain("mill a card");
    expect(manaAbilityRider({ ...millikin, mill_cost: 4 })).toContain("mill 4 cards");
    expect(manaAbilityRider({ ...millikin, mill_cost: undefined })).not.toContain("mill");
  });
});
