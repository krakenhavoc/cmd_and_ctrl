import { writable, type Writable } from "svelte/store";
import {
  PROTOCOL_VERSION,
  uuid,
  type Frame,
  type PingPayload,
  type PongPayload,
  type ErrorPayload,
  type SnapshotPayload,
  type GameView,
} from "./protocol";

export type ConnectionStatus = "connecting" | "connected" | "disconnected";

export type LogDirection = "sent" | "received" | "error" | "info";

export interface LogEntry {
  id: string;
  at: Date;
  direction: LogDirection;
  text: string;
}

// GameClient wraps a WebSocket with the v0 protocol and exposes reactive
// Svelte stores for UI binding. As of S03 it understands `ping`/`pong`,
// `error`, and `snapshot` frames. `action` frames (client → server) are
// sent by the caller via sendAction; they are not yet exposed via a
// typed helper surface because the S05+ play UI hasn't landed.
export class GameClient {
  readonly status: Writable<ConnectionStatus> = writable("disconnected");
  readonly log: Writable<LogEntry[]> = writable([]);
  // snapshot is the latest GameView received from the server, or null
  // before the initial snapshot arrives. Updated on every snapshot
  // frame; UI components subscribe to it for the authoritative game
  // state. The snapshot sequence number is tracked separately so
  // consumers that care about ordering don't have to derive it.
  readonly snapshot: Writable<GameView | null> = writable(null);
  readonly lastSeq: Writable<number> = writable(0);

  private socket: WebSocket | null = null;
  // highestSeq tracks the largest `seq` seen so far. The protocol
  // guarantees monotonically non-decreasing seq; a duplicate is
  // possible during a join-while-action race and is treated as a
  // no-op state refresh (same state, same seq).
  private highestSeq = 0;

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
      this.append("info", `connected to ${this.url}`);
    });

    socket.addEventListener("message", (ev: MessageEvent<unknown>) => {
      if (!isCurrent()) return;
      this.handleMessage(ev.data);
    });

    socket.addEventListener("close", () => {
      if (!isCurrent()) return;
      this.status.set("disconnected");
      this.socket = null;
      this.append("info", "socket closed");
    });

    socket.addEventListener("error", () => {
      if (!isCurrent()) return;
      this.append("error", "socket error");
    });
  }

  disconnect(): void {
    const socket = this.socket;
    this.socket = null;
    // Explicitly flip status before closing: the close listener on the
    // captured socket will short-circuit (isCurrent() is now false), so
    // without this line the status store would stay stuck at whatever
    // it was, misrepresenting a deliberate disconnect in the UI.
    this.status.set("disconnected");
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

  // sendAction submits an action frame to the server. Returns the
  // generated frame id for correlation with subsequent snapshot/error
  // frames. The caller is responsible for typing the params shape
  // against the docs/protocol.md action catalog; this method does not
  // validate. Intended for the dev tooling and S05+ play UI.
  sendAction(type: string, player?: string, params?: unknown): string | null {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      this.append("error", "not connected");
      return null;
    }
    const id = uuid();
    const frame: Frame = {
      v: PROTOCOL_VERSION,
      kind: "action",
      id,
      payload: { type, player, params },
    };
    this.socket.send(JSON.stringify(frame));
    this.append("sent", `action ${type} id=${id.slice(0, 8)}`);
    return id;
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
      case "snapshot": {
        const p = frame.payload as SnapshotPayload | undefined;
        if (!p) {
          this.append("error", "snapshot frame had no payload");
          break;
        }
        // Protocol guarantees seq is monotonically non-decreasing.
        // Duplicate seqs can legitimately occur during a join-while-
        // action race (same state captured twice); only older seqs
        // indicate a real out-of-order delivery.
        if (p.seq < this.highestSeq) {
          this.append(
            "error",
            `snapshot seq=${p.seq} older than highest=${this.highestSeq}; ignored`,
          );
          break;
        }
        this.highestSeq = p.seq;
        this.snapshot.set(p.game);
        this.lastSeq.set(p.seq);
        this.append(
          "received",
          `snapshot seq=${p.seq} turn=${p.game.turn.number} seat=${p.game.turn.active_seat} step=${p.game.turn.step}`,
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
