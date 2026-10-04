// myDecks.ts — the signed-in player's deck library (ADR 0051 decision
// 7, S34 sub-PR 5): types, the "signed in" check, and the one line the
// picker puts under each saved deck's name.
//
// Mirrors GET /me/decks (server/internal/lobby/http.go's
// myDeckInfo/myDecksResponse). Hand-maintained, like prebuiltDecks.ts
// and protocol.ts — update both sides in lockstep.
//
import type { CoverageBucket } from "./deckcheck";

// The logic lives here rather than in the component for the same
// reason prebuiltDecks.ts's summariser does: this project has no
// jsdom, so a `.svelte` file cannot be unit-tested, and both "is this
// caller signed in" and "what does this row say" are worth pinning
// down with a test.

/**
 * A saved deck's coverage on read (ADR 0110 section 6), mirroring
 * lobby.myDeckCoverageInfo. Never stored: the server computes it per
 * request, so it is right after every deploy.
 */
export interface MyDeckCoverage {
  counts: Record<CoverageBucket, number>;
  unknown: number;
  /** Distinct cards that play exactly as printed (automated + no_effect). */
  as_printed: number;
  /** Every distinct card the report bucketed. */
  resolved: number;
  /** `counts` weighted by copies (#2220). */
  copies: Record<CoverageBucket, number>;
  /** Copies the card index could not resolve. */
  unknown_copies: number;
  /** Copies that play exactly as printed (automated + no_effect). */
  as_printed_copies: number;
  /** The deck's size: every copy, unresolved ones included. */
  deck_size: number;
}

/** One saved deck, as GET /me/decks reports it. */
export interface MyDeckInfo {
  id: string;
  name: string;
  commanders: string[];
  card_count: number;
  updated_at: string;
  /** The link a deck was imported from; absent for a pasted one. */
  source_url?: string;
  /** Absent when the server has no card index or the list no longer parses. */
  coverage?: MyDeckCoverage;
}

export interface MyDecksResponse {
  decks: MyDeckInfo[];
}

// ZERO_USER_ID is the all-zero uuid a Go uuid.UUID marshals to when
// it's the zero value. Principal.user_id (server:
// server/internal/auth/auth.go) carries `json:"...,omitempty"`, but
// `omitempty` never actually omits it: encoding/json only calls an
// array "empty" at length zero, and uuid.UUID is a fixed [16]byte, so
// the field always renders as a string rather than being left out.
// This is true of every uuid.UUID field on Principal (admin_id,
// game_id, player_id too), not something specific to user_id — see
// ADR 0051's sub-PR 5 implementation notes.
export const ZERO_USER_ID = "00000000-0000-0000-0000-000000000000";

/**
 * isSignedIn reports whether a principal's user_id names a real
 * users(id) row — i.e. whether "your decks" has anywhere to read
 * from. Deliberately NOT a truthiness check on user_id: the field is
 * always present (see ZERO_USER_ID above), so `!!user_id` is always
 * true and would show the picker to every guest.
 */
export function isSignedIn(userID: string | undefined): boolean {
  return !!userID && userID !== ZERO_USER_ID;
}

/**
 * deckSubtitle is the "commander(s) · N cards" line under a saved
 * deck's name. Partners/backgrounds join with " / ", matching how a
 * player would read two commanders named together; an empty
 * commanders list (parsing hadn't resolved one, or the format never
 * carries one) just leaves the count.
 */
export function deckSubtitle(deck: MyDeckInfo): string {
  const commanders = deck.commanders.filter((c) => c.trim() !== "").join(" / ");
  const count = `${deck.card_count} card${deck.card_count === 1 ? "" : "s"}`;
  return [commanders, count].filter((s) => s !== "").join(" · ");
}

/**
 * coverageLine is the headline under a saved deck: "N of M cards play
 * as printed" (ADR 0110 section 6, the strict reading: automated plus
 * nothing-to-automate), counted by copies so M is the deck's size
 * (#2220). Empty when the server sent no coverage, so the
 * row simply has no line rather than a wrong one.
 */
export function coverageLine(deck: MyDeckInfo): string {
  const c = deck.coverage;
  if (!c || c.deck_size === 0) return "";
  return `${c.as_printed_copies} of ${c.deck_size} cards play as printed`;
}

/**
 * coverageDetail is the second line: what the rest of the deck needs,
 * in the report's words for the buckets. Zero buckets are left out.
 */
export function coverageDetail(deck: MyDeckInfo): string {
  const c = deck.coverage;
  if (!c) return "";
  const parts: string[] = [];
  if (c.copies.caveats > 0) parts.push(`${c.copies.caveats} simplified`);
  if (c.copies.unreviewed > 0) parts.push(`${c.copies.unreviewed} not checked yet`);
  if (c.copies.manual > 0) parts.push(`${c.copies.manual} you resolve by hand`);
  if (c.unknown_copies > 0) parts.push(`${c.unknown_copies} not found`);
  return parts.join(" · ");
}

/** sourceHost is the short label for a deck's source link ("moxfield.com"). */
export function sourceHost(url: string | undefined): string {
  if (!url) return "";
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return url;
  }
}
