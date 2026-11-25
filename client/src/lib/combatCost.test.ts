// combatCost.test.ts — #6190: the picks a non-mana cost to attack or
// block is paid with (CR 508.1h, 509.1d).

import { describe, expect, it } from "vitest";

import {
  attackCostAsks,
  blockCostAsks,
  canConfirmCombatCost,
  chooseCombatCostForMe,
  combatCostNote,
  combatCostParams,
  combatCostPrompt,
  combatCostUnpayable,
  combatCostVerb,
  emptyPicks,
  pickedElsewhere,
  togglePick,
} from "./combatCost";
import type { CardView, CombatCostDemandView, GameView } from "./protocol";

function card(id: string, name: string, extra: Partial<CardView> = {}): CardView {
  return {
    instance_id: id,
    name,
    owner: "p1",
    controller: "p1",
    ...extra,
  } as CardView;
}

const landDemand: CombatCostDemandView = {
  kind: "sacrifice",
  clause: "a land",
  source: "Exalted Dragon",
  count: 1,
  candidates: ["l1", "l2"],
};

const tapDemand: CombatCostDemandView = {
  kind: "tap",
  clause: "an untapped creature you control",
  source: "Hollow Warrior",
  count: 1,
  candidates: ["c1"],
};

function viewWith(cards: CardView[]): GameView {
  return { battlefield: { cards } } as GameView;
}

describe("attackCostAsks", () => {
  it("is empty when attacking costs no permanents", () => {
    const view = viewWith([card("a1", "Grizzly Bears")]);
    expect(attackCostAsks(view, ["a1"])).toEqual([]);
  });

  it("resolves each demand's candidates to cards, in the server's order", () => {
    const view = viewWith([
      card("a1", "Exalted Dragon", { attack_cost: [landDemand] }),
      card("l1", "Plains"),
      card("l2", "Island"),
    ]);
    const asks = attackCostAsks(view, ["a1"]);
    expect(asks).toHaveLength(1);
    expect(asks[0].creatureName).toBe("Exalted Dragon");
    expect(asks[0].options.map((c) => c.name)).toEqual(["Plains", "Island"]);
  });

  it("reads the staged blockers for the block half", () => {
    const view = viewWith([
      card("b1", "Hollow Warrior", { block_cost: [tapDemand], blocking_target: "x" }),
      card("b2", "Hollow Warrior", { block_cost: [tapDemand] }),
      card("c1", "Helper"),
    ]);
    const asks = blockCostAsks(view, "p1");
    expect(asks).toHaveLength(1);
    expect(asks[0].creature).toBe("b1");
  });
});

describe("combatCostPrompt", () => {
  it("is null with no demands and carries the verb otherwise", () => {
    expect(combatCostPrompt([], "Attack with 1")).toBeNull();
    const view = viewWith([
      card("a1", "Exalted Dragon", { attack_cost: [landDemand] }),
      card("l1", "Plains"),
    ]);
    const prompt = combatCostPrompt(attackCostAsks(view, ["a1"]), "Attack with 1");
    expect(prompt?.verb).toBe("Attack with 1");
  });
});

describe("the picks", () => {
  const view = viewWith([
    card("a1", "Exalted Dragon", { attack_cost: [landDemand] }),
    card("a2", "Leviathan", {
      attack_cost: [
        { kind: "sacrifice", clause: "two Islands", count: 2, candidates: ["l2", "l3"] },
      ],
    }),
    card("l1", "Plains"),
    card("l2", "Island"),
    card("l3", "Island"),
  ]);
  const asks = attackCostAsks(view, ["a1", "a2"]);

  it("confirms only when every clause has exactly its count", () => {
    let picks = emptyPicks(asks);
    expect(canConfirmCombatCost(picks, asks)).toBe(false);
    picks = togglePick(picks, asks, 0, "l1");
    expect(canConfirmCombatCost(picks, asks)).toBe(false);
    picks = togglePick(picks, asks, 1, "l2");
    picks = togglePick(picks, asks, 1, "l3");
    expect(canConfirmCombatCost(picks, asks)).toBe(true);
  });

  it("holds a clause inside its count, dropping the oldest pick", () => {
    let picks = emptyPicks(asks);
    picks = togglePick(picks, asks, 0, "l1");
    picks = togglePick(picks, asks, 0, "l2");
    expect(picks[0]).toEqual(["l2"]);
  });

  it("lets a permanent pay only one clause", () => {
    let picks = emptyPicks(asks);
    picks = togglePick(picks, asks, 1, "l2");
    picks = togglePick(picks, asks, 0, "l2");
    expect(picks[1]).toEqual([]);
    expect(picks[0]).toEqual(["l2"]);
    expect(pickedElsewhere(picks, 1, "l2")).toBe(true);
  });

  it("unpicks on a second press", () => {
    let picks = emptyPicks(asks);
    picks = togglePick(picks, asks, 0, "l1");
    picks = togglePick(picks, asks, 0, "l1");
    expect(picks[0]).toEqual([]);
  });

  it('fills every clause with distinct candidates for "choose for me"', () => {
    const picks = chooseCombatCostForMe(asks);
    expect(canConfirmCombatCost(picks, asks)).toBe(true);
    expect(new Set(picks.flat()).size).toBe(3);
  });

  it("sends one id list per cost component", () => {
    const tapView = viewWith([
      card("a1", "Hollow Warrior", { attack_cost: [tapDemand] }),
      card("a2", "Exalted Dragon", { attack_cost: [landDemand] }),
      card("c1", "Helper"),
      card("l1", "Plains"),
    ]);
    const mixed = attackCostAsks(tapView, ["a1", "a2"]);
    const picks = chooseCombatCostForMe(mixed);
    expect(combatCostParams(picks, mixed)).toEqual({ tap_ids: ["c1"], sacrifice_ids: ["l1"] });
  });

  it("omits a component nothing was picked for", () => {
    expect(combatCostParams([[]], [asks[0]])).toEqual({});
  });
});

describe("the sentences", () => {
  it("names the verb of each component", () => {
    expect(combatCostVerb(landDemand)).toBe("Sacrifice a land");
    expect(combatCostVerb(tapDemand)).toBe("Tap an untapped creature you control");
    expect(combatCostVerb({ kind: "return", clause: "an enchantment you control", count: 1 })).toBe(
      "Return an enchantment you control to its owner's hand",
    );
  });

  it("summarises a declaration's cost once per clause", () => {
    const view = viewWith([
      card("a1", "Exalted Dragon", { attack_cost: [landDemand] }),
      card("a2", "Exalted Dragon", { attack_cost: [landDemand] }),
      card("l1", "Plains"),
    ]);
    expect(combatCostNote(attackCostAsks(view, ["a1", "a2"]))).toBe("sacrificing a land");
  });

  it("reports a clause the board cannot fill", () => {
    const view = viewWith([
      card("a1", "Leviathan", {
        attack_cost: [{ kind: "sacrifice", clause: "two Islands", count: 2, candidates: ["l2"] }],
      }),
      card("l2", "Island"),
    ]);
    const asks = attackCostAsks(view, ["a1"]);
    expect(combatCostUnpayable(asks)?.demand.clause).toBe("two Islands");
  });
});
