import { describe, expect, it } from "vitest";

import { abilityBlocked, exertCostWords } from "./contextMenu.logic";
import { manaAbilityRider } from "./manaSource";
import type { ManaAbilityView } from "./protocol";

// ADR 0130 §4: an exert cost component on a mana or activated ability.
describe("exert cost", () => {
  const arena: ManaAbilityView = {
    index: 1,
    ref: "own:1",
    label: "{R}, {T}, Exert this land: Add {R}{R}.",
    produced: "{R}{R}",
    tap_cost: true,
    mana_cost: "{R}",
    exert: true,
  } as ManaAbilityView;

  it("names the exert in a mana row's rider", () => {
    expect(manaAbilityRider(arena)).toContain("exert it");
    expect(manaAbilityRider({ ...arena, exert: undefined })).not.toContain("exert");
  });

  it("says what an exert costs, and nothing for a row without one", () => {
    expect(exertCostWords({ exert: true })).toBe(
      "exert it: it won't untap during your next untap step",
    );
    expect(exertCostWords({})).toBe("");
  });

  it("never blocks a row: an exert can always be paid", () => {
    expect(abilityBlocked({ exert: true, tap_cost: true }, false, false)).toBeFalsy();
  });
});
