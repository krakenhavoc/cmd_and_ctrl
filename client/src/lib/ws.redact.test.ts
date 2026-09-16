import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { get } from "svelte/store";

import { GameClient, connectLogLine } from "./ws";
import { PROTOCOL_VERSION } from "./protocol";
import { collectBugLog } from "./bugReport";
import { recentClientErrors, recordClientError, resetClientErrors } from "./clientErrors";

// Guard for #721: in-app bug reports inline the client log in a GitHub
// issue, and the log used to carry the WebSocket URL verbatim — session
// token included. Durable sessions (#517) would keep such a token valid
// across restarts, so these tests are the tripwire for any log line
// that reintroduces a credential.

const TOKEN = "Zk3q9_Rb-2xVn8LmPq4tYw7cHs1dJe6uKo0aBf5gTiA";
const WS_URL = `wss://cmd.example/ws?game=11111111-2222-3333-4444-555555555555&token=${TOKEN}&player=p-1`;

// UNREDACTED_TOKEN matches token=<anything but REDACTED>, in either the
// literal or percent-encoded form. A future log line that trips it has
// put a credential in bug reports.
const UNREDACTED_TOKEN = /token(=|%3d)(?!REDACTED)/i;

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
  resetClientErrors();
});

afterEach(() => {
  (globalThis as Record<string, unknown>).WebSocket = realWebSocket;
});

describe("#721 — the ws connect log line never carries the token", () => {
  it("connectLogLine redacts the token and keeps host, path and game", () => {
    const line = connectLogLine(WS_URL);
    expect(line).not.toContain(TOKEN);
    expect(line).not.toMatch(UNREDACTED_TOKEN);
    expect(line).toContain("wss://cmd.example/ws?game=11111111-2222-3333-4444-555555555555");
    expect(line).toContain("token=REDACTED");
  });

  it("a real connect, drop and reconnect leaves no token in the protocol log", () => {
    const client = new GameClient(WS_URL);
    client.connect();
    const socket = FakeSocket.last!;
    // The socket itself must still dial with the real token: redaction
    // is for the log, not the wire.
    expect(socket.url).toBe(WS_URL);
    socket.emit("open", {});

    const lines = get(client.log).map((e) => e.text);
    expect(lines.some((l) => l.startsWith("connected to "))).toBe(true);
    for (const l of lines) {
      expect(l).not.toContain(TOKEN);
      expect(l).not.toMatch(UNREDACTED_TOKEN);
    }
    client.disconnect();
  });

  it("redacts credentials that arrive in server frames (chat, error messages)", () => {
    const client = new GameClient(WS_URL);
    client.connect();
    const socket = FakeSocket.last!;
    socket.emit("open", {});
    socket.emit("message", {
      data: JSON.stringify({
        v: PROTOCOL_VERSION,
        kind: "error",
        id: "e1",
        payload: { code: "bad_request", message: `retry /ws?token=${TOKEN}` },
      }),
    });
    for (const e of get(client.log)) {
      expect(e.text).not.toContain(TOKEN);
    }
    client.disconnect();
  });
});

describe("#721 — the bug-report log is redacted end to end", () => {
  it("console capture and collectBugLog strip tokens from every line", () => {
    recordClientError(`console.error: fetch failed for /avatars/1/a.png?token=${TOKEN}`);
    recordClientError(`console.warn: {"token":"${TOKEN}"}`);
    expect(recentClientErrors().every((e) => !e.text.includes(TOKEN))).toBe(true);

    // A buffer that forgot to redact (a stale entry, a new source) is
    // still caught at the merge.
    const merged = collectBugLog(
      [{ at: new Date(1), direction: "info", text: `connected to ${WS_URL}` }],
      [{ at: 2, text: `Authorization: Bearer ${TOKEN}` }],
    );
    expect(merged).toHaveLength(2);
    for (const e of merged) {
      expect(e.text).not.toContain(TOKEN);
      expect(e.text).not.toMatch(UNREDACTED_TOKEN);
    }
  });
});
