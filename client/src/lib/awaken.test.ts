import { describe, it, expect } from "vitest";

import { castableAlternativeCostsOf, castTargetOverride, stepsFor } from "./targeting";
import type { AlternativeCostView, CardView } from "./protocol";

// awaken.test.ts — ADR 0135 §3 (#2411). Awaken needs no client code of
// its own: the offer is listed by its label, and claiming it replaces the
// spell's target statement with the one the server stamped on the offer,
// the spell's own clause followed by "target land you control", which the
// targeting walk takes in order. This pins that chain.

const awaken: AlternativeCostView = {
  key: "awaken",
  label: "Awaken 4—{5}{B}{B}",
  mana_cost: "{5}{B}{B}",
  target_mode: "permanent",
  legal_targets: { cards: ["victim"], min: 1, max: 1 },
  clauses: [
    { label: "target creature or planeswalker", cards: ["victim"], min: 1, max: 1 },
    { label: "target land you control", cards: ["swamp"], min: 1, max: 1 },
  ],
  purpose: { awaken_land: 4 },
};

function ruinousPath(): CardView {
  return {
    instance_id: "path",
    name: "Ruinous Path",
    owner: "me",
    controller: "me",
    type_line: "Sorcery",
    castable_here: true,
    target_mode: "permanent",
    legal_targets: { cards: ["victim"], min: 1, max: 1 },
    alternative_costs: [awaken],
  };
}

describe("awaken", () => {
  it("lists the awaken offer by its printed label", () => {
    expect(castableAlternativeCostsOf(ruinousPath()).map((o) => o.label)).toEqual([
      "Awaken 4—{5}{B}{B}",
    ]);
  });

  it("walks the spell's clause, then the land, when awaken is claimed", () => {
    const override = castTargetOverride(ruinousPath(), { altCost: "awaken" });
    expect(override?.clauses?.length).toBe(2);
    const steps = stepsFor("permanent", override?.legal_targets, override?.clauses, 0);
    expect(steps.map((s) => [s.slot, s.label, [...(s.legal?.cards ?? [])]])).toEqual([
      [0, "target creature or planeswalker", ["victim"]],
      [1, "target land you control", ["swamp"]],
    ]);
    // CR 601.2c: the land clause may name an object the first names.
    expect(steps[1].distinct).toBe(false);
  });

  it("keeps the card's own single clause for a cast at the mana cost", () => {
    expect(castTargetOverride(ruinousPath(), {})).toBeUndefined();
  });
});
