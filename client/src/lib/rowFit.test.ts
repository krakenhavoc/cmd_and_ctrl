import { describe, expect, it } from "vitest";
import { MAX_OVERLAP, fitOverlap } from "./rowFit";

describe("fitOverlap", () => {
  it("is 0 when the row fits, or has fewer than two piles", () => {
    expect(fitOverlap([100, 100, 100], 8, 316, 100)).toBe(0);
    expect(fitOverlap([400], 8, 100, 100)).toBe(0);
    expect(fitOverlap([], 8, 100, 100)).toBe(0);
  });

  it("overlaps each pile after the first by the same amount, just enough to fit", () => {
    // 4 × 150 + 3 × 8 = 624 natural, into 600: 24 over 3 joins.
    expect(fitOverlap([150, 150, 150, 150], 8, 600, 150)).toBe(8);
  });

  it("stops at the cap, so a huge board still shows each card", () => {
    const px = fitOverlap(Array(30).fill(100), 8, 300, 100);
    expect(px).toBe(100 * MAX_OVERLAP + 8);
  });

  it("does nothing before the row has a width", () => {
    expect(fitOverlap([100, 100], 8, 0, 100)).toBe(0);
  });
});
