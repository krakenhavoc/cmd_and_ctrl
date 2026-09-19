import { describe, expect, it } from "vitest";
import type { ActivatedAbilityView } from "./protocol";
import {
  counterChoiceKey,
  counterChoices,
  counterCostBlocked,
  counterCostNeedsPrompt,
  counterPaymentParams,
  counterPaymentReady,
  isMultiCounterCost,
} from "./counterCost";

// counterCostAnyKind.test.ts — #943, the client half of the last
// counter-cost shape: an ANY-KIND removal split across permanents.
//
// The point of the shape is that it needs no new picker. A row has
// always been a (permanent, KIND) pair, so a permanent holding two
// kinds simply offers two rows, and the many-pick stepper on each row
// IS the kind choice. What is new is only the payload: one kind while
// one kind will do, a parallel array when the parts actually differ.

// tekuthal is "{1}{U/P}{U/P}, Remove three counters from among other
// artifacts, creatures, and planeswalkers you control" — among, with
// no printed kind.
const tekuthal: ActivatedAbilityView = {
  index: 0,
  label:
    "{1}{U/P}{U/P}, Remove three counters from among other artifacts, creatures, and planeswalkers you control: Put an indestructible counter on Tekuthal.",
  mana_cost: "{1}{U/P}{U/P}",
  counter_cost_n: 3,
  counter_cost_label: "other artifacts, creatures, and planeswalkers you control",
  counter_cost_among: true,
  counter_cost_options: [
    {
      card_id: "bear",
      kinds: [
        { kind: "+1/+1", count: 2 },
        { kind: "shield", count: 1 },
      ],
    },
    { card_id: "walker", kinds: [{ kind: "loyalty", count: 1 }] },
  ],
};

describe("an any-kind among counter cost", () => {
  it("is a many-pick cost that always prompts", () => {
    expect(isMultiCounterCost(tekuthal)).toBe(true);
    expect(counterCostNeedsPrompt(tekuthal)).toBe(true);
  });

  it("offers one row per (permanent, kind), most counters first", () => {
    const rows = counterChoices(tekuthal);
    expect(rows.map((r) => [r.cardID, r.kind, r.count])).toEqual([
      ["bear", "+1/+1", 2],
      ["bear", "shield", 1],
      ["walker", "loyalty", 1],
    ]);
    // Two rows on one permanent are two distinct selections, which is
    // what lets it pay in two kinds.
    expect(new Set(rows.map(counterChoiceKey)).size).toBe(3);
  });

  it("sends a kind per permanent when the picks mix kinds", () => {
    const picks = [
      { cardID: "bear", kind: "+1/+1", count: 2, n: 2 },
      { cardID: "walker", kind: "loyalty", count: 1, n: 1 },
    ];
    expect(counterPaymentReady(tekuthal, picks)).toBe(true);
    expect(counterPaymentParams(tekuthal, picks)).toEqual({
      counter_source_ids: ["bear", "walker"],
      counter_counts: [2, 1],
      counter_kinds: ["+1/+1", "loyalty"],
    });
  });

  it("names one permanent twice when it pays in two kinds", () => {
    const picks = [
      { cardID: "bear", kind: "+1/+1", count: 2, n: 2 },
      { cardID: "bear", kind: "shield", count: 1, n: 1 },
    ];
    expect(counterPaymentReady(tekuthal, picks)).toBe(true);
    expect(counterPaymentParams(tekuthal, picks)).toEqual({
      counter_source_ids: ["bear", "bear"],
      counter_counts: [2, 1],
      counter_kinds: ["+1/+1", "shield"],
    });
  });

  it("keeps the older single-kind payload when one kind pays", () => {
    const picks = [
      { cardID: "bear", kind: "+1/+1", count: 2, n: 2 },
      { cardID: "servo", kind: "+1/+1", count: 1, n: 1 },
    ];
    expect(counterPaymentParams(tekuthal, picks)).toEqual({
      counter_source_ids: ["bear", "servo"],
      counter_counts: [2, 1],
      counter_kind: "+1/+1",
    });
  });

  it("still refuses a split that misses the printed total", () => {
    expect(counterPaymentReady(tekuthal, [{ cardID: "bear", kind: "+1/+1", count: 2, n: 2 }])).toBe(
      false,
    );
    expect(
      counterPaymentReady(tekuthal, [
        { cardID: "bear", kind: "+1/+1", count: 2, n: 2 },
        { cardID: "bear", kind: "shield", count: 1, n: 1 },
        { cardID: "walker", kind: "loyalty", count: 1, n: 1 },
      ]),
    ).toBe(false);
  });

  it("counts every kind when it says how short the pool is", () => {
    expect(counterCostBlocked(tekuthal)).toBe("");
    const short = {
      ...tekuthal,
      counter_cost_options: [{ card_id: "bear", kinds: [{ kind: "+1/+1", count: 2 }] }],
    };
    expect(counterCostBlocked(short)).toBe(
      "only 2 counters among other artifacts, creatures, and planeswalkers you control",
    );
  });
});
