import { type Writable } from "svelte/store";
import { recordClientError } from "./clientErrors";
import { OFFLINE_ERROR_CODE, offlineSendMessage } from "./connectionBanner";
import { describeThrown, guardedWritable } from "./guardedStore";
import { redactSecrets, redactURL } from "./redact";
import { checkSessionAlive, currentSession } from "./session";
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
  type ChatKind,
  type GameView,
} from "./protocol";

// "reconnecting" is the automatic-retry state after a non-deliberate
// close (network blip, server restart). Deliberate teardown via
// disconnect() lands on "disconnected" and never retries.
//
// "session_ended" is the other terminal state (#1475, per the
// owner's 2026-09-24 decision 5 on #515): the backoff ladder gave up
// asking "is the server just restarting?" and asked the server
// directly instead — GET /me came back 401, so the credential this
// socket is dialling with is gone, not merely unreachable. Nothing
// redials from here; the player has to sign in again.
export type ConnectionStatus =
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected"
  | "session_ended";

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
  // kind / reason carry the S31 bot lines. "say" for everything a
  // human typed. See botChat.ts for what the UI does with them.
  kind: ChatKind;
  reason: string;
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

// connectLogLine is the "connected" entry in the protocol log. The
// socket URL carries the session token as ?token= (browsers can't set
// headers on a WebSocket upgrade), and this log is inlined into
// GitHub issues by the bug-report button — so the URL is redacted
// here, at the source, rather than trusted to a later filter (#721).
// Host, path, game and player survive: they are what triage needs.
export function connectLogLine(url: string): string {
  return `connected to ${redactURL(url)}`;
}

// FRAME_LOG_LIMIT caps the dev frame inspector's ring buffer. A busy
// four-player turn is a few dozen frames, so 500 covers "what just
// happened" without letting a long game grow the heap unbounded.
export const FRAME_LOG_LIMIT = 500;

// FrameRecord is one raw WebSocket frame captured for the dev frame
// inspector (docs/decisions/0023-develop-environment.md). Recording is
// off unless a dev deployment turns it on via setFrameRecording, so
// production pays one boolean check per frame and nothing else.
export interface FrameRecord {
  // "in" is server -> client, "out" is client -> server.
  dir: "in" | "out";
  at: number;
  kind: string;
  // Frame id, for correlating an outbound action with the snapshot or
  // error it produced.
  id: string;
  // Snapshot sequence number, when the frame carries one.
  seq?: number;
  // Serialised size, so an unexpectedly fat snapshot is visible.
  bytes: number;
  // The parsed frame, for the detail pane.
  frame: unknown;
}

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

// DEAD_SESSION_CHECK_AFTER_FAILURES is how many consecutive failed
// dials the ladder serves before it stops guessing and asks the
// server directly whether this session is still good (#1475, per the
// owner's 2026-09-24 decision 5 on #515). Small enough to catch a
// revoked or expired session quickly; large enough that one ordinary
// drop — the close that opens every reconnect, deploy included —
// never fires a network call nobody asked for.
export const DEAD_SESSION_CHECK_AFTER_FAILURES = 3;

// Close codes the server sends on purpose. They used to share one
// terminal branch, which is the whole of #518: hub.Shutdown writes
// CLOSE_GOING_AWAY on every `systemctl restart`, the game outlives the
// process (ADR 0041), and the client hung up on it for good.
export const CLOSE_NORMAL = 1000;
export const CLOSE_GOING_AWAY = 1001;

// isTerminalClose reports whether a server-initiated close means "stop
// dialling" (ADR 0044 decision 2).
//
// Only 1000 does. hub.EvictGame sends it when the game is deleted, and
// a session revocation sends it when the credential the socket
// authenticated with is gone — redialling either one is dialling for
// something that is not there any more.
//
// 1001 "server shutting down" is a deploy, and a deploy is exactly the
// case the backoff ladder below was written for: the process goes away
// for 15-60s and the table comes back with it. Everything else — a
// network blip, a crash, a rejected upgrade that surfaces as 1006 —
// keeps retrying as it always did.
//
// Deliberate teardown from THIS side does not reach here at all:
// disconnect() nulls the socket first, so its close event is ignored.
export function isTerminalClose(code: number): boolean {
  return code === CLOSE_NORMAL;
}

// SESSION_TOKEN_PARAM is the query parameter that carries the session
// token on a WS upgrade — browsers cannot set headers on a WebSocket
// handshake, so gameURL.ts bakes it into the URL.
export const SESSION_TOKEN_PARAM = "token";

