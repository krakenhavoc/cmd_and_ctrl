// ADR 0119 amendment 2026-10-16 §D — the arrival trace, as pure
// functions: what counts as a new item (never the first frame, a
// reconnect or an undo), the burst's cap and stagger, the three gates,
// the curve, and the head and tail it is drawn with.

import { describe, expect, it } from "vitest";

import {
  ARRIVAL_MAX,
  ARRIVAL_MS,
  ARRIVAL_STAGGER_MS,
  arrivalAllowed,
  arrivalCurve,
  arrivalDuration,
  curveLength,
  emptyArrivalTracker,
  startShows,
  tracePlan,
  trackArrivals,
  TRACE_TAIL_MAX,
  TRACE_TAIL_MIN,
  type ArrivalItem,
  type ArrivalTracker,
} from "./stackArrival";
import { BOARD_BOW } from "./stackArrows";

const spell = (id: string, controller = "b"): ArrivalItem => ({
  id,
  kind: "spell",
  controller,
  source_card_id: id,
});
const trigger = (id: string, source: string, controller = "a"): ArrivalItem => ({
  id,
  kind: "triggered",
  controller,
  source_card_id: source,
});
const log = (top: number) => Array.from({ length: top }, (_, i) => ({ seq: i + 1 }));

function primed(items: ArrivalItem[] = [], seq = 1): ArrivalTracker {
  const res = trackArrivals(emptyArrivalTracker(), { stackItems: items, log: log(seq) });
  expect(res.arrivals).toEqual([]);
  return res.tracker;
}

describe("trackArrivals", () => {
  it("traces nothing on the first frame, whatever is already on the stack", () => {
    const res = trackArrivals(emptyArrivalTracker(), {
      stackItems: [spell("bolt"), trigger("t1", "blood-artist")],
      log: log(5),
    });
    expect(res.arrivals).toEqual([]);
    expect([...res.tracker.ids]).toEqual(["bolt", "t1"]);
  });

  it("traces a spell new on the stack from its caster, and nothing for one already there", () => {
    const res = trackArrivals(primed([spell("bolt")]), {
      stackItems: [spell("bolt"), spell("counter", "a")],
      log: log(2),
    });
    expect(res.arrivals).toEqual([
      { itemID: "counter", casterSeat: "a", sourceCardID: null, delayMs: 0 },
    ]);
  });

  it("traces an ability from its source permanent", () => {
    const res = trackArrivals(primed(), {
      stackItems: [trigger("t1", "purphoros")],
      log: log(2),
    });
    expect(res.arrivals).toEqual([
      { itemID: "t1", casterSeat: "a", sourceCardID: "purphoros", delayMs: 0 },
    ]);
  });

  it("traces nothing on a reconnect snapshot, even with new items in it", () => {
    const res = trackArrivals(
      primed([spell("bolt")]),
      { stackItems: [spell("bolt"), spell("missed-1"), trigger("missed-2", "x")], log: log(9) },
      { reprime: true },
    );
    expect(res.arrivals).toEqual([]);
    // …and the next frame traces only what is new after it.
    const next = trackArrivals(res.tracker, {
      stackItems: [spell("bolt"), spell("missed-1"), trigger("missed-2", "x"), spell("new")],
      log: log(10),
    });
    expect(next.arrivals.map((a) => a.itemID)).toEqual(["new"]);
  });

  it("traces nothing on an undo, which brings items back as the log rewinds", () => {
    const t = primed([], 8);
    const res = trackArrivals(t, { stackItems: [spell("bolt")], log: log(6) });
    expect(res.arrivals).toEqual([]);
  });

  it("staggers a burst in the order it went on, bottom first", () => {
    const res = trackArrivals(primed(), {
      stackItems: [trigger("t1", "s"), trigger("t2", "s"), trigger("t3", "s")],
      log: log(3),
    });
    expect(res.arrivals.map((a) => [a.itemID, a.delayMs])).toEqual([
      ["t1", 0],
      ["t2", ARRIVAL_STAGGER_MS],
      ["t3", 2 * ARRIVAL_STAGGER_MS],
    ]);
  });

  it("runs at most four of a burst over four, the newest; the rest arrive without one", () => {
    const burst = Array.from({ length: 7 }, (_, i) => trigger(`t${i + 1}`, "blood-artist"));
    const res = trackArrivals(primed(), { stackItems: burst, log: log(4) });
    expect(ARRIVAL_MAX).toBe(4);
    expect(res.arrivals.map((a) => a.itemID)).toEqual(["t4", "t5", "t6", "t7"]);
    expect(res.arrivals.map((a) => a.delayMs)).toEqual([0, 80, 160, 240]);
  });

  it("does not mutate the tracker it was given", () => {
    const t = primed([spell("bolt")]);
    trackArrivals(t, { stackItems: [spell("bolt"), spell("x")], log: log(2) });
    expect([...t.ids]).toEqual(["bolt"]);
  });

  it("treats an absent stack as empty", () => {
    const res = trackArrivals(primed([spell("bolt")]), { stackItems: undefined, log: log(2) });
    expect(res.arrivals).toEqual([]);
    expect(res.tracker.ids.size).toBe(0);
  });
});

