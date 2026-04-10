import { writable, type Writable } from "svelte/store";
import {
  PROTOCOL_VERSION,
  uuid,
  type Frame,
  type PingPayload,
  type PongPayload,
  type ErrorPayload,
} from "./protocol";

export type ConnectionStatus = "connecting" | "connected" | "disconnected";

export interface LogEntry {
  id: string;
  at: Date;
  direction: "sent" | "received" | "error";
  text: string;
}

// GameClient wraps a WebSocket with the v0 protocol and exposes reactive
// Svelte stores for UI binding. At S01 it only knows ping/pong; S02 will
// add game action types.
export class GameClient {
  readonly status: Writable<ConnectionStatus> = writable("disconnected");
  readonly log: Writable<LogEntry[]> = writable([]);

  private socket: WebSocket | null = null;

  constructor(private readonly url: string) {}

  connect(): void {
    if (this.socket) {
      return;
    }
    this.status.set("connecting");
    const socket = new WebSocket(this.url);
    this.socket = socket;

    // Each listener captures `socket` in its closure and ignores events
    // from any socket other than the currently active one. Without this
    // guard, a previous socket's async close event can fire after a new
    // socket has already been assigned during reconnect, silently
    // nulling out the live socket.
    const isCurrent = (): boolean => this.socket === socket;

    socket.addEventListener("open", () => {
      if (!isCurrent()) return;
      this.status.set("connected");
      this.append("sent", `connected to ${this.url}`);
    });

    socket.addEventListener("message", (ev) => {
      if (!isCurrent()) return;
      this.handleMessage(ev.data);
    });

    socket.addEventListener("close", () => {
      if (!isCurrent()) return;
      this.status.set("disconnected");
      this.socket = null;
      this.append("error", "socket closed");
    });

    socket.addEventListener("error", () => {
      if (!isCurrent()) return;
      this.append("error", "socket error");
    });
  }

  disconnect(): void {
    const socket = this.socket;
    this.socket = null;
    socket?.close();
  }

  sendPing(msg: string): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      this.append("error", "not connected");
      return;
    }
    const frame: Frame<PingPayload> = {
      v: PROTOCOL_VERSION,
      kind: "ping",
      id: uuid(),
      payload: { msg },
    };
    this.socket.send(JSON.stringify(frame));
    this.append("sent", `ping id=${frame.id.slice(0, 8)} msg=${JSON.stringify(msg)}`);
  }

  private handleMessage(raw: unknown): void {
    if (typeof raw !== "string") {
      this.append("error", "non-text frame received");
      return;
    }
    let frame: Frame;
    try {
      frame = JSON.parse(raw) as Frame;
    } catch {
      this.append("error", "malformed JSON from server");
      return;
    }
    if (frame.v !== PROTOCOL_VERSION) {
      this.append("error", `server spoke v${frame.v}, expected v${PROTOCOL_VERSION}`);
      return;
    }
    switch (frame.kind) {
      case "pong": {
        const p = frame.payload as PongPayload | undefined;
        this.append(
          "received",
          `pong id=${frame.id.slice(0, 8)} msg=${JSON.stringify(p?.msg ?? "")} at=${p?.server_time ?? "?"}`,
        );
        break;
      }
      case "error": {
        const p = frame.payload as ErrorPayload | undefined;
        this.append("error", `server error code=${p?.code ?? "?"} message=${p?.message ?? "?"}`);
        break;
      }
      default:
        this.append("error", `unknown kind from server: ${frame.kind}`);
    }
  }

  private append(direction: LogEntry["direction"], text: string): void {
    const entry: LogEntry = {
      id: uuid(),
      at: new Date(),
      direction,
      text,
    };
    this.log.update((entries) => [...entries.slice(-99), entry]);
  }
}
