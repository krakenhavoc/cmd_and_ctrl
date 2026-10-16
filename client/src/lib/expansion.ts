// expansion decides, per seat, whether that seat's panel renders as a
// SeatSummary read-out or as a full PlayerPanel of cards.
//
// It is the other half of ADR 0077. seatSummary.ts answers "what does
// a summary say"; this module answers "when does a summary stop being
// enough". Both are pure and unit-tested, for the same reason: a rule
// that decides what the player sees during a priority window is a rule
// that belongs in a test rather than in a template.
//
// THE ONE INVARIANT. A panel expands only because the VIEWER did
// something, or because the turn changed. Never because something
// happened on the table.
//
// That is not a stylistic preference. Expansion moves every other
// panel — reflow resizes them, overlay covers them — and a player in a
// priority window is usually mid-click on a specific card. A trigger
// resolving on an opponent's board while you are clicking your land is
// the exact moment the table must NOT move. Things that happen to you
// go to the attention strip, which is designed to be interrupted; the
// board is not.
//
// So the trigger list below is closed, and every entry on it is either
// (a) a mode the viewer entered by clicking — a targeting prompt,
// block mode, attack mode — or (b) the active player changing, which
// only happens on a turn boundary and therefore cannot land inside a
// click the viewer has already started.
//
// The reflow below (decideReflow) obeys the same invariant: it follows
// the active seat, so it changes on a turn boundary, and it never
// shrinks a seat an open prompt needs.
//
// ADR 0077 also had a pin the viewer asked for, which expanded a
// summary in place. ADR 0120 §6 retired it: a viewer who wants a seat's
// board opens it in the expanded overlay (boardExpand.ts), drawn over
// the table, which does not move the table at all.
//
// A spell being cast, a permanent entering, a trigger going on the
// stack, life totals moving: none of these expand anything. If you are
// adding a trigger here, the question to answer first is "can this fire
// while the viewer's mouse is already down?" — and if it can, it does
// not belong.

import type { CardView, GameView } from "./protocol";
import type { TableLayout } from "./tableLayout";
import { isLegalCardTarget, isLegalPlayerTarget, type TargetingState } from "./targeting";

export type SeatRendering = "summary" | "full";

/**
 * Why a seat is rendering full. Kept on the decision rather than
 * discarded because it is the only way to debug "why did the table
 * just move" after the fact, and because `"setting"` and `"self"` are
 * worth distinguishing from a live expansion when reading a snapshot.
 */
export type ExpansionReason =
  | "self"
  | "spectator"
  | "setting"
  | "targeting"
  | "blocking"
  | "attacking"
  | "active-player";

export interface SeatDecision {
  rendering: SeatRendering;
  /** Null exactly when `rendering` is `"summary"`. */
  reason: ExpansionReason | null;
}

/** Per-seat facts. Every one is derived elsewhere and passed in. */
export interface SeatSignals {
  /** The viewer's own seat. Never a summary — you need your own cards. */
  isSelf: boolean;
  isActiveSeat: boolean;
  /** A targeting prompt is live and this seat holds a legal target. */
  controlsLegalTarget: boolean;
  /** This seat controls a creature attacking the viewer. */
  hasAttackersOnViewer: boolean;
  /** The server published this seat as a legal defender for the viewer. */
  isLegalDefender: boolean;
}

export interface TableSignals {
  spectator: boolean;
  combatMode: "idle" | "attack" | "block";
}

export interface ExpansionSettings {
  opponentDetail: "summary" | "full";
  expandActivePlayer: boolean;
}

/**
 * expansionSettingsFor reads the two expansion settings through the
 * table layout. The "focus" layout is opponents as summaries in the
 * top half, so it overrides both: every opponent is a summary
 * whatever "Opponent boards" says, and the active player's board does
 * not grow on their turn — you hover or click an avatar to see it
 * (ADR 0120). The interaction triggers in decideSeatRendering still
 * apply, because a prompt you cannot click is unanswerable.
 */
