import { describe, it, expect } from "vitest";

import { autopassDecision, isBluff, stepStopFor, type AutopassGates } from "./autopassDecision";
import { defaultSettings } from "./settings";
import { ownsEveryStackItem } from "./holdPriority";
import type { CardView, GameView, LegalMoveView, PlayerView, StackItemView } from "./protocol";
import { DEFAULT_RESPONSES, hasPlay, hasResponse, keyWindow } from "./responseWindow";
import { noteStackSeen, stackHoldRemainingMs } from "./stackHold";
import { stackEmpty } from "./timing";

// A viewer holding priority on an empty stack at their opponent's
// upkeep with the default settings: nothing pinned, nothing owed, the
// autopass toggle off. `autopassDecision` on this returns "pass",
// which is the S13 default ("auto-pass every step you haven't
// pinned"). Every test below states only what it changes.
function gates(overrides: Partial<AutopassGates> = {}): AutopassGates {
  return {
    viewerHasPriority: true,
    tableBusy: false,
    hasPendingChoice: false,
    owesBlockDecision: false,
    owesAttackRequirement: false,
    loopSuspended: false,
    step: "upkeep",
    autopassToggle: false,
    viewerIsActive: false,
    manualStop: false,
    passMode: "smart",
    stackEmpty: true,
    holdPriority: false,
    autoPassOwnStack: true,
    ownsEveryStackItem: false,
    stepStop: undefined,
    stepStopsOnlyWhenCanAct: true,
    hasResponse: false,
    hasPlay: false,
    engineMayMissMana: false,
    combatWindow: false,
    oppEndWindow: false,
    bluffCounter: false,
    bluffInstant: false,
    bluffManual: false,
    ...overrides,
  };
}

describe("autopassDecision — the baseline", () => {
  it("passes an unpinned, unstopped window", () => {
    expect(autopassDecision(gates())).toBe("pass");
  });

  it("holds when the viewer does not hold priority", () => {
    expect(autopassDecision(gates({ viewerHasPriority: false }))).toBe("hold");
  });

  it("holds in Manual", () => {
    expect(autopassDecision(gates({ passMode: "manual" }))).toBe("hold");
  });

  it("holds on an unknown step", () => {
    expect(autopassDecision(gates({ step: null }))).toBe("hold");
    expect(autopassDecision(gates({ step: undefined }))).toBe("hold");
  });
});

// #526 — the reported bug. The player turns autopass on, then clicks
// a phase icon to pin a stop, and the game runs straight through it.
// The pin used to be read inside the `if (!autopass)` branch, so it
// was only ever consulted while the toggle was OFF — the one state in
// which the player had least need of it.
describe("autopassDecision — #526: a manual stop beats autopass", () => {
  it("holds a pinned step with the autopass toggle ON", () => {
    expect(autopassDecision(gates({ autopassToggle: true, manualStop: true }))).toBe("hold");
  });

  it("still passes every unpinned step with the toggle ON", () => {
    // The fix must not turn the toggle into a no-op: autopass keeps
    // doing its job everywhere the player has not asked to stop.
    expect(autopassDecision(gates({ autopassToggle: true, manualStop: false }))).toBe("pass");
  });

  it("holds a pinned step with the toggle OFF (the pre-fix behaviour, kept)", () => {
    expect(autopassDecision(gates({ autopassToggle: false, manualStop: true }))).toBe("hold");
  });

  it("holds a pinned step that only-when-I-can-act would otherwise skip", () => {
    // The stops grid says stop, the predicate says "nothing to do",
    // and without the pin that combination passes. ADR 0009 §5: a pin
    // is "I want the cursor even though the engine sees no reason for
    // it" — bluffing, thinking, or an action the catalog doesn't
    // model yet.
    const g = gates({ stepStop: true, hasPlay: false });
    expect(autopassDecision({ ...g, manualStop: false })).toBe("pass");
    expect(autopassDecision({ ...g, manualStop: true })).toBe("hold");
  });

  it("holds a pinned step with the viewer's own spell on the stack (#323)", () => {
    // The #323 own-stack pass is below the pin too: a pinned step
    // hands you the cursor even with your own spell up.
    const g = gates({
      stackEmpty: false,
      ownsEveryStackItem: true,
      autoPassOwnStack: true,
    });
    expect(autopassDecision({ ...g, manualStop: false })).toBe("pass");
    expect(autopassDecision({ ...g, manualStop: true })).toBe("hold");
  });

  it("holds a pinned step under autopass even when nothing else would", () => {
    // The full reported configuration: toggle on, grid says pass,
    // smart says nothing to do, stack empty, opponent's turn.
    const g = gates({
      autopassToggle: true,
      stepStop: false,
      hasPlay: false,
      step: "declare_blockers",
      manualStop: true,
    });
    expect(autopassDecision(g)).toBe("hold");
  });
});

