// Client-side timing legality predicates.
//
// S13.3 shipped these as a reimplementation of the rules in a second
// language: sorcery-speed windows, land-drop gating, split second,
// flash, priority derivation, and type-line classification by string
// matching. That layer is exactly where "greyed out in the UI but
// accepted by the server" and "offered by the UI and then refused by
// the server" bugs come from, and it had both.
//
// S31 sub-PR 2 deletes it. The server now enumerates the seat's legal
// moves (`server/internal/legal`, ADR 0033 §1) and ships them as
// `GameView.legal_moves` for the viewer's own seat, so the question
// "can I play this card right now?" is a lookup, not a derivation.
//
// What survives, and why:
//
//   - The snapshot readers (hasPriority / isActivePlayer /
//     isMainPhase / stackEmpty). These read four fields off the
//     frame; they are not rules, and other modules consume them.
//   - canCastFromHand's target / mode / additional-cost branches.
//     Those never were a reimplementation — they read fields the
//     server stamps (legal_targets, modes, additional_cost) — and
//     they are what produces a USEFUL tooltip. The server's move
//     list is the verdict; these supply the sentence.
//   - canActivateSorcerySpeedAbility and canActivateLoyalty. Both
//     gate verbs the enumerator deliberately does NOT enumerate —
//     `activate_loyalty` is a sandbox affordance where the players
//     resolve the ability's text between themselves — so there is no
//     server list to look them up in and the CR 602.5d / 606.3
//     window has to stay here. Between them they are the only rules
//     derivation left in this file.
//
// What went: canCastFromHand's timing gate (castTimingForTypeLine /
// castableFaceTypeLines), canActivateAbility (its only caller was the
// auto-pass heuristic, which now reads legal_moves directly), and
// canPassPriority (no callers at all, in or out of production).

import type { CardView, GameView, LegalMoveView, LegalTargetsView } from "./protocol";
import { castableFaces } from "./faces";
import { sacrificeCount } from "./sacrificeCost";
import { printedCostClaimable } from "./targeting";

// Legality is a predicate result: legal=true means "the action
// would succeed if dispatched right now"; legal=false carries a
// plain-English reason for the tooltip. Reasons are NOT i18n-ised
// but are short enough to render in a tooltip without wrapping.
export interface Legality {
  legal: boolean;
  reason?: string;
}

const LEGAL: Legality = { legal: true };

function deny(reason: string): Legality {
  return { legal: false, reason };
}

// --- the move-list lookup ------------------------------------------

// movesFor returns the viewer's enumerated moves for one source card.
//
// Returns undefined — distinct from an empty array — when the server
// shipped no move list at all. That happens on every frame where the
// seat owes no decision, and on any server predating S31, and the two
// cases are indistinguishable from here. Callers MUST treat undefined
// as "no information" and stay permissive; treating it as "nothing is
// legal" would grey the whole hand every time a snapshot arrives
// without the field.
export function movesFor(
  snap: GameView | null | undefined,
  instanceID: string,
  kinds?: readonly LegalMoveView["kind"][],
): LegalMoveView[] | undefined {
  const all = snap?.legal_moves;
  if (!all) return undefined;
  return all.filter(
    (m) => m.source === instanceID && (kinds === undefined || kinds.includes(m.kind)),
  );
}

// hasNonPassMove reports whether the viewer's seat has anything to do
// beyond yielding. This is the whole of the auto-pass question, and
// it is the server's answer to it rather than a reconstruction.
//
// Undefined move list → undefined answer, for the same reason as
// movesFor: the caller decides what to do with "don't know".
export function hasNonPassMove(snap: GameView | null | undefined): boolean | undefined {
  const all = snap?.legal_moves;
  if (!all) return undefined;
  return all.some((m) => m.kind !== "pass");
}

