import { describe, it, expect } from "vitest";

import { RECONNECT_BASE_MS, RECONNECT_CAP_MS, reconnectDelayMs } from "./ws";

describe("reconnectDelayMs", () => {
  it("spans [base/2, base] on the first attempt", () => {
    expect(reconnectDelayMs(0, () => 0)).toBe(RECONNECT_BASE_MS / 2);
    expect(reconnectDelayMs(0, () => 1)).toBe(RECONNECT_BASE_MS);
  });

  it("doubles the nominal delay per attempt", () => {
    expect(reconnectDelayMs(1, () => 1)).toBe(RECONNECT_BASE_MS * 2);
    expect(reconnectDelayMs(2, () => 1)).toBe(RECONNECT_BASE_MS * 4);
    expect(reconnectDelayMs(3, () => 1)).toBe(RECONNECT_BASE_MS * 8);
  });

  it("caps the nominal delay at RECONNECT_CAP_MS", () => {
    expect(reconnectDelayMs(6, () => 1)).toBe(RECONNECT_CAP_MS);
    expect(reconnectDelayMs(20, () => 1)).toBe(RECONNECT_CAP_MS);
    // 2 ** 1100 overflows to Infinity; Math.min must still cap.
    expect(reconnectDelayMs(1100, () => 1)).toBe(RECONNECT_CAP_MS);
  });

  it("never drops below half the nominal delay", () => {
    expect(reconnectDelayMs(6, () => 0)).toBe(RECONNECT_CAP_MS / 2);
    expect(reconnectDelayMs(20, () => 0)).toBe(RECONNECT_CAP_MS / 2);
  });

  it("jitters with the default RNG inside the [nominal/2, nominal] band", () => {
    const nominal = RECONNECT_BASE_MS * 2 ** 4;
    for (let i = 0; i < 100; i++) {
      const d = reconnectDelayMs(4);
      expect(d).toBeGreaterThanOrEqual(nominal / 2);
      expect(d).toBeLessThanOrEqual(nominal);
    }
  });

  it("returns integer milliseconds", () => {
    expect(Number.isInteger(reconnectDelayMs(3, () => 0.3337))).toBe(true);
  });
});
