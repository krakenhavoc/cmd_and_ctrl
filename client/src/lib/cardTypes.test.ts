import { describe, expect, it } from "vitest";
import { bucketForBattlefield } from "./cardTypes";
import type { CardView } from "./protocol";

const card = (type_line: string): CardView =>
  ({ instance_id: type_line, name: type_line, owner: "a", controller: "a", type_line }) as CardView;

// #4087: one group per permanent type. A permanent lands in exactly one,
// the first that applies in: creature > planeswalker > battle > land >
// artifact > enchantment.
describe("bucketForBattlefield", () => {
  it("gives each permanent type its own group", () => {
    expect(bucketForBattlefield(card("Creature — Elf"))).toBe("creature");
    expect(bucketForBattlefield(card("Basic Land — Forest"))).toBe("land");
    expect(bucketForBattlefield(card("Legendary Planeswalker — Sorin"))).toBe("planeswalker");
    expect(bucketForBattlefield(card("Battle — Siege"))).toBe("battle");
    expect(bucketForBattlefield(card("Legendary Artifact — Equipment"))).toBe("artifact");
    expect(bucketForBattlefield(card("Enchantment — Aura"))).toBe("enchantment");
  });

  it("puts a permanent with several types in the first group that applies", () => {
    expect(bucketForBattlefield(card("Artifact Creature — Golem"))).toBe("creature");
    expect(bucketForBattlefield(card("Enchantment Creature — Satyr"))).toBe("creature");
    expect(bucketForBattlefield(card("Land Creature — Forest Dryad"))).toBe("creature");
    expect(bucketForBattlefield(card("Planeswalker Creature — Gideon"))).toBe("creature");
    expect(bucketForBattlefield(card("Artifact Land"))).toBe("land");
    expect(bucketForBattlefield(card("Enchantment Land"))).toBe("land");
    expect(bucketForBattlefield(card("Legendary Enchantment Artifact"))).toBe("artifact");
    expect(bucketForBattlefield(card("Artifact Vehicle — Crew 3"))).toBe("artifact");
  });

  it("does not match a word that merely contains a type", () => {
    expect(bucketForBattlefield(card("Enchantment — Saga"))).toBe("enchantment");
    expect(bucketForBattlefield(card("Planeswalkers' Rest"))).toBe("other");
  });

  it("keeps a card that names no type in the other group", () => {
    expect(bucketForBattlefield({ ...card(""), type_line: undefined } as CardView)).toBe("other");
    expect(bucketForBattlefield(card("Kindred"))).toBe("other");
  });
});
