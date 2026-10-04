// engineMayMissMana.test.ts — ADR 0118 owner decision 8 (2026-10-04),
// "stop if the engine may be wrong": smart autopass holds the viewer's
// own main phase when a spell in hand is left out of the move list for
// mana alone AND the viewer controls a mana source the engine does not
// run. The gate is read from a real snapshot shape here, then fed to
// autopassDecision to show the verdict it produces.

import { describe, it, expect } from "vitest";

import {
  controlsManualManaSource,
  engineMayMissMana,
  isManualManaSource,
  unpayableSpellInHand,
} from "./engineMayMissMana";
import { autopassDecision, type AutopassGates } from "./autopassDecision";
import type { CardView, GameView, LegalMoveView, PlayerView, ZoneView } from "./protocol";

const ME = "me";
const THEM = "them";

const zone = (kind: string, owner: string, cards: CardView[] = []): ZoneView => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

function seat(id: string, idx: number, hand: CardView[] = []): PlayerView {
  return {
    id,
    name: id === ME ? "Me" : "Them",
    seat: idx,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
  };
}

function card(name: string, typeLine: string, extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: name.toLowerCase().replace(/\W+/g, "-"),
    name,
    owner: ME,
    controller: ME,
    type_line: typeLine,
    ...extras,
  } as CardView;
}

// Craw Wurm: the board cannot pay {4}{G}{G}, so the move list has no
// cast for it.
const wurm = () => card("Craw Wurm", "Creature — Wurm", { mana_cost: "{4}{G}{G}" });
// An uncatalogued nonbasic land: it prints a mana ability the engine
// does not run, so it carries `unimplemented` and no `mana_abilities`.
const manualLand = (extras: Partial<CardView> = {}) =>
  card("Ancient Den Of Mana", "Land", { unimplemented: true, ...extras });
const forest = () =>
  card("Forest", "Basic Land — Forest", {
    mana_abilities: [{ index: 0, ref: "own:0", label: "Add {G}", produced: "{G}" }],
  } as unknown as Partial<CardView>);

interface Opts {
  hand?: CardView[];
  battlefield?: CardView[];
  active?: number;
  priority?: number;
  step?: string;
  moves?: LegalMoveView[] | null;
}

function view(o: Opts = {}): GameView {
  const step = o.step ?? "precombat_main";
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, 0, o.hand ?? [wurm()]), seat(THEM, 1)],
    battlefield: zone("battlefield", "", o.battlefield ?? [manualLand()]),
    stack: zone("stack", ""),
    exile: zone("exile", ""),
    stack_items: [],
    turn: {
      seq: 1,
      number: 3,
      active_seat: o.active ?? 0,
      priority_holder: o.priority ?? o.active ?? 0,
      phase: step,
      step,
    },
    // The seat owes a decision, and the server offered only a pass.
    legal_moves:
      o.moves === null
        ? undefined
        : (o.moves ?? [{ type: "pass_priority", player: ME, kind: "pass", label: "Pass" }]),
  } as unknown as GameView;
}

