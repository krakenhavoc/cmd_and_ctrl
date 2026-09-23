import { describe, it, expect } from "vitest";

import { autopassDecision, isBluff, type AutopassGates } from "./autopassDecision";

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
    loopSuspended: false,
    step: "upkeep",
    autopassToggle: false,
    viewerIsActive: false,
    autopassPersistThroughTurns: false,
    manualStop: false,
    autoPassPriority: true,
    stackEmpty: true,
    holdPriority: false,
    autoPassOwnStack: true,
    ownsEveryStackItem: false,
    stepStop: undefined,
    smartAutoPass: true,
    alwaysStopOpponentStack: false,
    hasResponse: false,
    hasPlay: false,
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

  it("holds with autoPassPriority off", () => {
    expect(autopassDecision(gates({ autoPassPriority: false }))).toBe("hold");
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

  it("holds a pinned step that smartAutoPass would otherwise skip", () => {
    // The stops grid says stop, the predicate says "nothing to do",
    // and without the pin that combination passes. ADR 0009 §5: a pin
    // is "I want the cursor even though the engine sees no reason for
    // it" — bluffing, thinking, or an action the catalog doesn't
    // model yet.
    const g = gates({ stepStop: true, smartAutoPass: true, hasPlay: false });
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
      smartAutoPass: true,
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
      smartAutoPass: true,
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

  it("passes it with smartAutoPass off and the step unstopped", () => {
    // Nothing to review and no stop: smart-skip's own case.
    expect(autopassDecision(declareAttackers({ stepStop: false, hasPlay: false }))).toBe("pass");
  });

  it("an explicit autopass toggle still passes the window", () => {
    // Unchanged by #526: the toggle out-votes the smart-autopass
    // guard, as PR #606 stated. Only a manual pin now out-votes the
    // toggle.
    expect(autopassDecision(declareAttackers({ autopassToggle: true }))).toBe("pass");
  });

  it("smartAutoPass off keeps a stopped declare-attackers window", () => {
    expect(autopassDecision(declareAttackers({ smartAutoPass: false, hasPlay: false }))).toBe(
      "hold",
    );
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
    expect(autopassDecision(gates({ owesBlockDecision: true, autopassToggle: true }))).toBe("hold");
  });

  it("holds while the CR 726 loop breaker is up (#628)", () => {
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

  it("does not fire with the danger setting on", () => {
    expect(autopassDecision(ownMain({ autopassPersistThroughTurns: true }))).toBe("pass");
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

  it("holds a stopped step with smartAutoPass off", () => {
    expect(autopassDecision(gates({ stepStop: true, smartAutoPass: false, hasPlay: false }))).toBe(
      "hold",
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

  it("alwaysStopOpponentStack restores the old stop", () => {
    expect(autopassDecision(oppStack({ alwaysStopOpponentStack: true }))).toBe("hold");
  });

  it("smart autopass off keeps the legacy stop", () => {
    expect(autopassDecision(oppStack({ smartAutoPass: false }))).toBe("hold");
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

  it("smart autopass off ignores key windows", () => {
    const g = gates({ step: "end", oppEndWindow: true, hasResponse: true, smartAutoPass: false });
    expect(autopassDecision(g)).toBe("pass");
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

  it("still passes key windows and ticked steps with an empty stack", () => {
    const g = gates({
      autopassToggle: true,
      step: "end",
      oppEndWindow: true,
      hasResponse: true,
      stepStop: true,
      hasPlay: true,
      bluffInstant: true,
    });
    expect(autopassDecision(g)).toBe("pass");
  });

  it("ignores alwaysStopOpponentStack", () => {
    expect(autopassDecision(oppStack({ alwaysStopOpponentStack: true }))).toBe("pass");
  });
});
