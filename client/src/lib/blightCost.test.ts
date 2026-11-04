import { describe, expect, it } from "vitest";

import { abilityBlocked, blightCostLabel, blightShortfall } from "./contextMenu.logic";

// #4684: "Blight N" as part of an activated ability's cost (Gristle
// Glutton's "{T}, Blight 1:"). The server stamps `blight` (N) and
// `blight_options` (the creatures you control that could take the
// counters) on the row; the client greys the row with none and asks which
// creature with more than one.

describe("a blight cost with no creature to blight", () => {
  it("greys the row (CR 701.68b)", () => {
    expect(blightShortfall(1, { cards: [], min: 1, max: 1 })).toBe("no creature to blight 1");
    expect(abilityBlocked({ blight: 2, blight_options: { min: 1, max: 1 } }, false, false)).toBe(
      "no creature to blight 2",
    );
  });
  it("leaves it live with a creature, and without the component", () => {
    expect(blightShortfall(1, { cards: ["glutton"], min: 1, max: 1 })).toBe("");
    expect(blightShortfall(undefined, undefined)).toBe("");
    expect(blightShortfall(undefined, { cards: [], min: 1, max: 1 })).toBe("");
  });
});

describe("blightCostLabel", () => {
  it("names the counters the pick takes", () => {
    expect(blightCostLabel(1)).toBe("a creature you control (1 -1/-1 counter)");
    expect(blightCostLabel(2)).toBe("a creature you control (2 -1/-1 counters)");
    expect(blightCostLabel(undefined)).toBe("a creature you control (1 -1/-1 counter)");
  });
});
