import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";

import {
  BLUFF_MAX_MS,
  BLUFF_MIN_MS,
  _resetForTests,
  bluffArmed,
  bluffDelayMs,
  bluffStatus,
  bluffStatusText,
  initBluffArmed,
  setBluffStatus,
  toggleBluffArmed,
} from "./bluff";

describe("bluffDelayMs", () => {
  it("spans the range with the injected random", () => {
    expect(bluffDelayMs(1500, 4000, () => 0)).toBe(1500);
    expect(bluffDelayMs(1500, 4000, () => 1)).toBe(4000);
    expect(bluffDelayMs(1500, 4000, () => 0.5)).toBe(2750);
  });

  it("clamps both bounds to 500–15000", () => {
    expect(bluffDelayMs(0, 0, () => 0.5)).toBe(BLUFF_MIN_MS);
    expect(bluffDelayMs(60_000, 90_000, () => 0.5)).toBe(BLUFF_MAX_MS);
    expect(bluffDelayMs(-5, 100_000, () => 0)).toBe(BLUFF_MIN_MS);
    expect(bluffDelayMs(-5, 100_000, () => 1)).toBe(BLUFF_MAX_MS);
  });

  it("swaps bounds given the wrong way round", () => {
    expect(bluffDelayMs(4000, 1500, () => 0)).toBe(1500);
    expect(bluffDelayMs(4000, 1500, () => 1)).toBe(4000);
  });

  it("treats a non-number bound as the floor", () => {
    expect(bluffDelayMs(Number.NaN, 2000, () => 0)).toBe(BLUFF_MIN_MS);
  });

  it("uses Math.random by default and stays in range", () => {
    for (let i = 0; i < 50; i++) {
      const d = bluffDelayMs(1500, 4000);
      expect(d).toBeGreaterThanOrEqual(1500);
      expect(d).toBeLessThanOrEqual(4000);
    }
  });
});

describe("bluffArmed", () => {
  beforeEach(() => _resetForTests());

  it("defaults from the settings at game mount", () => {
    initBluffArmed({ bluffCounterspell: false, bluffInstant: false });
    expect(get(bluffArmed)).toBe(false);
    initBluffArmed({ bluffCounterspell: true, bluffInstant: false });
    expect(get(bluffArmed)).toBe(true);
    initBluffArmed({ bluffCounterspell: false, bluffInstant: true });
    expect(get(bluffArmed)).toBe(true);
  });

  it("toggles and reports the new state", () => {
    expect(toggleBluffArmed()).toBe(true);
    expect(get(bluffArmed)).toBe(true);
    expect(toggleBluffArmed()).toBe(false);
  });
});

describe("bluffStatusText", () => {
  beforeEach(() => _resetForTests());

  it("is empty with no bluff running", () => {
    expect(bluffStatusText(null, 0)).toBe("");
  });

  it("counts a timed bluff down in whole seconds", () => {
    expect(bluffStatusText({ manual: false, passesAt: 3_000 }, 0)).toBe("bluffing — passes in 3s");
    expect(bluffStatusText({ manual: false, passesAt: 3_000 }, 2_100)).toBe(
      "bluffing — passes in 1s",
    );
    expect(bluffStatusText({ manual: false, passesAt: 3_000 }, 5_000)).toBe(
      "bluffing — passes in 0s",
    );
  });

  it("tells a manual bluffer to click next", () => {
    expect(bluffStatusText({ manual: true }, 0)).toBe("bluffing — click next");
  });

  it("the status store round-trips", () => {
    setBluffStatus({ manual: true });
    expect(get(bluffStatus)).toEqual({ manual: true });
    setBluffStatus(null);
    expect(get(bluffStatus)).toBeNull();
  });
});
