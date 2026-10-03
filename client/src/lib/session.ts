import { get, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { signedInUserID } from "./myGames";
import { navigate, route } from "./router";

// Session is the client-side view of a server-issued Principal.
// Mirrors auth.Principal in server/internal/auth/auth.go — new
// fields should be added in lockstep with the server.
export interface Session {
  token: string;
  expiresAt: string; // ISO 8601, for display
  principal: {
    // "identified" is a Discord sign-in that has not claimed a seat
    // yet (server: auth.RoleIdentified). It can do exactly one
    // thing — POST /join with an invite code, which swaps it for a
    // player session — so no game-facing route ever sees one.
    role: "player" | "admin" | "spectator" | "identified";
    // user_id is the server's users-table row (ADR 0051 decision 3).
    // Present on a Discord sign-in, and on a seat claimed from one,
    // when the server has a user database; absent (or the nil uuid)
    // for admin, guest and spectator sessions, and for everyone on a
    // server with no database. Its presence is what makes the session
    // revocable, so it gates "sign out everywhere"
    // (canSignOutEverywhere); lib/myGames.ts's signedInUserID gates
    // every "My games" link on it.
    user_id?: string;
    admin_id?: string;
    game_id?: string;
    player_id?: string;
    name?: string;
    issued_at: string;
    expires_at: string;
  };
  // For RolePlayer sessions: the server echoes the game metadata +
  // player ID on the join response. Stored locally so the Game route
  // can dial /ws with the correct ?game=&player= without another
  // round-trip.
  playerID?: string;
  gameID?: string;
  // admin is GET /me's computed answer for a session with a user: true
  // when the person's Discord ID is on the server's admin allowlist
  // (ADR 0110 §3). It is not in the token and the server never trusts
  // it. Absent until /me has answered; read it through lib/admin.ts's
  // isAdmin, which also covers the shared token's `role: "admin"`.
  admin?: boolean;
}

const STORAGE_KEY = "cmdctrl.session";

// Read any previously-persisted session out of localStorage. Runs
// once at module load — components just read the store.
function loadSession(): Session | null {
  if (typeof localStorage === "undefined") return null;
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    const s = JSON.parse(raw) as Session;
    // Drop expired sessions at load time so the UI doesn't flash
    // "logged in" before a 401 hits the next API call.
    if (Date.parse(s.expiresAt) < Date.now()) return null;
    return s;
  } catch {
    return null;
  }
}

export const session: Writable<Session | null> = guardedWritable(loadSession(), "session");

// expiryNotice is raised when we clear the session proactively
// because its TTL elapsed. The Login route surfaces this as a banner
// so the user understands why they landed back on login instead of
// mid-game. Cleared on the next successful setSession.
export const expiryNotice: Writable<string> = guardedWritable("", "expiryNotice");

// Track the pending expiry timer so we cancel + rearm it on every
// setSession call. Module-scoped rather than per-subscriber so there
// is always exactly one timer in flight.
let expiryTimer: ReturnType<typeof setTimeout> | null = null;

function scheduleExpiry(s: Session | null): void {
  if (expiryTimer !== null) {
    clearTimeout(expiryTimer);
    expiryTimer = null;
  }
  scheduleRenewal(s);
  if (s === null) return;
  const ms = Date.parse(s.expiresAt) - Date.now();
  if (ms <= 0) {
    // Already expired at setSession time — clear on the next tick so
    // subscribers see the setSession first, then the clear.
    expiryTimer = setTimeout(() => expireSession(), 0);
    return;
  }
  // setTimeout overflows past 2^31-1 ms (~24.8 days) and fires at
  // once, so the delay is clamped. A Discord sign-in lives 30 days
  // (CMDCTRL_IDENTITY_TTL, ADR 0051 decision 3), which is past the
  // clamp — so when the clamped timer fires early, it re-arms for the
  // remainder instead of clearing a session that is still good.
  expiryTimer = setTimeout(() => onExpiryTimer(s), Math.min(ms, MAX_TIMER_MS));
}

// MAX_TIMER_MS is the longest delay handed to setTimeout, below the
// 2^31-1 overflow with room to spare.
export const MAX_TIMER_MS = 2_000_000_000;

