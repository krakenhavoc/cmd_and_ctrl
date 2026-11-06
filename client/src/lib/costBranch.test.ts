import { describe, expect, it } from "vitest";
import {
  abilityCostBranches,
  costBranchModeSpec,
  payableCostBranches,
  withCostBranch,
} from "./costBranch";
import { abilityBlocked } from "./contextMenu.logic";
import { branchLabel, branchNamesCards, branchReady, branchWhyNot } from "./payEither";
import type { ActivatedAbilityView, PayOptionView } from "./protocol";

// #4566: Bullseye, Death Dealer's "{3}, {T}, Sacrifice an artifact or
// discard a nonland card", and Titania's "Ward—Discard a card or pay {2}".

const row: ActivatedAbilityView = {
  index: 0,
  ref: "r0",
  label:
    "{3}, {T}, Sacrifice an artifact or discard a nonland card: Bullseye deals 2 damage to any target.",
  mana_cost: "{3}",
  tap_cost: true,
  cost_branches: [
    {
      label: "Sacrifice an artifact",
      payable: true,
      sacrifice_label: "an artifact",
      sacrifice_options: { cards: ["a1"], min: 1, max: 1 },
    },
    {
      label: "Discard a nonland card",
      payable: false,
      discard_cost_n: 1,
      discard_cost_label: "a nonland card",
      discard_cost_options: [],
    },
  ],
} as ActivatedAbilityView;

describe("either/or activation costs", () => {
  it("lists the branches and which can be paid", () => {
    expect(abilityCostBranches(row)).toHaveLength(2);
    expect(abilityCostBranches({ ...row, cost_branches: undefined })).toHaveLength(0);
    expect(payableCostBranches(row)).toEqual([0]);
  });

  it("asks the branch as a one-of question with the unpayable one disabled", () => {
    const spec = costBranchModeSpec(row);
    expect(spec.min).toBe(1);
    expect(spec.max).toBe(1);
    expect(spec.options.map((o) => o.label)).toEqual([
      "Sacrifice an artifact",
      "Discard a nonland card",
    ]);
    expect(spec.options.map((o) => o.used)).toEqual([false, true]);
  });

  it("folds the chosen branch into the row as its own clause", () => {
    const sac = withCostBranch(row, 0);
    expect(sac.cost_branches).toBeUndefined();
    expect(sac.sacrifice_options?.cards).toEqual(["a1"]);
    expect(sac.discard_cost_n).toBeUndefined();
    const disc = withCostBranch(row, 1);
    expect(disc.discard_cost_n).toBe(1);
    expect(disc.discard_cost_label).toBe("a nonland card");
    expect(disc.sacrifice_options).toBeUndefined();
    expect(withCostBranch(row, 5)).toBe(row);
  });

  it("greys the row in the menu only when no branch can be paid", () => {
    expect(abilityBlocked(row, false, false)).toBe("");
    const none = {
      ...row,
      cost_branches: row.cost_branches!.map((b) => ({ ...b, payable: false })),
    };
    expect(abilityBlocked(none, false, false)).toBe(
      "needs: sacrifice an artifact or discard a nonland card",
    );
  });
});

// #5184: Camellia, the Seedmiser's "{2}, Forage:" — exile three cards
// from your graveyard, or sacrifice a Food.
describe("a forage cost", () => {
  const forage: ActivatedAbilityView = {
    index: 0,
    ref: "r0",
    label: "{2}, Forage: Put a +1/+1 counter on each other Squirrel you control.",
    mana_cost: "{2}",
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
  } as ActivatedAbilityView;

  it("folds the graveyard branch into the row's exile picker", () => {
    const exile = withCostBranch(forage, 0);
    expect(exile.cost_branches).toBeUndefined();
    expect(exile.exile_cost_n).toBe(3);
    expect(exile.exile_cost_label).toBe("three cards");
    expect(exile.exile_cost_options).toEqual(["g1", "g2", "g3", "g4"]);
    expect(exile.exile_cost_zone).toBe("graveyard");
    expect(exile.sacrifice_options).toBeUndefined();
    const food = withCostBranch(forage, 1);
    expect(food.exile_cost_n).toBeUndefined();
    expect(food.sacrifice_label).toBe("a Food");
    expect(payableCostBranches(forage)).toEqual([0]);
  });
});

describe("a ward's mana branch", () => {
  const mana: PayOptionView = { action: "mana", words: "pay {2}", payable: true };
  it("is answered without picking cards", () => {
    expect(branchNamesCards(mana)).toBe(false);
    expect(branchReady(mana, [])).toBe(true);
    expect(branchLabel(mana)).toBe("Pay {2}");
  });
  it("says why when it can't be paid", () => {
    expect(branchWhyNot({ ...mana, payable: false })).toBe("You don't have enough to pay {2}.");
  });
});
