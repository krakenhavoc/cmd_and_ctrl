// attackDefault.ts — what the declare-attackers step starts with (#3586).
//
// The gameplay setting "When you declare attackers" is one of
//   ask   today's behaviour: nothing staged, the primary reads No attacks.
//   all   pre-stage every creature the server lets attack, but ONLY with
//         exactly one attackable opponent. With two or more the client
//         never guesses a defender (attackAll.ts), so the step opens empty.
//         Nothing is sent: the player still presses Attack with N.
//   none  send the empty declaration (the "No attacks" primary, a plain
//         pass_priority) when the step opens, unless a stop, a priority
//         hold or an open combat prompt says the player wants to look.
//
// All pure: Game.svelte owns the state and the once-per-step latch.

import { eligibleAt, type AttackAllPlan } from "./attackAll";
import type { StagedAttack } from "./attackStage";
import type { AttackDefaultMode } from "./settings";

export interface AttackDefaultInput {
  mode: AttackDefaultMode;
  /** The viewer holds priority in their own declare-attackers step. */
  canDeclare: boolean;
  plan: AttackAllPlan;
  /** A stop on declare attackers: the stops grid or a manual pin. */
  stopped: boolean;
  /** The session "hold priority" toggle is on. */
  holding: boolean;
  /** An exert, attack-tax or requirement prompt, or a pending choice, is open. */
  promptOpen: boolean;
}

export type AttackDefaultAction =
  | { kind: "stage"; stage: StagedAttack[] }
  | { kind: "declare-none" }
  | null;

/**
 * attackDefaultAction is what to do once for this declare-attackers
 * step, or null for nothing.
 */
export function attackDefaultAction(i: AttackDefaultInput): AttackDefaultAction {
  if (!i.canDeclare) return null;
  // Already declared (a reload mid-step) or nothing that can attack.
  if (i.plan.declared.length > 0 || i.plan.eligible.length === 0) return null;
  if (i.mode === "all") {
    const stage = preStage(i.plan);
    return stage.length > 0 ? { kind: "stage", stage } : null;
  }
  if (i.mode === "none") {
    if (i.stopped || i.holding || i.promptOpen) return null;
    return { kind: "declare-none" };
  }
  return null;
}

/**
 * preStage is the eligible set aimed at the one attackable opponent,
 * or empty with two or more (or none). No exert is chosen: the stage
 * never answers that question for the player, as attackAllParams
 * does not either. Every creature the server lists as able to attack
 * is kept, so one that must attack (CR 508.1d) is never dropped.
 */
export function preStage(plan: AttackAllPlan): StagedAttack[] {
  const only = plan.defenders[0];
  if (plan.defenders.length !== 1 || !only) return [];
  return eligibleAt(plan, only.id).map((c) => ({ attacker: c.instance_id, target: only.id }));
}

/**
 * attackStepKey identifies one declare-attackers step, so the default
 * fires once per step. Null outside it. `seq` is the turn's own
 * sequence, so an extra turn with the same number gets its own key.
 */
export function attackStepKey(
  turn: { seq: number; number: number; step: string } | null | undefined,
): string | null {
  if (!turn || turn.step !== "declare_attackers") return null;
  return `${turn.seq}:${turn.number}`;
}

export interface AttackDefaultStep {
  /** The latch to keep: the step already handled, or null outside one. */
  latch: string | null;
  action: AttackDefaultAction;
}

/**
 * attackDefaultStep is the once-per-step rule. `latch` is the key of
 * the last step this ran for. The first frame of a step on which the
 * viewer could declare (and the setting is not "ask") takes the latch
 * whatever the action is, so a player who un-stages is not re-staged and
 * a refused empty declaration is not sent again. Outside the step the
 * latch clears, so an extra combat gets its own.
 */
export function attackDefaultStep(
  latch: string | null,
  key: string | null,
  input: AttackDefaultInput,
): AttackDefaultStep {
  if (!key) return { latch: null, action: null };
  if (latch === key) return { latch, action: null };
  if (!input.canDeclare || input.mode === "ask") return { latch, action: null };
  return { latch: key, action: attackDefaultAction(input) };
}
