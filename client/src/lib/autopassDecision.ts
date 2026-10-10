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
// PRECEDENCE, strongest first. See ADR 0009 §5, §6, the #1307
// amendment, and ADR 0143 §2.5, which replaces rules 5-7.
//
//  1. Table- and viewer-level blocks that are not questions about
//     what the viewer *may* do: no priority, mulligans open, game
//     over, eliminated, an open pending choice, an owed
//     declare-blockers decision (#328), an owed attack requirement
//     (#1571), the CR 732 loop breaker (#628). None of these can be out-voted by any toggle.
//  2. The autopass safety belt (ADR 0009 §7): entering the viewer's
//     own precombat_main clears the toggle instead of passing.
//  3. A manual one-time stop on this step → hold (#526). Beats the
//     autopass toggle, the stops grid, and the pass mode.
//  4. The autopass toggle: an opponent's stack item the viewer can
//     answer → hold; one they could bluff at → bluff; else pass.
//  5. passMode "manual" → hold (ADR 0143 §2.1).
//  6. A non-empty stack: hold-priority armed → hold; entirely the
//     viewer's own → the #323 carve-out; passMode "careful" → hold;
//     a response → hold; a bluff if armed (Smart only); else pass
//     (#1307).
//  7. Empty stack, a combat or opponent's-end-step key window: a
//     response → hold; an instant bluff (Smart only) → bluff; else
//     fall through. ADR 0143 §2.2: this runs in Smart and Careful
//     alike, with no switch, so no setting can cost the viewer a
//     window they can respond in.
//  8. The active player's column ticks this step (ADR 0143 §2.3:
//     stepStops on the viewer's turn, stepStopsOpponents otherwise;
//     stepStopFor) → hold, unless "only when I can act"
//     (#2871, stepStopsOnlyWhenCanAct) says there is nothing to play
//     (#599 keeps the declare-attackers review window open through
//     this rule, via hasPlay). On the
//     viewer's own main phase, a spell left out for mana alone while
//     a manual mana source is out counts as something to play (ADR
//     0118 owner decision 8, engineMayMissMana).
//  9. Otherwise → pass.
//
// Bluffing is Smart's alone (ADR 0143 §4.3). Careful stops at every
// opponent item anyway, Manual stops everywhere, and the dock's bluff
// chip is disabled in both, so the decision never bluffs there either.
//
// Rule 3 sits *below* rule 2 deliberately, and it costs nothing:
// "clear-toggle" holds the cursor too, so a pinned own-precombat_main
// still stops — it just also disarms the forgotten toggle, which is
// the whole point of the belt. Putting the pin first would let a pin
// on your own main phase keep autopass armed for the rest of the
// turn.

import type { PassMode } from "./settings";
import { stopKeyFor, type StepID } from "./turn";

/**
 * stepStopFor reads rule 8's tick for this step from the column for
 * whoever is active (ADR 0143 §2.3): the My-turn column (stepStops)
 * on the viewer's own turn, the Opponents'-turns column on everyone
 * else's. The two combat damage steps share one stop (turn.ts
 * stopKeyFor). Undefined when there is no step or the map omits it.
 */
