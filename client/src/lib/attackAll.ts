// attackAll — the "attack with everything" planner behind the
// declare-attackers cluster (#318).
//
// Declaring a wide Commander board one creature at a time is the
// single most tedious thing in the client: click creature, click
// seat, repeat eleven times. This module works out, for one viewer,
// which of their creatures can actually attack right now, why the
// rest can't, and who they're allowed to swing at — so the UI can
// offer a per-opponent "attack with all N" button whose count is
// honest before the click, not a hopeful guess that produces a wall
// of server rejections.
//
// MULTI-OPPONENT: Commander is 2–4 players, so "attack all" has no
// single meaning. This module deliberately refuses to guess. It
// enumerates the attackable opponents and the caller names ONE of
// them per invocation; every eligible creature is pointed at that
// seat. Spreading an attack across several opponents is a strategic
// decision with no defensible default, so it stays a manual, per-
// creature act (the existing two-click flow and the card context
// menu both still do it). The UI renders one labelled button per
// opponent rather than a bare "attack all", so the target is always
// stated on the control the player presses.
//
// Nothing here is authoritative. The server re-checks every entry in
// game.DeclareAttackers and silently skips whatever it disagrees with,
// which matters while the printed-keyword pipeline is still being
// repaired (#317): if `abilities` is missing a `defender` or
// `vigilance` today, the worst case is a count that's off by one and
// a creature the server declines to declare — never a failed action.

import { isCreature } from "./cardTypes";
import type { CardView, GameView, PlayerView } from "./protocol";

// AttackBlocker is why a creature the viewer controls can't be added
// to an attack-with-all. "declared" is tracked separately in the plan
// because it isn't a problem — that creature is already attacking.
export type AttackBlocker = "restricted" | "tapped" | "summoning-sick" | "defender";

export const BLOCKER_LABELS: Record<AttackBlocker, string> = {
  restricted: "can't attack",
  tapped: "tapped",
  "summoning-sick": "summoning sick",
  defender: "defender",
};

export interface BlockedAttacker {
  card: CardView;
  reason: AttackBlocker;
}

export interface AttackAllPlan {
  // Creatures that would be declared by an attack-with-all right now.
  eligible: CardView[];
  // Creatures already declared this combat. Left alone — re-pointing
  // an existing attacker is a correction, and the bulk verb refuses
  // to second-guess a choice the player already made.
  declared: CardView[];
  // Creatures that can't attack, with the reason to show in the hint.
  blocked: BlockedAttacker[];
  // Opponents who may legally be attacked (seated, not the viewer,
  // not eliminated), in seat order.
  defenders: PlayerView[];
}

// attackBlocker reports why a creature can't be swept into an
// attack-with-all, or null when it can. Precedence runs from the most
// permanent reason to the most transient so the hint names the thing
// the player would have to fix first.
//
// Mirrors the server's eligibility check in game.DeclareAttackers.
// Kept deliberately strict — the single-card declare_attacker verb is
// laxer on purpose (sandbox hand-forcing), but a bulk button must not
// quietly tap a creature that had no business attacking.
export function attackBlocker(card: CardView): AttackBlocker | null {
  // S24: a "can't attack" restriction (Pacifism, Arrest, Faith's
  // Fetters) is the most permanent reason of the four and goes
  // first. The flag is READ off the wire, not derived — the server
  // computes the restriction set and this renders it, so a card the
  // bulk verb would silently skip is never counted in the button.
  if ((card.restrictions ?? []).includes("cant_attack")) return "restricted";
  if ((card.abilities ?? []).includes("defender")) return "defender";
  if (card.summoning_sick) return "summoning-sick";
  if (card.tapped) return "tapped";
  return null;
}

// planAttackAll buckets every creature the viewer controls on the
// battlefield and lists the seats they may attack. Step gating is the
// caller's job — both call sites already know whether the cursor is
// in declare_attackers.
export function planAttackAll(
  view: GameView | null | undefined,
  viewerID: string | null | undefined,
): AttackAllPlan {
  const plan: AttackAllPlan = { eligible: [], declared: [], blocked: [], defenders: [] };
  if (!view || !viewerID) return plan;

  for (const seat of view.seats ?? []) {
    if (seat.id === viewerID || seat.eliminated) continue;
    plan.defenders.push(seat);
  }

  for (const card of view.battlefield?.cards ?? []) {
    if (card.controller !== viewerID || !isCreature(card)) continue;
    if (card.attacking_target) {
      plan.declared.push(card);
      continue;
    }
    const reason = attackBlocker(card);
    if (reason) plan.blocked.push({ card, reason });
    else plan.eligible.push(card);
  }
  return plan;
}

// AttackAllParams is the wire shape of the bulk declare_attackers
// action. Mirrors the params struct in
// server/internal/actions/actions.go.
// Declared as a type alias rather than an interface so it carries an
// implicit index signature and drops straight into the context menu's
// `Record<string, unknown>` params slot without a cast.
export type AttackAllParams = {
  attackers: { attacker: string; target: string }[];
};

// attackAllParams builds the action payload aiming every eligible
// creature at one named seat. Returns null when there is nothing to
// send, so callers never fire an empty batch the server would reject.
export function attackAllParams(
  plan: AttackAllPlan,
  defenderSeatID: string,
): AttackAllParams | null {
  if (plan.eligible.length === 0) return null;
  if (!plan.defenders.some((s) => s.id === defenderSeatID)) return null;
  return {
    attackers: plan.eligible.map((c) => ({
      attacker: c.instance_id,
      target: defenderSeatID,
    })),
  };
}

// blockedSummary renders the "why not everything" hint: counts by
// reason, most-common first, e.g. "3 tapped, 1 defender". Empty
// string when nothing is blocked, so the caller can drop the clause
// entirely rather than printing a stray separator.
export function blockedSummary(blocked: readonly BlockedAttacker[]): string {
  if (blocked.length === 0) return "";
  const counts = new Map<AttackBlocker, number>();
  for (const b of blocked) counts.set(b.reason, (counts.get(b.reason) ?? 0) + 1);
  return [...counts.entries()]
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
    .map(([reason, n]) => `${n} ${BLOCKER_LABELS[reason]}`)
    .join(", ");
}

// seatLabel is the name to put on an attack-all button. Prefers the
// Discord display name the way the seat headers do.
export function seatLabel(seat: PlayerView): string {
  return seat.display_name || seat.name;
}

// attackAllLabel is the full button copy. It always names the target,
// because "attack all" on its own is ambiguous at a four-player
// table — the control the player presses has to say who gets hit.
export function attackAllLabel(plan: AttackAllPlan, seat: PlayerView): string {
  const n = plan.eligible.length;
  return `Attack ${seatLabel(seat)} with all ${n} creature${n === 1 ? "" : "s"}`;
}
