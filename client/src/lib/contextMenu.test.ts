import { describe, expect, it } from "vitest";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import {
  buildMenuSections,
  canOverride,
  findCard,
  locateCard,
  moveDestinations,
  type MenuItem,
  type MenuSection,
} from "./contextMenu.logic";

// Builders mirror zoneBrowser.test.ts: only the fields the logic
// reads are populated, so unrelated protocol churn doesn't ripple in.

function card(instance_id: string, owner: string, extra: Partial<CardView> = {}): CardView {
  return {
    instance_id,
    name: instance_id,
    owner,
    controller: owner,
    ...extra,
  };
}

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

interface SeatZones {
  hand?: CardView[];
  graveyard?: CardView[];
  command?: CardView[];
  library?: CardView[];
}

function seat(id: string, name: string, zones: SeatZones = {}): PlayerView {
  return {
    id,
    name,
    seat: 0,
    life: 40,
    library: zone("library", id, zones.library ?? []),
    hand: zone("hand", id, zones.hand ?? []),
    graveyard: zone("graveyard", id, zones.graveyard ?? []),
    command: zone("command", id, zones.command ?? []),
    commander_damage: {},
    life_history: [],
  };
}

interface ViewOpts {
  battlefield?: CardView[];
  exile?: CardView[];
  stack?: CardView[];
  step?: string;
}

function view(seats: PlayerView[], opts: ViewOpts = {}): GameView {
  return {
    id: "g1",
    state: "active",
    seats,
    battlefield: zone("battlefield", undefined, opts.battlefield ?? []),
    stack: zone("stack", undefined, opts.stack ?? []),
    exile: zone("exile", undefined, opts.exile ?? []),
    turn: {
      number: 1,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: opts.step ?? "precombat_main",
    },
    mulligans_open: false,
  };
}

function itemById(sections: MenuSection[], id: string): MenuItem | undefined {
  for (const s of sections) {
    for (const i of s.items) {
      if (i.id === id) return i;
      const nested = (i.items ?? []).find((n) => n.id === id);
      if (nested) return nested;
    }
  }
  return undefined;
}

function sectionIDs(sections: MenuSection[]): string[] {
  return sections.map((s) => s.id);
}

describe("locateCard", () => {
  it("finds a card on the battlefield and reports its owner", () => {
    const c = card("c1", "a");
    const v = view([seat("a", "Alice")], { battlefield: [c] });
    expect(locateCard(v, "c1")).toEqual({ zone: "battlefield", ownerID: "a" });
  });

  it("finds cards in each per-player zone", () => {
    const a = seat("a", "Alice", {
      hand: [card("h1", "a")],
      graveyard: [card("g1", "a")],
      command: [card("cmd1", "a")],
      library: [card("l1", "a")],
    });
    const v = view([a]);
    expect(locateCard(v, "h1")?.zone).toBe("hand");
    expect(locateCard(v, "g1")?.zone).toBe("graveyard");
    expect(locateCard(v, "cmd1")?.zone).toBe("command");
    expect(locateCard(v, "l1")?.zone).toBe("library");
  });

  it("finds a card in the shared exile and stack zones", () => {
    const v = view([seat("a", "Alice")], {
      exile: [card("x1", "b")],
      stack: [card("s1", "a")],
    });
    expect(locateCard(v, "x1")).toEqual({ zone: "exile", ownerID: "b" });
    expect(locateCard(v, "s1")).toEqual({ zone: "stack", ownerID: "a" });
  });

  it("returns null for an unknown instance", () => {
    expect(locateCard(view([seat("a", "Alice")]), "nope")).toBeNull();
  });
});

describe("findCard", () => {
  it("re-resolves a card from the snapshot", () => {
    const c = card("c1", "a", { counters: { "+1/+1": 3 } });
    const v = view([seat("a", "Alice")], { battlefield: [c] });
    expect(findCard(v, "c1")?.counters).toEqual({ "+1/+1": 3 });
    expect(findCard(v, "missing")).toBeNull();
  });
});

