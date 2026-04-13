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

// Persist store writes to localStorage so a page reload restores
// the session.
session.subscribe((s) => {
  if (typeof localStorage === "undefined") return;
  if (s === null) localStorage.removeItem(STORAGE_KEY);
  else localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
});

// setSession replaces the current session with s and persists it.
export function setSession(s: Session | null): void {
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
    try {
      const body = (await res.clone().json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // body wasn't JSON — keep the default message
    }
    throw new LobbyApiError(res.status, message);
  }
  return res;
}

export class LobbyApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "LobbyApiError";
  }
}
