import { describe, expect, it } from "vitest";
import { damageCantBePreventedLine } from "./turnRules";

describe("damageCantBePreventedLine", () => {
  it("is empty when no grant is live", () => {
    expect(damageCantBePreventedLine(undefined)).toBe("");
    expect(damageCantBePreventedLine([])).toBe("");
  });

  it("names the sources once each, in order", () => {
    expect(damageCantBePreventedLine(["Skullcrack"])).toBe(
      "Damage can't be prevented this turn — Skullcrack",
    );
    expect(damageCantBePreventedLine(["Skullcrack", "Stomp", "Skullcrack"])).toBe(
      "Damage can't be prevented this turn — Skullcrack, Stomp",
    );
  });

  it("still states the rule when the sources are unnamed", () => {
    expect(damageCantBePreventedLine([""])).toBe("Damage can't be prevented this turn");
  });
});