describe("canOverride", () => {
  it("allows the controller", () => {
    expect(canOverride(card("c1", "a"), "a", false)).toBe(true);
  });

  it("refuses a different seat", () => {
    expect(canOverride(card("c1", "a"), "b", false)).toBe(false);
  });

  it("refuses a spectator", () => {
    expect(canOverride(card("c1", "a"), null, false)).toBe(false);
  });

  it("allows an admin on anyone's card", () => {
    expect(canOverride(card("c1", "a"), "b", true)).toBe(true);
    expect(canOverride(card("c1", "a"), null, true)).toBe(true);
  });
});

describe("moveDestinations", () => {
  it("drops the zone the card already occupies", () => {
    const ids = moveDestinations("hand").map((d) => d.id);
    expect(ids).not.toContain("hand");
    expect(ids).toContain("battlefield");
    expect(ids).toContain("library-top");
    expect(ids).toContain("library-bottom");
  });

  it("never offers the stack as a destination", () => {
    for (const from of ["battlefield", "hand", "stack"] as const) {
      expect(moveDestinations(from).map((d) => d.zone)).not.toContain("stack");
    }
  });
});

describe("buildMenuSections — gating", () => {
  it("is empty for a card the viewer does not control", () => {
    const v = view([seat("a", "Alice"), seat("b", "Bob")], {
      battlefield: [card("c1", "a")],
    });
    expect(buildMenuSections(v, card("c1", "a"), "b", false)).toEqual([]);
  });

  it("is empty for a card that is not in the snapshot", () => {
    const v = view([seat("a", "Alice")]);
    expect(buildMenuSections(v, card("ghost", "a"), "a", false)).toEqual([]);
  });

  it("gives an admin the full menu on someone else's card", () => {
    const v = view([seat("a", "Alice"), seat("b", "Bob")], {
      battlefield: [card("c1", "a")],
    });
    const sections = buildMenuSections(v, card("c1", "a"), "b", true);
    expect(sectionIDs(sections)).toContain("move");
  });
});

