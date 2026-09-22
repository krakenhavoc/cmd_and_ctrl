import { describe, it, expect } from "vitest";
import {
  abilityBlocked,
  ABILITY_EXHAUSTED,
  ACTIVATION_CONDITION_UNMET,
  NO_COMMANDER_IDENTITY,
} from "./contextMenu.logic";
import type { ManaAbilityView } from "./protocol";

// exhaust.test.ts — the client half of #1181 and #1183. The server
// ships `exhausted` on an ability row it has already spent ("Activate
// each exhaust ability only once"), and the menu greys the row with
// its own reason.
//
// Its own reason and not ACTIVATION_CONDITION_UNMET because the two
// recover differently: a condition may hold again next turn, an
// exhaust only if the permanent becomes a new object.
//
// #1183 puts the same flag on a MANA ability (Loot, the Pathfinder's
// "Exhaust — {G}, {T}: Add three mana of any one color"). Nothing in
// `abilityBlocked` changed for it: the predicate is structural, so one
// flag under one wire name covers both ability kinds — and that is the
// property worth a test rather than an assumption.

describe("abilityBlocked and exhausted", () => {
  it("names the exhaust for a spent ability", () => {
    expect(abilityBlocked({ exhausted: true }, false, false)).toBe(ABILITY_EXHAUSTED);
  });

  it("is silent for an exhaust ability that has not been used", () => {
    expect(abilityBlocked({}, false, false)).toBe("");
  });

  // Bitter Work prints both gates. "Already activated" is the one that
  // will still be true tomorrow, so it wins.
  it("beats condition_unmet on a row that carries both", () => {
    expect(abilityBlocked({ exhausted: true, condition_unmet: true }, false, false)).toBe(
      ABILITY_EXHAUSTED,
    );
    expect(abilityBlocked({ condition_unmet: true }, false, false)).toBe(
      ACTIVATION_CONDITION_UNMET,
    );
  });

  // A tapped source is the more immediate truth, exactly as it is for
  // every other reason in this function.
  it("keeps the tap reason for a tapped source", () => {
    expect(abilityBlocked({ tap_cost: true, exhausted: true }, true, false)).toBe("already tapped");
  });

  // #1183: a MANA row, typed as the wire types it, through the same
  // predicate the menu's mana loop calls. The row is greyed and kept,
  // never hidden — a player has to be able to see that the permanent
  // prints the ability.
  it("greys a spent exhaust MANA ability with the same reason", () => {
    const loot: ManaAbilityView = {
      index: 0,
      label: "Exhaust — {G}, {T}: Add three mana of any one color.",
      tap_cost: true,
      mana_cost: "{G}",
      exhausted: true,
    };
    expect(abilityBlocked(loot, false, false)).toBe(ABILITY_EXHAUSTED);
    // Untapped and unspent, the same row is offered.
    expect(abilityBlocked({ ...loot, exhausted: false }, false, false)).toBe("");
  });

  // The two mana-only reasons keep their order around it: exhaust is
  // read before adds_no_mana, because "already activated" is the one
  // that can never stop being true.
  it("beats adds_no_mana on a mana row that carries both", () => {
    expect(abilityBlocked({ exhausted: true, adds_no_mana: true }, false, false)).toBe(
      ABILITY_EXHAUSTED,
    );
    expect(abilityBlocked({ adds_no_mana: true }, false, false)).toBe(NO_COMMANDER_IDENTITY);
  });
});
