// tablemates.ts — the people you have shared a table with (ADR 0051
// decision 8, S34 sub-PR 6): the wire type, who may be offered the
// invite picker, and the one line each row shows.
//
// Mirrors GET /me/tablemates (server/internal/lobby/tablemates.go's
// Tablemate / tablematesResponse). Hand-maintained, like myDecks.ts
// and myGames.ts — update both sides in lockstep.
//
// The logic lives here rather than in the component for the reason
// myDecks.ts gives: the decisions ("can this session invite", "what
// does this row say") are worth a test, and a `.svelte` file is the
// awkward place to write one.

import { isSignedIn } from "./myDecks";
import type { Session } from "./session";

/** One tablemate, as GET /me/tablemates reports them. */
export interface Tablemate {
  user_id: string;
  display_name: string;
  /** Same-origin avatar path, or absent for an account with no avatar. */
  avatar_url?: string;
  /** When the most recent shared table was created, Unix milliseconds. */
  last_played_at: number;
}

export interface TablematesResponse {
  tablemates: Tablemate[];
}

/** The 200 body of POST /games/{id}/invites/dm. */
export interface InviteDMResponse {
  sent: boolean;
  user_id?: string;
  display_name?: string;
}

/**
 * canInviteTablemates reports whether this session should be offered
 * the picker for a given table. It mirrors the server's rule
 * (lobby.canInviteDM) closely enough to keep the UI honest without
 * duplicating it: a signed-in person who is seated at this table, or
 * who created it.
 *
 * `isCreator` is the table's `is_creator`, which the server computes
 * per viewer (redactMetaFor, #1098) without ever sending the creator's
 * identity. A creator who has not sat down yet is exactly who the
 * create flow offers the picker to (ADR 0110 §5 item 3).
 */
export function canInviteTablemates(
  s: Session | null | undefined,
  seatedGameID: string | null | undefined,
  gameID: string,
  isCreator: boolean = false,
): boolean {
  if (!s) return false;
  // An admin session is a server credential, not a person: it has no
  // UserID, so GET /me/tablemates refuses it and the list could only
  // ever be empty. Mounting the picker anyway cost an admin their
  // session — the refusal used to be a 401, which authFetch reads as
  // "expired" and clears (#1154). The status is 403 now, but the
  // picker still has nothing to show an admin.
  if (s.principal.role === "admin") return false;
  if (!isSignedIn(s.principal.user_id)) return false;
  if (isCreator) return true;
  return !!seatedGameID && seatedGameID === gameID;
}

/**
 * tablemateSubtitle is the line under a tablemate's name: when you
 * last shared a table. Coarse on purpose — "today", "3 days ago" —
 * because the picker is a list to recognise people in, not a history.
 *
 * `now` is injected so the test does not depend on the clock.
 */
export function tablemateSubtitle(mate: Tablemate, now: number = Date.now()): string {
  if (!mate.last_played_at || mate.last_played_at <= 0) return "played together";
  const days = Math.floor((now - mate.last_played_at) / 86_400_000);
  if (days <= 0) return "played today";
  if (days === 1) return "played yesterday";
  if (days < 30) return `played ${days} days ago`;
  const months = Math.floor(days / 30);
  if (months === 1) return "played a month ago";
  if (months < 12) return `played ${months} months ago`;
  const years = Math.floor(days / 365);
  return years === 1 ? "played a year ago" : `played ${years} years ago`;
}

/**
 * inviteSentMessage is what the picker says after a DM lands. The
 * server echoes the display name when it knows one; fall back to the
 * name the picker already had rather than saying nothing.
 */
export function inviteSentMessage(res: InviteDMResponse, fallbackName: string): string {
  const who = res.display_name || fallbackName;
  return who ? `invite sent to ${who}` : "invite sent";
}
