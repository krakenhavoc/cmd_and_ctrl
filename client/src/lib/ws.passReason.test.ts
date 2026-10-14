import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { get } from "svelte/store";

import { GameClient } from "./ws";
import { PROTOCOL_VERSION } from "./protocol";
import { collectBugLog } from "./bugReport";

// #3343: an automatic pass carries its reason into the client log the
// bug report attaches, the reason never goes on the wire, and the
// client keeps the ack and the action count the auto-pass chip reads.

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

let realWebSocket: unknown;

beforeEach(() => {
  realWebSocket = (globalThis as Record<string, unknown>).WebSocket;
  (globalThis as Record<string, unknown>).WebSocket = FakeSocket;
  FakeSocket.last = null;
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

describe("sendAction's reason (#3343)", () => {
  it("writes the reason into the sent line, and not onto the wire", () => {
    const { client, socket } = connected();
    const id = client.sendAction("pass_priority", undefined, undefined, "nothing to respond with");
    expect(id).not.toBeNull();

    const line = get(client.log).find((e) => e.direction === "sent")!;
    expect(line.text).toBe(
      `action pass_priority id=${id!.slice(0, 8)} reason="nothing to respond with"`,
    );
    // Not part of the protocol: the frame is unchanged.
    expect(socket.sent[0]).not.toContain("nothing to respond with");
    expect(JSON.parse(socket.sent[0]).payload).toEqual({ type: "pass_priority" });
  });

  it("leaves the line as it was when there is no reason", () => {
    const { client } = connected();
    const id = client.sendAction("pass_priority");
    const line = get(client.log).find((e) => e.direction === "sent")!;
    expect(line.text).toBe(`action pass_priority id=${id!.slice(0, 8)}`);
  });

  it("reaches the log a bug report attaches", () => {
    const { client } = connected();
    client.sendAction("pass_priority", undefined, undefined, "skipping to your turn");
    const bug = collectBugLog(get(client.log), []);
    expect(
      bug.some((e) => e.kind === "sent" && /reason="skipping to your turn"/.test(e.text)),
    ).toBe(true);
  });
});

describe("the stores the auto-pass chip reads (#3343)", () => {
  it("keeps the newest ack by its frame id", () => {
    const { client, socket } = connected();
    expect(get(client.lastAck)).toBeNull();
    socket.emit("message", {
      data: JSON.stringify({
        v: PROTOCOL_VERSION,
        kind: "ack",
        id: "a4f7b8e1-2c3d-4e5f-9a2c-0d8e7f6a5b4c",
        payload: { seq: 43, generation: 0 },
      }),
    });
    expect(get(client.lastAck)).toEqual({ id: "a4f7b8e1-2c3d-4e5f-9a2c-0d8e7f6a5b4c", seq: 43 });
  });

  it("counts actions as a store, in step with actionsSent", () => {
    const { client } = connected();
    expect(get(client.actionCount)).toBe(0);
    client.sendAction("pass_priority");
    client.sendAction("undo");
    expect(get(client.actionCount)).toBe(2);
    expect(client.actionsSent).toBe(2);
  });
});
