// @vitest-environment jsdom
//
// targetSet.test.ts — #1559. A clause's rule over the chosen SET of
// its targets (CR 601.2c): "any number of target creature cards that
// each have a different mana value X or less" (Agadeem's Awakening),
// "two target creatures controlled by different players" (Run Away
// Together). The server ships the rule as keys; the picker greys a
// candidate whose key a pick already holds, refuses the click, says
// the rule in the dock's targeting prompt (the banner's line until
// ADR 0111 PR 4), and narrows an X-bounded clause by the X
// collected in the cost prompts.

import { describe, it, expect, afterEach } from "vitest";
import { get } from "svelte/store";

import { targetingDetail } from "./targetingDock";
import {
  begin,
  beginChoice,
  breaksSetRule,
  cancel,
  isLegalCardTarget,
  pickedTotal,
  targeting,
  togglePick,
  type TargetingState,
} from "./targeting";
import type { CardView, PendingChoiceView } from "./protocol";
import { cleanup } from "./test/render.svelte";

afterEach(() => {
  targeting.set(null);
  cleanup();
});

// Agadeem's Awakening as the hand snapshot ships it: every creature
// card, their mana values, and the mana-value keys.
function agadeem(): CardView {
  return {
    instance_id: "agadeem",
    name: "Agadeem's Awakening",
    owner: "me",
    controller: "me",
    mana_cost: "{X}{B}{B}{B}",
    target_mode: "card_in_graveyard",
    legal_targets: {
      cards: ["one", "two-a", "two-b", "three"],
      min: 0,
      max: 0,
      different: {
        label: "each have a different mana value",
        keys: { one: "1", "two-a": "2", "two-b": "2", three: "3" },
      },
      mana_value_at_most_x: true,
      mana_values: { one: 1, "two-a": 2, "two-b": 2, three: 3 },
    },
  };
}

describe("a set rule over the chosen targets (#1559)", () => {
  it("narrows an X-bounded clause to the cards the announced X admits", () => {
    begin(agadeem(), "card_in_graveyard", { xValue: 2 });
    const t = get(targeting)!;
    expect(isLegalCardTarget(t, "one")).toBe(true);
    expect(isLegalCardTarget(t, "two-a")).toBe(true);
    expect(isLegalCardTarget(t, "three")).toBe(false);
    cancel();
  });

  it("greys and refuses a second pick that shares a key, and lets the first be un-picked", () => {
    begin(agadeem(), "card_in_graveyard", { xValue: 3 });
    let t: TargetingState = get(targeting)!;
    t = togglePick(t, { kind: "card", id: "two-a" });
    expect(breaksSetRule(t, "two-b")).toBe(true);
    expect(isLegalCardTarget(t, "two-b")).toBe(false);
    expect(isLegalCardTarget(t, "two-a")).toBe(true);
    expect(isLegalCardTarget(t, "three")).toBe(true);
    const refused = togglePick(t, { kind: "card", id: "two-b" });
    expect(refused.picked.map((p) => p.id)).toEqual(["two-a"]);
    // Un-picking the 2-drop frees the other one.
    t = togglePick(t, { kind: "card", id: "two-a" });
    expect(isLegalCardTarget(t, "two-b")).toBe(true);
    cancel();
  });

  it("a clause with no rule is untouched", () => {
    begin(
      {
        instance_id: "s",
        name: "Spell",
        owner: "me",
        controller: "me",
        legal_targets: { cards: ["a", "b"], min: 2, max: 2 },
      },
      "creature",
    );
    let t: TargetingState = get(targeting)!;
    t = togglePick(t, { kind: "card", id: "a" });
    expect(breaksSetRule(t, "b")).toBe(false);
    expect(togglePick(t, { kind: "card", id: "b" }).picked).toHaveLength(2);
    cancel();
  });

  it("a trigger's pick_target prompt carries the rule too", () => {
    const choice: PendingChoiceView = {
      id: "ch",
      kind: "pick_target",
      chooser: "me",
      from_player: "me",
      count: 0,
      reason: "exile any number of other target creatures controlled by different players",
      pick_target: {
        cards: ["x", "y", "z"],
        min: 0,
        max: 0,
        different: {
          label: "be controlled by different players",
          keys: { x: "p1", y: "p1", z: "p2" },
        },
      },
    } as PendingChoiceView;
    beginChoice(choice, {
      instance_id: "lagrella",
      name: "Lagrella",
      owner: "me",
      controller: "me",
    });
    let t: TargetingState = get(targeting)!;
    t = togglePick(t, { kind: "card", id: "x" });
    expect(isLegalCardTarget(t, "y")).toBe(false);
    expect(isLegalCardTarget(t, "z")).toBe(true);
  });

  it("the dock's prompt says the rule", () => {
    begin(agadeem(), "card_in_graveyard", { xValue: 2 });
    expect(targetingDetail(get(targeting)!)).toContain(
      "targets must each have a different mana value",
    );
    cancel();
  });
});

