import { afterEach, beforeEach, describe, expect, it, vi, type Mock } from "vitest";
import { get } from "svelte/store";

import { CLOSE_NORMAL, GameClient, type TableClosedCheck } from "./ws";
import { SEAT_REMOVED_MESSAGE } from "./session";

// #5082, ADR 0211 §4 on the client: an archived table refuses every
// non-admin WebSocket upgrade with 409 (#5042). A browser never sees
// that status, only a bare 1006, so without help a seat or spectator
// whose table was closed redials it every 30 seconds forever under
// "Connection lost — reconnecting". The client now asks the server
// whether the table was closed (the Game route installs api.ts's
// tableIsClosed, GET /games/{id}) after a dial that never opened, and
// after the 1000 an archive closes sockets with, and stops on
// "table_closed" when it was.

const TABLE = "11111111-2222-3333-4444-555555555555";
const WS_URL = `wss://cmd.example/ws?game=${TABLE}&token=seat-token&player=p-1`;
// Far past the ladder's 30-second cap, many times over.
const HALF_AN_HOUR = 30 * 60_000;

class FakeSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
  static opened: FakeSocket[] = [];
  // The server's answer to each upgrade, as WSAuthorizer gives it.
  static server: () => "accept" | "refuse" = () => "accept";

  readyState = FakeSocket.CONNECTING;
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};

  constructor(readonly url: string) {
    FakeSocket.opened.push(this);
    // The upgrade's answer arrives after the constructor returns, as a
    // browser's does.
    queueMicrotask(() => {
      if (this.readyState === FakeSocket.CLOSED) return;
      if (FakeSocket.server() === "accept") {
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

function lastSocket(): FakeSocket {
  const s = FakeSocket.opened[FakeSocket.opened.length - 1];
  if (!s) throw new Error("no socket was dialled");
  return s;
}

async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) await Promise.resolve();
}

let realWebSocket: unknown;

beforeEach(() => {
  realWebSocket = (globalThis as Record<string, unknown>).WebSocket;
  (globalThis as Record<string, unknown>).WebSocket = FakeSocket;
  FakeSocket.opened = [];
  FakeSocket.server = () => "accept";
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
  (globalThis as Record<string, unknown>).WebSocket = realWebSocket;
});

// archivedTable is the server once the table was archived: every upgrade
// refused, and GET /games/{id} answering with archived_at.
function archivedTable(): Mock<TableClosedCheck> {
  FakeSocket.server = () => "refuse";
  return vi.fn(async () => true);
}

describe("#5082 — a seat whose table was closed", () => {
  it("stops redialling after the first refused dial and says the table was closed", async () => {
    const check = archivedTable();
    const client = new GameClient(WS_URL);
    client.setTableClosedCheck(check);
    client.connect();
    await settle();

    expect(check).toHaveBeenCalledTimes(1);
    expect(get(client.status)).toBe("table_closed");
    const dials = FakeSocket.opened.length;
    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);
    await settle();
    expect(FakeSocket.opened).toHaveLength(dials);
    expect(get(client.status)).toBe("table_closed");
  });

  it("stops after a live socket drops and the redial is refused", async () => {
    const client = new GameClient(WS_URL);
    const check = vi.fn(async () => true);
    client.setTableClosedCheck(check);
    client.connect();
    await settle();
    expect(get(client.status)).toBe("connected");

    // The table is archived while the player's laptop sleeps; the socket
    // drops without a close frame, and the redial is refused.
    FakeSocket.server = () => "refuse";
    lastSocket().emit("close", { code: 1006, reason: "" });
    // A drop of a socket that was up is an ordinary drop: no question yet.
    expect(check).not.toHaveBeenCalled();
    expect(get(client.status)).toBe("reconnecting");

    await vi.advanceTimersByTimeAsync(30_000);
    await settle();
    expect(check).toHaveBeenCalledTimes(1);
    expect(get(client.status)).toBe("table_closed");
    const dials = FakeSocket.opened.length;
    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);
    await settle();
    expect(FakeSocket.opened).toHaveLength(dials);
  });

  // EvictGame closes an archived table's sockets with 1000, the same
  // close a deleted table gets. Asked which it was, a closed table says
  // so instead of offering a Reconnect the server would refuse.
  it("reads the archive's 1000 as table_closed", async () => {
    const client = new GameClient(WS_URL);
    const check = vi.fn(async () => true);
    client.setTableClosedCheck(check);
    client.connect();
    await settle();

    lastSocket().emit("close", { code: CLOSE_NORMAL, reason: "game deleted" });
    expect(get(client.status)).toBe("disconnected");
    await settle();
    expect(check).toHaveBeenCalledTimes(1);
    expect(get(client.status)).toBe("table_closed");
  });

  it("leaves a deleted table's 1000 on disconnected", async () => {
    const client = new GameClient(WS_URL);
    client.setTableClosedCheck(async () => false);
    client.connect();
    await settle();

    lastSocket().emit("close", { code: CLOSE_NORMAL, reason: "game deleted" });
    await settle();
    expect(get(client.status)).toBe("disconnected");
  });

  it("does not ask about a removed seat, which has its own notice", async () => {
    const client = new GameClient(WS_URL);
    const check = vi.fn(async () => true);
    client.setTableClosedCheck(check);
    client.connect();
    await settle();

    lastSocket().emit("close", { code: CLOSE_NORMAL, reason: SEAT_REMOVED_MESSAGE });
    await settle();
    expect(check).not.toHaveBeenCalled();
    expect(get(client.status)).toBe("disconnected");
  });
});

