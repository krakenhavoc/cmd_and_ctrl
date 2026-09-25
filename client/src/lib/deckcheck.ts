// deckcheck.ts — the public deck coverage checker: types, fetch, and
// the pure grouping/formatting the #/deck-check page renders.
//
// Mirrors server/internal/deckcoverage's wire shapes for POST
// /deck-coverage and POST /deck-requests (ADR 0095 §5, docs/lobby.md).
// Hand-maintained, like catalog.ts and roadmap.ts — update all three
// sides (server, this file, the page) in lockstep.
//
// The report carries names, oracle IDs, buckets and caveat sentences
// only, never art or oracle text — the same line ADR 0092 drew for
// GET /roadmap, which is why POST /deck-coverage needs no session.

import { currentSession, setSession } from "./session";

/** How much of a card's printed text the engine automates. */
export type CoverageBucket = "manual" | "unreviewed" | "caveats" | "automated" | "no_effect";

/** Sorted, most actionable first — the order the server sends `cards` in. */
export const BUCKET_ORDER: readonly CoverageBucket[] = [
  "manual",
  "unreviewed",
  "caveats",
  "automated",
  "no_effect",
];

export interface CoverageCard {
  name: string;
  oracle_id: string;
  /** Copies in the list. */
  count: number;
  bucket: CoverageBucket;
  /** Present only on a `caveats` card: its catalogue Caveats sentences. */
  caveats?: string[];
}

export interface CoverageCounts {
  manual: number;
  unreviewed: number;
  caveats: number;
  automated: number;
  no_effect: number;
}

/** deck.Violation — informational only; never blocks a report. */
export interface DeckCoverageViolation {
  code: string;
  message: string;
  card?: string;
}

export type DeckCoverageSource = "moxfield" | "archidekt" | "text";

export interface CoverageReport {
  deck_name: string;
  source: DeckCoverageSource;
  /** Absent for a pasted-text check. */
  source_url?: string;
  /** "moxfield:<id>" | "archidekt:<id>"; absent for a pasted-text check. */
  deck_key?: string;
  commanders: string[];
  counts: CoverageCounts;
  cards: CoverageCard[];
  /** Names the card index could not resolve. */
  unknown: string[];
  violations: DeckCoverageViolation[];
}

export type DeckCheckRequest = { url: string } | { text: string };

/**
 * DeckCheckError is thrown by both checkDeck and requestDeck for any
 * response the caller should show. `message` is always the sentence to
 * render as-is — including the friendly rewrite of a bare rate-limit
 * 429 (one with no `status` field, i.e. the IP limiter's shape rather
 * than /deck-requests' own `rate_limited` outcome, which is not an
 * error at all — see requestDeck).
 */
export class DeckCheckError extends Error {
  constructor(
    public status: number,
    message: string,
    public code?: string,
    public violations?: DeckCoverageViolation[],
  ) {
    super(message);
    this.name = "DeckCheckError";
  }
}

interface RawErrorBody {
  error?: string;
  code?: string;
  violations?: DeckCoverageViolation[];
  status?: string;
  retry_after?: number;
}

const TOO_MANY_CHECKS = "Too many checks — try again in a few seconds.";

async function readErrorBody(res: Response): Promise<RawErrorBody> {
  try {
    return (await res.json()) as RawErrorBody;
  } catch {
    return {};
  }
}

async function deckCheckErrorFrom(res: Response): Promise<DeckCheckError> {
  const body = await readErrorBody(res);
  // The IP limiter's 429 — {"error": "too many requests"} — carries no
  // `status` field, unlike /deck-requests' own rate_limited outcome
  // (handled separately, as a normal response, in requestDeck). Any
  // other bare 429 without a `status` gets the same friendly rewrite.
  if (res.status === 429 && !body.status) {
    return new DeckCheckError(429, TOO_MANY_CHECKS, body.code, body.violations);
  }
  const message = body.error ?? `${res.status} ${res.statusText}`;
  return new DeckCheckError(res.status, message, body.code, body.violations);
}

/**
 * checkDeck submits a link or a pasted list to POST /deck-coverage.
 *
 * Plain `fetch`, not `authFetch`: the route is public — see
 * roadmap.ts's fetchRoadmap and catalog.ts's fetchCatalog for the same
 * reasoning. The bot's larger rate bucket rides its own admin session
 * server-side; the site page never needs to send one.
 */
export async function checkDeck(
  req: DeckCheckRequest,
  signal?: AbortSignal,
): Promise<CoverageReport> {
  const res = await fetch("/deck-coverage", {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    credentials: "same-origin",
    body: JSON.stringify(req),
    signal,
  });
  if (!res.ok) {
    throw await deckCheckErrorFrom(res);
  }
  return (await res.json()) as CoverageReport;
}

export type DeckRequestStatus = "filed" | "joined" | "nothing_to_add" | "rate_limited";

/** POST /deck-requests' 200/201/429 body. */
export interface DeckRequestResponse {
  status: DeckRequestStatus;
  issue_url?: string;
  issue_number?: number;
  /** Present only when status is "rate_limited": whole seconds to wait. */
  retry_after?: number;
  /** Present only on "joined": true when this requester had already asked. */
  already_requested?: boolean;
  report?: CoverageReport;
}