// #599 (PR #606) — smart auto-pass must not close the
// declare-attackers window straight after a bulk declaration, because
// #318's Undo button lives inside it and nothing else in the client
// can put a declared attacker back. The predicate half of this is
// `hasDeclaredAttackers` in priority.ts (tested there); the half that
// lives here is that a `true` hasPlay on a stopped step
// still HOLDS, and that the autopass toggle still passes through it
// ("Scope is the smart-autopass path only. An explicit autopass
// toggle still passes." — #599).
describe("autopassDecision — #599: the declare-attackers review window", () => {
  const declareAttackers = (overrides: Partial<AutopassGates> = {}) =>
    gates({
      step: "declare_attackers",
      viewerIsActive: true,
      stepStop: true,
      // hasPlay returns true here via hasDeclaredAttackers
      // even though the enumerator has only `pass` left to offer.
      hasPlay: true,
      ...overrides,
    });

  it("holds the window so the Undo button stays mounted", () => {
    expect(autopassDecision(declareAttackers())).toBe("hold");
  });

  it("still holds it when the step is also pinned", () => {
    expect(autopassDecision(declareAttackers({ manualStop: true }))).toBe("hold");
  });

  it("passes it with nothing to review and the step unstopped", () => {
    // Nothing to review and no stop: smart-skip's own case.
    expect(autopassDecision(declareAttackers({ stepStop: false, hasPlay: false }))).toBe("pass");
  });

  it("an explicit autopass toggle still passes the window", () => {
    // Unchanged by #526: the toggle out-votes the smart-autopass
    // guard, as PR #606 stated. Only a manual pin now out-votes the
    // toggle.
    expect(autopassDecision(declareAttackers({ autopassToggle: true }))).toBe("pass");
  });

  it("stop-only-when-I-can-act off keeps a stopped declare-attackers window", () => {
    expect(
      autopassDecision(declareAttackers({ stepStopsOnlyWhenCanAct: false, hasPlay: false })),
    ).toBe("hold");
  });
});

// Everything above the toggle in the chain. These are the guards no
// player-facing switch may out-vote, so a manual pin cannot weaken
// them either — and since they all hold, a pin cannot strengthen them.
describe("autopassDecision — guards above every toggle", () => {
  it("holds while mulligans are open, the game is over, or the viewer is out", () => {
    expect(autopassDecision(gates({ tableBusy: true, autopassToggle: true }))).toBe("hold");
  });

  it("holds while a pending choice is unanswered", () => {
    expect(autopassDecision(gates({ hasPendingChoice: true, autopassToggle: true }))).toBe("hold");
  });

  it("holds an owed declare-blockers decision (#328), toggle or not", () => {
    expect(autopassDecision(gates({ owesBlockDecision: true }))).toBe("hold");
    // #1571: an owed attack requirement holds too — the server would
    // refuse the pass, even with the autopass toggle armed.
    expect(autopassDecision(gates({ owesAttackRequirement: true }))).toBe("hold");
    expect(autopassDecision(gates({ owesAttackRequirement: true, autopassToggle: true }))).toBe(
      "hold",
    );
    expect(autopassDecision(gates({ owesBlockDecision: true, autopassToggle: true }))).toBe("hold");
  });

  it("holds while the CR 732 loop breaker is up (#628)", () => {
    expect(autopassDecision(gates({ loopSuspended: true, autopassToggle: true }))).toBe("hold");
  });
});

// ADR 0009 §7. The belt clears the toggle on entering the viewer's
// own precombat_main rather than passing, so the cursor holds either
// way; what matters is that it still fires when the step is ALSO
// pinned, otherwise a pin on your own main phase would leave a
// forgotten toggle armed for the rest of the turn.
describe("autopassDecision — the autopass safety belt", () => {
  const ownMain = (overrides: Partial<AutopassGates> = {}) =>
    gates({
      step: "precombat_main",
      viewerIsActive: true,
      autopassToggle: true,
      stepStop: true,
      ...overrides,
    });

  it("clears the toggle on the viewer's own precombat_main", () => {
    expect(autopassDecision(ownMain())).toBe("clear-toggle");
  });

  it("clears it even when the step is pinned, and the cursor holds anyway", () => {
    expect(autopassDecision(ownMain({ manualStop: true }))).toBe("clear-toggle");
  });

  // ADR 0143 §4.2: there is no setting that keeps Skip to my turn on
  // through the viewer's own main phase any more. It clears whatever
  // else is going on.
  it("always fires: no setting, pass mode or stop keeps it on", () => {
    for (const passMode of ["smart", "careful", "manual"] as const) {
      expect(autopassDecision(ownMain({ passMode })), passMode).toBe("clear-toggle");
    }
    expect(autopassDecision(ownMain({ stepStop: false, hasPlay: false }))).toBe("clear-toggle");
    expect(autopassDecision(ownMain({ hasResponse: true }))).toBe("clear-toggle");
  });

  it("does not fire on someone else's precombat_main", () => {
    expect(autopassDecision(ownMain({ viewerIsActive: false }))).toBe("pass");
  });

  it("does not fire with the toggle already off", () => {
    // Toggle off, own main phase, stopped with something to do: the
    // conventional path holds.
    expect(autopassDecision(ownMain({ autopassToggle: false, hasPlay: true }))).toBe("hold");
  });
});

// ADR 0118 owner decision 8: "stop if the engine may be wrong". The
// gate itself (own main phase, a spell left out for mana, a manual mana
// source) is engineMayMissMana.ts's; here it is one input to rule 8.
describe("autopassDecision — stop if the engine may be wrong", () => {
  const ownMain = (overrides: Partial<AutopassGates> = {}) =>
    gates({
      step: "precombat_main",
      viewerIsActive: true,
      stepStop: true,
      hasPlay: false,
      ...overrides,
    });

  it("holds the viewer's own stopped main phase with nothing else to play", () => {
    expect(autopassDecision(ownMain())).toBe("pass");
    expect(autopassDecision(ownMain({ engineMayMissMana: true }))).toBe("hold");
  });

  it("changes nothing on a step the player did not tick", () => {
    expect(autopassDecision(ownMain({ engineMayMissMana: true, stepStop: false }))).toBe("pass");
  });

  it("changes nothing with stop-only-when-I-can-act off: the stop already holds", () => {
    expect(autopassDecision(ownMain({ stepStopsOnlyWhenCanAct: false }))).toBe("hold");
    expect(
      autopassDecision(ownMain({ stepStopsOnlyWhenCanAct: false, engineMayMissMana: true })),
    ).toBe("hold");
  });

  it("does not out-vote the autopass toggle or the safety belt", () => {
    expect(autopassDecision(ownMain({ engineMayMissMana: true, autopassToggle: true }))).toBe(
      "clear-toggle",
    );
    // Past the safety belt (the viewer's own postcombat main), the
    // toggle still passes over it.
    expect(
      autopassDecision(
        ownMain({ step: "postcombat_main", engineMayMissMana: true, autopassToggle: true }),
      ),
    ).toBe("pass");
  });
});

