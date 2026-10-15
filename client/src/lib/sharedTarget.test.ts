// sharedTarget.test.ts — #3265: one target for every instance of a
// repeated trigger; #3817: or, as a second choice, for every copy of the
// card's (ADR 0127, 2026-10-15 amendment: the owner chose to offer both).

import { describe, it, expect, vi } from "vitest";
import { get } from "svelte/store";

import {
  beginChoice,
  cardRepeats,
  effectiveShareScope,
  repeatsOf,
  sharedAfterAnswer,
  sharedFrom,
  sharedKeyOf,
  sharedPick,
  sharedStillWanted,
  targeting,
  type TargetingState,
} from "./targeting";
import { sameTargetCardLabel, sameTargetSourceLabel, targetingRequest } from "./targetingDock";
import type { CardView, GameView, PendingChoiceView } from "./protocol";

const source = { instance_id: "pyre", name: "Impact Tremors" } as CardView;
const REASON = "Choose target opponent";
const KEY = "impact-oracle|triggered|0|Impact Tremors — 1 damage|step0";

function prompt(id: string, over: Partial<PendingChoiceView> = {}): PendingChoiceView {
  return {
    id,
    kind: "pick_target",
    chooser: "me",
    source: "pyre",
    reason: REASON,
    shared_target_key: KEY,
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
    sharedKey: sharedKeyOf(prompt("t1")),
    ...over,
  };
}

// A second copy of the card: another source, the same key.
const copy = (id: string, over: Partial<PendingChoiceView> = {}) =>
  prompt(id, { source: "pyre-2", ...over });

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

describe("cardRepeats (#3817)", () => {
  it("reaches every copy of the card: the other prompts, and the sources in all", () => {
    const v = viewWith([prompt("t1"), prompt("t2"), copy("c1"), copy("c2"), copy("c3")]);
    expect(cardRepeats(state(), v)).toEqual({ count: 4, sources: 2 });
  });

  it("never reaches another ability, another clause, another player or a keyless prompt", () => {
    const v = viewWith([
      prompt("t1"),
      // Falkenrath Noble: the same clause wording, another card.
      copy("x1", { shared_target_key: "noble|triggered|0|Falkenrath Noble — drain 1|step0" }),
      copy("x2", { reason: "Choose target creature" }),
      // A copy the viewer answers for a player they control (ADR 0166).
      copy("x3", { on_behalf_of: "them" }),
      copy("x4", { chooser: "them" }),
      copy("x5", { shared_target_key: undefined }),
    ]);
    expect(cardRepeats(state(), v)).toEqual({ count: 0, sources: 1 });
  });

  it("is nothing for a prompt with no card identity", () => {
    const v = viewWith([prompt("t1"), copy("c1")]);
    expect(cardRepeats(state({ sharedKey: undefined }), v)).toEqual({ count: 0, sources: 1 });
  });
});

describe("beginChoice keeps the prompt's card identity, qualified by its player", () => {
  it("sets sharedKey, and leaves it unset without a key", () => {
    beginChoice(prompt("t1"), source);
    expect(get(targeting)?.sharedKey).toBe(`me\u0000${KEY}`);
    beginChoice(prompt("t2", { on_behalf_of: "them" }), source);
    expect(get(targeting)?.sharedKey).toBe(`them\u0000${KEY}`);
    beginChoice(prompt("t3", { shared_target_key: undefined }), source);
    expect(get(targeting)?.sharedKey).toBeUndefined();
    targeting.set(null);
  });
});

