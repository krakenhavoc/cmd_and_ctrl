// sharedTarget.test.ts — #3265: one target for every instance of a
// repeated trigger.

import { describe, it, expect } from "vitest";

import {
  repeatsOf,
  sharedFrom,
  sharedPick,
  sharedStillWanted,
  type TargetingState,
} from "./targeting";
import { sameTargetLabel, targetingRequest } from "./targetingDock";
import type { CardView, GameView, PendingChoiceView } from "./protocol";

const source = { instance_id: "pyre", name: "Impact Tremors" } as CardView;
const REASON = "Choose target opponent";

function prompt(id: string, over: Partial<PendingChoiceView> = {}): PendingChoiceView {
  return {
    id,
    kind: "pick_target",
    chooser: "me",
    source: "pyre",
    reason: REASON,
    ...over,
  } as PendingChoiceView;
}

const viewWith = (choices: PendingChoiceView[]): GameView =>
  ({ pending_choices: choices }) as unknown as GameView;

function state(over: Partial<TargetingState> = {}): TargetingState {
  return {
    card: source,
    mode: "any",
    legal: { players: new Set(["ann", "bob"]), cards: new Set() },
    label: REASON,
    min: 1,
    max: 1,
    picked: [],
    steps: [],
    step: 0,
    done: [],
    choiceID: "t1",
    choiceKind: "pick_target",
    ...over,
  };
}

describe("repeatsOf", () => {
  it("counts the other prompts of the same ability for the same chooser", () => {
    const v = viewWith([prompt("t1"), prompt("t2"), prompt("t3")]);
    expect(repeatsOf(state(), v)).toBe(2);
  });

  it("leaves out another source, another clause and another chooser", () => {
    const v = viewWith([
      prompt("t1"),
      prompt("x1", { source: "other-copy" }),
      prompt("x2", { reason: "Choose target creature" }),
      prompt("x3", { chooser: "them" }),
      prompt("x4", { kind: "retarget" }),
    ]);
    expect(repeatsOf(state(), v)).toBe(0);
  });

  it("is 0 for a cast, a multi-target clause, or no view", () => {
    const v = viewWith([prompt("t1"), prompt("t2")]);
    expect(repeatsOf(state({ choiceID: undefined, choiceKind: undefined }), v)).toBe(0);
    expect(repeatsOf(state({ max: 2 }), v)).toBe(0);
    expect(repeatsOf(state(), null)).toBe(0);
  });
});

describe("sharedPick", () => {
  const shared = sharedFrom(state(), { kind: "player", id: "bob" });

  it("keeps the source, the clause and the pick", () => {
    expect(shared).toEqual({ source: "pyre", reason: REASON, ref: { kind: "player", id: "bob" } });
  });

  it("answers a sibling with the kept target while it is legal", () => {
    expect(sharedPick(state({ choiceID: "t2" }), shared)).toEqual({ kind: "player", id: "bob" });
  });

  it("asks when the kept target is no longer legal for this instance", () => {
    const legal = { players: new Set(["ann"]), cards: new Set<string>() };
    expect(sharedPick(state({ choiceID: "t2", legal }), shared)).toBeNull();
  });

  it("asks for another ability, another source, or with nothing kept", () => {
    expect(sharedPick(state({ label: "Choose target creature" }), shared)).toBeNull();
    expect(sharedPick(state({ card: { ...source, instance_id: "copy" } }), shared)).toBeNull();
    expect(sharedPick(state(), null)).toBeNull();
  });
});

describe("sharedStillWanted", () => {
  const shared = sharedFrom(state(), { kind: "player", id: "bob" });

  it("is true while a sibling prompt is waiting, false once none is", () => {
    expect(sharedStillWanted(shared, viewWith([prompt("t2")]))).toBe(true);
    expect(sharedStillWanted(shared, viewWith([]))).toBe(false);
    expect(sharedStillWanted(null, viewWith([prompt("t2")]))).toBe(false);
  });
});

describe("the dock's Same target toggle", () => {
  const handlers = { onDone: () => {}, onCancel: () => {} };

  it("is offered when the ability has other instances waiting, and counts all of them", () => {
    const req = targetingRequest(state(), null, {
      ...handlers,
      share: { count: 2, on: true, onToggle: () => {} },
    });
    expect(req.row?.map((a) => [a.label, a.pressed])).toEqual([[sameTargetLabel(2), true]]);
    expect(sameTargetLabel(2)).toBe("Same target for all 3");
  });

  it("is not offered for a lone instance", () => {
    const req = targetingRequest(state(), null, {
      ...handlers,
      share: { count: 0, on: true, onToggle: () => {} },
    });
    expect(req.row).toBeUndefined();
  });
});
