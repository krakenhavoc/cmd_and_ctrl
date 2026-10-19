// deckUpdate.ts — updating a saved deck from its own link, a new link or a
// pasted list (#2036, ADR 0095 amendment 2026-10-19).
//
// Mirrors POST /me/decks/{id}/update (the preview) and POST
// /me/decks/{id}/update/confirm (the commit) in
// server/internal/lobby/mydecks_update.go. Hand-maintained, like the other
// wire mirrors: change both sides together.

import type { CoverageReport } from "./deckcheck";
import { sourceHost, type MyDeckCoverage, type MyDeckInfo } from "./myDecks";

/** The three ways to give the new list. */
export type DeckUpdateChoice = "relink" | "url" | "list";

/** What POST /me/decks/{id}/update takes: exactly one of the three. */
export type DeckUpdateInput = { source: "relink" } | { url: string } | { list: string };

export interface DeckUpdateCard {
  name: string;
  count: number;
  /** A row in the sideboard or companion slot. */
  sideboard?: boolean;
}

export interface DeckUpdateChange {
  name: string;
  from: number;
  to: number;
  sideboard?: boolean;
}

export interface DeckUpdateDiff {
  added: DeckUpdateCard[];
  removed: DeckUpdateCard[];
  changed: DeckUpdateChange[];
  /** Distinct cards that are the same in both lists. */
  unchanged: number;
  commanders_before: string[];
  commanders_after: string[];
  commander_changed: boolean;
}

export interface DeckUpdateRequestPreview {
  /** This server files requests and the caller can file one. */
  available: boolean;
  /** Why not, in a sentence. */
  reason?: string;
  /** Missing cards a request would add: new to this deck's request. */
  new_cards: string[];
  /** Missing cards left out because the deck's earlier request lists them. */
  already_requested: number;
}

export interface DeckUpdatePreview {
  preview_id: string;
  expires_at: string;
  source: DeckUpdateChoice;
  /** The link the deck carries after the update; absent for a pasted list. */
  source_url?: string;
  /** True when the update drops a link the deck had. */
  link_cleared?: boolean;
  diff: DeckUpdateDiff;
  no_change: boolean;
  /** Card rows that keep the printing they had. */
  printings_kept: number;
  coverage_before?: MyDeckCoverage;
  coverage_after?: MyDeckCoverage;
  report: CoverageReport;
  request: DeckUpdateRequestPreview;
  /** Tables that have not ended and are seated with this deck. */
  live_tables: number;
}

export type DeckUpdateRequestStatus =
  | "filed"
  | "joined"
  | "nothing_new"
  | "unavailable"
  | "rate_limited"
  | "failed";

export interface DeckUpdateRequestResult {
  status: DeckUpdateRequestStatus;
  message?: string;
  issue_url?: string;
  issue_number?: number;
  new_cards?: string[];
  already_requested?: number;
  retry_after?: number;
}

export interface DeckUpdateConfirm {
  deck: MyDeckInfo;
  diff: DeckUpdateDiff;
  /** Present only when the update was confirmed with request_missing. */
  request?: DeckUpdateRequestResult;
}

export interface UpdateChoiceOption {
  key: DeckUpdateChoice;
  label: string;
}

/**
 * updateChoices are the sheet's three ways in, "Re-pull from ‹site›" first
 * and only when the deck has a link to re-pull from.
 */
export function updateChoices(deck: Pick<MyDeckInfo, "source_url">): UpdateChoiceOption[] {
  const out: UpdateChoiceOption[] = [];
  if (deck.source_url) {
    out.push({ key: "relink", label: `Re-pull from ${sourceHost(deck.source_url)}` });
  }
  out.push({ key: "url", label: "New link" }, { key: "list", label: "Paste list" });
  return out;
}

/** The first choice the sheet opens on. */
export function defaultChoice(deck: Pick<MyDeckInfo, "source_url">): DeckUpdateChoice {
  return deck.source_url ? "relink" : "list";
}

