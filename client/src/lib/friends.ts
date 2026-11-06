// friends.ts — friends, requests, blocks and the two social settings
// (ADR 0213 §2, §3 and §9, slice 3 of #3313): the wire types, the calls,
// and the words each row and button says.
//
// Mirrors server/internal/lobby/friends_http.go. Hand-maintained, like
// tablemates.ts; update both sides in lockstep.
//
// Pure where it can be, like playedWith.ts: the decisions ("which button
// does this row get", "what does this answer say") are worth a test, and a
// `.svelte` file is the awkward place to write one.

import { L } from "./labels";
import { authFetch, LobbyApiError } from "./session";
import type { Tablemate } from "./tablemates";

/** The hash route of the page that manages requests sent, blocks and the settings. */
export const FRIENDS_HASH = "#/friends";

/** One person in GET /me/friends or GET /me/blocks. `since` is Unix milliseconds. */
export interface FriendEntry {
  user_id: string;
  display_name: string;
  avatar_url?: string;
  since: number;
}

/** GET /me/friends. Friends in name order, requests newest first. */
export interface FriendsResponse {
  friends: FriendEntry[];
  incoming: FriendEntry[];
  outgoing: FriendEntry[];
  max_friends?: number;
  max_outgoing?: number;
}

/** What POST /me/friends/requests did. */
export type FriendOutcome = "sent" | "pending" | "friends" | "already_friends";

export interface FriendRequestResponse {
  outcome: FriendOutcome;
  user_id?: string;
  display_name?: string;
}

/** "Who can see what I'm doing on CMD & CTRL" (§5). Read by slice 4. */
export type PresenceSetting = "friends" | "played_with" | "nobody";
/** "Who can send me friend requests" (§2). */
export type RequestsSetting = "played_with" | "nobody";

export interface SocialSettings {
  presence: PresenceSetting;
  requests: RequestsSetting;
}

/** The two settings' choices, in the order the page shows them, default first. */
export const PRESENCE_CHOICES: ReadonlyArray<{ id: PresenceSetting; label: string }> = [
  { id: "friends", label: "My friends" },
  { id: "played_with", label: "My friends and people I’ve played with" },
  { id: "nobody", label: "Nobody" },
];
export const REQUESTS_CHOICES: ReadonlyArray<{ id: RequestsSetting; label: string }> = [
  { id: "played_with", label: "People I’ve played with" },
  { id: "nobody", label: "Nobody" },
];

// --- calls -------------------------------------------------------------

const enc = encodeURIComponent;

export async function fetchFriends(): Promise<FriendsResponse> {
  const res = await authFetch("/me/friends");
  const body = (await res.json()) as Partial<FriendsResponse>;
  return {
    friends: body.friends ?? [],
    incoming: body.incoming ?? [],
    outgoing: body.outgoing ?? [],
    max_friends: body.max_friends,
    max_outgoing: body.max_outgoing,
  };
}

/** Ask by account id: a Played with row, or a request to answer by asking back. */
export async function askFriend(userID: string): Promise<FriendRequestResponse> {
  const res = await authFetch("/me/friends/requests", {
    method: "POST",
    body: JSON.stringify({ user_id: userID }),
  });
  return (await res.json()) as FriendRequestResponse;
}

/** Ask from a seat at a table you sit at: the server resolves the account. */
export async function askFriendAtSeat(
  gameID: string,
  playerID: string,
): Promise<FriendRequestResponse> {
  const res = await authFetch("/me/friends/requests", {
    method: "POST",
    body: JSON.stringify({ game_id: gameID, player_id: playerID }),
  });
  return (await res.json()) as FriendRequestResponse;
}

export async function acceptFriend(userID: string): Promise<FriendRequestResponse> {
  const res = await authFetch(`/me/friends/requests/${enc(userID)}/accept`, { method: "POST" });
  return (await res.json()) as FriendRequestResponse;
}

export async function ignoreFriend(userID: string): Promise<void> {
  await authFetch(`/me/friends/requests/${enc(userID)}/ignore`, { method: "POST" });
}

export async function cancelFriendRequest(userID: string): Promise<void> {
  await authFetch(`/me/friends/requests/${enc(userID)}`, { method: "DELETE" });
}

export async function removeFriend(userID: string): Promise<void> {
  await authFetch(`/me/friends/${enc(userID)}`, { method: "DELETE" });
}

export async function fetchBlocks(): Promise<FriendEntry[]> {
  const res = await authFetch("/me/blocks");
  const body = (await res.json()) as { blocks?: FriendEntry[] };
  return body.blocks ?? [];
}

export async function blockPerson(userID: string): Promise<void> {
  await authFetch(`/me/blocks/${enc(userID)}`, { method: "PUT" });
}