// hasPassMove reports whether the seat's enumerated moves include a
// pass — i.e. whether yielding priority is a thing this seat can do
// on this frame. The keyboard layer (ADR 0047) reads it to decide
// whether the pass-priority key is live, for the same reason
// everything else in this file reads the move list: the alternative
// is deriving the priority rules a second time in TypeScript, and
// S31 sub-PR 2 deleted the last client that did.
//
// Undefined move list → undefined answer, same contract as movesFor
// and hasNonPassMove. `pass` is flagged `always_legal` server-side,
// so when the list exists at all this is an exact answer.
export function hasPassMove(snap: GameView | null | undefined): boolean | undefined {
  const all = snap?.legal_moves;
  if (!all) return undefined;
  return all.some((m) => m.kind === "pass");
}

// --- snapshot readers ----------------------------------------------

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
// (precombat or postcombat).
export function isMainPhase(snap: GameView | null | undefined): boolean {
  const step = snap?.turn?.step;
  return step === "precombat_main" || step === "postcombat_main";
}

// --- casting -------------------------------------------------------

// hasSatisfiableTargets reports whether a target clause has enough
// legal candidates on the board to be announced (CR 601.2c). An
// absent clause is trivially satisfiable — the spell targets
// nothing, which is always fine; "up to N" (min 0) likewise.
//
// Exported since #1157 because the ABILITY rows needed it too and had
// their own copy that stopped at "the list is empty". A clause's
// count is the whole question — an empty list is a refusal at min 1
// and a shrug at min 0 — and The Aetherspark's "+1: Attach The
// Aetherspark to up to one target creature you control" is the report
// that proved one client can hold both answers at once: the cast path
// read `min`, the ability path did not, so a loyalty ability the
// engine accepts and the enumerator offers to a BOT was greyed out
// for the human who owned it.
export function hasSatisfiableTargets(lt: LegalTargetsView | undefined): boolean {
  if (!lt) return true;
  const n = (lt.players?.length ?? 0) + (lt.cards?.length ?? 0);
  return n >= (lt.min ?? 1);
}

