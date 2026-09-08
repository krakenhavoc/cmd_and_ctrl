// Client-side timing legality predicates (S13.3). The server is
// authoritative — every cast / activate / pass goes through the
// dispatcher and gets rejected with a clean error if illegal. These
// predicates are best-effort UI greying so players see the rejection
// *before* clicking, with a plain-English reason in the tooltip
// instead of an error toast after the click.
//
// Built on the engine state shipped by S13 (priority sentinel,
// step-grant table) + S13.1 (sorcery-speed window, split-second,
// loyalty-once-per-turn). Mana-source / target-legality / alt-cast-
// path predicates depend on later sprints (S15, S20, S29) and are
// out of scope here.

import type { CardView, GameView } from "./protocol";
import { grantsPriority, NO_PRIORITY_STEPS } from "./turn";

// Legality is a predicate result: legal=true means "the action
// would succeed if dispatched right now"; legal=false carries a
// plain-English reason for the tooltip. Reasons are NOT i18n-ised
// (sandbox at hour 0 of S13.3) but are short enough to render in a
// tooltip without wrapping.
export interface Legality {
  legal: boolean;
  reason?: string;
}

const LEGAL: Legality = { legal: true };

function deny(reason: string): Legality {
  return { legal: false, reason };
}

// hasPriority reports whether the given viewer ID currently holds
// priority. Returns false if the viewer ID is empty (admin /
// spectator) or if the priority holder is the no-priority sentinel
// (-1 — Untap / Cleanup steps don't grant priority per CR 502.4 /
// 514.3).
export function hasPriority(snap: GameView | null | undefined, viewerID: string | null): boolean {
  if (!snap || !viewerID) return false;
  const ph = snap.turn?.priority_holder ?? -1;
  if (ph < 0) return false;
  const seat = snap.seats?.[ph];
  return seat?.id === viewerID;
}

// isActivePlayer reports whether the viewer is the active-turn
// player. Different from hasPriority — the active player keeps
// priority into their main phases but loses it during opponents'
// instant-speed responses.
export function isActivePlayer(
  snap: GameView | null | undefined,
  viewerID: string | null,
): boolean {
  if (!snap || !viewerID) return false;
  const seat = snap.seats?.[snap.turn?.active_seat ?? -1];
  return seat?.id === viewerID;
}

// stackEmpty reports whether the stack has no items.
export function stackEmpty(snap: GameView | null | undefined): boolean {
  if (!snap) return true;
  if ((snap.stack_items?.length ?? 0) > 0) return false;
  return (snap.stack?.cards?.length ?? 0) === 0;
}

// isMainPhase reports whether the cursor is on a main-phase step
// (precombat or postcombat). Sorceries / sorcery-speed activations
// require this.
export function isMainPhase(snap: GameView | null | undefined): boolean {
  const step = snap?.turn?.step;
  return step === "precombat_main" || step === "postcombat_main";
}

// canCastFromHand returns the legality of casting `card` from the
// viewer's hand right now. Mirrors the server's CastSpell guards:
// caller holds priority, split-second clear, and (for non-instant
// non-land cards) sorcery-speed open. Lands fall through the same
// sorcery-speed gate per CR 305.
//
// Uses card.type_line classification — empty type lines (placeholder
// demo cards) are treated as non-land, non-instant, requiring
// sorcery speed. That's the conservative default: the server may
// still accept them if they're real sorceries, and the placeholder
// case is sandbox-only anyway.
export function canCastFromHand(
  card: CardView,
  snap: GameView | null | undefined,
  viewerID: string | null,
): Legality {
  if (!snap || !viewerID) return deny("Spectator can't cast");
  if (!hasPriority(snap, viewerID)) return deny("Not your priority");
  if (snap.split_second_active) return deny("Split second on the stack");
  // S20: a targeted spell with nothing legal to point at can't be
  // cast (CR 601.2c — you must choose a legal target to cast it).
  // S20 sub-PR 5: a clause needs at least `min` legal candidates
  // ("two target creatures" with one creature out is uncastable);
  // "up to N" (min 0) is always castable.
  if (card.legal_targets) {
    const lt = card.legal_targets;
    const n = (lt.players?.length ?? 0) + (lt.cards?.length ?? 0);
    const min = lt.min ?? 1;
    if (n < min) return deny(min > 1 ? `Needs ${min} legal targets` : "No legal target");
  }
  // S20 sub-PR 4: a modal spell needs enough castable options to
  // meet its minimum — untargeted options always count, targeted
  // ones only with a legal target.
  if (card.modes && card.modes.options.length > 0) {
    const castable = card.modes.options.filter((o) => {
      if (!o.legal_targets) return true;
      const n = (o.legal_targets.players?.length ?? 0) + (o.legal_targets.cards?.length ?? 0);
      return n >= (o.legal_targets.min ?? 1);
    }).length;
    if (castable < card.modes.min) return deny("No castable mode");
  }
  const type = (card.type_line ?? "").toLowerCase();
  const isLand = type.includes("land");
  const isInstant = type.includes("instant");
  // Flash (CR 702.8) lets a card be cast as if it had instant timing.
  // Server-side Abilities for hand cards are sourced from the S18
  // CatalogPrintedKeywords hook, surfaced through the same wire field
  // the battlefield uses (characteristic.go:printedCharacteristic).
  const hasFlash = (card.abilities ?? []).includes("flash");
  if (isLand) {
    if (!isMainPhase(snap)) return deny("Lands only on your main phase");
    if (!stackEmpty(snap)) return deny("Stack isn't empty");
    if (!isActivePlayer(snap, viewerID)) return deny("Not your turn");
    return LEGAL;
  }
  if (isInstant || hasFlash) {
    return LEGAL;
  }
  // Sorcery / non-instant non-land permanent. Sorcery-speed gate.
  if (!isMainPhase(snap)) return deny("Only at sorcery speed");
  if (!stackEmpty(snap)) return deny("Stack isn't empty");
  if (!isActivePlayer(snap, viewerID)) return deny("Not your turn");
  return LEGAL;
}

