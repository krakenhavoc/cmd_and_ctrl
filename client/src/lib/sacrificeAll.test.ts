import { describe, expect, it } from "vitest";

import {
  castSacrificeFloor,
  castSacrificeRange,
  sacrificeAllWarning,
  sacrificesAll,
} from "./sacrificeCost";

// #2097: "As an additional cost to cast this spell, sacrifice all
// creatures you control" (Soulblast). The server ships the clause with
// `all: true` and min = max = the number of permanents listed.

describe("sacrifice all", () => {
  it("is read off the view's all flag", () => {
    expect(sacrificesAll({ all: true, cards: ["a"], min: 1, max: 1 })).toBe(true);
    expect(sacrificesAll({ cards: ["a"], min: 1, max: 1 })).toBe(false);
    expect(sacrificesAll(undefined)).toBe(false);
  });

  it("never blocks the cast: an empty board pays it with nothing", () => {
    expect(castSacrificeFloor({ all: true, cards: [], min: 0, max: 0 })).toBe(0);
    expect(castSacrificeFloor({ all: true, cards: ["a", "b"], min: 2, max: 2 })).toBe(0);
  });

  it("bounds the sheet at exactly what it lists", () => {
    expect(castSacrificeRange({ all: true, cards: ["a", "b", "c"], min: 3, max: 3 })).toEqual({
      min: 3,
      max: 3,
    });
    expect(castSacrificeRange({ all: true, cards: [], min: 0, max: 0 })).toEqual({
      min: 0,
      max: 0,
    });
  });

  it("warns how many permanents go", () => {
    expect(sacrificeAllWarning(1)).toBe(
      "This sacrifices the one permanent below. You don't choose.",
    );
    expect(sacrificeAllWarning(4)).toBe(
      "This sacrifices all 4 permanents below. You don't choose.",
    );
  });
});
