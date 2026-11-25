import { describe, expect, it } from "vitest";

import {
  castSacrificeFloor,
  castSacrificeRange,
  sacrificeAllWarning,
  sacrificeRangeShortfall,
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

  // #6208: an ACTIVATED ability prints the clause too (Tomb of Urami's
  // "Sacrifice all lands you control"), and the menu row reads the same
  // flag: the cost is paid by whatever it lists, so it is never a
  // shortfall and the row is never greyed for it (CR 118.3).
  it("never greys an ability row, even with nothing to sacrifice", () => {
    expect(
      sacrificeRangeShortfall({ all: true, cards: [], min: 0, max: 0 }, "all lands you control"),
    ).toBe("");
    expect(
      sacrificeRangeShortfall(
        { all: true, cards: ["a", "b"], min: 2, max: 2 },
        "all lands you control",
      ),
    ).toBe("");
    // A clause WITHOUT the flag is still judged against its count.
    expect(sacrificeRangeShortfall({ cards: [], min: 1, max: 1 }, "a creature")).toBe(
      "nothing to sacrifice (a creature)",
    );
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