describe("sharedPick", () => {
  const shared = sharedFrom(state(), { kind: "player", id: "bob" });

  it("keeps the scope, the source, the clause and the pick", () => {
    expect(shared).toEqual({
      scope: "source",
      source: "pyre",
      reason: REASON,
      ref: { kind: "player", id: "bob" },
    });
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

  it("#3817: the card scope answers another copy, and nothing of another card or player", () => {
    const card = sharedFrom(state(), { kind: "player", id: "bob" }, "card");
    const other = { ...source, instance_id: "pyre-2" };
    expect(sharedPick(state({ choiceID: "c1", card: other }), card)).toEqual({
      kind: "player",
      id: "bob",
    });
    const noble = sharedKeyOf(copy("n", { shared_target_key: "noble|triggered|0|x|step0" }));
    expect(sharedPick(state({ card: other, sharedKey: noble }), card)).toBeNull();
    const theirs = sharedKeyOf(copy("t", { on_behalf_of: "them" }));
    expect(sharedPick(state({ card: other, sharedKey: theirs }), card)).toBeNull();
    expect(sharedPick(state({ card: other, sharedKey: undefined }), card)).toBeNull();
  });

  it("keeps nothing for the card scope without a card identity", () => {
    expect(sharedFrom(state({ sharedKey: undefined }), { kind: "player", id: "bob" }, "card")).toBe(
      null,
    );
  });
});

describe("sharedAfterAnswer (#3817)", () => {
  const bob = [{ kind: "player" as const, id: "bob" }];
  const both = viewWith([prompt("t1"), prompt("t2"), copy("c1")]);

  it("keeps the card scope where it reaches another source", () => {
    expect(sharedAfterAnswer(state(), bob, "card", both)?.scope).toBe("card");
    expect(sharedAfterAnswer(state(), bob, "source", both)?.scope).toBe("source");
  });

  it("falls back to the source scope with one source, and keeps nothing when off or alone", () => {
    const one = viewWith([prompt("t1"), prompt("t2")]);
    expect(sharedAfterAnswer(state(), bob, "card", one)?.scope).toBe("source");
    expect(sharedAfterAnswer(state(), bob, "off", both)).toBeNull();
    expect(sharedAfterAnswer(state(), bob, "source", viewWith([prompt("t1"), copy("c1")]))).toBe(
      null,
    );
    expect(sharedAfterAnswer(state(), [...bob, ...bob], "card", both)).toBeNull();
  });

  it("effectiveShareScope never turns sharing off by asking for more", () => {
    expect(effectiveShareScope("card", false)).toBe("source");
    expect(effectiveShareScope("card", true)).toBe("card");
    expect(effectiveShareScope("off", true)).toBe("off");
  });
});

describe("sharedStillWanted", () => {
  const shared = sharedFrom(state(), { kind: "player", id: "bob" });

  it("is true while a sibling prompt is waiting, false once none is", () => {
    expect(sharedStillWanted(shared, viewWith([prompt("t2")]))).toBe(true);
    expect(sharedStillWanted(shared, viewWith([]))).toBe(false);
    expect(sharedStillWanted(null, viewWith([prompt("t2")]))).toBe(false);
  });

  it("#3817: the card scope is wanted while any copy's prompt waits", () => {
    const card = sharedFrom(state(), { kind: "player", id: "bob" }, "card");
    expect(sharedStillWanted(card, viewWith([copy("c1")]))).toBe(true);
    expect(sharedStillWanted(shared, viewWith([copy("c1")]))).toBe(false);
  });
});

describe("the dock's Same target choices", () => {
  const handlers = { onDone: () => {}, onCancel: () => {} };
  const rowOf = (share: Parameters<typeof targetingRequest>[2]["share"]) =>
    targetingRequest(state(), null, { ...handlers, share }).row;

  it("offers this source's repeats alone when only this source has them", () => {
    const row = rowOf({
      scope: "source",
      count: 2,
      card: { count: 2, sources: 1 },
      onScope: vi.fn(),
    });
    expect(row?.map((a) => [a.label, a.pressed])).toEqual([
      ["Same target for this Impact Tremors's 3 triggers", true],
    ]);
    expect(sameTargetSourceLabel("Blood Artist", 8)).toBe(
      "Same target for this Blood Artist's 8 triggers",
    );
  });

  it("#3817: offers every copy's as a second choice when more than one source waits", () => {
    const onScope = vi.fn();
    const row = rowOf({ scope: "source", count: 7, card: { count: 39, sources: 20 }, onScope });
    expect(row?.map((a) => [a.label, a.pressed])).toEqual([
      ["Same target for this Impact Tremors's 8 triggers", true],
      ["Same target for all 40 Impact Tremors triggers", false],
    ]);
    expect(sameTargetCardLabel("Blood Artist", 40)).toBe(
      "Same target for all 40 Blood Artist triggers",
    );
    row![1].onPress();
    expect(onScope).toHaveBeenLastCalledWith("card");
    row![0].onPress();
    expect(onScope).toHaveBeenLastCalledWith("off");
  });

  it("presses the card choice when chosen, and a second press turns sharing off", () => {
    const onScope = vi.fn();
    const row = rowOf({ scope: "card", count: 1, card: { count: 5, sources: 3 }, onScope });
    expect(row?.map((a) => a.pressed)).toEqual([false, true]);
    row![1].onPress();
    expect(onScope).toHaveBeenLastCalledWith("off");
    row![0].onPress();
    expect(onScope).toHaveBeenLastCalledWith("source");
  });

  it("offers only the card choice when this source has no repeats of its own", () => {
    const row = rowOf({
      scope: "card",
      count: 0,
      card: { count: 3, sources: 4 },
      onScope: vi.fn(),
    });
    expect(row?.map((a) => [a.label, a.pressed])).toEqual([
      ["Same target for all 4 Impact Tremors triggers", true],
    ]);
  });

  it("is not offered for a lone instance", () => {
    expect(
      rowOf({ scope: "source", count: 0, card: { count: 0, sources: 1 }, onScope: vi.fn() }),
    ).toBe(undefined);
  });
});
