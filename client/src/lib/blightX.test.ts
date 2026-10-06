import { describe, expect, it } from "vitest";
import { castBlightOffer, hasXCost } from "./targeting";
import type { CardView } from "./protocol";

// #2174: "blight X" (CR 701.68a) — the X prompt opens, and the cast
// chain asks which one creature takes the counters.

function soul(options: string[]): CardView {
  return {
    instance_id: "spell",
    name: "Soul Immolation",
    owner: "me",
    controller: "me",
    mana_cost: "{3}{R}{R}",
    additional_cost: {
      label: "Blight X",
      demands_x: true,
      blight_x: true,
      blight_x_max: options.length > 0 ? 3 : undefined,
      blight_options: { cards: options, min: 1, max: 1 },
    },
  };
}

describe("blight X", () => {
  it("opens the X prompt even though the printed cost has no {X}", () => {
    expect(hasXCost(soul(["bear"]))).toBe(true);
  });

  it("asks for one creature, with X named rather than a number", () => {
    const offer = castBlightOffer(soul(["bear", "ogre"]), {});
    expect(offer?.n).toBe(0);
    expect(offer?.options).toEqual(["bear", "ogre"]);
  });

  it("asks nothing when there is no creature (X can only be 0)", () => {
    expect(castBlightOffer(soul([]), {})).toBeUndefined();
  });
});
