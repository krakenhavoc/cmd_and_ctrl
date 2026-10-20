// attackHandoff.ts — #4154: the active player's half of an attack
// declaration another player chose (Master Warcraft's "you choose which
// creatures attack this turn").
//
// The declarer chose which of the active player's creatures attack and
// proposed a target for each. The active player now decides what each
// one attacks (the card's ruling, CR 508.1b) and the costs: re-pointing
// a chosen attacker (its menu's "Re-declare attacker") charges nothing
// yet; "Confirm attacks" sends declare_attackers naming every chosen
// attacker at its target, which pays every attack tax at once; "Decline
// the costs" passes, which takes out each chosen attacker whose cost
// the player won't pay and hands the declaration back to the declarer.
// The server refuses anything else (actions/game declaration_control).
//
// Pure: the dock request is built from the frame, and Game.svelte wires
// the two sends.

import type { DockRequest } from "./dock";
import type { GameView } from "./protocol";

/** One chosen attacker and what it attacks right now. */
export interface ChosenAttack {
  attacker: string;
  target: string;
}

/**
 * attackTargetsOwed says whether `seatID` is the active player owing the
 * targets half of a chosen attack declaration.
 */
export function attackTargetsOwed(
  view: GameView | null | undefined,
  seatID: string | null,
): boolean {
  if (!view || !seatID) return false;
  if (view.attack_handoff?.stage !== "targets" || view.turn?.step !== "declare_attackers")
    return false;
  return view.seats[view.turn.active_seat]?.id === seatID;
}

/** chosenAttacks lists `seatID`'s attacking creatures, battlefield order. */
export function chosenAttacks(view: GameView, seatID: string): ChosenAttack[] {
  return (view.battlefield?.cards ?? [])
    .filter((c) => c.controller === seatID && !!c.attacking_target)
    .map((c) => ({ attacker: c.instance_id, target: c.attacking_target as string }));
}

/**
 * chosenAttacksCost is what confirming the attacks as they stand costs:
 * each attack's server-priced tax (`turn.attack_targets[].tax`), once per
 * attacker, concatenated as the server charges it. "" when free.
 */
export function chosenAttacksCost(view: GameView, attacks: ChosenAttack[]): string {
  const rows = view.turn?.attack_targets ?? [];
  return attacks.map((a) => rows.find((t) => t.id === a.target)?.tax ?? "").join("");
}

/** The declare_attackers payload that confirms the attacks as they stand. */
export function confirmAttacksParams(attacks: ChosenAttack[]): {
  attackers: ChosenAttack[];
  auto_tap: boolean;
} {
  return { attackers: attacks, auto_tap: true };
}

/**
 * attackHandoffRequest is the dock's request for the active player's
 * half, or null when `seatID` owes none.
 */
export function attackHandoffRequest(
  view: GameView | null | undefined,
  seatID: string | null,
  onConfirm: (attacks: ChosenAttack[]) => void,
  onDecline: () => void,
): DockRequest | null {
  if (!view || !seatID || !attackTargetsOwed(view, seatID)) return null;
  const h = view.attack_handoff!;
  const declarer = view.seats.find((s) => s.id === h.declarer)?.name || "Another player";
  const attacks = chosenAttacks(view, seatID);
  const cost = chosenAttacksCost(view, attacks);
  const n = attacks.length;
  return {
    rank: "step",
    label: "chosen attackers",
    tag: "attack",
    tone: "gold",
    question: `${declarer} chose ${n} attacker${n === 1 ? "" : "s"} for you`,
    detail: h.source_name,
    hint: cost
      ? `Re-declare any of them to change what it attacks. Confirming pays ${cost}; declining takes out each one whose cost you won't pay, and ${declarer} chooses again.`
      : "Re-declare any of them to change what it attacks, then confirm.",
    row: [
      {
        id: "confirm-chosen-attacks",
        label: cost ? `Confirm attacks (pay ${cost})` : "Confirm attacks",
        onPress: () => onConfirm(attacks),
        disabled: n === 0,
      },
      ...(cost
        ? [{ id: "decline-attack-costs", label: "Decline the costs", onPress: onDecline }]
        : []),
    ],
  };
}
