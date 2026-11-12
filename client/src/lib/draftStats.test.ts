// #5702: what a pool or a deck is made of.

import { describe, expect, it } from "vitest";

import type { DraftCard } from "./draft";
import { cardTypes, curveSummary, tally } from "./draftStats";

function card(id: string, type_line: string, over: Partial<DraftCard> = {}): DraftCard {
  return { id, name: id, type_line, colors: [], mana_value: 2, ...over };
}

describe("the types of a card", () => {
  it("reads them from the front of the type line", () => {
    expect(cardTypes(card("a", "Creature — Elf Druid"))).toEqual(["creature"]);
    expect(cardTypes(card("b", "Legendary Planeswalker — Jace"))).toEqual(["planeswalker"]);
    expect(cardTypes(card("c", "Instant"))).toEqual(["instant"]);
    expect(cardTypes(card("d", "Sorcery — Arcane"))).toEqual(["sorcery"]);
    expect(cardTypes(card("e", "Enchantment — Aura"))).toEqual(["enchantment"]);
  });

  it("gives a card with two types both, in a fixed order", () => {
    expect(cardTypes(card("a", "Artifact Creature — Golem"))).toEqual(["creature", "artifact"]);
    expect(cardTypes(card("b", "Enchantment Creature — Nymph"))).toEqual([
      "creature",
      "enchantment",
    ]);
    expect(cardTypes(card("c", "Artifact Land"))).toEqual(["artifact", "land"]);
    expect(cardTypes(card("d", "Land Creature — Forest Dryad"))).toEqual(["creature", "land"]);
  });

  it("does not read a subtype as a type", () => {
    expect(cardTypes(card("a", "Artifact — Equipment"))).toEqual(["artifact"]);
    // "Creature — Land Mine" would be a silly card, but the dash is the line.
    expect(cardTypes(card("b", "Creature — Enchantment Spirit"))).toEqual(["creature"]);
    expect(cardTypes(card("c", "Basic Land — Forest"))).toEqual(["land"]);
  });

  it("reads only the front face of a double-faced card", () => {
    expect(cardTypes(card("a", "Creature — Human // Land"))).toEqual(["creature"]);
    expect(cardTypes(card("b", "Sorcery // Land — Forest"))).toEqual(["sorcery"]);
    expect(cardTypes(card("c", "Instant // Instant"))).toEqual(["instant"]);
  });

  it("calls a basic land a land from its flag, with no type line", () => {
    expect(cardTypes({ basic: true })).toEqual(["land"]);
    expect(cardTypes({})).toEqual([]);
  });

  it("does not take a word that only contains a type for the type", () => {
    expect(cardTypes(card("a", "Kindred Instant — Elf"))).toEqual(["instant"]);
    expect(cardTypes(card("b", "Battle — Siege"))).toEqual([]);
  });
});

describe("the tally of a pool", () => {
  const pool = [
    card("bear", "Creature — Bear", { colors: ["G"], mana_value: 2 }),
    card("golem", "Artifact Creature — Golem", { colors: [], mana_value: 4 }),
    card("bolt", "Instant", { colors: ["R"], mana_value: 1 }),
    card("ritual", "Sorcery", { colors: ["B"], mana_value: 2 }),
    card("aura", "Enchantment — Aura", { colors: ["W"], mana_value: 3 }),
    card("jace", "Legendary Planeswalker — Jace", { colors: ["U"], mana_value: 4 }),
    card("gold", "Creature — Elf", { colors: ["G", "W"], mana_value: 7 }),
    card("rock", "Artifact", { colors: [], mana_value: 0 }),
    card("forest", "Basic Land — Forest", { basic: true, mana_value: 0 }),
    card("shrine", "Land", { mana_value: 0 }),
  ];
  const t = tally(pool);

  it("counts creatures against everything else, a card once", () => {
    expect(t.total).toBe(10);
    expect(t.creatures).toBe(3);
    expect(t.nonCreatures).toBe(7);
    expect(t.creatures + t.nonCreatures).toBe(t.total);
  });

  it("counts a card under each of its types", () => {
    expect(t.types).toEqual({
      creature: 3,
      instant: 1,
      sorcery: 1,
      artifact: 2,
      enchantment: 1,
      planeswalker: 1,
      land: 2,
    });
  });

  it("puts cards on the curve by mana value, lands left out, the top column open", () => {
    expect(t.curve.map((c) => c.count)).toEqual([1, 1, 2, 1, 2, 0, 1]);
    expect(t.curve.map((c) => c.label)).toEqual(["0", "1", "2", "3", "4", "5", "6+"]);
  });

  it("counts a colour once per card, several colours apart, lands never", () => {
    expect(t.colours.byColour).toEqual({ W: 2, U: 1, B: 1, R: 1, G: 2 });
    expect(t.colours.multi).toBe(1);
    expect(t.colours.colourless).toBe(2);
  });

  it("is all zeros for no cards", () => {
    const none = tally([]);
    expect(none.total).toBe(0);
    expect(none.creatures).toBe(0);
    expect(Object.values(none.types).every((n) => n === 0)).toBe(true);
    expect(none.curve.every((c) => c.count === 0)).toBe(true);
    expect(none.colours.multi).toBe(0);
  });

  it("counts a land that is also a creature as a creature and a land, off the curve", () => {
    const x = tally([card("arbor", "Land Creature — Forest Dryad", { colors: ["G"] })]);
    expect(x.creatures).toBe(1);
    expect(x.types.land).toBe(1);
    expect(x.types.creature).toBe(1);
    expect(x.curve.every((c) => c.count === 0)).toBe(true);
    expect(x.colours.byColour.G).toBe(0);
  });

  it("moves with the cards: taking one out lowers its counts", () => {
    const less = tally(pool.filter((c) => c.id !== "golem"));
    expect(less.creatures).toBe(2);
    expect(less.types.artifact).toBe(1);
    expect(less.curve[4].count).toBe(1);
  });
});

describe("the curve in words", () => {
  it("says each column, the last as six or more", () => {
    const t = tally([
      card("a", "Creature", { mana_value: 8 }),
      card("b", "Instant", { mana_value: 1 }),
    ]);
    expect(curveSummary(t.curve)).toBe(
      "0 at 0, 1 at 1, 0 at 2, 0 at 3, 0 at 4, 0 at 5, 1 at 6 or more",
    );
  });
});
