// loopShortcut.ts — the client half of ADR 0193 (CR 732.2): demonstrate
// a loop once, propose N more iterations, and answer another player's
// proposal. The server records, checks and replays; this file only
// reads GameView.loop_recording and a loop_proposal prompt and builds
// the params the five loop actions take.

import type {
  GameView,
  LoopPointView,
  LoopProposalView,
  LoopRecordingView,
  PendingChoiceView,
} from "./protocol";

// What the dock's loop control shows for the viewer right now.
//   none      nothing stands
//   armed     the viewer asked to record where they could not begin
//             (partway through a resolution, or without priority); the
//             recording begins at their next priority with nothing open
//   recording the viewer is recording; Stop and Cancel
//   ready     the viewer's recording closed; Propose (the sheet)
//   proposed  the viewer's proposal is waiting on the other seats
//   watching  another seat is recording or proposing
// Starting a recording is the ⋯ menu's "Record loop" (recordLoopEntry).
export type LoopControlMode = "none" | "armed" | "recording" | "ready" | "proposed" | "watching";

export interface LoopControlState {
  mode: LoopControlMode;
  // Steps recorded so far (recording) or in the iteration.
  steps: number;
  // The "watching" line: "Ann is recording a loop".
  text: string;
  // recording: the viewer asked to stop without priority, so the server
  // checks the loop at their next priority with nothing open (CR 732.2a
  // ends a shortcut where a player has priority).
  stopping: boolean;
  // ready: the viewer may propose now, holding priority with no prompt
  // open (CR 732.2a: "the player with priority may suggest a shortcut").
  // Otherwise the proposal waits for their next priority.
  canPropose: boolean;
}

// seatName is a seat's display name, or "A player".
function seatName(view: GameView, id: string | undefined): string {
  const s = view.seats.find((p) => p.id === id);
  return s?.display_name || s?.name || "A player";
}

// quiet: no prompt open and no trigger waiting to go on the stack, the
// state a recording begins and ends in (with the recorder's priority).
function quiet(view: GameView): boolean {
  return (view.pending_choices?.length ?? 0) === 0 && (view.pending_triggers?.length ?? 0) === 0;
}

// What another seat's loop reads as, for every other seat (CR 732.1a).
function watchingText(view: GameView, lr: LoopRecordingView): string {
  const who = seatName(view, lr.recorder);
  switch (lr.status) {
    case "armed":
      return `${who} will record a loop from their next priority`;
    case "recording":
      return `${who} is recording a loop`;
    case "proposed":
      return `${who} proposes repeating ${lr.name ?? "a loop"} ${lr.count ?? 0} more times`;
    default:
      return `${who} recorded a loop`;
  }
}

export function loopControl(
  view: GameView,
  viewerID: string | null,
  viewerHasPriority: boolean,
): LoopControlState {
  const lr = view.loop_recording;
  const base = { steps: lr?.steps ?? 0, text: "", stopping: false, canPropose: false };
  if (!lr) return { ...base, mode: "none" };
  if (viewerID && lr.recorder === viewerID) {
    switch (lr.status) {
      case "armed":
        return { ...base, mode: "armed" };
      case "recording":
        return { ...base, mode: "recording", stopping: lr.stopping === true };
      case "ready":
        return {
          ...base,
          mode: "ready",
          canPropose: viewerHasPriority && quiet(view) && view.state === "active",
        };
      default:
        return { ...base, mode: "proposed" };
    }
  }
  return { ...base, mode: "watching", text: watchingText(view, lr) };
}

// RecordLoopEntry is the ⋯ menu's "Record loop" (ADR 0193's amendment
// of 2026-10-30, #4736). It is always drawn: enabled with an optional
// hint, or disabled with the one-line reason, never hidden.
export interface RecordLoopEntry {
  enabled: boolean;
  // The line under the name: why it is disabled, or, enabled, when the
  // recording will begin. "" for none.
  note: string;
}