// canCastFromHand returns the legality of playing `card` from the
// viewer's hand or command zone right now.
//
// The VERDICT is the server's: the card is playable exactly when the
// enumerator listed a cast or land move for it. That covers timing
// (CR 307.1 / 304.1 / 305), the land drop (CR 305.2 — which the old
// client never checked at all), split second, flash, modal DFC faces,
// commander tax, and mana affordability (which the old client also
// never checked), all without a line of rules here.
//
// The REASON is still derived locally, because the move list carries
// no denials. Each branch below reads a field the server stamped, and
// none of them can flip a verdict the server disagrees with: they run
// only to explain a "no" the server already gave, or to explain a
// missing move list.
//
// #1168: the target / mode / additional-cost branches walk EVERY
// castable face (`castableFaces`, faces.ts) rather than reading the
// card's own top-level block — which is the face that HAPPENS TO BE
// UP, always face 0 for a card in hand. A modal DFC or an adventure
// card whose front half has a target clause with nothing legal to
// point at used to deny with "No legal target" even when the back
// or the Adventure half needs no target at all; a gate now denies
// only when EVERY castable face fails IT specifically, so a face that
// answers yes on its own terms stops a denial that was really about
// its sibling. This still can't flip the verdict itself — a single
// face's local pass says nothing about mana or timing, which stay the
// server's alone — only which of these branches, if any, has
// something to say. `named()` below prefixes the blocked face's own
// name once there is more than one to name.
export function canCastFromHand(
  card: CardView,
  snap: GameView | null | undefined,
  viewerID: string | null,
): Legality {
  if (!snap || !viewerID) return deny("Spectator can't cast");

  const moves = movesFor(snap, card.instance_id, ["cast", "land"]);
  if (moves && moves.length > 0) return LEGAL;

  // No move list at all. The server owes this seat no decision (or is
  // older than S31); either way we know nothing, so fall back to the
  // two coarse facts the frame does carry and otherwise stay out of
  // the way. Erring permissive is deliberate — a false "yes" costs a
  // rejected click, a false "no" costs the player a window they were
  // entitled to.
  if (moves === undefined) {
    if (!hasPriority(snap, viewerID)) return deny("Not your priority");
    if (snap.split_second_active) return deny("Split second on the stack");
    return LEGAL;
  }

  // The server said no. Everything from here is tooltip copy.
  if (!hasPriority(snap, viewerID)) return deny("Not your priority");
  if (snap.split_second_active) return deny("Split second on the stack");

  // #1168: every face a cast may choose, materialised — `[card]` for
  // the ~33,000 ordinary single-faced oracle IDs, which makes every
  // gate below read identically to the single-face version it
  // replaces. `named` prefixes a denial with the blocked face's own
  // name, but only once there is more than one candidate to name —
  // a single-faced denial reads exactly as it always has.
  const faces = castableFaces(card);
  const named = (face: CardView, reason: string): string =>
    faces.length > 1 ? `${face.name}: ${reason}` : reason;

  // A targeted spell with nothing legal to point at can't be cast
  // (CR 601.2c). A clause needs at least `min` legal candidates —
  // "two target creatures" with one creature out is uncastable; "up
  // to N" (min 0) is always castable. An alternative cost can rewrite
  // the clause away entirely (an overloaded Cyclonic Rift has none),
  // so the card is only blocked when no cost option has a satisfiable
  // one.
  //
  // #1012: "no cost option" counts the printed cost only when the
  // printed cost is one of the prices this cast may claim. The server
  // says which (`alternative_cost_required`); the client used to
  // assume the printed clause was always on the table, which reads a
  // flashback-only cast at the wrong price.
  const targetOK = (face: CardView): boolean => {
    if (!face.legal_targets) return true;
    const printedCounts = printedCostClaimable(face) && hasSatisfiableTargets(face.legal_targets);
    return (
      printedCounts ||
      (face.alternative_costs ?? []).some((a) => hasSatisfiableTargets(a.legal_targets))
    );
  };
  if (!faces.some(targetOK)) {
    const blocked = faces.find((f) => !targetOK(f)) ?? card;
    const min = blocked.legal_targets?.min ?? 1;
    return deny(named(blocked, min > 1 ? `Needs ${min} legal targets` : "No legal target"));
  }
  // A modal spell needs enough castable options to meet its minimum —
  // untargeted options always count, targeted ones only with a legal
  // target.
  const modesOK = (face: CardView): boolean => {
    if (!face.modes || face.modes.options.length === 0) return true;
    const castable = face.modes.options.filter((o) => {
      if (!o.legal_targets) return true;
      const n = (o.legal_targets.players?.length ?? 0) + (o.legal_targets.cards?.length ?? 0);
      return n >= (o.legal_targets.min ?? 1);
    }).length;
    return castable >= face.modes.min;
  };
  if (!faces.some(modesOK)) {
    const blocked = faces.find((f) => !modesOK(f)) ?? card;
    return deny(named(blocked, "No castable mode"));
  }
  // An additional cost you can't pay makes the spell uncastable
  // (CR 601.2h). "Discard a card" with an empty hand is the whole
  // case — the spell itself doesn't count, since it's on the stack by
  // the time costs are paid. `payable` is the same for every face —
  // it's the viewer's hand, not the card's — so it's computed once.
  const hand = snap.seats.find((s) => s.id === viewerID)?.hand.cards ?? [];
  const payable = hand.filter((c) => c.instance_id !== card.instance_id).length;
  const discardOK = (face: CardView): boolean =>
    payable >= (face.additional_cost?.discard_cards ?? 0);
  if (!faces.some(discardOK)) {
    const blocked = faces.find((f) => !discardOK(f)) ?? card;
    const discards = blocked.additional_cost?.discard_cards ?? 0;
    return deny(
      named(blocked, discards > 1 ? `Needs ${discards} cards to discard` : "No card to discard"),
    );
  }
  // Same rule for a sacrifice clause. The server has already filtered
  // the options to permanents this caster controls, so an empty list
  // is exactly "nothing to sacrifice" — Village Rites with an empty
  // board is uncastable, not a failed click.
  // #747: "sacrifice two creatures" with one creature is the same
  // verdict, and the reason says how many.
  const sacrificeOK = (face: CardView): boolean => {
    const opts = face.additional_cost?.sacrifice_options;
    if (!opts) return true;
    return (opts.cards?.length ?? 0) >= sacrificeCount(opts);
  };
  if (!faces.some(sacrificeOK)) {
    const blocked = faces.find((f) => !sacrificeOK(f)) ?? card;
    const need = sacrificeCount(blocked.additional_cost?.sacrifice_options);
    return deny(
      named(blocked, need > 1 ? `Needs ${need} permanents to sacrifice` : "Nothing to sacrifice"),
    );
  }

  // Nothing card-specific to say. The seat holds priority and the
  // server still did not offer this card, which leaves timing, the
  // spent land drop, and mana. The frame does tell us whether the
  // sorcery-speed window (CR 307.1) is open, and a shut window is the
  // overwhelmingly common answer for a greyed hand — a hand full of
  // sorceries and creatures during combat.
  //
  // It is a HINT, not a derivation, and it can be imprecise in one
  // direction: an unaffordable instant during combat gets told
  // "only at sorcery speed" when the real reason is the mana. The
  // verdict is right either way, which is the part that was broken
  // before; sharpening the sentence needs the server to ship a
  // denial reason alongside the move list, which it does not yet.
  if (!canActivateSorcerySpeedAbility(snap, viewerID).legal) return deny("Only at sorcery speed");
  return deny("Can't play this right now");
}

