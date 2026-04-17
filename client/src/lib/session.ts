import { writable, get, type Writable } from "svelte/store";

// Session is the client-side view of a server-issued Principal.
// Mirrors auth.Principal in server/internal/auth/auth.go — new
// fields should be added in lockstep with the server.
export interface Session {
  token: string;
  expiresAt: string; // ISO 8601, for display
  principal: {
    role: "player" | "admin";
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

export const session: Writable<Session | null> = writable(loadSession());

// expiryNotice is raised when we clear the session proactively
// because its TTL elapsed. The Login route surfaces this as a banner
// so the user understands why they landed back on login instead of
// mid-game. Cleared on the next successful setSession.
export const expiryNotice: Writable<string> = writable("");

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
  // setTimeout caps at ~24 days on most runtimes — well above our
  // 12h default TTL. Clamp defensively so tabs left open across a
  // browser suspend don't fire with a negative delay on resume.
  expiryTimer = setTimeout(() => expireSession(), Math.min(ms, 2_000_000_000));
}

function expireSession(): void {
  expiryTimer = null;
  session.set(null);
  expiryNotice.set("Your session expired — please sign in again.");
}

// Persist store writes to localStorage so a page reload restores
// the session. Also rearms the expiry timer so a long-running tab
// clears state at the TTL rather than quietly 401-ing mid-match.
session.subscribe((s) => {
  scheduleExpiry(s);
  if (typeof localStorage === "undefined") return;
  if (s === null) localStorage.removeItem(STORAGE_KEY);
  else localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
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
  if (!headers.has("Content-Type") && init.body) {
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
