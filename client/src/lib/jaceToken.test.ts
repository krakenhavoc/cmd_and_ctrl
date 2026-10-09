import { describe, expect, it } from "vitest";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { abilityPopoverModel, battlefieldClickPlan } from "./contextMenu.logic";

// jaceToken.test.ts — ADR 0139 (#2796). Empower Jace makes a Jace
// planeswalker TOKEN whose two loyalty abilities come from a token
// template rather than a printed card. The wire carries them in the same
// `activated_abilities` rows a printed planeswalker's use, so the board
// must treat the token exactly like one: a click activates the one
// payable ability, the popover lists both, and the manual loyalty rows
// an uncatalogued planeswalker gets are not offered on top.

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, name: string): PlayerView {
  return {
    id,
    name,
    seat: 0,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}

function jaceToken(loyalty: number): CardView {
  return {
    instance_id: "jace1",
    name: "Jace",
    owner: "a",
    controller: "a",
    type_line: "Token Planeswalker — Jace",
    colors: ["U"],
    counters: { loyalty },
    token_text: "−1: Surveil 1.\n−3: Draw a card.",
    activated_abilities: [
      { index: 0, ref: "own:0", label: "−1: Surveil 1.", loyalty_cost: -1, sorcery_speed: true },
      { index: 1, ref: "own:1", label: "−3: Draw a card.", loyalty_cost: -3, sorcery_speed: true },
    ],
  };
}

function view(battlefield: CardView[]): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat("a", "Alice"), seat("b", "Bob")],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      seq: 1,
      number: 1,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
  };
}

describe("the Jace planeswalker token (ADR 0139)", () => {
  it("activates its −1 on a click when that is the only payable ability", () => {
    // Two loyalty: the −3 cannot be paid (CR 606.6).
    const tok = jaceToken(2);
    expect(battlefieldClickPlan(tok, "a", false, { view: view([tok]), special: true })).toEqual({
      intent: "activate",
      row: { kind: "activated", index: 0 },
    });
  });

  it("opens the popover with both loyalty abilities and no manual loyalty rows", () => {
    const tok = jaceToken(4);
    expect(battlefieldClickPlan(tok, "a", false, { view: view([tok]), special: true }).intent).toBe(
      "popover",
    );
    const model = abilityPopoverModel({
      card: tok,
      viewerID: "a",
      view: view([tok]),
      mana: true,
      activated: true,
      special: true,
    });
    expect(model.activated.map((r) => r.a.label)).toEqual(["−1: Surveil 1.", "−3: Draw a card."]);
    expect(model.activated.every((r) => !r.blocked)).toBe(true);
    expect(model.loyalty).toEqual([]);
  });

  it("greys its loyalty abilities once one was activated this turn", () => {
    const tok = { ...jaceToken(4), loyalty_activated: true };
    const model = abilityPopoverModel({
      card: tok,
      viewerID: "a",
      view: view([tok]),
      mana: true,
      activated: true,
      special: true,
    });
    expect(model.activated.every((r) => r.blocked)).toBe(true);
  });

  it("offers nothing to a player who does not control it", () => {
    const tok = { ...jaceToken(4), controller: "b", owner: "b" };
    expect(battlefieldClickPlan(tok, "a", false, { view: view([tok]), special: true }).intent).toBe(
      "none",
    );
  });
});
