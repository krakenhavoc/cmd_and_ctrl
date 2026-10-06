import { describe, expect, it } from "vitest";
import { MAX_OVERLAP, MIN_CARD_H, fitOverlap, fitScale } from "./rowFit";

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

// #2438: a row shrinks its cards before it overlaps them.
describe("fitScale", () => {
  it("is 1 when the row fits, or before it has a width", () => {
    expect(fitScale(300, 16, 316, 0.4)).toBe(1);
    expect(fitScale(300, 16, 500, 0.4)).toBe(1);
    expect(fitScale(300, 16, 0, 0.4)).toBe(1);
    expect(fitScale(0, 0, 300, 0.4)).toBe(1);
  });

  it("shrinks what scales with the card just enough to fit, leaving the gaps", () => {
    // 5 cards of 100 + 4 gaps of 8 = 532 into 432: (432 - 32) / 500.
    expect(fitScale(500, 32, 432, 0.4)).toBeCloseTo(0.8, 6);
  });

  it("stops at the safe minimum, after which the row overlaps", () => {
    // A 140px card may shrink to MIN_CARD_H and no further.
    const min = MIN_CARD_H / 140;
    expect(fitScale(2000, 0, 300, min)).toBeCloseTo(min, 6);
    // A card already at or under the minimum never grows to it.
    expect(fitScale(2000, 0, 300, MIN_CARD_H / 48)).toBe(1);
  });
});
