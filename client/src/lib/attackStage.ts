// attackStage.ts — declaring attackers by toggling creatures (#3266).
//
// A click on a creature that can attack no longer declares it on the
// spot. It is STAGED: drawn as attacking, with nothing on the wire, and
// a second click takes it back out. The dock's primary then reads
// "Attack with N" and sends the whole staged set as ONE declare_attackers
// (one undo entry, one broadcast; the server takes a target per
// attacker, so a split attack is one declaration too). With nothing
// staged the primary reads No attacks, as before (#3267).
//
// Where a staged creature attacks (owner decision, the mockup's option
// 2): a creature with exactly one legal attack target is staged at it
// on the first click (every duel without a planeswalker or battle);
// otherwise the click selects it and a click on a player, planeswalker
// or battle stages it there, as the old two-click flow declared it.
// "Attack all", token piles and the attackers picker still declare at
// once; they replace whatever was staged.
//
// All pure: Game.svelte owns the state.

import type { AttackerSelection } from "./exert";
import { exertUnanswered } from "./exert";

export interface StagedAttack {
  attacker: string;
  target: string;
  exert?: true;
}

/** isStaged reports whether a creature is in the staged set. */
export function isStaged(stage: readonly StagedAttack[], cardID: string): boolean {
  return stage.some((s) => s.attacker === cardID);
}

/** unstage takes a creature out of the staged set. */
export function unstage(stage: readonly StagedAttack[], cardID: string): StagedAttack[] {
  return stage.filter((s) => s.attacker !== cardID);
}

/**
 * stageAt puts the selected creature in the staged set aimed at
 * `target`, replacing any earlier aim. Null while the exert question
 * is still open: nothing is staged until the player answers it.
 */
export function stageAt(
  stage: readonly StagedAttack[],
  sel: AttackerSelection,
  target: string,
): StagedAttack[] | null {
  if (exertUnanswered(sel)) return null;
  const entry: StagedAttack =
    sel.exert === true
      ? { attacker: sel.cardID, target, exert: true }
      : { attacker: sel.cardID, target };
  return [...unstage(stage, sel.cardID), entry];
}

/**
 * soleTarget is where a first click stages a creature without asking:
 * its one legal attack target, or null when it has several (or none)
 * and the player must point it at one.
 */
export function soleTarget(targets: readonly string[]): string | null {
  return targets.length === 1 ? targets[0]! : null;
}

/**
 * pruneStage drops staged creatures that can no longer attack (one that
 * died, was tapped or was declared some other way), keeping the rest in
 * order. Returns the same array when nothing changed.
 */
export function pruneStage(
  stage: StagedAttack[],
  canAttack: (cardID: string) => boolean,
): StagedAttack[] {
  const kept = stage.filter((s) => canAttack(s.attacker));
  return kept.length === stage.length ? stage : kept;
}

export interface StagedAttackParams {
  attackers: StagedAttack[];
  auto_tap: true;
  // #4198: Phyrexian symbols of the attack tax paid with 2 life each.
  phyrexian_life?: number;
}

/**
 * stagedParams is the one declare_attackers the primary sends, or null
 * with nothing staged. auto_tap pays a CR 508.1h attack tax from lands
 * (ADR 0080), as every other declaration does; it is inert without one.
 */
export function stagedParams(
  stage: readonly StagedAttack[],
  phyrexianLife = 0,
): StagedAttackParams | null {
  if (stage.length === 0) return null;
  const params: StagedAttackParams = { attackers: stage.map((s) => ({ ...s })), auto_tap: true };
  if (phyrexianLife > 0) params.phyrexian_life = phyrexianLife;
  return params;
}
