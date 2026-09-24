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
import { ErrorCode } from "./protocol";
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
  // ADR 0080 (#1063): let the server tap lands for the CR 508.1a
  // attack tax when the mana pool alone cannot cover it. Always sent,
  // for the reason the cast chain sends it after its preview — a
  // declaration the player has confirmed at a price they were shown
  // should not then fail because the mana is in lands rather than in
  // the pool. It is inert at a table with no attack tax on it.
  auto_tap: true;
  // #1162: permanents the auto-tapper must not reach for when paying
  // the tax — the attack-tax picker's lock-a-land toggle
  // (AttackDeclarationModal.svelte), the declaration-shaped sibling
  // of the cast modal's `locked_sources`. Omitted rather than sent
  // empty, matching every other caller of this field on the wire.
  locked_sources?: string[];
};

// AttackAllOptions narrows a declaration built from an AttackAllPlan.
// Both fields are additive over the "attack with everything" default:
// omitting `only` keeps every eligible creature, and omitting
// `lockedSources` keeps the auto-tapper unrestricted.
export interface AttackAllOptions {
  // #1162: restrict the declaration to exactly these instance IDs
  // rather than every eligible creature — the attack-tax picker's
  // chosen subset, offered when "attack with all" is refused for want
  // of the CR 508.1a tax (ADR 0080) and the player has to choose which
  // attacks to pay for instead. An ID the plan no longer considers
  // eligible (a stale selection after a snapshot changed the board) is
  // dropped rather than sent, the same silent-skip posture
  // declare_attackers itself takes on the server.
  only?: readonly string[];
  // #1162: locked sources for the same declaration, carried straight
  // through to AttackAllParams.locked_sources.
  lockedSources?: readonly string[];
}

// attackAllParams builds the action payload aiming every eligible
// creature — or, with `opts.only`, an explicit subset of them — at one
// named seat. Returns null when there is nothing to send, so callers
// never fire an empty batch the server would reject.
export function attackAllParams(
  plan: AttackAllPlan,
  defenderSeatID: string,
  opts: AttackAllOptions = {},
): AttackAllParams | null {
  if (!plan.defenders.some((s) => s.id === defenderSeatID)) return null;
  const attackers = opts.only
    ? plan.eligible.filter((c) => opts.only!.includes(c.instance_id))
    : plan.eligible;
  if (attackers.length === 0) return null;
  const params: AttackAllParams = {
    attackers: attackers.map((c) => ({
      attacker: c.instance_id,
      target: defenderSeatID,
    })),
    auto_tap: true,
  };
  if (opts.lockedSources && opts.lockedSources.length > 0) {
    params.locked_sources = [...opts.lockedSources];
  }
  return params;
}

// attackTaxOn is the CR 508.1a price of attacking one seat, read off
// the turn's server-priced attack_targets (ADR 0080, #1063). "" when
// attacking it is free, or when the field is absent — an older server,
// or a step that is not declare_attackers.
//
// PER CREATURE. The wire field is a per-target flat rate because every
// printed attack tax charges the same for every creature; a whole
// declaration pays it once per attacker, which is what
// attackAllTaxLabel spells out.
export function attackTaxOn(view: GameView | null | undefined, defenderSeatID: string): string {
  const row = view?.turn?.attack_targets?.find(
    (t) => t.kind === "player" && t.id === defenderSeatID,
  );
  return row?.tax ?? "";
}

// attackTaxLabelForCount is attackAllTaxLabel's arithmetic,
// parameterised by an arbitrary attacker count rather than
// `plan.eligible.length` — factored out for #1162's subset picker,
// whose running total changes as the player checks attackers on and
// off rather than always naming the whole eligible set.
//
// The TOTAL is not re-derived from the per-creature string — it is
// that string repeated, which is exactly what the server charges and
// concatenates (three attackers into Propaganda is "{2}{2}{2}"). No
// arithmetic here means no way for the client to disagree with the
// price it is about to commit the player to.
export function attackTaxLabelForCount(each: string, n: number): string {
  if (!each || n <= 0) return "";
  if (n === 1) return `costs ${each}`;
  return `costs ${each} each, ${each.repeat(n)} for all ${n}`;
}

