import { describe, it, expect } from "vitest";

import {
  ALL_RESPONSES,
  DEFAULT_RESPONSES,
  classifyMove,
  hasPlay,
  hasResponse,
  inCombatWindow,
  inSorceryWindow,
  isDefending,
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
      seq: 1,
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
  untargeted: false,
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

  it("any other cast is an instant, targeted or not", () => {
    expect(classifyMove(move("cast"), false)).toBe("instant");
    expect(classifyMove(move("cast", { has_targets: true }), false)).toBe("instant");
  });

  it("#2853: an activation is an ability when it targets, untargeted when it does not", () => {
    expect(classifyMove(move("activate", { has_targets: true }), false)).toBe("ability");
    expect(classifyMove(move("activate"), false)).toBe("untargeted");
    // Owner answer 2: an untargeted answer is an ability too.
    expect(classifyMove(move("activate", { interacts: true }), false)).toBe("ability");
    expect(classifyMove(move("activate", { interacts: true }), true)).toBe("play");
    // An older server sends neither bit.
    expect(classifyMove(move("activate", { has_targets: undefined }), false)).toBe("untargeted");
    // A counter is a counter, whatever has_targets says.
    expect(classifyMove(move("activate", { targets_stack: true, has_targets: true }), false)).toBe(
      "counter",
    );
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
      [move("activate", { has_targets: true }), "ability"],
      [move("activate"), "untargeted"],
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

// #2853, owner decision 1: an opponent's stack item stops you only for
// real interaction. The issue's board: three Forests, Llanowar Elves,
// Mind Stone and Evolving Wilds, with an opponent's spell on the stack.
describe("hasResponse — #2853's default categories on an opponent's stack item", () => {
  const onOppStack = (moves: LegalMoveView[]) =>
    snap({ step: "precombat_main", active: 1, holder: 0, moves, stackItems: [stackItem("p1")] });
  const valueBoard = [
    pass,
    move("mana", { source: "forest-1" }),
    move("mana", { source: "forest-2" }),
    move("mana", { source: "forest-3" }),
    move("mana", { source: "elves" }),
    move("mana", { source: "mind-stone" }),
    move("activate", { source: "mind-stone", label: "Mind Stone: draw a card" }),
    move("activate", { source: "evolving-wilds", label: "Evolving Wilds: search" }),
  ];

  it("the Mind Stone / Evolving Wilds board has no response by default", () => {
    expect(hasResponse(onOppStack(valueBoard), "p0", DEFAULT_RESPONSES)).toBe(false);
  });

  it("the old behaviour comes back with untargeted abilities ticked", () => {
    expect(
      hasResponse(onOppStack(valueBoard), "p0", { ...DEFAULT_RESPONSES, untargeted: true }),
    ).toBe(true);
  });

  it("cycling from hand is an untargeted activation and does not stop you", () => {
    const cycling = move("activate", { source: "hand-card", label: "Cycle Lonely Sandbar" });
    expect(hasResponse(onOppStack([pass, cycling]), "p0", DEFAULT_RESPONSES)).toBe(false);
  });

  // Owner answer 2: an untargeted ability that can answer the stack
  // still stops you. The server marks it `interacts`.
  it("a sacrifice outlet, a regeneration shield and a pump hold", () => {
    for (const [source, label] of [
      ["viscera-seer", "Viscera Seer: Sacrifice a creature: Scry 1."],
      ["undercity-troll", "Undercity Troll: {2}{G}: Regenerate this creature."],
      ["evernight-shade", "Evernight Shade: {B}: This creature gets +1/+1 until end of turn."],
    ]) {
      const m = move("activate", { source, label, interacts: true });
      expect(hasResponse(onOppStack([...valueBoard, m]), "p0", DEFAULT_RESPONSES), source).toBe(
        true,
      );
    }
  });

  it("a mana ability that is a sacrifice outlet holds; plain mana does not", () => {
    const altar = move("mana", { source: "ashnods-altar", interacts: true });
    expect(hasResponse(onOppStack([...valueBoard, altar]), "p0", DEFAULT_RESPONSES)).toBe(true);
    expect(classifyMove(altar, true)).toBe("none");
    expect(classifyMove(move("mana"), false)).toBe("none");
  });

  it("an ability that targets holds", () => {
    const ping = move("activate", { source: "pinger", has_targets: true });
    expect(hasResponse(onOppStack([...valueBoard, ping]), "p0", DEFAULT_RESPONSES)).toBe(true);
  });

  it("an instant holds, targeted or not", () => {
    for (const extras of [{}, { has_targets: true }]) {
      const inst = move("cast", { source: "instant", ...extras });
      expect(hasResponse(onOppStack([...valueBoard, inst]), "p0", DEFAULT_RESPONSES)).toBe(true);
    }
  });

  it("a counter holds, cast or activated", () => {
    for (const kind of ["cast", "activate"] as const) {
      const counter = move(kind, { source: "counter", targets_stack: true, has_targets: true });
      expect(hasResponse(onOppStack([...valueBoard, counter]), "p0", DEFAULT_RESPONSES)).toBe(true);
    }
  });
});

describe("hasPlay", () => {
  it("a land is a play in the viewer's own main phase", () => {
    expect(hasPlay(snap({ moves: [pass, move("land")] }), "p0", ALL_RESPONSES)).toBe(true);
  });

  it("a sorcery-speed cast is a play even with every response category off", () => {
    const none = {
      counter: false,
      instant: false,
      ability: false,
      untargeted: false,
      special: false,
    };
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

// ADR 0106 §1 decision 7 (owner decision 1, #1793): an "Any player may
// activate this ability" row on a permanent somebody else controls is a
// legal move at every window, so it must not hold smart autopass.
describe("an activation of another player's permanent", () => {
  const xantcha = card("xantcha", { owner: "p1", controller: "p1" });
  const mine = card("mine");
  const across = move("activate", { source: "xantcha", has_targets: true });
  const own = move("activate", { source: "mine", has_targets: true });
  const board = [xantcha, mine];

  it("classifies as none with the frame's controllers, and as before without them", () => {
    const controllers = new Map([
      ["xantcha", "p1"],
      ["mine", "p0"],
    ]);
    expect(classifyMove(across, false, controllers, "p0")).toBe("none");
    expect(classifyMove(across, true, controllers, "p0")).toBe("none");
    expect(classifyMove(own, false, controllers, "p0")).toBe("ability");
    // The existing two-argument call is unchanged.
    expect(classifyMove(across, false)).toBe("ability");
    // A source the map does not know keeps today's class.
    const elsewhere = move("activate", { source: "elsewhere", has_targets: true });
    expect(classifyMove(elsewhere, false, controllers, "p0")).toBe("ability");
  });

  it("hasResponse ignores it on an opponent's upkeep, and still counts the viewer's own", () => {
    const opp = (moves: LegalMoveView[]) =>
      snap({ step: "upkeep", active: 1, holder: 0, moves, battlefield: board });
    expect(hasResponse(opp([pass, across]), "p0", ALL_RESPONSES)).toBe(false);
    expect(hasResponse(opp([pass, across, own]), "p0", ALL_RESPONSES)).toBe(true);
  });

  it("hasPlay ignores it in the viewer's own main, and still counts the viewer's own", () => {
    const main = (moves: LegalMoveView[]) => snap({ moves, battlefield: board });
    expect(hasPlay(main([pass, across]), "p0", ALL_RESPONSES)).toBe(false);
    expect(hasPlay(main([pass, own]), "p0", ALL_RESPONSES)).toBe(true);
  });

  it("the controller's own move on it still counts for the controller", () => {
    // p1 is the viewer now, and Xantcha is theirs.
    const s = snap({
      step: "upkeep",
      active: 0,
      holder: 1,
      moves: [pass, across],
      battlefield: board,
    });
    expect(hasResponse(s, "p1", ALL_RESPONSES)).toBe(true);
  });
});

// #2871: a combat ability is a response in a combat window only.
describe("combat_interacts", () => {
  const crew = move("activate", { source: "copter", combat_interacts: true });

  it("classifies as ability in combat and untargeted elsewhere", () => {
    expect(classifyMove(crew, false, undefined, undefined, true)).toBe("ability");
    expect(classifyMove(crew, false, undefined, undefined, false)).toBe("untargeted");
    expect(classifyMove(crew, false)).toBe("untargeted");
    // The viewer's own main phase: a play, as any activation is.
    expect(classifyMove(crew, true, undefined, undefined, true)).toBe("play");
  });

  it("inCombatWindow: beginning of combat, attackers and blockers, nothing later", () => {
    for (const step of ["begin_combat", "declare_attackers", "declare_blockers"]) {
      expect(inCombatWindow(snap({ step })), step).toBe(true);
    }
    for (const step of ["upkeep", "precombat_main", "combat_damage", "end_combat", "end"]) {
      expect(inCombatWindow(snap({ step })), step).toBe(false);
    }
  });

  it("inCombatWindow: an attack or block trigger on the stack, not a spell", () => {
    const trig = (label: string, kind: StackItemView["kind"] = "triggered"): StackItemView => ({
      ...stackItem("p1"),
      kind,
      label,
    });
    const at = (it: StackItemView) => inCombatWindow(snap({ step: "end", stackItems: [it] }));
    expect(at(trig("Whenever this creature blocks, it gets +1/+1."))).toBe(true);
    expect(at(trig("Whenever a creature attacks you, draw a card."))).toBe(true);
    expect(at(trig("At the beginning of your end step, draw a card."))).toBe(false);
    expect(at(trig("Blocking Party", "spell"))).toBe(false);
  });

  it("hasResponse counts it at declare blockers and not on an opponent's main phase", () => {
    const blockers = snap({ step: "declare_blockers", active: 1, moves: [pass, crew] });
    expect(hasResponse(blockers, "p0", DEFAULT_RESPONSES)).toBe(true);
    const main = snap({
      step: "precombat_main",
      active: 1,
      moves: [pass, crew],
      stackItems: [stackItem("p1")],
    });
    expect(hasResponse(main, "p0", DEFAULT_RESPONSES)).toBe(false);
    // With value abilities ticked it counts anywhere, as before.
    expect(hasResponse(main, "p0", { ...DEFAULT_RESPONSES, untargeted: true })).toBe(true);
  });

  // Owner answer (#2871): a creature-token maker is a blocker, so it
  // counts only while the viewer is being attacked.
  const tokens = move("activate", {
    source: "castle",
    combat_interacts: true,
    combat_defender_only: true,
  });
  const attacker = (target: string, defending: string = target): CardView =>
    card("attacker", {
      owner: "p1",
      controller: "p1",
      attacking_target: target,
      defending_player: defending,
    });

  it("classifies a token maker as ability only when defending", () => {
    expect(classifyMove(tokens, false, undefined, undefined, true, true)).toBe("ability");
    expect(classifyMove(tokens, false, undefined, undefined, true, false)).toBe("untargeted");
    expect(classifyMove(tokens, false, undefined, undefined, false, true)).toBe("untargeted");
    // A crew is not narrowed.
    expect(classifyMove(crew, false, undefined, undefined, true, false)).toBe("ability");
  });

  it("isDefending: a creature attacks me, or a planeswalker I defend", () => {
    expect(isDefending(snap({ battlefield: [attacker("p0")] }), "p0")).toBe(true);
    expect(isDefending(snap({ battlefield: [attacker("pw-1", "p0")] }), "p0")).toBe(true);
    expect(isDefending(snap({ battlefield: [attacker("p2")] }), "p0")).toBe(false);
    expect(isDefending(snap({ battlefield: [] }), "p0")).toBe(false);
  });

  it("a token maker stops me when I am attacked, not when another player is", () => {
    const at = (target: string) =>
      snap({
        step: "declare_attackers",
        active: 1,
        moves: [pass, tokens],
        battlefield: [attacker(target)],
      });
    expect(hasResponse(at("p0"), "p0", DEFAULT_RESPONSES)).toBe(true);
    expect(hasResponse(at("p2"), "p0", DEFAULT_RESPONSES)).toBe(false);
  });
});