describe("buildMenuSections — battlefield", () => {
  const c = card("c1", "a", { type_line: "Creature — Bear", tapped: true });
  const v = view([seat("a", "Alice"), seat("b", "Bob")], { battlefield: [c] });

  it("offers state, counters, damage, move and sacrifice", () => {
    const ids = sectionIDs(buildMenuSections(v, c, "a", false));
    expect(ids).toEqual(["state", "counters", "damage", "move", "extras"]);
  });

  // #1194 / ADR 0082 decision 9. A face-down permanent's only menu
  // row is the CR 116.2g one — while it is face down it has no
  // abilities at all (CR 708.2a) — and the server decides both that
  // the row exists and whether it is available.
  it("offers turn face up on a face-down permanent it controls", () => {
    const morph = card("m1", "a", {
      name: "",
      type_line: "Creature",
      face_down: true,
      face_down_kind: "morphed",
      special_actions: [
        { kind: "turn_face_up", label: "Turn face up {1}{U}", cost: "{1}{U}", available: true },
      ],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [morph] });
    const sections = buildMenuSections(v2, morph, "a", false);
    expect(sectionIDs(sections)[0]).toBe("special_actions");
    const row = itemById(sections, "special-turn_face_up");
    expect(row?.label).toBe("Turn face up {1}{U}");
    expect(row?.disabled).toBeFalsy();
    expect(row?.action).toEqual({
      type: "special_action",
      params: { card_id: "m1", kind: "turn_face_up", strict: true, auto_tap: true },
      player: "a",
    });
  });

  // CR 708.6 says the CONTROLLER turns it face up, so a stolen morph
  // is turned up by the thief — the one kind whose actor is not the
  // card's owner.
  it("sends a stolen morph's turn-face-up as its controller", () => {
    const stolen: CardView = {
      ...card("m2", "b"),
      controller: "a",
      face_down: true,
      face_down_kind: "morphed",
      special_actions: [
        { kind: "turn_face_up", label: "Turn face up {2}", cost: "{2}", available: true },
      ],
    };
    const v2 = view([seat("a", "Alice"), seat("b", "Bob")], { battlefield: [stolen] });
    const sections = buildMenuSections(v2, stolen, "a", false);
    expect(itemById(sections, "special-turn_face_up")?.action?.player).toBe("a");
  });

  it("gives an ordinary permanent no special-action section", () => {
    expect(sectionIDs(buildMenuSections(v, c, "a", false))).not.toContain("special_actions");
  });

  it("disables the no-op half of the tap toggle", () => {
    const sections = buildMenuSections(v, c, "a", false);
    expect(itemById(sections, "tap")?.disabled).toBe(true);
    expect(itemById(sections, "untap")?.disabled).toBe(false);
  });

  it("disables counter removal when the card has none of that kind", () => {
    const sections = buildMenuSections(v, c, "a", false);
    expect(itemById(sections, "counter-remove-+1/+1")?.disabled).toBe(true);
    expect(itemById(sections, "counter-add-+1/+1")?.action).toEqual({
      type: "add_counter",
      params: { instance_id: "c1", name: "+1/+1", delta: 1 },
    });
  });

  it("enables counter removal once the card carries one", () => {
    const withCounter = card("c2", "a", { counters: { "+1/+1": 2 } });
    const v2 = view([seat("a", "Alice")], { battlefield: [withCounter] });
    const sections = buildMenuSections(v2, withCounter, "a", false);
    expect(itemById(sections, "counter-remove-+1/+1")?.disabled).toBe(false);
  });

  it("surfaces a homebrew counter already on the card for removal", () => {
    const odd = card("c3", "a", { counters: { quest: 1 } });
    const v2 = view([seat("a", "Alice")], { battlefield: [odd] });
    const sections = buildMenuSections(v2, odd, "a", false);
    expect(itemById(sections, "counter-other-remove-quest")?.disabled).toBe(false);
  });

  it("clears exactly the damage that is marked", () => {
    const hurt = card("c4", "a", { damage_marked: 3 });
    const v2 = view([seat("a", "Alice")], { battlefield: [hurt] });
    const sections = buildMenuSections(v2, hurt, "a", false);
    expect(itemById(sections, "damage-clear")?.action).toEqual({
      type: "mark_damage",
      params: { instance_id: "c4", delta: -3 },
    });
    expect(itemById(sections, "damage-remove")?.disabled).toBe(false);
  });

  it("hides the damage-removal rows when nothing is marked", () => {
    const sections = buildMenuSections(v, c, "a", false);
    expect(itemById(sections, "damage-remove")?.disabled).toBe(true);
    expect(itemById(sections, "damage-clear")?.disabled).toBe(true);
  });

  it("lists mana and activated abilities first when the card has them", () => {
    const rock = card("c5", "a", {
      mana_abilities: [{ index: 0, label: "Add {C}{C}", produced: "{C}{C}" }],
      activated_abilities: [{ index: 0, label: "{2}: Draw a card" }],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [rock] });
    const sections = buildMenuSections(v2, rock, "a", false);
    expect(sections[0]?.id).toBe("abilities");
    expect(itemById(sections, "mana-0")?.activate).toEqual({ kind: "mana", index: 0 });
    expect(itemById(sections, "ability-0")?.activate).toEqual({ kind: "ability", index: 0 });
  });

  it("greys a tap-cost ability on an already-tapped permanent", () => {
    const rock = card("c6", "a", {
      tapped: true,
      mana_abilities: [{ index: 0, label: "{T}: Add {G}", tap_cost: true }],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [rock] });
    const sections = buildMenuSections(v2, rock, "a", false);
    expect(itemById(sections, "mana-0")?.disabled).toBe(true);
    expect(itemById(sections, "mana-0")?.hint).toBe("already tapped");
  });

  // --- #530 / #365 / #368 ------------------------------------------
  //
  // The menu reads `summoning_sick` straight off the wire and derives
  // nothing of its own — CR 302.6 is the server's call, and the
  // server now answers it correctly (it used to ship the flag for
  // every permanent that entered this turn, which greyed every
  // Treasure, mana rock and fetchland played that turn). These pin
  // that contract from the client side so nobody reintroduces a
  // client-side "it came down this turn" rule.
  it("does not grey a noncreature's tap ability the server calls healthy", () => {
    const treasure = card("c7", "a", {
      type_line: "Token Artifact — Treasure",
      // The server's answer for an artifact that entered this turn.
      summoning_sick: false,
      mana_abilities: [
        {
          index: 0,
          label: "{T}, Sacrifice: Add one mana of any color",
          tap_cost: true,
          sacrifice_cost: true,
        },
      ],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [treasure] });
    const sections = buildMenuSections(v2, treasure, "a", false);
    expect(itemById(sections, "mana-0")?.disabled).toBe(false);
    expect(itemById(sections, "mana-0")?.hint).toBeUndefined();
  });

  it("still greys a tap ability the server calls summoning sick", () => {
    const bird = card("c8", "a", {
      type_line: "Creature — Bird",
      summoning_sick: true,
      mana_abilities: [{ index: 0, label: "{T}: Add {G}", tap_cost: true }],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [bird] });
    const sections = buildMenuSections(v2, bird, "a", false);
    expect(itemById(sections, "mana-0")?.disabled).toBe(true);
    expect(itemById(sections, "mana-0")?.hint).toBe("summoning sickness");
  });

  // #1190: a discounted ability's row shows the charged cost where the
  // wire used to carry only the printed one, baked into `label`
  // freehand. The row itself never re-renders the cost — that stays
  // free text the catalog author wrote — but the hint says what the
  // engine will actually take, with the printed cost named, whenever
  // charged_mana_cost differs from mana_cost.
  it("hints the printed cost on an ability the engine discounts", () => {
    const discounted = card("c11", "a", {
      activated_abilities: [
        {
          index: 0,
          label: "{3}{R}: Do a thing.",
          mana_cost: "{3}{R}",
          charged_mana_cost: "{1}{R}",
        },
      ],
      mana_abilities: [
        {
          index: 0,
          label: "{3}, {T}: Add {C}{C}{C}.",
          tap_cost: true,
          mana_cost: "{3}",
          charged_mana_cost: "{1}",
        },
      ],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [discounted] });
    const sections = buildMenuSections(v2, discounted, "a", false);
    expect(itemById(sections, "ability-0")?.hint).toBe("printed cost {3}{R}");
    expect(itemById(sections, "ability-0")?.disabled).toBe(false);
    expect(itemById(sections, "mana-0")?.hint).toBe("printed cost {3}");
    expect(itemById(sections, "mana-0")?.disabled).toBe(false);
  });

  // #1296: Dragonfire Blade's equip has no single price — it is {1}
  // cheaper for each colour of the creature it targets — so the row
  // names the range, not a printed-cost note that would read {4}/{4}.
  it("hints the price range on an ability whose price reads its target", () => {
    const blade = card("c13", "a", {
      activated_abilities: [
        {
          index: 0,
          label: "Equip {4}",
          mana_cost: "{4}",
          charged_mana_cost: "{4}",
          target_charged_mana_costs: { vivi: "{2}", golem: "{4}", queen: "" },
        },
      ],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [blade] });
    const sections = buildMenuSections(v2, blade, "a", false);
    expect(itemById(sections, "ability-0")?.hint).toBe("free–{4} depending on the target");
  });

  it("is silent about the cost on an undiscounted ability", () => {
    const plain = card("c12", "a", {
      activated_abilities: [
        {
          index: 0,
          label: "{3}{R}: Do a thing.",
          mana_cost: "{3}{R}",
          charged_mana_cost: "{3}{R}",
        },
      ],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [plain] });
    const sections = buildMenuSections(v2, plain, "a", false);
    expect(itemById(sections, "ability-0")?.hint).toBeUndefined();
  });

  // --- S24 restrictions --------------------------------------------
  //
  // "Its activated abilities can't be activated" (Arrest, Faith's
  // Fetters). The server refuses these activations outright, so the
  // row has to say so rather than open a rejection toast. As with
  // summoning sickness, the client READS the server's answer and
  // derives nothing.
  it("greys every ability of an Arrested permanent", () => {
    const arrested = card("c9", "a", {
      type_line: "Creature — Bird",
      restrictions: ["cant_attack", "cant_block", "cant_activate", "cant_activate_mana"],
      mana_abilities: [{ index: 0, label: "{T}: Add {G}", tap_cost: true }],
      activated_abilities: [{ index: 0, label: "{2}: Draw a card" }],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [arrested] });
    const sections = buildMenuSections(v2, arrested, "a", false);
    expect(itemById(sections, "mana-0")?.disabled).toBe(true);
    expect(itemById(sections, "ability-0")?.disabled).toBe(true);
    expect(itemById(sections, "ability-0")?.hint).toBe("an effect stops its abilities");
  });

  // Faith's Fetters prints "unless they're mana abilities", and the
  // difference is visible on the menu: one row greys, the other does
  // not. A single "restricted" flag would have lost this.
  it("leaves the mana ability of a Fettered permanent alone", () => {
    const fettered = card("c10", "a", {
      type_line: "Land",
      restrictions: ["cant_attack", "cant_block", "cant_activate"],
      mana_abilities: [{ index: 0, label: "{T}: Add {C}", tap_cost: true }],
      activated_abilities: [{ index: 0, label: "{4}, {T}: Target creature can't be blocked" }],
    });
    const v2 = view([seat("a", "Alice")], { battlefield: [fettered] });
    const sections = buildMenuSections(v2, fettered, "a", false);
    expect(itemById(sections, "mana-0")?.disabled).toBe(false);
    expect(itemById(sections, "ability-0")?.disabled).toBe(true);
  });
});

