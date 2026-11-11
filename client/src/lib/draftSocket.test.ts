// #5438 (ADR 0222 §5): the draft socket. One frame in (`draft`), one out
// (`draft_pick`), and the game socket's reconnect convention: retry on
// the ladder, stop for good on close 1000 ("draft closed"), and ask over
// HTTP whether the draft is still there after dials that failed.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { DraftView } from "./draft";
import {
  DRAFT_CLOSED_REASON,
  DRAFT_HEARTBEAT_INTERVAL_MS,
  DRAFT_HEARTBEAT_TIMEOUT_MS,
  DraftSocket,
  type DraftError,
  type DraftGoneCheck,
  type DraftSocketStatus,
} from "./draftSocket";
import { DEAD_SESSION_CHECK_AFTER_FAILURES, RECONNECT_CAP_MS } from "./ws";

const URL = "wss://c2.example/draft/d1/ws?token=seat";

class FakeSocket {
  static opened: FakeSocket[] = [];
  readyState = 1;
  sent: string[] = [];
  closeCode: number | null = null;
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};
  constructor(readonly url: string) {
    FakeSocket.opened.push(this);
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(type: string, fn: (ev: unknown) => void): void {
    this.listeners[type] = (this.listeners[type] ?? []).filter((f) => f !== fn);
  }
  send(data: string): void {
    this.sent.push(data);
  }
  close(code = 1000): void {
    this.readyState = 3;
    this.closeCode = code;
  }
  emit(type: string, ev: unknown): void {
    for (const fn of [...(this.listeners[type] ?? [])]) fn(ev);
  }
  frame(kind: string, payload: unknown, id = "srv"): void {
    this.emit("message", { data: JSON.stringify({ v: 0, kind, id, payload }) });
  }
  drop(code = 1006, reason = ""): void {
    this.readyState = 3;
    this.emit("close", { code, reason });
  }
  frames(): { kind: string; id: string; payload: unknown; v: number }[] {
    return this.sent.map((s) => JSON.parse(s));
  }
}

const last = (): FakeSocket => FakeSocket.opened[FakeSocket.opened.length - 1];
const view = (seq: number, over: Partial<DraftView> = {}): DraftView =>
  ({
    id: "d1",
    name: "n",
    set: "blb",
    mode: "pick_one",
    seats: 2,
    state: "lobby",
    seq,
    you: 0,
    players: [],
    ...over,
  }) as DraftView;

function open(check?: DraftGoneCheck) {
  const views: DraftView[] = [];
  const statuses: DraftSocketStatus[] = [];
  const errors: DraftError[] = [];
  const sock = new DraftSocket({
    url: URL,
    onView: (v) => views.push(v),
    onStatus: (s) => statuses.push(s),
    onError: (e) => errors.push(e),
    check,
    socket: FakeSocket as never,
    rand: () => 1,
  });
  sock.connect();
  return { sock, views, statuses, errors };
}

beforeEach(() => {
  FakeSocket.opened = [];
  vi.useFakeTimers();
});
afterEach(() => {
  vi.useRealTimers();
});

