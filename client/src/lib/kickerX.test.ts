// kickerX.test.ts — #3232. "Kicker {X}. X can't be 0." (Thieving Skydiver)
// puts the X on the KICKER: the cast asks for X only once the kicker is
// chosen, and the picker's floor is the kicker's `x_min`.

import { describe, it, expect } from "vitest";

import { hasXCost, optionalCostXFloor } from "./targeting";
import type { CardView } from "./protocol";

const skydiver = {
  instance_id: "sky-1",
  name: "Thieving Skydiver",
  owner: "me",
  controller: "me",
  type_line: "Creature — Merfolk Rogue",
  mana_cost: "{1}{U}",
  optional_costs: [{ index: 0, key: "kicker", label: "Kicker {X}", mana_cost: "{X}", x_min: 1 }],
} as CardView;

const verdeloth = {
  ...skydiver,
  name: "Verdeloth the Ancient",
  mana_cost: "{4}{G}{G}",
  optional_costs: [{ index: 0, key: "kicker", label: "Kicker {X}", mana_cost: "{X}" }],
} as CardView;

const burst = {
  ...skydiver,
  name: "Burst Lightning",
  mana_cost: "{R}",
  optional_costs: [{ index: 0, key: "kicker", label: "Kicker {4}", mana_cost: "{4}" }],
} as CardView;

describe("hasXCost with a kicker {X}", () => {
  it("asks for X only when the kicker is chosen", () => {
    expect(hasXCost(skydiver, undefined, [0])).toBe(true);
    expect(hasXCost(skydiver, undefined, [])).toBe(false);
    expect(hasXCost(skydiver)).toBe(false);
  });

  it("does not ask for a kicker with a flat cost", () => {
    expect(hasXCost(burst, undefined, [0])).toBe(false);
  });

  it("still asks for a printed {X}", () => {
    const fireball = { ...burst, mana_cost: "{X}{R}" } as CardView;
    expect(hasXCost(fireball, undefined, [])).toBe(true);
  });
});

describe("optionalCostXFloor", () => {
  it("is the chosen kicker's floor", () => {
    expect(optionalCostXFloor(skydiver, [0])).toBe(1);
  });

  it("is 0 when the kicker is not chosen", () => {
    expect(optionalCostXFloor(skydiver, [])).toBe(0);
    expect(optionalCostXFloor(skydiver, undefined)).toBe(0);
  });

  it("is 0 for a kicker with no floor", () => {
    expect(optionalCostXFloor(verdeloth, [0])).toBe(0);
  });

  it("ignores an index the card does not have", () => {
    expect(optionalCostXFloor(skydiver, [7])).toBe(0);
  });
});