describe("buildMenuSections — combat", () => {
  it("stays hidden outside combat", () => {
    const c = card("c1", "a");
    const v = view([seat("a", "Alice"), seat("b", "Bob")], { battlefield: [c] });
    expect(sectionIDs(buildMenuSections(v, c, "a", false))).not.toContain("combat");
  });

  it("offers each opponent as an attack target during combat", () => {
    const c = card("c1", "a");
    const v = view([seat("a", "Alice"), seat("b", "Bob")], {
      battlefield: [c],
      step: "declare_attackers",
    });
    const sections = buildMenuSections(v, c, "a", false);
    expect(sectionIDs(sections)).toContain("combat");
    expect(itemById(sections, "combat-attack-b")?.action).toEqual({
      type: "declare_attacker",
      // ADR 0080 (#1063): auto_tap rides every attack payload so the
      // CR 508.1a tax can reach the lands. Inert without a tax.
      params: { attacker: "c1", target: "b", auto_tap: true },
    });
    expect(itemById(sections, "combat-attack-a")).toBeUndefined();
  });

  it("offers declared attackers as block assignments", () => {
    const mine = card("c1", "a");
    const theirs = card("c2", "b", { attacking_target: "a" });
    const v = view([seat("a", "Alice"), seat("b", "Bob")], {
      battlefield: [mine, theirs],
      step: "declare_blockers",
    });
    const sections = buildMenuSections(v, mine, "a", false);
    expect(itemById(sections, "combat-block-c2")?.action).toEqual({
      type: "declare_blocker",
      params: { blocker: "c1", attacker: "c2" },
    });
  });

  // #1339, CR 802.4a: at a four-seat table a creature may block only
  // what its controller is defending against. The server names that
  // seat on each attacker as defending_player — for a battle it is the
  // PROTECTOR, not the controller.
  it("offers only the attackers the blocker's controller defends against", () => {
    const mine = card("c1", "b");
    const atB = card("x-b", "a", {
      attacking_target: "b",
      attacking_target_kind: "player",
      defending_player: "b",
    });
    const atC = card("x-c", "a", {
      attacking_target: "c",
      attacking_target_kind: "player",
      defending_player: "c",
    });
    const atMyWalker = card("x-w", "a", {
      attacking_target: "walker",
      attacking_target_kind: "planeswalker",
      defending_player: "b",
    });
    // Seat d controls the battle; seat b protects it.
    const atBattle = card("x-s", "a", {
      attacking_target: "siege",
      attacking_target_kind: "battle",
      defending_player: "b",
    });
    const atTheirBattle = card("x-t", "a", {
      attacking_target: "siege2",
      attacking_target_kind: "battle",
      defending_player: "d",
    });
    const v = view([seat("a", "Alice"), seat("b", "Bob"), seat("c", "Carol"), seat("d", "Dan")], {
      battlefield: [mine, atB, atC, atMyWalker, atBattle, atTheirBattle],
      step: "declare_blockers",
    });
    const sections = buildMenuSections(v, mine, "b", false);
    const offered = (itemById(sections, "combat-block")?.items ?? []).map((i) => i.id);
    expect(offered.sort()).toEqual(["combat-block-x-b", "combat-block-x-s", "combat-block-x-w"]);
  });

  it("offers no block row when every attacker is aimed at somebody else", () => {
    const mine = card("c1", "b");
    const atC = card("x-c", "a", {
      attacking_target: "c",
      attacking_target_kind: "player",
      defending_player: "c",
    });
    const v = view([seat("a", "Alice"), seat("b", "Bob"), seat("c", "Carol")], {
      battlefield: [mine, atC],
      step: "declare_blockers",
    });
    expect(itemById(buildMenuSections(v, mine, "b", false), "combat-block")).toBeUndefined();
  });

  // #318: the bulk affordance sits next to the per-card one and
  // reuses its "pick a defender" submenu shape.
  it("offers attack-with-all per opponent, aiming the whole board at one seat", () => {
    const bf = [
      card("c1", "a", { type_line: "Creature — Bear" }),
      card("c2", "a", { type_line: "Creature — Bear" }),
      card("c3", "a", { type_line: "Creature — Wall", tapped: true }),
    ];
    const v = view([seat("a", "Alice"), seat("b", "Bob"), seat("c", "Carol")], {
      battlefield: bf,
      step: "declare_attackers",
    });
    const sections = buildMenuSections(v, bf[0], "a", false);
    expect(itemById(sections, "combat-attack-all")?.label).toBe("Attack with all 2");
    expect(itemById(sections, "combat-attack-all-c")?.action).toEqual({
      type: "declare_attackers",
      params: {
        attackers: [
          { attacker: "c1", target: "c" },
          { attacker: "c2", target: "c" },
        ],
        auto_tap: true,
      },
    });
    // Never the controller's own seat.
    expect(itemById(sections, "combat-attack-all-a")).toBeUndefined();
  });

  it("disables attack-with-all when nothing on the board can attack", () => {
    const c = card("c1", "a", { type_line: "Creature — Wall", abilities: ["defender"] });
    const v = view([seat("a", "Alice"), seat("b", "Bob")], {
      battlefield: [c],
      step: "declare_attackers",
    });
    const item = itemById(buildMenuSections(v, c, "a", false), "combat-attack-all");
    expect(item?.disabled).toBe(true);
    expect(item?.items).toBeUndefined();
  });

  it("always offers the clear-combat escape hatch during combat", () => {
    const c = card("c1", "a");
    const v = view([seat("a", "Alice")], { battlefield: [c], step: "combat_damage" });
    const sections = buildMenuSections(v, c, "a", false);
    expect(itemById(sections, "combat-clear")?.action).toEqual({ type: "clear_combat" });
    expect(itemById(sections, "combat-clear")?.danger).toBe(true);
  });
});