function onExpiryTimer(s: Session): void {
  expiryTimer = null;
  // Only the session this timer was armed for. scheduleExpiry cancels
  // a stale timer when a new session is installed, so this is belt and
  // braces: a mismatch is a no-op, never a logout.
  if (get(session) !== s) return;
  if (Date.parse(s.expiresAt) > Date.now()) {
    scheduleExpiry(s);
    return;
  }
  expireSession();
}

function expireSession(): void {
  expiryTimer = null;
  const dead = get(session);
  if (dead === null) return;
  sessionDied(dead, EXPIRED_NOTICE);
}

const EXPIRED_NOTICE = "Your session expired — please sign in again.";

// --- the saved identity (ADR 0110 §1 item 6) -------------------------
//
// A browser holds one session at a time. When it installs a session
// with no user (the admin token, a reclaim ticket for a seat that is not
// yours) over a signed-in person's session, it keeps the person's
// session aside under IDENTITY_KEY first. When the session with no user
// ends (it expires, or the server answers 401 for it), the saved one is
// put back through POST /me/session, sent with the saved token as its
// bearer. The browser has already dropped the expired cookie, so the
// bearer is the credential the server reads, and the server sets the
// cookie back to it. The person is still signed in instead of being sent
// to #/login. Leaving a practice table whose own saved session has gone
// puts it back the same way (practiceTable.ts's endPractice).
//
// localStorage, like the session itself: readable by this origin only.
// Signing out, or signing out everywhere, clears it (api.ts's logout and
// logoutEverywhere call clearSavedIdentity). Installing any session with
// a user clears it too, since that session is now the person's.

export const IDENTITY_KEY = "cmdctrl.identity";

function isExpired(s: Session, now: number = Date.now()): boolean {
  return !(Date.parse(s.expiresAt) > now);
}

// savedIdentity returns the signed-in session kept aside, or null when
// there is none, it is malformed, it has expired, or it is not a
// person's.
export function savedIdentity(now: number = Date.now()): Session | null {
  let raw: string | null = null;
  try {
    raw = localStorage.getItem(IDENTITY_KEY);
  } catch {
    return null;
  }
  if (!raw) return null;
  try {
    const s = JSON.parse(raw) as Session;
    if (typeof s?.token !== "string" || s.token === "" || !s.principal) return null;
    if (isExpired(s, now) || signedInUserID(s) === null) {
      clearSavedIdentity();
      return null;
    }
    return s;
  } catch {
    return null;
  }
}

function saveIdentity(s: Session): void {
  try {
    localStorage.setItem(IDENTITY_KEY, JSON.stringify(s));
  } catch {
    // Best-effort: without it this browser falls back to #/login when
    // the session with no user ends, which is what happened before.
  }
}

// clearSavedIdentity forgets the session kept aside. Sign-out calls it.
export function clearSavedIdentity(): void {
  try {
    localStorage.removeItem(IDENTITY_KEY);
  } catch {
    // Nothing to do.
  }
}

// RenewResponse is POST /me/session's body (lobby.sessionResponse).
interface RenewResponse {
  token: string;
  expires_at: string;
  principal: Session["principal"];
}

// renewedFrom is `from` with the token and principal the server
// returned. The principal is the same one with new timestamps, so the
// seat fields the client stored for `from` still hold.
function renewedFrom(from: Session, body: RenewResponse): Session {
  return { ...from, token: body.token, expiresAt: body.expires_at, principal: body.principal };
}

// postMeSession is POST /me/session with `token` as the bearer: the
// server sets the cookie to the session it returns, renewed first when
// more than half spent. Plain fetch, not authFetch: a refusal here is an
// answer about one token, and must never sign the tab out by itself.
async function postMeSession(token: string): Promise<RenewResponse> {
  const res = await fetch("/me/session", {
    method: "POST",
    headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
    credentials: "same-origin",
  });
  if (!res.ok) throw new LobbyApiError(res.status, `${res.status} ${res.statusText}`);
  const body = (await res.json()) as RenewResponse;
  if (typeof body?.token !== "string" || body.token === "" || !body.principal) {
    throw new LobbyApiError(500, "malformed session response");
  }
  return body;
}

let reinstalling: Promise<void> | null = null;