// The #323 own-stack carve-out and the stops grid, unchanged by #526
// but pinned here now that they are testable.
describe("autopassDecision — the conventional path", () => {
  it("holds a stack with an opponent's item on it when the viewer can respond", () => {
    expect(
      autopassDecision(gates({ stackEmpty: false, ownsEveryStackItem: false, hasResponse: true })),
    ).toBe("hold");
  });

  it("passes a stack that is entirely the viewer's own (#323)", () => {
    expect(
      autopassDecision(gates({ stackEmpty: false, ownsEveryStackItem: true, stepStop: true })),
    ).toBe("pass");
  });

  it("holds the viewer's own stack when hold-priority is armed", () => {
    expect(
      autopassDecision(gates({ stackEmpty: false, ownsEveryStackItem: true, holdPriority: true })),
    ).toBe("hold");
  });

  it("holds the viewer's own stack with autoPassOwnStack off", () => {
    expect(
      autopassDecision(
        gates({ stackEmpty: false, ownsEveryStackItem: true, autoPassOwnStack: false }),
      ),
    ).toBe("hold");
  });

  it("holds a stopped step with something to consider", () => {
    expect(autopassDecision(gates({ stepStop: true, hasPlay: true }))).toBe("hold");
  });

  it("passes a stopped step with nothing to consider", () => {
    expect(autopassDecision(gates({ stepStop: true, hasPlay: false }))).toBe("pass");
  });

  it("holds a stopped step with stop-only-when-I-can-act off", () => {
    expect(
      autopassDecision(gates({ stepStop: true, stepStopsOnlyWhenCanAct: false, hasPlay: false })),
    ).toBe("hold");
  });

  // #2871: the setting is its own now. Careful (every opponent stack
  // item stops) leaves a ticked step skipping when there is nothing to
  // do, and the other way round.
  it("skips an empty ticked step in Careful", () => {
    expect(autopassDecision(gates({ stepStop: true, passMode: "careful", hasPlay: false }))).toBe(
      "pass",
    );
  });

  it("passes an unstopped quiet step even with something to play or respond with", () => {
    // #1307: outside the ticked steps only the key windows stop, and
    // an opponent's upkeep with an empty stack is not one of them.
    expect(autopassDecision(gates({ stepStop: false, hasPlay: true }))).toBe("pass");
    expect(autopassDecision(gates({ stepStop: undefined, hasPlay: true }))).toBe("pass");
    expect(autopassDecision(gates({ stepStop: undefined, hasPlay: true, hasResponse: true }))).toBe(
      "pass",
    );
  });

  it("holds a land-only own main phase that is ticked", () => {
    // A land is a play, not a response: it keeps your own ticked main
    // phase open and never holds anyone else's window.
    const g = gates({
      step: "precombat_main",
      viewerIsActive: true,
      stepStop: true,
      hasPlay: true,
      hasResponse: false,
    });
    expect(autopassDecision(g)).toBe("hold");
  });
});

// #1307 — smart autopass on an opponent's stack item. It used to
// hold every time; now it holds only when the viewer has an answer.
describe("autopassDecision — #1307: an opponent's stack", () => {
  const oppStack = (overrides: Partial<AutopassGates> = {}) =>
    gates({ stackEmpty: false, ownsEveryStackItem: false, ...overrides });

  it("passes a spell the viewer cannot answer", () => {
    expect(autopassDecision(oppStack())).toBe("pass");
  });

  it("holds when the viewer has a response", () => {
    expect(autopassDecision(oppStack({ hasResponse: true }))).toBe("hold");
  });

  it("does not hold on a play with no response, even on a ticked step", () => {
    expect(autopassDecision(oppStack({ hasPlay: true, stepStop: true }))).toBe("pass");
  });

  it("Careful stops at every opponent item, answer or not", () => {
    expect(autopassDecision(oppStack({ passMode: "careful" }))).toBe("hold");
  });

  it("hold-priority still holds", () => {
    expect(autopassDecision(oppStack({ holdPriority: true }))).toBe("hold");
  });

  it("bluffs when a bluff is armed and there is no answer", () => {
    expect(autopassDecision(oppStack({ bluffCounter: true }))).toEqual({
      kind: "bluff",
      manual: false,
    });
    expect(autopassDecision(oppStack({ bluffInstant: true, bluffManual: true }))).toEqual({
      kind: "bluff",
      manual: true,
    });
  });

  it("a real response beats a bluff", () => {
    expect(autopassDecision(oppStack({ hasResponse: true, bluffCounter: true }))).toBe("hold");
  });

  it("a stack of the viewer's own never bluffs", () => {
    expect(
      autopassDecision(gates({ stackEmpty: false, ownsEveryStackItem: true, bluffCounter: true })),
    ).toBe("pass");
  });
});

