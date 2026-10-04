// castAnyway.test.ts — ADR 0118 §2 (#2188), the pure half of "Cast
// anyway (don't pay)": when the row is greyed (castAnywayBlocked, one
// case per denial), the row's builder, the dock confirmation it opens
// (castAnywayConfirmRequest), the admin menu's "cast" section, and what
// a confirmed cast puts on the wire (CastChoices.forceCast through the
// cast chain).

import { describe, it, expect, afterEach, vi } from "vitest";
import { get } from "svelte/store";

import {
  castAnywayBlocked,
  castAnywayOffered,
  canCastFromHand,
  type CastAnywayZone,
} from "./timing";
import { castAnywayItem, buildMenuSections } from "./contextMenu.logic";
import { castAnywayConfirmLabel, castAnywayConfirmRequest } from "./targetingDock";
import {
  CAST_ANYWAY_LABEL,
  CAST_ANYWAY_TITLE,
  castAnywayPending,
  clearCastAnyway,
  requestCastAnyway,
} from "./castAnyway";
import { castStripCastAnywayBlocked, castStripEntries } from "./castStrip";
import {
  allPicks,
  applyCastChoices,
  begin,
  beginForModes,
  cancel,
  castChoicesBase,
  targeting,
  togglePick,
} from "./targeting";
import { stampManaEnforcement } from "./manaEnforcement";
import { dockKeyAction, takesBar } from "./dock";
import type { CardView, GameView, LegalMoveView, PlayerView, ZoneView } from "./protocol";

afterEach(() => {
  cancel();
  clearCastAnyway();
});

const ME = "me";
const THEM = "them";

const zone = (kind: string, owner: string, cards: CardView[] = []): ZoneView => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

function seat(
  id: string,
  idx: number,
  hand: CardView[] = [],
  command: CardView[] = [],
): PlayerView {
  return {
    id,
    name: id === ME ? "Me" : "Them",
    seat: idx,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id),
    command: zone("command", id, command),
    commander_damage: {},
    life_history: [],
  };
}

interface SnapOpts {
  hand?: CardView[];
  command?: CardView[];
  exile?: CardView[];
  priority?: number;
  splitSecond?: boolean;
  moves?: LegalMoveView[];
}

function snap(o: SnapOpts = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, 0, o.hand, o.command), seat(THEM, 1)],
    battlefield: zone("battlefield", ""),
    stack: zone("stack", ""),
    exile: zone("exile", "", o.exile ?? []),
    turn: {
      seq: 1,
      number: 1,
      active_seat: 0,
      priority_holder: o.priority ?? 0,
      phase: "precombat_main",
      step: "precombat_main",
    },
    split_second_active: o.splitSecond,
    // The seat owes a decision and the server enumerated it: only a pass.
    // A Craw Wurm with no lands is not castable, which is the whole point.
    legal_moves: o.moves ?? [{ type: "pass_priority", player: ME, kind: "pass", label: "Pass" }],
  } as unknown as GameView;
}

function card(name: string, typeLine: string, extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: name.toLowerCase().replace(/\W+/g, "-"),
    name,
    owner: ME,
    controller: ME,
    type_line: typeLine,
    mana_cost: "{4}{G}{G}",
    ...extras,
  } as CardView;
}

const wurm = () => card("Craw Wurm", "Creature — Wurm");