describe("buildMenuSections — move payloads", () => {
  it("stamps owner on private zones and leaves shared zones bare", () => {
    const c = card("c1", "a");
    const v = view([seat("a", "Alice")], { battlefield: [c] });
    const sections = buildMenuSections(v, c, "a", false);
    expect(itemById(sections, "move-graveyard")?.action).toEqual({
      type: "move_card",
      params: {
        src: { kind: "battlefield" },
        dst: { kind: "graveyard", owner: "a" },
        instance_id: "c1",
      },
    });
    expect(itemById(sections, "move-exile")?.action?.params).toEqual({
      src: { kind: "battlefield" },
      dst: { kind: "exile" },
      instance_id: "c1",
    });
  });

  it("routes a stolen creature to its OWNER's graveyard", () => {
    const stolen = card("c1", "b", { controller: "a" });
    const v = view([seat("a", "Alice"), seat("b", "Bob")], { battlefield: [stolen] });
    const sections = buildMenuSections(v, stolen, "a", false);
    const params = itemById(sections, "move-graveyard")?.action?.params;
    expect(params?.dst).toEqual({ kind: "graveyard", owner: "b" });
  });

  it("marks the library-bottom destination with to_bottom", () => {
    const c = card("c1", "a");
    const v = view([seat("a", "Alice", { hand: [c] })]);
    const sections = buildMenuSections(v, c, "a", false);
    expect(itemById(sections, "move-library-bottom")?.action?.params).toEqual({
      src: { kind: "hand", owner: "a" },
      dst: { kind: "library", owner: "a" },
      instance_id: "c1",
      to_bottom: true,
    });
    expect(itemById(sections, "move-library-top")?.action?.params).toEqual({
      src: { kind: "hand", owner: "a" },
      dst: { kind: "library", owner: "a" },
      instance_id: "c1",
    });
  });

  it("adds the CR 903.9 commander route only for a commander on the field", () => {
    const cmd = card("c1", "a", { is_commander: true });
    const v = view([seat("a", "Alice")], { battlefield: [cmd] });
    const sections = buildMenuSections(v, cmd, "a", false);
    expect(itemById(sections, "move-commander-903-9")?.action?.params).toEqual({
      src: { kind: "battlefield" },
      dst: { kind: "graveyard", owner: "a" },
      instance_id: "c1",
      as_commander: true,
    });

    const plain = card("c2", "a");
    const v2 = view([seat("a", "Alice")], { battlefield: [plain] });
    const plainSections = buildMenuSections(v2, plain, "a", false);
    expect(itemById(plainSections, "move-commander-903-9")).toBeUndefined();
  });
});