// canActivateAbility returns the legality of activating a generic
// ability on `card`. Sandbox: the engine doesn't know which
// abilities a card has, so the predicate just asks "could you put
// an ability on the stack right now?" — same shape as casting an
// instant (any priority window).
//
// `sorcerySpeed` callers (loyalty, sorcery-speed activated abilities
// that the player marked as such) should use `canActivateLoyalty`
// or pass the same predicate; this helper assumes instant-speed
// activations.
export function canActivateAbility(
  _card: CardView,
  snap: GameView | null | undefined,
  viewerID: string | null,
): Legality {
  if (!snap || !viewerID) return deny("Spectator can't activate");
  if (!hasPriority(snap, viewerID)) return deny("Not your priority");
  if (snap.split_second_active) return deny("Split second on the stack");
  return LEGAL;
}

// canActivateLoyalty mirrors the server's ActivateLoyalty guards:
// sorcery-speed window + once-per-turn. The once-per-turn bit
// requires server state we don't currently surface on the wire
// (Game.LoyaltyActivatedThisTurn is server-only); the predicate is
// best-effort and falls back to the server rejection if a player
// races a second activation through the dialog.
//
// `alreadyActivated` is the optional caller-tracked "has this
// planeswalker activated this turn" hint — the ability dialog can
// memoise it from the last successful activation. When unset the
// predicate optimistically returns legal (server still gates).
export function canActivateLoyalty(
  card: CardView,
  snap: GameView | null | undefined,
  viewerID: string | null,
  alreadyActivated = false,
): Legality {
  if (!snap || !viewerID) return deny("Spectator can't activate");
  if (!hasPriority(snap, viewerID)) return deny("Not your priority");
  if (snap.split_second_active) return deny("Split second on the stack");
  if (!isMainPhase(snap)) return deny("Sorcery-speed only");
  if (!stackEmpty(snap)) return deny("Stack isn't empty");
  if (!isActivePlayer(snap, viewerID)) return deny("Not your turn");
  if (alreadyActivated) return deny("Already activated this turn");
  // Source must be on the battlefield.
  const onBattlefield = snap.battlefield?.cards?.some((c) => c.instance_id === card.instance_id);
  if (!onBattlefield) return deny("Planeswalker not on the battlefield");
  return LEGAL;
}

// canPassPriority returns the legality of clicking "pass priority"
// right now. False when the cursor is on a no-priority step
// (Untap / Cleanup) OR the viewer doesn't hold priority OR a pending
// choice addressed to *anyone* is still open — advancing past a
// damage-assignment modal (CR 510.1c) would have the server reject
// with invalid-parameters, so gate it here.
export function canPassPriority(
  snap: GameView | null | undefined,
  viewerID: string | null,
): Legality {
  if (!snap || !viewerID) return deny("Spectator can't pass priority");
  const step = snap.turn?.step;
  if (step && NO_PRIORITY_STEPS.has(step as never)) {
    return deny("No one holds priority this step");
  }
  if (!grantsPriority(step)) {
    return deny("No one holds priority this step");
  }
  if (!hasPriority(snap, viewerID)) return deny("Not your priority");
  if ((snap.pending_choices ?? []).length > 0) {
    return deny("Pending choice in progress");
  }
  return LEGAL;
}
