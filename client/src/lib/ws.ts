import { writable, type Writable } from "svelte/store";
import {
  PROTOCOL_VERSION,
  uuid,
  type ActionType,
  type Frame,
  type PingPayload,
  type PongPayload,
  type ErrorPayload,
  type SnapshotPayload,
  type ChatPayload,
  type GameView,
} from "./protocol";

// "reconnecting" is the automatic-retry state after a non-deliberate
// close (network blip, server restart). Deliberate teardown via
// disconnect() lands on "disconnected" and never retries.
export type ConnectionStatus = "connecting" | "connected" | "reconnecting" | "disconnected";

export type LogDirection = "sent" | "received" | "error" | "info";

export interface LogEntry {
  id: string;
  at: Date;
  direction: LogDirection;
  text: string;
}

// ChatMessage is the local view of a server-stamped chat frame. The
// frame `id` doubles as the message id so de-duped renders are easy;
// `at` is parsed once on receipt so the UI can format without
// re-parsing on every reactive tick.
export interface ChatMessage {
  id: string;
  authorID: string;
  authorName: string;
  text: string;
  at: Date;
}

// CHAT_LOG_LIMIT caps the number of messages retained client-side. Old
// entries fall off the front of the log; chat history is ephemeral
// and not part of crash recovery.
const CHAT_LOG_LIMIT = 200;

// WS_LOG_LIMIT caps the protocol log ring buffer. This buffer is what
// an in-app bug report attaches (ADR 0017 §7), so it is sized to match
// the report's own cap — a smaller window here would silently truncate
// reports, and a larger one would be trimmed server-side anyway.
export const WS_LOG_LIMIT = 200;

// ERROR_TOAST_TTL_MS is how long a server error sticks in the
// lastError store before auto-clearing. Long enough to read, short
// enough that an unread error doesn't stay on screen indefinitely.
const ERROR_TOAST_TTL_MS = 5000;

// Automatic-reconnect backoff bounds: first retry after ~500ms,
// doubling per consecutive failure, capped at 30s, retried forever.
export const RECONNECT_BASE_MS = 500;
export const RECONNECT_CAP_MS = 30_000;

