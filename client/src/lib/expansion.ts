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
  if (seat.controlsLegalTarget) return { rendering: "full", reason: "targeting" };
  if (table.combatMode === "block" && seat.hasAttackersOnViewer) {
    return { rendering: "full", reason: "blocking" };
  }
  if (table.combatMode === "attack" && seat.isLegalDefender) {
    return { rendering: "full", reason: "attacking" };
  }

  // Turn boundary, and opt-out-able. Unlike the three above this one
  // is a convenience rather than a requirement — nothing breaks with
  // it off, you just read a summary during someone else's turn.
  if (settings.expandActivePlayer && seat.isActiveSeat) {
    return { rendering: "full", reason: "active-player" };
  }

  return { rendering: "summary", reason: null };
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
