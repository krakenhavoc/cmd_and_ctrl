import { describe, expect, it } from "vitest";
import { castPreviewParams, castPreviewParamsFromPayload } from "./castPreview";
import type { CardView } from "./protocol";
import {
  applyCastChoices,
  castAdditionalCost,
  castBlightOffer,
  castSacrificeClause,
  castSacrificeLabel,
  costBranchesOf,
  discardCostOf,
  firstPayableBranch,
} from "./targeting";

// eitherCost.test.ts — ADR 0100 sub-PR 3: an either/or additional cost
// on the client. The branch radio's default, the chosen branch driving
// the discard / sacrifice / blight pickers, and cost_branch on the cast
// payload and the preview query.

function demandAnswers(payable: [boolean, boolean] = [true, true]): CardView {
  return {
    instance_id: "da",
    name: "Demand Answers",
    owner: "p0",
    controller: "p0",
    additional_cost: {
      label: "Sacrifice an artifact or discard a card",
      branches: [
        {
          key: "sacrifice",
          label: "Sacrifice an artifact",
          sacrifice_options: { cards: ["relic"], min: 1, max: 1 },
          payable: payable[0],
        },
        { key: "discard", label: "Discard a card", discard_cards: 1, payable: payable[1] },
      ],
    },
  } as CardView;
}

function wildUnraveling(): CardView {
  return {
    instance_id: "wu",
    name: "Wild Unraveling",
    owner: "p0",
    controller: "p0",
    additional_cost: {
      label: "Blight 2 or pay {1}",
      branches: [
        {
          key: "blight",
          label: "Blight 2",
          blight: 2,
          blight_options: { cards: ["bear"], min: 1, max: 1 },
          payable: true,
        },
        { key: "mana", label: "Pay {1}", mana_cost: "{1}", payable: true },
      ],
    },
  } as CardView;
}

describe("either/or additional costs", () => {
  it("lists the branches, and none on an ordinary card", () => {
    expect(costBranchesOf(demandAnswers())).toHaveLength(2);
    expect(costBranchesOf({ instance_id: "x" } as CardView)).toEqual([]);
  });

  it("defaults the radio to the first payable branch", () => {
    expect(firstPayableBranch(demandAnswers([true, true]))).toBe(0);
    expect(firstPayableBranch(demandAnswers([false, true]))).toBe(1);
    expect(firstPayableBranch(demandAnswers([false, false]))).toBeUndefined();
  });

  it("pays the chosen branch and nothing else", () => {
    const card = demandAnswers();
    expect(castAdditionalCost(card, {})).toBeUndefined();
    expect(discardCostOf(card, { costBranch: 0 })).toBe(0);
    expect(discardCostOf(card, { costBranch: 1 })).toBe(1);
    expect(castSacrificeClause(card, { costBranch: 0 })?.cards).toEqual(["relic"]);
    expect(castSacrificeClause(card, { costBranch: 1 })).toBeUndefined();
    expect(castSacrificeLabel(card, { costBranch: 0 })).toBe("Sacrifice an artifact");
  });

  it("keeps an ordinary card's cost as it was", () => {
    const thrill = {
      instance_id: "t",
      additional_cost: { discard_cards: 1, label: "Discard a card" },
    } as CardView;
    expect(discardCostOf(thrill)).toBe(1);
    expect(castAdditionalCost(thrill, undefined)?.discard_cards).toBe(1);
  });

  it("asks for a creature to blight when the blight branch is chosen", () => {
    const card = wildUnraveling();
    const claimed = castBlightOffer(card, { costBranch: 0 });
    expect(claimed?.n).toBe(2);
    expect(claimed?.options).toEqual(["bear"]);
    expect(claimed?.offer.label).toBe("Blight 2");
    expect(castBlightOffer(card, { costBranch: 1 })).toBeUndefined();
  });

  it("sends cost_branch on the cast, branch 0 included", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { costBranch: 0 });
    expect(params.cost_branch).toBe(0);
    const none: Record<string, unknown> = {};
    applyCastChoices(none, {});
    expect("cost_branch" in none).toBe(false);
  });

  it("prices the preview with the branch", () => {
    expect(castPreviewParams({ costBranch: 1 }).costBranch).toBe(1);
    expect(castPreviewParams({}).costBranch).toBeUndefined();
    expect(castPreviewParamsFromPayload({ cost_branch: 0 }).costBranch).toBe(0);
    expect(castPreviewParamsFromPayload({ cost_branch: -1 }).costBranch).toBeUndefined();
    expect(castPreviewParamsFromPayload({ cost_branch: "1" }).costBranch).toBeUndefined();
  });
});

// ADR 0135 §4: the emerge creature rides the preview, because its mana
// value changes the price.
describe("the alternative cost's payment on the preview", () => {
  it("carries alt_cost_ids both ways", () => {
    expect(castPreviewParams({ altCostIDs: ["a"] }).altCostIDs).toEqual(["a"]);
    expect(castPreviewParams({}).altCostIDs).toBeUndefined();
    expect(castPreviewParamsFromPayload({ alt_cost_ids: ["b"] }).altCostIDs).toEqual(["b"]);
  });
});