export interface RecordLoopContext {
  // A seat of the viewer's own at this table (not a spectator).
  seated: boolean;
  eliminated: boolean;
  // ADR 0166: another player makes the viewer's decisions.
  controlled: boolean;
  // The viewer holds priority now (as themselves).
  viewerHasPriority: boolean;
}

// recordLoopEntry: any seat still in an active game may ask to record
// while no loop stands, a prompt of theirs partway through a resolution
// included. Recording only captures what happens next; the proposal is
// what needs priority (CR 732.2a). Where they cannot begin now, the
// server arms it for their next priority with nothing open.
export function recordLoopEntry(
  view: GameView | null | undefined,
  viewerID: string | null,
  ctx: RecordLoopContext,
): RecordLoopEntry {
  const no = (note: string): RecordLoopEntry => ({ enabled: false, note });
  if (!view || !viewerID || !ctx.seated) return no("Only a player at the table can record a loop");
  if (ctx.eliminated) return no("You have left the game");
  if (view.state === "ended") return no("The game is over");
  if (view.state !== "active") return no("The game has not started");
  if (view.mulligans_open || view.opening_roll) return no("Not until every opening hand is kept");
  if (ctx.controlled) return no("Not while another player makes your decisions");
  const lr = view.loop_recording;
  if (lr) {
    return no(
      lr.recorder === viewerID
        ? "You have a loop going: its controls are on the dock"
        : watchingText(view, lr),
    );
  }
  if (ctx.viewerHasPriority && quiet(view)) return { enabled: true, note: "" };
  return { enabled: true, note: "Starts when you next hold priority" };
}

// LoopDockProps is what the action dock's loop control takes
// (LoopControl.svelte, the status line); Game.svelte builds it. Starting
// a recording is not the dock's: "Record loop" is a ⋯ menu entry
// (GameMenu.svelte, recordLoopEntry).
export interface LoopDockProps {
  state: LoopControlState;
  onStop: () => void;
  onCancel: () => void;
  onPropose: () => void;
}

// clampLoopCount is a typed count held to 1..max (a non-number is 1).
export function clampLoopCount(n: number, max: number): number {
  if (!Number.isFinite(n)) return 1;
  return Math.min(Math.max(1, Math.trunc(n)), Math.max(1, max));
}

// The viewer's loop_proposal prompt, or null.
export function loopProposalFor(
  view: GameView,
  viewerID: string | null,
): (PendingChoiceView & { loop_proposal: LoopProposalView }) | null {
  if (!viewerID) return null;
  for (const c of view.pending_choices ?? []) {
    if (c.kind === "loop_proposal" && c.chooser === viewerID && c.loop_proposal) {
      return c as PendingChoiceView & { loop_proposal: LoopProposalView };
    }
  }
  return null;
}

// defaultShorten is ADR 0193 owner decision 6's default: the chooser's
// first choice point in iteration 1. Null when they have none.
export function defaultShorten(p: LoopProposalView): LoopPointView | null {
  const first = p.choice_points?.[0];
  return first === undefined ? null : { iteration: 1, step: first };
}

// shortenAllowed: the point is one of the chooser's own steps, in an
// iteration of the proposal, and no later than its current end
// (CR 732.2b; the server refuses anything else).
export function shortenAllowed(p: LoopProposalView, at: LoopPointView): boolean {
  if (!Number.isInteger(at.iteration) || at.iteration < 1 || at.iteration > p.count) return false;
  if (!(p.choice_points ?? []).includes(at.step)) return false;
  if (at.iteration !== p.end.iteration) return at.iteration < p.end.iteration;
  return at.step <= p.end.step;
}

export function acceptParams(choiceID: string): { choice_id: string; accept: true } {
  return { choice_id: choiceID, accept: true };
}

export function shortenParams(
  choiceID: string,
  at: LoopPointView,
): { choice_id: string; shorten: LoopPointView } {
  return { choice_id: choiceID, shorten: { iteration: at.iteration, step: at.step } };
}
