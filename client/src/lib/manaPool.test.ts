import { describe, expect, it } from "vitest";
import { manaPoolOf, manaPoolRuns, manaPoolTotal } from "./manaPool";
import type { ManaPoolRun } from "./protocol";

describe("mana_pool runs (#3041)", () => {
  it("totals the counts, not the runs", () => {
    const pool: ManaPoolRun[] = [
      { color: "G", count: 1048576 },
      { color: "R", count: 1 },
    ];
    expect(manaPoolTotal(pool)).toBe(1048577);
    expect(manaPoolTotal(undefined)).toBe(0);
  });

  it("folds adjacent letters of one colour, as the server does", () => {
    expect(manaPoolOf("G", "G", "C", "G")).toEqual([
      { color: "G", count: 2 },
      { color: "C", count: 1 },
      { color: "G", count: 1 },
    ]);
  });

  it("reads a frame recorded before runs, one letter per mana", () => {
    const legacy = ["G", "G", "U"] as unknown as ManaPoolRun[];
    expect(manaPoolRuns(legacy)).toEqual([
      { color: "G", count: 1 },
      { color: "G", count: 1 },
      { color: "U", count: 1 },
    ]);
    expect(manaPoolTotal(legacy)).toBe(3);
  });
});