describe("castAnywayBlocked — greyed only for what is not mana (ADR 0118 §2)", () => {
  it("is live on a card the board cannot pay for", () => {
    const w = wurm();
    const s = snap({ hand: [w] });
    // The hand dims it: the move list has no cast for it.
    expect(canCastFromHand(w, s, ME).legal).toBe(false);
    // The row is still offered, live: owner decision 2.
    expect(castAnywayBlocked(w, s, ME)).toBe("");
  });

  it("is live on a card the board CAN pay for, too", () => {
    const w = wurm();
    const s = snap({
      hand: [w],
      moves: [
        { type: "cast_spell", player: ME, kind: "cast", label: "Cast", source: w.instance_id },
      ],
    });
    expect(castAnywayBlocked(w, s, ME)).toBe("");
  });

  const cases: Array<[string, () => [CardView, GameView, CastAnywayZone?], string]> = [
    ["a spectator", () => [wurm(), snap()], "Spectator can't cast"],
    ["no priority", () => [wurm(), snap({ priority: 1 })], "Not your priority"],
    ["split second", () => [wurm(), snap({ splitSecond: true })], "Split second on the stack"],
    [
      "a land",
      () => [card("Forest", "Basic Land — Forest", { mana_cost: "" }), snap()],
      "A land is played, not cast",
    ],
    [
      "cant_cast",
      () => [
        card("Bolt", "Instant", { mana_cost: "{R}", cant_cast: "Can't cast spells this turn" }),
        snap(),
      ],
      "Can't cast spells this turn",
    ],
    [
      "no legal target",
      () => [
        card("Doom Blade", "Instant", {
          mana_cost: "{1}{B}",
          legal_targets: { players: [], cards: [] },
        }),
        snap(),
      ],
      "No legal target",
    ],
    [
      "no castable mode",
      () => [
        card("Charm", "Instant", {
          mana_cost: "{G}",
          modes: {
            min: 1,
            max: 1,
            options: [{ label: "A", target_mode: "permanent", legal_targets: { cards: [] } }],
          },
        } as unknown as Partial<CardView>),
        snap(),
      ],
      "No castable mode",
    ],
    [
      "nothing to discard",
      () => {
        const c = card("Thrill", "Sorcery", {
          mana_cost: "{1}{R}",
          additional_cost: { discard_cards: 1 },
        });
        return [c, snap({ hand: [c] })];
      },
      "No card to discard",
    ],
    [
      "nothing to sacrifice",
      () => [
        card("Village Rites", "Instant", {
          mana_cost: "{B}",
          additional_cost: { sacrifice_options: { cards: [] } },
        }),
        snap(),
      ],
      "Nothing to sacrifice",
    ],
    [
      "no payable either/or branch",
      () => [
        card("Either", "Sorcery", {
          mana_cost: "{B}",
          additional_cost: { branches: [{ label: "Sacrifice", payable: false }] },
        } as unknown as Partial<CardView>),
        snap(),
      ],
      "No additional cost you can pay",
    ],
    [
      "no mana cost and no alternative cost (CR 118.6)",
      () => [card("Ancestral Vision", "Sorcery", { mana_cost: "" }), snap()],
      "It has no mana cost, so it can't be cast",
    ],
    [
      "an exile card the server says is not castable from here",
      () => {
        const c = card("Impulse Bolt", "Instant", { mana_cost: "{R}", castable_here: false });
        return [c, snap({ exile: [c] }), "exile"];
      },
      "Not castable from exile right now",
    ],
  ];

  for (const [name, build, want] of cases) {
    it(`greys on ${name}`, () => {
      const [c, s, z] = build();
      const viewer = name === "a spectator" ? null : ME;
      expect(castAnywayBlocked(c, s, viewer, z ?? "hand")).toBe(want);
    });
  }

  it("an alternative cost rescues a card with no mana cost (CR 118.6a)", () => {
    const c = card("Ancestral Vision", "Sorcery", {
      mana_cost: "",
      alternative_costs: [{ key: "granted-free", label: "Free", mana_cost: "" }],
    } as unknown as Partial<CardView>);
    expect(castAnywayBlocked(c, snap(), ME)).toBe("");
  });

  it("an exile card castable from here is live, payable or not", () => {
    const c = card("Impulse Bolt", "Instant", { mana_cost: "{R}", castable_here: true });
    expect(castAnywayBlocked(c, snap({ exile: [c] }), ME, "exile")).toBe("");
  });
});

describe("castAnywayOffered", () => {
  it("offers on a spell and not on a card that can only be played as a land", () => {
    expect(castAnywayOffered(wurm())).toBe(true);
    expect(castAnywayOffered(card("Forest", "Basic Land — Forest", { mana_cost: "" }))).toBe(false);
  });

  it("offers on a modal DFC with a spell face", () => {
    const mdfc = card("Sea Gate Restoration", "Sorcery", {
      mana_cost: "{4}{U}{U}{U}",
      layout: "modal_dfc",
      faces: [
        { name: "Sea Gate Restoration", type_line: "Sorcery", mana_cost: "{4}{U}{U}{U}" },
        { name: "Sea Gate, Reborn", type_line: "Land", mana_cost: "" },
      ],
    } as unknown as Partial<CardView>);
    expect(castAnywayOffered(mdfc)).toBe(true);
  });
});

