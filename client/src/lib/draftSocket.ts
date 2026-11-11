// draftSocket.ts — the draft socket, /draft/{id}/ws (ADR 0222 §5, #5438;
// docs/protocol.md "The draft socket").
//
// A small sibling of ws.ts's GameClient, not a fork of it: a draft has
// one frame in (`draft`, the whole per-seat view) and one out
// (`draft_pick`). What it shares with the game socket is the reconnect
// convention, imported from ws.ts so the two cannot drift:
//
//   - the backoff ladder (reconnectDelayMs), retried for as long as the
//     page is open;
//   - close 1000 is terminal (isTerminalClose): the server sends it with
//     the reason "draft closed" when a draft closes, and redialling
//     would be dialling for something that is not there;
//   - a browser cannot read a refused upgrade (401, 403 and 404 all
//     surface as 1006), so after a few failed dials the socket asks over
//     HTTP whether the draft and the seat are still there, and stops for
//     good on a definite no;
//   - an application heartbeat, so a half-open connection is noticed in
//     seconds. It matters more here than at a table: a seat with no
//     socket is picked for ten seconds into a pick.
//
// It holds no Svelte store. The route owns the state and is told through
// the handlers, which keeps this file testable with a fake socket and
// fake timers.

import type { DraftView } from "./draft";
import { ErrorCode, PROTOCOL_VERSION, uuid, type Kind } from "./protocol";
import { DEAD_SESSION_CHECK_AFTER_FAILURES, isTerminalClose, reconnectDelayMs } from "./ws";

/**
 * connecting: the first dial. connected: frames are flowing.
 * reconnecting: the socket dropped and the ladder is redialling.
 * closed: the server closed the draft; no redial. gone: the draft or
 * this seat is no longer there; no redial, and the seat token is dead.
 */
export type DraftSocketStatus = "connecting" | "connected" | "reconnecting" | "closed" | "gone";

/** The reason the server gives when it closes a draft's sockets. */
export const DRAFT_CLOSED_REASON = "draft closed";

/** The refusal of a `draft_pick`. */
export const DRAFT_PICK_REFUSED = ErrorCode.DraftPickRefused;

// The same cadence as the game socket's heartbeat (ws.ts, #3047).
export const DRAFT_HEARTBEAT_INTERVAL_MS = 5_000;
export const DRAFT_HEARTBEAT_TIMEOUT_MS = DRAFT_HEARTBEAT_INTERVAL_MS * 2;
const CLOSE_HEARTBEAT_TIMEOUT = 4002;

export interface DraftFrame<P = unknown> {
  v: number;
  kind: string;
  id: string;
  payload?: P;
}

export interface DraftError {
  /** The id of the frame it answers, when it answers one. */
  id: string;
  code: string;
  message: string;
}

/** What a failed dial learns over HTTP: is the draft, and the seat, still there? */
export type DraftGoneCheck = () => Promise<"here" | "gone" | "unknown">;

interface SocketLike {
  readyState: number;
  send(data: string): void;
  close(code?: number, reason?: string): void;
  addEventListener(type: string, fn: (ev: never) => void): void;
  removeEventListener(type: string, fn: (ev: never) => void): void;
}

export interface DraftSocketOptions {
  url: string;
  onView: (view: DraftView) => void;
  onStatus: (status: DraftSocketStatus) => void;
  onError: (err: DraftError) => void;
  /** Asked after failed dials; "gone" stops the socket for good. */
  check?: DraftGoneCheck;
  /** For tests: the socket constructor and the jitter source. */
  socket?: new (url: string) => SocketLike;
  rand?: () => number;
}

const OPEN = 1;

export class DraftSocket {
  private socket: SocketLike | null = null;
  private status: DraftSocketStatus = "connecting";
  private stopped = false;
  // Consecutive dials that ended without a `draft` frame.
  private failures = 0;
  private everHadView = false;
  private gotViewThisDial = false;
  private lastSeq = -1;
  private lastFrameAt = 0;
  private retry: ReturnType<typeof setTimeout> | null = null;
  private beat: ReturnType<typeof setInterval> | null = null;
  private checking = false;
  private detach: (() => void) | null = null;

  constructor(private readonly opts: DraftSocketOptions) {}

  /** Dial. Safe to call once; disconnect() ends it. */
  connect(): void {
    if (this.stopped || this.socket) return;
    this.dial();
  }

  /** Hang up for good, from this side: no redial, no status change. */
  disconnect(): void {
    this.stopped = true;
    this.clearTimers();
    this.dropSocket(1000, "leaving");
  }

  /** True when a frame sent now would reach the server. */
  isOpen(): boolean {
    return this.socket?.readyState === OPEN;
  }

  /**
   * Send a pick. Returns the frame's id, which a refusal echoes, or
   * null when the socket is not open and nothing was sent.
   */
  pick(cards: readonly string[]): string | null {
    return this.send("draft_pick", { cards: [...cards] });
  }

  private send(kind: Kind, payload: unknown): string | null {
    const socket = this.socket;
    if (!socket || socket.readyState !== OPEN) return null;
    const frame: DraftFrame = { v: PROTOCOL_VERSION, kind, id: uuid(), payload };
    try {
      socket.send(JSON.stringify(frame));
    } catch {
      return null;
    }
    return frame.id;
  }