export async function unblockPerson(userID: string): Promise<void> {
  await authFetch(`/me/blocks/${enc(userID)}`, { method: "DELETE" });
}

export async function fetchSocialSettings(): Promise<SocialSettings> {
  const res = await authFetch("/me/social");
  return (await res.json()) as SocialSettings;
}

export async function putSocialSettings(change: Partial<SocialSettings>): Promise<SocialSettings> {
  const res = await authFetch("/me/social", { method: "PUT", body: JSON.stringify(change) });
  return (await res.json()) as SocialSettings;
}

// --- words -------------------------------------------------------------

/**
 * FriendState is where a Played with row (or a seat) stands with you,
 * which picks its friend control: Add friend, Accept (they asked you),
 * Request sent, or Friends.
 */
export type FriendState = "none" | "incoming" | "outgoing" | "friend";

export function friendStateOf(m: Pick<Tablemate, "friend" | "request">): FriendState {
  if (m.friend) return "friend";
  if (m.request === "incoming") return "incoming";
  if (m.request === "outgoing") return "outgoing";
  return "none";
}

/** The state after an answer from POST /me/friends/requests. */
export function stateAfter(outcome: FriendOutcome): FriendState {
  return outcome === "friends" || outcome === "already_friends" ? "friend" : "outgoing";
}

/** outcomeLine is what a row says once asking is done. */
export function outcomeLine(outcome: FriendOutcome, name: string): string {
  switch (outcome) {
    case "sent":
      return `Friend request sent to ${name}`;
    case "pending":
      return `Your request to ${name} is still waiting`;
    case "friends":
      return `You and ${name} are friends now`;
    case "already_friends":
      return `You and ${name} are already friends`;
  }
}

/** friendErrorLine is what a row says when a call failed. */
export function friendErrorLine(err: unknown): string {
  if (err instanceof LobbyApiError) {
    if (err.status === 503) return "Friends aren’t available on this server.";
    const m = err.message.trim();
    if (m) return m.charAt(0).toUpperCase() + m.slice(1) + (/[.!?]$/.test(m) ? "" : ".");
  }
  return "That didn’t work. Try again.";
}

/**
 * friendControlText is the visible word on a row's friend control, by
 * state. Add friend is the contract label (L.addFriend); the others are
 * a status, not an action, except Accept.
 */
export function friendControlText(state: FriendState): string {
  switch (state) {
    case "none":
      return L.addFriend;
    case "incoming":
      return L.acceptFriend;
    case "outgoing":
      return "Request sent";
    case "friend":
      return "Friends";
  }
}

/** sinceLine is the line under a name in the drawer and on the page. */
export function sinceLine(
  kind: "friend" | "incoming" | "outgoing" | "blocked",
  at: number,
  now: number = Date.now(),
): string {
  const ago = agoText(at, now);
  switch (kind) {
    case "friend":
      return ago ? `friends since ${ago}` : "friends";
    case "incoming":
      return ago ? `asked you ${ago}` : "asked you";
    case "outgoing":
      return ago ? `you asked ${ago}` : "you asked";
    case "blocked":
      return ago ? `blocked ${ago}` : "blocked";
  }
}

function agoText(at: number, now: number): string {
  if (!at || at <= 0) return "";
  const days = Math.floor((now - at) / 86_400_000);
  if (days <= 0) return "today";
  if (days === 1) return "yesterday";
  if (days < 30) return `${days} days ago`;
  const months = Math.floor(days / 30);
  if (months < 12) return months === 1 ? "a month ago" : `${months} months ago`;
  const years = Math.floor(days / 365);
  return years <= 1 ? "a year ago" : `${years} years ago`;
}

/**
 * playedWithRest is the Played with section once friends are listed above
 * it: the people who are not your friends yet. A friend shows once, under
 * Friends.
 */
export function playedWithRest(mates: Tablemate[]): Tablemate[] {
  return mates.filter((m) => !m.friend);
}

/**
 * canBefriendSeat reports whether a seat on your own table offers Add
 * friend (ADR 0213 §2): you sit at the table (`mySeatNo`, through this
 * session or another, as the server checks by account), and the seat is
 * another signed-in person (a Discord identity), not a bot, an agent or a guest.
 */
export function canBefriendSeat(
  seat: { seat: number; discord_id?: string; is_bot?: boolean; is_agent?: boolean },
  mySeatNo: number | null | undefined,
): boolean {
  if (mySeatNo === null || mySeatNo === undefined || seat.seat === mySeatNo) return false;
  if (seat.is_bot || seat.is_agent) return false;
  return !!seat.discord_id;
}