describe("castAnywayItem — the row's builder", () => {
  it("is the named row, live, with its title", () => {
    const item = castAnywayItem("");
    expect(item.label).toBe("Cast anyway (don't pay)");
    expect(item.label).toBe(CAST_ANYWAY_LABEL);
    expect(item.disabled).toBeFalsy();
    expect(item.hint).toBe(CAST_ANYWAY_TITLE);
    expect(item.hint).toBe("Cast it without paying its mana cost. The game log shows the table.");
    expect(item.castAnyway).toBe("hand");
    // It sends nothing itself: no action rides it.
    expect(item.action).toBeUndefined();
  });

  it("greys with the reason", () => {
    const item = castAnywayItem("No legal target", "exile");
    expect(item.disabled).toBe(true);
    expect(item.hint).toBe("No legal target");
    expect(item.castAnyway).toBe("exile");
  });
});

describe("the row opens the confirmation and sends nothing", () => {
  it("requestCastAnyway only sets the pending question", () => {
    const w = wurm();
    requestCastAnyway(w, "hand");
    expect(get(castAnywayPending)).toEqual({ card: w, zone: "hand" });
    clearCastAnyway();
    expect(get(castAnywayPending)).toBeNull();
  });

  it("an exile grant's face rides along", () => {
    const w = wurm();
    requestCastAnyway(w, "exile", 1);
    expect(get(castAnywayPending)).toEqual({ card: w, zone: "exile", face: 1 });
  });
});

describe("castAnywayConfirmRequest (owner decision 6)", () => {
  it("names the dialog by the question, with Cast and Cancel", () => {
    const h = { onCast: vi.fn(), onCancel: vi.fn() };
    const r = castAnywayConfirmRequest("Craw Wurm", h);
    expect(r.label).toBe("Cast Craw Wurm without paying its mana cost?");
    expect(castAnywayConfirmLabel("Craw Wurm")).toBe(r.label);
    expect(r.question).toBe(r.label);
    expect(r.rank).toBe("flow");
    expect(takesBar(r)).toBe(true);
    expect(r.primary?.label).toBe("Cast");
    expect(r.secondary?.map((a) => a.label)).toEqual(["Cancel"]);
  });

  it("gives Cancel Escape and gives Cast no key", () => {
    const h = { onCast: vi.fn(), onCancel: vi.fn() };
    const r = castAnywayConfirmRequest("Craw Wurm", h);
    expect(r.primary?.keyShortcuts).toBeUndefined();
    expect(r.primary?.chord).toBeUndefined();
    expect(dockKeyAction(r, "Enter")).toBeNull();
    const esc = dockKeyAction(r, "Escape");
    expect(esc?.label).toBe("Cancel");
    esc!.onPress();
    expect(h.onCancel).toHaveBeenCalledTimes(1);
    expect(h.onCast).not.toHaveBeenCalled();
  });
});

