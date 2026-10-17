import { describe, expect, it } from "vitest";
import { BACK_ZONES, ZONE_LABEL, ZONE_WEIGHT, backRowShares, isSpreadZone } from "./backRow";

// #3028 / #4087: how the back row's lands, artifacts, enchantments,
// planeswalkers and battles share its width.
describe("backRowShares", () => {
  it("is the old two-zone split when only artifacts sit beside the lands", () => {
    const s = backRowShares(6, { artifact: 2 });
    expect(s.artifact).toBeCloseTo(2 / 8, 9);
    expect(s.planeswalker).toBe(1);
    expect(s.enchantment).toBe(1);
  });

  it("lets a lone zone take the whole row", () => {
    expect(backRowShares(0, { artifact: 5 }).artifact).toBe(1);
    expect(backRowShares(4, {}).artifact).toBe(1);
    expect(backRowShares(0, { planeswalker: 3 }).planeswalker).toBe(1);
  });

  it("weighs a planeswalker pile up against an artifact pile", () => {
    const s = backRowShares(4, { artifact: 4, planeswalker: 2 });
    const total = 4 + 4 + 2 * ZONE_WEIGHT.planeswalker;
    expect(s.artifact).toBeCloseTo(4 / total, 9);
    expect(s.planeswalker).toBeCloseTo((2 * ZONE_WEIGHT.planeswalker) / total, 9);
    expect(s.planeswalker).toBeGreaterThan(2 / (4 + 4 + 2));
  });

  it("splits the row between the zones when there are no lands", () => {
    const s = backRowShares(0, { artifact: 3, enchantment: 3, battle: 1 });
    const sum = s.artifact + s.enchantment + s.battle;
    expect(sum).toBeCloseTo(1, 9);
  });

  it("never gives the zones more than the row between them", () => {
    const s = backRowShares(5, { planeswalker: 3, battle: 1, artifact: 6, enchantment: 4 });
    // "other" has no piles and is not drawn; the four that are share less than the row.
    const sum = s.planeswalker + s.battle + s.artifact + s.enchantment;
    expect(sum).toBeCloseTo(1 - 5 / (5 + 6 + 2 + 6 + 4), 9);
  });
});

describe("back row zones", () => {
  it("labels every zone and spreads only planeswalkers and battles", () => {
    expect(BACK_ZONES.map((z) => ZONE_LABEL[z])).toEqual([
      "planeswalkers",
      "battles",
      "artifacts",
      "enchantments",
      "other",
    ]);
    expect(BACK_ZONES.filter(isSpreadZone)).toEqual(["planeswalker", "battle"]);
  });
});
