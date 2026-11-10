import { describe, it, expect } from "vitest";

import { exiledTotalManaValue } from "./exileTotalX";
import type { CardView } from "./protocol";

// #5401: "Exile one or more other artifacts you control with total mana
// value X" (Fabrication Foundry). The permanents the player picks ARE the
// announcement: the client sends their total mana value as x_value and
// narrows the target clause by it, and the engine refuses any other
// number.

function card(id: string, manaCost?: string): CardView {
  return {
    instance_id: id,
    name: id,
    owner: "me",
    controller: "me",
    mana_cost: manaCost,
  } as CardView;
}

const battlefield = [
  card("treasure"),
  card("two", "{2}"),
  card("three", "{1}{W}{W}"),
  card("x-artifact", "{X}{X}"),
];

describe("exiledTotalManaValue", () => {
  it("adds the picked permanents' mana values", () => {
    expect(exiledTotalManaValue(battlefield, ["two", "three"])).toBe(5);
  });

  it("counts a token with no mana cost as zero", () => {
    expect(exiledTotalManaValue(battlefield, ["treasure"])).toBe(0);
    expect(exiledTotalManaValue(battlefield, ["treasure", "two"])).toBe(2);
  });

  it("counts {X} on a permanent as zero", () => {
    expect(exiledTotalManaValue(battlefield, ["x-artifact"])).toBe(0);
  });

  it("is undefined with nothing picked or a pick off the battlefield", () => {
    expect(exiledTotalManaValue(battlefield, [])).toBeUndefined();
    expect(exiledTotalManaValue(battlefield, ["two", "gone"])).toBeUndefined();
    expect(exiledTotalManaValue(undefined, ["two"])).toBeUndefined();
  });
});
