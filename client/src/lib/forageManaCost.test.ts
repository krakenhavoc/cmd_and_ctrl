import { describe, expect, it } from "vitest";
import { payableCostBranches, withCostBranch } from "./costBranch";
import { manaAbilityNeedsPrompt, manaCostBranchPayment } from "./manaAbilityCost";
import { manaAbilityRider } from "./manaSource";
import type { ManaAbilityView } from "./protocol";

// #5397: forage as a mana ability's cost — Thornvault Forager's "{T},
// Forage: Add two mana in any combination of colors". The row ships the
// same `cost_branches` an activated row does; the branch folds into the
// mana row's own exile or sacrifice fields, so the existing pickers run,
// and its index rides activate_mana_ability as `cost_branch`.

const forage: ManaAbilityView = {
  index: 1,
  ref: "own:1",
  label: "{T}, Forage: Add two mana in any combination of colors.",
  tap_cost: true,
  cost_branches: [
    {
      label: "Exile three cards from your graveyard",
      payable: true,
      exile_cost_n: 3,
      exile_cost_label: "three cards",
      exile_cost_options: ["g1", "g2", "g3", "g4"],
      exile_cost_zone: "graveyard",
    },
    {
      label: "Sacrifice a Food",
      payable: false,
      sacrifice_label: "a Food",
      sacrifice_options: { cards: [], min: 1, max: 1 },
    },
  ],
} as ManaAbilityView;

describe("forage on a mana ability", () => {
  it("needs a prompt, and the plain row does not", () => {
    expect(manaAbilityNeedsPrompt(forage)).toBe(true);
    expect(
      manaAbilityNeedsPrompt({ index: 0, ref: "own:0", tap_cost: true } as ManaAbilityView),
    ).toBe(false);
  });

  it("offers only the payable branch", () => {
    expect(payableCostBranches(forage)).toEqual([0]);
  });

  it("folds the graveyard branch into the row's exile fields", () => {
    const folded = withCostBranch(forage, 0);
    expect(folded.cost_branches).toBeUndefined();
    expect(folded.exile_cost_n).toBe(3);
    expect(folded.exile_cost_zone).toBe("graveyard");
    expect(folded.exile_cost_options).toEqual(["g1", "g2", "g3", "g4"]);
    expect(folded.sacrifice_options).toBeUndefined();
  });

  it("folds the Food branch into the row's sacrifice fields", () => {
    const folded = withCostBranch(forage, 1);
    expect(folded.sacrifice_label).toBe("a Food");
    expect(folded.exile_cost_n).toBeUndefined();
  });

  it("sends cost_branch only when a branch was chosen", () => {
    expect(manaCostBranchPayment(0)).toEqual({ cost_branch: 0 });
    expect(manaCostBranchPayment(1)).toEqual({ cost_branch: 1 });
    expect(manaCostBranchPayment(undefined)).toEqual({});
  });

  it("names the forage in the source picker's rider", () => {
    expect(manaAbilityRider(forage)).toBe(
      "exile three cards from your graveyard or sacrifice a Food",
    );
  });
});
