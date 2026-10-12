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
});