// withSessionToken returns `url` with the session token parameter set
// to `token`, which is how a dial picks up a session that changed
// after the URL was built (#518).
//
// A missing token leaves the URL untouched: the captured string is
// still the best guess available, and a spectator dial legitimately
// carries no credential. So does a URL the platform will not parse —
// dialling a stale token 401s and retries, while dialling a mangled
// URL fails forever.
export function withSessionToken(url: string, token: string | null | undefined): string {
  if (!token) return url;
  try {
    const parsed = new URL(url);
    // The common case: the captured URL already carries this exact
    // token. Returned verbatim rather than re-serialised, so a dial
    // that changes nothing cannot change the string either.
    if (parsed.searchParams.get(SESSION_TOKEN_PARAM) === token) return url;
    parsed.searchParams.set(SESSION_TOKEN_PARAM, token);
    return parsed.toString();
  } catch {
    return url;
  }
}

// guardedWritable / describeThrown moved to guardedStore.ts (#720).
// svelte/store's `subscriber_queue` is module-GLOBAL, so guarding only
// GameClient's stores — which is all #284 did — protects nothing: a
// throw from a subscriber on `settings`, `targeting` or any other
// store stalls these guarded stores just as thoroughly. The guard
// belongs to every store in the app, so it lives in one place now.

