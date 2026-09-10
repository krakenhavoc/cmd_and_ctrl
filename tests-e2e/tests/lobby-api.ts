import type { APIRequestContext } from "@playwright/test";
import { ADMIN_TOKEN } from "./env";

// Thin wrapper over the lobby's HTTP surface. The UI already covers
// most happy paths in lobby.spec.ts; this helper is for tests that
// need to bypass the UI — mostly to grab an invite token server-side
// so the join flow can be driven without scraping the clipboard.

export interface SeatInfo {
  player_id: string;
  name: string;
  seat: number;
  deck_name?: string;
  deck_uploaded: boolean;
}

export interface GameMeta {
  id: string;
  name: string;
  created_at: string;
  invite_token?: string;
  // The spectator invite is a distinct token from the player invite:
  // handing a spectator the player link would let them claim a seat.
  // Only present on the create response (List strips both).
  spectator_invite?: string;
  players: SeatInfo[];
  state: "lobby" | "active" | "ended";
}

export async function adminLogin(req: APIRequestContext): Promise<string> {
  const res = await req.post("/admin/login", { data: { token: ADMIN_TOKEN } });
  if (!res.ok()) throw new Error(`admin login failed: ${res.status()}`);
  const body = (await res.json()) as { token: string };
  return body.token;
}

export async function createGame(
  req: APIRequestContext,
  token: string,
  name: string,
): Promise<GameMeta> {
  const res = await req.post("/games", {
    data: { name },
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok()) throw new Error(`create game failed: ${res.status()} ${await res.text()}`);
  return (await res.json()) as GameMeta;
}

// uploadDeckAs uploads a decklist for (gameID, playerID). An admin
// session can upload for any seat; a RolePlayer session must supply
// its own player_id — the server rejects cross-seat uploads with 403.
export async function uploadDeckAs(
  req: APIRequestContext,
  token: string,
  gameID: string,
  playerID: string,
  source: string,
): Promise<{ deck_name: string; card_count: number; commanders: string[] }> {
  // The server reports HTTP-ready before the ~500MB Scryfall index
  // finishes loading and answers uploads with 503 until it has. The
  // webServer readiness URL can't see that, so retry here — bounded,
  // and only for that specific 503 — instead of sleeping in specs.
  // Bounded well under the 90s test timeout so a genuinely missing
  // index reports the 503 rather than a mute test-timeout.
  const deadline = Date.now() + 45_000;
  for (;;) {
    const res = await req.post(`/games/${gameID}/decks`, {
      data: { player_id: playerID, source, format: "text" },
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok()) {
      return (await res.json()) as { deck_name: string; card_count: number; commanders: string[] };
    }
    const body = await res.text();
    const indexLoading = res.status() === 503 && body.includes("card index not loaded");
    if (!indexLoading || Date.now() > deadline) {
      throw new Error(`upload deck failed: ${res.status()} ${body}`);
    }
    await new Promise((r) => setTimeout(r, 2000));
  }
}

export async function startGameAs(
  req: APIRequestContext,
  token: string,
  gameID: string,
): Promise<GameMeta> {
  const res = await req.post(`/games/${gameID}/start`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok()) throw new Error(`start game failed: ${res.status()} ${await res.text()}`);
  return (await res.json()) as GameMeta;
}

export async function getGameAs(
  req: APIRequestContext,
  token: string,
  gameID: string,
): Promise<GameMeta> {
  const res = await req.get(`/games/${gameID}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok()) throw new Error(`get game failed: ${res.status()} ${await res.text()}`);
  return (await res.json()) as GameMeta;
}

// joinViaAPI claims a seat without a browser. Used to stage a table
// into a particular shape (partly seated, full) before a test loads
// the invite page — walking N browser contexts through the UI just to
// fill seats is slow and tests nothing the join spec doesn't already.
export async function joinViaAPI(
  req: APIRequestContext,
  gameID: string,
  inviteToken: string,
  name: string,
): Promise<{ token: string; player_id: string }> {
  const res = await req.post(`/games/${gameID}/join`, {
    data: { invite_token: inviteToken, name },
  });
  if (!res.ok()) throw new Error(`join failed: ${res.status()} ${await res.text()}`);
  return (await res.json()) as { token: string; player_id: string };
}
