import { describe, it, expect } from "vitest";

import { discardedManaValue } from "./discardCostX";
import { manaValueOf } from "./targetPrices";
import type { CardView } from "./protocol";

// #2190: "Discard a card with mana value X" (Kozilek, the Great
// Distortion). The card the player picks IS the announcement: the
// client sends its mana value as x_value and narrows the target clause
// by it, and the engine refuses any other number.

function card(id: string, manaCost?: string): CardView {
  return {
    instance_id: id,
    name: id,
    owner: "me",
    controller: "me",
    mana_cost: manaCost,
  } as CardView;
}

const hand = [
  card("land"),
  card("one", "{G}"),
  card("three", "{2}{G}"),
  card("x-spell", "{X}{R}"),
  card("twobrid", "{2/W}{2/W}"),
];

describe("discardedManaValue", () => {
  it("is the mana value of the one card picked", () => {
    expect(discardedManaValue(hand, ["land"])).toBe(0);
    expect(discardedManaValue(hand, ["one"])).toBe(1);
    expect(discardedManaValue(hand, ["three"])).toBe(3);
  });

  it("counts X as zero in a hand, as the engine does", () => {
    expect(discardedManaValue(hand, ["x-spell"])).toBe(1);
  });

  it("counts a monocoloured hybrid as its larger half (CR 202.3f)", () => {
    expect(manaValueOf("{2/W}")).toBe(2);
    expect(discardedManaValue(hand, ["twobrid"])).toBe(4);
  });

  it("is undefined when nothing, several cards, or a stranger was picked", () => {
    expect(discardedManaValue(hand, [])).toBeUndefined();
    expect(discardedManaValue(hand, ["one", "three"])).toBeUndefined();
    expect(discardedManaValue(hand, ["not-in-hand"])).toBeUndefined();
    expect(discardedManaValue(undefined, ["one"])).toBeUndefined();
  });
});
