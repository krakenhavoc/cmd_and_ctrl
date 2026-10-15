import { describe, expect, it } from "vitest";
import { STRIP_DEPTH, namePileKey, pileDepths } from "./landStrip";
import type { CardView } from "./protocol";

// #3044: a land pile reads as one card and a count at any size.
describe("pileDepths", () => {
  it("steps every copy of a small pile, as before", () => {
    expect(pileDepths(1)).toEqual([0]);
    expect(pileDepths(3)).toEqual([0, 1, 2]);
    expect(pileDepths(STRIP_DEPTH)).toEqual([0, 1, 2, 3]);
  });

  it("caps a big pile's stack at STRIP_DEPTH steps, the rest under the bottom copy", () => {
    const d = pileDepths(28);
    expect(d).toHaveLength(28);
    expect(Math.max(...d)).toBe(STRIP_DEPTH - 1);
    expect(d.filter((x) => x === 0)).toHaveLength(28 - STRIP_DEPTH + 1);
    // The top copy, last in the DOM, is the highest step.
    expect(d[27]).toBe(STRIP_DEPTH - 1);
  });
});

describe("namePileKey", () => {
  const forest = (extra: Partial<CardView> = {}): CardView => ({
    instance_id: "f1",
    name: "Forest",
    owner: "me",
    controller: "me",
    type_line: "Basic Land — Forest",
    ...extra,
  });

  it("piles by name, and keeps tapped and untapped lands apart", () => {
    expect(namePileKey(forest(), false)).toBe("pile:Forest");
    expect(namePileKey(forest({ tapped: true }), false)).toBe("tapped:Forest");
  });

  it("leaves a commander or an enchanted land on its own", () => {
    expect(namePileKey(forest({ is_commander: true }), false)).toBeNull();
    expect(namePileKey(forest(), true)).toBeNull();
  });
});
