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

// --- report kinds ---
//
// What the player is telling us: something is broken, something
// should exist, or they don't understand something. The kind picks
// the GitHub label and — just as importantly — what the report
// attaches (ADR 0017 §8).
//
// This table MIRRORS `bugKinds` in server/internal/lobby/bugreport.go,
// which is authoritative: the client sends a kind name, never a
// label, and the server re-derives everything from its own copy. What
// lives here is only what the modal needs in order to tell the
// reporter the truth about what they are about to send.

export type BugReportKind = "bug" | "idea" | "question";

export interface BugKindSpec {
  kind: BugReportKind;
  /** the picker's chip */
  chip: string;
  /** modal heading and submit button once picked */
  heading: string;
  cta: string;
  hint: string;
  titlePlaceholder: string;
  detailsPlaceholder: string;
  /** the existing repo label the server applies */
  label: string;
  /** the client activity log rides along */
  log: boolean;
  /** the server pins the game's replay (admin-only, and the expensive one) */
  replay: boolean;
  /** the server pins the public game log (admin-only, a few KiB) */
  gameLog: boolean;
}

export const BUG_KINDS: readonly BugKindSpec[] = [
  {
    kind: "bug",
    chip: "Something's broken",
    heading: "Report a bug",
    cta: "File bug",
    hint: "What happened, and what did you expect instead? This files an issue in the project's GitHub repo — no account needed on your side.",
    titlePlaceholder: "e.g. cast dialog ignored my X value",
    detailsPlaceholder: "Steps to reproduce, what you saw, what you expected…",
    label: "bug",
    log: true,
    replay: true,
    gameLog: true,
  },
  {
    kind: "idea",
    chip: "Something's missing",
    heading: "Suggest an improvement",
    cta: "Send idea",
    hint: "What would you like to be able to do, and what would it let you do better? This files an issue in the project's GitHub repo.",
    titlePlaceholder: "e.g. let me widen the stack panel",
    detailsPlaceholder: "What you're trying to do, and what would make it easier…",
    label: "enhancement",
    log: false,
    replay: false,
    gameLog: false,
  },
  {
    kind: "question",
    chip: "I don't understand something",
    heading: "Ask a question",
    cta: "Ask question",
    hint: "What are you trying to work out? This files an issue in the project's GitHub repo and you'll get an answer there.",
    titlePlaceholder: "e.g. why did my commander go to the graveyard?",
    detailsPlaceholder: "What you did, and what you expected to understand…",
    label: "question",
    log: true,
    replay: false,
    gameLog: true,
  },
];

// DEFAULT_BUG_KIND is what the modal opens on and what the server
// assumes when the field is absent. It is "bug" on both sides
// because that is what every client built before the picker existed
// is filing.
export const DEFAULT_BUG_KIND: BugReportKind = "bug";

// bugKindSpec looks up a kind, falling back to the default rather
// than throwing — a modal that can't render is worse than one that
// opens on the wrong tab.
export function bugKindSpec(kind: BugReportKind | string): BugKindSpec {
  return BUG_KINDS.find((k) => k.kind === kind) ?? BUG_KINDS[0];
}

// describeBugAttachments lists, in the reporter's words, what this
// kind sends beyond their prose and screenshots. Empty when the kind
// attaches nothing automatic.
//
// The modal shows this next to the picker because the kind is the
// thing that changes it: switching from "broken" to "missing" quietly
// stops sending a replay, and the reporter should see that happen
// rather than discover it in an issue.
export function describeBugAttachments(spec: BugKindSpec, inGame: boolean): string[] {
  const bits: string[] = [];
  if (spec.log) bits.push("your recent activity log");
  const replay = inGame && spec.replay;
  const gameLog = inGame && spec.gameLog;
  // The two pins are described as one phrase when both ride along:
  // they are the same artifact to a reporter ("what the server saw"),
  // and a three-item list in a 12px line is noise.
  if (replay && gameLog) {
    bits.push("an admin-only snapshot of this game's replay and log");
  } else if (replay) {
    bits.push("an admin-only snapshot of this game's replay");
  } else if (gameLog) {
    bits.push("an admin-only snapshot of this game's log");
  }
  return bits;
}

// joinPhrases renders a short list as English. Kept here with the
// phrases it joins so the modal has no string assembly of its own.
export function joinPhrases(parts: readonly string[]): string {
  if (parts.length <= 1) return parts[0] ?? "";
  return `${parts.slice(0, -1).join(", ")} and ${parts[parts.length - 1]}`;
}

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

// --- attachments ---
//
// These mirror the server caps in server/internal/bugstore/store.go.
// Duplicated so the modal can refuse a 12 MiB screenshot before
// spending thirty seconds uploading it; the server is still the
// authority (it re-checks, and sniffs the bytes rather than trusting
// the type the browser reports).
export const BUG_MAX_IMAGES = 4;
export const BUG_MAX_IMAGE_BYTES = 4 * 1024 * 1024;
export const BUG_MAX_TOTAL_IMAGE_BYTES = 10 * 1024 * 1024;

