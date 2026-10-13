// lonePick.test.ts — #3264: which prompts answer themselves when one
// target is legal.

import { describe, it, expect } from "vitest";

import { lonePick, type TargetingState } from "./targeting";
import type { CardView } from "./protocol";

const source = { instance_id: "src", name: "Source" } as CardView;

function state(over: Partial<TargetingState>): TargetingState {
  return {
    card: source,
    mode: "any",
    legal: { players: new Set(), cards: new Set(["bear"]) },
    min: 1,
    max: 1,
    picked: [],
    steps: [],
    step: 0,
    done: [],
    ...over,
  };
}

const trigger = { choiceID: "c1", choiceKind: "pick_target" };
const ability = { ability: { index: 0, sacrificeIDs: [], crewIDs: [] } };

describe("lonePick", () => {
  it("answers a trigger's prompt with its one legal card or player", () => {
    expect(lonePick(state(trigger))).toEqual({ kind: "card", id: "bear" });
    const opp = state({ ...trigger, legal: { players: new Set(["opp"]), cards: new Set() } });
    expect(lonePick(opp)).toEqual({ kind: "player", id: "opp" });
  });

  it("answers an activated ability's step", () => {
    expect(lonePick(state(ability))).toEqual({ kind: "card", id: "bear" });
  });

  it("leaves a spell's targets to the click", () => {
    expect(lonePick(state({}))).toBeNull();
  });

  it("asks when more than one target is legal, or none", () => {
    const two = { players: new Set(["opp"]), cards: new Set(["bear"]) };
    expect(lonePick(state({ ...trigger, legal: two }))).toBeNull();
    const none = { players: new Set<string>(), cards: new Set<string>() };
    expect(lonePick(state({ ...trigger, legal: none }))).toBeNull();
    expect(lonePick(state({ ...trigger, legal: undefined }))).toBeNull();
  });

  it("asks for an optional or multi-target step", () => {
    expect(lonePick(state({ ...trigger, min: 0 }))).toBeNull();
    expect(lonePick(state({ ...trigger, max: 2 }))).toBeNull();
    expect(lonePick(state({ ...trigger, max: 0 }))).toBeNull();
  });

  it("asks for a retarget or a card-set pick, which are not a trigger's target", () => {
    expect(lonePick(state({ choiceID: "c1", choiceKind: "retarget" }))).toBeNull();
    expect(lonePick(state({ choiceID: "c1", choiceKind: "choose_cards" }))).toBeNull();
  });

  it("asks when the ability's price reads the target, so the price is seen first", () => {
    const priced = { ability: { ...ability.ability, prices: { bear: "{1}" } } };
    expect(lonePick(state(priced))).toBeNull();
  });

  it("does not answer a step that already has its pick", () => {
    expect(lonePick(state({ ...trigger, picked: [{ kind: "card", id: "bear" }] }))).toBeNull();
  });
});
