import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { get } from "svelte/store";

import { GameClient } from "./ws";
import { PROTOCOL_VERSION } from "./protocol";

// #3619: the chat log is rendered by a keyed {#each} (BotFeed), so two
// lines with the same id throw each_key_duplicate and take the board
// down. A chat frame with no id (an AI agent seat's `say`) must still
// get a key of its own.

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

describe("chat frame ids (#3619)", () => {
  it("gives chat frames that carry no id distinct, non-empty ids", () => {
    const client = new GameClient("ws://test/ws");
    client.connect();
    const socket = FakeSocket.last!;
    socket.emit("open", {});

    for (const text of ["GG", "thanks for the rematch"]) {
      socket.emit("message", {
        data: JSON.stringify({
          v: PROTOCOL_VERSION,
          kind: "chat",
          payload: {
            author_id: "a",
            author_name: "Claude",
            text,
            timestamp: "2026-10-14T18:36:12Z",
          },
        }),
      });
    }

    const ids = get(client.chat).map((m) => m.id);
    expect(ids).toHaveLength(2);
    expect(ids.every((id) => id !== "")).toBe(true);
    expect(new Set(ids).size).toBe(2);
  });
});