// #1807: the opposite rule — "up to three target cards from a single
// graveyard" (Decompose). Every pick must share a key, the card's
// owner; the picker greys every candidate outside the first pick's
// group and says the rule.
function decompose(): CardView {
  return {
    instance_id: "decompose",
    name: "Decompose",
    owner: "me",
    controller: "me",
    mana_cost: "{1}{B}",
    target_mode: "card_in_graveyard",
    legal_targets: {
      cards: ["mine-a", "mine-b", "theirs-a", "theirs-b"],
      min: 0,
      max: 3,
      same: {
        label: "come from a single graveyard",
        keys: { "mine-a": "me", "mine-b": "me", "theirs-a": "opp", "theirs-b": "opp" },
      },
    },
  };
}

describe("a sameness rule over the chosen targets (#1807)", () => {
  it("leaves every candidate open until the first pick", () => {
    begin(decompose(), "card_in_graveyard");
    const t = get(targeting)!;
    for (const id of ["mine-a", "mine-b", "theirs-a", "theirs-b"]) {
      expect(isLegalCardTarget(t, id)).toBe(true);
    }
    cancel();
  });

  it("greys and refuses a card from another graveyard once one is picked", () => {
    begin(decompose(), "card_in_graveyard");
    let t: TargetingState = get(targeting)!;
    t = togglePick(t, { kind: "card", id: "theirs-a" });
    expect(breaksSetRule(t, "mine-a")).toBe(true);
    expect(isLegalCardTarget(t, "mine-a")).toBe(false);
    expect(isLegalCardTarget(t, "theirs-b")).toBe(true);
    expect(isLegalCardTarget(t, "theirs-a")).toBe(true);
    const refused = togglePick(t, { kind: "card", id: "mine-a" });
    expect(refused.picked.map((p) => p.id)).toEqual(["theirs-a"]);
    t = togglePick(t, { kind: "card", id: "theirs-b" });
    expect(t.picked).toHaveLength(2);
    // Un-picking both opens the other graveyard again.
    t = togglePick(t, { kind: "card", id: "theirs-a" });
    t = togglePick(t, { kind: "card", id: "theirs-b" });
    expect(isLegalCardTarget(t, "mine-a")).toBe(true);
    cancel();
  });

  it("a trigger's pick_target prompt carries the rule too", () => {
    const choice: PendingChoiceView = {
      id: "ch",
      kind: "pick_target",
      chooser: "me",
      from_player: "me",
      count: 0,
      reason: "exile up to two target cards from a single graveyard",
      pick_target: {
        cards: ["x", "y", "z"],
        min: 0,
        max: 2,
        same: {
          label: "come from a single graveyard",
          keys: { x: "p1", y: "p1", z: "p2" },
        },
      },
    } as PendingChoiceView;
    beginChoice(choice, {
      instance_id: "tracker",
      name: "Griffnaut Tracker",
      owner: "me",
      controller: "me",
    });
    let t: TargetingState = get(targeting)!;
    t = togglePick(t, { kind: "card", id: "z" });
    expect(isLegalCardTarget(t, "x")).toBe(false);
    expect(isLegalCardTarget(t, "y")).toBe(false);
  });

  it("the dock's prompt says the rule", () => {
    begin(decompose(), "card_in_graveyard");
    expect(targetingDetail(get(targeting)!)).toContain("targets must come from a single graveyard");
    cancel();
  });
});

