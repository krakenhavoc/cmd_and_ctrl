// bugReport.ts — pure helpers behind the in-app "report a bug"
// button (BugReportModal.svelte). Kept out of the component so
// vitest can exercise them without a Svelte renderer, same pattern
// as zoneBrowser.logic.ts (ADR 0015 §2).

import type { GameView } from "./protocol";

// Client-side mirrors of the server caps in
// server/internal/lobby/bugreport.go — enforced there
// authoritatively; duplicated here so the modal can show a live
// character count instead of a post-submit 400.
export const BUG_TITLE_MAX = 200;
export const BUG_DESC_MAX = 5000;

// BugReportContext mirrors lobby.bugReportContext. Every field is
// optional; the server clips and never trusts the values.
export interface BugReportContext {
  game_id?: string;
  turn?: number;
  phase?: string;
  step?: string;
  seq?: number;
  connection?: string;
}

// buildBugContext snapshots the game situation at click time. Null
// view (connection still opening, or the game evaporated) degrades
// to just the game ID + connection status — a report with thin
// context beats no report.
export function buildBugContext(
  gameID: string,
  view: GameView | null,
  seq: number,
  connection: string,
): BugReportContext {
  const ctx: BugReportContext = { game_id: gameID };
  if (connection) ctx.connection = connection;
  if (seq > 0) ctx.seq = seq;
  if (view?.turn) {
    ctx.turn = view.turn.number;
    ctx.phase = view.turn.phase;
    ctx.step = view.turn.step;
  }
  return ctx;
}

// validateBugReport returns a human-readable problem or null when
// the draft is submittable. Mirrors the server's 400 rules.
export function validateBugReport(title: string, description: string): string | null {
  const t = title.trim();
  if (t.length === 0) return "a short title is required";
  if (t.length > BUG_TITLE_MAX) return `title is too long (max ${BUG_TITLE_MAX} characters)`;
  if (description.length > BUG_DESC_MAX)
    return `description is too long (max ${BUG_DESC_MAX} characters)`;
  return null;
}
