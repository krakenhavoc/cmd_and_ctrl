import { describe, expect, it } from "vitest";
import { fanAngle, fanLift, handOverlap } from "./handFan";

describe("fanAngle", () => {
  it("is flat for zero or one card", () => {
    expect(fanAngle(0, 0)).toBe(0);
    expect(fanAngle(0, 1)).toBe(0);
  });

  it("is symmetric about the centre", () => {
    expect(fanAngle(0, 5)).toBeCloseTo(-fanAngle(4, 5));
    expect(fanAngle(2, 5)).toBeCloseTo(0);
  });

  it("caps the total spread at 60 degrees", () => {
    const n = 40;
    expect(fanAngle(n - 1, n) - fanAngle(0, n)).toBeLessThanOrEqual(60.0001);
  });

  it("uses the 7 degree step until the cap binds", () => {
    expect(fanAngle(1, 3) - fanAngle(0, 3)).toBeCloseTo(7);
  });
});

describe("fanLift", () => {
  it("lifts the outer cards and leaves the centre alone", () => {
    expect(fanLift(2, 5)).toBe(0);
    expect(fanLift(0, 5)).toBe(6);
    expect(fanLift(4, 5)).toBe(6);
  });
});

// The fan's width in card-widths: the first card, plus the part of
// each later card that is not pulled back over its predecessor.
const fanWidth = (n: number, base = 0.5): number => 1 + (n - 1) * (1 - handOverlap(n, base));

describe("handOverlap", () => {
  it("leaves small hands at the layout's resting overlap", () => {
    expect(handOverlap(1, 0.5)).toBe(0.5);
    expect(handOverlap(7, 0.5)).toBe(0.5);
    expect(handOverlap(7, 0.62)).toBe(0.62);
  });

  it("tightens past seven cards", () => {
    expect(handOverlap(8, 0.5)).toBeCloseTo(0.545);
    expect(handOverlap(12, 0.5)).toBeCloseTo(0.725);
  });

  it("never exceeds the cap", () => {
    expect(handOverlap(40, 0.5)).toBe(0.85);
    expect(handOverlap(40, 0.85)).toBe(0.85);
  });

  it("keeps the fan inside 4.5 card-widths for every hand size that occurs", () => {
    // #956 is a 7-14 card hand in a half-width panel; 20 is already
    // past anything a Commander game produces.
    for (const n of [1, 5, 7, 10, 14, 20]) {
      expect(fanWidth(n)).toBeLessThanOrEqual(4.5);
    }
  });

  it("widens again past the cap, which is the accepted trade", () => {
    // Once CAP binds (~15 cards) the overlap stops tightening, so the
    // width resumes growing at 0.15 card-widths per card. Holding 4.5
    // out here would need a 0.94 overlap — a 6% sliver of each card.
    // This records the boundary so it is a decision, not a surprise.
    expect(fanWidth(20)).toBeLessThanOrEqual(4.5);
    expect(fanWidth(60)).toBeGreaterThan(4.5);
  });
});
