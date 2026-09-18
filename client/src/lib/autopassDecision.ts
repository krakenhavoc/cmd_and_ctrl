// The auto-pass gate chain, lifted out of Game.svelte's $effect so
// it can be tested (#526).
//
// Every rule that decides whether the client passes priority on the
// viewer's behalf used to live inline in one ~90-line `$effect`.
// That is how #526 shipped: the manual one-time stop (priorityStops
// .ts) was checked *inside* the `if (!autopass)` branch, so a pin
// was only ever consulted while the autopass toggle was off — which
// is backwards, because a pin is the one thing a player clicks
// specifically to interrupt automatic passing. A player with
// autopass on who pinned a step watched the game run straight
// through it.
//
// The component still owns the side effects (sending pass_priority,
// clearing the toggle) and the reactive reads; this module owns the
// precedence, as data in and a verdict out.
//
// PRECEDENCE, strongest first. See ADR 0009 §5 and §6.
//
//  1. Table- and viewer-level blocks that are not questions about
//     what the viewer *may* do: no priority, mulligans open, game
//     over, eliminated, an open pending choice, an owed
//     declare-blockers decision (#328), the CR 726 loop breaker
//     (#628). None of these can be out-voted by any toggle.
//  2. The autopass safety belt (ADR 0009 §7): entering the viewer's
//     own precombat_main clears the toggle instead of passing.
//  3. A manual one-time stop on this step → hold (#526). Beats the
//     autopass toggle, the stops grid, and smartAutoPass.
//  4. The autopass toggle → pass.
//  5. settings.autoPassPriority off → hold.
//  6. A non-empty stack → hold, except the #323 own-stack carve-out.
//  7. stepStops[step] === true → hold, unless smartAutoPass says
//     there is nothing to consider (#599 keeps the declare-attackers
//     review window open through this rule, via hasAnyLegalResponse).
//  8. Otherwise → pass.
//
// Rule 3 sits *below* rule 2 deliberately, and it costs nothing:
// "clear-toggle" holds the cursor too, so a pinned own-precombat_main
// still stops — it just also disarms the forgotten toggle, which is
// the whole point of the belt. Putting the pin first would let a pin
// on your own main phase keep autopass armed for the rest of the
// turn.

// AutopassGates is the fully-resolved state the decision reads. All
// fields are plain data so the caller does the reactive reads and
// this module stays pure.
export interface AutopassGates {
  // The viewer holds priority right now.
  viewerHasPriority: boolean;
  // Mulligans open, game over, or the viewer is eliminated.
  tableBusy: boolean;
  // The viewer has an unanswered pending_choice.
  hasPendingChoice: boolean;
  // #328: the viewer owes a declare-blockers decision they can act on.
  owesBlockDecision: boolean;
  // #628 / CR 726: the server suspended automatic passing table-wide.
  loopSuspended: boolean;
  // The snapshot's current step, or null/undefined if unknown.
  step: string | null | undefined;
  // The session autopass toggle.
  autopassToggle: boolean;
  // The viewer is the active player.
  viewerIsActive: boolean;
  // gameplay.autopassPersistThroughTurns — the safety-belt opt-out.
  autopassPersistThroughTurns: boolean;
  // A manual one-time stop is pinned on `step`.
  manualStop: boolean;
  // gameplay.autoPassPriority.
  autoPassPriority: boolean;
  // The stack is empty.
  stackEmpty: boolean;
  // The session hold-priority toggle (#323's escape hatch).
  holdPriority: boolean;
  // gameplay.autoPassOwnStack.
  autoPassOwnStack: boolean;
  // Every item on the stack belongs to the viewer (#323).
  ownsEveryStackItem: boolean;
  // gameplay.stepStops[step] — undefined for steps the map omits.
  stepStop: boolean | undefined;
  // gameplay.smartAutoPass.
  smartAutoPass: boolean;
  // hasAnyLegalResponse(view, viewerID) — "is there anything to
  // consider here?". Carries the #328 and #599 windows.
  hasLegalResponse: boolean;
}

// AutopassVerdict is what the component should do:
//   "hold"         — leave the cursor with the player, send nothing.
//   "pass"         — send pass_priority (subject to the caller's
//                    per-seq dedupe).
//   "clear-toggle" — disarm the autopass toggle and hold.
export type AutopassVerdict = "hold" | "pass" | "clear-toggle";

export function autopassDecision(g: AutopassGates): AutopassVerdict {
  // 1. Blocks nothing can out-vote.
  if (!g.viewerHasPriority) return "hold";
  if (g.tableBusy) return "hold";
  if (g.hasPendingChoice) return "hold";
  if (g.owesBlockDecision) return "hold";
  if (g.loopSuspended) return "hold";
  if (!g.step) return "hold";

  // 2. The safety belt (ADR 0009 §7).
  if (
    g.autopassToggle &&
    g.step === "precombat_main" &&
    g.viewerIsActive &&
    !g.autopassPersistThroughTurns
  ) {
    return "clear-toggle";
  }

  // 3. #526: a manual one-time stop is the player asking, this
  // cycle, for the cursor. It outranks the autopass toggle below as
  // well as the stops grid and smartAutoPass further down — a pin is
  // a later, narrower instruction than either. The pin is consumed
  // on the next step transition, so autopass resumes on its own
  // immediately afterwards without another click.
  if (g.manualStop) return "hold";

  // 4. The session toggle: "I'm out, stop asking."
  if (g.autopassToggle) return "pass";

  // 5-8. The conventional path.
  if (!g.autoPassPriority) return "hold";

  if (!g.stackEmpty) {
    // Anything an opponent put up is a window the viewer answers.
    // #323 carves out a stack that is entirely the viewer's own.
    if (g.holdPriority) return "hold";
    if (!g.autoPassOwnStack) return "hold";
    if (!g.ownsEveryStackItem) return "hold";
    // Fall through: casting during this step already answered the
    // "give me the cursor here" question the stops grid asks.
    return "pass";
  }

  if (g.stepStop === true) {
    // smartAutoPass is the escape hatch: a stop the viewer cannot
    // act on is dead air. The predicate errs toward stopping
    // (ADR 0009 §3), and carries #328's block window and #599's
    // declare-attackers review window.
    const canRespond = g.smartAutoPass ? g.hasLegalResponse : true;
    if (canRespond) return "hold";
  }

  return "pass";
}
