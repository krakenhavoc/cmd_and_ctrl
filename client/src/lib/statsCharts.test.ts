import { describe, expect, it } from "vitest";

import {
  barShares,
  columns,
  evennessWord,
  linearScale,
  linePath,
  linePoints,
  niceMax,
  plotBox,
  sparkCells,
  stackSegments,
  ticks,
  valueX,
} from "./statsCharts";

describe("statsCharts", () => {
  const plot = plotBox(100, 60, { top: 10, right: 0, bottom: 10, left: 0 });

  it("maps linearly, and a flat domain to the range's start", () => {
    const s = linearScale([0, 10], [0, 100]);
    expect(s(5)).toBe(50);
    expect(linearScale([3, 3], [7, 9])(3)).toBe(7);
  });

  it("rounds an axis top up to a readable value", () => {
    expect([0, -1, 0.7, 3, 7, 12, 26, 240].map(niceMax)).toEqual([1, 1, 1, 5, 10, 20, 50, 250]);
    expect(ticks(7, 4)).toEqual([0, 2, 4, 6, 8]);
    expect(ticks(3.8, 4)).toEqual([0, 1, 2, 3, 4]);
    expect(ticks(9, 4)).toEqual([0, 2.5, 5, 7.5, 10]);
  });

  it("gives bar shares of the widest, clamped", () => {
    expect(barShares([2, 4, 0])).toEqual([0.5, 1, 0]);
    expect(barShares([0, 0])).toEqual([0, 0]);
    expect(barShares([5], 4)).toEqual([1]);
  });

  it("stacks a 100% bar, skipping empty parts", () => {
    const segs = stackSegments([1, 0, 3]);
    expect(segs.map((s) => [s.index, s.start, s.share])).toEqual([
      [0, 0, 0.25],
      [2, 0.25, 0.75],
    ]);
    expect(stackSegments([0, 0])).toEqual([]);
  });

  it("lays columns out from the plot's bottom", () => {
    const [a, b] = columns(plot, [5, 10]);
    expect(b.y).toBe(plot.top);
    expect(b.height).toBe(plot.bottom - plot.top);
    expect(a.height).toBeCloseTo(b.height / 2);
    expect(a.x).toBeLessThan(b.x);
  });

  it("lifts the pen at a gap, so missing data is never drawn as zero", () => {
    const d = linePath(plot, [1, null, 2, 2], 2);
    expect(d.match(/M/g)?.length).toBe(2);
    expect(d.match(/L/g)?.length).toBe(1);
    expect(linePoints(plot, [1, null, 2], 2).map((p) => p.index)).toEqual([0, 2]);
  });

  it("places a median between categories", () => {
    const x6 = valueX(plot, [5, 6, 7], 6)!;
    const x65 = valueX(plot, [5, 6, 7], 6.5)!;
    expect(x65).toBeGreaterThan(x6);
    expect(valueX(plot, [], 3)).toBeNull();
  });

  it("orders a sparkline oldest first, height carrying the result", () => {
    expect(sparkCells(["win", "loss", "draw"]).map((c) => [c.result, c.height])).toEqual([
      ["draw", 0.5],
      ["loss", 0.25],
      ["win", 1],
    ]);
    expect(sparkCells(["odd"])[0].result).toBe("none");
  });

  it("names evenness in words", () => {
    expect([null, 0.1, 0.3, 0.6, 0.9].map(evennessWord)).toEqual([
      "-",
      "all at one",
      "focused",
      "fairly even",
      "even",
    ]);
  });
});