// reconnectDelayMs computes the delay before reconnect attempt
// `attempt` (0-based): exponential in the attempt count, capped at
// RECONNECT_CAP_MS, with half-jitter — the result lands uniformly in
// [nominal/2, nominal] so four clients dropped by the same server
// restart don't stampede back in lockstep, while still honouring a
// meaningful minimum wait. `rand` is injectable for tests.
export function reconnectDelayMs(attempt: number, rand: () => number = Math.random): number {
  const nominal = Math.min(RECONNECT_CAP_MS, RECONNECT_BASE_MS * 2 ** attempt);
  return Math.floor(nominal / 2 + rand() * (nominal / 2));
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
  // chat is the rolling list of server-stamped chat messages received
  // on this connection. Capped to CHAT_LOG_LIMIT entries; older
  // messages drop off the front. Chat is ephemeral — a fresh connect
  // always starts with an empty log.
  readonly chat: Writable<ChatMessage[]> = writable([]);
  // lastError holds the most recent error frame received from the
  // server. Routes subscribe to render a toast / banner so users
  // can see why an action was rejected (the v0 protocol replies
  // with error frames addressed only to the originating client).
  // Auto-cleared after ERROR_TOAST_TTL_MS so a stale error doesn't
  // sit forever; a new error replaces the old one immediately.
  readonly lastError: Writable<{
    code: string;
    message: string;
    at: Date;
    missing?: string[];
    cardID?: string;
  } | null> = writable(null);
  private errorClearTimer: ReturnType<typeof setTimeout> | null = null;

  private socket: WebSocket | null = null;
  // highestSeq tracks the largest `seq` seen so far. The protocol
  // guarantees monotonically non-decreasing seq; a duplicate is
  // possible during a join-while-action race and is treated as a
  // no-op state refresh (same state, same seq). Reset on setURL /
  // connect / socket open — the watermark only orders frames within
  // one connection to one server incarnation.
  private highestSeq = 0;
  // reconnectTimer is the pending automatic-reconnect timeout, or
  // null when none is scheduled. Non-null doubles as the "we are in
  // a retry loop" flag; deliberate disconnect()/connect() cancel it.
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  // reconnectAttempts counts consecutive failed opens since the last
  // successful one — the exponent for the backoff ladder.
  private reconnectAttempts = 0;

  constructor(private url: string) {}

  // setURL swaps the target URL. Does not affect an already-open
  // socket — callers who want the new URL to take effect should
  // disconnect() + connect() after setting. Exposed so route
  // components can rebuild the WS URL on session change without
  // throwing away the GameClient instance and its subscribers.
  //
  // Retargeting (game A → game B) invalidates the seq watermark and
  // the rendered snapshot: game B starts its own seq numbering, and
  // without the reset every one of its snapshots would be dropped as
  // "out of order" while the board kept showing game A's state under
  // a green "connected" tag. (App.svelte's {#key gameID} remount is
  // the second belt for the same failure mode.)
  setURL(url: string): void {
    this.url = url;
    this.resetSnapshotTracking();
  }

  connect(): void {
    if (this.socket) {
      return;
    }
    // A deliberate connect supersedes any scheduled auto-reconnect
    // and restarts the backoff ladder and seq watermark fresh.
    this.cancelReconnect();
    this.reconnectAttempts = 0;
    this.resetSnapshotTracking();
    this.status.set("connecting");
    this.open();
  }

  // open dials the socket and installs listeners. Shared by the
  // deliberate connect() path and the automatic-reconnect timer;
  // status is set by the caller ("connecting" vs "reconnecting").
  private open(): void {
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
      this.reconnectAttempts = 0;
      // Reset the seq watermark on every open, including automatic
      // reconnects to the same URL: a server restart restarts seq from
      // scratch, and keeping the old watermark would silently drop
      // every snapshot on the new connection. Snapshots are idempotent
      // full states, so re-accepting one duplicate frame after a
      // same-incarnation reconnect is harmless — a frozen board is not.
      this.highestSeq = 0;
      this.append("info", `connected to ${this.url}`);
    });

    socket.addEventListener("message", (ev: MessageEvent<unknown>) => {
      if (!isCurrent()) return;
      this.handleMessage(ev.data);
    });

    socket.addEventListener("close", (ev: CloseEvent) => {
      if (!isCurrent()) return;
      // Deliberate teardown (disconnect()) nulls this.socket before
      // closing, so a close event that reaches this point was not
      // requested by this client. But the SERVER's deliberate closes
      // are terminal, not retryable: hub.EvictGame sends 1000 "game
      // deleted" and shutdown sends 1001 "server shutting down"
      // precisely so the peer renders an ended state instead of
      // redialling a game that is gone. Anything else (network blip,
      // crash, rejected upgrade surfacing as 1006) gets the retry
      // loop.
      this.socket = null;
      if (ev.code === 1000 || ev.code === 1001) {
        this.append(
          "info",
          `socket closed by server (${ev.code}${ev.reason ? `: ${ev.reason}` : ""})`,
        );
        this.status.set("disconnected");
        return;
      }
      this.append("info", "socket closed");
      this.scheduleReconnect();
    });

    socket.addEventListener("error", () => {
      if (!isCurrent()) return;
      this.append("error", "socket error");
    });
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer !== null) return;
    this.status.set("reconnecting");
    const delay = reconnectDelayMs(this.reconnectAttempts);
    this.reconnectAttempts += 1;
    this.append("info", `reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      // A deliberate connect() may have raced the timer; don't stack
      // a second socket on top of it.
      if (this.socket) return;
      this.open();
    }, delay);
  }

  private cancelReconnect(): void {
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  // resetSnapshotTracking clears the seq watermark and the rendered
  // snapshot — used on deliberate retarget (setURL) / connect so a
  // new game's frames are never compared against an old game's seqs.
  // NOT used on automatic reconnect: the stale board stays rendered
  // under the "reconnecting" tag until the fresh snapshot lands.
  private resetSnapshotTracking(): void {
    this.highestSeq = 0;
    this.snapshot.set(null);
    this.lastSeq.set(0);
  }

  disconnect(): void {
    this.cancelReconnect();
    this.reconnectAttempts = 0;
    const socket = this.socket;
    this.socket = null;
    // Explicitly flip status before closing: the close listener on the
    // captured socket will short-circuit (isCurrent() is now false), so
    // without this line the status store would stay stuck at whatever
    // it was, misrepresenting a deliberate disconnect in the UI. This
    // is also what makes the close deliberate — the nulled socket means
    // no reconnect gets scheduled.
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

  // sendChat submits a chat frame to the server. The server stamps
  // author_id, author_name, and timestamp from the connection's
  // principal — only the text is honoured from this side. Returns the
  // generated frame id, or null if the socket isn't open.
  sendChat(text: string): string | null {
    const trimmed = text.trim();
    if (!trimmed) {
      return null;
    }
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      this.append("error", "not connected");
      return null;
    }
    const id = uuid();
    const frame: Frame<ChatPayload> = {
      v: PROTOCOL_VERSION,
      kind: "chat",
      id,
      payload: { author_name: "", text: trimmed, timestamp: "" },
    };
    this.socket.send(JSON.stringify(frame));
    this.append("sent", `chat id=${id.slice(0, 8)} text=${JSON.stringify(trimmed)}`);
    return id;
  }

  // sendAction submits an action frame to the server. Returns the
  // generated frame id for correlation with subsequent snapshot/error
  // frames. The caller is responsible for typing the params shape
  // against the docs/protocol.md action catalog; this method does not
  // validate params. The `type` union (protocol.ts ActionType) keeps
  // action names typo-proof against the server registry.
  sendAction(type: ActionType, player?: string, params?: unknown): string | null {
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
      case "chat": {
        const p = frame.payload as ChatPayload | undefined;
        if (!p) {
          this.append("error", "chat frame had no payload");
          break;
        }
        const msg: ChatMessage = {
          id: frame.id,
          authorID: p.author_id ?? "",
          authorName: p.author_name,
          text: p.text,
          at: new Date(p.timestamp),
        };
        this.chat.update((entries) => {
          const next = [...entries, msg];
          return next.length > CHAT_LOG_LIMIT ? next.slice(-CHAT_LOG_LIMIT) : next;
        });
        this.append("received", `chat from=${msg.authorName} text=${JSON.stringify(msg.text)}`);
        break;
      }
      case "error": {
        const p = frame.payload as ErrorPayload | undefined;
        const code = p?.code ?? "?";
        const message = p?.message ?? "?";
        this.append("error", `server error code=${code} message=${message}`);
        // Surface visibly so the user sees why an action was
        // rejected. Reset the auto-clear timer so a fresh error
        // gets its full TTL even if a previous one is still showing.
        this.lastError.set({
          code,
          message,
          at: new Date(),
          missing: p?.missing,
          cardID: p?.card_id,
        });
        if (this.errorClearTimer !== null) {
          clearTimeout(this.errorClearTimer);
        }
        this.errorClearTimer = setTimeout(() => {
          this.lastError.set(null);
          this.errorClearTimer = null;
        }, ERROR_TOAST_TTL_MS);
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
    this.log.update((entries) => [...entries.slice(-(WS_LOG_LIMIT - 1)), entry]);
  }
}
