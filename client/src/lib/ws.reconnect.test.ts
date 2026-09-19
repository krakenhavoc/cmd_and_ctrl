import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

import {
  CLOSE_GOING_AWAY,
  CLOSE_NORMAL,
  GameClient,
  isTerminalClose,
  withSessionToken,
} from "./ws";
import { session, type Session } from "./session";

// Cover for #518: a deploy must not disconnect the table permanently.
//
// `hub.Shutdown` writes close 1001 "server shutting down" to every
// client on `systemctl restart`, and ADR 0041 persists and restores the
// game across that restart — but the close handler treated 1001 and
// 1000 alike as terminal, so the client went straight to `disconnected`
// and never dialled again. The exponential-backoff ladder was dead code
// on the one path it was written for, and ws.test.ts only ever exercised
// that ladder's arithmetic, which is why the regression was invisible.
//
// 1000 must STAY terminal: hub.EvictGame sends it when the game is
// deleted, and a session revocation sends it when the credential is
// gone (ADR 0044 decisions 2 and 3). Redialling either is dialling for
// something that is not coming back.

const GAME_ID = "11111111-2222-3333-4444-555555555555";
const STALE_TOKEN = "stale-token-from-the-previous-process";
const FRESH_TOKEN = "fresh-token-minted-after-the-deploy";
const WS_URL = `wss://cmd.example/ws?game=${GAME_ID}&token=${STALE_TOKEN}&player=p-1`;

// CLOSE_ABNORMAL is what a browser reports for a connection that died
// without a close frame — a network blip, a crash, or a rejected
// upgrade. It always retried and must keep retrying.
const CLOSE_ABNORMAL = 1006;

class FakeSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
  // Every socket this test file's WebSocket stand-in has constructed,
  // in dial order. The length IS the "did it redial?" assertion, and
  // the last entry's url is what the redial actually asked for.
  static opened: FakeSocket[] = [];

  readyState = FakeSocket.OPEN;
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};

  constructor(readonly url: string) {
    FakeSocket.opened.push(this);
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(): void {}
  send(): void {}
  close(): void {
    this.readyState = FakeSocket.CLOSED;
  }
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
}

function lastSocket(): FakeSocket {
  const s = FakeSocket.opened[FakeSocket.opened.length - 1];
  if (!s) throw new Error("no socket was dialled");
  return s;
}

// closeFrom delivers a server close to the live socket, the way the
// browser would: code plus the reason string the server wrote.
function closeFrom(socket: FakeSocket, code: number, reason = ""): void {
  socket.emit("close", { code, reason });
}

// connected dials and completes the handshake, leaving the client in
// the state a mid-game player is in when a deploy lands.
function connected(url = WS_URL): { client: GameClient; socket: FakeSocket } {
  const client = new GameClient(url);
  client.connect();
  const socket = lastSocket();
  socket.emit("open", {});
  return { client, socket };
}

// LADDER_MS is comfortably past the first rung of the backoff ladder
// (500ms nominal, half-jittered), so advancing by it always fires a
// scheduled reconnect and never silently proves nothing.
const LADDER_MS = 60_000;

// playerSession is a live session, expiring far enough out that
// advancing the fake clock through the reconnect ladder cannot trip
// session.ts's expiry timer and clear it mid-test.
function playerSession(token: string): Session {
  const expiresAt = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token,
    expiresAt,
    principal: {
      role: "player",
      game_id: GAME_ID,
      player_id: "p-1",
      issued_at: new Date().toISOString(),
      expires_at: expiresAt,
    },
    playerID: "p-1",
    gameID: GAME_ID,
  };
}

let realWebSocket: unknown;

beforeEach(() => {
  realWebSocket = (globalThis as Record<string, unknown>).WebSocket;
  (globalThis as Record<string, unknown>).WebSocket = FakeSocket;
  FakeSocket.opened = [];
  session.set(null);
  vi.useFakeTimers();
});

afterEach(() => {
  // Cleared before the clock is handed back, so session.ts's expiry
  // timer is cancelled on the same timer implementation that armed it.
  session.set(null);
  vi.useRealTimers();
  (globalThis as Record<string, unknown>).WebSocket = realWebSocket;
});

describe("isTerminalClose", () => {
  it("is true only for 1000 — the game is gone", () => {
    expect(isTerminalClose(CLOSE_NORMAL)).toBe(true);
  });

  it("is false for 1001 — the process is restarting, the game is not gone", () => {
    expect(isTerminalClose(CLOSE_GOING_AWAY)).toBe(false);
  });

  it("is false for every abnormal close", () => {
    for (const code of [CLOSE_ABNORMAL, 1005, 1011, 1012, 1013, 4000]) {
      expect(isTerminalClose(code)).toBe(false);
    }
  });
});

