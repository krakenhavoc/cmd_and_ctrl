// loopShortcut.ts — the client half of ADR 0193 (CR 732.2): demonstrate
// a loop once, propose N more iterations, and answer another player's
// proposal. The server records, checks and replays; this file only
// reads GameView.loop_recording and a loop_proposal prompt and builds
// the params the five loop actions take.

import type { GameView, LoopPointView, LoopProposalView, PendingChoiceView } from "./protocol";

// What the dock's loop control shows for the viewer right now.
//   none      nothing (no priority, or a prompt is open)
//   record    the viewer may start recording
//   recording the viewer is recording; Stop and Cancel
//   ready     the viewer's recording closed; Propose (the sheet)
//   proposed  the viewer's proposal is waiting on the other seats
//   watching  another seat is recording or proposing
export type LoopControlMode = "none" | "record" | "recording" | "ready" | "proposed" | "watching";

export interface LoopControlState {
  mode: LoopControlMode;
  // Steps recorded so far (recording) or in the iteration.
  steps: number;
  // The "watching" line: "Ann is recording a loop".
  text: string;
}

// seatName is a seat's display name, or "A player".
function seatName(view: GameView, id: string | undefined): string {
  const s = view.seats.find((p) => p.id === id);
  return s?.display_name || s?.name || "A player";
}

export function loopControl(
  view: GameView,
  viewerID: string | null,
  viewerHasPriority: boolean,
): LoopControlState {
  const lr = view.loop_recording;
  if (!lr) {
    const quiet = (view.pending_choices?.length ?? 0) === 0 && view.state === "active";
    return { mode: viewerHasPriority && quiet && viewerID ? "record" : "none", steps: 0, text: "" };
  }
  if (viewerID && lr.recorder === viewerID) {
    switch (lr.status) {
      case "recording":
        return { mode: "recording", steps: lr.steps, text: "" };
      case "ready":
        return { mode: "ready", steps: lr.steps, text: "" };
      default:
        return { mode: "proposed", steps: lr.steps, text: "" };
    }
  }
  const who = seatName(view, lr.recorder);
  const text =
    lr.status === "recording"
      ? `${who} is recording a loop`
      : lr.status === "proposed"
        ? `${who} proposes repeating ${lr.name ?? "a loop"} ${lr.count ?? 0} more times`
        : `${who} recorded a loop`;
  return { mode: "watching", steps: lr.steps, text };
}

// LoopDockProps is what the action dock's loop control takes
// (LoopControl.svelte); Game.svelte builds it.
export interface LoopDockProps {
  state: LoopControlState;
  onRecord: () => void;
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
