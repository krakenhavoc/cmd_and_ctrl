import { describe, expect, it } from "vitest";

import { FUSE_FLASH_MS, fuseColour, fuseFlashing, fuseFraction, fuseRGB } from "./draftFuse";

// #5848 §7: the fuse burns blue to about 55 % time left, orange by 35 %,
// red by 15 %, and flashes in the last eight seconds.

describe("fuseFraction", () => {
  it("is the part of the total that is left, from 0 to 1", () => {
    expect(fuseFraction(60_000, 60_000)).toBe(1);
    expect(fuseFraction(30_000, 60_000)).toBe(0.5);
    expect(fuseFraction(0, 60_000)).toBe(0);
  });

  it("clamps more time than the total, a negative time and a bad total", () => {
    expect(fuseFraction(90_000, 60_000)).toBe(1);
    expect(fuseFraction(-5, 60_000)).toBe(0);
    expect(fuseFraction(10, 0)).toBe(0);
    expect(fuseFraction(Number.NaN, 60_000)).toBe(0);
  });
});

describe("fuseRGB / fuseColour", () => {
  const BLUE = [61, 139, 255];
  const ORANGE = [255, 140, 26];

  it("is blue with all of the time left, down to 55 %, and still mostly blue at half", () => {
    expect(fuseRGB(1)).toEqual(BLUE);
    expect(fuseRGB(0.55)).toEqual(BLUE);
    // Past 55 % it starts to turn: a quarter of the way to orange at 50 %.
    const [r, g, b] = fuseRGB(0.5);
    expect(r).toBeGreaterThan(BLUE[0]);
    expect(r).toBeLessThan((BLUE[0] + ORANGE[0]) / 2);
    expect(g).toBeCloseTo(BLUE[1], -1);
    expect(b).toBeGreaterThan((BLUE[2] + ORANGE[2]) / 2);
  });

  it("is fully orange by 35 %, and between orange and red at 30 %", () => {
    expect(fuseRGB(0.35)).toEqual(ORANGE);
    const [r, g] = fuseRGB(0.3);
    expect(r).toBe(255);
    // A quarter of the way from orange (35 %) to red (15 %).
    expect(g).toBeLessThan(ORANGE[1]);
    expect(g).toBeGreaterThan(60);
  });

  it("is red by 15 %, and is a deeper red at 10 % and at the end", () => {
    expect(fuseRGB(0.15)).toEqual([255, 60, 40]);
    const [r10, g10] = fuseRGB(0.1);
    expect(r10).toBeLessThan(255);
    expect(r10).toBeGreaterThan(225);
    expect(g10).toBeLessThan(60);
    expect(fuseRGB(0)).toEqual([225, 20, 20]);
  });

  it("blends between the stops, in one direction only", () => {
    // Between blue and orange the red rises and the blue falls.
    const a = fuseRGB(0.5);
    const b = fuseRGB(0.45);
    const c = fuseRGB(0.4);
    expect(a[0]).toBeLessThanOrEqual(b[0]);
    expect(b[0]).toBeLessThan(c[0]);
    expect(b[2]).toBeGreaterThan(c[2]);
  });

  it("says each as a CSS colour, and holds to the ends for a fraction out of range", () => {
    expect(fuseColour(1)).toBe("rgb(61, 139, 255)");
    expect(fuseColour(0)).toBe("rgb(225, 20, 20)");
    expect(fuseColour(2)).toBe("rgb(61, 139, 255)");
    expect(fuseColour(-1)).toBe("rgb(225, 20, 20)");
    expect(fuseColour(Number.NaN)).toBe("rgb(225, 20, 20)");
  });
});

describe("fuseFlashing", () => {
  it("is on in the last eight seconds, and off before them", () => {
    expect(FUSE_FLASH_MS).toBe(8000);
    expect(fuseFlashing(8000)).toBe(true);
    expect(fuseFlashing(1)).toBe(true);
    expect(fuseFlashing(8001)).toBe(false);
    expect(fuseFlashing(60_000)).toBe(false);
  });

  it("is off when the clock is out and when there is no clock", () => {
    expect(fuseFlashing(0)).toBe(false);
    expect(fuseFlashing(null)).toBe(false);
  });
});
