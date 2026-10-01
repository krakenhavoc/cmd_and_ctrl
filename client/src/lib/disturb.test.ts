import { describe, it, expect } from "vitest";

import { castableFromZone } from "./zoneBrowser.logic";
import {
  applyCastChoices,
  castableAlternativeCostsOf,
  castTargetOverride,
  printedCostClaimable,
} from "./targeting";
import type { AlternativeCostView, CardView } from "./protocol";

// disturb.test.ts — ADR 0107 §4 (#1855). Disturb needs no client code
// of its own: the graveyard menu lists alternative costs by label, and
// the server stamps the disturb offer with the BACK face's target
// clause, because the spell is the back face (CR 712.8c). This pins
// that the existing chain reads the frame the server sends the way the
// cast needs: the button is there, disturb is the only price, a
// disturbed Aura targets with its own clause, and the payload names
// the graveyard and the cost with no face (the server turns the card
// over itself).

const disturb: AlternativeCostView = {
  key: "disturb",
  label: "Disturb {3}{W}",
  mana_cost: "{3}{W}",
  target_mode: "creature",
  legal_targets: { cards: ["bear"] },
};

function drogskolInfantry(): CardView {
  return {
    instance_id: "infantry",
    name: "Drogskol Infantry",
    owner: "me",
    controller: "me",
    type_line: "Creature — Spirit Soldier",
    layout: "transform",
    castable_here: true,
    alternative_cost_required: true,
    alternative_costs: [disturb],
    faces: [
      {
        name: "Drogskol Infantry",
        type_line: "Creature — Spirit Soldier",
        castable_here: true,
        alternative_cost_required: true,
        alternative_costs: [disturb],
      },
      { name: "Drogskol Armaments", type_line: "Enchantment — Aura" },
    ],
  };
}

describe("disturb from the graveyard", () => {
  it("offers the graveyard cast, at the disturb cost alone", () => {
    const card = drogskolInfantry();
    expect(castableFromZone(card, "graveyard")).toBe(true);
    expect(printedCostClaimable(card)).toBe(false);
    expect(castableAlternativeCostsOf(card).map((o) => o.label)).toEqual(["Disturb {3}{W}"]);
  });

  it("targets with the back face's clause the server stamped on the offer", () => {
    const override = castTargetOverride(drogskolInfantry(), { altCost: "disturb" });
    expect(override?.target_mode).toBe("creature");
    expect(override?.legal_targets?.cards).toEqual(["bear"]);
  });

  it("sends the graveyard and the cost, and leaves the face to the server", () => {
    const params: Record<string, unknown> = { instance_id: "infantry" };
    applyCastChoices(params, { fromZone: "graveyard", altCost: "disturb" });
    expect(params).toEqual({
      instance_id: "infantry",
      from_zone: "graveyard",
      alternative_cost: "disturb",
    });
  });
});
