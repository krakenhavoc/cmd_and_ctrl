import { get, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";

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
  session.set(null);
  expiryNotice.set("Your session expired — please sign in again.");
}

// canSignOutEverywhere reports whether "sign out everywhere" means
// anything for s: only a session tied to a user (user_id) can be
// revoked server-side (POST /logout/everywhere, ADR 0051 decision 6).
// A guest, admin or spectator session, or any session on a server with
// no user database, gets the plain sign-out only.
export function canSignOutEverywhere(s: Session | null): boolean {
  return Boolean(s?.principal.user_id);
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
export function setSession(s: Session | null): void {
  if (s !== null) expiryNotice.set("");
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
    setSession(null);
    throw new LobbyApiError(401, "session expired");
  }
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`;
    let violations: ApiViolation[] | undefined;
    let warnings: ApiViolation[] | undefined;
    try {
      const body = (await res.clone().json()) as {
        error?: string;
        violations?: ApiViolation[];
        warnings?: ApiViolation[];
      };
      if (body.error) message = body.error;
      if (Array.isArray(body.violations)) violations = body.violations;
      if (Array.isArray(body.warnings)) warnings = body.warnings;
    } catch {
      // body wasn't JSON — keep the default message
    }
    throw new LobbyApiError(res.status, message, violations, warnings);
  }
  return res;
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
  ) {
    super(message);
    this.name = "LobbyApiError";
  }
}
