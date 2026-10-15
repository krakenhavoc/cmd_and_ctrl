import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import {
  MAX_LINES,
  MAX_OVERLAP,
  MIN_CARD_H,
  WRAP_BELOW,
  allocateRowHeights,
  fitOverlap,
  fitScale,
  fitWrap,
  lineCount,
  lineHeight,
  overrideStale,
  rowDemand,
  scaleForLines,
  stripOpen,
  tappedRoom,
  type RowShape,
} from "./rowFit";

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

// #3028: a shrunk row's override is in px, so a resize that moves the
// row's own size leaves its cards at the old size until it is redone.
describe("overrideStale", () => {
  it("is false when the row is not shrunk", () => {
    expect(overrideStale(1, 0, 144)).toBe(false);
  });

  it("is false while the row's size has not moved", () => {
    expect(overrideStale(0.985, 194, 194)).toBe(false);
    expect(overrideStale(0.985, 194, 194.3)).toBe(false);
  });

  it("is true when the row's size moves under the override, either way", () => {
    expect(overrideStale(0.985, 194, 144)).toBe(true);
    expect(overrideStale(0.99, 83, 194)).toBe(true);
  });

  it("is false for an unmeasured row", () => {
    expect(overrideStale(0.5, 194, 0)).toBe(false);
  });
});

// #3044: a crowded row wraps into the panel's height instead of
// shrinking toward the floor.
describe("lineCount", () => {
  it("breaks lines the way a wrapping flex row does", () => {
    expect(lineCount([], 8, 100)).toBe(0);
    expect(lineCount([100, 100, 100], 8, 316)).toBe(1);
    expect(lineCount([100, 100, 100], 8, 315)).toBe(2);
    expect(lineCount([100, 100, 100], 8, 100)).toBe(3);
    // A pile wider than the row still takes a line of its own.
    expect(lineCount([300, 50], 8, 200)).toBe(2);
  });
});

describe("scaleForLines", () => {
  it("is 1 when the piles already fit on that many lines", () => {
    expect(scaleForLines([100, 100, 100], 8, 316, 1)).toBe(1);
    expect(scaleForLines([100, 100, 100], 8, 208, 2)).toBe(1);
  });

  it("is the largest share at which they fit", () => {
    // Ten piles of 100 on two lines of 400: five a line, 5w + 32 <= 400.
    const s = scaleForLines(Array(10).fill(100), 8, 400, 2);
    expect(s).toBeCloseTo(368 / 500, 4);
    expect(lineCount(Array(10).fill(100 * s), 8, 400)).toBe(2);
  });
});

// A creature row as the owner saw it (#3044): a row 1700px wide, 220px
// cards (157 wide), 24px of padding with the label.
const crowd = (n: number): RowShape => ({
  widths: Array(n).fill(157),
  gap: 8,
  available: 1700,
  lineH: 220,
  pad: 24,
});

describe("rowDemand", () => {
  it("asks for one line when one line draws its cards near full size", () => {
    expect(rowDemand(crowd(8))).toEqual({ base: 244, want: 244 });
  });

  it("asks for the lines it needs at full size, up to MAX_LINES", () => {
    // Ten cards a line at full size: four lines for 39.
    expect(rowDemand(crowd(39), MIN_CARD_H / 220).want).toBe(24 + 4 * 220 + 3 * 8);
    expect(rowDemand(crowd(200)).want).toBe(24 + MAX_LINES * 220 + (MAX_LINES - 1) * 8);
  });

  it("counts its one line at the size it is drawn, so a shrunk row frees height", () => {
    // Fourteen cards on one line at (1700 - 104) / 2198 of full size.
    expect(rowDemand(crowd(14)).base).toBeCloseTo(24 + 220 * (1596 / 2198), 6);
    // 39 would be 0.23, under the floor: the line is drawn at the floor.
    expect(rowDemand(crowd(39), MIN_CARD_H / 220).base).toBeCloseTo(24 + MIN_CARD_H, 6);
  });

  it("is one label high for an empty row", () => {
    expect(rowDemand({ ...crowd(0), lineH: 0 })).toEqual({ base: 24, want: 24 });
  });
});

