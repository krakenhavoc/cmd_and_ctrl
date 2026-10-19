// playingAs.ts — ADR 0167 §9: "Playing as Bob".
//
// While Alice plays Bob's whole turn (Emrakul, the Promised End, CR
// 723.1), her board can act for either of two seats: her own, and
// Bob's. Game.svelte keeps two ids from here on:
//
//   - `viewerID`, who the viewer IS: settings sync, undo, the game menu
//     and concede, chat, fanfare, table management and Discord;
//   - `actingID`, whom the board ACTS FOR: the bottom seat, the dock's
//     priority, pass and end turn, combat, the legal-action lookup and
//     the `player` of every rules action it sends.
//
// Everything here is a pure read of the view, so it is unit-tested
// without a board. Nothing here is a rule: the server routes and
// refuses (actions/controlled_player.go); this only decides what the
// board shows and whose name an action carries.

import type { GameView, PlayerControlView, ScheduledPlayerControlView } from "./protocol";
import { blockDeclarationPending } from "./priority";

/**
 * playsForSeat says whether a control of this length has the controller
 * play the controlled seat's game on its board: a whole turn (ADR 0167)
 * or one combat phase (#3999). The server's ControlUntil.PlaysForThem.
 */
export function playsForSeat(until: PlayerControlView["until"] | undefined): boolean {
  return until === "turn" || until === "combat";
}

/**
 * The turn-long (or combat-phase, #3999) control the viewer holds over
 * another seat right now, or null. A Word of Command control (`until:
 * "resolution"`) is not this: it is played from its own sheet (ADR 0166).
 */
export function controlledByViewer(
  view: GameView | null | undefined,
  viewerID: string | null,
): PlayerControlView | null {
  if (!view || !viewerID) return null;
  return (
    view.player_controls?.find(
      (pc) => playsForSeat(pc.until) && pc.controller === viewerID && pc.controlled !== viewerID,
    ) ?? null
  );
}

/** The turn-long (or combat-phase) control another player holds over the viewer, or null. */
export function controllerOfViewer(
  view: GameView | null | undefined,
  viewerID: string | null,
): PlayerControlView | null {
  if (!view || !viewerID) return null;
  return (
    view.player_controls?.find(
      (pc) => playsForSeat(pc.until) && pc.controlled === viewerID && pc.controller !== viewerID,
    ) ?? null
  );
}

// The rules verbs a turn's controller may send as the controlled seat
// (ADR 0167 §4, actions/controlled_player.go turnRulesVerbs). Every
// other verb stays the viewer's own: blocks, settings, concede, undo.
const RULES_VERBS: ReadonlySet<string> = new Set([
  "pass_priority",
  "end_turn",
  "cast_spell",
  "activate_ability",
  "activate_loyalty",
  "activate_mana_ability",
  "special_action",
  "declare_attacker",
  "declare_attackers",
  "resolve_choice",
  "discard_selection",
]);

/**
 * actsForSeat says whether an action of this type, sent from a board
 * acting for another seat, carries that seat as its `player`.
 */
export function actsForSeat(type: string): boolean {
  return RULES_VERBS.has(type);
}

/**
 * endTurnClosed says whether End turn is closed to a board acting for
 * another seat: during a combat phase the viewer controls (#3999) it
 * would walk that seat's turn on past the control, and the server
 * refuses it (actions/controlled_player.go combatRulesVerbs).
 */
export function endTurnClosed(
  view: GameView | null | undefined,
  viewerID: string | null,
  actingID: string | null,
): boolean {
  const c = controlledByViewer(view, viewerID);
  return !!c && c.until === "combat" && actingID === c.controlled;
}

/** Which of the viewer's two seats the board acts for. */
export type PlayingAsPref = "self" | "controlled";

/**
 * actingSeatID is the seat the board acts for: the controlled seat while
 * the viewer plays its turn and has not switched to their own board,
 * the viewer otherwise (null for a spectator).
 */
export function actingSeatID(
  view: GameView | null | undefined,
  viewerID: string | null,
  pref: PlayingAsPref,
): string | null {
  if (!viewerID) return viewerID;
  const c = controlledByViewer(view, viewerID);
  return c && pref === "controlled" ? c.controlled : viewerID;
}

/**
 * decisionOwer names which of the viewer's two seats owes the next
 * decision: a prompt the viewer answers (for themselves, or for the
 * controlled seat, `on_behalf_of`), a cleanup discard, a block
 * declaration, or priority. Null when neither does (a third player is
 * deciding).
 */
