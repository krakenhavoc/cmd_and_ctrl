import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

import { afterAdminModeChanged, resetAdminChecksForTest } from "./admin";
import { CLOSE_ADMIN_MODE_CHANGED, GameClient } from "./ws";
import { session, type Session } from "./session";

// ADR 0112 §2 item 5 on the client: a 4001 "admin mode changed" close.
//
// The server closes a person's sockets with 4001 when their admin mode
// is switched or lapses, because the socket's admin bit is now wrong.
// Before ADR 0112 PR 4 the client treated it like any non-1000 close
// and redialled the same binding. When that binding is one only an
// admin may hold (the seatless unfiltered view of a table that is not
// yours, another seat), the server refuses the upgrade, the browser
// sees a bare 1006, and the ladder redials it every 30 seconds forever.
//
// The fix: ask GET /me first, then reconnect with the binding this
// session may now hold, or stop and go to the Lobby. These tests drive
// GameClient with lib/admin.ts's real handler against a fake server
// that refuses every admin-only upgrade from a person in player mode,
// and count the dials.

const TABLE = "11111111-2222-3333-4444-555555555555";
const OTHER_TABLE = "99999999-8888-7777-6666-555555555555";
const USER = "cccccccc-cccc-cccc-cccc-cccccccccccc";
const SEAT = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
const HOUR = 3_600_000;
// Far past the ladder's 30-second cap, many times over.
const HALF_AN_HOUR = 30 * 60_000;

class FakeSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
  static opened: FakeSocket[] = [];
  // server decides each upgrade the way WSAuthorizer would.
  static server: (url: URL) => "accept" | "refuse" = () => "accept";

  readyState = FakeSocket.CONNECTING;
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};

  constructor(readonly url: string) {
    FakeSocket.opened.push(this);
    // The upgrade's answer arrives after the constructor returns, as a
    // browser's does.
    queueMicrotask(() => {
      if (this.readyState === FakeSocket.CLOSED) return;
      if (FakeSocket.server(new URL(url)) === "accept") {
        this.readyState = FakeSocket.OPEN;
        this.emit("open", {});
      } else {
        // A refused upgrade: the browser reports only an abnormal close.
        this.readyState = FakeSocket.CLOSED;
        this.emit("close", { code: 1006, reason: "" });
      }
    });
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

// sessionFor is a person's session with what /me last said about them.
function sessionFor(
  base: { role: Session["principal"]["role"]; gameID?: string; playerID?: string },
  mode: "admin" | "player",
): Session {
  const expiresAt = new Date(Date.now() + 24 * HOUR).toISOString();
  return {
    token: "person-token",
    expiresAt,
    principal: {
      role: base.role,
      user_id: USER,
      game_id: base.gameID,
      player_id: base.playerID,
      issued_at: new Date().toISOString(),
      expires_at: expiresAt,
    },
    gameID: base.gameID,
    playerID: base.playerID,
    admin: mode === "admin",
    admin_allowed: true,
    admin_mode: mode === "admin",
    admin_mode_ends_at: mode === "admin" ? Date.now() + 12 * HOUR : undefined,
  };
}

// meSays is the server's current answer to GET /me, which the tests
// flip when the person switches. Every call is counted.
let meSays: "admin" | "player" | "offline" = "player";
let meCalls = 0;

function stubMe(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string) => {
      if (input !== "/me") throw new Error(`unexpected fetch ${input}`);
      meCalls += 1;
      if (meSays === "offline") throw new TypeError("Failed to fetch");
      const on = meSays === "admin";
      return new Response(
        JSON.stringify({
          admin: on,
          admin_allowed: true,
          admin_mode: on,
          admin_mode_ends_at: on ? Date.now() + 12 * HOUR : undefined,
        }),
        { status: 200 },
      );
    }),
  );
}

// adminOnlyRefused is WSAuthorizer for a person: in admin mode any
// binding; in player mode only a player session's own table.
function adminOnlyRefused(s: () => Session | null) {
  return (url: URL): "accept" | "refuse" => {
    if (meSays === "admin") return "accept";
    const cur = s();
    const game = url.searchParams.get("game");
    if (cur?.principal.role === "player" && cur.gameID === game) return "accept";
    return "refuse";
  };
}

// dialTo connects a client to `table`, with the 4001 handler Game.svelte
// installs, and lets the first upgrade answer.
async function dialTo(table: string, url: string): Promise<{ client: GameClient; left: string[] }> {
  const client = new GameClient(url);
  const left: string[] = [];
  client.setAdminModeHandler(async () => {
    const verdict = await afterAdminModeChanged(table);
    if (verdict === "leave") left.push(table);
    return verdict;
  });
  client.connect();
  await vi.advanceTimersByTimeAsync(0);
  return { client, left };
}

function live(): FakeSocket {
  const s = FakeSocket.opened[FakeSocket.opened.length - 1];
  if (!s) throw new Error("nothing dialled");
  return s;
}

let realWebSocket: unknown;

beforeEach(() => {
  realWebSocket = (globalThis as Record<string, unknown>).WebSocket;
  (globalThis as Record<string, unknown>).WebSocket = FakeSocket;
  FakeSocket.opened = [];
  FakeSocket.server = adminOnlyRefused(() => get(session));
  meCalls = 0;
  resetAdminChecksForTest();
  session.set(null);
  vi.useFakeTimers();
  stubMe();
});