// --- activated and loyalty abilities --------------------------------

// canActivateSorcerySpeedAbility is CR 602.5d's "activate only as a
// sorcery" window, shared by every activated ability that declares
// it — equip (CR 702.6a) is the first in the catalog, and a
// planeswalker's loyalty ability answers to the same three gates
// plus two of its own (see canActivateLoyalty).
//
// S31 note: this is the LAST rules derivation left in this file, and
// it survives because its callers gate actions the server does not
// enumerate. `activate_loyalty` is a sandbox affordance — the engine
// charges the counters and enforces CR 606.3, and the players resolve
// the ability text between themselves — so `internal/legal` skips it
// by design (see its package doc on sandbox verbs) and there is no
// move list to look it up in.
//
// Advisory, like every predicate in this file: the server rejects
// with ErrSorcerySpeedRequired regardless. This exists so the menu
// row greys with a reason instead of looking available and failing.
export function canActivateSorcerySpeedAbility(
  snap: GameView | null | undefined,
  viewerID: string | null,
): Legality {
  if (!snap || !viewerID) return deny("Spectator can't activate");
  if (!hasPriority(snap, viewerID)) return deny("Not your priority");
  if (snap.split_second_active) return deny("Split second on the stack");
  if (!isMainPhase(snap)) return deny("Sorcery-speed only");
  if (!stackEmpty(snap)) return deny("Stack isn't empty");
  if (!isActivePlayer(snap, viewerID)) return deny("Not your turn");
  return LEGAL;
}

// canActivateLoyalty mirrors the engine's CR 606.3 gates: the
// sorcery-speed window plus once per turn per planeswalker. It greys
// a loyalty row in the card menu — see contextMenu.logic.ts
// `abilityBlocked` and `loyaltyAbilityItems`, its production callers.
//
// The once-per-turn bit comes off the wire: CardView carries
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
  const window = canActivateSorcerySpeedAbility(snap, viewerID);
  if (!window.legal) return window;
  if (alreadyActivated || card.loyalty_activated) return deny("Already activated this turn");
  // Source must be on the battlefield.
  const onBattlefield = snap?.battlefield?.cards?.some((c) => c.instance_id === card.instance_id);
  if (!onBattlefield) return deny("Planeswalker not on the battlefield");
  return LEGAL;
}

// loyaltyOf reads a planeswalker's current loyalty counters, which
// is the number CR 606.6 measures a −N cost against.
export function loyaltyOf(card: CardView): number {
  return card.counters?.loyalty ?? 0;
}

// canPayLoyaltyCost is CR 606.6: a cost that REMOVES N loyalty
// counters can only be activated with at least N there. A + or [0]
// cost is always payable. Returns "" when payable, otherwise the
// reason to show in the greyed row's hint.
export function canPayLoyaltyCost(card: CardView, cost: number | undefined): string {
  if (cost === undefined || cost >= 0) return "";
  const have = loyaltyOf(card);
  if (have >= -cost) return "";
  return `not enough loyalty (${have} of ${-cost})`;
}