export function expansionSettingsFor(
  display: ExpansionSettings & { tableLayout: TableLayout },
): ExpansionSettings {
  if (display.tableLayout === "focus") {
    return { opponentDetail: "summary", expandActivePlayer: false };
  }
  return {
    opponentDetail: display.opponentDetail,
    expandActivePlayer: display.expandActivePlayer,
  };
}

/**
 * decideSeatRendering is the whole decision, in priority order. The
 * reasons are disjoint in practice and ordered for readability.
 */
export function decideSeatRendering(
  seat: SeatSignals,
  table: TableSignals,
  settings: ExpansionSettings,
): SeatDecision {
  // Your own board is always cards. A summary of your own seat would
  // hide the hand you are about to play out of.
  if (seat.isSelf) return { rendering: "full", reason: "self" };

  // Spectators have no viewer seat to protect and no priority windows
  // to be interrupted in, so they get the old uniform grid.
  if (table.spectator) return { rendering: "full", reason: "spectator" };

  // The escape hatch. A player who chose "Full boards" gets exactly
  // the table they had before this feature existed.
  if (settings.opponentDetail === "full") return { rendering: "full", reason: "setting" };

  // The three interaction triggers. Each one exists because the
  // summary renders pips rather than cards, and a pip is not
  // something you can always finish an interaction against — see
  // SeatSummary's click routing. Without these, a targeting prompt
  // could point at a seat the viewer has no way to click.
  const live = interactionReason(seat, table);
  if (live) return { rendering: "full", reason: live };

  // Turn boundary, and opt-out-able. Unlike the three above this one
  // is a convenience rather than a requirement — nothing breaks with
  // it off, you just read a summary during someone else's turn.
  if (settings.expandActivePlayer && seat.isActiveSeat) {
    return { rendering: "full", reason: "active-player" };
  }

  return { rendering: "summary", reason: null };
}

/**
 * interactionReason is the interaction trigger, if any, that needs this
 * seat's cards on screen for a prompt the viewer has open: a targeting
 * prompt with a legal target there, block mode against its attackers,
 * or attack mode with it as a legal defender. Null when none applies.
 *
 * decideSeatRendering expands a summary for these, and decideReflow
 * never shrinks a seat to an avatar card while one applies: a prompt
 * that can only be answered by clicking a seat's card must find that
 * card drawn.
 */
export function interactionReason(
  seat: SeatSignals,
  table: TableSignals,
): "targeting" | "blocking" | "attacking" | null {
  if (seat.controlsLegalTarget) return "targeting";
  if (table.combatMode === "block" && seat.hasAttackersOnViewer) return "blocking";
  if (table.combatMode === "attack" && seat.isLegalDefender) return "attacking";
  return null;
}

// ---- Reflow (ADR 0077 §6, amended 2026-10-16, #4065) ----------------
//
// The "reflow" expand style. On an opponent's turn the active opponent's
// board takes most of the opponents' row and the other opponents shrink
// to an avatar card: name, avatar and life, the same identity the focus
// layout's summaries lead with. Hovering or clicking that avatar opens
// the seat's board in the expanded overlay (ADR 0120), so nothing is out
// of reach.
//
// It changes only on a turn boundary, because the active seat changes
// only there, and it never shrinks a seat a live prompt needs
// (interactionReason): such a seat stays a full board beside the active
// one. On the viewer's own turn there is no reflow at all, so the row is
// the usual equal boards.

/** The two expand styles (`display.expandStyle`). */
export type ExpandStyle = "reflow" | "overlay";

/** One opponent in the row. */
export interface ReflowSeat {
  id: string;
  isActiveSeat: boolean;
  /** interactionReason is not null for this seat. */
  needsBoard: boolean;
}

export interface ReflowContext {
  style: ExpandStyle;
  /** After expansionSettingsFor, so the focus layout's "off" holds. */
  expandActivePlayer: boolean;
  /**
   * Every opponent sits in the one top row: the row layout, and three
   * players in the quadrant layout. A four-player quadrant table seats
   * the next player beside you, so there is no one row to reflow.
   */
  opponentsShareRow: boolean;
  spectator: boolean;
}

