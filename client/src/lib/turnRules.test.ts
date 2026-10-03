import { describe, expect, it } from "vitest";
import {
  damageCantBePreventedLine,
  damageMultiplierLines,
  damageRedirectionLines,
  damageShieldsLine,
  exileIfCreaturesDieLine,
} from "./turnRules";

describe("damageMultiplierLines", () => {
  it("is empty when no multiplier is live", () => {
    expect(damageMultiplierLines(undefined)).toEqual([]);
    expect(damageMultiplierLines([])).toEqual([]);
  });

  it("keeps every line in order, repeats included, and drops blanks", () => {
    const insult = "Alice's sources deal double damage this turn — Insult";
    expect(damageMultiplierLines([insult, "", insult])).toEqual([insult, insult]);
  });
});

describe("damageRedirectionLines", () => {
  it("is empty when no redirection is live", () => {
    expect(damageRedirectionLines(undefined)).toEqual([]);
    expect(damageRedirectionLines([])).toEqual([]);
  });

  it("keeps every line in order, repeats included, and drops blanks", () => {
    const beacon =
      "Damage to Alice from Goblin Guide is dealt to Beacon of Destiny instead, the next time — Beacon of Destiny";
    const harm =
      "Damage to Bob and their permanents is dealt to Carol instead, the next 2 — Harm's Way";
    expect(damageRedirectionLines([beacon, " ", harm, beacon])).toEqual([beacon, harm, beacon]);
  });
});

describe("exileIfCreaturesDieLine", () => {
  it("is empty when no effect is live", () => {
    expect(exileIfCreaturesDieLine(undefined)).toBe("");
    expect(exileIfCreaturesDieLine([])).toBe("");
  });

  it("names the sources once each, in order", () => {
    expect(
      exileIfCreaturesDieLine(["Flaying Tendrils", "Malicious Eclipse", "Flaying Tendrils"]),
    ).toBe(
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

describe("damageShieldsLine", () => {
  it("is empty when no shield is live", () => {
    expect(damageShieldsLine(undefined)).toBe("");
    expect(damageShieldsLine([])).toBe("");
  });

  it("names each shield and its source once, in order", () => {
    expect(
      damageShieldsLine([
        "Pay No Heed (Goblin Guide)",
        "Healing Grace (Lightning Bolt) — 3 left",
        "Pay No Heed (Goblin Guide)",
      ]),
    ).toBe(
      "Damage prevented this turn — Pay No Heed (Goblin Guide), Healing Grace (Lightning Bolt) — 3 left",
    );
  });
});
