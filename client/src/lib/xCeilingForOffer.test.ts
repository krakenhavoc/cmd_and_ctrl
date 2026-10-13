// xCeilingForOffer.test.ts — #3225. "You may cast this spell as though it
// had flash if X is 3 or less" (Silver Scrutiny, CR 601.3c) is an offer
// that carries its own X ceiling. The X picker takes the claimed offer's
// `x_max`; the cast for the printed cost is not bound by it.

import { describe, it, expect } from "vitest";

import { xCeilingFor } from "./targeting";
import type { CardView } from "./protocol";

const silverScrutiny = {
  instance_id: "ss-1",
  name: "Silver Scrutiny",
  owner: "me",
  controller: "me",
  type_line: "Sorcery",
  mana_cost: "{X}{U}{U}",
  alternative_costs: [
    { key: "flash-x", label: "Cast with flash (X is 3 or less)", mana_cost: "{X}{U}{U}", x_max: 3 },
  ],
} as CardView;

describe("xCeilingFor", () => {
  it("caps the picker at the claimed offer's ceiling", () => {
    expect(xCeilingFor(silverScrutiny, "flash-x")).toBe(3);
  });

  it("leaves the cast for the printed cost uncapped", () => {
    expect(xCeilingFor(silverScrutiny, undefined)).toBeUndefined();
  });

  it("falls back to the card's own ceiling for an offer without one", () => {
    const card = {
      ...silverScrutiny,
      x_max: 5,
      alternative_costs: [{ key: "other", mana_cost: "{X}{U}" }],
    } as CardView;
    expect(xCeilingFor(card, "other")).toBe(5);
    expect(xCeilingFor(card, undefined)).toBe(5);
  });

  it("keeps a ceiling of 0 (X = 0 is the only announcement)", () => {
    const card = {
      ...silverScrutiny,
      alternative_costs: [{ key: "flash-x", mana_cost: "{X}{U}{U}", x_max: 0 }],
    } as CardView;
    expect(xCeilingFor(card, "flash-x")).toBe(0);
  });
});