describe("arrivalAllowed", () => {
  const s = (enabled: boolean, stackArrival: boolean, reduceMotion: boolean) => ({
    animations: { enabled, stackArrival },
    accessibility: { reduceMotion },
  });

  it("runs only with the master switch and its own toggle on and no reduced motion", () => {
    expect(arrivalAllowed(s(true, true, false))).toBe(true);
    expect(arrivalAllowed(s(false, true, false))).toBe(false);
    expect(arrivalAllowed(s(true, false, false))).toBe(false);
    expect(arrivalAllowed(s(true, true, true))).toBe(false);
  });
});

describe("arrivalDuration", () => {
  it("scales the run with animations.speed, and a bad speed counts as 1", () => {
    expect(ARRIVAL_MS).toBe(320);
    expect(arrivalDuration(1)).toBe(320);
    expect(arrivalDuration(2)).toBe(640);
    expect(arrivalDuration(0.5)).toBe(160);
    expect(arrivalDuration(0)).toBe(320);
    expect(arrivalDuration(Number.NaN)).toBe(320);
  });
});

describe("arrivalCurve", () => {
  const board = { width: 1400, height: 900 };

  it("leaves the source's edge, ends just short of the card, and bows toward the centre", () => {
    const avatar = { left: 1200, top: 60, width: 60, height: 60 };
    const card = { left: 20, top: 300, width: 200, height: 280 };
    const c = arrivalCurve(avatar, card, board);
    // Starts on the avatar's rim, facing the card (its left or bottom edge).
    expect(c.x1).toBeLessThanOrEqual(1200 + 60);
    expect(c.x1).toBeGreaterThanOrEqual(1200);
    // Ends outside the card, by its edge.
    expect(c.x2).toBeGreaterThan(220);
    expect(c.x2).toBeLessThan(230);
    // Bows toward the board's centre (below the straight line here).
    const mid = { x: (c.x1 + c.x2) / 2, y: (c.y1 + c.y2) / 2 };
    expect(c.cy).toBeGreaterThan(mid.y);
    const chord = Math.hypot(c.x2 - c.x1, c.y2 - c.y1);
    expect(Math.hypot(c.cx - mid.x, c.cy - mid.y)).toBeCloseTo(chord * BOARD_BOW, 3);
  });

  it("bows away from the stack panel when the near side would cross it", () => {
    const source = { left: 1000, top: 100, width: 60, height: 60 };
    const card = { left: 20, top: 500, width: 200, height: 280 };
    const free = arrivalCurve(source, card, board);
    // A panel over the near side's control point, not the far side's.
    const lane = { left: 300, top: 300, width: 400, height: 400 };
    const inLane = (x: number, y: number) => x >= 300 && x <= 700 && y >= 300 && y <= 700;
    expect(inLane(free.cx, free.cy)).toBe(true);
    const kept = arrivalCurve(source, card, board, lane);
    expect(inLane(kept.cx, kept.cy)).toBe(false);
    expect(kept.x1).toBe(free.x1);
    expect(kept.x2).toBe(free.x2);
  });
});

describe("startShows", () => {
  const board = { width: 1400, height: 900 };
  it("needs a start on the board, with a size, and not under the stack panel", () => {
    expect(startShows({ left: 1200, top: 40, width: 60, height: 60 }, board)).toBe(true);
    expect(startShows({ left: 1500, top: 40, width: 60, height: 60 }, board)).toBe(false);
    expect(startShows({ left: 100, top: 40, width: 0, height: 0 }, board)).toBe(false);
    // A two-seat table: the pile docks over the opponent's avatar.
    const pile = { left: 12, top: 150, width: 280, height: 540 };
    expect(startShows({ left: 30, top: 310, width: 84, height: 84 }, board, pile)).toBe(false);
    expect(startShows({ left: 1200, top: 40, width: 60, height: 60 }, board, pile)).toBe(true);
  });
});

describe("curveLength", () => {
  it("is the chord for a straight curve and longer for a bowed one", () => {
    expect(curveLength({ x1: 0, y1: 0, cx: 50, cy: 0, x2: 100, y2: 0 })).toBeCloseTo(100, 6);
    expect(curveLength({ x1: 0, y1: 0, cx: 50, cy: 40, x2: 100, y2: 0 })).toBeGreaterThan(100);
  });
});

describe("tracePlan", () => {
  it("is a bright head and a fading tail of the arrows' 6px dashes", () => {
    const p = tracePlan(480);
    expect(p.segments[0]).toMatchObject({ head: true, back: 0, opacity: 1 });
    const tail = p.segments.slice(1);
    expect(tail.length).toBe(TRACE_TAIL_MAX);
    for (const s of tail) expect(s.len).toBe(6);
    // Each dash 11px behind the last, fading as it goes.
    for (let i = 1; i < tail.length; i++) {
      expect(tail[i].back - tail[i - 1].back).toBe(11);
      expect(tail[i].opacity).toBeLessThan(tail[i - 1].opacity);
    }
  });

  it("runs the front past the end by the tail, so it all drains into the card", () => {
    const p = tracePlan(300);
    const last = p.segments[p.segments.length - 1];
    expect(p.travel).toBeGreaterThanOrEqual(300 + last.back + last.len);
  });

  it("keeps a short trace's tail short, and never shorter than two dashes", () => {
    expect(tracePlan(40).segments.length - 1).toBe(TRACE_TAIL_MIN);
    expect(tracePlan(0).travel).toBeGreaterThan(0);
  });
});
