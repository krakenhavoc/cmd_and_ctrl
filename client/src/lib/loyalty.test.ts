import { describe, expect, it } from "vitest";
import { get } from "svelte/store";
import type { ActionType, CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { battlefieldClickIntent, buildMenuSections, type MenuSection } from "./contextMenu.logic";
import { beginForAbility, canConfirm, isMultiPick, targeting } from "./targeting";
import { canActivateLoyalty } from "./timing";

// loyalty.test.ts — issues #329 and #334.
//
// #329: "I cast teferi and when I click on him to choose one of his
// abilities it just tapped him. I right clicked him and nothing but
// the normal admin abilities showed up."
// #334: "Planeswalker loyalty abilities are not able to be
// activated."
//
// Both replays confirm the same board state: a Teferi with four
// loyalty counters, `activated_abilities: []`, and a `tapped` bit
// flipping true / false across consecutive snapshots as the player
// clicked him. So there are two client-side failures to pin — the
// click routes to tap, and the menu has nothing to show.

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

function walker(extra: Partial<CardView> = {}): CardView {
  return {
    instance_id: "pw1",
    name: "Teferi, Time Raveler",
    owner: "a",
    controller: "a",
    type_line: "Legendary Planeswalker — Teferi",
    counters: { loyalty: 4 },
    activated_abilities: [
      { index: 0, label: "+1: …", loyalty_cost: 1, sorcery_speed: true },
      { index: 1, label: "−3: …", loyalty_cost: -3, sorcery_speed: true },
    ],
    ...extra,
  };
}

// closedWalker is the same planeswalker with the server's #1208
// verdict on its rows. Since #1208 the client does not derive the
// window for a catalogued ability — `timing_closed` is the stamp of
// game.ActivationTimingOpenLocked — so a fixture that wants a shut
// window has to say so the way the wire does.
function closedWalker(extra: Partial<CardView> = {}): CardView {
  const pw = walker(extra);
  return {
    ...pw,
    activated_abilities: (pw.activated_abilities ?? []).map((a) => ({
      ...a,
      timing_closed: true,
    })),
  };
}

interface ViewOpts {
  battlefield?: CardView[];
  stack?: CardView[];
  step?: string;
  priority?: number;
  active?: number;
}

function view(opts: ViewOpts = {}): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat("a", "Alice"), seat("b", "Bob")],
    battlefield: zone("battlefield", undefined, opts.battlefield ?? []),
    stack: zone("stack", undefined, opts.stack ?? []),
    exile: zone("exile", undefined, []),
    turn: {
      number: 1,
      active_seat: opts.active ?? 0,
      priority_holder: opts.priority ?? 0,
      phase: "main1",
      step: opts.step ?? "precombat_main",
    },
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

// --- the wire ----------------------------------------------------

describe("activate_loyalty on the wire", () => {
  it("is a member of the ActionType union", () => {
    // A compile-time assertion first: before #334 this line was a
    // TypeScript error, so no client code could send the action the
    // server had been exposing since S13.1.
    const t: ActionType = "activate_loyalty";
    expect(t).toBe("activate_loyalty");
  });
});

// --- #329: the click ---------------------------------------------

