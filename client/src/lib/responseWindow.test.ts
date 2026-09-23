import { describe, it, expect } from "vitest";

import {
  ALL_RESPONSES,
  classifyMove,
  hasPlay,
  hasResponse,
  inSorceryWindow,
  keyWindow,
  type ResponseCategories,
} from "./responseWindow";
import type {
  CardView,
  GameView,
  LegalMoveView,
  PlayerView,
  StackItemView,
  ZoneView,
} from "./protocol";

function zone(kind: string, owner = "", cards: CardView[] = []): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, idx: number): PlayerView {
  return {
    id,
    name: `seat ${idx}`,
    seat: idx,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
  };
}

function card(id: string, extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: id,
    name: id,
    owner: "p0",
    controller: "p0",
    type_line: "Creature",
    ...extras,
  };
}

function move(kind: LegalMoveView["kind"], extras: Partial<LegalMoveView> = {}): LegalMoveView {
  return { type: "x", player: "p0", kind, label: kind, source: `c-${kind}`, ...extras };
}

const pass = move("pass");

interface Opts {
  step?: string;
  active?: number;
  holder?: number;
  moves?: LegalMoveView[];
  stackItems?: StackItemView[];
  battlefield?: CardView[];
}

// The viewer is p0 (seat 0). Defaults: p0's own precombat main, p0
// holds priority, empty stack.
function snap(o: Opts = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seat("p0", 0), seat("p1", 1)],
    battlefield: zone("battlefield", "", o.battlefield ?? []),
    stack: zone("stack"),
    exile: zone("exile"),
    turn: {
      number: 1,
      active_seat: o.active ?? 0,
      priority_holder: o.holder ?? 0,
      phase: "x",
      step: o.step ?? "precombat_main",
    },
    mulligans_open: false,
    stack_items: o.stackItems ?? [],
    split_second_active: false,
    legal_moves: o.moves,
  };
}

function stackItem(controller: string): StackItemView {
  return {
    id: `s-${controller}`,
    kind: "spell",
    controller,
    owner: controller,
    source_card_id: `c-${controller}`,
  };
}

const only = (k: keyof ResponseCategories): ResponseCategories => ({
  counter: false,
  instant: false,
  ability: false,
  special: false,
  [k]: true,
});

describe("classifyMove", () => {
  it("pass and mana are never anything", () => {
    for (const sw of [true, false]) {
      expect(classifyMove(pass, sw)).toBe("none");
      expect(classifyMove(move("mana"), sw)).toBe("none");
    }
  });

  it("a land is always a play", () => {
    expect(classifyMove(move("land"), true)).toBe("play");
    expect(classifyMove(move("land"), false)).toBe("play");
  });

  it("a cast or activation in the sorcery window is a play", () => {
    expect(classifyMove(move("cast"), true)).toBe("play");
    expect(classifyMove(move("activate"), true)).toBe("play");
  });

  it("a stack-targeting cast or activation is a counter", () => {
    expect(classifyMove(move("cast", { targets_stack: true }), false)).toBe("counter");
    expect(classifyMove(move("activate", { targets_stack: true }), false)).toBe("counter");
  });

  it("any other cast is an instant, any other activation an ability", () => {
    expect(classifyMove(move("cast"), false)).toBe("instant");
    expect(classifyMove(move("activate"), false)).toBe("ability");
  });

  it("an older server with no targets_stack reads a counterspell as an instant", () => {
    expect(classifyMove(move("cast", { targets_stack: undefined }), false)).toBe("instant");
  });

  it("a special action is special in any window", () => {
    expect(classifyMove(move("special_action"), false)).toBe("special");
    expect(classifyMove(move("special_action"), true)).toBe("special");
  });

  it("attacks and blocks are declarations", () => {
    expect(classifyMove(move("attack"), false)).toBe("declaration");
    expect(classifyMove(move("block"), false)).toBe("declaration");
  });

  it("choices and mulligans are other", () => {
    expect(classifyMove(move("choice"), false)).toBe("other");
    expect(classifyMove(move("mulligan"), false)).toBe("other");
  });
});

describe("inSorceryWindow", () => {
  it("is the viewer's own main phase with an empty stack", () => {
    expect(inSorceryWindow(snap(), "p0")).toBe(true);
    expect(inSorceryWindow(snap({ step: "postcombat_main" }), "p0")).toBe(true);
  });

  it("is not an opponent's main, a non-main step, or a live stack", () => {
    expect(inSorceryWindow(snap({ active: 1 }), "p0")).toBe(false);
    expect(inSorceryWindow(snap({ step: "upkeep" }), "p0")).toBe(false);
    expect(inSorceryWindow(snap({ stackItems: [stackItem("p1")] }), "p0")).toBe(false);
  });
});