// GameClient wraps a WebSocket with the v0 protocol and exposes reactive
// Svelte stores for UI binding. As of S03 it understands `ping`/`pong`,
// `error`, and `snapshot` frames. `action` frames (client → server) are
// sent by the caller via sendAction; they are not yet exposed via a
// typed helper surface because the S05+ play UI hasn't landed.
export class GameClient {
  readonly status: Writable<ConnectionStatus> = guardedWritable("disconnected", "status");
  readonly log: Writable<LogEntry[]> = guardedWritable([], "log");
  // snapshot is the latest GameView received from the server, or null
  // before the initial snapshot arrives. Updated on every snapshot
  // frame; UI components subscribe to it for the authoritative game
  // state. The snapshot sequence number is tracked separately so
  // consumers that care about ordering don't have to derive it.
  readonly snapshot: Writable<GameView | null> = guardedWritable(null, "snapshot");
  readonly lastSeq: Writable<number> = guardedWritable(0, "lastSeq");
  // frames is the dev frame inspector's ring buffer. Empty and never
  // written unless setFrameRecording(true) has been called, which
  // only happens on a dev deployment with the frame_inspector feature
  // enabled — see lib/env.ts.
  readonly frames: Writable<FrameRecord[]> = guardedWritable([], "frames");
  // chat is the rolling list of server-stamped chat messages received
  // on this connection. Capped to CHAT_LOG_LIMIT entries; older
  // messages drop off the front. Chat is ephemeral — a fresh connect
  // always starts with an empty log.
  readonly chat: Writable<ChatMessage[]> = guardedWritable([], "chat");
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
    // #1063: the attack-tax refusal's whole price as a cost string
    // ("{2}{2}"), or a block refusal's stable reason token — whichever
    // the error frame's `reason` field carries for this code. See
    // ErrorPayload.reason (protocol.ts).
    reason?: string;
  } | null> = guardedWritable(null, "lastError");
  // reconnectAttempt is how many automatic reconnects have been
  // scheduled since the last successful open, and 0 whenever the
  // socket is up or the client was torn down deliberately. The ladder
  // used to count privately and retry forever in silence; this is the
  // number a "reconnecting (attempt 3)" banner reads (#518, rendered
  // by #519).
  readonly reconnectAttempt: Writable<number> = guardedWritable(0, "reconnectAttempt");
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
  // reconnectGeneration increments every time the backoff ladder is
  // reset for real — a successful open, or a deliberate connect() /
  // disconnect() — never on a mere failed dial. A dead-session check
  // captures it before making its network call; if it no longer
  // matches when the check resolves, the world has moved on (the
  // socket came back, or the client was torn down) and the stale
  // verdict is discarded rather than acted on.
  private reconnectGeneration = 0;
  // sessionCheckDone / sessionCheckInFlight bound the dead-session
  // probe to ONE per failure streak: done once a probe has resolved
  // (so a long outage doesn't keep hitting the server), and in-flight
  // while its fetch is still out (so two probes never race). Both are
  // cleared alongside reconnectAttempts on the next successful open,
  // so a fresh run of failures after a healthy stretch gets a fresh
  // probe.
  private sessionCheckDone = false;
  private sessionCheckInFlight = false;
  // actionsSent counts action frames that left the socket. A timed
  // bluff (#1307) reads it to tell whether the viewer did anything
  // while it was waiting, whichever of the many send paths they used.
  private sentActions = 0;
  get actionsSent(): number {
    return this.sentActions;
  }
  // recordFrames gates every capture. Kept as a plain boolean rather
  // than reading a store per frame so the disabled path is a single
  // branch on the hot receive loop.
  private recordFrames = false;

  constructor(private url: string) {}

  // setFrameRecording turns raw frame capture on or off. Callers are
  // responsible for only enabling it when the server reports the
  // frame_inspector dev feature; this method itself is unprivileged
  // (it reveals nothing the client did not already send or receive).
  // Turning it off clears the buffer so a captured game state does not
  // linger in memory after the drawer is closed.
  setFrameRecording(on: boolean): void {
    this.recordFrames = on;
    if (!on) {
      this.frames.set([]);
    }
  }

  // clearFrames empties the ring buffer without changing the
  // recording state — the inspector's "clear" button.
  clearFrames(): void {
    this.frames.set([]);
  }

  private recordFrame(dir: FrameRecord["dir"], frame: Frame, serialised: string): void {
    if (!this.recordFrames) return;
    const seq = (frame.payload as { seq?: unknown } | undefined)?.seq;
    const rec: FrameRecord = {
      dir,
      at: Date.now(),
      kind: String(frame.kind ?? "?"),
      id: String(frame.id ?? ""),
      seq: typeof seq === "number" ? seq : undefined,
      bytes: serialised.length,
      frame,
    };
    this.frames.update((prev) => {
      const next = [...prev, rec];
      return next.length > FRAME_LOG_LIMIT ? next.slice(-FRAME_LOG_LIMIT) : next;
    });
  }

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
    this.resetReconnectAttempts();
    this.resetSnapshotTracking();
    this.status.set("connecting");
    this.open();
  }

  // open dials the socket and installs listeners. Shared by the
  // deliberate connect() path and the automatic-reconnect timer;
  // status is set by the caller ("connecting" vs "reconnecting").
  private open(): void {
    // Resolved per dial, not per mount. `this.url` was built when the
    // route mounted and gameURL.ts baked the session token of that
    // moment into it; the reconnect ladder can easily outlive that
    // token. Re-reading the session store here is what lets a retry
    // after a deploy present the credential the new process will
    // actually accept, and picks up a session refreshed in another
    // tab (#518). The URL is captured in this closure so the log line
    // and the socket can never disagree about what was dialled.
    const url = this.dialURL();
    const socket = new WebSocket(url);
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
      this.resetReconnectAttempts();
      // Reset the seq watermark on every open, including automatic
      // reconnects to the same URL: a server restart restarts seq from
      // scratch, and keeping the old watermark would silently drop
      // every snapshot on the new connection. Snapshots are idempotent
      // full states, so re-accepting one duplicate frame after a
      // same-incarnation reconnect is harmless — a frozen board is not.
      this.highestSeq = 0;
      this.append("info", connectLogLine(url));
    });

    socket.addEventListener("message", (ev: MessageEvent<unknown>) => {
      if (!isCurrent()) return;
      this.handleMessage(ev.data);
    });

    socket.addEventListener("close", (ev: CloseEvent) => {
      if (!isCurrent()) return;
      // Deliberate teardown (disconnect()) nulls this.socket before
      // closing, so a close event that reaches this point was not
      // requested by this client. Of the server's own deliberate
      // closes only 1000 is terminal — see isTerminalClose. 1001 is a
      // deploy, and ADR 0041 persists and restores the game across
      // one, so it takes the retry ladder like any other drop.
      this.socket = null;
      if (isTerminalClose(ev.code)) {
        this.append(
          "info",
          `socket closed by server (${ev.code}${ev.reason ? `: ${ev.reason}` : ""})`,
        );
        this.status.set("disconnected");
        return;
      }
      this.append("info", `socket closed (${ev.code}${ev.reason ? `: ${ev.reason}` : ""})`);
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
    this.reconnectAttempt.set(this.reconnectAttempts);
    this.append("info", `reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
    this.maybeCheckDeadSession();
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      // A deliberate connect() may have raced the timer; don't stack
      // a second socket on top of it.
      if (this.socket) return;
      this.open();
    }, delay);
  }

  // maybeCheckDeadSession fires the #1475 probe once a streak of
  // failed dials reaches DEAD_SESSION_CHECK_AFTER_FAILURES, and runs
  // ALONGSIDE the backoff ladder rather than instead of it — a server
  // that is only restarting must keep being retried at full speed
  // while the probe is in flight, so this never delays or cancels the
  // reconnect timer scheduleReconnect just set. Only a definite "dead"
  // verdict changes anything, from inside the probe's own resolution.
  private maybeCheckDeadSession(): void {
    if (this.sessionCheckDone || this.sessionCheckInFlight) return;
    if (this.reconnectAttempts < DEAD_SESSION_CHECK_AFTER_FAILURES) return;
    this.sessionCheckInFlight = true;
    const generation = this.reconnectGeneration;
    void checkSessionAlive().then((result) => {
      this.sessionCheckInFlight = false;
      this.sessionCheckDone = true;
      // Stale: a successful open or a deliberate connect()/disconnect()
      // happened while the fetch was out. Whatever it found no longer
      // describes the world this client is in.
      if (generation !== this.reconnectGeneration) return;
      if (result === "dead") {
        this.append("info", "session check: server says 401 — the session is gone");
        this.enterSessionEnded();
        return;
      }
      this.append("info", `session check: ${result} — keeping the reconnect ladder`);
    });
  }

  // enterSessionEnded stops the reconnect ladder for good and moves to
  // the terminal "session_ended" status the connection banner reads as
  // "sign in again" (#1475). It closes any live socket the same way
  // disconnect() does, so a close event that socket still delivers is
  // ignored (isCurrent() below is keyed off `this.socket`) rather than
  // scheduling one more reconnect on top of this.
  private enterSessionEnded(): void {
    this.cancelReconnect();
    const socket = this.socket;
    this.socket = null;
    this.status.set("session_ended");
    socket?.close();
  }

  // resetReconnectAttempts zeroes the backoff exponent and the store
  // the UI reads off it, and — since #1475 — the dead-session probe's
  // own bookkeeping and generation counter. The three must move
  // together: a stale attempt count on a healthy connection is a
  // banner that lies, and a probe left "in flight" or "done" from a
  // previous failure streak would either never fire again or fire on
  // top of a connection that has since recovered.
  private resetReconnectAttempts(): void {
    this.reconnectAttempts = 0;
    this.reconnectAttempt.set(0);
    this.sessionCheckDone = false;
    this.sessionCheckInFlight = false;
    this.reconnectGeneration += 1;
  }

  // dialURL is the URL to dial right now: the target this client was
  // pointed at, carrying the session token as the store holds it at
  // this instant.
  private dialURL(): string {
    return withSessionToken(this.url, currentSession()?.token);
  }

  private cancelReconnect(): void {
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  // retryNow dials immediately instead of waiting out the rung of the
  // backoff ladder currently being served — the connection banner's
  // "Try now" button (#519).
  //
  // Deliberately NOT connect(). connect() calls resetSnapshotTracking,
  // which blanks the board; the stale board staying rendered under the
  // banner is the design (ADR 0044 / #519), and the fix for a dropped
  // socket is admitting the board is stale, not erasing it. So this
  // short-circuits the wait and changes nothing else.
  //
  // The attempt counter is left alone too: if this dial fails, the
  // close handler schedules the next rung from where the ladder
  // already was, so a player leaning on the button cannot reset the
  // backoff into a tight dial loop against a server that is still down.
  //
  // A no-op while a socket exists — one is already open or dialling,
  // and stacking a second on top of it is the bug the isCurrent()
  // guard in open() exists to survive.
  retryNow(): void {
    if (this.socket) return;
    this.cancelReconnect();
    this.status.set("reconnecting");
    this.append("info", "manual reconnect requested");
    this.open();
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
    this.resetReconnectAttempts();
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

  // failSend is what every send path does when the socket is not open.
  //
  // It used to be one `append("error", "not connected")` per method,
  // and the `log` store that line lands in is read by exactly one
  // thing in the whole app: the bug-report modal. So an action
  // attempted during a deploy produced a line nobody would ever see
  // and a silent no-op — indistinguishable from the board freezing,
  // and duly reported as one (#519). Routing it through lastError as
  // well puts it in the same toast every server rejection already
  // uses. The log line stays: it is what triage reads afterwards.
  private failSend(what: string): void {
    this.append("error", `not connected: ${what} not sent`);
    this.raiseError(OFFLINE_ERROR_CODE, offlineSendMessage(what));
  }

  // raiseError publishes to the toast store and (re)starts the
  // auto-clear timer. Shared by the server `error` frame handler and
  // the client-side offline path above so both age out identically —
  // a fresh error always gets its full TTL, replacing whatever was
  // still on screen.
  private raiseError(
    code: string,
    message: string,
    extra: { missing?: string[]; cardID?: string; reason?: string } = {},
  ): void {
    this.lastError.set({ code, message, at: new Date(), ...extra });
    if (this.errorClearTimer !== null) {
      clearTimeout(this.errorClearTimer);
    }
    this.errorClearTimer = setTimeout(() => {
      this.lastError.set(null);
      this.errorClearTimer = null;
    }, ERROR_TOAST_TTL_MS);
  }

  sendPing(msg: string): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      this.failSend("ping");
      return;
    }
    const frame: Frame<PingPayload> = {
      v: PROTOCOL_VERSION,
      kind: "ping",
      id: uuid(),
      payload: { msg },
    };
    const wire = JSON.stringify(frame);
    this.socket.send(wire);
    this.recordFrame("out", frame, wire);
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
      this.failSend("chat message");
      return null;
    }
    const id = uuid();
    const frame: Frame<ChatPayload> = {
      v: PROTOCOL_VERSION,
      kind: "chat",
      id,
      payload: { author_name: "", text: trimmed, timestamp: "" },
    };
    const wire = JSON.stringify(frame);
    this.socket.send(wire);
    this.recordFrame("out", frame, wire);
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
      // Every card click, every priority pass, every mana tap funnels
      // through here. This branch is the one a player meets during a
      // deploy, so it is the one that has to speak (#519).
      this.failSend(`action "${type}"`);
      return null;
    }
    const id = uuid();
    const frame: Frame = {
      v: PROTOCOL_VERSION,
      kind: "action",
      id,
      payload: { type, player, params },
    };
    const wire = JSON.stringify(frame);
    this.socket.send(wire);
    this.sentActions++;
    this.recordFrame("out", frame, wire);
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
    // Recorded before the version gate so a v-mismatch — the exact
    // thing you would open the inspector to diagnose — is visible.
    this.recordFrame("in", frame, raw);
    if (frame.v !== PROTOCOL_VERSION) {
      this.append("error", `server spoke v${frame.v}, expected v${PROTOCOL_VERSION}`);
      return;
    }
    // #266: belt to guardedWritable's braces. Nothing below is allowed
    // to escape into the socket's message listener — an uncaught throw
    // there is invisible (no console in this app) and leaves the
    // connection alive but the client mid-update. Record it instead so
    // it reaches the next bug report, and keep serving later frames:
    // snapshots are full states, so the frame after a bad one heals
    // the board on its own.
    try {
      this.dispatchFrame(frame);
    } catch (err) {
      this.append("error", `frame handling failed: ${describeThrown(err)}`);
      recordClientError(`ws frame handling failed: ${describeThrown(err)}`);
    }
  }

  private dispatchFrame(frame: Frame): void {
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
        // Log the frame BEFORE applying it. #266: with the log line
        // last, a frame whose application throws is never logged, so
        // the attached log ends at the last frame that worked and the
        // culprit is invisible. Logging on receipt means the final
        // line in a freeze report names the frame that broke it.
        // Optional-chained because a malformed `turn` must not take
        // out the log line that would have named the malformed frame.
        this.append(
          "received",
          `snapshot seq=${p.seq} turn=${p.game?.turn?.number} seat=${p.game?.turn?.active_seat} step=${p.game?.turn?.step}`,
        );
        this.highestSeq = p.seq;
        this.snapshot.set(p.game);
        this.lastSeq.set(p.seq);
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
          kind: p.kind ?? "say",
          reason: p.reason ?? "",
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
        // rejected. raiseError resets the auto-clear timer so a fresh
        // error gets its full TTL even if a previous one is still
        // showing.
        this.raiseError(code, message, {
          missing: p?.missing,
          cardID: p?.card_id,
          reason: p?.reason,
        });
        break;
      }
      default:
        this.append("error", `unknown kind from server: ${frame.kind}`);
    }
  }

  // append records one protocol-log entry. Every entry is redacted on
  // the way in (#721): this buffer is what a bug report attaches, and
  // some of what lands here is not ours to vouch for — chat text other
  // players typed, server error messages, a URL someone logs next year.
  private append(direction: LogEntry["direction"], text: string): void {
    const entry: LogEntry = {
      id: uuid(),
      at: new Date(),
      direction,
      text: redactSecrets(text),
    };
    this.log.update((entries) => [...entries.slice(-(WS_LOG_LIMIT - 1)), entry]);
  }
}