describe("battlefieldClickIntent", () => {
  it("offers a planeswalker's abilities instead of tapping it", () => {
    expect(battlefieldClickIntent(walker(), "a", false)).toBe("abilities");
  });

  it("still offers the menu for a planeswalker with no catalog abilities", () => {
    // #329's actual card was Teferi, Who Slows the Sunset — no
    // catalog entry, so no activated abilities. Tapping is still
    // the wrong answer: the menu carries the manual loyalty rows.
    const bare = walker({ activated_abilities: undefined });
    expect(battlefieldClickIntent(bare, "a", false)).toBe("abilities");
  });

  it("taps everything that is not a planeswalker", () => {
    const bear: CardView = {
      instance_id: "c1",
      name: "Bear",
      owner: "a",
      controller: "a",
      type_line: "Creature — Bear",
    };
    expect(battlefieldClickIntent(bear, "a", false)).toBe("tap");
  });

  it("does nothing for a card the viewer neither controls nor admins", () => {
    expect(battlefieldClickIntent(walker({ controller: "b" }), "a", false)).toBe("none");
  });

  it("lets an admin drive someone else's planeswalker", () => {
    expect(battlefieldClickIntent(walker({ controller: "b" }), "a", true)).toBe("abilities");
  });

  // --- #368: the same shape on a utility land ----------------------

  it("offers a fetchland's abilities instead of tapping it", () => {
    const passage: CardView = {
      instance_id: "l1",
      name: "Fabled Passage",
      owner: "a",
      controller: "a",
      type_line: "Land",
      activated_abilities: [
        { index: 0, label: "{T}, Sacrifice this land: Search for a basic land" },
      ],
    };
    expect(battlefieldClickIntent(passage, "a", false)).toBe("abilities");
  });

  it("still taps a plain land, which has mana abilities and nothing else", () => {
    const forest: CardView = {
      instance_id: "l2",
      name: "Forest",
      owner: "a",
      controller: "a",
      type_line: "Basic Land — Forest",
      mana_abilities: [{ index: 0, label: "{T}: Add {G}", tap_cost: true }],
    };
    expect(battlefieldClickIntent(forest, "a", false)).toBe("tap");
  });

  it("still taps an animated manland so it stays combat-selectable", () => {
    const manland: CardView = {
      instance_id: "l3",
      name: "Celestial Colonnade",
      owner: "a",
      controller: "a",
      type_line: "Creature Land — Elemental",
      activated_abilities: [{ index: 0, label: "{3}{W}{U}: becomes a 4/4" }],
    };
    expect(battlefieldClickIntent(manland, "a", false)).toBe("tap");
  });
});

// --- #334: the menu ----------------------------------------------

describe("loyalty abilities in the card menu", () => {
  it("lists each loyalty ability as its own row", () => {
    const pw = walker();
    const sections = buildMenuSections(view({ battlefield: [pw] }), pw, "a", false);
    expect(itemIn(sections, "ability-0")?.label).toBe("+1: …");
    expect(itemIn(sections, "ability-1")?.label).toBe("−3: …");
    expect(itemIn(sections, "ability-0")?.disabled).toBeFalsy();
  });

  it("greys a −N the planeswalker cannot pay for (CR 606.6)", () => {
    const pw = walker({ counters: { loyalty: 2 } });
    const sections = buildMenuSections(view({ battlefield: [pw] }), pw, "a", false);
    expect(itemIn(sections, "ability-1")?.disabled).toBe(true);
    expect(itemIn(sections, "ability-1")?.hint).toMatch(/loyalty/i);
    // The +1 is unaffected — it costs nothing to pay.
    expect(itemIn(sections, "ability-0")?.disabled).toBeFalsy();
  });

  it("greys every loyalty ability once one has been activated this turn", () => {
    const pw = walker({ loyalty_activated: true });
    const sections = buildMenuSections(view({ battlefield: [pw] }), pw, "a", false);
    expect(itemIn(sections, "ability-0")?.disabled).toBe(true);
    expect(itemIn(sections, "ability-1")?.disabled).toBe(true);
  });

  it("greys loyalty abilities outside the sorcery-speed window (CR 606.3)", () => {
    const pw = closedWalker();
    const v = view({ battlefield: [pw], step: "upkeep" });
    const sections = buildMenuSections(v, pw, "a", false);
    expect(itemIn(sections, "ability-0")?.disabled).toBe(true);
  });

  // #1208, The Wandering Emperor: "As long as she entered this turn,
  // you may activate her loyalty abilities any time you could cast an
  // instant." CR 606.3 has no printed clause behind it, so this is
  // the half sorcery_speed could never have carried — the row is live
  // in an upkeep step because the SERVER says so.
  it("leaves a loyalty row live when the server says the window is open", () => {
    const pw = walker();
    const v = view({ battlefield: [pw], step: "upkeep" });
    const sections = buildMenuSections(v, pw, "a", false);
    expect(itemIn(sections, "ability-0")?.disabled).toBeFalsy();
  });

  // The once-per-turn half of CR 606.3 is not timing and stays the
  // client's to read off the card, so an open window does not undo it.
  it("still greys an already-activated walker when the window is open", () => {
    const pw = walker({ loyalty_activated: true });
    const v = view({ battlefield: [pw], step: "upkeep" });
    const sections = buildMenuSections(v, pw, "a", false);
    expect(itemIn(sections, "ability-0")?.hint).toBe("Already activated this turn");
  });

  it("leaves an instant-speed ability alone outside the main phase", () => {
    const rock: CardView = {
      instance_id: "r1",
      name: "Rock",
      owner: "a",
      controller: "a",
      type_line: "Artifact",
      activated_abilities: [{ index: 0, label: "{T}: do a thing" }],
    };
    const v = view({ battlefield: [rock], step: "upkeep" });
    const sections = buildMenuSections(v, rock, "a", false);
    expect(itemIn(sections, "ability-0")?.disabled).toBeFalsy();
  });
});

