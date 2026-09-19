import { describe, expect, it } from "vitest";
import type { ActivatedAbilityView, ManaAbilityView } from "./protocol";
import {
  autoCounterChoice,
  counterCostBlocked,
  counterCostNeedsPrompt,
  counterPaymentParams,
  counterPaymentReady,
  hasCounterAddCost,
  hasCounterCost,
  isMultiCounterCost,
} from "./counterCost";
import { abilityBlocked } from "./contextMenu.logic";

// counterCostRest.test.ts — #789, the client half of the rest of the
// counter-cost seam: the same helpers reading a MANA ability, the
// among and variable shapes, and the add-a-counter cost.
//
// The invariant under test is the one the ADR argues for: ONE set of
// helpers, one payload builder and one greyed-row reason, whichever
// ability kind carries the component.

const vividLand: ManaAbilityView = {
  index: 1,
  label: "Add one mana of any color",
  tap_cost: true,
  counter_cost_n: 1,
  counter_cost_kind: "charge",
  counter_cost_self: true,
  counter_cost_options: [{ card_id: "vivid", kinds: [{ kind: "charge", count: 2 }] }],
};

const emptyVividLand: ManaAbilityView = { ...vividLand, counter_cost_options: [] };

const mageRing: ManaAbilityView = {
  index: 1,
  label: "Add {C} for each storage counter removed this way",
  tap_cost: true,
  counter_cost_n: 0,
  counter_cost_kind: "storage",
  counter_cost_self: true,
  counter_cost_variable: true,
  counter_cost_max: 3,
  counter_cost_options: [{ card_id: "network", kinds: [{ kind: "storage", count: 3 }] }],
};

const ironSpider: ActivatedAbilityView = {
  index: 1,
  label: "{2}, Remove two +1/+1 counters from among artifacts you control: Draw a card.",
  mana_cost: "{2}",
  counter_cost_n: 2,
  counter_cost_kind: "+1/+1",
  counter_cost_label: "artifacts you control",
  counter_cost_among: true,
  counter_cost_options: [
    { card_id: "art-a", kinds: [{ kind: "+1/+1", count: 1 }] },
    { card_id: "art-b", kinds: [{ kind: "+1/+1", count: 1 }] },
  ],
};

const devotedDruid: ActivatedAbilityView = {
  index: 1,
  label: "Put a -1/-1 counter on this creature: Untap this creature.",
  counter_cost_add: 1,
  counter_cost_add_kind: "-1/-1",
};

describe("a mana ability's counter cost", () => {
  it("reads through the same helpers an activated ability does", () => {
    expect(hasCounterCost(vividLand)).toBe(true);
    expect(isMultiCounterCost(vividLand)).toBe(false);
  });

  it("never opens a prompt when the source is the only thing that can pay", () => {
    expect(counterCostNeedsPrompt(vividLand)).toBe(false);
    const auto = autoCounterChoice(vividLand);
    expect(auto).toEqual({ cardID: "vivid", kind: "charge", count: 2 });
    // The self form with a printed kind sends nothing at all — the
    // server knows which permanent and which counters.
    expect(counterPaymentParams(vividLand, auto ?? undefined)).toEqual({});
  });

  it("greys the row once the charge counters are gone", () => {
    expect(counterCostBlocked(emptyVividLand)).toBe("no charge counter to remove");
    expect(abilityBlocked(emptyVividLand, false, false)).toBe("no charge counter to remove");
    expect(counterCostNeedsPrompt(emptyVividLand)).toBe(false);
  });
});

describe("a variable counter cost", () => {
  it("always asks, even with one permanent", () => {
    expect(isMultiCounterCost(mageRing)).toBe(true);
    expect(autoCounterChoice(mageRing)).toBeNull();
    expect(counterCostNeedsPrompt(mageRing)).toBe(true);
  });

  it("accepts any count at or above the floor and sends it", () => {
    const pick = [{ cardID: "network", kind: "storage", count: 3, n: 2 }];
    expect(counterPaymentReady(mageRing, pick)).toBe(true);
    expect(counterPaymentParams(mageRing, pick)).toEqual({ counter_counts: [2] });
  });

  it("refuses a payment of nothing", () => {
    expect(counterPaymentReady(mageRing, [])).toBe(false);
  });

  it("honours a printed floor", () => {
    const withFloor = { ...mageRing, counter_cost_n: 2 };
    expect(
      counterPaymentReady(withFloor, [{ cardID: "network", kind: "storage", count: 3, n: 1 }]),
    ).toBe(false);
    expect(
      counterPaymentReady(withFloor, [{ cardID: "network", kind: "storage", count: 3, n: 2 }]),
    ).toBe(true);
  });
});

describe("an among counter cost", () => {
  it("asks for a split and sends one count per permanent", () => {
    expect(counterCostNeedsPrompt(ironSpider)).toBe(true);
    const picks = [
      { cardID: "art-a", kind: "+1/+1", count: 1, n: 1 },
      { cardID: "art-b", kind: "+1/+1", count: 1, n: 1 },
    ];
    expect(counterPaymentReady(ironSpider, picks)).toBe(true);
    expect(counterPaymentParams(ironSpider, picks)).toEqual({
      counter_source_ids: ["art-a", "art-b"],
      counter_counts: [1, 1],
    });
  });

  it("refuses a split that misses the printed total in either direction", () => {
    expect(
      counterPaymentReady(ironSpider, [{ cardID: "art-a", kind: "+1/+1", count: 1, n: 1 }]),
    ).toBe(false);
    expect(
      counterPaymentReady(ironSpider, [
        { cardID: "art-a", kind: "+1/+1", count: 1, n: 1 },
        { cardID: "art-b", kind: "+1/+1", count: 1, n: 2 },
      ]),
    ).toBe(false);
  });

  it("says how short the pool is, which no other row can", () => {
    const short = {
      ...ironSpider,
      counter_cost_options: [{ card_id: "art-a", kinds: [{ kind: "+1/+1", count: 1 }] }],
    };
    expect(counterCostBlocked(short)).toBe("only 1 +1/+1 counters among artifacts you control");
  });
});

describe("an add-a-counter cost", () => {
  it("is a cost with nothing to choose, so it opens no prompt", () => {
    expect(hasCounterAddCost(devotedDruid)).toBe(true);
    expect(hasCounterCost(devotedDruid)).toBe(false);
    expect(counterCostNeedsPrompt(devotedDruid)).toBe(false);
    expect(counterCostBlocked(devotedDruid)).toBe("");
  });

  it("greys the row when CR 118.3 says the counter can't go on", () => {
    const blocked = { ...devotedDruid, counter_add_blocked: true };
    expect(counterCostBlocked(blocked)).toBe("it can't have -1/-1 counters put on it");
    expect(abilityBlocked(blocked, false, false)).toBe("it can't have -1/-1 counters put on it");
  });
});