afterEach(() => {
  session.set(null);
  vi.useRealTimers();
  (globalThis as Record<string, unknown>).WebSocket = realWebSocket;
  vi.unstubAllGlobals();
});

describe("a 4001 at a table that needs admin mode", () => {
  it("asks /me, stops, and never redials — no loop", async () => {
    // An admin in admin mode watching another pod's table, seatless.
    meSays = "admin";
    session.set(sessionFor({ role: "player", gameID: OTHER_TABLE, playerID: SEAT }, "admin"));
    const { client, left } = await dialTo(
      TABLE,
      `wss://cmd.example/ws?game=${TABLE}&token=person-token`,
    );
    expect(get(client.status)).toBe("connected");
    expect(FakeSocket.opened).toHaveLength(1);

    // They switch to player mode; the server closes the socket.
    meSays = "player";
    live().emit("close", { code: CLOSE_ADMIN_MODE_CHANGED, reason: "admin mode changed" });
    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);

    expect(FakeSocket.opened).toHaveLength(1);
    expect(meCalls).toBe(1);
    expect(get(client.status)).toBe("disconnected");
    expect(left).toEqual([TABLE]);
    expect(get(session)?.admin_mode).toBe(false);
  });

  it("is the loop it replaces: the same close with no handler redials a refused binding forever", async () => {
    meSays = "admin";
    session.set(sessionFor({ role: "player", gameID: OTHER_TABLE, playerID: SEAT }, "admin"));
    const client = new GameClient(`wss://cmd.example/ws?game=${TABLE}&token=person-token`);
    client.connect();
    await vi.advanceTimersByTimeAsync(0);
    meSays = "player";
    live().emit("close", { code: CLOSE_ADMIN_MODE_CHANGED, reason: "admin mode changed" });
    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);
    // Dozens of refused dials, and still going.
    expect(FakeSocket.opened.length).toBeGreaterThan(20);
    expect(get(client.status)).toBe("reconnecting");
    client.disconnect();
  });
});

describe("a 4001 at a table the session may hold", () => {
  it("reconnects at once, exactly once, to its own seat", async () => {
    meSays = "admin";
    session.set(sessionFor({ role: "player", gameID: TABLE, playerID: SEAT }, "admin"));
    const { client, left } = await dialTo(
      TABLE,
      `wss://cmd.example/ws?game=${TABLE}&token=person-token&player=${SEAT}`,
    );
    meSays = "player";
    live().emit("close", { code: CLOSE_ADMIN_MODE_CHANGED, reason: "admin mode changed" });
    await vi.advanceTimersByTimeAsync(0);
    expect(FakeSocket.opened).toHaveLength(2);
    expect(get(client.status)).toBe("connected");
    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);
    expect(FakeSocket.opened).toHaveLength(2);
    expect(meCalls).toBe(1);
    expect(left).toEqual([]);
    client.disconnect();
  });

  it("switching admin mode on reconnects anywhere", async () => {
    meSays = "player";
    session.set(sessionFor({ role: "player", gameID: TABLE, playerID: SEAT }, "player"));
    const { client } = await dialTo(
      TABLE,
      `wss://cmd.example/ws?game=${TABLE}&token=person-token&player=${SEAT}`,
    );
    meSays = "admin";
    live().emit("close", { code: CLOSE_ADMIN_MODE_CHANGED, reason: "admin mode changed" });
    await vi.advanceTimersByTimeAsync(0);
    expect(FakeSocket.opened).toHaveLength(2);
    expect(get(client.status)).toBe("connected");
    expect(get(session)?.admin_mode).toBe(true);
    client.disconnect();
  });
});

describe("a 4001 while /me cannot answer", () => {
  it("never dials blind: it waits out the ladder, asking /me each rung", async () => {
    meSays = "admin";
    session.set(sessionFor({ role: "player", gameID: OTHER_TABLE, playerID: SEAT }, "admin"));
    const { client, left } = await dialTo(
      TABLE,
      `wss://cmd.example/ws?game=${TABLE}&token=person-token`,
    );
    meSays = "offline";
    live().emit("close", { code: CLOSE_ADMIN_MODE_CHANGED, reason: "admin mode changed" });
    await vi.advanceTimersByTimeAsync(5 * 60_000);
    expect(FakeSocket.opened).toHaveLength(1);
    expect(meCalls).toBeGreaterThan(1);
    expect(get(client.status)).toBe("reconnecting");

    // /me answers again: player mode, so the table is left.
    meSays = "player";
    await vi.advanceTimersByTimeAsync(60_000);
    expect(FakeSocket.opened).toHaveLength(1);
    expect(left).toEqual([TABLE]);
    expect(get(client.status)).toBe("disconnected");
  });

  it("drops a verdict that arrives after the client was torn down", async () => {
    meSays = "admin";
    session.set(sessionFor({ role: "player", gameID: TABLE, playerID: SEAT }, "admin"));
    const { client } = await dialTo(
      TABLE,
      `wss://cmd.example/ws?game=${TABLE}&token=person-token&player=${SEAT}`,
    );
    live().emit("close", { code: CLOSE_ADMIN_MODE_CHANGED, reason: "admin mode changed" });
    client.disconnect();
    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);
    expect(FakeSocket.opened).toHaveLength(1);
    expect(get(client.status)).toBe("disconnected");
  });
});