describe("a confirmed Cast anyway on the wire (CastChoices.forceCast)", () => {
  function payload(choices: Parameters<typeof applyCastChoices>[1]): Record<string, unknown> {
    const params: Record<string, unknown> = { instance_id: "spell" };
    applyCastChoices(params, choices);
    return params;
  }

  it("starts with the flag, beside the zone", () => {
    expect(castChoicesBase(undefined, false, true)).toEqual({ forceCast: true });
    expect(castChoicesBase("exile", false, true)).toEqual({ fromZone: "exile", forceCast: true });
  });

  it("sends strict and force_cast, and no auto_tap", () => {
    expect(payload(castChoicesBase(undefined, false, true))).toEqual({
      instance_id: "spell",
      strict: true,
      force_cast: true,
    });
  });

  it("keeps X, the zone and the face beside it", () => {
    expect(payload({ ...castChoicesBase("command", false, true), xValue: 3, face: 1 })).toEqual({
      instance_id: "spell",
      x_value: 3,
      face: 1,
      from_zone: "command",
      strict: true,
      force_cast: true,
    });
  });

  it("survives a targeted spell's picker", () => {
    const bolt = {
      instance_id: "spell",
      name: "Bolt",
      owner: ME,
      controller: ME,
      legal_targets: { players: ["opp"], cards: [] },
    } as CardView;
    begin(bolt, "any", { ...castChoicesBase(undefined, false, true), xValue: 0 });
    let t = get(targeting)!;
    t = togglePick(t, { kind: "player", id: "opp" });
    const params: Record<string, unknown> = {
      instance_id: t.card.instance_id,
      targets: allPicks(t),
    };
    applyCastChoices(params, t.choices);
    expect(params).toMatchObject({
      instance_id: "spell",
      targets: [{ kind: "player", id: "opp" }],
      x_value: 0,
      strict: true,
      force_cast: true,
    });
    expect(params.auto_tap).toBeUndefined();
  });

  it("survives a modal spell's targeted mode", () => {
    const command = {
      instance_id: "spell",
      name: "Command",
      owner: ME,
      controller: ME,
      modes: {
        min: 1,
        max: 1,
        options: [
          {
            label: "Destroy target artifact",
            target_mode: "permanent",
            legal_targets: { players: [], cards: ["rock"] },
          },
        ],
      },
    } as unknown as CardView;
    beginForModes(command, [0], castChoicesBase(undefined, false, true));
    expect(get(targeting)?.choices?.forceCast).toBe(true);
  });

  it("is left alone by the strictMana stamp", () => {
    const forced = payload(castChoicesBase(undefined, false, true));
    expect(stampManaEnforcement("cast_spell", forced, false)).toEqual(forced);
    expect(stampManaEnforcement("cast_spell", forced, true)).toEqual(forced);
  });
});

describe("the admin override menu's cast section", () => {
  const sections = (view: GameView, c: CardView, strict: boolean) =>
    buildMenuSections(view, c, ME, false, undefined, undefined, strict);

  it("puts the row in a cast section above move to, for a hand card", () => {
    const w = wurm();
    const s = sections(snap({ hand: [w] }), w, true);
    const ids = s.map((x) => x.id);
    expect(ids).toContain("cast");
    expect(ids.indexOf("cast")).toBe(ids.indexOf("move") - 1);
    const row = s.find((x) => x.id === "cast")!.items[0];
    expect(row.label).toBe("Cast anyway (don't pay)");
    expect(row.castAnyway).toBe("hand");
    expect(row.disabled).toBeFalsy();
  });

  it("offers none with strict off, on a land, or on another seat's card", () => {
    const w = wurm();
    expect(sections(snap({ hand: [w] }), w, false).map((x) => x.id)).not.toContain("cast");
    const forest = card("Forest", "Basic Land — Forest", { mana_cost: "" });
    expect(sections(snap({ hand: [forest] }), forest, true).map((x) => x.id)).not.toContain("cast");
    const theirs = { ...wurm(), owner: THEM, controller: THEM };
    const view = snap();
    view.seats[1].hand = zone("hand", THEM, [theirs]);
    expect(
      buildMenuSections(view, theirs, ME, true, undefined, undefined, true).map((x) => x.id),
    ).not.toContain("cast");
  });

  it("offers it on the viewer's commander, out of the command zone", () => {
    const cmdr = card("Kenrith", "Legendary Creature — Human Noble", {
      mana_cost: "{4}{W}",
      is_commander: true,
    });
    const s = sections(snap({ command: [cmdr] }), cmdr, true);
    const row = s.find((x) => x.id === "cast")?.items[0];
    expect(row?.castAnyway).toBe("command");
  });
});

describe("the strip's greying", () => {
  it("an exile grant waiting on a later turn says so first", () => {
    const c = card("Warped Wurm", "Creature — Wurm", {
      owner: THEM,
      controller: THEM,
      exile_play: { player: ME, from_turn: 9 },
    } as unknown as Partial<CardView>);
    const view = snap({ exile: [c] });
    const entry = castStripEntries(view, ME).find((e) => e.card.instance_id === c.instance_id);
    expect(entry).toBeDefined();
    expect(entry!.state).not.toBe("now");
    expect(castStripCastAnywayBlocked(entry!, view, ME)).not.toBe("");
  });

  it("a commander is live, payable or not", () => {
    const cmdr = card("Kenrith", "Legendary Creature — Human Noble", { mana_cost: "{4}{W}" });
    const view = snap({ command: [cmdr] });
    const entry = castStripEntries(view, ME)[0];
    expect(entry.zone).toBe("command");
    expect(castStripCastAnywayBlocked(entry, view, ME)).toBe("");
  });
});