describe("#518 — close 1001 from a deploy schedules a reconnect", () => {
  it("goes to reconnecting and redials after the backoff", () => {
    const { client, socket } = connected();
    closeFrom(socket, CLOSE_GOING_AWAY, "server shutting down");

    expect(get(client.status)).toBe("reconnecting");
    expect(FakeSocket.opened).toHaveLength(1);

    vi.advanceTimersByTime(LADDER_MS);
    expect(FakeSocket.opened).toHaveLength(2);
    expect(lastSocket()).not.toBe(socket);
  });

  it("comes back connected when the restarted server answers", () => {
    const { client, socket } = connected();
    closeFrom(socket, CLOSE_GOING_AWAY, "server shutting down");
    vi.advanceTimersByTime(LADDER_MS);

    lastSocket().emit("open", {});
    expect(get(client.status)).toBe("connected");
    expect(get(client.reconnectAttempt)).toBe(0);
  });

  it("climbs the ladder while the server is still down", () => {
    const { client, socket } = connected();
    closeFrom(socket, CLOSE_GOING_AWAY, "server shutting down");
    expect(get(client.reconnectAttempt)).toBe(1);

    for (let i = 2; i <= 4; i++) {
      vi.advanceTimersByTime(LADDER_MS);
      closeFrom(lastSocket(), CLOSE_ABNORMAL);
      expect(get(client.reconnectAttempt)).toBe(i);
    }
    expect(get(client.status)).toBe("reconnecting");
    expect(FakeSocket.opened).toHaveLength(4);
  });

  it("names the close code in the protocol log", () => {
    const { client, socket } = connected();
    closeFrom(socket, CLOSE_GOING_AWAY, "server shutting down");

    const texts = get(client.log).map((e) => e.text);
    expect(
      texts.some((t) => t.includes("socket closed (1001")),
      `log was: ${texts.join(" | ")}`,
    ).toBe(true);
  });
});

describe("#518 — close 1000 stays terminal", () => {
  it("goes to disconnected and never redials", () => {
    const { client, socket } = connected();
    closeFrom(socket, CLOSE_NORMAL, "game deleted");

    expect(get(client.status)).toBe("disconnected");
    expect(get(client.reconnectAttempt)).toBe(0);

    vi.advanceTimersByTime(LADDER_MS);
    expect(FakeSocket.opened).toHaveLength(1);
    expect(get(client.status)).toBe("disconnected");
  });

  it("is terminal for a revoked session too (ADR 0044 decision 3)", () => {
    const { client, socket } = connected();
    closeFrom(socket, CLOSE_NORMAL, "session revoked");

    vi.advanceTimersByTime(LADDER_MS);
    expect(FakeSocket.opened).toHaveLength(1);
    expect(get(client.status)).toBe("disconnected");
  });
});

describe("#518 — an abnormal close behaves as it always did", () => {
  it("retries after 1006", () => {
    const { client, socket } = connected();
    closeFrom(socket, CLOSE_ABNORMAL);

    expect(get(client.status)).toBe("reconnecting");
    vi.advanceTimersByTime(LADDER_MS);
    expect(FakeSocket.opened).toHaveLength(2);
  });

  it("a deliberate disconnect() still retries nothing", () => {
    const { client } = connected();
    client.disconnect();

    expect(get(client.status)).toBe("disconnected");
    vi.advanceTimersByTime(LADDER_MS);
    expect(FakeSocket.opened).toHaveLength(1);
  });
});

describe("withSessionToken", () => {
  it("replaces a stale token with the one held now", () => {
    const url = withSessionToken(WS_URL, FRESH_TOKEN);
    expect(new URL(url).searchParams.get("token")).toBe(FRESH_TOKEN);
    expect(url).not.toContain(STALE_TOKEN);
  });

  it("keeps every other parameter, including the seat", () => {
    const params = new URL(withSessionToken(WS_URL, FRESH_TOKEN)).searchParams;
    expect(params.get("game")).toBe(GAME_ID);
    expect(params.get("player")).toBe("p-1");
  });

  it("returns the URL verbatim when the token already matches", () => {
    expect(withSessionToken(WS_URL, STALE_TOKEN)).toBe(WS_URL);
  });

  it("leaves a tokenless URL alone when there is no session", () => {
    const spectator = `wss://cmd.example/ws?game=${GAME_ID}`;
    expect(withSessionToken(spectator, null)).toBe(spectator);
    expect(withSessionToken(spectator, undefined)).toBe(spectator);
    expect(withSessionToken(spectator, "")).toBe(spectator);
  });

  it("adds the token to a URL that was built without one", () => {
    const url = withSessionToken(`wss://cmd.example/ws?game=${GAME_ID}`, FRESH_TOKEN);
    expect(new URL(url).searchParams.get("token")).toBe(FRESH_TOKEN);
  });

  it("returns an unparseable URL untouched rather than mangling it", () => {
    expect(withSessionToken("not a url", FRESH_TOKEN)).toBe("not a url");
  });
});

describe("#518 — the dial reads the session store, not the mount-time URL", () => {
  it("redials with a token refreshed after the URL was captured", () => {
    session.set(playerSession(STALE_TOKEN));
    const { socket } = connected();
    expect(socket.url).toBe(WS_URL);

    // The deploy lands, and the session is re-minted somewhere else —
    // another tab, a refresh, #517's durable credential. The captured
    // URL still carries the old one.
    session.set(playerSession(FRESH_TOKEN));
    closeFrom(socket, CLOSE_GOING_AWAY, "server shutting down");
    vi.advanceTimersByTime(LADDER_MS);

    expect(FakeSocket.opened).toHaveLength(2);
    expect(new URL(lastSocket().url).searchParams.get("token")).toBe(FRESH_TOKEN);
  });

  it("dials the captured URL unchanged when no session is stored", () => {
    connected();
    expect(lastSocket().url).toBe(WS_URL);
  });
});
