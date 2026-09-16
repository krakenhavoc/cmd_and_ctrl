import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import type { ActivatedAbilityView } from "./protocol";
import {
  autoCounterChoice,
  counterChoiceKey,
  counterChoices,
  counterCostBlocked,
  counterPaymentParams,
  hasCounterCost,
} from "./counterCost";
import { abilityBlocked } from "./contextMenu.logic";

// counterCost.test.ts — #625, the client half of the counter-removal
// cost: what the picker offers, when it is skipped, what the payload
// carries, and why the menu row greys. crew.test.ts is the precedent.

const heartAlt: ActivatedAbilityView = {
  index: 1,
  label: "Crew — remove a loyalty counter from a planeswalker you control",
  counter_cost_n: 1,
  counter_cost_kind: "loyalty",
  counter_cost_label: "a planeswalker you control",
  counter_cost_options: [
    { card_id: "pw-big", kinds: [{ kind: "loyalty", count: 5 }] },
    { card_id: "pw-small", kinds: [{ kind: "loyalty", count: 1 }] },
  ],
};

const hoardDraw: ActivatedAbilityView = {
  index: 0,
  label: "{T}, Remove a gold counter from this artifact: Draw a card.",
  tap_cost: true,
  counter_cost_n: 1,
  counter_cost_kind: "gold",
  counter_cost_self: true,
  counter_cost_options: [{ card_id: "hoard", kinds: [{ kind: "gold", count: 2 }] }],
};

const fainTreasure: ActivatedAbilityView = {
  index: 1,
  label: "{T}, Remove a counter from a creature you control: Create a Treasure token.",
  tap_cost: true,
  counter_cost_n: 1,
  counter_cost_label: "a creature you control",
  counter_cost_options: [
    {
      card_id: "bear",
      kinds: [
        { kind: "+1/+1", count: 1 },
        { kind: "stun", count: 1 },
      ],
    },
    { card_id: "hydra", kinds: [{ kind: "+1/+1", count: 4 }] },
  ],
};

describe("hasCounterCost", () => {
  it("is keyed on counter_cost_n", () => {
    expect(hasCounterCost(heartAlt)).toBe(true);
    expect(hasCounterCost({ counter_cost_label: "unused" })).toBe(false);
  });
});

describe("counterChoices", () => {
  it("is one choice per (permanent, kind), most counters first", () => {
    expect(counterChoices(fainTreasure)).toEqual([
      { cardID: "hydra", kind: "+1/+1", count: 4 },
      { cardID: "bear", kind: "+1/+1", count: 1 },
      { cardID: "bear", kind: "stun", count: 1 },
    ]);
  });

  it("is empty when nothing can pay", () => {
    expect(counterChoices({ ...heartAlt, counter_cost_options: undefined })).toEqual([]);
  });
});

describe("autoCounterChoice", () => {
  it("skips the prompt for one permanent with one kind", () => {
    expect(autoCounterChoice(hoardDraw)).toEqual({ cardID: "hoard", kind: "gold", count: 2 });
    const oneWalker = { ...heartAlt, counter_cost_options: [heartAlt.counter_cost_options![1]] };
    expect(autoCounterChoice(oneWalker)).toEqual({ cardID: "pw-small", kind: "loyalty", count: 1 });
  });

  it("asks when there are two planeswalkers", () => {
    expect(autoCounterChoice(heartAlt)).toBeNull();
  });

  it("asks when one creature holds two kinds — the kind is part of the answer", () => {
    const oneBear = {
      ...fainTreasure,
      counter_cost_options: [fainTreasure.counter_cost_options![0]],
    };
    expect(autoCounterChoice(oneBear)).toBeNull();
  });

  it("has nothing to choose when nothing can pay", () => {
    expect(autoCounterChoice({ ...heartAlt, counter_cost_options: [] })).toBeNull();
  });
});

describe("counterPaymentParams", () => {
  it("names the planeswalker, not the printed kind", () => {
    expect(counterPaymentParams(heartAlt, { cardID: "pw-big", kind: "loyalty", count: 5 })).toEqual(
      {
        counter_source_ids: ["pw-big"],
      },
    );
  });

  it("names the creature and the kind for 'a counter'", () => {
    expect(counterPaymentParams(fainTreasure, { cardID: "bear", kind: "stun", count: 1 })).toEqual({
      counter_source_ids: ["bear"],
      counter_kind: "stun",
    });
  });

  it("sends nothing for a self cost that prints its kind", () => {
    expect(counterPaymentParams(hoardDraw, { cardID: "hoard", kind: "gold", count: 2 })).toEqual(
      {},
    );
  });

  it("sends nothing for an ability without the component", () => {
    expect(counterPaymentParams({}, { cardID: "x", kind: "gold", count: 1 })).toEqual({});
  });
});

describe("abilityBlocked for a counter cost", () => {
  it("greys Heart of Kiran's alternative with no planeswalker to pay", () => {
    const none = { ...heartAlt, counter_cost_options: undefined };
    expect(abilityBlocked(none, false, false)).toBe(
      "nothing to remove a loyalty counter from (a planeswalker you control)",
    );
    expect(counterCostBlocked(none)).toBe(abilityBlocked(none, false, false));
  });

  it("greys a self cost with no counter", () => {
    const empty = { ...hoardDraw, counter_cost_options: undefined };
    expect(abilityBlocked(empty, false, false)).toBe("no gold counter to remove");
  });

  it("says 'a counter' for the any-kind form", () => {
    const empty = { ...fainTreasure, counter_cost_options: [] };
    expect(abilityBlocked(empty, false, false)).toBe(
      "nothing to remove a counter from (a creature you control)",
    );
  });

  it("leaves a payable row open, even on a tapped Vehicle — crew by counter taps nothing", () => {
    expect(abilityBlocked(heartAlt, true, true)).toBe("");
  });

  it("still reports the tap first on a tap-and-counter cost", () => {
    expect(abilityBlocked(hoardDraw, true, false)).toBe("already tapped");
  });
});

describe("counterChoiceKey", () => {
  it("tells apart the same permanent's kinds, and the same kind on two permanents", () => {
    const keys = new Set([
      counterChoiceKey({ cardID: "a", kind: "+1/+1", count: 1 }),
      counterChoiceKey({ cardID: "a", kind: "stun", count: 1 }),
      counterChoiceKey({ cardID: "b", kind: "+1/+1", count: 1 }),
    ]);
    expect(keys.size).toBe(3);
  });

  // The separator was once a raw NUL byte in the source, which made git
  // treat counterCost.ts as binary: every diff and PR review of it
  // showed "Bin" and no content. The escape keeps the same string.
  it("keeps its source text-diffable (no raw NUL byte)", () => {
    const src = readFileSync(fileURLToPath(new URL("./counterCost.ts", import.meta.url)), "utf8");
    expect(src.includes("\0")).toBe(false);
  });
});
