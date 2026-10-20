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
//
// The one table event that does move the row (ADR 0077 as amended
// 2026-10-18, #4115): under a reflow, a seat a stack item targets is a
// full board beside the active one, so the target can be seen. It
// answers that question by waiting: Board passes the targeted seats
// through settleGate.ts, which holds the change while a pointer button
// is down and until the pointer has been still for a moment.

import type { CardView, GameView } from "./protocol";
import type { TableLayout } from "./tableLayout";
import { isLegalCardTarget, isLegalPlayerTarget, type TargetingState } from "./targeting";
import { L } from "./labels";

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
// to an avatar card: the identity the focus layout's summaries lead
// with, and since #4102 a compact read-out of hand, mana, creatures and
// piles. The cards stack in one column at the row's left end, in the
// order the seats play (reflowAvatarOrder, reflowGrid; ADR 0077 as
// amended 2026-10-17). Hovering or clicking that avatar opens the seat's
// board in the expanded overlay (ADR 0120), so nothing is out of reach.
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
  /**
   * A stack item targets this seat or a permanent it controls
   * (stackTargetSeatIDs, after Board's settle gate; #4115).
   */
  targeted?: boolean;
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
  /**
   * A targeted seat is a full board beside the focus, so the focus
   * gives up its second share and the boards split the row evenly
   * (#4115). Absent means false.
   */
  evenSplit?: boolean;
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
  let evenSplit = false;
  for (const o of opponents) {
    if (o.id === focus.id) continue;
    if (o.targeted) evenSplit = true;
    else if (!o.needsBoard) avatars.add(o.id);
  }
  return { focusSeatID: focus.id, avatarSeatIDs: avatars, evenSplit };
}

/**
 * stackTargetSeatIDs is every seat a stack item targets (ADR 0077 as
 * amended 2026-10-18, #4115): a targeted player, and the controller of
 * a targeted permanent. A target off the battlefield (a spell on the
 * stack, a card in a graveyard) names no seat's board, so it is left
 * out, as are "self" and "none".
 */
export function stackTargetSeatIDs(view: GameView): Set<string> {
  const seats = new Set<string>();
  const controllerOf = new Map(
    (view.battlefield?.cards ?? []).map((c) => [c.instance_id, c.controller ?? c.owner]),
  );
  for (const item of view.stack_items ?? []) {
    for (const t of item.targets ?? []) {
      if (!t.id) continue;
      if (t.kind === "player") seats.add(t.id);
      else if (t.kind === "card") {
        const controller = controllerOf.get(t.id);
        if (controller) seats.add(controller);
      }
    }
  }
  return seats;
}

/** The width of the avatar cards' column: a token, 176px (app.css). */
export const REFLOW_AVATAR_COLUMN = "var(--reflow-avatar-w)";

/** One opponent's cell in the reflow's grid. */
export interface ReflowCell {
  id: string;
  /** The seat's grid-area name (its SeatPosition). */
  area: string;
}

/** The board's grid under a reflow, as inline style values. */
export interface ReflowGrid {
  columns: string;
  rows: string;
  areas: string;
}

/**
 * reflowGrid is the board's grid under a reflow (ADR 0077, amended
 * 2026-10-17, #4102). Every avatar card goes in ONE column at the left
 * end of the opponents' row, stacked top to bottom in `avatarOrder`
 * (reflowAvatarOrder's order), each in a grid row of its own, so two
 * cards are each half the row's height. The full boards fill the rest
 * of the row in `columnCells` order (turn order): the focus with two
 * shares, a seat a prompt needs with one. The viewer's board spans the
 * bottom row. `split` is the layout's top and bottom row shares (1 and 1
 * in the row layout, 0.7 and 1.3 in the quadrant one), and the avatar
 * rows share the top one evenly. Every seat stays a cell of the board's
 * one grid, so CombatArrows and boardAnchor keep measuring against it.
 */