export function decisionOwer(
  view: GameView | null | undefined,
  viewerID: string | null,
  controlledID: string | null,
): string | null {
  if (!view || !viewerID) return null;
  const prompt = view.pending_choices?.find((c) => c.chooser === viewerID);
  if (prompt) {
    return controlledID && prompt.on_behalf_of === controlledID ? controlledID : viewerID;
  }
  if (controlledID && (view.discard_pending?.[controlledID] ?? 0) > 0) return controlledID;
  if ((view.discard_pending?.[viewerID] ?? 0) > 0) return viewerID;
  if (blockDeclarationPending(view, viewerID)) return viewerID;
  const holder = view.turn ? view.seats[view.turn.priority_holder]?.id : undefined;
  if (holder && (holder === viewerID || holder === controlledID)) return holder;
  return null;
}

/**
 * PlayingAsState is the toggle's memory, one per board. `control` is
 * the seat whose turn the viewer is playing (null when none), `pref`
 * the board the viewer is on, and `lastOwer` the seat that owed the
 * last decision, so the board switches by itself only at a handoff.
 */
export interface PlayingAsState {
  control: string | null;
  pref: PlayingAsPref;
  lastOwer: string | null;
}

export const NOT_PLAYING_AS: PlayingAsState = Object.freeze({
  control: null,
  pref: "controlled",
  lastOwer: null,
}) as PlayingAsState;

/**
 * nextPlayingAs folds a new frame into the toggle (ADR 0167 §9). The
 * swap is on as the controlled turn begins. After that the board
 * switches by itself only when the seat that owes the next decision
 * changes between the viewer and the controlled seat; it never switches
 * while the same seat is still deciding, so a manual toggle stands
 * until the next handoff. When the control ends, the board is the
 * viewer's again. Returns `prev` itself when nothing changed.
 */
export function nextPlayingAs(
  prev: PlayingAsState,
  view: GameView | null | undefined,
  viewerID: string | null,
): PlayingAsState {
  const control = controlledByViewer(view, viewerID)?.controlled ?? null;
  if (!control) return prev.control === null ? prev : NOT_PLAYING_AS;
  const base: PlayingAsState =
    prev.control === control ? prev : { control, pref: "controlled", lastOwer: null };
  const ower = decisionOwer(view, viewerID, control);
  if (ower === null || ower === base.lastOwer) return base;
  const pref: PlayingAsPref = ower === control ? "controlled" : "self";
  return { control, pref, lastOwer: ower };
}

/** togglePlayingAs is the banner's button: the other board. */
export function togglePlayingAs(s: PlayingAsState): PlayingAsState {
  if (!s.control) return s;
  return { ...s, pref: s.pref === "controlled" ? "self" : "controlled" };
}

/**
 * viewForSeat is the frame as the seat `seatID` acts in it: `legal_moves`
 * holds only that seat's moves and `legal_actions` is that seat's
 * digest (`controlled_legal_actions[seatID]` for a controlled seat,
 * ADR 0167 §7). Everything that reads the move list or the digest
 * (autopass, the response window, the hand, the shortcuts) then answers
 * for that seat. A prompt the controller answers on the seat's behalf
 * (`on_behalf_of`) is that seat's own on its board, so the board and
 * the prompt sheet treat it as they treat any prompt the bottom seat
 * owes; that half needs `viewerID`, and only the seat's controller gets
 * it. A frame with no turn-long control is returned as is.
 *
 * A controlled seat's digest is only ever its controller's to read, so
 * the controlled player's own frame (no `controlled_legal_actions`)
 * reads as "no information" for their seat, which is the truth: they
 * decide nothing this turn.
 */
export function viewForSeat(
  view: GameView,
  seatID: string | null,
  viewerID: string | null = null,
): GameView {
  if (!seatID) return view;
  const control = view.player_controls?.find(
    (pc) => playsForSeat(pc.until) && pc.controlled === seatID && pc.controller !== seatID,
  );
  if (!control && !view.controlled_legal_actions) return view;
  const moves = view.legal_moves?.filter((m) => !m.player || m.player === seatID);
  const out: GameView = {
    ...view,
    legal_moves: moves && moves.length > 0 ? moves : undefined,
    legal_actions: control ? view.controlled_legal_actions?.[seatID] : view.legal_actions,
    controlled_legal_actions: undefined,
  };
  if (
    control &&
    viewerID === control.controller &&
    view.pending_choices?.some((c) => c.on_behalf_of === seatID)
  ) {
    out.pending_choices = view.pending_choices.map((c) =>
      c.on_behalf_of === seatID && c.chooser === control.controller ? { ...c, chooser: seatID } : c,
    );
  }
  return out;
}

