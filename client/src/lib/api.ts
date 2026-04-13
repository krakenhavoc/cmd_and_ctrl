import { authFetch, currentSession, setSession, type Session } from "./session";

// GameMeta mirrors lobby.GameMeta in server/internal/lobby/lobby.go.
export interface GameMeta {
  id: string;
  name: string;
  created_at: string;
  invite_token?: string;
  players: SeatInfo[];
  state: "lobby" | "active" | "ended";
}

export interface SeatInfo {
  player_id: string;
  name: string;
  seat: number;
}

interface SessionResponse {
  token: string;
  expires_at: string;
  principal: Session["principal"];
  game?: GameMeta;
  player_id?: string;
}

// adminLogin exchanges the shared admin token for a session and
// stores it. Throws LobbyApiError on wrong token.
export async function adminLogin(token: string): Promise<Session> {
  const res = await authFetch("/admin/login", {
    method: "POST",
    body: JSON.stringify({ token }),
  });
  const body = (await res.json()) as SessionResponse;
  const s: Session = {
    token: body.token,
    expiresAt: body.expires_at,
    principal: body.principal,
  };
  setSession(s);
  return s;
}

export async function listGames(): Promise<GameMeta[]> {
  const res = await authFetch("/games", { method: "GET" });
  const body = (await res.json()) as { games: GameMeta[] };
  return body.games;
}

export async function getGame(id: string): Promise<GameMeta> {
  const res = await authFetch(`/games/${id}`, { method: "GET" });
  return (await res.json()) as GameMeta;
}

export async function createGame(name: string): Promise<GameMeta> {
  const res = await authFetch("/games", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
  return (await res.json()) as GameMeta;
}

// joinGame is the public invite-link flow: call with the invite
// token and a display name, receive a fresh RolePlayer session bound
// to (gameID, playerID). The session is installed in the session
// store on success so subsequent API calls use it.
export async function joinGame(
  gameID: string,
  inviteToken: string,
  name: string,
): Promise<Session> {
  const res = await authFetch(`/games/${gameID}/join`, {
    method: "POST",
    body: JSON.stringify({ invite_token: inviteToken, name }),
  });
  const body = (await res.json()) as SessionResponse;
  const s: Session = {
    token: body.token,
    expiresAt: body.expires_at,
    principal: body.principal,
    playerID: body.player_id,
    gameID: body.game?.id,
  };
  setSession(s);
  return s;
}

export async function startGame(id: string): Promise<GameMeta> {
  const res = await authFetch(`/games/${id}/start`, { method: "POST" });
  return (await res.json()) as GameMeta;
}

// logout revokes the current session server-side and clears local
// state. Best-effort: a network failure still clears the store so
// the user isn't stranded in a half-logged-out UI. We bypass
// authFetch because its 401 handler would double-clear the session
// and throw — /logout accepts stale credentials and always returns
// 204, so there's nothing to interpret from the body.
export async function logout(): Promise<void> {
  const s = currentSession();
  try {
    await fetch("/logout", {
      method: "POST",
      headers: s?.token ? { Authorization: `Bearer ${s.token}` } : {},
      credentials: "same-origin",
    });
  } catch {
    // Swallow network errors — we still want to drop the local
    // session so the UI recovers.
  }
  setSession(null);
}
