// @vitest-environment jsdom
//
// boundedTargets.test.ts — ADR 0109 §9 (#1842): a target bounded by X
// on a statistic other than mana value ("with power X or less"), by an
// exact mana value ("with mana value X"), and by the counters an
// activation's cost removes rather than its X ("power less than or
// equal to the number of +1/+1 counters removed this way"). The server
// ships the superset and each card's value; the picker narrows it.

import { describe, it, expect, afterEach } from "vitest";
import { get } from "svelte/store";

import {
  begin,
  beginForAbility,
  cancel,
  countersRemovedOf,
  isLegalCardTarget,
  targeting,
  withinX,
} from "./targeting";
import type { ActivatedAbilityView, CardView } from "./protocol";
import { cleanup } from "./test/render.svelte";

afterEach(() => {
  targeting.set(null);
  cleanup();
});

function killingGlare(): CardView {
  return {
    instance_id: "glare",
    name: "Killing Glare",
    owner: "me",
    controller: "me",
    mana_cost: "{X}{B}",
    target_mode: "creature",
    legal_targets: {
      cards: ["two", "four"],
      min: 1,
      max: 1,
      power_at_most_x: true,
      powers: { two: 2, four: 4 },
    },
  };
}

describe("X bounds on power, toughness and exact mana value", () => {
  it("narrows a power bound by the announced X", () => {
    begin(killingGlare(), "creature", { xValue: 3 });
    const t = get(targeting)!;
    expect(isLegalCardTarget(t, "two")).toBe(true);
    expect(isLegalCardTarget(t, "four")).toBe(false);
    cancel();
  });

  it("narrows a toughness bound and an exact mana value", () => {
    expect(
      withinX(
        { cards: ["a", "b"], toughness_at_most_x: true, toughnesses: { a: 1, b: 5 } },
        { xValue: 2 },
      ),
    ).toEqual(["a"]);
    expect(
      withinX(
        { cards: ["a", "b", "c"], mana_value_equals_x: true, mana_values: { a: 1, b: 2, c: 3 } },
        { xValue: 2 },
      ),
    ).toEqual(["b"]);
  });

  it("a card with no value meets no bound", () => {
    expect(
      withinX({ cards: ["a", "b"], power_at_most_x: true, powers: { a: 0 } }, { xValue: 9 }),
    ).toEqual(["a"]);
  });
});

describe("a bound read off the counters removed", () => {
  const ability: ActivatedAbilityView = {
    index: 0,
    label: "{T}, Remove one or more +1/+1 counters from this creature: Gain control of …",
    target_mode: "creature",
    counter_cost_n: 1,
    counter_cost_variable: true,
    counter_cost_self: true,
    legal_targets: {
      cards: ["two", "four"],
      min: 1,
      max: 1,
      power_at_most_x: true,
      x_from_counters_removed: true,
      powers: { two: 2, four: 4 },
    },
  } as ActivatedAbilityView;
  const manipulator: CardView = {
    instance_id: "manipulator",
    name: "Simic Manipulator",
    owner: "me",
    controller: "me",
  };

  it("counts the counters the payment removes", () => {
    expect(countersRemovedOf({ counter_counts: [2, 1] }, 1)).toBe(3);
    expect(countersRemovedOf({}, 2)).toBe(2);
    expect(countersRemovedOf(undefined, 2)).toBe(0);
  });

  it("narrows the walk by the counters, not by X", () => {
    beginForAbility(manipulator, ability, [], [], undefined, { counter_counts: [3] });
    let t = get(targeting)!;
    expect(isLegalCardTarget(t, "two")).toBe(true);
    expect(isLegalCardTarget(t, "four")).toBe(false);
    cancel();
    beginForAbility(manipulator, ability, [], [], undefined, { counter_counts: [1] });
    t = get(targeting)!;
    expect(isLegalCardTarget(t, "two")).toBe(false);
    cancel();
  });
});
