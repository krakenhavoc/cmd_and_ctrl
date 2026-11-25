import { describe, it, expect } from "vitest";

import { applyCastChoices, withCastCreatureType } from "./targeting";
import type { CardView } from "./protocol";

// #6245: "Return X target creatures of the creature type of your choice"
// (Selective Snare). The server ships each type with its legal cards; the
// chosen type narrows the picker and rides cast_spell.
describe("cast_creature_types", () => {
  const card = {
    instance_id: "snare",
    name: "Selective Snare",
    legal_targets: { min: 1, max: 1, cards: [] },
    cast_creature_types: [
      { type: "Elf", cards: ["elf1", "elf2"] },
      { type: "Goblin", cards: ["gob"] },
    ],
  } as unknown as CardView;

  it("narrows the clause to the chosen type's cards", () => {
    expect(withCastCreatureType(card, "Elf").legal_targets?.cards).toEqual(["elf1", "elf2"]);
    expect(withCastCreatureType(card, "Goblin").legal_targets?.cards).toEqual(["gob"]);
  });

  it("leaves the card alone for a type it was not offered", () => {
    expect(withCastCreatureType(card, "Merfolk")).toBe(card);
    expect(withCastCreatureType(card, undefined)).toBe(card);
  });

  it("sends the type as chosen_creature_type", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { chosenCreatureType: "Elf" });
    expect(params.chosen_creature_type).toBe("Elf");
    const none: Record<string, unknown> = {};
    applyCastChoices(none, {});
    expect(none.chosen_creature_type).toBeUndefined();
  });
});