/** updateInput builds the request body for a choice, or null while its field is empty. */
export function updateInput(
  choice: DeckUpdateChoice,
  url: string,
  list: string,
): DeckUpdateInput | null {
  switch (choice) {
    case "relink":
      return { source: "relink" };
    case "url":
      return url.trim() ? { url: url.trim() } : null;
    case "list":
      return list.trim() ? { list } : null;
  }
}

const rows = (n: number, word: string): string => `${n} ${word}`;

/** diffSummary is the one line over the diff: "3 added, 2 removed, 1 changed". */
export function diffSummary(d: DeckUpdateDiff, noChange = false): string {
  const parts: string[] = [];
  if (d.added.length) parts.push(rows(d.added.length, "added"));
  if (d.removed.length) parts.push(rows(d.removed.length, "removed"));
  if (d.changed.length) parts.push(rows(d.changed.length, "changed"));
  if (d.commander_changed) parts.push("commander changed");
  if (parts.length === 0) return noChange ? "No changes" : "The cards are the same";
  return parts.join(", ");
}

/** diffCardLabel is one added or removed row: "3× Forest", "Sol Ring (sideboard)". */
export function diffCardLabel(c: DeckUpdateCard): string {
  const n = c.count > 1 ? `${c.count}× ` : "";
  return `${n}${c.name}${c.sideboard ? " (sideboard)" : ""}`;
}

/** diffChangeLabel is one changed row: "Forest 10 → 12". */
export function diffChangeLabel(c: DeckUpdateChange): string {
  return `${c.name} ${c.from} → ${c.to}${c.sideboard ? " (sideboard)" : ""}`;
}

/** commanderLine reads the commander change: "Atraxa → Tymna / Thrasios". */
export function commanderLine(d: DeckUpdateDiff): string {
  const names = (l: string[]): string => (l.length ? l.join(" / ") : "none");
  return `${names(d.commanders_before)} → ${names(d.commanders_after)}`;
}

/** coverageLine is "95 of 100 cards play as printed", or "" when there is no coverage. */
export function coverageLine(c: MyDeckCoverage | undefined): string {
  if (!c || c.deck_size <= 0) return "";
  return `${c.as_printed_copies} of ${c.deck_size} cards play as printed`;
}

/** missingCount is the distinct cards the engine does not run (manual + unreviewed). */
export function missingCount(c: MyDeckCoverage | undefined): number {
  return c ? c.counts.manual + c.counts.unreviewed : 0;
}

/**
 * requestOptionText is the checkbox's label: how many cards a request would
 * add, and how many are left out because the deck already asked for them.
 */
export function requestOptionText(r: DeckUpdateRequestPreview): string {
  const n = r.new_cards.length;
  const base = `Request the ${n} new missing card${n === 1 ? "" : "s"}`;
  return r.already_requested > 0 ? `${base} (${r.already_requested} already requested)` : base;
}

/** wait reads a number of seconds as "45 seconds", "12 minutes" or "3 hours". */
function wait(seconds: number): string {
  if (seconds < 90) return `${Math.max(1, Math.round(seconds))} seconds`;
  if (seconds < 5400) return `${Math.round(seconds / 60)} minutes`;
  return `${Math.round(seconds / 3600)} hours`;
}

/** requestResultMessage is the sentence the page shows for the request an update made. */
export function requestResultMessage(r: DeckUpdateRequestResult): string {
  const n = r.new_cards?.length ?? 0;
  const cards = `${n} new card${n === 1 ? "" : "s"}`;
  switch (r.status) {
    case "filed":
      return `Filed a new issue for ${cards}.`;
    case "joined":
      return `Added ${cards} to the deck's open issue.`;
    case "nothing_new":
      return "Nothing new to request: the deck's earlier request already lists every missing card.";
    case "rate_limited":
      return `You've asked a few times already — the request can go in ${wait(r.retry_after ?? 0)}.`;
    case "unavailable":
    case "failed":
      return r.message || "The request could not be made.";
  }
}

/** updateDoneMessage is the page's line after an update: the diff, then the request. */
export function updateDoneMessage(res: DeckUpdateConfirm): string {
  const head = `Updated ${res.deck.name}: ${diffSummary(res.diff)}.`;
  return res.request ? `${head} ${requestResultMessage(res.request)}` : head;
}