// #1307 — the empty-stack key windows: combat with attackers
// declared, and an opponent's end step.
describe("autopassDecision — #1307: key windows", () => {
  it("holds an unticked combat window with a response", () => {
    const g = gates({ step: "declare_blockers", combatWindow: true, hasResponse: true });
    expect(autopassDecision(g)).toBe("hold");
  });

  it("holds an unticked opponent's end step with a response", () => {
    const g = gates({ step: "end", oppEndWindow: true, hasResponse: true });
    expect(autopassDecision(g)).toBe("hold");
  });

  it("passes a key window with nothing to respond with", () => {
    expect(autopassDecision(gates({ step: "end", oppEndWindow: true, hasPlay: true }))).toBe(
      "pass",
    );
  });

  it("an instant bluff bluffs in a key window, a counter bluff does not", () => {
    const g = gates({ step: "end", oppEndWindow: true });
    expect(isBluff(autopassDecision({ ...g, bluffInstant: true }))).toBe(true);
    expect(autopassDecision({ ...g, bluffCounter: true })).toBe("pass");
  });

  it("Careful keeps the key windows (ADR 0143 §2.2)", () => {
    // Before ADR 0143, turning smart auto-pass off dropped these.
    const g = gates({ step: "end", oppEndWindow: true, hasResponse: true, passMode: "careful" });
    expect(autopassDecision(g)).toBe("hold");
  });

  it("a ticked key window with no response falls through to the stops grid", () => {
    const g = gates({ step: "end", oppEndWindow: true, stepStop: true, hasPlay: true });
    expect(autopassDecision(g)).toBe("hold");
  });
});

// #1307 — the Shift+P toggle stops for a real answer to an
// opponent's stack item, and bluffs there if asked. Everywhere else
// it passes as before.
describe("autopassDecision — #1307: the autopass toggle", () => {
  const oppStack = (overrides: Partial<AutopassGates> = {}) =>
    gates({ autopassToggle: true, stackEmpty: false, ownsEveryStackItem: false, ...overrides });

  it("holds an opponent's stack item the viewer can answer", () => {
    expect(autopassDecision(oppStack({ hasResponse: true }))).toBe("hold");
  });

  it("passes one the viewer cannot answer", () => {
    expect(autopassDecision(oppStack())).toBe("pass");
  });

  it("bluffs at one when a bluff is armed", () => {
    expect(isBluff(autopassDecision(oppStack({ bluffCounter: true })))).toBe(true);
  });

  // ADR 0143 §4.1/§4.2: Skip to my turn is Smart's passing until you
  // are next active, so it keeps Smart's key windows. With a response
  // in hand it stops; with none, it passes, ticked step or not.
  it("stops at a key window when the viewer can respond", () => {
    const end = gates({ autopassToggle: true, step: "end", oppEndWindow: true });
    expect(autopassDecision({ ...end, hasResponse: true })).toBe("hold");
    expect(autopassDecision({ ...end, hasResponse: false })).toBe("pass");
    for (const step of ["declare_attackers", "declare_blockers"]) {
      const combat = gates({ autopassToggle: true, step, combatWindow: true });
      expect(autopassDecision({ ...combat, hasResponse: true }), step).toBe("hold");
      expect(autopassDecision({ ...combat, hasResponse: false }), step).toBe("pass");
    }
  });

  it("passes ticked steps and quiet windows, and never bluffs a key window", () => {
    const g = gates({
      autopassToggle: true,
      step: "end",
      oppEndWindow: true,
      hasResponse: false,
      stepStop: true,
      hasPlay: true,
      bluffInstant: true,
    });
    expect(autopassDecision(g)).toBe("pass");
    expect(
      autopassDecision(
        gates({ autopassToggle: true, step: "upkeep", stepStop: true, hasPlay: true }),
      ),
    ).toBe("pass");
    // A response outside the key windows (an opponent's main phase with
    // an instant in hand) is not a reason to stop.
    expect(
      autopassDecision(gates({ autopassToggle: true, step: "precombat_main", hasResponse: true })),
    ).toBe("pass");
  });

  it("still stops for a choice the table waits on", () => {
    const base = { autopassToggle: true, step: "declare_blockers" };
    expect(autopassDecision(gates({ ...base, hasPendingChoice: true }))).toBe("hold");
    expect(autopassDecision(gates({ ...base, owesBlockDecision: true }))).toBe("hold");
    expect(autopassDecision(gates({ ...base, owesAttackRequirement: true }))).toBe("hold");
    expect(autopassDecision(gates({ ...base, loopSuspended: true }))).toBe("hold");
  });

  it("clears at the viewer's own main 1, and only there", () => {
    const own = { autopassToggle: true, viewerIsActive: true };
    expect(autopassDecision(gates({ ...own, step: "precombat_main" }))).toBe("clear-toggle");
    for (const step of ["upkeep", "draw", "declare_attackers", "postcombat_main", "end"]) {
      expect(autopassDecision(gates({ ...own, step })), step).not.toBe("clear-toggle");
    }
    expect(
      autopassDecision(
        gates({ autopassToggle: true, viewerIsActive: false, step: "precombat_main" }),
      ),
    ).toBe("pass");
  });

  it("ignores Careful and Manual", () => {
    expect(autopassDecision(oppStack({ passMode: "careful" }))).toBe("pass");
    expect(autopassDecision(oppStack({ passMode: "manual" }))).toBe("pass");
    expect(autopassDecision(oppStack({ passMode: "careful", hasResponse: true }))).toBe("hold");
  });

  it("bluffs only in Smart (ADR 0143 §4.3)", () => {
    expect(autopassDecision(oppStack({ passMode: "careful", bluffCounter: true }))).toBe("pass");
    expect(autopassDecision(oppStack({ passMode: "manual", bluffCounter: true }))).toBe("pass");
  });
});