// BUG_IMAGE_TYPES is the set GitHub renders inline. SVG is absent on
// purpose: the server refuses it because an unauthenticated route
// serving attacker-supplied SVG is a stored-XSS hole, so offering it
// here would only produce a confusing rejection.
export const BUG_IMAGE_TYPES = ["image/png", "image/jpeg", "image/gif", "image/webp"];

// --- client log ---

export const BUG_LOG_MAX_ENTRIES = 200;

export type BugLogKind = "sent" | "received" | "error" | "info" | "console";

// BugLogEntry mirrors lobby.bugLogEntry. `at` is epoch milliseconds
// from this browser's clock; the server formats it and labels it as
// the reporter's clock rather than pretending it is server time.
export interface BugLogEntry {
  at: number;
  kind: BugLogKind;
  text: string;
}

// WsLogLike is the shape of ws.LogEntry that this module needs.
// Structural rather than an import so these helpers stay free of the
// WebSocket client — the same reason they live outside the component.
export interface WsLogLike {
  at: Date;
  direction: string;
  text: string;
}

// ConsoleLogLike is the shape of clientErrors.ClientErrorEntry.
export interface ConsoleLogLike {
  at: number;
  text: string;
}

const LOG_KINDS: BugLogKind[] = ["sent", "received", "error", "info", "console"];

// collectBugLog merges the protocol log with the captured JS errors
// into one timeline and keeps the newest `limit` entries.
//
// Merged rather than sent as two lists because the useful signal is
// the interleaving: "action sent, then a TypeError, then no snapshot"
// is a diagnosis, whereas the same three facts in two separate blocks
// are three facts.
//
// The tail is what's kept — a bug report is filed just after the thing
// went wrong, so the end of the log is the relevant part.
export function collectBugLog(
  wsLog: readonly WsLogLike[],
  consoleLog: readonly ConsoleLogLike[],
  limit: number = BUG_LOG_MAX_ENTRIES,
): BugLogEntry[] {
  const merged: BugLogEntry[] = [
    ...wsLog.map((e) => ({
      at: e.at instanceof Date ? e.at.getTime() : 0,
      kind: (LOG_KINDS as string[]).includes(e.direction) ? (e.direction as BugLogKind) : "info",
      text: e.text,
    })),
    ...consoleLog.map((e) => ({ at: e.at, kind: "console" as BugLogKind, text: e.text })),
  ];
  // Stable sort by timestamp: entries stamped in the same millisecond
  // keep the order they were recorded in, which for the protocol log is
  // causal order.
  merged.sort((a, b) => a.at - b.at);
  return limit > 0 && merged.length > limit ? merged.slice(-limit) : merged;
}

// formatBytes renders a size for the modal's counter. Deliberately
// coarse — the reporter needs "3.2 MB of 10 MB", not exact bytes.
export function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${Math.round(n / 1024)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

// AttachmentLike is the subset of File the validator reads, so tests
// don't need a real File.
export interface AttachmentLike {
  name: string;
  type: string;
  size: number;
}

// validateAttachments returns a human-readable problem with the
// proposed set, or null when it is submittable. Checks the whole set at
// once because two of the three limits (count, aggregate size) are
// properties of the set rather than of any one file.
export function validateAttachments(files: readonly AttachmentLike[]): string | null {
  if (files.length > BUG_MAX_IMAGES) {
    return `at most ${BUG_MAX_IMAGES} images per report`;
  }
  let total = 0;
  for (const f of files) {
    if (!BUG_IMAGE_TYPES.includes(f.type)) {
      return `${f.name || "that file"} isn't a PNG, JPEG, GIF, or WebP image`;
    }
    if (f.size > BUG_MAX_IMAGE_BYTES) {
      return `${f.name || "that file"} is ${formatBytes(f.size)} — the limit is ${formatBytes(BUG_MAX_IMAGE_BYTES)} per image`;
    }
    total += f.size;
  }
  if (total > BUG_MAX_TOTAL_IMAGE_BYTES) {
    return `attachments total ${formatBytes(total)} — the limit is ${formatBytes(BUG_MAX_TOTAL_IMAGE_BYTES)}`;
  }
  return null;
}

// acceptableImages filters a dropped/pasted/picked FileList down to
// the image types the server will take, so a paste that also carried
// text/html parts doesn't surface as a validation error the reporter
// can't act on.
export function acceptableImages(files: readonly AttachmentLike[]): AttachmentLike[] {
  return files.filter((f) => BUG_IMAGE_TYPES.includes(f.type));
}