  private setStatus(next: DraftSocketStatus): void {
    if (this.status === next) return;
    this.status = next;
    this.opts.onStatus(next);
  }

  private dial(): void {
    const Ctor = this.opts.socket ?? (WebSocket as unknown as new (url: string) => SocketLike);
    let socket: SocketLike;
    try {
      socket = new Ctor(this.opts.url);
    } catch {
      this.dropped();
      return;
    }
    this.socket = socket;
    this.gotViewThisDial = false;
    this.lastFrameAt = Date.now();

    const onMessage = (ev: { data: unknown }): void => this.message(socket, ev.data);
    const onClose = (ev: { code?: number; reason?: string }): void => {
      if (this.socket !== socket) return;
      this.release();
      if (isTerminalClose(ev.code ?? 0)) {
        // 1000: the draft closed (reason "draft closed"). Do not redial.
        this.stopped = true;
        this.clearTimers();
        this.setStatus("closed");
        return;
      }
      this.dropped();
    };
    socket.addEventListener("message", onMessage as (ev: never) => void);
    socket.addEventListener("close", onClose as (ev: never) => void);
    this.detach = () => {
      socket.removeEventListener("message", onMessage as (ev: never) => void);
      socket.removeEventListener("close", onClose as (ev: never) => void);
    };

    this.beat = setInterval(() => this.heartbeat(socket), DRAFT_HEARTBEAT_INTERVAL_MS);
  }

  private heartbeat(socket: SocketLike): void {
    if (this.socket !== socket) return;
    if (Date.now() - this.lastFrameAt > DRAFT_HEARTBEAT_TIMEOUT_MS) {
      // Half-open: nothing has come back for two beats. Treat it as a
      // drop without waiting for a close event that may never come.
      this.dropSocket(CLOSE_HEARTBEAT_TIMEOUT, "heartbeat timeout");
      this.dropped();
      return;
    }
    this.send("ping", { msg: "heartbeat" });
  }

  private message(socket: SocketLike, data: unknown): void {
    if (this.socket !== socket || typeof data !== "string") return;
    let frame: DraftFrame;
    try {
      frame = JSON.parse(data) as DraftFrame;
    } catch {
      return;
    }
    if (!frame || typeof frame.kind !== "string") return;
    this.lastFrameAt = Date.now();
    switch (frame.kind) {
      case "draft": {
        const view = frame.payload as DraftView | undefined;
        if (!view || typeof view.seq !== "number") return;
        this.gotViewThisDial = true;
        this.everHadView = true;
        this.failures = 0;
        this.setStatus("connected");
        // `seq` only rises. A second tab's echo or a late frame from
        // before a reconnect never rolls the page back.
        if (view.seq < this.lastSeq) return;
        this.lastSeq = view.seq;
        this.opts.onView(view);
        return;
      }
      case "error": {
        const p = frame.payload as { code?: unknown; message?: unknown } | undefined;
        this.opts.onError({
          id: frame.id ?? "",
          code: typeof p?.code === "string" ? p.code : "",
          message: typeof p?.message === "string" ? p.message : "",
        });
        return;
      }
      default:
        // pong, and anything a newer server adds: nothing to do.
        return;
    }
  }

  // release forgets the current socket without closing it.
  private release(): void {
    this.detach?.();
    this.detach = null;
    this.socket = null;
    if (this.beat) clearInterval(this.beat);
    this.beat = null;
  }

  private dropSocket(code: number, reason: string): void {
    const socket = this.socket;
    this.release();
    try {
      socket?.close(code, reason);
    } catch {
      // already closed
    }
  }

  private clearTimers(): void {
    if (this.retry) clearTimeout(this.retry);
    this.retry = null;
    if (this.beat) clearInterval(this.beat);
    this.beat = null;
  }

  // dropped: the socket went away without the server saying the draft
  // closed. Redial on the ladder, and ask whether there is still
  // something to dial for.
  private dropped(): void {
    if (this.stopped) return;
    if (!this.gotViewThisDial) this.failures += 1;
    this.setStatus(this.everHadView ? "reconnecting" : "connecting");
    // A seat that has never been let in is asked about at once: the
    // usual cause is a dead token. One that was connected gets the same
    // patience as the game socket, so a deploy costs no extra request.
    const askAfter = this.everHadView ? DEAD_SESSION_CHECK_AFTER_FAILURES : 1;
    if (this.failures >= askAfter) this.ask();
    const delay = reconnectDelayMs(Math.max(0, this.failures - 1), this.opts.rand);
    this.retry = setTimeout(() => {
      this.retry = null;
      if (!this.stopped && !this.socket) this.dial();
    }, delay);
  }

  private ask(): void {
    const check = this.opts.check;
    if (!check || this.checking) return;
    this.checking = true;
    void check()
      .catch((): "unknown" => "unknown")
      .then((answer) => {
        this.checking = false;
        if (answer !== "gone" || this.stopped) return;
        this.stopped = true;
        this.clearTimers();
        this.dropSocket(1000, "gone");
        this.setStatus("gone");
      });
  }
}
