import { describe, expect, it } from "vitest";
import type { ActivatedAbilityView } from "./protocol";
import {
  autoCounterChoice,
  counterChoices,
  counterCostBlocked,
  counterCostNeedsPrompt,
  counterPaymentParams,
  counterPaymentReady,
  isSelfAmongCounterCost,
} from "./counterCost";

// counterCostSelfAnyKinds.test.ts — #3987, "Remove two counters from
// Ghost-Spider": two counters of any kinds off the source. The server
// sends it as self AND among with no printed kind, and its one option
// is the source with a row per kind.

function ghostSpider(kinds: { kind: string; count: number }[]): ActivatedAbilityView {
  return {
    index: 0,
    label:
      "Remove two counters from Ghost-Spider: Exile the top card of your library. You may play that card this turn.",
    counter_cost_n: 2,
    counter_cost_self: true,
    counter_cost_among: true,
    counter_cost_options: kinds.length ? [{ card_id: "spider", kinds }] : [],
  };
}

describe("a self any-kinds counter cost", () => {
  it("pays straight from the menu when the source holds one kind", () => {
    const a = ghostSpider([{ kind: "+1/+1", count: 3 }]);
    expect(isSelfAmongCounterCost(a)).toBe(true);
    expect(counterCostNeedsPrompt(a)).toBe(false);
    const auto = autoCounterChoice(a);
    expect(auto).toEqual({ cardID: "spider", kind: "+1/+1", count: 3, n: 2 });
    // No permanents on the wire (a self cost), the printed count, one kind.
    expect(counterPaymentParams(a, auto ?? undefined)).toEqual({
      counter_counts: [2],
      counter_kind: "+1/+1",
    });
  });

  it("asks for the split when the source holds two kinds", () => {
    const a = ghostSpider([
      { kind: "+1/+1", count: 1 },
      { kind: "shield", count: 1 },
    ]);
    expect(autoCounterChoice(a)).toBeNull();
    expect(counterCostNeedsPrompt(a)).toBe(true);
    const picks = counterChoices(a).map((c) => ({ ...c, n: 1 }));
    expect(counterPaymentReady(a, picks)).toBe(true);
    expect(counterPaymentParams(a, picks)).toEqual({
      counter_counts: [1, 1],
      counter_kinds: ["+1/+1", "shield"],
    });
    expect(counterPaymentReady(a, picks.slice(0, 1))).toBe(false);
  });

  it("greys the row when the source holds fewer than two counters", () => {
    expect(counterCostBlocked(ghostSpider([{ kind: "+1/+1", count: 1 }]))).toBe(
      "fewer than 2 counters",
    );
    expect(counterCostBlocked(ghostSpider([]))).toBe("fewer than 2 counters");
    expect(counterCostBlocked(ghostSpider([{ kind: "+1/+1", count: 2 }]))).toBe("");
  });
});
