import { describe, expect, it } from "vitest";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import {
  NOT_ENOUGH_LIFE,
  abilityBlocked,
  buildMenuSections,
  notEnoughLife,
  type MenuSection,
} from "./contextMenu.logic";

// lifeCost.test.ts — #1690. The server ships each ability's
// life_cost (the printed "Pay N life", or since #1688 a computed one
// — War Room, Murderous Betrayal), and the client refused the
// activation anyway. The row itself never greyed, so a player found
// out only after clicking. CR 119.4: paying life equal to your life
// total is legal, so the test is strictly greater than, not "at
// least".

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, name: string, life: number): PlayerView {
  return {
    id,
    name,
    seat: 0,
    life,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}

function view(battlefield: CardView[], life: number): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat("a", "Alice", life), seat("b", "Bob", 40)],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: { seq: 1, number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main1" },
    mulligans_open: false,
  };
}

function itemIn(sections: MenuSection[], id: string) {
  for (const s of sections) {
    for (const i of s.items) {
      if (i.id === id) return i;
    }
  }
  return undefined;
}

// cardWithLifeCost builds a Greed-shaped permanent ("{1}{B}, Pay N
// life: Draw a card.") with the given life-cost component, so each
// scenario below can sit right at the CR 119.4 boundary it's testing
// rather than tying the assertion to Greed's own printed "Pay 1 life".
function cardWithLifeCost(lifeCost: number): CardView {
  return {
    instance_id: "greed",
    name: "Greed",
    owner: "a",
    controller: "a",
    type_line: "Enchantment",
    activated_abilities: [
      {
        index: 0,
        label: `{1}{B}, Pay ${lifeCost} life: Draw a card.`,
        mana_cost: "{1}{B}",
        life_cost: lifeCost,
      },
    ],
  };
}

describe("notEnoughLife", () => {
  it("blocks when the cost is more than the player's life", () => {
    expect(notEnoughLife(2, 1)).toBe(NOT_ENOUGH_LIFE);
  });

  it("allows paying life down to exactly zero (CR 119.4)", () => {
    expect(notEnoughLife(2, 2)).toBe("");
  });

  it("is silent when there's no life cost at all", () => {
    expect(notEnoughLife(undefined, 1)).toBe("");
    expect(notEnoughLife(0, 1)).toBe("");
  });

  it("is silent when there's no life total to check against", () => {
    expect(notEnoughLife(5, undefined)).toBe("");
  });
});

describe("abilityBlocked with a life cost", () => {
  it("blocks a fixed life cost the player can't afford — Greed at 1 life", () => {
    expect(abilityBlocked({ life_cost: 2 }, false, false, undefined, 1)).toBe(NOT_ENOUGH_LIFE);
  });

  it("allows the same cost at exactly enough life", () => {
    expect(abilityBlocked({ life_cost: 2 }, false, false, undefined, 2)).toBe("");
  });

  it("leaves a row with no life cost unaffected", () => {
    expect(abilityBlocked({ tap_cost: true }, false, false, undefined, 1)).toBe("");
  });

  it("keeps the tap reason for a tapped source over an unpayable life cost", () => {
    expect(abilityBlocked({ tap_cost: true, life_cost: 5 }, true, false, undefined, 1)).toBe(
      "already tapped",
    );
  });
});

describe("the context menu greys an unpayable life cost", () => {
  it("greys a 2-life draw ability at 1 life", () => {
    const card = cardWithLifeCost(2);
    const sections = buildMenuSections(view([card], 1), card, "a", false);
    const row = itemIn(sections, "ability-0");
    expect(row?.disabled).toBe(true);
    expect(row?.hint).toBe(NOT_ENOUGH_LIFE);
  });

  it("offers the same ability at 2 life", () => {
    const card = cardWithLifeCost(2);
    const sections = buildMenuSections(view([card], 2), card, "a", false);
    const row = itemIn(sections, "ability-0");
    expect(row?.disabled).toBeFalsy();
  });

  it("leaves a card with no life cost unaffected at 1 life", () => {
    const card = cardWithLifeCost(0);
    card.activated_abilities![0].life_cost = undefined;
    const sections = buildMenuSections(view([card], 1), card, "a", false);
    const row = itemIn(sections, "ability-0");
    expect(row?.disabled).toBeFalsy();
  });
});