// #1307 bluffing. A bluff only ever replaces a pass: it never beats a
// guard, a pin, a real answer or anything else that holds.
describe("autopassDecision — #1307: bluffs", () => {
  const armed = { bluffCounter: true, bluffInstant: true };
  const oppStack = (overrides: Partial<AutopassGates> = {}) =>
    gates({ stackEmpty: false, ownsEveryStackItem: false, ...armed, ...overrides });

  it("carries the manual flag through", () => {
    expect(autopassDecision(oppStack({ bluffManual: false }))).toEqual({
      kind: "bluff",
      manual: false,
    });
    expect(autopassDecision(oppStack({ bluffManual: true }))).toEqual({
      kind: "bluff",
      manual: true,
    });
  });

  it("never beats a guard or a pin", () => {
    expect(autopassDecision(oppStack({ viewerHasPriority: false }))).toBe("hold");
    expect(autopassDecision(oppStack({ hasPendingChoice: true }))).toBe("hold");
    expect(autopassDecision(oppStack({ loopSuspended: true }))).toBe("hold");
    expect(autopassDecision(oppStack({ manualStop: true }))).toBe("hold");
  });

  it("does not bluff where smart autopass would hold anyway", () => {
    expect(autopassDecision(oppStack({ passMode: "careful" }))).toBe("hold");
    expect(autopassDecision(oppStack({ passMode: "manual" }))).toBe("hold");
  });

  it("does not bluff a key window outside Smart (ADR 0143 §4.3)", () => {
    const g = gates({ step: "end", oppEndWindow: true, ...armed });
    expect(isBluff(autopassDecision(g))).toBe(true);
    expect(autopassDecision({ ...g, passMode: "careful" })).toBe("pass");
    expect(autopassDecision({ ...g, passMode: "manual" })).toBe("hold");
  });

  it("does not bluff a quiet step, ticked or not", () => {
    expect(autopassDecision(gates({ ...armed }))).toBe("pass");
    expect(autopassDecision(gates({ ...armed, stepStop: true, hasPlay: true }))).toBe("hold");
    expect(autopassDecision(gates({ ...armed, stepStop: true, hasPlay: false }))).toBe("pass");
  });

  it("bluffs combat under an instant bluff only", () => {
    const g = gates({ step: "declare_attackers", combatWindow: true });
    expect(isBluff(autopassDecision({ ...g, bluffInstant: true }))).toBe(true);
    expect(autopassDecision({ ...g, bluffCounter: true })).toBe("pass");
  });
});

// #2853 end to end: the gates built from a real frame through
// responseWindow, the way Game.svelte builds them. The issue's board —
// lands, Llanowar Elves, Mind Stone and Evolving Wilds — with an
// opponent's spell on the stack. With the default "Stop for" list the
// ADR 0119 stack hold runs its course and then auto-pass passes.
describe("autopassDecision — #2853: an untargeted value ability is not a response", () => {
  const me = "p0";
  const seatOf = (id: string, seat: number): PlayerView => ({
    id,
    name: id,
    seat,
    life: 40,
    library: { kind: "library", owner: id, count: 0, cards: [] },
    hand: { kind: "hand", owner: id, count: 0, cards: [] },
    graveyard: { kind: "graveyard", owner: id, count: 0, cards: [] },
    command: { kind: "command", owner: id, count: 0, cards: [] },
    commander_damage: {},
    life_history: [],
  });
  const mv = (kind: LegalMoveView["kind"], extras: Partial<LegalMoveView> = {}): LegalMoveView => ({
    type: "x",
    player: me,
    kind,
    label: kind,
    ...extras,
  });
  const valueBoard: LegalMoveView[] = [
    mv("pass"),
    mv("mana", { source: "forest" }),
    mv("mana", { source: "elves" }),
    mv("activate", { source: "mind-stone" }),
    mv("activate", { source: "evolving-wilds" }),
  ];
  const frame = (moves: LegalMoveView[]): GameView => ({
    id: "g",
    state: "active",
    seats: [seatOf("p0", 0), seatOf("p1", 1)],
    battlefield: { kind: "battlefield", owner: "", count: 0, cards: [] },
    stack: { kind: "stack", owner: "", count: 0, cards: [] },
    exile: { kind: "exile", owner: "", count: 0, cards: [] },
    turn: {
      seq: 1,
      number: 1,
      active_seat: 1,
      priority_holder: 0,
      phase: "x",
      step: "precombat_main",
    },
    mulligans_open: false,
    stack_items: [
      { id: "opp-spell", kind: "spell", controller: "p1", owner: "p1", source_card_id: "c" },
    ],
    split_second_active: false,
    legal_moves: moves,
  });
  const gatesFor = (view: GameView, cats = DEFAULT_RESPONSES): AutopassGates =>
    gates({
      step: view.turn.step,
      stackEmpty: false,
      ownsEveryStackItem: ownsEveryStackItem(view, me),
      hasResponse: hasResponse(view, me, cats),
      hasPlay: hasPlay(view, me, cats),
    });

  it("passes once the stack hold has run out", () => {
    const view = frame(valueBoard);
    expect(autopassDecision(gatesFor(view))).toBe("pass");
    const firstSeen = noteStackSeen(new Map(), view, 0);
    const hold = (now: number) =>
      stackHoldRemainingMs({ view, viewerID: me, firstSeen, holdMs: 2000, now });
    expect(hold(0)).toBe(2000);
    expect(hold(2000)).toBe(0);
  });

  it("holds with untargeted abilities ticked, as before #2853", () => {
    const view = frame(valueBoard);
    expect(autopassDecision(gatesFor(view, { ...DEFAULT_RESPONSES, untargeted: true }))).toBe(
      "hold",
    );
  });

  // The owner's other half: no missed windows. A real answer means the
  // decision is hold, so neither the immediate pass nor the stack
  // hold's timed pass (which re-asks this decision when it fires) goes.
  it("holds for a targeted ability, an interacting one, an instant and a counter", () => {
    for (const extra of [
      mv("activate", { source: "pinger", has_targets: true }),
      mv("activate", { source: "viscera-seer", interacts: true }),
      mv("activate", { source: "undercity-troll", interacts: true }),
      mv("activate", { source: "evernight-shade", interacts: true }),
      mv("mana", { source: "ashnods-altar", interacts: true }),
      mv("cast", { source: "instant" }),
      mv("cast", { source: "counter", targets_stack: true, has_targets: true }),
    ]) {
      expect(autopassDecision(gatesFor(frame([...valueBoard, extra]))), extra.source).toBe("hold");
    }
  });

  it("the hold toggle holds it whatever the categories say", () => {
    expect(autopassDecision({ ...gatesFor(frame(valueBoard)), holdPriority: true })).toBe("hold");
  });

  // ADR 0143 finding 3: the categories decide the autopass toggle's
  // hold too, in every mode, so the UI never greys them out.
  it("the autopass toggle reads the categories, in every mode", () => {
    const view = frame([...valueBoard, mv("cast", { source: "instant" })]);
    for (const passMode of ["smart", "careful", "manual"] as const) {
      const on = { ...gatesFor(view), autopassToggle: true, passMode };
      expect(autopassDecision(on), passMode).toBe("hold");
      const off = {
        ...gatesFor(view, { ...DEFAULT_RESPONSES, instant: false }),
        autopassToggle: true,
        passMode,
      };
      expect(autopassDecision(off), passMode).toBe("pass");
    }
  });
});