// sessionDied handles the end of the session the tab is holding: its
// expiry timer fired, or the server answered 401 for it. With a saved
// identity behind a session that had no user, it reinstalls that
// identity. Otherwise, or when the reinstall fails, it clears the
// session and shows `notice` on the login page. While the reinstall is
// in flight the dead session stays in the store, so the router does not
// flash the login page in between.
function sessionDied(dead: Session, notice: string): void {
  const saved = signedInUserID(dead) === null ? savedIdentity() : null;
  if (saved === null) {
    session.set(null);
    if (notice) expiryNotice.set(notice);
    return;
  }
  if (reinstalling !== null) return;
  reinstalling = (async () => {
    try {
      const body = await postMeSession(saved.token);
      if (get(session) !== dead) return;
      if (!samePrincipal(body.principal, saved.principal)) {
        // The cookie still held a live session, and the server read it
        // before the bearer. It is not the one saved; do not adopt it.
        throw new Error("the server answered for a different session");
      }
      const next = renewedFrom(saved, body);
      setSession(next);
      // The dead session's table is not the person's: a seat ticket's
      // game, or a game the admin token was watching.
      const r = get(route);
      if (r.name === "game" && r.gameID !== next.gameID) navigate("#/lobby");
    } catch (err) {
      if (err instanceof LobbyApiError && (err.status === 401 || err.status === 403)) {
        clearSavedIdentity();
      }
      if (get(session) === dead) {
        session.set(null);
        if (notice) expiryNotice.set(notice);
      }
    } finally {
      reinstalling = null;
    }
  })();
}

// --- renewal on use (ADR 0110 §1 item 5, owner answer 1) -------------
//
// The server re-issues a signed-in session more than half spent for a
// fresh CMDCTRL_IDENTITY_TTL when the client calls POST /me/session. The
// client calls it when the session it holds passes half its life: at
// page load if it already has (the store's first subscriber arms the
// timer with no delay), and otherwise on a timer armed for the half-life
// beside the expiry timer. The server makes the same test on the same
// two fields (server/internal/lobby/session_renew.go's halfSpent), so a
// client clock a little ahead of the server's only costs a no-op answer,
// and the client asks again an hour later.
//
// A failed renewal is quiet: no notice, no sign-out. The session still
// lasts until it expires, and the client tries again in an hour.

// RENEW_RETRY_MS is how long the client waits to ask again after a
// renewal that renewed nothing or failed.
export const RENEW_RETRY_MS = 60 * 60 * 1000;

let renewTimer: ReturnType<typeof setTimeout> | null = null;
let renewing: Promise<void> | null = null;

// pastHalfLife reports whether more than half of s's lifetime, from
// its issued_at to its expiry, has passed at now.
export function pastHalfLife(s: Session, now: number = Date.now()): boolean {
  const issued = Date.parse(s.principal.issued_at);
  const expires = Date.parse(s.expiresAt);
  if (!Number.isFinite(issued) || !Number.isFinite(expires)) return false;
  return now - issued > (expires - issued) / 2;
}

function scheduleRenewal(s: Session | null, delay?: number): void {
  if (renewTimer !== null) {
    clearTimeout(renewTimer);
    renewTimer = null;
  }
  if (s === null || signedInUserID(s) === null) return;
  let ms = delay;
  if (ms === undefined) {
    const issued = Date.parse(s.principal.issued_at);
    const expires = Date.parse(s.expiresAt);
    if (!Number.isFinite(issued) || !Number.isFinite(expires)) return;
    ms = Math.max(0, issued + (expires - issued) / 2 - Date.now() + 1);
  }
  renewTimer = setTimeout(
    () => {
      renewTimer = null;
      if (get(session) !== s) return;
      if (!pastHalfLife(s)) {
        scheduleRenewal(s);
        return;
      }
      void renewSession();
    },
    Math.min(ms, MAX_TIMER_MS),
  );
}

// sessionSettled resolves once no renewal or reinstall is in flight.
//
// Both set the session cookie when their answer arrives. A request that
// mints another session (a join, a reclaim) and is sent while one of
// them is in flight could have its own cookie overwritten by the later
// answer, leaving the cookie on the old session while the tab holds the
// new one. So those requests wait: authFetch does, and so does every
// plain-fetch mint in api.ts. A renewal is one short request, so the
// wait is too.
export async function sessionSettled(): Promise<void> {
  while (sessionBusy()) {
    await (renewing ?? reinstalling)?.catch(() => {});
  }
}

function sessionBusy(): boolean {
  return renewing !== null || reinstalling !== null;
}

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

function idOrNone(id: string | undefined): string {
  return !id || id === NIL_UUID ? "" : id;
}

