import { describe, it, expect } from "vitest";

import { castableAlternativeCostsOf, castsFaceName } from "./targeting";
import type { AlternativeCostView, CardView } from "./protocol";

// moreThanMeetsTheEye.test.ts — ADR 0152 §5 (#3196). The offer carries
// `casts_face`, and the alternative-cost picker names the face it casts
// ("casts Goldbug, Scrappy Scout"), as it does for disturb.

const mtmte: AlternativeCostView = {
  key: "more_than_meets_the_eye",
  label: "More Than Meets the Eye {W}{U}",
  mana_cost: "{W}{U}",
  casts_face: 1,
};

function goldbug(offers: AlternativeCostView[]): CardView {
  return {
    instance_id: "goldbug",
    name: "Goldbug, Humanity's Ally",
    owner: "me",
    controller: "me",
    type_line: "Legendary Artifact Creature — Robot",
    mana_cost: "{1}{W}{U}",
    layout: "transform",
    castable_here: true,
    alternative_costs: offers,
    faces: [
      { name: "Goldbug, Humanity's Ally", type_line: "Legendary Artifact Creature — Robot" },
      { name: "Goldbug, Scrappy Scout", type_line: "Legendary Artifact — Vehicle" },
    ],
  };
}

describe("an offer that casts another face", () => {
  it("names the face it casts", () => {
    const card = goldbug([mtmte]);
    const [offer] = castableAlternativeCostsOf(card);
    expect(castsFaceName(card, offer)).toBe("Goldbug, Scrappy Scout");
  });

  it("names the back face for disturb too", () => {
    const disturb: AlternativeCostView = { key: "disturb", label: "Disturb {3}{W}", casts_face: 1 };
    const card = goldbug([disturb]);
    expect(castsFaceName(card, disturb)).toBe("Goldbug, Scrappy Scout");
  });

  it("says nothing for an offer that casts the face as it sits", () => {
    const flashback: AlternativeCostView = { key: "flashback", label: "Flashback {2}{R}" };
    expect(castsFaceName(goldbug([flashback]), flashback)).toBeUndefined();
  });

  it("says nothing when the card carries no such face", () => {
    const card = goldbug([mtmte]);
    card.faces = undefined;
    expect(castsFaceName(card, mtmte)).toBeUndefined();
  });
});
