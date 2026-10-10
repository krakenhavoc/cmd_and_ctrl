import { describe, expect, it } from "vitest";

import {
  abilityDemandsX,
  abilityEnergyMaxX,
  abilityLoyaltyMaxX,
  abilityMaxX,
  abilityMinX,
  abilityXSlots,
  suggestedAbilityX,
} from "./abilityX";
import { energyCostSymbols, energyCostWords } from "./contextMenu.logic";
import { beginForAbility, targeting, type TargetingState } from "./targeting";
import type { ActivatedAbilityView, CardView } from "./protocol";

// abilityX.test.ts — {X} on an activated ability (CR 602.2b), the
// client half.

function ability(over: Partial<ActivatedAbilityView> = {}): ActivatedAbilityView {
  return { index: 0, label: "an ability", ...over };
}

function card(): CardView {
  return {
    instance_id: "src",
    name: "Helm of Obedience",
    type_line: "Artifact",
    owner: "me",
    controller: "me",
  } as CardView;
}

describe("abilityDemandsX", () => {
  it("is false for an ability with no X", () => {
    expect(abilityDemandsX(ability({ mana_cost: "{2}" }))).toBe(false);
  });

  it("reads the server's flag rather than the cost string", () => {
    // The braces are the server's to parse. A client that hunted for
    // "{X}" itself would be a second parser of the same syntax.
    expect(abilityDemandsX(ability({ mana_cost: "{X}", demands_x: true }))).toBe(true);
    expect(abilityDemandsX(ability({ mana_cost: "{X}" }))).toBe(false);
  });
});

describe("abilityMinX", () => {
  it("defaults to zero", () => {
    expect(abilityMinX(ability({ demands_x: true }))).toBe(0);
  });

  it('carries "X can\'t be 0"', () => {
    expect(abilityMinX(ability({ demands_x: true, min_x: 1 }))).toBe(1);
  });

  it("clamps nonsense off the wire to zero", () => {
    expect(abilityMinX(ability({ demands_x: true, min_x: -4 }))).toBe(0);
  });
});

describe("abilityXSlots", () => {
  it("defaults to one", () => {
    expect(abilityXSlots(ability({ demands_x: true }))).toBe(1);
  });

  it('reads two for Treasure Vault\'s "{X}{X}"', () => {
    expect(abilityXSlots(ability({ demands_x: true, x_slots: 2 }))).toBe(2);
  });

  it("treats a bogus slot count as one — over-pricing, never under", () => {
    expect(abilityXSlots(ability({ demands_x: true, x_slots: 0 }))).toBe(1);
  });
});

describe("suggestedAbilityX", () => {
  it("spends everything on a one-slot cost", () => {
    expect(suggestedAbilityX(ability({ demands_x: true }), 6)).toBe(6);
  });

  it("halves the guess for a two-slot cost", () => {
    // Treasure Vault: six mana over "{X}{X}" is X=3, not X=6.
    expect(suggestedAbilityX(ability({ demands_x: true, x_slots: 2 }), 6)).toBe(3);
    expect(suggestedAbilityX(ability({ demands_x: true, x_slots: 2 }), 7)).toBe(3);
  });

  it("never suggests below the printed floor", () => {
    // Helm of Obedience with no mana: the suggestion is still 1,
    // because 0 is not a legal announcement. The server refuses it
    // and the enumerator never offers it either.
    expect(suggestedAbilityX(ability({ demands_x: true, min_x: 1 }), 0)).toBe(1);
  });

  it("treats negative mana as none", () => {
    expect(suggestedAbilityX(ability({ demands_x: true }), -3)).toBe(0);
  });
});

// ADR 0129 §8: "Pay X {E}" — X is capped at the seat's energy (CR
// 118.3), and with no {X} in the mana the energy alone is the guess.
describe("energy X", () => {
  const sphinx = ability({ mana_cost: "{W}{U}{U}", demands_x: true, energy_cost_x: true });

  it("caps X at the seat's energy", () => {
    expect(abilityEnergyMaxX(sphinx, 4)).toBe(4);
    expect(abilityEnergyMaxX(ability({ energy_cost_x: true, energy_cost: 1 }), 4)).toBe(3);
    expect(abilityEnergyMaxX(ability({ energy_cost: 3 }), 4)).toBeUndefined();
  });

  it("suggests the energy for an X paid only in energy", () => {
    expect(suggestedAbilityX(sphinx, 9, 4)).toBe(4);
  });

  it("keeps a mana X under the energy too", () => {
    const both = ability({ mana_cost: "{X}", demands_x: true, x_slots: 1, energy_cost_x: true });
    expect(suggestedAbilityX(both, 9, 4)).toBe(4);
    expect(suggestedAbilityX(both, 2, 4)).toBe(2);
  });

  it("draws and says the energy component", () => {
    expect(energyCostSymbols(ability({ energy_cost: 2 }))).toBe("{E}{E}");
    expect(energyCostSymbols(sphinx)).toBe("{X}{E}");
    expect(energyCostSymbols(ability())).toBe("");
    expect(energyCostWords(ability({ energy_cost: 2 }))).toBe("pay 2 energy");
    expect(energyCostWords(sphinx)).toBe("pay X energy");
  });
});

describe("beginForAbility carries the announced X into targeting", () => {
  it("locks X alongside the other announce-time choices", () => {
    // CR 602.2b: X is chosen as the ability is activated, before any
    // cost is paid, and cannot change once the targets are picked.
    beginForAbility(
      card(),
      ability({
        demands_x: true,
        min_x: 1,
        legal_targets: { players: ["opp"], cards: [], min: 1, max: 1 },
      }),
      [],
      [],
      4,
    );
    const state = getTargeting();
    expect(state?.ability?.xValue).toBe(4);
    targeting.set(null);
  });

  it("leaves X undefined for an ability that has none", () => {
    beginForAbility(
      card(),
      ability({ legal_targets: { players: ["opp"], cards: [], min: 1, max: 1 } }),
      [],
    );
    const state = getTargeting();
    expect(state?.ability).toBeTruthy();
    expect(state?.ability?.xValue).toBeUndefined();
    targeting.set(null);
  });
});

// #1944: a −X loyalty cost. The ceiling is the loyalty on the
// permanent (CR 606.6), and the opening guess leaves it standing.
describe("a −X loyalty cost", () => {
  const minusX = ability({ loyalty_cost: 0, loyalty_cost_x: true, demands_x: true });

  it("caps X at the planeswalker's loyalty", () => {
    expect(abilityLoyaltyMaxX(minusX, 6)).toBe(6);
    expect(abilityMaxX(minusX, 9, 6)).toBe(6);
    expect(abilityLoyaltyMaxX(ability({ loyalty_cost: -3 }), 6)).toBeUndefined();
  });

  it("suggests the largest X that keeps the planeswalker alive", () => {
    expect(suggestedAbilityX(minusX, 10, 0, 6)).toBe(5);
    expect(suggestedAbilityX(minusX, 10, 0, 0)).toBe(0);
  });

  it("leaves energy abilities to the energy ceiling", () => {
    expect(abilityMaxX(ability({ energy_cost_x: true }), 4, 6)).toBe(4);
    expect(abilityMaxX(ability({ mana_cost: "{X}" }), 4, 6)).toBeUndefined();
  });
});

function getTargeting(): TargetingState | null {
  let out: TargetingState | null = null;
  const unsub = targeting.subscribe((v) => {
    out = v;
  });
  unsub();
  return out;
}