export function reflowGrid(
  columnCells: readonly ReflowCell[],
  avatarOrder: readonly string[],
  reflow: Reflow,
  split: readonly [number, number] = [1, 1],
): ReflowGrid {
  const areaOf = new Map(columnCells.map((c) => [c.id, c.area]));
  const avatars = avatarOrder.flatMap((id) => {
    const area = areaOf.get(id);
    return area && reflow.avatarSeatIDs.has(id) ? [area] : [];
  });
  const boards = columnCells.filter((c) => !reflow.avatarSeatIDs.has(c.id));
  const tracks: string[] = boards.map((c) =>
    c.id === reflow.focusSeatID && !reflow.evenSplit ? "minmax(0, 2fr)" : "minmax(0, 1fr)",
  );
  if (avatars.length > 0) tracks.unshift(REFLOW_AVATAR_COLUMN);
  const boardAreas = boards.map((c) => c.area).join(" ");
  const topRows = Math.max(1, avatars.length);
  const share = Number((split[0] / topRows).toFixed(4));
  const rows = [
    ...Array.from({ length: topRows }, () => `minmax(0, ${share}fr)`),
    `minmax(0, ${split[1]}fr)`,
  ];
  const top =
    avatars.length > 0
      ? avatars.map((a) => `"${[a, boardAreas].filter(Boolean).join(" ")}"`)
      : [`"${boardAreas}"`];
  const self = `"${Array.from({ length: tracks.length }, () => "self").join(" ")}"`;
  return { columns: tracks.join(" "), rows: rows.join(" "), areas: [...top, self].join(" ") };
}

// ---- The avatar cards' turn order (ADR 0077, amended 2026-10-17) ------
//
// The column reads in the order the seats play: the top card is the
// seat that plays soonest after the active one, counting round the table
// in seat order. Each card says who plays right before it. It is the
// seating order, not a prediction: extra turns, skipped turns and
// turn-order changes are not foreseen (call 7), and the active board's
// gold ring stays the one signal of whose turn it is now.

/** One seat at the table, in seat (turn) order, the viewer included. */
export interface TurnOrderSeat {
  id: string;
  name: string;
  eliminated?: boolean;
}

/** One avatar card, in the column's order. */
export interface AvatarCardOrder {
  id: string;
  /**
   * Who plays right before this seat: "Next up", "After you" or "After
   * <name>" (labels.ts; drawn upper case). Null for a seat that is out.
   */
  chip: string | null;
  /** The seat is out of the game: it sorts last and shows "out". */
  out: boolean;
}

/**
 * reflowAvatarOrder sorts the avatar-card seats by how soon each plays
 * after `activeID`, counting round `seats` (every seat, in seat order)
 * from the active seat, and names who plays right before each. A seat
 * that is out plays no turn, so it is skipped when finding the seat
 * before, and an eliminated avatar seat sorts last with no chip. An ID
 * in `avatarIDs` that is not at the table, or is the active seat, is
 * dropped.
 */
export function reflowAvatarOrder(
  seats: readonly TurnOrderSeat[],
  activeID: string,
  viewerID: string | null,
  avatarIDs: Iterable<string>,
): AvatarCardOrder[] {
  const n = seats.length;
  const active = seats.findIndex((s) => s.id === activeID);
  if (active < 0) return [];
  const want = new Set(avatarIDs);
  const live: { order: AvatarCardOrder; d: number }[] = [];
  const out: { order: AvatarCardOrder; d: number }[] = [];
  seats.forEach((s, i) => {
    if (!want.has(s.id) || i === active) return;
    const d = (i - active + n) % n;
    if (s.eliminated) out.push({ order: { id: s.id, chip: null, out: true }, d });
    else
      live.push({ order: { id: s.id, chip: chipFor(seats, i, active, viewerID), out: false }, d });
  });
  const byDistance = (a: { d: number }, b: { d: number }) => a.d - b.d;
  return [...live.sort(byDistance), ...out.sort(byDistance)].map((x) => x.order);
}

// The chip for the seat at index `i`: the nearest seat before it that is
// still in the game, which is the active seat (Next up), the viewer
// (After you) or another opponent (After <name>).
function chipFor(
  seats: readonly TurnOrderSeat[],
  i: number,
  active: number,
  viewerID: string | null,
): string {
  const n = seats.length;
  for (let k = 1; k < n; k++) {
    const j = (i - k + n) % n;
    if (j === active) return L.reflowNextUp;
    const s = seats[j];
    if (s.eliminated) continue;
    if (s.id === viewerID) return L.reflowAfterYou;
    return L.reflowAfter(s.name);
  }
  return L.reflowNextUp;
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