// --- #1157: "up to one target" is satisfied by no target ----------
//
// "[in-app] Aetherspark abilities are not functioning — Aetherspark
// shows it has loyalty abilities but they are non-functioning (all
// grayed out)."
//
// The Aetherspark's +1 reads "Attach The Aetherspark to up to one
// target creature you control. Put a +1/+1 counter on that creature."
// — min 0, and declining is a complete, legal activation that still
// ticks the loyalty up. The server agrees twice over:
// ActivateCatalogAbility accepts the activation with no targets, and
// internal/legal enumerates it as a move for a bot. Only the human's
// menu refused it, because abilityBlocked asked "is the candidate
// list empty" instead of "does it hold `min` candidates" — so on a
// board with no creature to attach to, the one ability a 4-loyalty
// Aetherspark can actually pay for was greyed, and its −5 and −10
// were correctly greyed beside it. All three greyed out.
describe("an 'up to N' target clause with nothing to point at (#1157)", () => {
  function aetherspark(targets: { cards?: string[]; min?: number }): CardView {
    return {
      instance_id: "spark",
      name: "The Aetherspark",
      owner: "a",
      controller: "a",
      type_line: "Legendary Artifact Planeswalker — Equipment",
      counters: { loyalty: 4 },
      activated_abilities: [
        {
          index: 0,
          label: "+1: Attach The Aetherspark to up to one target creature you control.",
          loyalty_cost: 1,
          sorcery_speed: true,
          legal_targets: { cards: targets.cards ?? [], min: targets.min ?? 0, max: 1 },
        },
        { index: 1, label: "−5: Draw two cards.", loyalty_cost: -5, sorcery_speed: true },
      ],
    };
  }

  it("offers the +1 with no legal creature on the board", () => {
    const spark = aetherspark({ cards: [], min: 0 });
    const item = itemIn(
      buildMenuSections(view({ battlefield: [spark] }), spark, "a", false),
      "ability-0",
    );
    expect(item?.disabled).toBeFalsy();
    expect(item?.hint).toBeUndefined();
  });

  it("offers the +1 when a creature IS there", () => {
    const spark = aetherspark({ cards: ["bear"], min: 0 });
    const item = itemIn(
      buildMenuSections(view({ battlefield: [spark] }), spark, "a", false),
      "ability-0",
    );
    expect(item?.disabled).toBeFalsy();
  });

  it("still greys a REQUIRED target clause with nothing to point at (CR 601.2c)", () => {
    const spark = aetherspark({ cards: [], min: 1 });
    const item = itemIn(
      buildMenuSections(view({ battlefield: [spark] }), spark, "a", false),
      "ability-0",
    );
    expect(item?.disabled).toBe(true);
    expect(item?.hint).toBe("no legal target");
  });

  it("leaves the −5 correctly greyed at four loyalty (CR 606.6)", () => {
    const spark = aetherspark({ cards: [], min: 0 });
    const item = itemIn(
      buildMenuSections(view({ battlefield: [spark] }), spark, "a", false),
      "ability-1",
    );
    expect(item?.disabled).toBe(true);
    expect(item?.hint).toMatch(/loyalty/i);
  });
});

