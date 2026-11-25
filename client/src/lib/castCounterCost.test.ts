import { describe, expect, it } from "vitest";
import type { AdditionalCostView, AlternativeCostView } from "./protocol";
import {
  additionalCostAsksCounters,
  counterChoices,
  counterCostBlocked,
  counterPaymentParams,
  counterPaymentReady,
  hasCounterCost,
  isMultiCounterCost,
} from "./counterCost";
import { applyCastChoices } from "./targeting";

// castCounterCost.test.ts — #5390: an alternative cost that removes
// counters (Quilled Greatwurm's "You may cast this card from your
// graveyard by removing six counters from among creatures you control")
// carries the activated ability's counter block on the offer, so the
// one counter picker and payload builder serve it, and the payment
// rides cast_spell in activate_ability's words.

const greatwurm: AlternativeCostView = {
  key: "remove_counters_from_graveyard",
  label:
    "Remove six counters from among creatures you control to cast this card from your graveyard",
  mana_cost: "{4}{G}{G}",
  counter_cost_n: 6,
  counter_cost_among: true,
  counter_cost_label: "creatures you control",
  counter_cost_options: [
    { card_id: "bear", kinds: [{ kind: "+1/+1", count: 4 }] },
    {
      card_id: "golem",
      kinds: [
        { kind: "oil", count: 2 },
        { kind: "+1/+1", count: 1 },
      ],
    },
  ],
};

describe("a cast's counter-removal offer", () => {
  it("is the among picker, any kinds", () => {
    expect(hasCounterCost(greatwurm)).toBe(true);
    expect(isMultiCounterCost(greatwurm)).toBe(true);
    expect(counterChoices(greatwurm).map((c) => `${c.cardID}:${c.kind}`)).toEqual([
      "bear:+1/+1",
      "golem:oil",
      "golem:+1/+1",
    ]);
  });

  it("is ready at exactly six, not five or seven", () => {
    const five = [
      { cardID: "bear", kind: "+1/+1", count: 4, n: 4 },
      { cardID: "golem", kind: "oil", count: 2, n: 1 },
    ];
    const six = [
      { cardID: "bear", kind: "+1/+1", count: 4, n: 4 },
      { cardID: "golem", kind: "oil", count: 2, n: 2 },
    ];
    const seven = [...six, { cardID: "golem", kind: "+1/+1", count: 1, n: 1 }];
    expect(counterPaymentReady(greatwurm, five)).toBe(false);
    expect(counterPaymentReady(greatwurm, six)).toBe(true);
    expect(counterPaymentReady(greatwurm, seven)).toBe(false);
  });

  it("says why it can't be paid with five counters in all", () => {
    const short: AlternativeCostView = {
      ...greatwurm,
      counter_cost_options: [{ card_id: "bear", kinds: [{ kind: "+1/+1", count: 5 }] }],
    };
    expect(counterCostBlocked(short)).toBe("only 5 counters among creatures you control");
    expect(counterCostBlocked(greatwurm)).toBe("");
  });

  it("rides cast_spell as the activation's counter fields, mixed kinds per part", () => {
    const pay = counterPaymentParams(greatwurm, [
      { cardID: "bear", kind: "+1/+1", count: 4, n: 4 },
      { cardID: "golem", kind: "oil", count: 2, n: 2 },
    ]);
    const params: Record<string, unknown> = { instance_id: "wurm", from_zone: "graveyard" };
    applyCastChoices(params, { altCost: greatwurm.key, counter: pay });
    expect(params).toEqual({
      instance_id: "wurm",
      from_zone: "graveyard",
      alternative_cost: "remove_counters_from_graveyard",
      counter_source_ids: ["bear", "golem"],
      counter_counts: [4, 2],
      counter_kinds: ["+1/+1", "oil"],
    });
  });

  it("sends one counter_kind when the parts share it", () => {
    const pay = counterPaymentParams(greatwurm, [
      { cardID: "bear", kind: "+1/+1", count: 4, n: 4 },
      { cardID: "golem", kind: "+1/+1", count: 1, n: 1 },
    ]);
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { counter: pay });
    expect(params).toEqual({
      counter_source_ids: ["bear", "golem"],
      counter_counts: [4, 1],
      counter_kind: "+1/+1",
    });
  });

  it("adds nothing to a cast without one", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { altCost: "flashback" });
    expect(params).toEqual({ alternative_cost: "flashback" });
  });
});

// #6121: Hierophant Bio-Titan's additional cost — "you may remove any
// number of +1/+1 counters from among creatures you control" — rides the
// same counter block on the cast's additional_cost, with a floor of zero.
describe("a cast's additional counter-removal cost", () => {
  const titan: AdditionalCostView = {
    counter_cost_kind: "+1/+1",
    counter_cost_among: true,
    counter_cost_variable: true,
    counter_cost_label: "creatures you control",
    counter_cost_options: [{ card_id: "bear", kinds: [{ kind: "+1/+1", count: 3 }] }],
  };

  it("asks only while a counter is there to remove", () => {
    expect(additionalCostAsksCounters(titan)).toBe(true);
    expect(additionalCostAsksCounters({ ...titan, counter_cost_options: [] })).toBe(false);
    expect(additionalCostAsksCounters({})).toBe(false);
    expect(additionalCostAsksCounters(undefined)).toBe(false);
  });

  it("accepts removing none, and any number up to what is there", () => {
    expect(counterPaymentReady(titan, [])).toBe(true);
    const picks = counterChoices(titan).map((c) => ({ ...c, n: 2 }));
    expect(counterPaymentReady(titan, picks)).toBe(true);
    expect(counterPaymentParams(titan, picks).counter_counts).toEqual([2]);
  });
});
