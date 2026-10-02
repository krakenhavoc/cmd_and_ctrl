import { describe, expect, it } from "vitest";
import { damageCantBePreventedLine, exileIfCreaturesDieLine } from "./turnRules";

describe("exileIfCreaturesDieLine", () => {
  it("is empty when no effect is live", () => {
    expect(exileIfCreaturesDieLine(undefined)).toBe("");
    expect(exileIfCreaturesDieLine([])).toBe("");
  });

  it("names the sources once each, in order", () => {
    expect(exileIfCreaturesDieLine(["Flaying Tendrils", "Malicious Eclipse", "Flaying Tendrils"])).toBe(
      "Creatures that would die this turn are exiled instead — Flaying Tendrils, Malicious Eclipse",
    );
  });
});

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
