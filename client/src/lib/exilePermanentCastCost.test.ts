import { describe, expect, it } from "vitest";
import { applyCastChoices, castExilePermanentCost } from "./targeting";
import { castPreviewParams, castPreviewParamsFromPayload } from "./castPreview";
import type { CardView } from "./protocol";

// #3402: "exile a creature you control" as a spell's additional cost
// (Soul Exchange) — the claimed cost the cast chain asks about, and the
// one wire list.

const soulExchange: CardView = {
  instance_id: "spell",
  name: "Soul Exchange",
  owner: "me",
  controller: "me",
  additional_cost: {
    label: "Exile a creature you control",
    exile_permanent_label: "a creature you control",
    exile_permanent_options: { cards: ["thrull", "bear"], min: 1, max: 1 },
  },
};

describe("castExilePermanentCost", () => {
  it("reads the clause, its count and the permanents that could pay it", () => {
    expect(castExilePermanentCost(soulExchange, {})).toEqual({
      label: "a creature you control",
      count: 1,
      options: ["thrull", "bear"],
    });
  });

  it("is undefined for a card whose cost exiles nothing", () => {
    const plain: CardView = { ...soulExchange, additional_cost: { label: "Sacrifice a creature" } };
    expect(castExilePermanentCost(plain, {})).toBeUndefined();
    expect(castExilePermanentCost({ ...soulExchange, additional_cost: undefined }, {})).toBe(
      undefined,
    );
  });

  it("offers an empty list when nothing can pay", () => {
    const none: CardView = {
      ...soulExchange,
      additional_cost: { exile_permanent_options: { cards: [], min: 1, max: 1 } },
    };
    expect(castExilePermanentCost(none, {})?.options).toEqual([]);
  });
});

describe("the exile_permanent_ids wire list", () => {
  it("rides cast_spell, omitted when empty", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { exilePermanentIDs: ["thrull"] });
    expect(params.exile_permanent_ids).toEqual(["thrull"]);
    const empty: Record<string, unknown> = {};
    applyCastChoices(empty, { exilePermanentIDs: [] });
    expect(empty).not.toHaveProperty("exile_permanent_ids");
  });

  it("is kept by the auto-tap preview so the exiled creature is not tapped for mana", () => {
    expect(castPreviewParams({ exilePermanentIDs: ["thrull"] }).exilePermanentIDs).toEqual([
      "thrull",
    ]);
    expect(
      castPreviewParamsFromPayload({ exile_permanent_ids: ["thrull"] }).exilePermanentIDs,
    ).toEqual(["thrull"]);
  });
});