describe("hasResponse", () => {
  // An opponent's upkeep, where p0 holds priority: nothing is a play.
  const opp = (moves?: LegalMoveView[], extra: Opts = {}) =>
    snap({ step: "upkeep", active: 1, holder: 0, moves, ...extra });

  it("a mana-only board has no response", () => {
    expect(hasResponse(opp([pass, move("mana"), move("mana")]), "p0", ALL_RESPONSES)).toBe(false);
  });

  it("a land is never a response", () => {
    expect(hasResponse(opp([pass, move("land")]), "p0", ALL_RESPONSES)).toBe(false);
  });

  it("each category counts when enabled and not when disabled", () => {
    const rows: [LegalMoveView, keyof ResponseCategories][] = [
      [move("cast", { targets_stack: true }), "counter"],
      [move("cast"), "instant"],
      [move("activate"), "ability"],
      [move("special_action"), "special"],
    ];
    for (const [m, k] of rows) {
      const s = opp([pass, m]);
      expect(hasResponse(s, "p0", only(k)), `${k} on`).toBe(true);
      expect(hasResponse(s, "p0", { ...ALL_RESPONSES, [k]: false }), `${k} off`).toBe(false);
    }
  });

  it("a sorcery-speed cast in the viewer's own main is not a response", () => {
    expect(hasResponse(snap({ moves: [pass, move("cast")] }), "p0", ALL_RESPONSES)).toBe(false);
  });

  it("declarations and choices are not responses", () => {
    const s = opp([pass, move("attack"), move("block"), move("choice")]);
    expect(hasResponse(s, "p0", ALL_RESPONSES)).toBe(false);
  });

  it("is false without priority, and true on a missing move list with priority", () => {
    expect(hasResponse(opp([pass, move("cast")], { holder: 1 }), "p0", ALL_RESPONSES)).toBe(false);
    expect(hasResponse(opp(undefined), "p0", ALL_RESPONSES)).toBe(true);
    expect(hasResponse(opp(undefined, { holder: 1 }), "p0", ALL_RESPONSES)).toBe(false);
  });

  it("is false for a spectator or no snapshot", () => {
    expect(hasResponse(opp([pass, move("cast")]), null, ALL_RESPONSES)).toBe(false);
    expect(hasResponse(null, "p0", ALL_RESPONSES)).toBe(false);
  });
});

describe("hasPlay", () => {
  it("a land is a play in the viewer's own main phase", () => {
    expect(hasPlay(snap({ moves: [pass, move("land")] }), "p0", ALL_RESPONSES)).toBe(true);
  });

  it("a sorcery-speed cast is a play even with every response category off", () => {
    const none = { counter: false, instant: false, ability: false, special: false };
    expect(hasPlay(snap({ moves: [pass, move("cast")] }), "p0", none)).toBe(true);
  });

  it("an instant off-window follows its category", () => {
    const s = snap({ step: "upkeep", active: 1, moves: [pass, move("cast")] });
    expect(hasPlay(s, "p0", ALL_RESPONSES)).toBe(true);
    expect(hasPlay(s, "p0", { ...ALL_RESPONSES, instant: false })).toBe(false);
  });

  it("mana never counts", () => {
    expect(hasPlay(snap({ moves: [pass, move("mana")] }), "p0", ALL_RESPONSES)).toBe(false);
  });
});

describe("keyWindow", () => {
  it("stackOpp: an opponent's item on the stack", () => {
    const s = snap({ step: "upkeep", active: 1, stackItems: [stackItem("p1")] });
    expect(keyWindow(s, "p0")).toEqual({ stackOpp: true, combat: false, oppEnd: false });
  });

  it("stackOpp is false for a stack that is all the viewer's own, or empty", () => {
    expect(keyWindow(snap({ stackItems: [stackItem("p0")] }), "p0").stackOpp).toBe(false);
    expect(keyWindow(snap(), "p0").stackOpp).toBe(false);
  });

  it("stackOpp is true for a mixed stack", () => {
    const s = snap({ stackItems: [stackItem("p0"), stackItem("p1")] });
    expect(keyWindow(s, "p0").stackOpp).toBe(true);
  });

  it("combat: declare attackers or blockers with an attacker on the board", () => {
    const attacking = [card("bear", { controller: "p1", attacking_target: "p0" })];
    for (const step of ["declare_attackers", "declare_blockers"]) {
      expect(keyWindow(snap({ step, active: 1, battlefield: attacking }), "p0").combat).toBe(true);
    }
    expect(keyWindow(snap({ step: "combat_damage", battlefield: attacking }), "p0").combat).toBe(
      false,
    );
    expect(
      keyWindow(snap({ step: "declare_attackers", battlefield: [card("idle")] }), "p0").combat,
    ).toBe(false);
  });

  it("oppEnd: an opponent's end step, not the viewer's own", () => {
    expect(keyWindow(snap({ step: "end", active: 1 }), "p0").oppEnd).toBe(true);
    expect(keyWindow(snap({ step: "end", active: 0 }), "p0").oppEnd).toBe(false);
  });

  it("is all false for a spectator", () => {
    const s = snap({ step: "end", active: 1, stackItems: [stackItem("p1")] });
    expect(keyWindow(s, null)).toEqual({ stackOpp: false, combat: false, oppEnd: false });
  });
});