// ADR 0143 §2.1 and §2.2: the three modes against each key window.
// Smart and Careful stop at an opponent's item, at combat once
// attackers are declared, and at an opponent's end step whenever the
// viewer can respond, and pass when they can't (Careful still stops at
// every opponent item). Manual stops everywhere.
describe("autopassDecision — ADR 0143: pass modes and the key windows", () => {
  const windows: Record<string, Partial<AutopassGates>> = {
    "an opponent's spell": { step: "precombat_main", stackEmpty: false, ownsEveryStackItem: false },
    "combat once attackers are declared": { step: "declare_attackers", combatWindow: true },
    "combat at declare blockers": { step: "declare_blockers", combatWindow: true },
    "an opponent's end step": { step: "end", oppEndWindow: true },
  };
  // What each mode does there with nothing to respond with.
  const noResponse: Record<string, Record<string, "hold" | "pass">> = {
    smart: {
      "an opponent's spell": "pass",
      "combat once attackers are declared": "pass",
      "combat at declare blockers": "pass",
      "an opponent's end step": "pass",
    },
    careful: {
      "an opponent's spell": "hold",
      "combat once attackers are declared": "pass",
      "combat at declare blockers": "pass",
      "an opponent's end step": "pass",
    },
  };

  for (const passMode of ["smart", "careful"] as const) {
    for (const [name, w] of Object.entries(windows)) {
      it(`${passMode}: stops at ${name} when you can respond`, () => {
        expect(autopassDecision(gates({ ...w, passMode, hasResponse: true }))).toBe("hold");
      });
      it(`${passMode}: ${name} when you can't respond → ${noResponse[passMode][name]}`, () => {
        expect(autopassDecision(gates({ ...w, passMode, hasResponse: false }))).toBe(
          noResponse[passMode][name],
        );
      });
      it(`${passMode}: an unticked ${name} still stops for a response`, () => {
        expect(
          autopassDecision(gates({ ...w, passMode, stepStop: false, hasResponse: true })),
        ).toBe("hold");
      });
    }
  }

  for (const [name, w] of Object.entries(windows)) {
    it(`manual: stops at ${name} with or without a response`, () => {
      expect(autopassDecision(gates({ ...w, passMode: "manual", hasResponse: true }))).toBe("hold");
      expect(autopassDecision(gates({ ...w, passMode: "manual", hasResponse: false }))).toBe(
        "hold",
      );
    });
  }

  it("manual stops a quiet step too", () => {
    expect(autopassDecision(gates({ passMode: "manual" }))).toBe("hold");
    expect(autopassDecision(gates({ passMode: "smart" }))).toBe("pass");
    expect(autopassDecision(gates({ passMode: "careful" }))).toBe("pass");
  });

  it("no mode turns the key windows off: a held response always holds", () => {
    for (const passMode of ["smart", "careful", "manual"] as const) {
      for (const w of Object.values(windows)) {
        expect(autopassDecision(gates({ ...w, passMode, hasResponse: true }))).toBe("hold");
      }
    }
  });
});