export interface Reflow {
  /** The active opponent, whose board takes most of the row. */
  focusSeatID: string;
  /** The opponents drawn as avatar cards. */
  avatarSeatIDs: ReadonlySet<string>;
}

/**
 * decideReflow is the reflow, or null when the row stays equal boards:
 * the overlay style, the active-player expansion off, a spectator, no
 * single opponents' row, fewer than two opponents, or the viewer's own
 * turn. `opponents` is every opponent in the row, the viewer excluded.
 */
export function decideReflow(opponents: readonly ReflowSeat[], ctx: ReflowContext): Reflow | null {
  if (ctx.style !== "reflow" || !ctx.expandActivePlayer) return null;
  if (ctx.spectator || !ctx.opponentsShareRow) return null;
  if (opponents.length < 2) return null;
  const focus = opponents.find((o) => o.isActiveSeat);
  if (!focus) return null;
  const avatars = new Set<string>();
  for (const o of opponents) {
    if (o.id !== focus.id && !o.needsBoard) avatars.add(o.id);
  }
  return { focusSeatID: focus.id, avatarSeatIDs: avatars };
}

/** The width of an avatar card's grid column. */
export const REFLOW_AVATAR_COLUMN = "var(--reflow-avatar-w, 132px)";

/**
 * reflowColumns is the opponents' row's `grid-template-columns` under a
 * reflow, one track per seat in `columnSeatIDs` order: an avatar card
 * gets a fixed narrow column, the active opponent twice the share of
 * any other full board, and a seat a prompt needs one share.
 */
export function reflowColumns(columnSeatIDs: readonly string[], reflow: Reflow): string {
  return columnSeatIDs
    .map((id) => {
      if (reflow.avatarSeatIDs.has(id)) return REFLOW_AVATAR_COLUMN;
      return id === reflow.focusSeatID ? "minmax(0, 2fr)" : "minmax(0, 1fr)";
    })
    .join(" ");
}

/**
 * seatControlsLegalTarget answers "would a click on this seat's panel
 * complete the live prompt?" — for the seat itself (a player target)
 * or for anything it controls (a card target).
 *
 * Delegates both questions to targeting.ts rather than re-reading
 * `legal` here, so the expansion rule and the click handler can never
 * disagree about what is targetable. A prompt that can be answered by
 * clicking a seat MUST expand it, or the prompt is unanswerable.
 */
export function seatControlsLegalTarget(
  t: TargetingState | null,
  seatID: string,
  controlledCards: readonly CardView[],
): boolean {
  if (!t) return false;
  if (isLegalPlayerTarget(t, seatID)) return true;
  return controlledCards.some((c) => isLegalCardTarget(t, c.instance_id));
}

/**
 * seatHasAttackersOn reports whether this seat has a creature
 * attacking the viewer right now.
 *
 * `attacking_target` carries the DEFENDER's id, which may be a player
 * or a planeswalker or a battle; only the player case matters here,
 * because the other two are already the viewer's own permanents and
 * live in the viewer's own panel.
 */
export function seatHasAttackersOn(
  viewerID: string | null,
  controlledCards: readonly CardView[],
): boolean {
  if (!viewerID) return false;
  return controlledCards.some((c) => c.attacking_target === viewerID);
}

/**
 * legalDefenderIDs is the set of seats the viewer may declare an
 * attack against, taken from the server's published set rather than
 * re-derived.
 *
 * `turn.attack_targets` is only meaningful for the ACTIVE seat (the
 * only seat that declares attackers), so a viewer who is not the
 * active player gets an empty set and no seat expands for "attacking".
 * That is correct rather than defensive: you cannot be in attack mode
 * on someone else's turn.
 */
export function legalDefenderIDs(view: GameView, viewerID: string | null): Set<string> {
  const out = new Set<string>();
  if (!viewerID) return out;
  const active = view.seats[view.turn.active_seat];
  if (!active || active.id !== viewerID) return out;
  for (const t of view.turn.attack_targets ?? []) {
    if (t.kind === "player") out.add(t.id);
  }
  return out;
}