/** One line of the board's "Playing as" banner. */
export interface PlayingAsBannerModel {
  // "playing": the viewer is on the controlled seat's board.
  // "viewing": the viewer controls a turn and is on their own board.
  // "controlled": another player controls the viewer this turn.
  // "scheduled": a turn-long control is waiting for its turn.
  // "watching": a spectator's or a bystander's line.
  kind: "playing" | "viewing" | "controlled" | "scheduled" | "watching";
  text: string;
  // The effect's card, shown after the text.
  source?: string;
  // A second line: "Alice is choosing for you".
  note?: string;
  // The toggle, for the controller only.
  toggle?: { label: string };
  // The handoff offer: the other seat owes a decision.
  switchTo?: { label: string };
  // The seat the line is about, for its colour.
  seatID?: string;
}

function nameOf(view: GameView, id: string): string {
  return view.seats.find((s) => s.id === id)?.name || "another player";
}

/**
 * playingAsBanner is the line the board's attention strip shows, or
 * null. The controller gets the toggle; the controlled player gets the
 * notice and who is choosing for them; a spectator and a third player
 * get the text only.
 */
export function playingAsBanner(
  view: GameView | null | undefined,
  viewerID: string | null,
  actingID: string | null,
): PlayingAsBannerModel | null {
  if (!view) return null;
  const mine = controlledByViewer(view, viewerID);
  if (mine && viewerID) {
    const bob = nameOf(view, mine.controlled);
    const span = mine.until === "combat" ? "combat" : "turn";
    const ower = decisionOwer(view, viewerID, mine.controlled);
    const onBob = actingID === mine.controlled;
    const switchTo =
      ower && ower !== actingID
        ? { label: ower === mine.controlled ? `Switch to ${bob}` : "Switch to you" }
        : undefined;
    return onBob
      ? {
          kind: "playing",
          text: `Playing as ${bob}`,
          source: mine.source_name,
          toggle: { label: "View my board" },
          switchTo,
          seatID: mine.controlled,
        }
      : {
          kind: "viewing",
          text: `Viewing your board, ${bob}'s ${span} is yours`,
          source: mine.source_name,
          toggle: { label: `Play as ${bob}` },
          switchTo,
          seatID: mine.controlled,
        };
  }
  const over = controllerOfViewer(view, viewerID);
  if (over && viewerID) {
    const alice = nameOf(view, over.controller);
    const choosing = view.pending_choices?.some((c) => c.on_behalf_of === viewerID);
    return {
      kind: "controlled",
      text: `${alice} controls you this ${over.until === "combat" ? "combat" : "turn"}`,
      source: over.source_name,
      note: choosing ? `${alice} is choosing for you` : undefined,
      seatID: over.controller,
    };
  }
  const live = view.player_controls?.find((pc) => playsForSeat(pc.until));
  if (live) {
    return {
      kind: "watching",
      text: `${nameOf(view, live.controller)} controls ${nameOf(view, live.controlled)} this ${live.until === "combat" ? "combat" : "turn"}`,
      source: live.source_name,
      seatID: live.controller,
    };
  }
  return scheduledLine(view, viewerID);
}

// What a scheduled control waits for, in the banner's words.
function scheduledSpan(s: ScheduledPlayerControlView): string {
  return s.until === "combat" ? "combat phase" : "turn";
}

// The last scheduled record over a player is the one that will apply
// (CR 723.1a; the server lists them oldest first).
function scheduledLine(view: GameView, viewerID: string | null): PlayingAsBannerModel | null {
  const all = view.scheduled_player_controls ?? [];
  if (all.length === 0) return null;
  const pick = (f: (s: ScheduledPlayerControlView) => boolean) =>
    [...all].reverse().find(f) ?? null;
  const overMe = viewerID ? pick((s) => s.controlled === viewerID) : null;
  if (overMe) {
    return {
      kind: "scheduled",
      text: `${nameOf(view, overMe.controller)} will control your next ${scheduledSpan(overMe)}`,
      source: overMe.source_name,
      seatID: overMe.controller,
    };
  }
  const byMe = viewerID ? pick((s) => s.controller === viewerID) : null;
  if (byMe) {
    return {
      kind: "scheduled",
      text: `You will control ${nameOf(view, byMe.controlled)}'s next ${scheduledSpan(byMe)}`,
      source: byMe.source_name,
      seatID: byMe.controlled,
    };
  }
  const any = all[all.length - 1];
  return {
    kind: "scheduled",
    text: `${nameOf(view, any.controller)} will control ${nameOf(view, any.controlled)}'s next ${scheduledSpan(any)}`,
    source: any.source_name,
    seatID: any.controller,
  };
}