// #2871: a ticked step stops only when the viewer can do something
// there, and a combat ability (crew, a manland, a granted keyword) is
// something to do in combat and nowhere else. These run the frame
// through the same helpers Game.svelte does, with the default "Stop
// for" categories.
describe("autopassDecision — #2871: stop at a ticked step only when you can act", () => {
  const me = "p0";
  const seatOf = (id: string, seat: number): PlayerView => ({
    id,
    name: id,
    seat,
    life: 40,
    library: { kind: "library", owner: id, count: 0, cards: [] },
    hand: { kind: "hand", owner: id, count: 0, cards: [] },
    graveyard: { kind: "graveyard", owner: id, count: 0, cards: [] },
    command: { kind: "command", owner: id, count: 0, cards: [] },
    commander_damage: {},
    life_history: [],
  });
  const mv = (kind: LegalMoveView["kind"], extras: Partial<LegalMoveView> = {}): LegalMoveView => ({
    type: "x",
    player: me,
    kind,
    label: kind,
    ...extras,
  });
  // Lands, a mana rock and a value ability: nothing to do off your own
  // main phase.
  const quiet: LegalMoveView[] = [
    mv("pass"),
    mv("mana", { source: "forest" }),
    mv("mana", { source: "mind-stone" }),
    mv("activate", { source: "mind-stone" }),
  ];
  const instant = mv("cast", { source: "instant" });
  const crew = mv("activate", { source: "copter", combat_interacts: true });
  const manland = mv("activate", { source: "anchorage", combat_interacts: true });
  const castle = mv("activate", {
    source: "castle-ardenvale",
    combat_interacts: true,
    combat_defender_only: true,
  });

  interface Frame {
    step: string;
    active: number;
    moves: LegalMoveView[];
    stack?: StackItemView[];
    attacking?: boolean;
    // Whom the attacker attacks; the viewer by default.
    attacked?: string;
  }
  const frame = (f: Frame): GameView => {
    const attackers: CardView[] = f.attacking
      ? [
          {
            instance_id: "attacker",
            name: "attacker",
            owner: "p1",
            controller: "p1",
            type_line: "Creature",
            attacking_target: f.attacked ?? me,
            defending_player: f.attacked ?? me,
          },
        ]
      : [];
    return {
      id: "g",
      state: "active",
      seats: [seatOf("p0", 0), seatOf("p1", 1)],
      battlefield: { kind: "battlefield", owner: "", count: attackers.length, cards: attackers },
      stack: { kind: "stack", owner: "", count: 0, cards: [] },
      exile: { kind: "exile", owner: "", count: 0, cards: [] },
      turn: {
        seq: 1,
        number: 1,
        active_seat: f.active,
        priority_holder: 0,
        phase: "x",
        step: f.step,
      },
      mulligans_open: false,
      stack_items: f.stack ?? [],
      split_second_active: false,
      legal_moves: f.moves,
    };
  };
  const decide = (view: GameView, ticked: boolean, over: Partial<AutopassGates> = {}) => {
    const kw = keyWindow(view, me);
    return autopassDecision(
      gates({
        step: view.turn.step,
        viewerIsActive: view.turn.active_seat === 0,
        stackEmpty: stackEmpty(view),
        ownsEveryStackItem: ownsEveryStackItem(view, me),
        stepStop: ticked,
        hasResponse: hasResponse(view, me, DEFAULT_RESPONSES),
        hasPlay: hasPlay(view, me, DEFAULT_RESPONSES),
        combatWindow: kw.combat,
        oppEndWindow: kw.oppEnd,
        ...over,
      }),
    );
  };

  it("a ticked upkeep with nothing to do passes; with an instant in hand it stops", () => {
    for (const active of [0, 1]) {
      expect(decide(frame({ step: "upkeep", active, moves: quiet }), true)).toBe("pass");
      expect(decide(frame({ step: "upkeep", active, moves: [...quiet, instant] }), true)).toBe(
        "hold",
      );
    }
  });

  it("a ticked upkeep stops every time with the setting off", () => {
    const view = frame({ step: "upkeep", active: 1, moves: quiet });
    expect(decide(view, true, { stepStopsOnlyWhenCanAct: false })).toBe("hold");
  });

  it("a ticked own main phase stops for any play: a land, a sorcery, a value ability", () => {
    for (const play of [
      mv("land", { source: "land-in-hand" }),
      mv("cast", { source: "sorcery" }),
      mv("activate", { source: "mind-stone" }),
      crew,
    ]) {
      const view = frame({ step: "precombat_main", active: 0, moves: [mv("pass"), play] });
      expect(decide(view, true), play.source).toBe("hold");
      const post = frame({ step: "postcombat_main", active: 0, moves: [mv("pass"), play] });
      expect(decide(post, true), play.source).toBe("hold");
    }
  });

  it("a ticked own main phase with only mana passes", () => {
    const view = frame({
      step: "postcombat_main",
      active: 0,
      moves: [mv("pass"), mv("mana", { source: "forest" })],
    });
    expect(decide(view, true)).toBe("pass");
  });

  it("a crew or a manland stops you at declare blockers", () => {
    for (const ability of [crew, manland]) {
      const view = frame({
        step: "declare_blockers",
        active: 1,
        moves: [...quiet, ability],
        attacking: true,
      });
      // Ticked (the default) and unticked: combat is a key window.
      expect(decide(view, true), ability.source).toBe("hold");
      expect(decide(view, false), ability.source).toBe("hold");
    }
    // Without one, the same window passes.
    const none = frame({ step: "declare_blockers", active: 1, moves: quiet, attacking: true });
    expect(decide(none, true)).toBe("pass");
  });

  it("a crew or a manland stops you at beginning of combat when the step is ticked", () => {
    const view = frame({ step: "begin_combat", active: 1, moves: [...quiet, manland] });
    expect(decide(view, true)).toBe("hold");
    expect(decide(view, false)).toBe("pass");
  });

  it("a crew or a manland does not stop you for an opponent's main-phase sorcery", () => {
    const sorcery: StackItemView = {
      id: "s",
      kind: "spell",
      controller: "p1",
      owner: "p1",
      source_card_id: "c",
      label: "Divination",
    };
    for (const ability of [crew, manland]) {
      const view = frame({
        step: "precombat_main",
        active: 1,
        moves: [...quiet, ability],
        stack: [sorcery],
      });
      expect(decide(view, true), ability.source).toBe("pass");
    }
  });

  // Owner answer: a creature-token maker counts only for a player who
  // is being attacked; crew and manlands count for everyone.
  it("a token maker stops you when you are attacked, not when another player is", () => {
    for (const step of ["declare_attackers", "declare_blockers"]) {
      const me0 = frame({ step, active: 1, moves: [...quiet, castle], attacking: true });
      expect(decide(me0, false), step).toBe("hold");
      const other = frame({
        step,
        active: 1,
        moves: [...quiet, castle],
        attacking: true,
        attacked: "p2",
      });
      expect(decide(other, false), step).toBe("pass");
      expect(decide(other, true), step).toBe("pass");
    }
    // Before attackers are declared nobody is defending yet.
    expect(
      decide(frame({ step: "begin_combat", active: 1, moves: [...quiet, castle] }), true),
    ).toBe("pass");
  });

  it("crew and manlands stop you whoever is attacked", () => {
    for (const ability of [crew, manland]) {
      const other = frame({
        step: "declare_attackers",
        active: 1,
        moves: [...quiet, ability],
        attacking: true,
        attacked: "p2",
      });
      expect(decide(other, false), ability.source).toBe("hold");
    }
  });

  it("an attack trigger on the stack is a combat window", () => {
    const trigger: StackItemView = {
      id: "t",
      kind: "triggered",
      controller: "p1",
      owner: "p1",
      source_card_id: "c",
      label: "Whenever this creature attacks, create a 1/1 token.",
    };
    const view = frame({
      step: "postcombat_main",
      active: 1,
      moves: [...quiet, crew],
      stack: [trigger],
    });
    expect(decide(view, false)).toBe("hold");
  });
});