// #4616: the total rule — "up to two target creatures with total mana
// value 6 or less" (The Super Hero Civil War, chapter I). The picker
// greys a candidate whose value would take the picks past the ceiling,
// and one with no value at all, and says the rule and the running total.
function civilWarChoice(): PendingChoiceView {
  return {
    id: "ch",
    kind: "pick_target",
    chooser: "me",
    from_player: "me",
    count: 0,
    reason: "gain control of up to two target creatures with total mana value 6 or less",
    pick_target: {
      cards: ["four", "three", "two", "seven", "odd"],
      min: 0,
      max: 2,
      total: {
        label: "have total mana value 6 or less",
        max: 6,
        values: { four: 4, three: 3, two: 2, seven: 7 },
      },
    },
  } as PendingChoiceView;
}

const civilWarSaga: CardView = {
  instance_id: "saga",
  name: "The Super Hero Civil War",
  owner: "me",
  controller: "me",
};

describe("a total rule over the chosen targets (#4616)", () => {
  it("greys a card over the ceiling, and one with no value, from the start", () => {
    beginChoice(civilWarChoice(), civilWarSaga);
    const t = get(targeting)!;
    expect(isLegalCardTarget(t, "four")).toBe(true);
    expect(isLegalCardTarget(t, "seven")).toBe(false);
    expect(isLegalCardTarget(t, "odd")).toBe(false);
  });

  it("greys and refuses a second pick that passes the total, and lets the first be un-picked", () => {
    beginChoice(civilWarChoice(), civilWarSaga);
    let t: TargetingState = get(targeting)!;
    t = togglePick(t, { kind: "card", id: "four" });
    expect(pickedTotal(t)).toBe(4);
    expect(breaksSetRule(t, "three")).toBe(true);
    expect(isLegalCardTarget(t, "three")).toBe(false);
    expect(isLegalCardTarget(t, "two")).toBe(true);
    expect(isLegalCardTarget(t, "four")).toBe(true);
    const refused = togglePick(t, { kind: "card", id: "three" });
    expect(refused.picked.map((p) => p.id)).toEqual(["four"]);
    t = togglePick(t, { kind: "card", id: "two" });
    expect(t.picked.map((p) => p.id)).toEqual(["four", "two"]);
    // Un-picking the 4 makes room for the 3 beside the 2.
    t = togglePick(t, { kind: "card", id: "four" });
    expect(isLegalCardTarget(t, "three")).toBe(true);
  });

  it("a cast clause carries the rule too", () => {
    begin(
      {
        instance_id: "s",
        name: "Spell",
        owner: "me",
        controller: "me",
        legal_targets: {
          cards: ["a", "b"],
          min: 0,
          max: 2,
          total: { label: "have total mana value 3 or less", max: 3, values: { a: 2, b: 2 } },
        },
      },
      "creature",
    );
    let t: TargetingState = get(targeting)!;
    t = togglePick(t, { kind: "card", id: "a" });
    expect(isLegalCardTarget(t, "b")).toBe(false);
    cancel();
  });

  it("the dock's prompt says the rule and the running total", () => {
    beginChoice(civilWarChoice(), civilWarSaga);
    let t: TargetingState = get(targeting)!;
    expect(targetingDetail(t)).toContain("targets must have total mana value 6 or less");
    expect(targetingDetail(t)).not.toContain("total 0 of 6");
    t = togglePick(t, { kind: "card", id: "four" });
    expect(targetingDetail(t)).toContain("total 4 of 6");
  });
});
