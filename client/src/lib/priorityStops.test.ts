import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";

import {
  canManuallyStop,
  consumeManualStop,
  hasManualStop,
  manualStops,
  toggleManualStop,
  _resetForTests,
} from "./priorityStops";

describe("priorityStops", () => {
  beforeEach(() => _resetForTests());

  it("starts with an empty set", () => {
    expect(get(manualStops).size).toBe(0);
  });

  it("canManuallyStop allows priority-granting steps", () => {
    expect(canManuallyStop("upkeep")).toBe(true);
    expect(canManuallyStop("declare_attackers")).toBe(true);
    expect(canManuallyStop("end")).toBe(true);
  });

  it("canManuallyStop rejects no-priority steps", () => {
    expect(canManuallyStop("untap")).toBe(false);
    expect(canManuallyStop("cleanup")).toBe(false);
  });

  it("toggleManualStop adds and removes a priority-granting step", () => {
    expect(toggleManualStop("upkeep")).toBe(true);
    expect(hasManualStop("upkeep")).toBe(true);
    expect(toggleManualStop("upkeep")).toBe(false);
    expect(hasManualStop("upkeep")).toBe(false);
  });

  it("toggleManualStop is a no-op on no-priority steps", () => {
    expect(toggleManualStop("untap")).toBe(false);
    expect(hasManualStop("untap")).toBe(false);
    expect(get(manualStops).size).toBe(0);
  });

  it("supports multiple simultaneous pins", () => {
    toggleManualStop("upkeep");
    toggleManualStop("declare_attackers");
    toggleManualStop("end");
    expect(get(manualStops).size).toBe(3);
    expect(hasManualStop("upkeep")).toBe(true);
    expect(hasManualStop("declare_attackers")).toBe(true);
    expect(hasManualStop("end")).toBe(true);
  });

  it("consumeManualStop clears the pin", () => {
    toggleManualStop("upkeep");
    consumeManualStop("upkeep");
    expect(hasManualStop("upkeep")).toBe(false);
  });

  it("consumeManualStop is a no-op on unpinned steps", () => {
    toggleManualStop("upkeep");
    consumeManualStop("declare_attackers");
    expect(hasManualStop("upkeep")).toBe(true);
    expect(get(manualStops).size).toBe(1);
  });

  it("hasManualStop handles null / undefined gracefully", () => {
    expect(hasManualStop(null)).toBe(false);
    expect(hasManualStop(undefined)).toBe(false);
  });

  it("subscribers observe updates", () => {
    const observed: number[] = [];
    const unsub = manualStops.subscribe((s) => observed.push(s.size));
    toggleManualStop("upkeep");
    toggleManualStop("end");
    consumeManualStop("upkeep");
    unsub();
    expect(observed).toEqual([0, 1, 2, 1]);
  });
});
