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
      params: { attacker: "c1", target: "b" },
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