// samePrincipal reports whether a and b name the same session holder:
// the role, the user, and the game and seat.
function samePrincipal(a: Session["principal"], b: Session["principal"]): boolean {
  return (
    a.role === b.role &&
    idOrNone(a.user_id) === idOrNone(b.user_id) &&
    idOrNone(a.game_id) === idOrNone(b.game_id) &&
    idOrNone(a.player_id) === idOrNone(b.player_id)
  );
}

// renewSession asks the server to renew the session the tab is holding
// (POST /me/session) and installs the answer when it is a new token for
// the same principal. Never throws and never signs the tab out.
//
// The server reads the cookie before the bearer, so a tab whose cookie
// another tab has since moved to a different session (a seat at another
// table) gets that session back. It is not this tab's, so it is not
// installed: the answer is treated as a no-op.
export function renewSession(): Promise<void> {
  if (renewing !== null) return renewing;
  const s = get(session);
  if (s === null || signedInUserID(s) === null || isExpired(s)) return Promise.resolve();
  renewing = (async () => {
    try {
      const body = await postMeSession(s.token);
      if (get(session) !== s) return;
      if (body.token === s.token || !samePrincipal(body.principal, s.principal)) {
        scheduleRenewal(s, RENEW_RETRY_MS);
        return;
      }
      setSession(renewedFrom(s, body));
    } catch {
      if (get(session) === s) scheduleRenewal(s, RENEW_RETRY_MS);
    } finally {
      renewing = null;
    }
  })();
  return renewing;
}

// canSignOutEverywhere reports whether "sign out everywhere" means
// anything for s: only a session tied to a user (user_id) can be
// revoked server-side (POST /logout/everywhere, ADR 0051 decision 6).
// A guest, admin or spectator session, or any session on a server with
// no user database, gets the plain sign-out only. A guest's principal
// spells "no user" as the nil uuid, so this asks signedInUserID rather
// than whether the field is set (ADR 0112 §1 item 4 puts this in the
// header, where every guest sees it).
export function canSignOutEverywhere(s: Session | null): boolean {
  return signedInUserID(s) !== null;
}

// OAuthFragment is what the server's Discord callback puts in the
// #/oauth-complete fragment (router.ts parses it).
export interface OAuthFragment {
  token: string;
  expiresAt: string;
  gameID?: string;
  playerID?: string;
  displayName?: string;
  userID?: string;
}

// sessionFromOAuth builds the Session the SPA installs from the
// oauth-complete fragment, without a round trip to /me. With game and
// player_id the seat is already claimed (a player session); with
// neither it is the identity-only session from the login page.
export function sessionFromOAuth(f: OAuthFragment, now: Date = new Date()): Session {
  const seated = Boolean(f.gameID && f.playerID);
  return {
    token: f.token,
    expiresAt: f.expiresAt,
    principal: {
      role: seated ? "player" : "identified",
      user_id: f.userID,
      game_id: f.gameID,
      player_id: f.playerID,
      name: f.displayName,
      issued_at: now.toISOString(),
      expires_at: f.expiresAt,
    },
    playerID: f.playerID,
    gameID: f.gameID,
  };
}

// Persist store writes to localStorage so a page reload restores
// the session. Also rearms the expiry timer so a long-running tab
// clears state at the TTL rather than quietly 401-ing mid-match.
session.subscribe((s) => {
  scheduleExpiry(s);
  if (typeof localStorage === "undefined") return;
  // #720: persistence is best-effort, the same as settings.ts's
  // saveSettings. setItem throws on a full quota and in Safari
  // private browsing, and this runs inside a subscribe callback —
  // i.e. inside svelte/store's shared drain loop. `session` is
  // guarded now, so a throw here could not freeze the app, but it
  // would still cost the expiry timer its rearm on the way past, and
  // a failed WRITE is no reason to lose the live session.
  try {
    if (s === null) localStorage.removeItem(STORAGE_KEY);
    else localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
  } catch {
    // Swallowed: the store still holds the session for this tab.
  }
});

// setSession replaces the current session with s and persists it.
// Clears any pending expiry notice on a fresh login.
//
// It also keeps the saved identity (ADR 0110 §1 item 6): a session with
// no user installed over a signed-in person's live session moves that
// session aside first, and a session with a user replaces whatever was
// aside. Clearing the session (s === null) leaves the saved copy alone;
// signing out clears it explicitly.
export function setSession(s: Session | null): void {
  if (s !== null) {
    expiryNotice.set("");
    const prev = get(session);
    if (signedInUserID(s) !== null) {
      clearSavedIdentity();
    } else if (prev !== null && signedInUserID(prev) !== null && !isExpired(prev)) {
      saveIdentity(prev);
    }
  }
  session.set(s);
}

