import { describe, expect, it } from "vitest";
import { defaultBotTier, tierLabel, tierNote, unavailableNotes, type TierChoice } from "./botTier";

const t = (tier: string, available = true): TierChoice => ({
  tier,
  label: tier[0].toUpperCase() + tier.slice(1),
  description: `${tier} description`,
  available,
});

describe("defaultBotTier", () => {
  it("prefers heuristic over the server's first tier", () => {
    expect(defaultBotTier([t("random"), t("heuristic"), t("assisted", false)])).toBe("heuristic");
  });
  it("falls back to the first available tier", () => {
    expect(defaultBotTier([t("random"), t("heuristic", false), t("assisted")])).toBe("random");
    expect(defaultBotTier([t("heuristic", false), t("assisted")])).toBe("assisted");
  });
  it("is empty with nothing available", () => {
    expect(defaultBotTier([])).toBe("");
    expect(defaultBotTier([t("random", false)])).toBe("");
  });
});

describe("tierLabel", () => {
  it("marks random as testing only", () => {
    expect(tierLabel(t("random"))).toBe("Random (testing only)");
    expect(tierLabel(t("heuristic"))).toBe("Heuristic");
    expect(tierLabel(t("strong", false))).toBe("Strong");
  });
});

describe("unavailableNotes", () => {
  it("says why, once per reason", () => {
    const keyless = "this server has no model API key";
    expect(
      unavailableNotes([
        t("random"),
        t("heuristic"),
        { ...t("assisted", false), reason: keyless },
        { ...t("strong", false), reason: keyless },
        t("odd", false),
      ]),
    ).toEqual([`Assisted and Strong: ${keyless}`, "Odd: not available on this server"]);
  });
  it("is empty when every tier is offered", () => {
    expect(unavailableNotes([t("random"), t("heuristic")])).toEqual([]);
  });
});

describe("tierNote", () => {
  const tiers = [t("random"), t("heuristic")];
  it("warns for random", () => {
    const n = tierNote(tiers, "random");
    expect(n?.warning).toBe(true);
    expect(n?.text).toContain("not an opponent");
  });
  it("shows the plain description otherwise", () => {
    expect(tierNote(tiers, "heuristic")).toEqual({
      text: "heuristic description",
      warning: false,
    });
  });
  it("is null for an unknown tier", () => {
    expect(tierNote(tiers, "")).toBeNull();
  });
});
