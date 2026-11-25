// tapAcrossMana.test.ts — #6204 (Piracy): "Until end of turn, you may tap
// lands you don't control for mana." The server stamps `tap_across` on the
// viewer's own copy of another player's land's mana row; the client then
// treats it as it treats an any-player mana row: the click taps it, the
// menu lists it, and the opponent's panel lights what the digest lists.

import { describe, expect, it } from "vitest";
import type { CardView, GameView, ManaAbilityView, PlayerView, ZoneView } from "./protocol";
import {
  acrossManaRow,
  anyPlayerManaRows,
  battlefieldClickIntent,
  mayActivateAcross,
  menuManaRows,
} from "./contextMenu.logic";
import { acrossActions, legalActionsOf } from "./legalActions";

const ME = "me";
const ALICE = "alice";

function zone(kind: string, owner: string | undefined, cards: CardView[] = []): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string): PlayerView {
  return {
    id,
    name: id,
    seat: 0,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
  };
}

const mountain = (row: Partial<ManaAbilityView> = {}): CardView => ({
  instance_id: "mountain",
  name: "Mountain",
  owner: ALICE,
  controller: ALICE,
  type_line: "Basic Land — Mountain",
  mana_abilities: [{ index: 0, ref: "own:0", produced: "{R}", tap_cost: true, ...row }],
});

function gameView(cards: CardView[], extra: Partial<GameView> = {}): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME), seat(ALICE)],
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

describe("another player's land under Piracy", () => {
  it("is a mana click for the viewer whose copy is marked tap_across", () => {
    const land = mountain({ tap_across: true });
    expect(acrossManaRow(land.mana_abilities![0])).toBe(true);
    expect(mayActivateAcross(land, ME)).toBe(true);
    expect(battlefieldClickIntent(land, ME, false, { manaClick: true })).toBe("mana");
    expect(menuManaRows(land, ME)).toHaveLength(1);
    expect(anyPlayerManaRows(land)).toHaveLength(1);
  });

  it("is nothing without the mark", () => {
    const land = mountain();
    expect(mayActivateAcross(land, ME)).toBe(false);
    expect(battlefieldClickIntent(land, ME, false, { manaClick: true })).toBe("none");
    expect(menuManaRows(land, ME)).toHaveLength(0);
  });

  it("follows the digest: left out, the click does nothing", () => {
    const land = mountain({ tap_across: true });
    const listed = gameView([land], {
      legal_actions: {
        pass: true,
        sources: { mountain: { kinds: ["mana" as const], moves: 1, mana_abilities: ["own:0"] } },
      },
    });
    const live = acrossActions(legalActionsOf(listed), listed.battlefield.cards, ME);
    expect(battlefieldClickIntent(land, ME, false, { manaClick: true, legalGate: live })).toBe(
      "mana",
    );
    const empty = gameView([land], { legal_actions: { pass: true, sources: {} } });
    const gate = acrossActions(legalActionsOf(empty), empty.battlefield.cards, ME);
    expect(battlefieldClickIntent(land, ME, false, { manaClick: true, legalGate: gate })).toBe(
      "none",
    );
  });
});