// currentSession returns the current value synchronously. For use in
// non-reactive code paths (e.g. inside async functions that don't
// want to open a subscription).
export function currentSession(): Session | null {
  return get(session);
}

// authFetch wraps fetch() to attach the bearer token when a session
// is live and to surface JSON error bodies as rejected promises so
// callers don't have to manually check response.ok everywhere.
//
// On 401, it clears the session so the router re-renders to the
// login page. On every other non-2xx it throws a LobbyApiError with
// the server's error message.
export async function authFetch(input: string, init: RequestInit = {}): Promise<Response> {
  // Only when one is in flight: an idle tab's request goes out at once.
  if (sessionBusy()) await sessionSettled();
  const s = currentSession();
  const headers = new Headers(init.headers);
  if (s?.token) headers.set("Authorization", `Bearer ${s.token}`);
  // FormData must set its own Content-Type: the boundary token is
  // generated by the browser and a hand-set header omits it, which
  // makes the server's multipart parser fail on a body that is in fact
  // well-formed. Everything else here is JSON.
  if (!headers.has("Content-Type") && init.body && !(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }
  const res = await fetch(input, { ...init, headers, credentials: "same-origin" });
  if (res.status === 401) {
    dropDeadSession(s);
    throw new LobbyApiError(401, "session expired");
  }
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`;
    let violations: ApiViolation[] | undefined;
    let warnings: ApiViolation[] | undefined;
    let parsed: unknown;
    try {
      const body = (await res.clone().json()) as {
        error?: string;
        violations?: ApiViolation[];
        warnings?: ApiViolation[];
      };
      parsed = body;
      if (body.error) message = body.error;
      if (Array.isArray(body.violations)) violations = body.violations;
      if (Array.isArray(body.warnings)) warnings = body.warnings;
    } catch {
      // body wasn't JSON — keep the default message
    }
    throw new LobbyApiError(res.status, message, violations, warnings, parsed);
  }
  return res;
}

// dropDeadSession is what a 401 means: the server's own word that the
// session `sent` is gone. The tab's session is dropped or, when it had
// no user and a saved identity is waiting behind it, swapped for that
// identity (sessionDied). A session installed while the request was in
// flight is left alone, because the 401 was about a different token.
export function dropDeadSession(sent: Session | null): void {
  const live = get(session);
  if (live === null) return;
  if (sent !== null && live.token !== sent.token) return;
  sessionDied(live, "");
}

// SessionCheckResult is the answer to "is this session still good, as
// far as the server is concerned?" (#1475). "dead" is a definite
// answer — a 401, the server's own word that the credential is gone —
// and everything else is "unknown": a 5xx, a network failure, a
// timeout. ws.ts's reconnect ladder treats "unknown" exactly like
// "still trying", never like "dead", because a server that is merely
// restarting must not be mistaken for a revoked session.
export type SessionCheckResult = "alive" | "dead" | "unknown";

// checkSessionAlive asks GET /me — the cheapest authenticated route
// there is, and the one the server doc-comments "for client bootstrap"
// (server/cmd/server/main.go) — whether the session this tab is
// holding is one the server will still accept. authFetch already
// clears a 401'd session locally as a side effect; this just reports
// which of the three outcomes happened so a caller (ws.ts's dead-
// session probe) can act on the DEFINITE case only.
export async function checkSessionAlive(): Promise<SessionCheckResult> {
  try {
    await authFetch("/me");
    return "alive";
  } catch (err) {
    if (err instanceof LobbyApiError && err.status === 401) return "dead";
    return "unknown";
  }
}

// ApiViolation mirrors deck.Violation on the server. Kept here (not
// in api.ts) so LobbyApiError can carry the structured list without
// a circular import between session and api.
export interface ApiViolation {
  code: string;
  message: string;
  card?: string;
}

export class LobbyApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public violations?: ApiViolation[],
    public warnings?: ApiViolation[],
    // body is the whole parsed JSON error body, for the refusals that
    // carry data beside the message: PUT /me/settings' 412 and 409
    // return the account's current copy (ADR 0110 §4).
    public body?: unknown,
  ) {
    super(message);
    this.name = "LobbyApiError";
  }
}