describe("the draft socket", () => {
  it("dials the url it was given and hands every draft frame on", () => {
    const { views, statuses } = open();
    expect(FakeSocket.opened.map((s) => s.url)).toEqual([URL]);
    last().frame("draft", view(1));
    last().frame("draft", view(2, { state: "picking" }));
    expect(views.map((v) => v.seq)).toEqual([1, 2]);
    expect(statuses).toEqual(["connected"]);
  });

  it("never rolls the view back: seq only rises", () => {
    const { views } = open();
    last().frame("draft", view(5));
    last().frame("draft", view(3));
    last().frame("draft", view(5));
    expect(views.map((v) => v.seq)).toEqual([5, 5]);
  });

  it("sends a pick in the envelope, with the chosen ids", () => {
    const { sock } = open();
    last().frame("draft", view(1));
    const id = sock.pick(["c41", "c7"]);
    const sent = last()
      .frames()
      .filter((f) => f.kind === "draft_pick");
    expect(sent).toEqual([{ v: 0, kind: "draft_pick", id, payload: { cards: ["c41", "c7"] } }]);
    expect(id).toBeTruthy();
  });

  it("sends nothing, and says so, when the socket is not open", () => {
    const { sock } = open();
    last().readyState = 0;
    expect(sock.pick(["c1"])).toBeNull();
    expect(last().sent).toEqual([]);
    expect(sock.isOpen()).toBe(false);
  });

  it("hands a refusal on with the id it echoes", () => {
    const { errors } = open();
    last().frame("error", { code: "draft_pick_refused", message: "already picked" }, "f-1");
    expect(errors).toEqual([{ id: "f-1", code: "draft_pick_refused", message: "already picked" }]);
  });

  it("ignores frames it does not know, and junk", () => {
    const { views, errors } = open();
    last().frame("pong", { msg: "heartbeat" });
    last().frame("something_new", {});
    last().emit("message", { data: "{not json" });
    last().emit("message", { data: new ArrayBuffer(1) });
    expect(views).toEqual([]);
    expect(errors).toEqual([]);
  });

  it("does not redial after close 1000, draft closed", () => {
    const { statuses } = open();
    last().frame("draft", view(1));
    last().drop(1000, DRAFT_CLOSED_REASON);
    vi.advanceTimersByTime(RECONNECT_CAP_MS * 4);
    expect(FakeSocket.opened).toHaveLength(1);
    expect(statuses).toEqual(["connected", "closed"]);
  });

  it("redials a dropped socket on the ladder and picks up where it was", () => {
    const { views, statuses } = open();
    last().frame("draft", view(4));
    last().drop(1006);
    expect(statuses).toEqual(["connected", "reconnecting"]);
    expect(FakeSocket.opened).toHaveLength(1);
    vi.advanceTimersByTime(500);
    expect(FakeSocket.opened).toHaveLength(2);
    expect(last().url).toBe(URL);
    last().frame("draft", view(4));
    expect(statuses).toEqual(["connected", "reconnecting", "connected"]);
    expect(views.map((v) => v.seq)).toEqual([4, 4]);
  });

  it("treats a deploy's 1001 as a drop, not an ending", () => {
    open();
    last().frame("draft", view(1));
    last().drop(1001, "server shutting down");
    vi.advanceTimersByTime(500);
    expect(FakeSocket.opened).toHaveLength(2);
  });

  it("asks at once about a seat that was never let in, and stops when it is gone", async () => {
    const check = vi.fn<DraftGoneCheck>(async () => "gone");
    const { statuses } = open(check);
    last().drop(1006);
    expect(check).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(RECONNECT_CAP_MS * 4);
    expect(statuses).toEqual(["gone"]);
    expect(FakeSocket.opened).toHaveLength(1);
  });

  it("keeps dialling when the check cannot tell, or says the draft is there", async () => {
    const answers: ("here" | "unknown")[] = ["unknown", "here"];
    const check = vi.fn<DraftGoneCheck>(async () => answers.shift() ?? "here");
    const { statuses } = open(check);
    last().drop(1006);
    await vi.advanceTimersByTimeAsync(500);
    expect(FakeSocket.opened).toHaveLength(2);
    last().drop(1006);
    await vi.advanceTimersByTimeAsync(1000);
    expect(FakeSocket.opened).toHaveLength(3);
    expect(statuses).not.toContain("gone");
  });

  it("gives a seat that was connected the game socket's patience before asking", async () => {
    const check = vi.fn<DraftGoneCheck>(async () => "gone");
    const { statuses } = open(check);
    last().frame("draft", view(1));
    // The drop of a working connection is not a failed dial.
    last().drop(1006);
    await vi.advanceTimersByTimeAsync(500);
    // Then dials that never get a frame: one short of the threshold.
    for (let i = 1; i < DEAD_SESSION_CHECK_AFTER_FAILURES; i++) {
      last().drop(1006);
      await vi.advanceTimersByTimeAsync(2000);
      expect(check).not.toHaveBeenCalled();
    }
    last().drop(1006);
    await vi.advanceTimersByTimeAsync(0);
    expect(check).toHaveBeenCalledTimes(1);
    expect(statuses[statuses.length - 1]).toBe("gone");
  });

  it("sends a heartbeat, and redials a socket that has gone quiet", () => {
    const { statuses } = open();
    last().frame("draft", view(1));
    vi.advanceTimersByTime(DRAFT_HEARTBEAT_INTERVAL_MS);
    expect(
      last()
        .frames()
        .map((f) => f.kind),
    ).toEqual(["ping"]);
    const quiet = last();
    vi.advanceTimersByTime(DRAFT_HEARTBEAT_TIMEOUT_MS + DRAFT_HEARTBEAT_INTERVAL_MS);
    expect(quiet.closeCode).not.toBeNull();
    expect(quiet.closeCode).not.toBe(1000);
    expect(statuses).toContain("reconnecting");
    vi.advanceTimersByTime(RECONNECT_CAP_MS);
    expect(FakeSocket.opened.length).toBeGreaterThan(1);
  });

  it("stays up while frames keep coming", () => {
    open();
    for (let i = 0; i < 10; i++) {
      vi.advanceTimersByTime(DRAFT_HEARTBEAT_INTERVAL_MS);
      last().frame("pong", { msg: "heartbeat" });
    }
    expect(FakeSocket.opened).toHaveLength(1);
  });

  it("hangs up for good on disconnect, with no redial and no status", () => {
    const { sock, statuses } = open();
    last().frame("draft", view(1));
    const s = last();
    sock.disconnect();
    expect(s.closeCode).toBe(1000);
    s.drop(1006);
    vi.advanceTimersByTime(RECONNECT_CAP_MS * 4);
    expect(FakeSocket.opened).toHaveLength(1);
    expect(statuses).toEqual(["connected"]);
  });

  it("sends a deck as a draft_deck frame and returns its id (#5440)", () => {
    const { sock } = open();
    last().emit("open", {});
    const id = sock.deck(["c1", "c2"], { Forest: 9 });
    const frames = last().frames();
    expect(frames[frames.length - 1]).toMatchObject({
      kind: "draft_deck",
      id,
      payload: { main: ["c1", "c2"], basics: { Forest: 9 } },
    });
    sock.disconnect();
    expect(sock.deck(["c1"], {})).toBeNull();
  });
});
