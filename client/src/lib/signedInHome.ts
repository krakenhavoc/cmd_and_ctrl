import type { GameMeta } from "./api";
import { isDecksReturn } from "./decksPage";
import { canJoinByCode } from "./myGames";
import type { Route } from "./router";
import type { Session } from "./session";

// signedInHome.ts is the pure half of ADR 0112 §1, "the signed-in
// home": which routes a signed-out visitor may open, where a session
// on #/login or #/admin goes, where the Discord round trip lands, which
// join box the Lobby offers, and which tables it lists. No DOM and no
// fetch, so the whole decision table is unit-tested
// (signedInHome.test.ts); App.svelte and Lobby.svelte only render the
// answers.

// The routes a signed-out visitor may open.
const PUBLIC_ROUTES: ReadonlySet<Route["name"]> = new Set<Route["name"]>([
  "login",
  "adminLogin",
  // An invite or spectator link is how a new player arrives.
  "join",
  // The whole point of a reclaim link is that the holder has no session
  // yet: gating it behind one would bounce them to the login page they
  // cannot get past.
  "reclaim",
  "oauthComplete",
  // The site portal and the public roadmap (#1386): neither shows card
  // art or anything else session-gated.
  "home",
  "roadmap",
  // The one decks page (ADR 0112 §3 item 1), and its #/deck-check
  // alias: a signed-out visitor can check a deck, for the same reason
  // the roadmap is public, and the Discord bot links straight here.
  // The library on it shows only to a signed-in person.
  "decks",
]);

// isPublicRoute reports whether a signed-out visitor may open r.
export function isPublicRoute(r: Route): boolean {
  return PUBLIC_ROUTES.has(r.name);
}

// routeRedirect is the router's auth gate: where to send a visitor on
// route r holding session s, or null to stay.
//
//   - A signed-out visitor on a gated route goes to #/login.
//   - Every session on #/login goes to the Lobby (ADR 0112 §1 item 1):
//     Login is for signed-out visitors only. The Lobby has the header
//     and the join box now, so nobody is stranded one step short of a
//     table. An expired session is cleared before this runs, so that
//     visitor stays on #/login and sees the expiry notice.
//   - #/admin is the shared token's page (§2 item 8). The token's own
//     session has no use for it and goes to the Lobby; every other
//     session sees the token form, which says it replaces this
//     browser's session. (An allowlisted person also goes to the Lobby,
//     where the Admin chip is; that needs /me's admin_allowed.)
//
// Invite, spectator and reclaim links and both shapes of the Discord
// round trip are public and never redirected, whatever the browser
// holds (§1 item 2).
export function routeRedirect(r: Route, s: Session | null): string | null {
  if (!s) return isPublicRoute(r) ? null : "#/login";
  if (r.name === "login") return "#/lobby";
  if (r.name === "adminLogin" && s.principal.role === "admin") return "#/lobby";
  return null;
}

// oauthCompleteTarget is where a session installed from the Discord
// callback goes: the invite flow (a claimed seat) to its table, and the
// login-page flow (an identity-only session) to the Lobby, the
// signed-in home, or back to the decks page when that is where the
// sign-in started (ADR 0112 §3 item 7). `afterSignIn` is the route the
// decks page stored (decksPage.takeAfterSignIn), and it is followed
// only if it parses as the decks page.
export function oauthCompleteTarget(s: Session, afterSignIn?: string | null): string {
  if (s.principal.role === "player" && s.gameID) return `#/games/${s.gameID}`;
  if (isDecksReturn(afterSignIn)) return afterSignIn;
  return "#/lobby";
}

// JoinBox is the Lobby's "Join a table" card for a session (§1 item 3):
//
//   - "code": a signed-in person pastes a code or a link. A code posts
//     POST /join, which seats them as their Discord identity.
//   - "link": a guest seat or guest spectator pastes a link only. The
//     server refuses a guest's code (409), so the box says so first.
//   - "none": the admin token is not a person and cannot take a seat
//     by code, and a signed-out visitor is not on the Lobby at all.
export type JoinBox = "code" | "link" | "none";

export function joinBoxFor(s: Session | null | undefined): JoinBox {
  if (!s) return "none";
  if (canJoinByCode(s)) return "code";
  if (s.principal.role === "player" || s.principal.role === "spectator") return "link";
  return "none";
}

// GUEST_CODE_MESSAGE is what a guest's box says to a bare code, before
// anything is sent.
export const GUEST_CODE_MESSAGE =
  "This browser is seated as a guest. Open the invite link instead, or link Discord from your table's menu.";

// inviteHash pulls the #/games/…/join?t=… fragment out of a pasted
// invite URL. Returns "" for a bare code, which is the signal to ask
// the server which table the code belongs to instead.
export function inviteHash(raw: string): string {
  if (raw.startsWith("#")) return raw;
  if (!/^https?:\/\//i.test(raw)) return "";
  try {
    return new URL(raw).hash;
  } catch {
    return "";
  }
}

// myTables is what the Lobby lists (§1 item 6, owner answer 3).
//
//   - An admin (the token, or an allowlisted person in admin mode)
//     sees every table, as before.
//   - A signed-in person sees THEIR tables only: the session's own
//     table, every table they created (`is_creator`), and every open
//     table where they hold a seat through another session (`rejoinable`,
//     the GET /me/games entries with a `rejoin` path). Nothing else, and
//     no fallback to the room: an empty list is the empty state.
//   - A guest keeps the view it had: a guest seat sees its own table
//     and any it created, or the room when neither is listed; a guest
//     spectator sees the room.
//
// This is presentation only. GET /games is unchanged, and its invites
// are already redacted per caller on the server (redactMetaFor).
export function myTables(
  games: GameMeta[],
  s: Session | null | undefined,
  opts: { admin: boolean; rejoinable: ReadonlySet<string> },
): GameMeta[] {
  if (opts.admin) return games;
  if (canJoinByCode(s)) {
    return games.filter(
      (g) => g.id === s?.gameID || g.is_creator === true || opts.rejoinable.has(g.id),
    );
  }
  if (s?.principal.role !== "player" || !s.gameID) return games;
  const mine = games.filter((g) => g.id === s.gameID || g.is_creator === true);
  return mine.length > 0 ? mine : games;
}