// ADR 0143 §2.3: rule 8 reads the column for whoever is active. With
// the defaults a player holding an instant no longer stops at every
// opponent's main phases and end step; the key windows still stop
// them whenever they can respond.
describe("autopassDecision — ADR 0143: stops by whose turn it is", () => {
  type Columns = Parameters<typeof stepStopFor>[0];
  const defaults = (): Columns => {
    const gp = defaultSettings().gameplay;
    return { stepStops: gp.stepStops, stepStopsOpponents: gp.stepStopsOpponents };
  };
  // An instant in hand: something to play outside the viewer's own
  // sorcery window (hasPlay). Whether it is also a response is up to
  // each test.
  const holdingInstant = (
    columns: Columns,
    step: string,
    viewerIsActive: boolean,
    extra: Partial<AutopassGates> = {},
  ) =>
    gates({
      step,
      viewerIsActive,
      stepStop: stepStopFor(columns, step, viewerIsActive),
      hasPlay: true,
      ...extra,
    });

  it("stepStopFor reads My turn on your turn and Opponents' turns otherwise", () => {
    const columns = { stepStops: { upkeep: true }, stepStopsOpponents: { upkeep: false } };
    expect(stepStopFor(columns, "upkeep", true)).toBe(true);
    expect(stepStopFor(columns, "upkeep", false)).toBe(false);
    expect(stepStopFor(columns, "draw", true)).toBeUndefined();
    expect(stepStopFor(columns, null, true)).toBeUndefined();
    // The two combat damage steps share one stop, in both columns.
    const dmg = { stepStops: { combat_damage: true }, stepStopsOpponents: { combat_damage: true } };
    expect(stepStopFor(dmg, "first_strike_damage", true)).toBe(true);
    expect(stepStopFor(dmg, "first_strike_damage", false)).toBe(true);
  });

  it("defaults: an opponent's main 1, main 2 and end step pass with an instant in hand", () => {
    for (const step of ["precombat_main", "postcombat_main", "end"]) {
      expect(autopassDecision(holdingInstant(defaults(), step, false)), step).toBe("pass");
    }
  });

  it("defaults: your own main phases and combat declarations still stop", () => {
    for (const step of [
      "precombat_main",
      "declare_attackers",
      "declare_blockers",
      "postcombat_main",
    ]) {
      expect(autopassDecision(holdingInstant(defaults(), step, true)), step).toBe("hold");
    }
    // Your own end step is no longer ticked.
    expect(autopassDecision(holdingInstant(defaults(), "end", true))).toBe("pass");
  });

  it("defaults: an opponent's attack still stops you when you can respond", () => {
    for (const step of ["declare_attackers", "declare_blockers"]) {
      const g = holdingInstant(defaults(), step, false, { combatWindow: true, hasResponse: true });
      expect(autopassDecision(g), step).toBe("hold");
    }
  });

  it("defaults: an opponent's end step still stops you when you can respond", () => {
    // PR 1's key window, untouched by the columns.
    const g = holdingInstant(defaults(), "end", false, { oppEndWindow: true, hasResponse: true });
    expect(autopassDecision(g)).toBe("hold");
    expect(autopassDecision({ ...g, hasResponse: false })).toBe("pass");
  });

  it("defaults: an opponent's main phase passes even with a response, outside the key windows", () => {
    const g = holdingInstant(defaults(), "precombat_main", false, { hasResponse: true });
    expect(autopassDecision(g)).toBe("pass");
  });

  it("a step ticked for opponents stops on their turns, and not on yours", () => {
    const columns = { stepStops: defaults().stepStops, stepStopsOpponents: { upkeep: true } };
    expect(autopassDecision(holdingInstant(columns, "upkeep", false))).toBe("hold");
    expect(autopassDecision(holdingInstant(columns, "upkeep", true))).toBe("pass");
  });

  it("a player who tuned their grid stops where they did, on every turn", () => {
    // The v25 migration copies a tuned grid into both columns.
    const tuned = { upkeep: true, precombat_main: true, end: true };
    const columns = { stepStops: { ...tuned }, stepStopsOpponents: { ...tuned } };
    for (const step of ["upkeep", "precombat_main", "end"]) {
      for (const active of [true, false]) {
        expect(autopassDecision(holdingInstant(columns, step, active)), `${step} ${active}`).toBe(
          "hold",
        );
      }
    }
    expect(autopassDecision(holdingInstant(columns, "postcombat_main", false))).toBe("pass");
  });

  it("only-when-I-can-act applies to both columns", () => {
    const columns = { stepStops: { upkeep: true }, stepStopsOpponents: { upkeep: true } };
    for (const active of [true, false]) {
      const g = holdingInstant(columns, "upkeep", active, { hasPlay: false });
      expect(autopassDecision(g), String(active)).toBe("pass");
      expect(autopassDecision({ ...g, stepStopsOnlyWhenCanAct: false }), String(active)).toBe(
        "hold",
      );
    }
  });
});
