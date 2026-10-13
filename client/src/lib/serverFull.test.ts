import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

import { connectionBanner } from "./connectionBanner";
import { LobbyApiError, session } from "./session";
import {
  CLOSE_TRY_AGAIN_LATER,
  SERVER_FULL_CODE,
  SERVER_FULL_MESSAGE,
  isServerFull,
  lobbyErrorMessage,
  serverFull,
} from "./serverFull";
import { GameClient } from "./ws";
import { PROTOCOL_VERSION } from "./protocol";

// ADR 0148 (#2977): the server answers a request it will not take
// because it is full with a stable code, and the client words the
// refusal for a person. These tests pin the code the client matches,
// the sentence it shows on create / join, and the socket path.

const fullError = (status = 503) =>
  new LobbyApiError(status, "anything the server wrote", undefined, undefined, {
    error: "anything the server wrote",
    code: SERVER_FULL_CODE,
  });

describe("isServerFull / lobbyErrorMessage", () => {
  it("matches on the code, not on the status or the sentence", () => {
    expect(isServerFull(fullError())).toBe(true);
    // A proxy might change the status; the code is the contract.
    expect(isServerFull(fullError(502))).toBe(true);
    // The same sentence without the code is some other error.
    expect(
      isServerFull(new LobbyApiError(503, SERVER_FULL_MESSAGE, undefined, undefined, {})),
    ).toBe(false);
    expect(isServerFull(new LobbyApiError(503, "x"))).toBe(false);
    expect(isServerFull(new Error("boom"))).toBe(false);
    expect(isServerFull(null)).toBe(false);
  });

  it("shows the friendly sentence for a capacity refusal, whatever the body said", () => {
    expect(lobbyErrorMessage(fullError(), "create failed")).toBe(SERVER_FULL_MESSAGE);
  });

  it("keeps the server's own message for any other lobby error, and the fallback otherwise", () => {
    expect(lobbyErrorMessage(new LobbyApiError(409, "you already have 3 open tables"), "x")).toBe(
      "you already have 3 open tables",
    );
    expect(lobbyErrorMessage(new Error("network"), "create failed")).toBe("create failed");
  });

  it("is a sentence a player can act on, and not a raw status", () => {
    expect(SERVER_FULL_MESSAGE).toMatch(/full right now/i);
    expect(SERVER_FULL_MESSAGE).toMatch(/try again/i);
    expect(SERVER_FULL_MESSAGE).not.toMatch(/\b50\d\b/);
  });
});

describe("the connection banner", () => {
  it("says the server is full, and keeps retrying, when the socket was refused at the cap", () => {
    const b = connectionBanner("reconnecting", 2, false, true);
    expect(b?.headline).toBe(SERVER_FULL_MESSAGE);
    expect(b?.tone).toBe("retrying");
    expect(b?.retryLabel).toBe("Try now");
  });

  it("is the ordinary dropped-connection banner otherwise", () => {
    expect(connectionBanner("reconnecting", 2, false, false)?.headline).toMatch(/connection lost/i);
    expect(connectionBanner("reconnecting", 2)?.headline).toMatch(/connection lost/i);
  });
});

// A socket refused at the connection cap: the server accepts the
// upgrade, writes one error frame with the code, then closes 1013.
class FakeSocket {
  static readonly OPEN = 1;
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
  close(): void {}
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
}

describe("a socket refused at the connection cap", () => {
  let realWebSocket: unknown;
  beforeEach(() => {
    realWebSocket = (globalThis as Record<string, unknown>).WebSocket;
    (globalThis as Record<string, unknown>).WebSocket = FakeSocket;
    FakeSocket.opened = [];
    session.set(null);
    serverFull.set(false);
    vi.useFakeTimers();
  });
  afterEach(() => {
    session.set(null);
    vi.useRealTimers();
    (globalThis as Record<string, unknown>).WebSocket = realWebSocket;
  });

  it("raises serverFull on the error frame and retries after the 1013 close", () => {
    const client = new GameClient("wss://cmd.example/ws?game=g&token=t");
    client.connect();
    const socket = FakeSocket.opened[0];
    socket.emit("open", {});
    socket.emit("message", {
      data: JSON.stringify({
        v: PROTOCOL_VERSION,
        kind: "error",
        payload: { code: SERVER_FULL_CODE, message: SERVER_FULL_MESSAGE },
      }),
    });
    expect(get(serverFull)).toBe(true);

    socket.emit("close", { code: CLOSE_TRY_AGAIN_LATER, reason: SERVER_FULL_MESSAGE });
    // 1013 is not terminal: the client is on its reconnect ladder.
    expect(get(client.status)).toBe("reconnecting");
    vi.advanceTimersByTime(60_000);
    expect(FakeSocket.opened.length).toBe(2);
  });

  it("clears serverFull when a snapshot arrives (the cap let us in)", () => {
    const client = new GameClient("wss://cmd.example/ws?game=g&token=t");
    client.connect();
    const socket = FakeSocket.opened[0];
    socket.emit("open", {});
    serverFull.set(true);
    socket.emit("message", {
      data: JSON.stringify({
        v: PROTOCOL_VERSION,
        kind: "snapshot",
        payload: { seq: 1, generation: 1, view: { id: "g", state: "lobby", players: [] } },
      }),
    });
    expect(get(serverFull)).toBe(false);
  });
});
