// anyPlayerManaActivation.test.ts — #1864 (ADR 0106 §1 amendment
// 2026-10-16), the client half of an "Any player may activate this
// ability" MANA row (Mana Cache), as pure logic:
//
//   - a seated viewer who does not control the permanent clicks it for
//     mana when its any-player mana row is usable, and gets nothing from
//     an ordinary mana row of a permanent they do not control;
//   - the popover and the override menu list that row and no other mana
//     row, greyed when the exact digest leaves it out;
//   - an opponent's panel lights exactly what the viewer's digest lists
//     for it (acrossActions).

import { describe, expect, it } from "vitest";
import type { CardView, GameView, ManaAbilityView, PlayerView, ZoneView } from "./protocol";
import {
  ABILITY_NOT_RIGHT_NOW,
  anyPlayerManaRows,
  battlefieldClickIntent,
  buildMenuSections,
  mayActivateAcross,
  menuManaRows,
} from "./contextMenu.logic";
import { acrossActions, legalActionsOf } from "./legalActions";

const ME = "me";
const ALICE = "alice";

function zone(kind: string, owner: string | undefined, cards: CardView[] = []): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, life: number): PlayerView {
  return {
    id,
    name: id,
    seat: 0,
    life,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
  };
}

const LABEL =
  "Remove a charge counter from this enchantment: Add {C}. Any player may activate this ability but only during their turn before the end step.";

const cacheRow = (extra: Partial<ManaAbilityView> = {}): ManaAbilityView => ({
  index: 0,
  ref: "own:0",
  label: LABEL,
  produced: "{C}",
  any_player: true,
  ...extra,
});

const manaCache = (
  extra: Partial<CardView> = {},
  row: Partial<ManaAbilityView> = {},
): CardView => ({
  instance_id: "cache",
  name: "Mana Cache",
  owner: ALICE,
  controller: ALICE,
  type_line: "Enchantment",
  mana_abilities: [cacheRow(row)],
  ...extra,
});

const forest = (): CardView => ({
  instance_id: "forest",
  name: "Forest",
  owner: ALICE,
  controller: ALICE,
  type_line: "Basic Land — Forest",
  mana_abilities: [{ index: 0, ref: "own:0", produced: "{G}", tap_cost: true }],
});

function gameView(cards: CardView[], extra: Partial<GameView> = {}): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME, 40), seat(ALICE, 40)],
    battlefield: zone("battlefield", undefined, cards),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    turn: {
      seq: 1,
      number: 1,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
    ...extra,
  };
}

const listed = {
  legal_actions: {
    pass: true,
    sources: { cache: { kinds: ["mana" as const], moves: 1, mana_abilities: ["own:0"] } },
  },
};

describe("the click on another player's Mana Cache", () => {
  it("taps it for mana for a seated non-controller", () => {
    expect(mayActivateAcross(manaCache(), ME)).toBe(true);
    expect(battlefieldClickIntent(manaCache(), ME, false, { manaClick: true })).toBe("mana");
  });

  it("does nothing on another player's turn (the row's condition, stamped for the viewer)", () => {
    const notMine = manaCache({}, { condition_unmet: true });
    expect(battlefieldClickIntent(notMine, ME, false, { manaClick: true })).toBe("none");
  });

  it("does nothing when the exact digest leaves the row out", () => {
    const v = gameView([manaCache()], { legal_actions: { pass: true, sources: {} } });
    const gate = acrossActions(legalActionsOf(v), v.battlefield.cards, ME);
    expect(
      battlefieldClickIntent(manaCache(), ME, false, { manaClick: true, legalGate: gate }),
    ).toBe("none");
  });

  it("does nothing on an ordinary mana source the viewer does not control", () => {
    expect(mayActivateAcross(forest(), ME)).toBe(false);
    expect(battlefieldClickIntent(forest(), ME, false, { manaClick: true })).toBe("none");
  });

  it("does nothing for a spectator", () => {
    expect(mayActivateAcross(manaCache(), null)).toBe(false);
    expect(battlefieldClickIntent(manaCache(), null, false, { manaClick: true })).toBe("none");
  });
});

describe("menuManaRows on another player's permanent", () => {
  it("is the any-player mana rows for a non-controller, every row for the controller", () => {
    const both = manaCache({
      mana_abilities: [cacheRow(), { index: 1, ref: "own:1", produced: "{R}", tap_cost: true }],
    });
    expect(menuManaRows(both, ME).map((a) => a.index)).toEqual([0]);
    expect(menuManaRows(both, ALICE).map((a) => a.index)).toEqual([0, 1]);
    expect(anyPlayerManaRows(both).map((a) => a.index)).toEqual([0]);
    expect(menuManaRows(forest(), ME)).toEqual([]);
    expect(menuManaRows(manaCache(), null)).toEqual([]);
  });
});

describe("the override menu on another player's Mana Cache", () => {
  const sectionsFor = (extra: Partial<GameView> = {}) => {
    const v = gameView([manaCache()], extra);
    const legal = legalActionsOf(v);
    return buildMenuSections(v, manaCache(), ME, false, legal, legal);
  };

  it("lists the any-player mana row, activated as a mana row", () => {
    const sections = sectionsFor();
    expect(sections.map((s) => s.id)).toEqual(["abilities"]);
    expect(sections[0].items).toHaveLength(1);
    expect(sections[0].items[0].label).toBe(LABEL);
    expect(sections[0].items[0].activate).toEqual({ kind: "mana", index: 0 });
    expect(sections[0].items[0].disabled).toBe(false);
  });

  it("lights it when the digest lists it, and greys it when the exact digest leaves it out", () => {
    const lit = sectionsFor(listed)[0].items[0];
    expect(lit.ready).toBe(true);
    const leftOut = sectionsFor({ legal_actions: { pass: true, sources: {} } })[0].items[0];
    expect(leftOut.disabled).toBe(true);
    expect(leftOut.hint).toBe(ABILITY_NOT_RIGHT_NOW);
  });
});

describe("acrossActions and an any-player mana row", () => {
  it("lights the row the digest lists, and no ordinary mana row", () => {
    const v = gameView([manaCache(), forest()], {
      legal_actions: {
        pass: true,
        sources: {
          cache: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
          // Never sent for another player's land; here to show the lookup
          // does not pass it through.
          forest: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
        },
      },
    });
    const across = acrossActions(legalActionsOf(v), v.battlefield.cards, ME);
    expect(across.readyManaRefs("cache")).toEqual(["own:0"]);
    expect(across.kinds("cache")).toEqual(["mana"]);
    expect(across.isReady("cache")).toBe(true);
    expect(across.readyManaRefs("forest")).toEqual([]);
    expect(across.isReady("forest")).toBe(false);
  });
});