describe("the gate: own main phase, a spell left out for mana, a manual mana source", () => {
  it("fires on the viewer's own main phase, precombat and postcombat", () => {
    expect(engineMayMissMana(view(), ME)).toBe(true);
    expect(engineMayMissMana(view({ step: "postcombat_main" }), ME)).toBe(true);
  });

  it("does not fire with no manual mana source", () => {
    expect(engineMayMissMana(view({ battlefield: [] }), ME)).toBe(false);
    // A land the engine runs (a Forest) is not one.
    expect(engineMayMissMana(view({ battlefield: [forest()] }), ME)).toBe(false);
    // Nor is someone else's manual land.
    expect(engineMayMissMana(view({ battlefield: [manualLand({ controller: THEM })] }), ME)).toBe(
      false,
    );
  });

  it("does not fire on another player's turn", () => {
    // The opponent's main phase, with the viewer holding priority.
    expect(engineMayMissMana(view({ active: 1, priority: 0 }), ME)).toBe(false);
  });

  it("does not fire outside a main phase", () => {
    expect(engineMayMissMana(view({ step: "beginning_of_combat" }), ME)).toBe(false);
  });

  it("does not fire when the spell is left out for something other than mana", () => {
    // `cant_cast` refuses it whatever the pool holds: the Cast anyway
    // row is greyed too, so there is nothing for the stop to show.
    const refused = card("Craw Wurm", "Creature — Wurm", {
      mana_cost: "{4}{G}{G}",
      cant_cast: "Each player can't cast more than one spell each turn",
    });
    expect(unpayableSpellInHand(view({ hand: [refused] }), ME)).toBe(false);
    // A land in hand is played, not cast.
    expect(unpayableSpellInHand(view({ hand: [forest()] }), ME)).toBe(false);
  });

  it("does not fire when the move list already offers the cast", () => {
    const w = wurm();
    const moves: LegalMoveView[] = [
      { type: "pass_priority", player: ME, kind: "pass", label: "Pass" },
      {
        type: "cast_spell",
        player: ME,
        kind: "cast",
        label: "Cast Craw Wurm",
        source: w.instance_id,
        params: { instance_id: w.instance_id },
      } as LegalMoveView,
    ];
    expect(unpayableSpellInHand(view({ hand: [w], moves }), ME)).toBe(false);
  });

  it("says no when the frame carries no move list (no information)", () => {
    expect(unpayableSpellInHand(view({ moves: null }), ME)).toBe(false);
  });
});

describe("what counts as a manual mana source (the wire's `unimplemented`)", () => {
  it("an unimplemented land, presumed to make mana", () => {
    expect(isManualManaSource(manualLand())).toBe(true);
  });

  it("an unimplemented permanent that has a mana ability on the wire", () => {
    const rock = card("Odd Rock", "Artifact", {
      unimplemented: true,
      mana_abilities: [{ index: 0, ref: "own:0", label: "Add {C}" }],
    } as unknown as Partial<CardView>);
    expect(isManualManaSource(rock)).toBe(true);
  });

  it("not an unimplemented non-land with no mana ability on the wire: the wire can't tell", () => {
    expect(isManualManaSource(card("Odd Elf", "Creature — Elf", { unimplemented: true }))).toBe(
      false,
    );
  });

  it("not a land the engine runs", () => {
    expect(isManualManaSource(forest())).toBe(false);
    expect(controlsManualManaSource(view({ battlefield: [forest()] }), ME)).toBe(false);
  });
});

describe("the verdict it produces", () => {
  // The stops grid ticks the viewer's own main phase; smart autopass is
  // on and nothing else is playable (no land, the Wurm unpayable).
  const gatesFor = (v: GameView): AutopassGates => ({
    viewerHasPriority: true,
    tableBusy: false,
    hasPendingChoice: false,
    owesBlockDecision: false,
    owesAttackRequirement: false,
    loopSuspended: false,
    step: v.turn?.step,
    autopassToggle: false,
    viewerIsActive: v.turn?.active_seat === 0,
    autopassPersistThroughTurns: false,
    manualStop: false,
    autoPassPriority: true,
    stackEmpty: true,
    holdPriority: false,
    autoPassOwnStack: true,
    ownsEveryStackItem: false,
    stepStop: true,
    smartAutoPass: true,
    alwaysStopOpponentStack: false,
    hasResponse: false,
    hasPlay: false,
    engineMayMissMana: engineMayMissMana(v, ME),
    combatWindow: false,
    oppEndWindow: false,
    bluffCounter: false,
    bluffInstant: false,
    bluffManual: false,
  });

  it("holds the main phase when the stop fires", () => {
    expect(autopassDecision(gatesFor(view()))).toBe("hold");
  });

  it("passes it, as before, with no manual mana source", () => {
    expect(autopassDecision(gatesFor(view({ battlefield: [forest()] })))).toBe("pass");
  });

  it("passes another player's main phase, as before", () => {
    expect(autopassDecision(gatesFor(view({ active: 1, priority: 0 })))).toBe("pass");
  });
});
