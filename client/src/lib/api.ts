import { authFetch, currentSession, setSession, type ApiViolation, type Session } from "./session";

// Re-export the violation shape so consumers of api.ts don't also
// have to import from session.ts. ApiViolation is the canonical
// name; DeckViolation is kept as an alias for existing callers.
export type DeckViolation = ApiViolation;
export type { ApiViolation };

// GameMeta mirrors lobby.GameMeta in server/internal/lobby/lobby.go.
export interface GameMeta {
  id: string;
  name: string;
  created_at: string;
  invite_token?: string;
  // Per-game spectator invite. Distinct from invite_token — sharing
  // the player invite with a spectator would let them claim a seat.
  // Only emitted to the admin and seated players; stripped from list
  // responses + spectator-session responses. Added in S11.
  spectator_invite?: string;
  players: SeatInfo[];
  state: "lobby" | "active" | "ended";
}

export interface SeatInfo {
  player_id: string;
  name: string;
  seat: number;
  deck_name?: string;
  deck_uploaded: boolean;
}

// UploadDeckResponse mirrors lobby.uploadDeckResponse.
export interface UploadDeckResponse {
  game: GameMeta;
  deck_name: string;
  card_count: number;
  commanders: string[];
  warnings?: ApiViolation[];
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
  // Defensive normalisation: a buggy server may emit `players: null`
  // for seat-less games (see lobby.copyMeta regression). Iterating
  // `g.players` then throws mid-render and Svelte silently bails on
  // the subtree — "no games yet" keeps showing. Coercing to [] here
  // means the UI degrades to a visible empty-seats row instead of a
  // vanished list.
  return body.games.map((g) => ({ ...g, players: g.players ?? [] }));
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

// spectateGame is the read-only counterpart to joinGame: posts the
// per-game spectator invite, receives a RoleSpectator session bound
// to the game (no player_id). The session is installed in the
// session store; the Game route uses session.principal.role to hide
// action affordances. Added in S11.
export async function spectateGame(
  gameID: string,
  inviteToken: string,
  name: string,
): Promise<Session> {
  const res = await authFetch(`/games/${gameID}/spectate`, {
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

// replayURL returns an authenticated download URL for a game's
// replay JSONL. Ships the session token as ?token= because browser
// downloads can't set Authorization headers. Consumed by the lobby
// "download replay" link via a plain <a href>. Added in S11.
export function replayURL(gameID: string): string | null {
  const s = currentSession();
  if (!s?.token) return null;
  return `/games/${gameID}/replay?token=${encodeURIComponent(s.token)}`;
}

// uploadDeck ships a decklist (plain text or Moxfield JSON) to the
// server for parse + validate + install. Format can be omitted — the
// server auto-detects by checking the first non-whitespace byte for
// '{' (Moxfield) vs anything else (text).
//
// On validation failure (422), authFetch throws a LobbyApiError whose
// `.violations` carries the full structured list (one per offending
// card) and `.warnings` carries non-fatal advisories (e.g. sideboard
// ignored). Callers that want to highlight individual rows should
// read `.violations`; the plain `.message` is the human-readable
// summary.
export async function uploadDeck(
  gameID: string,
  playerID: string,
  source: string,
  format?: "text" | "moxfield",
): Promise<UploadDeckResponse> {
  const res = await authFetch(`/games/${gameID}/decks`, {
    method: "POST",
    body: JSON.stringify({ player_id: playerID, source, format }),
  });
  return (await res.json()) as UploadDeckResponse;
}

// avatarURL builds the cached-Discord-avatar URL for a (discord_id,
// avatar_hash) pair. Returns null when either value is missing so
// callers can use it as a render gate:
//   const url = avatarURL(seat.discord_id, seat.discord_avatar_hash);
//   {#if url}<img src={url}>{/if}
// The endpoint is session-gated and <img> tags can't set an
// Authorization header, so the token rides as ?token= (same pattern
// as replayURL) — the session cookie alone isn't reliable (Secure-
// flag mismatches, cleared cookies with a live localStorage session).
// The /avatars handler 503s when the disk cache is unconfigured;
// callers that care should treat a 503 as "fall back to initials".
export function avatarURL(discordID?: string, avatarHash?: string): string | null {
  if (!discordID || !avatarHash) return null;
  const base = `/avatars/${encodeURIComponent(discordID)}/${encodeURIComponent(avatarHash)}.png`;
  const token = currentSession()?.token;
  return token ? `${base}?token=${encodeURIComponent(token)}` : base;
}

// discordAuthEnabled probes /auth/discord/config and reports whether
// the server has the Discord OAuth three-tuple configured. Used by
// Join.svelte to decide whether to render the "Sign in with Discord"
// button. A 503 / network failure → false; the manual-name fallback
// is always safe, so a down probe shouldn't block the join flow.
export async function discordAuthEnabled(): Promise<boolean> {
  try {
    const res = await fetch("/auth/discord/config", {
      method: "GET",
      credentials: "same-origin",
    });
    if (!res.ok) return false;
    const body = (await res.json()) as { enabled?: boolean };
    return body.enabled === true;
  } catch {
    return false;
  }
}

// AutoTapPreview mirrors the JSON returned by
// `GET /games/:id/auto-tap-preview` (server/internal/lobby/http.go).
// `ok=false` means the auto-tapper couldn't satisfy the cost; the
// caller should keep showing the missing list and disable the
// "Auto-tap & cast" submit affordance. `plan` is the ordered list
// of permanent UUIDs the server proposes to tap; the modal renders
// them in the order returned (colored requirements first, generic
// recruits second). Added in S15 sub-PR 5.
export interface AutoTapPreview {
  ok: boolean;
  plan?: string[];
  missing?: string[];
  cost: string;
}

// fetchAutoTapPreview asks the server which permanents the auto-
// tapper would tap to cast `cardID` right now. Read-only — calling
// it does not mutate game state. `excluded` lets the caller pass
// the lock-tap UI's reservation list. `xValue` is the announced
// X for spells with {X} in their cost (defaults to 0).
export async function fetchAutoTapPreview(
  gameID: string,
  cardID: string,
  opts: { xValue?: number; excluded?: string[] } = {},
): Promise<AutoTapPreview> {
  const params = new URLSearchParams({ card: cardID });
  if (opts.xValue && opts.xValue > 0) {
    params.set("x", String(opts.xValue));
  }
  if (opts.excluded && opts.excluded.length > 0) {
    params.set("exclude", opts.excluded.join(","));
  }
  const res = await authFetch(`/games/${gameID}/auto-tap-preview?${params.toString()}`, {
    method: "GET",
  });
  return (await res.json()) as AutoTapPreview;
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