describe("#5082 — a table that is still open", () => {
  // A deploy, a crash, a blip: the dial fails, the answer is "not
  // closed", and the ladder carries on exactly as before.
  it("keeps the reconnect ladder when the answer is no", async () => {
    FakeSocket.server = () => "refuse";
    const check = vi.fn(async () => false);
    const client = new GameClient(WS_URL);
    client.setTableClosedCheck(check);
    client.connect();
    await settle();
    expect(get(client.status)).toBe("reconnecting");

    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);
    await settle();
    expect(FakeSocket.opened.length).toBeGreaterThan(3);
    expect(get(client.status)).toBe("reconnecting");

    // And comes back when the server does.
    // Only the next rung: running the clock on past the open would let
    // the heartbeat time out on this fake, which never answers a ping.
    FakeSocket.server = () => "accept";
    await vi.advanceTimersToNextTimerAsync();
    await settle();
    expect(get(client.status)).toBe("connected");
  });

  it("keeps the ladder when the question itself fails", async () => {
    FakeSocket.server = () => "refuse";
    const client = new GameClient(WS_URL);
    client.setTableClosedCheck(async () => {
      throw new Error("offline");
    });
    client.connect();
    await settle();
    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);
    await settle();
    expect(get(client.status)).toBe("reconnecting");
  });

  it("asks one question at a time", async () => {
    FakeSocket.server = () => "refuse";
    // Never answers: every later failed dial finds one already out.
    const check = vi.fn(() => new Promise<boolean>(() => {}));
    const client = new GameClient(WS_URL);
    client.setTableClosedCheck(check);
    client.connect();
    await settle();
    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);
    await settle();
    expect(FakeSocket.opened.length).toBeGreaterThan(3);
    expect(check).toHaveBeenCalledTimes(1);
  });

  // A "closed" that comes back after the socket is up again describes a
  // world the client has left (the table was unarchived meanwhile).
  it("drops a verdict that arrives after the socket came back", async () => {
    FakeSocket.server = () => "refuse";
    let answer: (closed: boolean) => void = () => {};
    const client = new GameClient(WS_URL);
    client.setTableClosedCheck(
      () =>
        new Promise<boolean>((resolve) => {
          answer = resolve;
        }),
    );
    client.connect();
    await settle();

    FakeSocket.server = () => "accept";
    await vi.advanceTimersToNextTimerAsync();
    await settle();
    expect(get(client.status)).toBe("connected");

    answer(true);
    await settle();
    expect(get(client.status)).toBe("connected");
  });

  it("with no check installed, a refused dial climbs the ladder as before", async () => {
    FakeSocket.server = () => "refuse";
    const client = new GameClient(WS_URL);
    client.connect();
    await settle();
    await vi.advanceTimersByTimeAsync(HALF_AN_HOUR);
    await settle();
    expect(get(client.status)).toBe("reconnecting");
  });
});
