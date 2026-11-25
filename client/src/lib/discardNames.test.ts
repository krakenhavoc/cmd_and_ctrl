import { describe, expect, it } from "vitest";

import {
  canPickDifferentNames,
  canPickSameName,
  cardNameKeys,
  discardNameJoins,
  discardNameRepeats,
  discardNamesDiffer,
  discardNamesShared,
} from "./discardNames";

const card = (instance_id: string, name: string) => ({ instance_id, name });

describe("discard with different names (#4896)", () => {
  const hand = [
    card("a", "Opt"),
    card("b", "Opt"),
    card("c", "Shock"),
    card("d", "Fire // Ice"),
    card("e", "Ice"),
  ];

  it("reads both halves of a split card", () => {
    expect(cardNameKeys(card("d", "Fire // Ice"))).toEqual(["fire", "ice"]);
    expect(cardNameKeys(card("x", ""))).toEqual([]);
  });

  it("greys a card whose name is already picked, and never the pick itself", () => {
    expect(discardNameRepeats(hand, ["a"], "b")).toBe(true);
    expect(discardNameRepeats(hand, ["a"], "a")).toBe(false);
    expect(discardNameRepeats(hand, ["a"], "c")).toBe(false);
    expect(discardNameRepeats(hand, ["d"], "e")).toBe(true);
    expect(discardNameRepeats(hand, ["e"], "d")).toBe(true);
  });

  it("counts the names a hand can pay", () => {
    expect(canPickDifferentNames(hand, 3)).toBe(true);
    expect(canPickDifferentNames(hand, 4)).toBe(false);
    expect(canPickDifferentNames([card("a", "Opt"), card("b", "Opt")], 2)).toBe(false);
    // Fire // Ice first would block Fire and Ice; the search gives it up.
    expect(
      canPickDifferentNames([card("d", "Fire // Ice"), card("f", "Fire"), card("e", "Ice")], 2),
    ).toBe(true);
  });

  it("checks a whole pick", () => {
    expect(discardNamesDiffer(hand, ["a", "c", "e"])).toBe(true);
    expect(discardNamesDiffer(hand, ["a", "b", "c"])).toBe(false);
    expect(discardNamesDiffer(hand, ["c", "d", "e"])).toBe(false);
  });
});

describe("discard with the same name (#6264)", () => {
  const hand = [
    card("a", "Opt"),
    card("b", "Opt"),
    card("c", "Shock"),
    card("d", "Fire // Ice"),
    card("e", "Ice"),
    card("x", ""),
  ];

  it("greys a card that shares no name with the picks, and never the pick itself", () => {
    expect(discardNameJoins(hand, ["a"], "b")).toBe(true);
    expect(discardNameJoins(hand, ["a"], "c")).toBe(false);
    expect(discardNameJoins(hand, ["a"], "a")).toBe(true);
    expect(discardNameJoins(hand, ["d"], "e")).toBe(true);
    expect(discardNameJoins(hand, [], "x")).toBe(false);
  });

  it("counts whether a hand can pay", () => {
    expect(canPickSameName(hand, 2)).toBe(true);
    expect(canPickSameName(hand, 3)).toBe(false);
    expect(canPickSameName([card("c", "Shock"), card("f", "Fire")], 2)).toBe(false);
  });

  it("checks a whole pick", () => {
    expect(discardNamesShared(hand, ["a", "b"])).toBe(true);
    expect(discardNamesShared(hand, ["d", "e"])).toBe(true);
    expect(discardNamesShared(hand, ["a", "c"])).toBe(false);
    expect(discardNamesShared(hand, [])).toBe(false);
  });
});
