// ADR 0119 amendment 2026-10-16 §B and §E: the ability card's kind line
// and its cost badge, read from the stack item's kind and the leading
// cost of its verbatim text.

import { describe, it, expect } from "vitest";

import {
  MINUS,
  abilityCardKind,
  abilityCostOf,
  abilityKindLine,
  activatedCostOf,
  loyaltyCostOf,
} from "./stackAbilityCard";
import type { StackLaneItemKind } from "./stackLane";
import { MANA_SYMBOL_META, costPips, pipWords } from "./manaSymbol";

const item = (kind: StackLaneItemKind, text: string) => ({ kind, text });

describe("the kind line (§B, call 20)", () => {
  it("names what the rules call the ability", () => {
    expect(abilityCardKind(item("triggered", "Whenever a creature dies, draw a card."))).toBe(
      "triggered",
    );
    expect(abilityCardKind(item("activated", "{T}: Add {G}."))).toBe("activated");
    expect(abilityCardKind(item("activated", "+1: Draw a card."))).toBe("loyalty");
    expect(abilityCardKind(item("activated", `${MINUS}3: Destroy target creature.`))).toBe(
      "loyalty",
    );
    expect(abilityCardKind(item("activated", "0: Untap target land."))).toBe("loyalty");
  });

  it("has none for a spell or a trigger still waiting to go on the stack", () => {
    expect(abilityCardKind(item("spell", "+1: not an ability"))).toBeNull();
    expect(abilityCardKind(item("pending", "Whenever …"))).toBeNull();
  });

  it("never reads a triggered ability as a loyalty ability", () => {
    expect(abilityCardKind(item("triggered", "+1: looks like a cost"))).toBe("triggered");
  });

  it("words each kind for the mono line", () => {
    expect(abilityKindLine("triggered")).toBe("triggered ability");
    expect(abilityKindLine("activated")).toBe("activated ability");
    expect(abilityKindLine("loyalty")).toBe("loyalty ability");
  });
});

describe("the loyalty cost (§E)", () => {
  it("signs it as printed, with a real minus sign", () => {
    expect(loyaltyCostOf("+1: Up to one target creature gets +2/+0 until end of turn.")).toBe("+1");
    expect(loyaltyCostOf(`${MINUS}2: Sorin deals 4 damage to any target.`)).toBe(`${MINUS}2`);
    expect(loyaltyCostOf("0: Draw a card.")).toBe("0");
    expect(loyaltyCostOf(`${MINUS}X: Destroy each creature with mana value X.`)).toBe(`${MINUS}X`);
    expect(loyaltyCostOf("+12: You get an emblem.")).toBe("+12");
  });

  it("reads a hyphen or an en dash as the minus sign", () => {
    expect(loyaltyCostOf("-2: Return target creature.")).toBe(`${MINUS}2`);
    expect(loyaltyCostOf("–8: You get an emblem.")).toBe(`${MINUS}8`);
  });

  it("reads a bracketed cost and an engine label's cost after its dash", () => {
    expect(loyaltyCostOf("[0]: Draw a card.")).toBe("0");
    expect(loyaltyCostOf("[+2]: Scry 2.")).toBe("+2");
    expect(loyaltyCostOf("Sorin Markov — +2: Sorin deals 2 damage.")).toBe("+2");
  });

  it("finds none where the text does not start with one", () => {
    expect(loyaltyCostOf("{T}: Add {C}{C}.")).toBeNull();
    expect(loyaltyCostOf("3: not signed")).toBeNull();
    expect(loyaltyCostOf("Creatures you control get +1/+1.")).toBeNull();
    expect(loyaltyCostOf("")).toBeNull();
  });
});

describe("the activated cost (§E)", () => {
  it("is the symbols before the colon, never the effect's", () => {
    expect(activatedCostOf("{T}: Add {G}.")).toBe("{T}");
    expect(activatedCostOf("{2}{B}, {T}, Sacrifice a creature: Draw a card.")).toBe("{2}{B}{T}");
    expect(activatedCostOf("{1}{U}{U}: Counter target spell.")).toBe("{1}{U}{U}");
    expect(activatedCostOf("{X}{R}: Deal X damage.")).toBe("{X}{R}");
  });

  it("reads a keyword ability's cost after its keyword", () => {
    expect(activatedCostOf("Equip {2}")).toBe("{2}");
    expect(activatedCostOf("Equip legendary creature {3}")).toBe("{3}");
    expect(activatedCostOf("Boast — {1}{R}: This deals 1 damage to any target.")).toBe("{1}{R}");
  });

  it("is empty for a cost with no symbols, or no cost at all", () => {
    expect(activatedCostOf("Sacrifice a creature: Scry 1.")).toBe("");
    expect(activatedCostOf("Vivi Ornitier — put a +1/+1 counter")).toBe("");
    // A colon inside a granted ability's quotes is not this one's cost.
    expect(activatedCostOf('Creatures you control have "{T}: Add {G}."')).toBe("");
  });
});

describe("the badge (§E, call 23)", () => {
  it("shows a loyalty ability's signed cost", () => {
    expect(abilityCostOf(item("activated", "+1: Draw a card."))).toEqual({
      kind: "loyalty",
      value: "+1",
    });
    expect(abilityCostOf(item("activated", `${MINUS}2: Deal 4 damage.`))).toEqual({
      kind: "loyalty",
      value: `${MINUS}2`,
    });
    expect(abilityCostOf(item("activated", "0: Draw a card."))).toEqual({
      kind: "loyalty",
      value: "0",
    });
  });

  it("shows an activated ability's symbols, {T} included", () => {
    expect(abilityCostOf(item("activated", "{T}: Add {C}{C}."))).toEqual({
      kind: "mana",
      cost: "{T}",
    });
    expect(abilityCostOf(item("activated", "{3}, {T}: Draw a card."))).toEqual({
      kind: "mana",
      cost: "{3}{T}",
    });
  });

  it("shows none for a triggered ability, a spell, or a cost with no symbols", () => {
    expect(abilityCostOf(item("triggered", "{T}: looks like a cost"))).toBeNull();
    expect(abilityCostOf(item("spell", "{1}{U}: not a cost"))).toBeNull();
    expect(abilityCostOf(item("pending", "+1: not on the stack"))).toBeNull();
    expect(abilityCostOf(item("activated", "Sacrifice a creature: Scry 1."))).toBeNull();
  });
});

describe("the tap symbol in a badge's cost", () => {
  it("is drawn as its own pip and spoken as tap", () => {
    expect(costPips("{2}{T}")).toEqual([
      { symbol: "2", kind: "generic" },
      { symbol: "T", kind: "tap" },
    ]);
    expect(pipWords({ symbol: "T", kind: "tap" })).toBe("tap");
    expect(MANA_SYMBOL_META.T.glyph.length).toBeGreaterThan(10);
  });
});