describe("fitWrap", () => {
  const min = MIN_CARD_H / 220;

  it("keeps a row on one line while one line is at WRAP_BELOW or more", () => {
    // Eleven cards: one line at (1700 - 80) / 1727 = 0.94.
    expect(fitWrap(crowd(11), 1000, min)).toBeNull();
    // Fifteen: (1700 - 112) / 2355 = 0.67, under WRAP_BELOW.
    expect((1700 - 112) / 2355).toBeLessThan(WRAP_BELOW);
    expect(fitWrap(crowd(15), 1000, min)).toEqual({ lines: 2, scale: 1 });
  });

  it("wraps 39 creatures onto two lines within the row's own height", () => {
    // One line would be 0.23 of full size, under the 56px floor. Two
    // lines in 244px: (244 - 24 - 8) / 440 = 0.48, about 106px cards.
    const fit = fitWrap(crowd(39), 244, min)!;
    expect(fit.lines).toBe(2);
    expect(fit.scale).toBeCloseTo(212 / 440, 4);
    expect(fit.scale * 220).toBeGreaterThan(100);
  });

  it("uses spare height for more lines of larger cards", () => {
    const fit = fitWrap(crowd(39), 600, min)!;
    expect(fit.lines).toBe(3);
    expect(fit.scale).toBeGreaterThan(0.75);
    // The lines it picked fit the budget and the width.
    expect(24 + fit.lines * 220 * fit.scale + (fit.lines - 1) * 8).toBeLessThanOrEqual(600);
    expect(lineCount(Array(39).fill(157 * fit.scale), 8, 1700)).toBe(fit.lines);
  });

  it("does not wrap without a budget, or when wrapping gains too little", () => {
    expect(fitWrap(crowd(39), 0, min)).toBeNull();
    // Too short a budget: two lines would be under the floor.
    expect(fitWrap(crowd(39), 120, min)).toBeNull();
    // Fourteen cards: one line at 0.73 of full size, and two lines in
    // 244px only reach 0.48, so it stays on one line.
    expect(fitWrap(crowd(14), 244, min)).toBeNull();
  });

  it("wraps an overlapped row for a small gain: its cards stop overlapping", () => {
    // One line needs 0.23 and is drawn at the 0.25 floor, overlapped.
    // Two lines in 147px: (147 - 32) / 440 = 0.26, every card whole.
    const fit = fitWrap(crowd(39), 147, min)!;
    expect(fit.lines).toBe(2);
    expect(fit.scale).toBeGreaterThanOrEqual(min);
  });

  it("never grows a card past full size", () => {
    expect(fitWrap(crowd(16), 10000, min)).toEqual({ lines: 2, scale: 1 });
  });
});

describe("allocateRowHeights", () => {
  it("gives each row its base when nothing is spare", () => {
    expect(
      allocateRowHeights(380, [
        { base: 244, want: 900 },
        { base: 142, want: 300 },
      ]),
    ).toEqual([244, 142]);
  });

  it("gives every row what it wants when there is room", () => {
    expect(
      allocateRowHeights(1000, [
        { base: 244, want: 500 },
        { base: 142, want: 142 },
      ]),
    ).toEqual([500, 142]);
  });

  it("shares the spare by how much more each row wants, within the total", () => {
    // 200 spare; they want 300 and 100 more, so each gets half of it.
    const [a, b] = allocateRowHeights(500, [
      { base: 200, want: 500 },
      { base: 100, want: 200 },
    ]);
    expect(a).toBeCloseTo(350, 9);
    expect(b).toBeCloseTo(150, 9);
  });
});

// #3893: a board with no creatures keeps the creature row's line, so
// the back row sits where it does beside a creature.
describe("lineHeight", () => {
  it("is the tallest pile at full size", () => {
    expect(lineHeight([110, 121, 99], 0.5, 220)).toBe(242);
  });

  it("is 0 for an empty row that does not hold its line", () => {
    expect(lineHeight([], 1, 220)).toBe(0);
  });

  it("is one card for an empty row that holds its line", () => {
    expect(lineHeight([], 1, 220, true)).toBe(220);
  });

  it("leaves the back row the budget it has beside one creature", () => {
    const back = { base: 142, want: 400 };
    const empty = rowDemand({ ...crowd(0), lineH: lineHeight([], 1, 220, true) });
    const one = rowDemand({ ...crowd(1), lineH: lineHeight([220], 1, 220, true) });
    expect(empty).toEqual(one);
    expect(allocateRowHeights(500, [empty, back])).toEqual(allocateRowHeights(500, [one, back]));
    // One label high, it handed the back row the creatures' height.
    const label = rowDemand({ ...crowd(0), lineH: lineHeight([], 1, 220) });
    expect(allocateRowHeights(500, [label, back])[1]).toBeGreaterThan(
      allocateRowHeights(500, [one, back])[1],
    );
  });
});
