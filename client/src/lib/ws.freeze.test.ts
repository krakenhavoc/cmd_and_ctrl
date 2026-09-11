import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { get, writable } from "svelte/store";

import { GameClient } from "./ws";
import { PROTOCOL_VERSION, type GameView } from "./protocol";
import { recentClientErrors, resetClientErrors } from "./clientErrors";

// Regression cover for #266 "[in-app] State freeze": the board stops
// updating while the socket stays green and the server keeps sending.
//
// The amplifier is svelte/store's module-global `subscriber_queue`.
// `set()` drains that queue in a bare loop and only clears it when the
// loop completes; a subscriber that throws skips the reset, leaving the
// queue permanently non-empty. Every later `set()` on EVERY store in
// the app then sees a non-empty queue, enqueues, and returns without
// flushing. One throw anywhere silently freezes the whole UI for the
// life of the page — exactly the reported symptom.
//
// GameClient must not be able to hand svelte/store a throwing
// subscriber, and handleMessage must survive one.

class FakeSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
  static last: FakeSocket | null = null;

  readyState = FakeSocket.OPEN;
  sent: string[] = [];
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};

  constructor(readonly url: string) {
    FakeSocket.last = this;
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(): void {}
  send(data: string): void {
    this.sent.push(data);
  }
  close(): void {
    this.readyState = FakeSocket.CLOSED;
  }
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
}

const viewAt = (step: string): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 2, active_seat: 0, priority_holder: 0, phase: "combat", step },
    mulligans_open: false,
  }) as unknown as GameView;

const snapshotFrame = (seq: number, step = "declare_attackers"): string =>
  JSON.stringify({
    v: PROTOCOL_VERSION,
    kind: "snapshot",
    id: "frame-" + seq,
    payload: { seq, game: viewAt(step) },
  });

let realWebSocket: unknown;

beforeEach(() => {
  realWebSocket = (globalThis as Record<string, unknown>).WebSocket;
  (globalThis as Record<string, unknown>).WebSocket = FakeSocket;
  FakeSocket.last = null;
  resetClientErrors();
});

afterEach(() => {
  (globalThis as Record<string, unknown>).WebSocket = realWebSocket;
});

function connected(): { client: GameClient; socket: FakeSocket } {
  const client = new GameClient("ws://test/ws");
  client.connect();
  const socket = FakeSocket.last!;
  socket.emit("open", {});
  return { client, socket };
}

describe("#266 — one throwing subscriber must not freeze the client", () => {
  it("keeps delivering snapshots to healthy subscribers after one throws", () => {
    const { client, socket } = connected();

    const healthy: number[] = [];
    client.snapshot.subscribe(() => {
      // A component whose render throws on a particular state.
      throw new Error("render blew up");
    });
    client.snapshot.subscribe((v) => {
      if (v) healthy.push(v.turn.number);
    });

    socket.emit("message", { data: snapshotFrame(77) });
    socket.emit("message", { data: snapshotFrame(78) });

    // The healthy subscriber must still be fed.
    expect(healthy.length).toBe(2);
    // And the failure must be visible to a bug report rather than silent.
    expect(recentClientErrors().some((e) => /render blew up/.test(e.text))).toBe(true);
    expect(get(client.lastSeq)).toBe(78);
  });

  it("does not poison unrelated stores elsewhere in the app", () => {
    const { client, socket } = connected();

    client.snapshot.subscribe(() => {
      throw new Error("render blew up");
    });

    // A store that belongs to some other module entirely — settings,
    // targeting, manual stops. It must keep notifying.
    const other = writable(0);
    const seen: number[] = [];
    other.subscribe((v) => seen.push(v));

    socket.emit("message", { data: snapshotFrame(77) });

    seen.length = 0;
    other.set(42);
    expect(seen).toEqual([42]);
  });

  it("logs the frame it received even when applying it throws", () => {
    const { client, socket } = connected();

    client.snapshot.subscribe((v) => {
      if (v) throw new Error("render blew up");
    });

    socket.emit("message", { data: snapshotFrame(77) });

    const texts = get(client.log).map((e) => e.text);
    expect(
      texts.some((t) => /snapshot seq=77/.test(t)),
      `log was: ${texts.join(" | ")}`,
    ).toBe(true);
  });

  it("keeps the socket usable so the player can still act", () => {
    const { client, socket } = connected();

    client.snapshot.subscribe(() => {
      throw new Error("render blew up");
    });
    socket.emit("message", { data: snapshotFrame(77) });

    expect(client.sendAction("pass_priority")).not.toBeNull();
    expect(socket.sent.some((s) => s.includes("pass_priority"))).toBe(true);
  });
});
