import { describe, expect, it } from "vitest";

import {
  canConfirmSacrificeRange,
  castSacrificeFloor,
  castSacrificeRange,
  sacrificeCeiling,
  toggleSacrificePickInRange,
} from "./sacrificeCost";
import { canCastFromHand } from "./timing";
import type { AdditionalCostView, CardView, GameView, PlayerView, ZoneView } from "./protocol";

// ADR 0100 §3: a variable sacrifice count on a CAST. "Sacrifice any
// number of creatures" ships min 0 / max 0 and "sacrifice X lands"
// ships count_from_x; in both, zero is a legal payment and nothing but
// the board bounds the count.

describe("castSacrificeRange", () => {
  it("reads an explicit 0 / 0 as any number, from zero", () => {
    expect(castSacrificeRange({ cards: ["a"], min: 0, max: 0 })).toEqual({ min: 0, max: 0 });
  });

  it("reads count_from_x as any number, from zero — the number picked is X", () => {
    expect(castSacrificeRange({ cards: [], min: 0, max: 0, count_from_x: true })).toEqual({
      min: 0,
      max: 0,
    });
  });

  it("reads a fixed clause exactly as the ability picker does", () => {
    expect(castSacrificeRange({ cards: [], min: 2, max: 2 })).toEqual({ min: 2, max: 2 });
    expect(castSacrificeRange({ cards: [], min: 1, max: 1 })).toEqual({ min: 1, max: 1 });
  });

  it("reads a view with no bounds as the one-permanent clause", () => {
    expect(castSacrificeRange(undefined)).toEqual({ min: 1, max: 1 });
    expect(castSacrificeRange({ cards: ["a"] })).toEqual({ min: 1, max: 1 });
  });
});

describe("the any-number picker", () => {
  it("confirms at zero and at every count up to the board", () => {
    const range = castSacrificeRange({ cards: ["a", "b", "c"], min: 0, max: 0 });
    expect(sacrificeCeiling(range, 3)).toBe(3);
    expect(canConfirmSacrificeRange([], range, 3)).toBe(true);
    expect(canConfirmSacrificeRange(["a", "b", "c"], range, 3)).toBe(true);
  });

  it("lets the one pick be cleared again when zero is legal", () => {
    expect(toggleSacrificePickInRange(["a"], "a", 1, 0)).toEqual([]);
    expect(toggleSacrificePickInRange([], "a", 1, 0)).toEqual(["a"]);
    // A fixed one-permanent clause still keeps its pick.
    expect(toggleSacrificePickInRange(["a"], "a", 1)).toEqual(["a"]);
  });
});

describe("canCastFromHand — a sacrifice that may be zero", () => {
  function zone(kind: string, cards: CardView[] = []): ZoneView {
    return { kind, owner: "p0", count: cards.length, cards };
  }

  // No cast move for the spell, so canCastFromHand explains why; the
  // sacrifice gate must not be the reason when zero is a payment.
  function snapshot(hand: CardView[]): GameView {
    const seat: PlayerView = {
      id: "p0",
      name: "Me",
      seat: 0,
      life: 40,
      library: zone("library"),
      hand: zone("hand", hand),
      graveyard: zone("graveyard"),
      command: zone("command"),
      commander_damage: {},
      life_history: [],
    };
    return {
      id: "g",
      state: "active",
      seats: [seat],
      battlefield: zone("battlefield"),
      stack: zone("stack"),
      exile: zone("exile"),
      turn: {
        seq: 1,
        number: 1,
        active_seat: 0,
        priority_holder: 0,
        phase: "precombat_main",
        step: "precombat_main",
      },
      mulligans_open: false,
      stack_items: [],
      split_second_active: false,
      legal_moves: [{ type: "pass_priority", player: "p0", kind: "pass", label: "Pass priority" }],
    };
  }

  function spell(cost: AdditionalCostView): CardView {
    return {
      instance_id: "spell",
      name: "Vicious Betrayal",
      type_line: "Sorcery",
      owner: "p0",
      controller: "p0",
      additional_cost: cost,
    };
  }

  it("an empty board does not block a cost that may be paid with none", () => {
    expect(castSacrificeFloor({ cards: [], min: 0, max: 0 })).toBe(0);
    for (const opts of [
      { cards: [], min: 0, max: 0 },
      { cards: [], min: 0, max: 0, count_from_x: true },
    ]) {
      const c = spell({ sacrifice_options: opts });
      expect(canCastFromHand(c, snapshot([c]), "p0").reason ?? "").not.toMatch(/sacrifice/i);
    }
  });

  it("an empty board still blocks a fixed sacrifice", () => {
    const c = spell({ sacrifice_options: { cards: [], min: 1, max: 1 } });
    const verdict = canCastFromHand(c, snapshot([c]), "p0");
    expect(verdict.legal).toBe(false);
    expect(verdict.reason).toBe("Nothing to sacrifice");
  });
});
