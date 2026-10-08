import type { APIRequestContext } from "@playwright/test";
import { ADMIN_TOKEN, SERVER_WS_ORIGIN } from "./env";

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

export interface StartOptions {
  // ADR 0121 §1: Start opens the opening roll — every seat rolls a d20
  // and the winner chooses who goes first — before anything is dealt.
  // By default the helper finishes it as the admin ("Roll for everyone
  // left" until one leader is left, who then takes the first turn), so
  // a spec that starts a table lands in the mulligan window as it
  // always has. A spec that plays the roll through the UI passes false.
  finishRoll?: boolean;
}

export async function startGameAs(
  req: APIRequestContext,
  token: string,
  gameID: string,
  opts: StartOptions = {},
): Promise<GameMeta & { starting_seat?: number }> {
  const res = await req.post(`/games/${gameID}/start`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok()) throw new Error(`start game failed: ${res.status()} ${await res.text()}`);
  const meta = (await res.json()) as GameMeta;
  if (opts.finishRoll === false) return meta;
  const startingSeat = await finishOpeningRollAsAdmin(token, gameID);
  return { ...meta, starting_seat: startingSeat };
}

// The few view fields the opening roll needs.
interface RollView {
  seats: { id: string; seat: number }[];
  starting_seat?: number;
  opening_roll?: {
    rounds: { seats: number[]; rolls: { seat: number; result: number }[] }[];
    chooser?: number;
  };
}

// finishOpeningRollAsAdmin plays an open opening roll out over an admin
// WebSocket with no seat binding: host_roll_remaining until a chooser
// exists, then that chooser's choose_starting_player for themselves (an
// admin may act for any seat). Returns the starting seat. A table with
// no roll open (already finished) returns at once.
export async function finishOpeningRollAsAdmin(
  adminToken: string,
  gameID: string,
): Promise<number | undefined> {
  if (typeof WebSocket === "undefined") {
    throw new Error(
      `global WebSocket is unavailable (Node ${process.version}); tests-e2e requires Node >= 22`,
    );
  }
  const url = `${SERVER_WS_ORIGIN}/ws?game=${gameID}&token=${encodeURIComponent(adminToken)}`;
  const ws = new WebSocket(url);
  let latest: RollView | null = null;
  let failure: string | null = null;
  const sentIDs = new Set<string>();
  ws.addEventListener("message", (ev) => {
    let frame: {
      kind: string;
      id?: string;
      payload?: { game?: RollView; code?: string; message?: string };
    };
    try {
      frame = JSON.parse(String(ev.data));
    } catch {
      return;
    }
    if (frame.kind === "snapshot" && frame.payload?.game) latest = frame.payload.game;
    if (frame.kind === "error" && frame.id && sentIDs.has(frame.id)) {
      failure = `${frame.payload?.code}: ${frame.payload?.message}`;
    }
  });
  await new Promise<void>((resolve, reject) => {
    const t = setTimeout(() => reject(new Error("admin WS open timed out")), 8000);
    ws.addEventListener("open", () => (clearTimeout(t), resolve()), { once: true });
    ws.addEventListener("error", () => (clearTimeout(t), reject(new Error("admin WS error"))), {
      once: true,
    });
  });

  // Wait until the view moves past `before` (or a refusal comes back).
  const waitPast = async (before: string, what: string): Promise<RollView> => {
    const deadline = Date.now() + 10_000;
    for (;;) {
      if (failure) throw new Error(`opening roll: ${what} refused: ${failure}`);
      const v = latest as RollView | null;
      if (v && JSON.stringify(v.opening_roll ?? null) !== before) return v;
      if (Date.now() > deadline) throw new Error(`opening roll: timed out waiting on ${what}`);
      await new Promise((r) => setTimeout(r, 50));
    }
  };
  const send = (type: string, player?: string, params?: Record<string, unknown>) => {
    const id = crypto.randomUUID();
    sentIDs.add(id);
    const payload: Record<string, unknown> = { type };
    if (player) payload.player = player;
    if (params) payload.params = params;
    ws.send(JSON.stringify({ v: 0, kind: "action", id, payload }));
  };

  try {
    let view = await waitPast("<none yet>", "the first snapshot");
    // Each round shrinks to its tied leaders, so this is generous.
    for (let i = 0; i < 32; i++) {
      const roll = view.opening_roll;
      if (!roll) return view.starting_seat;
      const before = JSON.stringify(roll);
      if (roll.chooser !== undefined) {
        const chooser = view.seats.find((s) => s.seat === roll.chooser);
        if (!chooser) throw new Error(`opening roll: no seat ${roll.chooser}`);
        send("choose_starting_player", chooser.id, { seat: roll.chooser });
        view = await waitPast(before, "choose_starting_player");
      } else {
        send("host_roll_remaining");
        view = await waitPast(before, "host_roll_remaining");
      }
    }
    throw new Error("opening roll: did not finish");
  } finally {
    ws.close();
  }
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