// attackAllTaxLabel is the clause the attack-all control appends when
// the target taxes attacks: "costs {2} each, {2}{2}{2} for all 3".
// Empty when the attack is free, so the caller drops the clause rather
// than printing a stray separator.
export function attackAllTaxLabel(
  view: GameView | null | undefined,
  plan: AttackAllPlan,
  defenderSeatID: string,
): string {
  return attackTaxLabelForCount(attackTaxOn(view, defenderSeatID), plan.eligible.length);
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

// ---- #1533: attack-with-all under a CR 508.1c count limit ----------
//
// Silent Arbiter ("no more than one creature can attack each combat")
// and Crawlspace ("no more than two creatures can attack you each
// combat") make the bulk verb refuse an over-full swing WHOLE, with
// `illegal_attack` / `attack_limit`: which creatures stay home is the
// attacking player's choice (ADR 0045 Decision 44). The answer is the
// attack-tax picker (AttackDeclarationModal.svelte), capped at the
// room the server publishes per target. Nothing here derives a limit.
// Who a limit protects and how many are already attacking are the
// engine's to count (ADR 0045 §6), and `attack_targets[].attack_limit`
// is that count.

// attackLimitOn is how many MORE creatures may be declared attacking
// one seat this combat, read off the turn's attack_targets (ADR 0045
// Decision 46). null when no limit counts that seat, or when the field
// is absent (an older server, or a step that is not declare_attackers).
// 0 is a real answer (the limit is used up), so callers test for
// null, never for truthiness.
export function attackLimitOn(
  view: GameView | null | undefined,
  defenderSeatID: string,
): number | null {
  const row = view?.turn?.attack_targets?.find(
    (t) => t.kind === "player" && t.id === defenderSeatID,
  );
  return row?.attack_limit ?? null;
}

// attackLimitBinds reports whether a limit keeps "attack with all"
// from sending every eligible creature at one seat: the room is
// smaller than the eligible set.
export function attackLimitBinds(
  view: GameView | null | undefined,
  plan: AttackAllPlan,
  defenderSeatID: string,
): boolean {
  const room = attackLimitOn(view, defenderSeatID);
  return room !== null && room < plan.eligible.length;
}

// offersAttackPicker is whether the attack-all cluster shows the
// "choose attackers…" control beside a seat: when attacking it is
// taxed (#1162, the seat may afford only some of the swing), or when a
// limit binds and still has room (#1533, only some of it may attack at
// all). A used-up limit (room 0) leaves nothing to choose.
export function offersAttackPicker(
  view: GameView | null | undefined,
  plan: AttackAllPlan,
  defenderSeatID: string,
): boolean {
  if (attackTaxOn(view, defenderSeatID)) return true;
  return attackLimitBinds(view, plan, defenderSeatID) && attackLimitOn(view, defenderSeatID) !== 0;
}

// seedAttackSelection is what the picker checks when it opens: every
// eligible creature or, under a limit, the first `cap` of them, so the
// confirm button is live at once and sends a declaration the server
// accepts. Board order, the order the rows are listed in.
export function seedAttackSelection(eligible: readonly CardView[], cap: number | null): string[] {
  const ids = eligible.map((c) => c.instance_id);
  return cap === null ? ids : ids.slice(0, Math.max(0, cap));
}

// toggleAttackSelection checks or unchecks one attacker, refusing to
// check one more than `cap` allows. The picker disables those rows as
// well; this is the rule it disables them by, kept in one place.
export function toggleAttackSelection(
  selected: readonly string[],
  id: string,
  cap: number | null,
): string[] {
  if (selected.includes(id)) return selected.filter((x) => x !== id);
  if (cap !== null && selected.length >= cap) return [...selected];
  return [...selected, id];
}

// attackLimitSentence is the picker's explanation when it opens without
// the server's refusal sentence (from the proactive "choose attackers…"
// control): the number and the seat, both read off the view.
export function attackLimitSentence(room: number, seatName: string): string {
  if (room <= 0) return `No more creatures can attack ${seatName} this combat.`;
  return `Only ${room} more creature${room === 1 ? "" : "s"} can attack ${seatName} this combat.`;
}

// BulkAttackAttempt is the one "attack with all" declaration that can
// be in flight: the seat it aimed at, and the action frame id
// sendAction returned for it.
export interface BulkAttackAttempt {
  defenderSeatID: string;
  frameID: string;
}

// BulkAttackRefusal is a refusal of that declaration the client can
// answer with the attackers picker: an unpaid tax (#1162) or a count
// limit (#1533).
export interface BulkAttackRefusal {
  kind: "tax" | "limit";
  defenderSeatID: string;
}

// bulkAttackRefusal claims the current error for the last "attack with
// all" only when it is the server's answer to THAT frame (an error
// frame echoes the refused action's id). Anything else stays with the
// generic toast. Under Silent Arbiter the usual refusal is a single
// click-declared attacker, and offering a picker for a bulk swing sent
// earlier would answer the wrong action.
export function bulkAttackRefusal(
  err: { code: string; reason?: string; replyTo?: string } | null | undefined,
  attempt: BulkAttackAttempt | null | undefined,
): BulkAttackRefusal | null {
  if (!err || !attempt || !err.replyTo || err.replyTo !== attempt.frameID) return null;
  if (err.code === ErrorCode.AttackTaxUnpaid) {
    return { kind: "tax", defenderSeatID: attempt.defenderSeatID };
  }
  if (err.code === ErrorCode.IllegalAttack && err.reason === "attack_limit") {
    return { kind: "limit", defenderSeatID: attempt.defenderSeatID };
  }
  return null;
}