export function stepStopFor(
  columns: {
    stepStops: Record<string, boolean>;
    stepStopsOpponents: Record<string, boolean>;
  },
  step: string | null | undefined,
  viewerIsActive: boolean,
): boolean | undefined {
  if (!step) return undefined;
  const column = viewerIsActive ? columns.stepStops : columns.stepStopsOpponents;
  return column[stopKeyFor(step as StepID)];
}

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
  // #1571: the viewer owes an attack a CR 508.1d requirement asks for
  // (a creature stamped must_attack); the server refuses their pass.
  owesAttackRequirement: boolean;
  // #628 / CR 732: the server suspended automatic passing table-wide.
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
  // gameplay.passMode (ADR 0143 §2.1): Smart, Careful or Manual.
  passMode: PassMode;
  // The stack is empty.
  stackEmpty: boolean;
  // The session hold-priority toggle (#323's escape hatch).
  holdPriority: boolean;
  // gameplay.autoPassOwnStack.
  autoPassOwnStack: boolean;
  // Every item on the stack belongs to the viewer (#323).
  ownsEveryStackItem: boolean;
  // The active player's column at this step (stepStopFor): stepStops
  // on the viewer's own turn, stepStopsOpponents on anyone else's.
  // Undefined for steps the map omits.
  stepStop: boolean | undefined;
  // gameplay.stepStopsOnlyWhenCanAct (#2871): a ticked step stops only
  // when hasPlay or engineMayMissMana says there is something to do.
  stepStopsOnlyWhenCanAct: boolean;
  // hasResponse(view, viewerID, categories) — the viewer could answer
  // (responseWindow.ts). Mana and land never count.
  hasResponse: boolean;
  // hasPlay(view, viewerID, categories) — a ticked step has something
  // in it. Carries the #328 and #599 windows.
  hasPlay: boolean;
  // engineMayMissMana(view, viewerID) — the viewer's own main phase,
  // a spell in hand the move list leaves out for mana alone, and a
  // manual mana source on the battlefield (ADR 0118 owner decision 8,
  // engineMayMissMana.ts). Holds rule 8 under only-when-can-act.
  engineMayMissMana: boolean;
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
  if (g.owesAttackRequirement) return "hold";
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
  // well as the stops grid and the pass mode further down — a pin is
  // a later, narrower instruction than either. The pin is consumed
  // on the next step transition, so autopass resumes on its own
  // immediately afterwards without another click.
  if (g.manualStop) return "hold";

  const stackOpp = !g.stackEmpty && !g.ownsEveryStackItem;
  const bluff: BluffVerdict = { kind: "bluff", manual: g.bluffManual };
  // ADR 0143 §4.3: only Smart bluffs, under the toggle as well. The
  // bluff chip is disabled in Careful and Manual, and a bluff left armed
  // from an earlier game must not slip through it.
  const smart = g.passMode === "smart";

  // 4. The session toggle: "I'm out, stop asking" — except for an
  // opponent's spell the viewer can actually answer (#1307). The
  // toggle is standing intent from a player who thought they were
  // tapped out; a live counterspell says otherwise.
  if (g.autopassToggle) {
    if (stackOpp && g.hasResponse) return "hold";
    if (stackOpp && smart && (g.bluffCounter || g.bluffInstant)) return bluff;
    return "pass";
  }

  // 5. Manual: every priority window waits for a click.
  if (g.passMode === "manual") return "hold";

  if (!g.stackEmpty) {
    if (g.holdPriority) return "hold";
    // #323: a stack that is entirely the viewer's own. Casting during
    // this step already answered the "give me the cursor here"
    // question the stops grid asks.
    if (g.ownsEveryStackItem) return g.autoPassOwnStack ? "pass" : "hold";
    // Something an opponent put up. Careful sees every one; Smart
    // holds only when there is an answer (#1307).
    if (g.passMode === "careful") return "hold";
    if (g.hasResponse) return "hold";
    if (smart && (g.bluffCounter || g.bluffInstant)) return bluff;
    return "pass";
  }

  // 7. Empty-stack key windows: attackers are in, or it is an
  // opponent's end step. Worth a stop for a real response even when
  // the step is not ticked. Always on in Smart and Careful (ADR 0143
  // §2.2): before it, turning smart auto-pass off to "see everything"
  // quietly dropped these stops.
  if (g.combatWindow || g.oppEndWindow) {
    if (g.hasResponse) return "hold";
    if (smart && g.bluffInstant) return bluff;
  }

  // 8. The stops grid. stepStopsOnlyWhenCanAct (#2871; until then
  // part of the old smartAutoPass setting) is the escape hatch: a stop the viewer
  // cannot act on is dead air. hasPlay counts lands and
  // sorcery-speed casts, and carries #328's block window and #599's
  // declare-attackers review window.
  //
  // ADR 0118 owner decision 8 ("stop if the engine may be wrong"): on
  // the viewer's own main phase, a spell the move list leaves out for
  // mana alone is not dead air while a manual mana source is out. The
  // engine may simply not see that mana, and Cast anyway exists for
  // exactly that, so the stop holds.
  if (g.stepStop === true) {
    const canAct = g.stepStopsOnlyWhenCanAct ? g.hasPlay || g.engineMayMissMana : true;
    if (canAct) return "hold";
  }

  return "pass";
}
