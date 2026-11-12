import { describe, expect, it } from "vitest";

import {
  colorSymbolCount,
  exileCostNote,
  exileCostOptionCards,
  exileCostSymbolTotal,
  exileCostVerb,
  exileCostWhere,
  exileCostZone,
} from "./exileCost";
import type { CardView, PlayerView, ZoneView } from "./protocol";

function card(id: string, name: string): CardView {
  return { instance_id: id, name } as CardView;
}

function zone(kind: string, cards: CardView[]): ZoneView {
  return { kind, count: cards.length, cards } as ZoneView;
}

function seat(hand: CardView[], graveyard: CardView[]): PlayerView {
  return { hand: zone("hand", hand), graveyard: zone("graveyard", graveyard) } as PlayerView;
}

describe("exileCost", () => {
  const me = seat(
    [card("h1", "Held Beast"), card("h2", "Held Spell")],
    [card("g1", "Dead Bear"), card("g2", "Spent Bolt"), card("g3", "Old Land")],
  );

  // #1297: Grim Lavamancer's options are graveyard ids, and resolving
  // them out of the HAND (the only pile #1283 knew) would find nothing.
  it("resolves a graveyard clause's options out of the graveyard", () => {
    const got = exileCostOptionCards(me, {
      exile_cost_options: ["g1", "g3"],
      exile_cost_zone: "graveyard",
    });
    expect(got.map((c) => c.name)).toEqual(["Dead Bear", "Old Land"]);
  });

  it("resolves a hand clause's options out of the hand", () => {
    const got = exileCostOptionCards(me, { exile_cost_options: ["h2"], exile_cost_zone: "hand" });
    expect(got.map((c) => c.name)).toEqual(["Held Spell"]);
  });

  // An older server sent no zone, and only ever meant the hand.
  it("reads a clause with no zone as the hand", () => {
    expect(exileCostZone({})).toBe("hand");
    expect(
      exileCostOptionCards(me, { exile_cost_options: ["h1", "g1"] }).map((c) => c.name),
    ).toEqual(["Held Beast"]);
  });

  // #5457: Void Attendant's options are an opponent's cards in the
  // shared exile zone, and the verb is not "Exile".
  it("resolves a processing clause's options out of exile", () => {
    const exile = zone("exile", [card("x1", "Lost Bear"), card("x2", "Lost Bolt")]);
    const clause = { exile_cost_options: ["x2"], exile_cost_zone: "exile" as const };
    expect(exileCostOptionCards(me, clause, exile).map((c) => c.name)).toEqual(["Lost Bolt"]);
    expect(exileCostOptionCards(me, clause)).toEqual([]);
    expect(exileCostVerb(clause)).toBe("Put into graveyard");
    expect(exileCostVerb({})).toBe("Exile");
    expect(exileCostNote(clause)).toContain("owner's graveyard");
    expect(exileCostWhere(clause)).toBe("in exile");
  });

  // #5655: Synthesis Pod's options are the viewer's spells on the
  // shared stack.
  it("resolves an exile-a-spell clause's options out of the stack", () => {
    const stack = zone("stack", [card("s1", "My Bolt"), card("s2", "Their Bolt")]);
    const clause = { exile_cost_options: ["s1"], exile_cost_zone: "stack" as const };
    expect(exileCostOptionCards(me, clause, undefined, stack).map((c) => c.name)).toEqual([
      "My Bolt",
    ]);
    expect(exileCostOptionCards(me, clause)).toEqual([]);
    expect(exileCostVerb(clause)).toBe("Exile");
    expect(exileCostNote(clause)).toContain("without resolving");
    expect(exileCostWhere(clause)).toBe("on the stack");
  });

  it("names the pile in the modal's copy", () => {
    expect(exileCostNote({ exile_cost_zone: "graveyard" })).toContain("graveyard");
    expect(exileCostNote({})).toContain("hand");
    expect(exileCostWhere({ exile_cost_zone: "graveyard" })).toBe("in your graveyard");
    expect(exileCostWhere({})).toBe("in hand");
  });

  it("offers nothing without a seat", () => {
    expect(exileCostOptionCards(undefined, { exile_cost_options: ["g1"] })).toEqual([]);
  });
});

describe("colour symbols among mana costs", () => {
  // Baron Helmut Zemo's boast: hybrid and Phyrexian black symbols count
  // once each, as the engine counts them.
  it("counts plain, hybrid, Phyrexian and two-brid symbols once", () => {
    expect(colorSymbolCount("{B}{B}{B}", "B")).toBe(3);
    expect(colorSymbolCount("{2}{B/G}{G}", "B")).toBe(1);
    expect(colorSymbolCount("{B/P}{B/G/P}", "B")).toBe(2);
    expect(colorSymbolCount("{2/B}{W}", "B")).toBe(1);
    expect(colorSymbolCount("{4}{U}{R}", "B")).toBe(0);
    expect(colorSymbolCount(undefined, "B")).toBe(0);
  });

  it("totals only the picked cards", () => {
    const cards = [
      { instance_id: "a", name: "A", mana_cost: "{B}{B}" },
      { instance_id: "b", name: "B", mana_cost: "{1}{B}" },
      { instance_id: "c", name: "C", mana_cost: "{B}{B}{B}{B}" },
    ] as CardView[];
    expect(exileCostSymbolTotal(cards, ["a", "c"], "B")).toBe(6);
    expect(exileCostSymbolTotal(cards, [], "B")).toBe(0);
  });
});
