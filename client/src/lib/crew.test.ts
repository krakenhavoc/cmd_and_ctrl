import { describe, expect, it } from "vitest";
import type { ActivatedAbilityView, CardView } from "./protocol";
import { crewAvailablePower, crewOptionsFor, crewPower, crewSatisfied } from "./crew";
import { abilityBlocked } from "./contextMenu.logic";

// crew.test.ts — S27, CR 702.122a.
//
// Crew is the first cost in the client that is COUNTED rather than
// itemised: "tap any number of creatures with total power N or
// more". Every assertion here is about that difference — a floor
// rather than a ceiling, overshooting legal, picking nothing not.

function creature(id: string, name: string, power: number): CardView {
  return {
    instance_id: id,
    name,
    power,
    toughness: power,
  } as unknown as CardView;
}

const bear = creature("c-bear", "Grizzly Bears", 2);
const cat = creature("c-cat", "Cat", 2);
const wall = creature("c-wall", "Wall", 0);

const crew4: ActivatedAbilityView = {
  index: 0,
  label: "Crew 4",
  crew_cost: 4,
  crew_options: { cards: ["c-bear", "c-cat", "c-wall"] },
};

describe("crewPower", () => {
  it("adds the chosen creatures' power", () => {
    expect(crewPower([bear, cat, wall], ["c-bear", "c-cat"])).toBe(4);
  });

  it("counts an unknown id as nothing rather than NaN", () => {
    // The snapshot can change under an open prompt; a creature that
    // has left the board should shrink the total, not poison it.
    expect(crewPower([bear], ["c-bear", "c-gone"])).toBe(2);
  });

  it("is zero for an empty selection", () => {
    expect(crewPower([bear, cat], [])).toBe(0);
  });
});

describe("crewSatisfied", () => {
  it("accepts an exact total", () => {
    expect(crewSatisfied([bear, cat], ["c-bear", "c-cat"], 4)).toBe(true);
  });

  it("accepts overshooting — the printed number is a floor", () => {
    const dragon = creature("c-dragon", "Dragon", 5);
    expect(crewSatisfied([dragon], ["c-dragon"], 3)).toBe(true);
  });

  it("rejects a total below the crew number", () => {
    expect(crewSatisfied([bear, cat, wall], ["c-bear", "c-wall"], 4)).toBe(false);
  });

  it("rejects an empty selection even when the cost is zero", () => {
    // There is no printed crew 0, and "tap nothing" is not a way to
    // pay a cost that says to tap creatures. Convoke's "tap nothing"
    // escape hatch is the OTHER cost.
    expect(crewSatisfied([bear], [], 0)).toBe(false);
  });
});

describe("crewAvailablePower", () => {
  it("sums the whole offered roster", () => {
    expect(crewAvailablePower([bear, cat, wall])).toBe(4);
  });

  it("is zero with nothing offered", () => {
    expect(crewAvailablePower([])).toBe(0);
  });
});

describe("crewOptionsFor", () => {
  it("narrows the board to the server's offered set", () => {
    const board = [bear, cat, wall, creature("c-other", "Someone else's bear", 9)];
    expect(crewOptionsFor(crew4, board).map((c) => c.instance_id)).toEqual([
      "c-bear",
      "c-cat",
      "c-wall",
    ]);
  });
});

describe("abilityBlocked for a crew ability", () => {
  it("blocks the row when nothing is untapped", () => {
    const empty: ActivatedAbilityView = { ...crew4, crew_options: { cards: [] } };
    expect(abilityBlocked(empty, false, false)).toBe("no untapped creatures to crew with");
  });

  it("leaves the row open when creatures exist, even if they fall short", () => {
    // Whether they ADD UP is the prompt's job, with the running
    // total in front of the player — two answers on screen would be
    // worse than one.
    const short: ActivatedAbilityView = { ...crew4, crew_options: { cards: ["c-wall"] } };
    expect(abilityBlocked(short, false, false)).toBe("");
  });

  it("does not block a crew ability on a tapped or summoning-sick source", () => {
    // Crewing does not tap the Vehicle — the CREATURES tap
    // (CR 702.122b) — so a tapped Vehicle is still crewable, and a
    // Vehicle that entered this turn is still crewable even though
    // CR 302.6 will stop the resulting creature from attacking.
    // Crew carries no tap_cost, which is what makes both fall out.
    expect(abilityBlocked(crew4, true, true)).toBe("");
  });
});
