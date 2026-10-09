import { describe, expect, it } from "vitest";
import { defaultBotTier, tierNote, tierOptionLabel, type TierChoice } from "./botTier";

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

describe("tierOptionLabel", () => {
  it("marks random as testing only", () => {
    expect(tierOptionLabel(t("random"))).toBe("Random (testing only)");
    expect(tierOptionLabel(t("heuristic"))).toBe("Heuristic");
    expect(tierOptionLabel(t("strong", false))).toBe("Strong — not built yet");
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
