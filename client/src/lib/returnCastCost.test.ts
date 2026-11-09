import { describe, expect, it } from "vitest";
import { applyCastChoices, castReturnCost, optionalCostPayOptions } from "./targeting";
import { castPreviewParams, castPreviewParamsFromPayload } from "./castPreview";
import type { CardView } from "./protocol";

// #5247: "return a permanent you control to its owner's hand" as a spell's
// additional cost (Fear of Isolation) — the claimed cost the cast chain
// asks about, and the one wire list.

const fearOfIsolation: CardView = {
  instance_id: "spell",
  name: "Fear of Isolation",
  owner: "me",
  controller: "me",
  additional_cost: {
    label: "Return a permanent you control to its owner's hand",
    return_label: "a permanent you control",
    return_options: { cards: ["island", "bear"], min: 1, max: 1 },
  },
};

describe("castReturnCost", () => {
  it("reads the clause, its count and the permanents that could pay it", () => {
    expect(castReturnCost(fearOfIsolation, {})).toEqual({
      label: "a permanent you control",
      count: 1,
      options: ["island", "bear"],
    });
  });

  it("is undefined for a card whose cost returns nothing", () => {
    const plain: CardView = {
      ...fearOfIsolation,
      additional_cost: { label: "Sacrifice a creature" },
    };
    expect(castReturnCost(plain, {})).toBeUndefined();
    expect(castReturnCost({ ...fearOfIsolation, additional_cost: undefined }, {})).toBeUndefined();
  });

  it("offers an empty list when nothing can pay", () => {
    const none: CardView = {
      ...fearOfIsolation,
      additional_cost: { return_options: { cards: [], min: 1, max: 1 } },
    };
    expect(castReturnCost(none, {})?.options).toEqual([]);
  });
});

describe("the return_ids wire list", () => {
  it("rides cast_spell, omitted when empty", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { returnIDs: ["island"] });
    expect(params.return_ids).toEqual(["island"]);
    const empty: Record<string, unknown> = {};
    applyCastChoices(empty, { returnIDs: [] });
    expect(empty).not.toHaveProperty("return_ids");
  });

  it("is kept by the auto-tap preview, so it plans around the returned permanent", () => {
    expect(castPreviewParams({ returnIDs: ["island"] }).returnIDs).toEqual(["island"]);
    expect(castPreviewParamsFromPayload({ return_ids: ["island"] }).returnIDs).toEqual(["island"]);
  });
});

// #5358: the same cost as a kicker (Chocobo Kick) — asked only once the
// kicker is claimed, and an offer with nothing to return can't be taken.
const chocoboKick: CardView = {
  instance_id: "kick",
  name: "Chocobo Kick",
  owner: "me",
  controller: "me",
  optional_costs: [
    {
      index: 0,
      key: "kicker",
      label: "Kicker—Return a land you control to its owner's hand",
      max_times: 1,
      return_label: "a land you control",
      return_options: { cards: ["forest"], min: 1, max: 1 },
    },
  ],
};

describe("castReturnCost on a return-to-hand kicker", () => {
  it("is undefined until the kicker is claimed", () => {
    expect(castReturnCost(chocoboKick, {})).toBeUndefined();
  });

  it("reads the claimed kicker's clause and options", () => {
    expect(castReturnCost(chocoboKick, { optionalCosts: [0] })).toEqual({
      label: "a land you control",
      count: 1,
      options: ["forest"],
    });
  });

  it("marks the kicker untakeable when nothing can pay it", () => {
    const offer = chocoboKick.optional_costs![0]!;
    expect(optionalCostPayOptions(offer)).toEqual(["forest"]);
    expect(
      optionalCostPayOptions({ ...offer, return_options: { cards: [], min: 1, max: 1 } }),
    ).toEqual([]);
  });
});