// --- S31: the sorcery_speed flag nothing read ---------------------
//
// `ActivatedAbilityView.sorcery_speed` has ridden the wire since S21
// and no client surface consulted it, so an "activate only as a
// sorcery" ability stayed live through combat and an opponent's turn
// and came back rejected by the server. ADR 0033 §1 cites it as the
// live specimen of the bug class the legal-move enumerator exists to
// end; S31 sub-PR 2 fixes it in both surfaces that render an ability
// row — the context menu here, and ManaAbilityMenu's popover.
describe("sorcery-speed activated abilities (not loyalty)", () => {
  function sorcerySpeedRock(extra: Partial<CardView> = {}): CardView {
    return {
      instance_id: "r1",
      name: "Thousand-Year Elixir",
      owner: "a",
      controller: "a",
      type_line: "Artifact",
      activated_abilities: [{ index: 0, label: "{2}: do a thing", sorcery_speed: true }],
      ...extra,
    };
  }

  // The same Elixir with the server's #1208 verdict on the row: since
  // #1208 `timing_closed` is what greys it, because a per-player
  // statement can open a sorcery-speed ability and the client cannot
  // see one.
  function closedRock(extra: Partial<CardView> = {}): CardView {
    return sorcerySpeedRock({
      activated_abilities: [
        { index: 0, label: "{2}: do a thing", sorcery_speed: true, timing_closed: true },
      ],
      ...extra,
    });
  }

  it("is offered inside the sorcery-speed window", () => {
    const rock = sorcerySpeedRock();
    const sections = buildMenuSections(view({ battlefield: [rock] }), rock, "a", false);
    expect(itemIn(sections, "ability-0")?.disabled).toBeFalsy();
  });

  it("is greyed outside a main phase — CR 307.1", () => {
    const rock = closedRock();
    const v = view({ battlefield: [rock], step: "upkeep" });
    const item = itemIn(buildMenuSections(v, rock, "a", false), "ability-0");
    expect(item?.disabled).toBe(true);
    expect(item?.hint).toBe("Sorcery-speed only");
  });

  it("is greyed on an opponent's turn", () => {
    const rock = closedRock();
    const v = view({ battlefield: [rock], active: 1, priority: 0 });
    const item = itemIn(buildMenuSections(v, rock, "a", false), "ability-0");
    expect(item?.disabled).toBe(true);
    expect(item?.hint).toBe("Not your turn");
  });

  it("is greyed while something is on the stack", () => {
    const rock = closedRock();
    const spell: CardView = { instance_id: "s1", name: "Spell", owner: "b", controller: "b" };
    const v = view({ battlefield: [rock], stack: [spell] });
    const item = itemIn(buildMenuSections(v, rock, "a", false), "ability-0");
    expect(item?.disabled).toBe(true);
    expect(item?.hint).toBe("Stack isn't empty");
  });

  it("does not gate an ability without the flag", () => {
    const rock = sorcerySpeedRock({
      activated_abilities: [{ index: 0, label: "{2}: do a thing" }],
    });
    const v = view({ battlefield: [rock], step: "upkeep" });
    expect(itemIn(buildMenuSections(v, rock, "a", false), "ability-0")?.disabled).toBeFalsy();
  });

  // #1208, Leonin Shikari: "You may activate equip abilities any time
  // you could cast an instant." The row still PRINTS "activate only
  // as a sorcery" — sorcery_speed stays true — and the engine accepts
  // it, so the absent timing_closed is what the client must obey.
  // Before #1208 the client re-derived CR 307.1 here and greyed a row
  // the server would have accepted.
  it("leaves a sorcery-speed row live when the server says the window is open", () => {
    const rock = sorcerySpeedRock();
    const v = view({ battlefield: [rock], step: "upkeep" });
    expect(itemIn(buildMenuSections(v, rock, "a", false), "ability-0")?.disabled).toBeFalsy();
  });

  it("does not gate mana abilities, which have no timing restriction (CR 605.1a)", () => {
    const rock = sorcerySpeedRock({
      activated_abilities: undefined,
      mana_abilities: [{ index: 0, label: "Add {C}", produced: "{C}", tap_cost: true }],
    });
    const v = view({ battlefield: [rock], step: "upkeep" });
    expect(itemIn(buildMenuSections(v, rock, "a", false), "mana-0")?.disabled).toBeFalsy();
  });
});

// --- #329's actual card: a planeswalker the catalog doesn't know --

