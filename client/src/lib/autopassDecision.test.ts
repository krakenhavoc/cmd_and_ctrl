import { describe, it, expect } from "vitest";

import { autopassDecision, type AutopassGates } from "./autopassDecision";

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
    hasLegalResponse: false,
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
    const g = gates({ stepStop: true, smartAutoPass: true, hasLegalResponse: false });
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
      hasLegalResponse: false,
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
// lives here is that a `true` hasLegalResponse on a stopped step
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
      // hasAnyLegalResponse returns true here via hasDeclaredAttackers
      // even though the enumerator has only `pass` left to offer.
      hasLegalResponse: true,
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
    expect(autopassDecision(declareAttackers({ stepStop: false, hasLegalResponse: false }))).toBe(
      "pass",
    );
  });

  it("an explicit autopass toggle still passes the window", () => {
    // Unchanged by #526: the toggle out-votes the smart-autopass
    // guard, as PR #606 stated. Only a manual pin now out-votes the
    // toggle.
    expect(autopassDecision(declareAttackers({ autopassToggle: true }))).toBe("pass");
  });

  it("smartAutoPass off keeps a stopped declare-attackers window", () => {
    expect(
      autopassDecision(declareAttackers({ smartAutoPass: false, hasLegalResponse: false })),
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
    expect(autopassDecision(ownMain({ autopassToggle: false, hasLegalResponse: true }))).toBe(
      "hold",
    );
  });
});

// The #323 own-stack carve-out and the stops grid, unchanged by #526
// but pinned here now that they are testable.
describe("autopassDecision — the conventional path", () => {
  it("holds a stack with an opponent's item on it", () => {
    expect(autopassDecision(gates({ stackEmpty: false, ownsEveryStackItem: false }))).toBe("hold");
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
    expect(autopassDecision(gates({ stepStop: true, hasLegalResponse: true }))).toBe("hold");
  });

  it("passes a stopped step with nothing to consider", () => {
    expect(autopassDecision(gates({ stepStop: true, hasLegalResponse: false }))).toBe("pass");
  });

  it("holds a stopped step with smartAutoPass off", () => {
    expect(
      autopassDecision(gates({ stepStop: true, smartAutoPass: false, hasLegalResponse: false })),
    ).toBe("hold");
  });

  it("passes an unstopped step even with something to consider", () => {
    expect(autopassDecision(gates({ stepStop: false, hasLegalResponse: true }))).toBe("pass");
    expect(autopassDecision(gates({ stepStop: undefined, hasLegalResponse: true }))).toBe("pass");
  });
});
