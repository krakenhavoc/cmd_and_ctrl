// moderation.ts — what a public table lets a player do about another
// (#2984, ADR 0075 §2.6): remove-and-block, report a player, and the two
// refusals the server answers with a stable `code`.
//
// The server decides everything; this file owns the sentences, so the
// four places that show them cannot word them four ways, and the bits of
// logic a test can reach without rendering Svelte.

import { LobbyApiError } from "./session";

/** The `code` of the join refusal for a person the table removed and blocked. */
export const BLOCKED_FROM_TABLE_CODE = "blocked_from_table";

/** The `code` of a refused display name or table name. */
export const NAME_NOT_ALLOWED_CODE = "name_not_allowed";

function errorCode(err: unknown): string | undefined {
  if (!(err instanceof LobbyApiError)) return undefined;
  return (err.body as { code?: unknown } | undefined)?.code as string | undefined;
}

/** isBlockedFromTable: the join was refused because the table removed and blocked you. */
export function isBlockedFromTable(err: unknown): boolean {
  return errorCode(err) === BLOCKED_FROM_TABLE_CODE;
}

/** isNameRefused: the name (a seat label or a table name) is on the filter's list. */
export function isNameRefused(err: unknown): boolean {
  return errorCode(err) === NAME_NOT_ALLOWED_CODE;
}

// --- remove and block ---------------------------------------------------

/** The checkbox in the Remove confirm. */
export const BLOCK_LABEL = "Also block them from this table";

/**
 * What the checkbox does, and what it cannot. Honest about the limit: a
 * guest has no account, so the block is their name, and a different name
 * gets past it. The invite link is the wall.
 */
export const BLOCK_HINT =
  "A signed-in person is blocked by their account. A guest is blocked by their name, " +
  "so a guest who types a different name can sit back down. Mint a new invite to shut the door for everyone.";

// --- report a player ----------------------------------------------------

/** The longest reason the server keeps, in characters. Keep in step with lobby.MaxReportReasonLen (Go). */
export const REPORT_REASON_MAX = 300;

export const REPORT_MENU_LABEL = "Report a player…";
export const REPORT_TITLE = "Report a player";
export const REPORT_INTRO =
  "Tell the people who run this server about someone's name, chat or behaviour. " +
  "They see your reason and the last few chat lines. The player is not told, and nothing happens to them automatically.";
export const REPORT_SENT = "Report sent. Thank you.";

/** A seat the viewer may report: a person at the table who is not them. */
export interface ReportTarget {
  id: string;
  name: string;
}

interface ReportableSeat {
  id: string;
  name: string;
  is_bot?: boolean;
}

/**
 * reportTargets lists who the viewer can report: every other human seat.
 * A bot is left out because nobody typed its name or its lines, and a
 * spectator has no seat to name.
 */
export function reportTargets(
  seats: readonly ReportableSeat[],
  viewerID: string | null,
): ReportTarget[] {
  if (!viewerID) return [];
  return seats
    .filter((s) => s.id !== viewerID && !s.is_bot)
    .map((s) => ({ id: s.id, name: s.name }));
}

/** canSendReport: a target is picked and the reason says something. */
export function canSendReport(targetID: string, reason: string): boolean {
  return targetID !== "" && reason.trim() !== "";
}

/** The sentence for a failed report: the server's own, or a plain one for the rate limit. */
export function reportErrorMessage(err: unknown): string {
  if (err instanceof LobbyApiError) {
    if (err.status === 429) {
      return "You've sent a few reports already. Try again in a few minutes.";
    }
    return err.message;
  }
  return "Couldn't send the report. Try again.";
}
