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

import type { CardView, GameView, LegalTargetsView } from "./protocol";
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

// hasSatisfiableTargets reports whether a target clause has enough
// legal candidates on the board to be announced (CR 601.2c). An
// absent clause is trivially satisfiable — the spell targets
// nothing, which is always fine; "up to N" (min 0) likewise.
function hasSatisfiableTargets(lt: LegalTargetsView | undefined): boolean {
  if (!lt) return true;
  const n = (lt.players?.length ?? 0) + (lt.cards?.length ?? 0);
  return n >= (lt.min ?? 1);
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
  // S22: the printed clause is not the only way to cast the card.
  // An alternative cost can rewrite it — an overloaded Cyclonic Rift
  // has no target clause at all — so the card is castable if ANY of
  // its cost options has a satisfiable one. Cards with no
  // alternative costs, which is nearly all of them, behave exactly
  // as before.
  if (card.legal_targets && !hasSatisfiableTargets(card.legal_targets)) {
    const castableSomehow = (card.alternative_costs ?? []).some((a) =>
      hasSatisfiableTargets(a.legal_targets),
    );
    if (!castableSomehow) {
      const min = card.legal_targets.min ?? 1;
      return deny(min > 1 ? `Needs ${min} legal targets` : "No legal target");
    }
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
  // S21 sub-PR 5: an additional cost you can't pay makes the spell
  // uncastable (CR 601.2h). "Discard a card" with an empty hand is
  // the whole case — the spell itself doesn't count, since it's on
  // the stack by the time costs are paid.
  const discards = card.additional_cost?.discard_cards ?? 0;
  if (discards > 0) {
    const hand = snap.seats.find((s) => s.id === viewerID)?.hand.cards ?? [];
    const payable = hand.filter((c) => c.instance_id !== card.instance_id).length;
    if (payable < discards)
      return deny(discards > 1 ? `Needs ${discards} cards to discard` : "No card to discard");
  }
  // S21 sub-PR 6: same rule for a sacrifice clause. The server has
  // already filtered the options to permanents this caster controls,
  // so an empty list is exactly "nothing to sacrifice" — Village
  // Rites with an empty board is uncastable, not a failed click.
  const sacrificeOptions = card.additional_cost?.sacrifice_options;
  if (sacrificeOptions && (sacrificeOptions.cards?.length ?? 0) === 0) {
    return deny("Nothing to sacrifice");
  }
  // ADR 0034: a modal DFC in hand is legal to PLAY if EITHER face is
  // legal right now, because the player has not chosen yet — the
  // face picker opens after this gate, not before it. Sea Gate
  // Restoration at instant speed is an illegal sorcery and a legal…
  // no, also illegal land; but Malakir Rebirth, whose front face is
  // an instant, stays castable in combat even though its land back
  // is not.
  //
  // This is the one client file whose BEHAVIOUR changes for faces.
  // Everything else keeps working untouched because the wire now
  // hands it one clean type line per face instead of a
  // concatenation — see the note on CardView.faces.
  const typeLines = castableFaceTypeLines(card);
  let lastDenial: Legality = LEGAL;
  for (const typeLine of typeLines) {
    const legality = castTimingForTypeLine(typeLine, card, snap, viewerID);
    if (legality.legal) return LEGAL;
    lastDenial = legality;
  }
  return lastDenial;
}

/**
 * castableFaceTypeLines returns the type lines the player could be
 * choosing between when playing this card from hand.
 *
 * Only a modal DFC offers a real choice (CR 712.12a). A transform
 * card is always cast as its front face (CR 712.4), and adventure /
 * split are deferred, so those all report exactly one type line —
 * which for a single-faced card is simply `card.type_line` and makes
 * this loop run once, as it always effectively did.
 */
function castableFaceTypeLines(card: CardView): string[] {
  if (card.layout === "modal_dfc" && card.faces && card.faces.length > 1) {
    return card.faces.map((f) => f.type_line ?? "");
  }
  return [card.type_line ?? ""];
}

/**
 * castTimingForTypeLine is the CR 307.1 / 305.1 / 702.8 timing gate
 * for one face. Lifted verbatim out of canCastCard so it can be run
 * once per castable face.
 */
function castTimingForTypeLine(
  rawTypeLine: string,
  card: CardView,
  snap: GameView,
  viewerID: string,
): Legality {
  const type = rawTypeLine.toLowerCase();
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

// canActivateLoyalty mirrors the engine's CR 606.5 gates: the
// sorcery-speed window plus once per turn per planeswalker. It is
// what greys a loyalty row in the card menu — see
// contextMenu.logic.ts `abilityItems`, its only production caller.
//
// Until #334 this function had NO callers at all. It was written in
// S13.1, ticked off as delivered in docs/sprints.md:733, and reached
// only by its own unit tests, while the action it gates
// (`activate_loyalty`) was not even a member of the ActionType
// union. Both halves of that are fixed here.
//
// The once-per-turn bit now comes off the wire: CardView carries
// `loyalty_activated`, stamped by the server from
// Game.LoyaltyActivatedThisTurn. `alreadyActivated` survives as an
// override for a caller that knows better (an optimistic local
// update between snapshots); it ORs with the server's flag rather
// than replacing it, so a stale `false` can never re-enable a row
// the server has already closed.
//
// Advisory, like every predicate in this file: the server re-checks.
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
  if (alreadyActivated || card.loyalty_activated) return deny("Already activated this turn");
  // Source must be on the battlefield.
  const onBattlefield = snap.battlefield?.cards?.some((c) => c.instance_id === card.instance_id);
  if (!onBattlefield) return deny("Planeswalker not on the battlefield");
  return LEGAL;
}

// loyaltyOf reads a planeswalker's current loyalty counters, which
// is the number CR 606.3 measures a −N cost against.
export function loyaltyOf(card: CardView): number {
  return card.counters?.loyalty ?? 0;
}

// canPayLoyaltyCost is CR 606.3: a cost that REMOVES N loyalty
// counters can only be activated with at least N there. A + or [0]
// cost is always payable. Returns "" when payable, otherwise the
// reason to show in the greyed row's hint.
export function canPayLoyaltyCost(card: CardView, cost: number | undefined): string {
  if (cost === undefined || cost >= 0) return "";
  const have = loyaltyOf(card);
  if (have >= -cost) return "";
  return `not enough loyalty (${have} of ${-cost})`;
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