describe("buildMenuSections — non-battlefield zones", () => {
  it("gives a hand card move options only", () => {
    const c = card("h1", "a");
    const v = view([seat("a", "Alice", { hand: [c] })]);
    const sections = buildMenuSections(v, c, "a", false);
    expect(sectionIDs(sections)).toEqual(["move"]);
    expect(itemById(sections, "move-hand")).toBeUndefined();
    expect(itemById(sections, "move-battlefield")).toBeDefined();
  });

  // #658 / #659, ADR 0062 Decision 4. A special-action row fires the
  // verb straight from the menu: no targeting, no cost picker, no
  // modal - the whole payload is the card and the kind.
  it("offers the CR 116.2 special actions a hand card declares", () => {
    const c = {
      ...card("h1", "a"),
      special_actions: [{ kind: "foretell", label: "Foretell {2}", cost: "{2}", available: true }],
    };
    const v = view([seat("a", "Alice", { hand: [c] })]);
    const sections = buildMenuSections(v, c, "a", false);
    expect(sectionIDs(sections)).toEqual(["special_actions", "move"]);
    const row = itemById(sections, "special-foretell");
    expect(row?.label).toBe("Foretell {2}");
    expect(row?.disabled).toBeFalsy();
    expect(row?.action).toEqual({
      type: "special_action",
      params: { card_id: "h1", kind: "foretell", strict: true, auto_tap: true },
      player: "a",
    });
  });

  // The server timing answer greys the row rather than hiding it - a
  // player has to be able to see the card has the keyword, and the
  // client must not re-derive a rule (split second differs per kind)
  // it would get backwards.
  // #1342: plot (CR 702.170a) is one more kind on the same surface -
  // the row is the plot button, fired with the kind the server sent.
  it("offers plot as a special-action row", () => {
    const c = {
      ...card("h1", "a"),
      special_actions: [{ kind: "plot", label: "Plot {3}{U}", cost: "{3}{U}", available: true }],
    };
    const v = view([seat("a", "Alice", { hand: [c] })]);
    const sections = buildMenuSections(v, c, "a", false);
    const row = itemById(sections, "special-plot");
    expect(row?.label).toBe("Plot {3}{U}");
    expect(row?.disabled).toBeFalsy();
    expect(row?.action).toEqual({
      type: "special_action",
      params: { card_id: "h1", kind: "plot", strict: true, auto_tap: true },
      player: "a",
    });
  });

  it("greys a special action the server says is unavailable", () => {
    const c = {
      ...card("h1", "a"),
      special_actions: [{ kind: "suspend", label: "Suspend 1", cost: "{R}" }],
    };
    const v = view([seat("a", "Alice", { hand: [c] })]);
    const sections = buildMenuSections(v, c, "a", false);
    const row = itemById(sections, "special-suspend");
    expect(row).toBeDefined();
    expect(row?.disabled).toBe(true);
    expect(row?.hint).toBe("not right now");
  });

  // #1319: Ranar the Ever-Watchful's "the first card you foretell
  // each turn costs {0} to foretell" is a cost modifier, not a
  // rewrite of the row's static label — "Foretell {2}" stays the
  // label, and the discount surfaces as the hint, exactly as an
  // activated ability's charged_mana_cost does.
  it("notes the printed cost as a hint when a discount makes charged_cost differ", () => {
    const c = {
      ...card("h1", "a"),
      special_actions: [
        { kind: "foretell", label: "Foretell {2}", cost: "{2}", charged_cost: "", available: true },
      ],
    };
    const v = view([seat("a", "Alice", { hand: [c] })]);
    const sections = buildMenuSections(v, c, "a", false);
    const row = itemById(sections, "special-foretell");
    expect(row?.label).toBe("Foretell {2}");
    expect(row?.disabled).toBeFalsy();
    expect(row?.hint).toBe("printed cost {2}");
  });

  // The common case: no modifier reached this action, charged_cost
  // equals cost, and there is nothing worth a hint.
  it("shows no cost hint when charged_cost matches the printed cost", () => {
    const c = {
      ...card("h1", "a"),
      special_actions: [
        {
          kind: "foretell",
          label: "Foretell {2}",
          cost: "{2}",
          charged_cost: "{2}",
          available: true,
        },
      ],
    };
    const v = view([seat("a", "Alice", { hand: [c] })]);
    const sections = buildMenuSections(v, c, "a", false);
    const row = itemById(sections, "special-foretell");
    expect(row?.hint).toBeUndefined();
  });

  it("gives a graveyard card a route back to hand and battlefield", () => {
    const c = card("g1", "a");
    const v = view([seat("a", "Alice", { graveyard: [c] })]);
    const sections = buildMenuSections(v, c, "a", false);
    expect(sectionIDs(sections)).toEqual(["move"]);
    expect(itemById(sections, "move-hand")?.action?.params).toEqual({
      src: { kind: "graveyard", owner: "a" },
      dst: { kind: "hand", owner: "a" },
      instance_id: "g1",
    });
    expect(itemById(sections, "move-graveyard")).toBeUndefined();
  });

  it("lets a stuck stack item be moved off the stack", () => {
    const c = card("s1", "a");
    const v = view([seat("a", "Alice")], { stack: [c] });
    const sections = buildMenuSections(v, c, "a", false);
    expect(itemById(sections, "move-graveyard")?.action?.params).toEqual({
      src: { kind: "stack" },
      dst: { kind: "graveyard", owner: "a" },
      instance_id: "s1",
    });
  });
});
