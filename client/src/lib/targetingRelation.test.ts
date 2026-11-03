import { describe, expect, it } from "vitest";
import { advance, openWalk, stepsFor } from "./targeting";
import type { CardView } from "./protocol";

const card: CardView = {
  instance_id: "welder",
  name: "Goblin Welder",
  owner: "p0",
  controller: "p0",
};
const clauses = [
  { min: 1, max: 1, cards: ["artifact-a", "artifact-b"] },
  {
    min: 1,
    max: 1,
    cards: ["grave-a", "grave-b"],
    relation: { slot: 0, pairs: { "artifact-a": ["grave-a"], "artifact-b": ["grave-b"] } },
  },
];

describe("target clauses related to earlier picks", () => {
  it("offers only the graveyard paired with the chosen artifact", () => {
    const state = openWalk(card, stepsFor("permanent", undefined, clauses, 0), {});
    state.picked = [{ kind: "card", id: "artifact-b" }];
    const next = advance(state);
    expect([...next!.legal!.cards]).toEqual(["grave-b"]);
  });

  it("does not use an anchor chosen for a different mode occurrence", () => {
    const steps = stepsFor("permanent", undefined, clauses, 1);
    const state = openWalk(card, steps, {});
    state.picked = [];
    state.done = [{ kind: "card", id: "artifact-a", mode: 0, slot: 0 }];
    expect([...advance(state)!.legal!.cards]).toEqual([]);
  });

  it("skips an optional dependent clause with no compatible candidates", () => {
    const steps = stepsFor(
      "permanent",
      undefined,
      [clauses[0], { ...clauses[1], min: 0, relation: { slot: 0, pairs: {} } }],
      0,
    );
    const state = openWalk(card, steps, {});
    state.picked = [{ kind: "card", id: "artifact-a" }];
    expect(advance(state)).toBeNull();
  });

  // #4862: Quandrix Command's "target player shuffles up to three target
  // cards from their graveyard": the anchor is a player, and the
  // dependent clause takes up to three of that player's cards.
  it("offers up to three cards from the chosen player's graveyard", () => {
    const steps = stepsFor(
      "player",
      undefined,
      [
        { min: 1, max: 1, players: ["p0", "p1"] },
        {
          min: 0,
          max: 3,
          cards: ["p0-a", "p1-a", "p1-b", "p1-c", "p1-d"],
          relation: { slot: 0, pairs: { p0: ["p0-a"], p1: ["p1-a", "p1-b", "p1-c", "p1-d"] } },
        },
      ],
      0,
    );
    const state = openWalk(card, steps, {});
    state.picked = [{ kind: "player", id: "p1" }];
    const next = advance(state)!;
    expect([...next.legal!.cards]).toEqual(["p1-a", "p1-b", "p1-c", "p1-d"]);
    expect(next.min).toBe(0);
    expect(next.max).toBe(3);
  });
});
