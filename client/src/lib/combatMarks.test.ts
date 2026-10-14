import { describe, expect, it } from "vitest";
import { noCombatDamageBadge } from "./combatMarks";

describe("noCombatDamageBadge", () => {
  it("is null for a creature with no mark", () => {
    expect(noCombatDamageBadge({})).toBeNull();
    expect(noCombatDamageBadge({ assigns_no_combat_damage: false })).toBeNull();
  });

  it("labels a creature that assigns no combat damage this turn", () => {
    expect(noCombatDamageBadge({ assigns_no_combat_damage: true })).toEqual({
      text: "NO COMBAT DMG",
      title: "Assigns no combat damage this turn",
    });
  });
});
