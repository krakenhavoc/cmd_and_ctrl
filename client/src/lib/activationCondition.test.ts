import { describe, expect, it } from "vitest";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import {
  ACTIVATION_CONDITION_UNMET,
  NO_COMMANDER_IDENTITY,
  abilityBlocked,
  buildMenuSections,
  type MenuSection,
} from "./contextMenu.logic";

// activationCondition.test.ts — #743, ADR 0020's activation-condition
// addendum. The server evaluates an ability's "Activate only if …"
// condition and ships `condition_unmet` on activated AND mana
// abilities (the owner's decision); the menu greys the row the way it
// greys sorcery_speed, rather than offering a click the server will
// refuse.

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

function view(battlefield: CardView[], step = "precombat_main"): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat("a", "Alice"), seat("b", "Bob")],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step },
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

function tectonicEdge(unmet: boolean): CardView {
  return {
    instance_id: "edge",
    name: "Tectonic Edge",
    owner: "a",
    controller: "a",
    type_line: "Land",
    mana_abilities: [{ index: 0, label: "Add {C}", produced: "{C}", tap_cost: true }],
    activated_abilities: [
      {
        index: 0,
        label:
          "{1}, {T}, Sacrifice this land: Destroy target nonbasic land. Activate only if an opponent controls four or more lands.",
        tap_cost: true,
        sacrifice_self: true,
        mana_cost: "{1}",
        condition_unmet: unmet || undefined,
        legal_targets: { cards: ["coffers"] },
      },
    ],
  };
}

describe("abilityBlocked and condition_unmet", () => {
  it("names the condition for an activated ability", () => {
    expect(abilityBlocked({ condition_unmet: true }, false, false)).toBe(
      ACTIVATION_CONDITION_UNMET,
    );
  });

  it("is silent when the flag is absent", () => {
    expect(abilityBlocked({}, false, false)).toBe("");
  });

  // #844, CR 903.4f: Command Tower with no commander, or a colourless
  // one, adds no mana. The server ships adds_no_mana and the row greys
  // with its own reason — a tap cost still comes first, because an
  // already-tapped source is the more immediate truth.
  it("names the missing commander identity for a mana ability", () => {
    expect(abilityBlocked({ adds_no_mana: true, tap_cost: true }, false, false)).toBe(
      NO_COMMANDER_IDENTITY,
    );
    expect(abilityBlocked({ adds_no_mana: true, tap_cost: true }, true, false)).toBe(
      "already tapped",
    );
  });

  it("keeps the tap reason for a tapped source", () => {
    expect(abilityBlocked({ tap_cost: true, condition_unmet: true }, true, false)).toBe(
      "already tapped",
    );
  });

  it("lets a shut sorcery-speed window keep its own, more specific reason", () => {
    const card: CardView = { instance_id: "s", name: "Speaker", owner: "a", controller: "a" };
    const v = view([card], "upkeep");
    const reason = abilityBlocked({ sorcery_speed: true, condition_unmet: true }, false, false, {
      card,
      view: v,
      viewerID: "a",
    });
    expect(reason).toBe("Sorcery-speed only");
  });

  it("names the condition once the sorcery-speed window is open", () => {
    const card: CardView = { instance_id: "s", name: "Speaker", owner: "a", controller: "a" };
    const v = view([card]);
    const reason = abilityBlocked({ sorcery_speed: true, condition_unmet: true }, false, false, {
      card,
      view: v,
      viewerID: "a",
    });
    expect(reason).toBe(ACTIVATION_CONDITION_UNMET);
  });
});

describe("the context menu", () => {
  it("greys Tectonic Edge's destroy while the condition is unmet, and not its mana row", () => {
    const edge = tectonicEdge(true);
    const sections = buildMenuSections(view([edge]), edge, "a", false);
    const destroy = itemIn(sections, "ability-0");
    expect(destroy?.disabled).toBe(true);
    expect(destroy?.hint).toBe(ACTIVATION_CONDITION_UNMET);
    expect(itemIn(sections, "mana-0")?.disabled).toBeFalsy();
  });

  it("offers the destroy once the condition holds", () => {
    const edge = tectonicEdge(false);
    const sections = buildMenuSections(view([edge]), edge, "a", false);
    expect(itemIn(sections, "ability-0")?.disabled).toBeFalsy();
  });

  it("greys a mana ability whose condition is unmet — Temple of the False God", () => {
    const temple: CardView = {
      instance_id: "temple",
      name: "Temple of the False God",
      owner: "a",
      controller: "a",
      type_line: "Land",
      mana_abilities: [
        {
          index: 0,
          label: "Add {C}{C}",
          produced: "{C}{C}",
          tap_cost: true,
          condition_unmet: true,
        },
      ],
    };
    const row = itemIn(buildMenuSections(view([temple]), temple, "a", false), "mana-0");
    expect(row?.disabled).toBe(true);
    expect(row?.hint).toBe(ACTIVATION_CONDITION_UNMET);
  });

  it("shows the restriction, not the condition, on a mana row that is both", () => {
    const temple: CardView = {
      instance_id: "temple",
      name: "Temple of the False God",
      owner: "a",
      controller: "a",
      type_line: "Land",
      restrictions: ["cant_activate_mana"],
      mana_abilities: [
        {
          index: 0,
          label: "Add {C}{C}",
          produced: "{C}{C}",
          tap_cost: true,
          condition_unmet: true,
        },
      ],
    };
    const row = itemIn(buildMenuSections(view([temple]), temple, "a", false), "mana-0");
    expect(row?.disabled).toBe(true);
    expect(row?.hint).toBe("an effect stops its abilities");
  });
});
