import { describe, expect, it } from "vitest";
import { applyCastChoices, castExileCardsCost, optionalCostPayOptions } from "./targeting";
import { exileCostManaValueTotal, exileCostOptionCards } from "./exileCost";
import type { CardView, PlayerView } from "./protocol";

// #5193: "you may collect evidence N" (CR 701.59a) — a ticked optional
// offer opens the graveyard exile picker as a variable pick with a mana
// value floor, its picks riding cast_spell as exile_ids.

function sprite(options: string[]): CardView {
  return {
    instance_id: "spell",
    name: "Crimestopper Sprite",
    owner: "me",
    controller: "me",
    optional_costs: [
      {
        index: 0,
        key: "collect-evidence",
        label: "Collect evidence 6",
        max_times: 1,
        collect_evidence: 6,
        evidence_options: { cards: options, min: 1, max: options.length },
      },
    ],
  } as CardView;
}

function gy(id: string, manaCost: string): CardView {
  return {
    instance_id: id,
    name: id,
    owner: "me",
    controller: "me",
    mana_cost: manaCost,
  } as CardView;
}

describe("collect evidence", () => {
  it("asks nothing until the offer is ticked", () => {
    expect(castExileCardsCost(sprite(["a", "b"]), {})).toBeUndefined();
  });

  it("opens a variable graveyard pick with the printed floor once ticked", () => {
    const got = castExileCardsCost(sprite(["a", "b"]), { optionalCosts: [0] });
    expect(got?.manaValueMin).toBe(6);
    expect(got?.count).toBe(0);
    expect(got?.options).toEqual(["a", "b"]);
    expect(got?.cost.exile_cost_zone).toBe("graveyard");
  });

  it("is an offer that cannot be taken when the graveyard is short", () => {
    expect(optionalCostPayOptions(sprite([]).optional_costs![0])).toEqual([]);
    expect(optionalCostPayOptions(sprite(["a"]).optional_costs![0])).toEqual(["a"]);
  });

  it("totals the picked cards' mana values", () => {
    const cards = [gy("a", "{2}{U}"), gy("b", "{4}"), gy("c", "")];
    expect(exileCostManaValueTotal(cards, ["a", "b"])).toBe(7);
    expect(exileCostManaValueTotal(cards, ["c"])).toBe(0);
  });

  it("resolves the options out of the graveyard and sends exile_ids", () => {
    const seat = {
      graveyard: { cards: [gy("a", "{3}"), gy("x", "{9}")] },
      hand: { cards: [] },
    } as unknown as PlayerView;
    const claim = castExileCardsCost(sprite(["a"]), { optionalCosts: [0] })!;
    expect(exileCostOptionCards(seat, claim.cost).map((c) => c.instance_id)).toEqual(["a"]);
    const sent: Record<string, unknown> = {};
    applyCastChoices(sent, { optionalCosts: [0], exileIDs: ["a"] });
    expect(sent.exile_ids).toEqual(["a"]);
    expect(sent.optional_costs).toEqual([0]);
  });
});
