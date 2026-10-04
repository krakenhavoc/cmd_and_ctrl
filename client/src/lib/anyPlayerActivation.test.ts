// anyPlayerActivation.test.ts — ADR 0106 §1 (#1793), the client half
// of "Any player may activate this ability" (CR 602.2), as pure logic:
//
//   - a seated viewer who does not control the permanent clicks it for
//     its abilities; without such a row, or as a spectator, the click
//     does nothing (decision 6);
//   - that viewer's menu is the any-player rows and nothing else, its
//     life check reads the VIEWER's life (CR 602.1a), and a row the
//     exact digest leaves out greys; the controller and an admin keep
//     the full menu;
//   - the popover lists the same rows (menuAbilityRows);
//   - an opponent's panel lights exactly what the viewer's digest lists
//     for those rows (acrossActions).

import { describe, expect, it } from "vitest";
import type { ActivatedAbilityView, CardView, GameView, PlayerView, ZoneView } from "./protocol";
import {
  ABILITY_NOT_RIGHT_NOW,
  NOT_ENOUGH_LIFE,
  battlefieldClickIntent,
  battlefieldClickPlan,
  buildMenuSections,
  mayActivateAcross,
  menuAbilityRows,
} from "./contextMenu.logic";
import { NO_LEGAL_ACTIONS, acrossActions, legalActionsOf } from "./legalActions";

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

const row = (index: number, extra: Partial<ActivatedAbilityView> = {}): ActivatedAbilityView => ({
  index,
  ref: `own:${index}`,
  label: `ability ${index}`,
  ...extra,
});

// Xantcha-shaped: the any-player row, plus an ordinary row only its
// controller may activate.
const xantcha = (extra: Partial<CardView> = {}): CardView => ({
  instance_id: "xantcha",
  name: "Xantcha, Sleeper Agent",
  owner: ALICE,
  controller: ALICE,
  type_line: "Legendary Creature — Phyrexian Minion",
  activated_abilities: [
    row(0, { label: "{3}: Draw; its controller loses 2 life", any_player: true }),
    row(1, { label: "{1}: Controller-only" }),
  ],
  ...extra,
});

const bear = (): CardView => ({
  instance_id: "bear",
  name: "Grizzly Bears",
  owner: ALICE,
  controller: ALICE,
  type_line: "Creature — Bear",
  activated_abilities: [row(0, { label: "{1}: Pump" })],
});

function gameView(cards: CardView[], extra: Partial<GameView> = {}): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME, 3), seat(ALICE, 40)],
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

describe("the click on another player's permanent", () => {
  it("activates the one any-player row for a seated non-controller (#2201)", () => {
    expect(mayActivateAcross(xantcha(), ME)).toBe(true);
    // The controller-only row is not the viewer's, so the any-player
    // row is the one usable row, and the click activates it.
    expect(battlefieldClickPlan(xantcha(), ME, false)).toEqual({
      intent: "activate",
      row: { kind: "activated", index: 0 },
    });
    // Even with the mana and raw-tap flags: none of it is the viewer's.
    expect(battlefieldClickIntent(xantcha(), ME, false, { manaClick: true, rawTap: true })).toBe(
      "activate",
    );
  });

  it("opens the light popover for a seated non-controller with two any-player rows", () => {
    const two = xantcha({
      activated_abilities: [
        row(0, { label: "{3}: Draw", any_player: true }),
        row(1, { label: "{2}: Scry 1", any_player: true }),
      ],
    });
    expect(battlefieldClickIntent(two, ME, false)).toBe("popover");
  });

  it("does nothing when the digest refuses the any-player row (ADR 0117 §2)", () => {
    // The exact digest lists no activation for the viewer on Xantcha.
    const view = gameView([xantcha()], { legal_actions: { pass: true, sources: {} } });
    const gate = acrossActions(legalActionsOf(view), view.battlefield.cards, ME);
    expect(battlefieldClickIntent(xantcha(), ME, false, { legalGate: gate })).toBe("none");
  });

  it("does nothing when an effect stops its abilities (Arrest)", () => {
    const arrested = xantcha({ restrictions: ["cant_activate"] });
    expect(battlefieldClickIntent(arrested, ME, false)).toBe("none");
  });

  it("does nothing without an any-player row", () => {
    expect(mayActivateAcross(bear(), ME)).toBe(false);
    expect(battlefieldClickIntent(bear(), ME, false)).toBe("none");
  });

  it("does nothing for a spectator", () => {
    expect(mayActivateAcross(xantcha(), null)).toBe(false);
    expect(battlefieldClickIntent(xantcha(), null, false)).toBe("none");
  });

  it("opens the controller's popover too: both rows are theirs", () => {
    expect(mayActivateAcross(xantcha(), ALICE)).toBe(false);
    expect(battlefieldClickIntent(xantcha(), ALICE, false)).toBe("popover");
  });
});

