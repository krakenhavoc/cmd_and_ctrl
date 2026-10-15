import { describe, expect, it } from "vitest";
import { altCostPayDefinesX, altCostPayX, applyCastChoices, castHandExileCost } from "./targeting";
import { castPreviewParams, castPreviewParamsFromPayload } from "./castPreview";
import type { AlternativeCostView, CardView } from "./protocol";

// #3175: the two hand-exile costs. March of Swirling Mist's "you may
// exile any number of blue cards from your hand" (a picker of zero or
// more, sent as exile_from_hand_ids), and Shining Shoal's pitch of "a
// white card with mana value X", whose pick IS the X announced.

const march: CardView = {
  instance_id: "march",
  name: "March of Swirling Mist",
  owner: "me",
  controller: "me",
  additional_cost: {
    label: "Exile any number of blue cards from your hand",
    exile_from_hand_label: "any number of blue cards from your hand",
    exile_from_hand_options: { cards: ["drake", "sphinx"], min: 0, max: 2 },
  },
};

describe("castHandExileCost", () => {
  it("reads the clause and the hand cards that could pay it", () => {
    expect(castHandExileCost(march, {})).toEqual({
      label: "any number of blue cards from your hand",
      options: ["drake", "sphinx"],
    });
  });

  it("is undefined for a card whose cost exiles nothing from hand", () => {
    expect(castHandExileCost({ ...march, additional_cost: undefined }, {})).toBeUndefined();
    expect(
      castHandExileCost({ ...march, additional_cost: { label: "Discard a card" } }, {}),
    ).toBeUndefined();
  });

  it("offers an empty list when the hand holds nothing that pays", () => {
    const none: CardView = {
      ...march,
      additional_cost: { exile_from_hand_options: { cards: [], min: 0, max: 0 } },
    };
    expect(castHandExileCost(none, {})?.options).toEqual([]);
  });
});

describe("the exile_from_hand_ids wire list", () => {
  it("rides cast_spell, omitted when none were exiled", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { exileFromHandIDs: ["drake"] });
    expect(params.exile_from_hand_ids).toEqual(["drake"]);
    const none: Record<string, unknown> = {};
    applyCastChoices(none, { exileFromHandIDs: [] });
    expect(none).not.toHaveProperty("exile_from_hand_ids");
  });

  it("reaches the auto-tap preview, since each card lowers the price", () => {
    expect(castPreviewParams({ exileFromHandIDs: ["drake"] }).exileFromHandIDs).toEqual(["drake"]);
    expect(
      castPreviewParamsFromPayload({ exile_from_hand_ids: ["drake"] }).exileFromHandIDs,
    ).toEqual(["drake"]);
  });
});

describe("a pitch bound to X (Shining Shoal)", () => {
  const pitch: AlternativeCostView = {
    key: "pitch",
    label: "Exile a white card with mana value X from your hand",
    pay_options: {
      cards: ["lions", "angel"],
      min: 1,
      max: 1,
      mana_value_equals_x: true,
      mana_values: { lions: 1, angel: 5 },
    },
  };

  it("announces the picked card's mana value as X", () => {
    expect(altCostPayDefinesX(pitch)).toBe(true);
    expect(altCostPayX(pitch, ["angel"])).toBe(5);
    expect(altCostPayX(pitch, ["lions"])).toBe(1);
  });

  it("announces nothing for an ordinary pitch or an empty pick", () => {
    const fow: AlternativeCostView = {
      key: "pitch",
      pay_options: { cards: ["x"], min: 1, max: 1 },
    };
    expect(altCostPayDefinesX(fow)).toBe(false);
    expect(altCostPayX(fow, ["x"])).toBeUndefined();
    expect(altCostPayX(pitch, [])).toBeUndefined();
    expect(altCostPayX(undefined, ["angel"])).toBeUndefined();
  });
});
