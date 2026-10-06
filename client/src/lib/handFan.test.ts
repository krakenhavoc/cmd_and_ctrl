import { describe, expect, it } from "vitest";
import {
  FAN_MIN_SLIVER,
  FAN_SNUG,
  fanAngle,
  fanLift,
  fanOverhang,
  fitFan,
  handOverlap,
} from "./handFan";

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

describe("fanOverhang (#2395)", () => {
  it("is nothing for an upright fan", () => {
    expect(fanOverhang(1, 168, 235)).toBe(0);
    expect(fanOverhang(7, 168, 235, 0)).toBe(0);
  });

  it("is how far the end card's top corner swings past its upright box", () => {
    // Seven cards: the end card turns 21 degrees about its bottom centre.
    const rad = (21 * Math.PI) / 180;
    expect(fanOverhang(7, 168, 235)).toBeCloseTo(235 * Math.sin(rad) - 84 * (1 - Math.cos(rad)));
    // Less tilt, less overhang.
    expect(fanOverhang(7, 168, 235, 0.5)).toBeLessThan(fanOverhang(7, 168, 235));
  });
});

// The width a fitted fan takes: its upright width plus both end cards'
// corners.
const fittedWidth = (n: number, w: number, h: number, f: { overlap: number; tilt: number }) =>
  w + (n - 1) * w * (1 - f.overlap) + 2 * fanOverhang(n, w, h, f.tilt);

describe("fitFan (#2395)", () => {
  it("keeps the resting fan when the row has room, or is not measured yet", () => {
    expect(fitFan(7, 0.5, 1800, 168, 235)).toEqual({ overlap: 0.5, tilt: 1, scroll: false });
    expect(fitFan(7, 0.5, 0, 168, 235)).toEqual({ overlap: 0.5, tilt: 1, scroll: false });
    expect(fitFan(7, 0.5, 800, 0, 0)).toEqual({ overlap: 0.5, tilt: 1, scroll: false });
    expect(fitFan(1, 0.5, 50, 168, 235)).toEqual({ overlap: 0.5, tilt: 1, scroll: false });
  });

  it("tightens a fan that would run out of its row, corners included", () => {
    // 1440x900 without the coach: the row is 800px.
    const f = fitFan(7, 0.5, 800, 168, 235);
    expect(f.tilt).toBe(1);
    expect(f.overlap).toBeGreaterThan(0.5);
    expect(f.overlap).toBeLessThanOrEqual(FAN_SNUG);
    expect(fittedWidth(7, 168, 235, f)).toBeCloseTo(800);
  });

  it("eases off the tilt before it goes past FAN_SNUG", () => {
    // 1280x720 beside the coach card: seven 120px cards in 410px.
    const f = fitFan(7, 0.5, 410, 120, 168);
    expect(f.tilt).toBeGreaterThan(0);
    expect(f.tilt).toBeLessThan(1);
    expect(f.overlap).toBeLessThanOrEqual(FAN_SNUG);
    expect(fittedWidth(7, 120, 168, f)).toBeLessThanOrEqual(410.0001);
  });

  it("goes flat and tighter still in the issue's row: seven 168px cards in 452px", () => {
    // 1440x900 beside the coach card (#2395).
    const f = fitFan(7, 0.5, 452, 168, 235);
    expect(f).toMatchObject({ tilt: 0, scroll: false });
    expect(f.overlap).toBeGreaterThan(FAN_SNUG);
    expect(f.overlap).toBeLessThanOrEqual(1 - FAN_MIN_SLIVER);
    expect(fittedWidth(7, 168, 235, f)).toBeCloseTo(452);
  });

  it("scrolls when even the thinnest flat sliver does not fit", () => {
    const f = fitFan(30, 0.5, 452, 168, 235);
    expect(f).toEqual({ overlap: 1 - FAN_MIN_SLIVER, tilt: 0, scroll: true });
  });

  it("never loosens a large hand past handOverlap", () => {
    const f = fitFan(14, 0.5, 5000, 168, 235);
    expect(f.overlap).toBe(handOverlap(14, 0.5));
  });

  it("has no tilt to give up in the stacked layout", () => {
    const f = fitFan(7, 0.85, 300, 168, 235, false);
    expect(f.overlap).toBeGreaterThan(0.85);
    expect(f.scroll).toBe(false);
    expect(168 + 6 * 168 * (1 - f.overlap)).toBeCloseTo(300);
  });
});
