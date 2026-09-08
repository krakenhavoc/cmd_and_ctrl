import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import {
  begin,
  cancel,
  hasXCost,
  isLegalCardTarget,
  isLegalPlayerTarget,
  legalTargetCount,
  targeting,
} from "./targeting";
import type { CardView } from "./protocol";

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "spell", name: "Spell", owner: "p0", controller: "p0", ...extras };
}

describe("targeting store — S20 legal sets", () => {
  it("uses the server legal set when the card carries one", () => {
    begin(card({ legal_targets: { players: ["p1"], cards: ["bear"] } }), "any");
    const t = get(targeting)!;
    expect(isLegalCardTarget(t, "bear")).toBe(true);
    expect(isLegalCardTarget(t, "rock")).toBe(false);
    expect(isLegalPlayerTarget(t, "p1")).toBe(true);
    expect(isLegalPlayerTarget(t, "p0")).toBe(false);
    expect(legalTargetCount(t)).toBe(2);
    cancel();
    expect(get(targeting)).toBeNull();
  });

  it("falls back to mode heuristics for free-form cards", () => {
    begin(card(), "creature");
    const t = get(targeting)!;
    expect(t.legal).toBeUndefined();
    expect(isLegalCardTarget(t, "anything")).toBe(true);
    expect(isLegalPlayerTarget(t, "p1")).toBe(false);
    expect(legalTargetCount(t)).toBe(-1);
    begin(card(), "player");
    expect(isLegalPlayerTarget(get(targeting)!, "p1")).toBe(true);
    cancel();
  });
});

describe("hasXCost + xValue on the prompt — S20 sub-PR 3", () => {
  it("detects {X} in the printed cost", () => {
    expect(hasXCost(card({ mana_cost: "{X}{R}" }))).toBe(true);
    expect(hasXCost(card({ mana_cost: "{2}{U}" }))).toBe(false);
    expect(hasXCost(card({}))).toBe(false);
  });

  it("carries the announced X through the targeting prompt", () => {
    begin(card({ mana_cost: "{X}{R}", legal_targets: { players: ["p1"] } }), "any", 4);
    expect(get(targeting)?.xValue).toBe(4);
    cancel();
  });
});
