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
// PRECEDENCE, strongest first. See ADR 0009 §5, §6 and the #1307
// amendment.
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
//  4. The autopass toggle: an opponent's stack item the viewer can
//     answer → hold; one they could bluff at → bluff; else pass.
//  5. settings.autoPassPriority off → hold.
//  6. A non-empty stack: hold-priority armed → hold; entirely the
//     viewer's own → the #323 carve-out; otherwise hold if smart
//     autopass is off or alwaysStopOpponentStack is on, hold on a
//     response, bluff if armed, else pass (#1307).
//  7. Empty stack, smart autopass on, a combat or opponent's-end-step
//     key window: a response → hold; an instant bluff → bluff; else
//     fall through.
//  8. stepStops[step] === true → hold, unless smart autopass says
//     there is nothing to play (#599 keeps the declare-attackers
//     review window open through this rule, via hasPlay).
//  9. Otherwise → pass.
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
  // gameplay.alwaysStopOpponentStack — the pre-#1307 "every opponent
  // stack item stops" behaviour.
  alwaysStopOpponentStack: boolean;
  // hasResponse(view, viewerID, categories) — the viewer could answer
  // (responseWindow.ts). Mana and land never count.
  hasResponse: boolean;
  // hasPlay(view, viewerID, categories) — a ticked step has something
  // in it. Carries the #328 and #599 windows.
  hasPlay: boolean;
  // keyWindow().combat / .oppEnd. The stack key window is derived
  // here from stackEmpty and ownsEveryStackItem.
  combatWindow: boolean;
  oppEndWindow: boolean;
  // Bluffing, each already ANDed with the session arm: represent a
  // counter (opponent stack items only), represent an instant (every
  // key window). bluffManual picks the verdict's flavour.
  bluffCounter: boolean;
  bluffInstant: boolean;
  bluffManual: boolean;
}

export interface BluffVerdict {
  kind: "bluff";
  manual: boolean;
}

// AutopassVerdict is what the component should do:
//   "hold"         — leave the cursor with the player, send nothing.
//   "pass"         — send pass_priority (subject to the caller's
//                    per-seq dedupe).
//   "clear-toggle" — disarm the autopass toggle and hold.
//   bluff          — nothing to answer with, but look as if there
//                    were: hold for a while then pass (timed), or
//                    hold until the player clicks (manual).
export type AutopassVerdict = "hold" | "pass" | "clear-toggle" | BluffVerdict;

export function isBluff(v: AutopassVerdict): v is BluffVerdict {
  return typeof v === "object" && v.kind === "bluff";
}

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

  const stackOpp = !g.stackEmpty && !g.ownsEveryStackItem;
  const bluff: BluffVerdict = { kind: "bluff", manual: g.bluffManual };

  // 4. The session toggle: "I'm out, stop asking" — except for an
  // opponent's spell the viewer can actually answer (#1307). The
  // toggle is standing intent from a player who thought they were
  // tapped out; a live counterspell says otherwise.
  if (g.autopassToggle) {
    if (stackOpp && g.hasResponse) return "hold";
    if (stackOpp && (g.bluffCounter || g.bluffInstant)) return bluff;
    return "pass";
  }

  // 5-9. The conventional path.
  if (!g.autoPassPriority) return "hold";

  if (!g.stackEmpty) {
    if (g.holdPriority) return "hold";
    // #323: a stack that is entirely the viewer's own. Casting during
    // this step already answered the "give me the cursor here"
    // question the stops grid asks.
    if (g.ownsEveryStackItem) return g.autoPassOwnStack ? "pass" : "hold";
    // Something an opponent put up. Before #1307 this always held;
    // smart autopass now holds only when there is an answer.
    if (!g.smartAutoPass || g.alwaysStopOpponentStack) return "hold";
    if (g.hasResponse) return "hold";
    if (g.bluffCounter || g.bluffInstant) return bluff;
    return "pass";
  }

  // 7. Empty-stack key windows: attackers are in, or it is an
  // opponent's end step. Worth a stop for a real response even when
  // the step is not ticked.
  if (g.smartAutoPass && (g.combatWindow || g.oppEndWindow)) {
    if (g.hasResponse) return "hold";
    if (g.bluffInstant) return bluff;
  }

  // 8. The stops grid. smartAutoPass is the escape hatch: a stop the
  // viewer cannot act on is dead air. hasPlay counts lands and
  // sorcery-speed casts, and carries #328's block window and #599's
  // declare-attackers review window.
  if (g.stepStop === true) {
    const canAct = g.smartAutoPass ? g.hasPlay : true;
    if (canAct) return "hold";
  }

  return "pass";
}
