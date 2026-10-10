import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { MAX_OVERLAP, MIN_CARD_H, fitOverlap, fitScale, stripOpen, tappedRoom } from "./rowFit";

// #2960: the land strip uses its room before it overlaps.
describe("stripOpen", () => {
  it("is 0 while the lands fit edge to edge", () => {
    expect(stripOpen(616, 360, 1000)).toBe(0);
    expect(stripOpen(616, 360, 616)).toBe(0);
  });

  it("opens exactly as far as the room allows", () => {
    expect(stripOpen(616, 360, 488)).toBeCloseTo(0.5, 9);
  });

  it("is fully overlapped when even the packed strip does not fit", () => {
    expect(stripOpen(616, 360, 300)).toBe(1);
  });

  it("is steady for a lone pile and an unmeasured row", () => {
    expect(stripOpen(88, 88, 50)).toBe(1);
    expect(stripOpen(616, 360, 0)).toBe(0);
  });
});

// #2442: a tapped card takes the width of its 90 degree turn.
describe("tappedRoom", () => {
  it("is half of what the turned card overhangs its box, per side", () => {
    expect(tappedRoom(88, 123)).toBe(17.5);
    expect(tappedRoom(100, 100)).toBe(0);
    expect(tappedRoom(120, 80)).toBe(0);
  });

  it("scales with the card, so a shrunk row keeps the same proportion", () => {
    expect(tappedRoom(88 * 0.5, 123 * 0.5)).toBeCloseTo(tappedRoom(88, 123) * 0.5, 9);
  });

  it("is counted by the fit: a row of tapped piles overlaps or shrinks sooner", () => {
    const w = 88;
    const h = 123;
    const upright = Array(5).fill(w);
    const turned = Array(5).fill(w + 2 * tappedRoom(w, h));
    // 5 x 88 + 4 x 8 = 472 fits in 480; five turned piles (123 each) do not.
    expect(fitOverlap(upright, 8, 480, w)).toBe(0);
    expect(fitOverlap(turned, 8, 480, w)).toBeGreaterThan(0);
    expect(
      fitScale(
        turned.reduce((a, b) => a + b, 0),
        32,
        480,
        0.4,
      ),
    ).toBeLessThan(1);
  });

  it("is the padding BattlefieldRow gives a tapped pile outside the strip", () => {
    const src = readFileSync("src/lib/components/board/BattlefieldRow.svelte", "utf8");
    expect(src).toMatch(
      /\.row:not\(\.strip\) \.pile\.tapped \{\s*padding-inline: calc\(\(var\(--card-h, 123px\) - var\(--card-w, 88px\)\) \/ 2\);/,
    );
  });
});

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
