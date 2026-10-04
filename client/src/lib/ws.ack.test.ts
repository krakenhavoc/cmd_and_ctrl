import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { get } from "svelte/store";

import { GameClient } from "./ws";
import { PROTOCOL_VERSION } from "./protocol";

// ADR 0122 §6.4: the server acknowledges every applied action to the
// client that sent it with an `ack` frame. The browser does nothing
// with it but record it — and must not log it as an unknown kind, or
// every action a player takes would put an error in the protocol log
// that bug reports attach.

class FakeSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
  static last: FakeSocket | null = null;

  readyState = FakeSocket.OPEN;
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};

  constructor(readonly url: string) {
    FakeSocket.last = this;
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

let realWebSocket: unknown;

beforeEach(() => {
  realWebSocket = (globalThis as Record<string, unknown>).WebSocket;
  (globalThis as Record<string, unknown>).WebSocket = FakeSocket;
  FakeSocket.last = null;
});

afterEach(() => {
  (globalThis as Record<string, unknown>).WebSocket = realWebSocket;
});

describe("ack frames (ADR 0122 §6.4)", () => {
  it("records an ack without logging a protocol error", () => {
    const client = new GameClient("ws://test/ws");
    client.connect();
    const socket = FakeSocket.last!;
    socket.emit("open", {});

    socket.emit("message", {
      data: JSON.stringify({
        v: PROTOCOL_VERSION,
        kind: "ack",
        id: "a4f7b8e1-2c3d-4e5f-9a2c-0d8e7f6a5b4c",
        payload: { seq: 43, generation: 0 },
      }),
    });

    const log = get(client.log);
    expect(log.some((e) => e.direction === "error")).toBe(false);
    expect(log.some((e) => /^ack id=a4f7b8e1 seq=43 generation=0$/.test(e.text))).toBe(true);
    // An ack applies nothing: it follows the snapshot it names.
    expect(get(client.snapshot)).toBeNull();
  });

  it("still logs a kind it does not know", () => {
    const client = new GameClient("ws://test/ws");
    client.connect();
    const socket = FakeSocket.last!;
    socket.emit("open", {});
    socket.emit("message", {
      data: JSON.stringify({ v: PROTOCOL_VERSION, kind: "no_such_kind", id: "x" }),
    });
    expect(
      get(client.log).some((e) => e.direction === "error" && /unknown kind/.test(e.text)),
    ).toBe(true);
  });
});
