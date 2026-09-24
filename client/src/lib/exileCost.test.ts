import { describe, expect, it } from "vitest";

import { exileCostNote, exileCostOptionCards, exileCostWhere, exileCostZone } from "./exileCost";
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
