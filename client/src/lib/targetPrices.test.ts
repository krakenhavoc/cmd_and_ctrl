import { describe, expect, it } from "vitest";
import {
  distinctPrices,
  manaValueOf,
  priceLabel,
  targetPriceRange,
  targetPriceSummary,
} from "./targetPrices";

// #1296: Dragonfire Blade — "Equip {4}. This ability costs {1} less to
// activate for each color of the creature it targets."
const blade = { vivi: "{2}", golem: "{4}", bird: "{4}", queen: "" };
const names: Record<string, string> = {
  vivi: "Vivi Ornitier",
  golem: "Golem",
  bird: "Ornithopter",
  queen: "Sliver Queen",
};

describe("targetPrices", () => {
  it("reads a cost's mana value for ordering", () => {
    expect(manaValueOf("")).toBe(0);
    expect(manaValueOf("{4}")).toBe(4);
    expect(manaValueOf("{2}{U}{R}")).toBe(4);
    expect(manaValueOf("{X}{G/P}")).toBe(1);
  });

  it('names an emptied cost "free"', () => {
    expect(priceLabel("")).toBe("free");
    expect(priceLabel("{2}")).toBe("{2}");
  });

  it("lists distinct prices cheapest first", () => {
    expect(distinctPrices(blade)).toEqual(["", "{2}", "{4}"]);
    expect(distinctPrices(undefined)).toEqual([]);
  });

  it("gives the menu a range only when the targets differ", () => {
    expect(targetPriceRange(blade)).toBe("free–{4} depending on the target");
    expect(targetPriceRange({ a: "{3}", b: "{3}" })).toBe("");
    expect(targetPriceRange(undefined)).toBe("");
  });

  it("gives the banner each price with the targets that pay it", () => {
    expect(targetPriceSummary(blade, (id) => names[id])).toBe(
      "free: Sliver Queen · {2}: Vivi Ornitier · {4}: Golem, Ornithopter",
    );
  });

  it("leaves out a target it cannot name rather than printing its ID", () => {
    expect(targetPriceSummary({ vivi: "{2}", hidden: "{4}" }, (id) => names[id])).toBe(
      "{2}: Vivi Ornitier",
    );
  });

  it("says nothing when every target costs the same", () => {
    expect(targetPriceSummary({ golem: "{4}", bird: "{4}" }, (id) => names[id])).toBe("");
  });
});
