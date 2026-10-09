import { describe, it, expect } from "vitest";

import {
  applyCastChoices,
  castableAlternativeCostsOf,
  castTargetOverride,
  printedCostClaimable,
} from "./targeting";
import type { AlternativeCostView, CardView } from "./protocol";

// bestow.test.ts — ADR 0141 (#2862). Bestow needs no client code of its
// own: a bestow card in hand carries the printed cost and a "Bestow
// {cost}" offer, the cast picker lists both, and the server stamps the
// offer with the enchant creature clause, which castTargetOverride takes
// as the spell's clause. The creature cast keeps the card's own clause,
// which is none.

const bestow: AlternativeCostView = {
  key: "bestow",
  label: "Bestow {2}{B}{B}",
  mana_cost: "{2}{B}{B}",
  target_mode: "creature",
  legal_targets: { cards: ["bear"] },
};

function nighthowler(): CardView {
  return {
    instance_id: "howler",
    name: "Nighthowler",
    owner: "me",
    controller: "me",
    type_line: "Enchantment Creature — Horror",
    mana_cost: "{1}{B}{B}",
    castable_here: true,
    alternative_costs: [bestow],
  };
}

describe("bestow from hand", () => {
  it("offers the creature cast and the bestow cast", () => {
    const card = nighthowler();
    expect(printedCostClaimable(card)).toBe(true);
    expect(castableAlternativeCostsOf(card).map((o) => o.label)).toEqual(["Bestow {2}{B}{B}"]);
  });

  it("targets a creature to enchant only when bestowed", () => {
    const override = castTargetOverride(nighthowler(), { altCost: "bestow" });
    expect(override?.target_mode).toBe("creature");
    expect(override?.legal_targets?.cards).toEqual(["bear"]);
    expect(castTargetOverride(nighthowler(), {})).toBeUndefined();
  });

  it("sends the bestow key with the cast", () => {
    const params: Record<string, unknown> = { instance_id: "howler" };
    applyCastChoices(params, { altCost: "bestow" });
    expect(params).toEqual({ instance_id: "howler", alternative_cost: "bestow" });
  });
});
