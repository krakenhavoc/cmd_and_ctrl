import { describe, expect, it } from "vitest";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { attackTargetHint, permanentAttackTargets } from "./attackTargets";
import { buildMenuSections } from "./contextMenu.logic";

// attackTargets.test.ts — S27, CR 506.2 / 508.1d.
//
// The set itself is the server's; what is tested here is that the
// client renders it, attributes it to the right seat, and does not
// invent rows when the server published none.

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, name: string, i: number): PlayerView {
  return {
    id,
    name,
    seat: i,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  } as unknown as PlayerView;
}

const me = seat("p-me", "Me", 0);
const them = seat("p-them", "Them", 1);
const third = seat("p-third", "Third", 2);

const bear: CardView = {
  instance_id: "c-bear",
  name: "Grizzly Bears",
  owner: me.id,
  controller: me.id,
  type_line: "Creature — Bear",
  power: 2,
  toughness: 2,
} as unknown as CardView;

const walker: CardView = {
  instance_id: "c-walker",
  name: "Their Walker",
  owner: them.id,
  controller: them.id,
  type_line: "Legendary Planeswalker — Teferi",
  counters: { loyalty: 4 },
} as unknown as CardView;

const siege: CardView = {
  instance_id: "c-siege",
  name: "Invasion of Innistrad",
  owner: them.id,
  controller: them.id,
  type_line: "Battle — Siege",
  protector_player: third.id,
  defense: 5,
} as unknown as CardView;

function view(overrides: Partial<GameView> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [me, them, third],
    battlefield: zone("battlefield", undefined, [bear, walker, siege]),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
      attack_targets: [
        { kind: "player", id: them.id },
        { kind: "player", id: third.id },
        { kind: "planeswalker", id: walker.instance_id },
        { kind: "battle", id: siege.instance_id },
      ],
    },
    ...overrides,
  } as unknown as GameView;
}

describe("permanentAttackTargets", () => {
  it("returns the planeswalker and battle rows, not the player rows", () => {
    // Player rows come from view.seats in the existing menu code and
    // carry the display-name fallback; folding them in here would
    // change what an unrelated row says.
    const rows = permanentAttackTargets(view(), me.id);
    expect(rows.map((r) => r.kind)).toEqual(["planeswalker", "battle"]);
    expect(rows.map((r) => r.id)).toEqual(["c-walker", "c-siege"]);
    expect(rows.map((r) => r.label)).toEqual(["Their Walker", "Invasion of Innistrad"]);
  });

  it("drops the rows for a seat the set was not published for", () => {
    // The wire set belongs to the ACTIVE player, because only the
    // active player declares attackers. An admin driving another
    // seat's card gets seats only, as before S27.
    expect(permanentAttackTargets(view(), them.id)).toEqual([]);
  });

  it("returns nothing when the server published no set", () => {
    const v = view();
    v.turn.attack_targets = [];
    expect(permanentAttackTargets(v, me.id)).toEqual([]);
  });
});

describe("attackTargetHint", () => {
  it("names a battle's PROTECTOR, not its controller", () => {
    // The inversion is the whole mechanic: a battle is cast by one
    // player and defended by another, and it is the protector whose
    // creatures block.
    const hint = attackTargetHint(view(), { id: "c-siege", label: "", kind: "battle" });
    expect(hint).toContain("Third");
    expect(hint).not.toContain("Them");
    expect(hint).toContain("5 defense");
  });

  it("names a planeswalker's controller and its loyalty", () => {
    const hint = attackTargetHint(view(), { id: "c-walker", label: "", kind: "planeswalker" });
    expect(hint).toContain("Them");
    expect(hint).toContain("4 loyalty");
  });
});

describe("the declare-attacker menu", () => {
  it("offers seats, the opponent's planeswalker and the battle", () => {
    const sections = buildMenuSections(view(), bear, me.id, false);
    const combat = sections.find((s) => s.items.some((i) => i.id === "combat-attack"));
    expect(combat).toBeTruthy();
    const attack = combat!.items.find((i) => i.id === "combat-attack")!;
    const ids = (attack.items ?? []).map((i) => i.id);
    expect(ids).toContain(`combat-attack-${them.id}`);
    expect(ids).toContain(`combat-attack-${third.id}`);
    expect(ids).toContain("combat-attack-c-walker");
    expect(ids).toContain("combat-attack-c-siege");
  });

  it("sends the permanent's instance id as declare_attacker's target", () => {
    // Same verb, same payload shape — only the id differs. That is
    // what makes the polymorphic target cheap on the client.
    const sections = buildMenuSections(view(), bear, me.id, false);
    const attack = sections.flatMap((s) => s.items).find((i) => i.id === "combat-attack")!;
    const row = (attack.items ?? []).find((i) => i.id === "combat-attack-c-walker")!;
    expect(row.action?.type).toBe("declare_attacker");
    expect(row.action?.params).toEqual({
      attacker: "c-bear",
      target: "c-walker",
      auto_tap: true,
    });
  });
});
