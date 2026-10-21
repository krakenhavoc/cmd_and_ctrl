// attackDirection.ts — ADR 0206 (#4211): the table says which way attacks
// go before anyone tries to attack. "Each player may attack only the
// nearest opponent in the chosen direction and planeswalkers controlled
// by that opponent" (Pramikon, Sky Rampart; Teyo, Geometric Tactician's
// −2) is otherwise visible only as greyed defenders and a refusal toast.
//
// READ, NOT DERIVED. The rows are the server's `attack_directions`, each
// with the opponent every player may attack already worked out
// (`nearest`, seat id to seat id); this module words them. Which
// defenders a selected attacker may click is still the legal_actions
// digest's answer (legalActions.ts `attackTargets`). Nothing here decides
// an attack.

import type { AttackDirectionView, PlayerView } from "./protocol";

/** One rule in force, worded for the dock. */
export interface AttackDirectionLine {
  /** "left" or "right", for a data attribute and a test. */
  direction: string;
  /** The visible line: "Attacks go left: you may attack only Bob (Pramikon, Sky Rampart)". */
  text: string;
  /** The hover and screen-reader explanation, in the rule's own terms. */
  title: string;
}

const MEANING: Record<string, string> = {
  left: "the next player in turn order",
  right: "the player before in turn order",
};

/**
 * attackDirectionLines words each rule in force, in the order the server
 * lists them, then one more line when the rules disagree: with two
 * rules that send the same player's attacks to different opponents
 * nobody at a table of three or more can attack at all (the Pramikon and
 * Teyo rulings). `viewerID` is the seat looking, or null for a spectator,
 * who is told the direction and its source only. A row with a direction
 * this build does not know is left out rather than claimed.
 */
export function attackDirectionLines(
  rows: readonly AttackDirectionView[] | undefined,
  seats: readonly PlayerView[] | undefined,
  viewerID: string | null | undefined,
): AttackDirectionLine[] {
  const known = (rows ?? []).filter((r) => r.direction === "left" || r.direction === "right");
  if (known.length === 0) return [];
  const names = new Map((seats ?? []).map((s) => [s.id, s.name]));
  const out: AttackDirectionLine[] = known.map((r) => {
    const source = r.source ? ` (${r.source})` : "";
    let text = `Attacks go ${r.direction}`;
    let who = "";
    if (viewerID) {
      const target = r.nearest?.[viewerID];
      if (target) {
        const name = names.get(target) ?? "another player";
        text += `: you may attack only ${name}`;
        who = ` You may attack only ${name} and planeswalkers ${name} controls.`;
      } else {
        text += ": you have no one to attack";
        who = " You have no opponent that way, so you can't attack.";
      }
    }
    return {
      direction: r.direction,
      text: text + source,
      title:
        `Each player may attack only the nearest opponent to the ${r.direction} ` +
        `(${MEANING[r.direction]}) and planeswalkers that opponent controls` +
        (r.source ? `: ${r.source}.` : ".") +
        who,
    };
  });
  if (rulesDisagree(known)) {
    out.push({
      direction: "conflict",
      text: "These rules disagree: no one can attack",
      title:
        "Two attack-direction rules send a player's attacks to different opponents, " +
        "so nobody can attack while both are in force.",
    });
  }
  return out;
}

/** True when some player is sent to a different opponent by two rules. */
function rulesDisagree(rows: readonly AttackDirectionView[]): boolean {
  if (rows.length < 2) return false;
  const players = new Set(rows.flatMap((r) => Object.keys(r.nearest ?? {})));
  for (const p of players) {
    const targets = new Set(rows.map((r) => r.nearest?.[p] ?? ""));
    if (targets.size > 1) return true;
  }
  return false;
}
