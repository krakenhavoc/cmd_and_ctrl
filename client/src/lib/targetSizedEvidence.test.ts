import { describe, expect, it } from "vitest";
import {
  castExileCardsCost,
  exileCostForTargets,
  targetsManaValue,
  type TargetRef,
} from "./targeting";
import type { AdditionalCostView, CardView } from "./protocol";

// #5240: "collect evidence X, where X is the total mana value of the
// permanents this spell targets" (Urgent Necropsy). The cast chain skips
// the graveyard picker before targeting and sizes it from the picks.

const cost: AdditionalCostView = {
  label: "Collect evidence X, where X is the total mana value of the permanents this spell targets",
  exile_cost_label: "cards from your graveyard with total mana value X or greater",
  exile_cost_options: ["gy1", "gy2"],
  exile_cost_zone: "graveyard",
  exile_cost_from_targets: true,
  exile_cost_target_mana_values: { golem: 4, bear: 3, token: 0 },
};

const necropsy: CardView = {
  instance_id: "spell",
  name: "Urgent Necropsy",
  owner: "me",
  controller: "me",
  additional_cost: cost,
};

const card = (id: string, slot = 0): TargetRef => ({ kind: "card", id, slot });

describe("castExileCardsCost for a target-sized collect evidence", () => {
  it("is marked fromTargets, with no count and the graveyard options", () => {
    const got = castExileCardsCost(necropsy, {});
    expect(got?.fromTargets).toBe(true);
    expect(got?.fromX).toBe(false);
    expect(got?.count).toBe(0);
    expect(got?.options).toEqual(["gy1", "gy2"]);
  });

  it("is not marked on an ordinary exile cost", () => {
    const oculus: CardView = {
      ...necropsy,
      additional_cost: {
        exile_cost_n: 6,
        exile_cost_options: ["a"],
        exile_cost_label: "six cards",
      },
    };
    expect(castExileCardsCost(oculus, {})?.fromTargets).toBeUndefined();
  });
});

describe("targetsManaValue", () => {
  it("sums the server's mana value of each distinct card target", () => {
    expect(targetsManaValue(cost, [card("golem", 0), card("bear", 1)])).toBe(7);
  });

  it("counts a permanent picked for two clauses once", () => {
    expect(targetsManaValue(cost, [card("golem", 0), card("golem", 1)])).toBe(4);
  });

  it("adds nothing for a player, a zero-cost token or an unknown card", () => {
    const player: TargetRef = { kind: "player", id: "bob" };
    expect(targetsManaValue(cost, [player, card("token"), card("ghost")])).toBe(0);
    expect(targetsManaValue(cost, [])).toBe(0);
  });
});

describe("exileCostForTargets", () => {
  it("sets the floor and the label's X from the targets", () => {
    const claimed = castExileCardsCost(necropsy, {})!;
    const sized = exileCostForTargets(claimed, [card("bear", 1)]);
    expect(sized.exileIDs).toBeUndefined();
    expect(sized.cost.manaValueMin).toBe(3);
    expect(sized.cost.label).toBe("cards from your graveyard with total mana value 3 or greater");
  });

  it("answers X = 0 with no cards, so nothing is asked", () => {
    const claimed = castExileCardsCost(necropsy, {})!;
    expect(exileCostForTargets(claimed, []).exileIDs).toEqual([]);
    expect(exileCostForTargets(claimed, [card("token")]).exileIDs).toEqual([]);
  });
});