describe("the menu on another player's permanent", () => {
  const sectionsFor = (
    viewer: string | null,
    isAdmin: boolean,
    card = xantcha(),
    extra: Partial<GameView> = {},
  ) => {
    const v = gameView([card], extra);
    const legal = legalActionsOf(v);
    return buildMenuSections(v, card, viewer, isAdmin, legal, legal);
  };

  it("lists only the any-player rows, in one abilities section", () => {
    const sections = sectionsFor(ME, false);
    expect(sections.map((s) => s.id)).toEqual(["abilities"]);
    expect(sections[0].items.map((i) => i.label)).toEqual([
      "{3}: Draw; its controller loses 2 life",
    ]);
    expect(sections[0].items[0].activate).toEqual({ kind: "ability", index: 0 });
  });

  it("is empty for a spectator, and for a permanent with no any-player row", () => {
    expect(sectionsFor(null, false)).toEqual([]);
    expect(sectionsFor(ME, false, bear())).toEqual([]);
  });

  it("is the full menu for the controller", () => {
    const sections = sectionsFor(ALICE, false);
    const ids = sections.map((s) => s.id);
    expect(ids).toContain("state");
    expect(ids).toContain("move");
    const abilities = sections.find((s) => s.id === "abilities")!.items;
    expect(abilities.map((i) => i.activate?.index)).toEqual([0, 1]);
  });

  it("is the full menu for an admin", () => {
    const sections = sectionsFor(ME, true);
    expect(sections.map((s) => s.id)).toContain("state");
    expect(sections.find((s) => s.id === "abilities")!.items).toHaveLength(2);
  });

  it("checks a life cost against the VIEWER's life, the activator's (CR 602.1a)", () => {
    // The viewer has 3 life and the controller 40: a "Pay 5 life" row is
    // unpayable for the viewer, whatever the controller could pay.
    const card = xantcha({
      activated_abilities: [row(0, { label: "Pay 5 life: Draw", life_cost: 5, any_player: true })],
    });
    const item = sectionsFor(ME, false, card)[0].items[0];
    expect(item.disabled).toBe(true);
    expect(item.hint).toBe(NOT_ENOUGH_LIFE);
    // The controller's own menu reads the controller's life.
    const own = sectionsFor(ALICE, false, card).find((s) => s.id === "abilities")!.items[0];
    expect(own.disabled).toBe(false);
  });

  it("lights the row the digest lists, and greys it when the exact digest leaves it out", () => {
    const listed = sectionsFor(ME, false, xantcha(), {
      legal_actions: {
        pass: true,
        sources: { xantcha: { kinds: ["activate"], moves: 1, abilities: ["own:0"] } },
      },
    })[0].items[0];
    expect(listed.ready).toBe(true);
    expect(listed.disabled).toBe(false);

    const leftOut = sectionsFor(ME, false, xantcha(), {
      legal_actions: { pass: true, sources: {} },
    })[0].items[0];
    expect(leftOut.ready).toBeUndefined();
    expect(leftOut.disabled).toBe(true);
    expect(leftOut.hint).toBe(ABILITY_NOT_RIGHT_NOW);
  });

  it("no digest greys nothing new", () => {
    const item = sectionsFor(ME, false)[0].items[0];
    expect(item.disabled).toBe(false);
    expect(item.ready).toBeUndefined();
  });
});

describe("menuAbilityRows: the popover's rows", () => {
  it("is every row for the controller, the any-player rows for anyone else", () => {
    expect(menuAbilityRows(xantcha(), ALICE).map((a) => a.index)).toEqual([0, 1]);
    expect(menuAbilityRows(xantcha(), ME).map((a) => a.index)).toEqual([0]);
    expect(menuAbilityRows(bear(), ME)).toEqual([]);
  });

  it("is nothing for a spectator, and unfiltered with no viewer in scope", () => {
    expect(menuAbilityRows(xantcha(), null)).toEqual([]);
    expect(menuAbilityRows(xantcha(), undefined).map((a) => a.index)).toEqual([0, 1]);
  });

  it("leaves a hand card's zone abilities alone", () => {
    const cycler: CardView = {
      instance_id: "c",
      name: "Cycler",
      owner: ME,
      controller: ME,
      zone_abilities: [row(0, { label: "Cycling {2}" })],
    };
    expect(menuAbilityRows(cycler, ME)).toHaveLength(1);
  });
});

describe("acrossActions: what an opponent's panel lights", () => {
  const digestView = (abilities: string[]) =>
    gameView([xantcha(), bear()], {
      legal_actions: {
        pass: true,
        sources: {
          xantcha: { kinds: ["activate"], moves: abilities.length, abilities },
          // Never sent for another player's ordinary row; here to show
          // the lookup does not pass it through.
          bear: { kinds: ["activate"], moves: 1, abilities: ["own:0"] },
        },
      },
    });

  it("answers for the any-player rows the digest lists", () => {
    const v = digestView(["own:0"]);
    const across = acrossActions(legalActionsOf(v), v.battlefield.cards, ME);
    expect(across.readyAbilityRefs("xantcha")).toEqual(["own:0"]);
    expect(across.isReady("xantcha")).toBe(true);
    expect(across.exact).toBe(true);
  });

  it("answers nothing for a row that is not an any-player row", () => {
    const v = digestView(["own:0", "own:1"]);
    const across = acrossActions(legalActionsOf(v), v.battlefield.cards, ME);
    expect(across.readyAbilityRefs("xantcha")).toEqual(["own:0"]);
    expect(across.readyAbilityRefs("bear")).toEqual([]);
    expect(across.isReady("bear")).toBe(false);
  });

  it("answers nothing for a permanent the viewer controls, or for a spectator", () => {
    const v = digestView(["own:0"]);
    expect(acrossActions(legalActionsOf(v), v.battlefield.cards, ALICE).isReady("xantcha")).toBe(
      false,
    );
    expect(acrossActions(legalActionsOf(v), v.battlefield.cards, null)).toBe(NO_LEGAL_ACTIONS);
  });

  it("knows nothing when the frame carries nothing", () => {
    const v = gameView([xantcha()]);
    expect(acrossActions(legalActionsOf(v), v.battlefield.cards, ME)).toBe(NO_LEGAL_ACTIONS);
  });
});