/**
 * requestDeck files or joins a deck request for a link check
 * (POST /deck-requests). Needs a session with a Discord identity —
 * the caller should gate the button on canRequestCards + a signed-in
 * Discord session and let the server's 403 be the final word.
 *
 * A 429 answers one of two shapes, and only one of them is an error:
 * /deck-requests' own per-requester limit comes back 429 with a
 * `status: "rate_limited"` body, which is a normal outcome (the UI
 * shows the wait) — not a thrown error. A 429 with no `status` is the
 * IP limiter's bare shape and throws, same as checkDeck.
 */
export async function requestDeck(url: string, signal?: AbortSignal): Promise<DeckRequestResponse> {
  const s = currentSession();
  const headers = new Headers({ "Content-Type": "application/json", Accept: "application/json" });
  if (s?.token) headers.set("Authorization", `Bearer ${s.token}`);
  const res = await fetch("/deck-requests", {
    method: "POST",
    headers,
    credentials: "same-origin",
    body: JSON.stringify({ url }),
    signal,
  });
  if (res.status === 401) {
    // Mirror authFetch: a 401 here is the server's own word that the
    // credential is gone, so clear it and let the page re-render
    // signed-out rather than leaving a stale session in the store.
    setSession(null);
  }
  if (res.ok) {
    return (await res.json()) as DeckRequestResponse;
  }
  if (res.status === 429) {
    const body = await readErrorBody(res);
    if (body.status === "rate_limited") {
      return { status: "rate_limited", retry_after: body.retry_after ?? 0 };
    }
    throw new DeckCheckError(429, TOO_MANY_CHECKS, body.code, body.violations);
  }
  throw await deckCheckErrorFrom(res);
}

/**
 * groupCardsByBucket splits a report's cards into their five buckets,
 * every bucket present (possibly empty) in BUCKET_ORDER. The server
 * already sorts `cards` by bucket then name, so each bucket's list
 * stays in that order.
 */
export function groupCardsByBucket(cards: CoverageCard[]): Record<CoverageBucket, CoverageCard[]> {
  const out: Record<CoverageBucket, CoverageCard[]> = {
    manual: [],
    unreviewed: [],
    caveats: [],
    automated: [],
    no_effect: [],
  };
  for (const c of cards) {
    out[c.bucket]?.push(c);
  }
  return out;
}

/** Bucket names in player words, not engine vocabulary. */
export const BUCKET_LABELS: Record<CoverageBucket, string> = {
  manual: "Not automated",
  unreviewed: "Not yet reviewed",
  caveats: "Automated, with caveats",
  automated: "Fully automated",
  no_effect: "Nothing to automate",
};

/** One-line explanation for each bucket, for a legend beside the counts. */
export const BUCKET_BLURBS: Record<CoverageBucket, string> = {
  manual: "Plays as a sandbox card — you resolve its rules text yourself.",
  unreviewed: "In the engine, but nobody has checked it against the printed text yet.",
  caveats: "The engine plays it, with a declared simplification or two.",
  automated: "The engine plays the whole card on its own.",
  no_effect:
    "Nothing to automate — a vanilla creature, a keyword the engine already enforces, a basic land.",
};

/**
 * canRequestCards reports whether "Request these cards" should show at
 * all: a link check (a stable deck key — pasted text has none, ADR
 * 0095 §2) with at least one manual or unreviewed card. Whether the
 * VIEWER may press it is a separate question — see the Discord-session
 * check the page makes; the server's 403 is the final word either way.
 */
export function canRequestCards(report: CoverageReport | null | undefined): boolean {
  if (!report) return false;
  if (report.source === "text" || !report.deck_key) return false;
  return report.counts.manual > 0 || report.counts.unreviewed > 0;
}

/**
 * formatRetryAfter turns whole seconds into a short duration a player
 * reads naturally: "45 seconds", "3 minutes", "8 hours". Rounds to the
 * nearest unit rather than truncating, so "89 seconds" reads as "1
 * minute" instead of a confusing "0 minutes".
 */
export function formatRetryAfter(seconds: number): string {
  const s = Math.max(0, Math.round(seconds));
  if (s < 60) return s === 1 ? "1 second" : `${s} seconds`;
  const minutes = Math.round(s / 60);
  if (minutes < 60) return minutes === 1 ? "1 minute" : `${minutes} minutes`;
  const hours = Math.round(minutes / 60);
  return hours === 1 ? "1 hour" : `${hours} hours`;
}

/**
 * deckRequestOutcomeMessage is the one-line summary of a POST
 * /deck-requests result, for the button's own feedback area. The
 * issue link itself is rendered separately (issue_url/issue_number),
 * since the caller usually wants it as a real `<a>`.
 */
export function deckRequestOutcomeMessage(res: DeckRequestResponse): string {
  switch (res.status) {
    case "filed":
      return "Filed a new issue for the missing cards.";
    case "joined":
      return res.already_requested
        ? "You already asked for this deck — the issue is still open."
        : "Already requested — added your name to the open issue.";
    case "nothing_to_add":
      return "Everything in this deck is already automated or caveated — nothing to add.";
    case "rate_limited":
      return `You've asked a few times already — try again in ${formatRetryAfter(res.retry_after ?? 0)}.`;
  }
}