describe("manual loyalty rows for a non-catalog planeswalker", () => {
  // #329 was Teferi, Who Slows the Sunset — four loyalty counters,
  // no catalog entry, `activated_abilities: []`.
  function bare(extra: Partial<CardView> = {}): CardView {
    return walker({
      name: "Teferi, Who Slows the Sunset",
      activated_abilities: undefined,
      ...extra,
    });
  }

  it("offers the costs the walker can actually pay", () => {
    const pw = bare();
    const sections = buildMenuSections(view({ battlefield: [pw] }), pw, "a", false);
    const ids = (sections.find((s) => s.id === "loyalty")?.items ?? []).map((i) => i.id);
    // +2 / +1 / [0] always, and −1..−4 for a walker at 4 loyalty.
    expect(ids).toEqual([
      "loyalty-2",
      "loyalty-1",
      "loyalty-0",
      "loyalty--1",
      "loyalty--2",
      "loyalty--3",
      "loyalty--4",
    ]);
  });

  it("stops the minus rows at the walker's current loyalty (CR 606.6)", () => {
    const pw = bare({ counters: { loyalty: 1 } });
    const sections = buildMenuSections(view({ battlefield: [pw] }), pw, "a", false);
    const ids = (sections.find((s) => s.id === "loyalty")?.items ?? []).map((i) => i.id);
    expect(ids).toEqual(["loyalty-2", "loyalty-1", "loyalty-0", "loyalty--1"]);
  });

  it("sends activate_loyalty, not a bare counter nudge", () => {
    const pw = bare();
    const sections = buildMenuSections(view({ battlefield: [pw] }), pw, "a", false);
    const minus = itemIn(sections, "loyalty--3");
    expect(minus?.action?.type).toBe("activate_loyalty");
    expect(minus?.action?.params).toEqual({
      planeswalker_id: "pw1",
      delta: -3,
      label: "−3",
    });
  });

  it("greys the rows outside the sorcery-speed window", () => {
    const pw = bare();
    const v = view({ battlefield: [pw], step: "upkeep" });
    const sections = buildMenuSections(v, pw, "a", false);
    expect(itemIn(sections, "loyalty-1")?.disabled).toBe(true);
  });

  it("stays out of the way once the catalog knows the card", () => {
    const pw = walker();
    const sections = buildMenuSections(view({ battlefield: [pw] }), pw, "a", false);
    expect(sections.find((s) => s.id === "loyalty")).toBeUndefined();
    expect(itemIn(sections, "ability-0")).toBeDefined();
  });

  it("does not appear on a non-planeswalker", () => {
    const bear: CardView = {
      instance_id: "c1",
      name: "Bear",
      owner: "a",
      controller: "a",
      type_line: "Creature — Bear",
    };
    const sections = buildMenuSections(view({ battlefield: [bear] }), bear, "a", false);
    expect(sections.find((s) => s.id === "loyalty")).toBeUndefined();
  });
});

// --- canActivateLoyalty, finally wired to something --------------

describe("canActivateLoyalty", () => {
  it("reads the server's once-per-turn flag off the card", () => {
    const pw = walker({ loyalty_activated: true });
    const v = view({ battlefield: [pw] });
    expect(canActivateLoyalty(pw, v, "a").legal).toBe(false);
  });

  it("is legal in the viewer's own main phase with an empty stack", () => {
    const pw = walker();
    const v = view({ battlefield: [pw] });
    expect(canActivateLoyalty(pw, v, "a").legal).toBe(true);
  });
});

// --- "up to one target" ------------------------------------------
//
// Teferi's −3 and the Emperor's −2 both print "up to one target".
// Server-side that is a Min 0 clause; the client used to pin every
// ability prompt to exactly one pick, so the player could never
// answer "none" and the draw half of Teferi's −3 was unreachable on
// a board with nothing to bounce.

describe("an ability clause that takes up to one target", () => {
  const source = walker();

  it("accumulates picks and offers Done instead of firing on the first click", () => {
    beginForAbility(source, { index: 1, legal_targets: { cards: ["c1"], min: 0, max: 1 } }, []);
    const t = get(targeting);
    expect(t).not.toBeNull();
    expect(t?.min).toBe(0);
    expect(t?.max).toBe(1);
    expect(isMultiPick(t!)).toBe(true);
    // Zero picks is a legal answer.
    expect(canConfirm(t!)).toBe(true);
    targeting.set(null);
  });

  it("still completes on the first click for an ordinary single-target ability", () => {
    beginForAbility(source, { index: 0, legal_targets: { cards: ["c1"], min: 1, max: 1 } }, []);
    const t = get(targeting);
    expect(isMultiPick(t!)).toBe(false);
    expect(canConfirm(t!)).toBe(false);
    targeting.set(null);
  });
});
