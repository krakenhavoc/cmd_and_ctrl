// ADR 0119 §3 — the linger on resolution, as pure functions: matching
// log entries to departed items, no entry no linger, priming, the
// burst schedule and its cap, the destination lookup, and placement.

import { describe, expect, it } from "vitest";

import {
  LINGER_BURST_MIN_MS,
  LINGER_GAP,
  LINGER_MS,
  LINGER_QUEUE_MAX,
  LingerQueue,
  OUTCOME_BADGE,
  counteredBy,
  destinationSelectors,
  emptyLingerTracker,
  enqueueLingers,
  entryItemID,
  flightAllowed,
  flightTransform,
  lingerAnnouncement,
  lingerDestination,
  lingerDwellMs,
  lingerShape,
  pileSelector,
  placeLinger,
  trackLinger,
  type Box,
  type Departure,
  type LingerItem,
  type LingerTimers,
  type LingerTracker,
  type QueuedLinger,
} from "./stackLinger";
import type { CardView, GameView, LogEvent } from "./protocol";

const item = (id: string, extra: Partial<LingerItem> = {}): LingerItem => ({
  id,
  kind: "spell",
  name: id,
  image: `/img/${id}`,
  thumb: `/thumb/${id}`,
  casterColor: "#f00",
  isCopy: false,
  ...extra,
});

const resolve = (seq: number, extra: Partial<LogEvent> = {}): LogEvent => ({
  seq,
  kind: "resolve",
  seat: 0,
  text: "x resolved",
  ...extra,
});

const box = (left: number, top: number, width = 100, height = 140): Box => ({
  left,
  top,
  width,
  height,
});

// A tracker that has seen `items` and the log up to `seq`.
function seen(items: LingerItem[], seq: number): LingerTracker {
  const primed = trackLinger(emptyLingerTracker(), [resolve(seq, { stack_item_id: "old" })], items);
  expect(primed.primed).toBe(true);
  return primed.tracker;
}

describe("entryItemID", () => {
  it("reads stack_item_id first, for spells and abilities alike", () => {
    expect(entryItemID(resolve(1, { stack_item_id: "item-1", card_id: "source" }))).toBe("item-1");
    expect(
      entryItemID({ seq: 1, kind: "fizzle", seat: 0, text: "", stack_item_id: "item-2" }),
    ).toBe("item-2");
  });

  it("falls back to a spell's card_id, and to a counter's target", () => {
    expect(entryItemID(resolve(1, { card_id: "bolt" }))).toBe("bolt");
    expect(
      entryItemID({ seq: 1, kind: "counter", seat: 1, card_id: "cs", target: "bolt", text: "" }),
    ).toBe("bolt");
  });

  it("is null for every other kind", () => {
    expect(entryItemID({ seq: 1, kind: "cast", seat: 0, card_id: "bolt", text: "" })).toBeNull();
    expect(entryItemID({ seq: 1, kind: "zone", seat: 0, card_id: "bolt", text: "" })).toBeNull();
  });
});

describe("counteredBy", () => {
  it("names the counter from the server's own line", () => {
    expect(
      counteredBy({
        seq: 1,
        kind: "counter",
        seat: 1,
        text: "Counterspell countered Lightning Bolt",
      }),
    ).toBe("Counterspell");
    expect(counteredBy(resolve(1))).toBeNull();
  });
});

describe("trackLinger", () => {
  it("primes on the first frame: nothing lingers or fades", () => {
    const r = trackLinger(emptyLingerTracker(), [resolve(5, { stack_item_id: "bolt" })], [], {
      rects: new Map([["bolt", box(0, 0)]]),
    });
    expect(r.primed).toBe(true);
    expect(r.lingers).toEqual([]);
    expect(r.fades).toEqual([]);
  });

  it("lingers an item a new resolve entry names, with where it was drawn", () => {
    const t = seen([item("bolt")], 10);
    const r = trackLinger(t, [resolve(11, { stack_item_id: "bolt" })], [], {
      rects: new Map([["bolt", box(12, 300)]]),
    });
    expect(r.primed).toBe(false);
    expect(r.lingers).toHaveLength(1);
    expect(r.lingers[0]).toMatchObject({ outcome: "resolved", seq: 11, rect: box(12, 300) });
    expect(r.fades).toEqual([]);
  });

  it("matches a countered spell by the counter's target and a fizzle by card_id", () => {
    const t = seen([item("bolt"), item("cs"), item("growth")], 10);
    const r = trackLinger(
      t,
      [
        {
          seq: 11,
          kind: "counter",
          seat: 1,
          card_id: "cs",
          target: "bolt",
          text: "Counterspell countered Lightning Bolt",
        },
        resolve(12, { card_id: "cs" }),
        { seq: 13, kind: "fizzle", seat: 1, card_id: "growth", text: "" },
      ],
      [],
    );
    expect(r.lingers.map((d) => [d.item.id, d.outcome, d.by])).toEqual([
      ["bolt", "countered", "Counterspell"],
      ["cs", "resolved", null],
      ["growth", "fizzled", null],
    ]);
  });

  it("matches an ability by stack_item_id, never by its source's card_id", () => {
    const trig = item("trig-1", { kind: "triggered" });
    const t = seen([trig], 10);
    const named = trackLinger(t, [resolve(11, { stack_item_id: "trig-1", card_id: "vivi" })], []);
    expect(named.lingers.map((d) => d.item.id)).toEqual(["trig-1"]);
    // An old server: the entry names only the source permanent.
    const unnamed = trackLinger(t, [resolve(11, { card_id: "vivi", label: "Vivi — x" })], []);
    expect(unnamed.lingers).toEqual([]);
    expect(unnamed.fades.map((d) => d.item.id)).toEqual(["trig-1"]);
  });

  it("gives an item that left with no entry a fade and no badge", () => {
    const t = seen([item("bolt")], 10);
    const r = trackLinger(t, [resolve(10, { stack_item_id: "old" })], []);
    expect(r.lingers).toEqual([]);
    expect(r.fades).toHaveLength(1);
    expect(r.fades[0]).toMatchObject({ outcome: null, by: null, seq: -1 });
  });

  it("ignores an entry the client already handled", () => {
    const t = seen([item("bolt")], 10);
    // Seq 9 is under the watermark: it is old news, not this departure.
    const r = trackLinger(t, [resolve(9, { stack_item_id: "bolt" }), resolve(10)], []);
    expect(r.lingers).toEqual([]);
    expect(r.fades).toHaveLength(1);
  });

  it("does nothing for an item still on the stack", () => {
    const t = seen([item("bolt"), item("cs")], 10);
    const r = trackLinger(t, [resolve(11, { stack_item_id: "cs" })], [item("bolt")]);
    expect(r.lingers.map((d) => d.item.id)).toEqual(["cs"]);
    expect(r.fades).toEqual([]);
  });

  it("re-primes on request (a reconnect, a replay toggle) without lingering", () => {
    const t = seen([item("bolt")], 10);
    const r = trackLinger(t, [resolve(11, { stack_item_id: "bolt" })], [], { reprime: true });
    expect(r.primed).toBe(true);
    expect(r.lingers).toEqual([]);
    expect(r.fades).toEqual([]);
    // And the next frame starts from there.
    const next = trackLinger(r.tracker, [resolve(11, { stack_item_id: "bolt" })], []);
    expect(next.lingers).toEqual([]);
  });

  it("treats an undo (a lower log) as no entry: the item fades", () => {
    const t = seen([item("bolt")], 10);
    const r = trackLinger(t, [resolve(8, { stack_item_id: "bolt" })], []);
    expect(r.primed).toBe(false);
    expect(r.lingers).toEqual([]);
    expect(r.fades.map((d) => d.item.id)).toEqual(["bolt"]);
  });

  it("keeps the frame's items for the next frame", () => {
    const t = seen([], 10);
    const r = trackLinger(t, [resolve(10)], [item("bolt")]);
    expect([...r.tracker.items.keys()]).toEqual(["bolt"]);
  });
});

describe("burst scheduling", () => {
  const dep = (id: string): Departure => ({
    item: item(id),
    outcome: "resolved",
    by: null,
    seq: 1,
    rect: box(0, 0),
  });

  it("shares 1.5 s among a burst, never under 400 ms", () => {
    expect(lingerDwellMs(1)).toBe(LINGER_MS);
    expect(lingerDwellMs(2)).toBe(750);
    expect(lingerDwellMs(3)).toBe(500);
    expect(lingerDwellMs(5)).toBe(LINGER_BURST_MIN_MS);
  });

  it("queues at most three, counting the one on screen, dropping the oldest", () => {
    const q = enqueueLingers([], [dep("a"), dep("b"), dep("c"), dep("d"), dep("e")], false);
    expect(q.map((l) => l.item.id)).toEqual(["c", "d", "e"]);
    expect(q.every((l) => l.dwellMs === LINGER_BURST_MIN_MS)).toBe(true);
    const behind = enqueueLingers(q.slice(0, 1), [dep("f"), dep("g")], true);
    expect(behind.map((l) => l.item.id)).toEqual(["f", "g"]);
    expect(behind).toHaveLength(LINGER_QUEUE_MAX - 1);
  });

  it("shows one at a time in resolution order, each for its dwell", () => {
    let now = 0;
    const pending: Array<{ at: number; fn: () => void }> = [];
    const timers: LingerTimers = {
      set: (fn, ms) => {
        const h = { at: now + ms, fn };
        pending.push(h);
        return h;
      },
      clear: (h) => {
        const i = pending.indexOf(h as (typeof pending)[number]);
        if (i >= 0) pending.splice(i, 1);
      },
    };
    const advance = (ms: number) => {
      now += ms;
      for (;;) {
        pending.sort((a, b) => a.at - b.at);
        const due = pending[0];
        if (!due || due.at > now) return;
        pending.shift();
        due.fn();
      }
    };
    const log: string[] = [];
    const q = new LingerQueue(
      (l: QueuedLinger) => log.push(`show ${l.item.id}@${now}`),
      (l: QueuedLinger) => log.push(`end ${l.item.id}@${now}`),
      timers,
    );
    q.push([dep("a"), dep("b")]);
    expect(q.current?.item.id).toBe("a");
    expect(q.waiting).toBe(1);
    advance(750);
    advance(750);
    expect(log).toEqual(["show a@0", "end a@750", "show b@750", "end b@1500"]);
    expect(q.current).toBeNull();

    // A re-prime drops everything without ending it.
    q.push([dep("c")]);
    q.clear();
    advance(5000);
    expect(log.at(-1)).toBe("show c@1500");
  });
});

describe("badges and the announcer", () => {
  it("says resolved, countered by, fizzled", () => {
    const bolt = item("bolt", { name: "Lightning Bolt" });
    expect(lingerAnnouncement({ item: bolt, outcome: "resolved", by: null })).toBe(
      "Lightning Bolt resolved",
    );
    expect(lingerAnnouncement({ item: bolt, outcome: "countered", by: "Counterspell" })).toBe(
      "Lightning Bolt was countered by Counterspell",
    );
    expect(lingerAnnouncement({ item: bolt, outcome: "fizzled", by: null })).toBe(
      "Lightning Bolt fizzled: no legal targets",
    );
    expect(lingerAnnouncement({ item: bolt, outcome: null, by: null })).toBe("");
  });

  it("titles a fizzle with the rule", () => {
    expect(OUTCOME_BADGE.fizzled.title).toBe(
      "countered on resolution: no legal targets (CR 608.2b)",
    );
    expect(OUTCOME_BADGE.resolved.label).toBe("Resolved");
    expect(OUTCOME_BADGE.countered.label).toBe("Countered");
  });
});

describe("lingerDestination", () => {
  const card = (id: string, owner: string, controller = owner): CardView =>
    ({ instance_id: id, name: id, owner, controller }) as CardView;
  const zone = (cards: CardView[] = []) => ({ kind: "z", count: cards.length, cards });
  const view = (opts: {
    field?: CardView[];
    grave?: CardView[];
    exile?: CardView[];
    command?: CardView[];
  }): GameView =>
    ({
      seats: [
        { id: "me", graveyard: zone(), command: zone() },
        { id: "opp", graveyard: zone(opts.grave), command: zone(opts.command) },
      ],
      battlefield: zone(opts.field),
      exile: zone(opts.exile),
    }) as unknown as GameView;

  it("finds a permanent spell on the battlefield (CR 608.3a)", () => {
    const d = lingerDestination(view({ field: [card("bears", "me", "opp")] }), item("bears"));
    expect(d).toEqual({ zone: "battlefield", cardID: "bears", seat: "opp" });
    expect(destinationSelectors(d!)).toEqual(['[data-instance-id="bears"]']);
  });

  it("sends an instant to its owner's graveyard pile (CR 608.2n, 701.6a, 608.2b)", () => {
    const d = lingerDestination(view({ grave: [card("bolt", "opp")] }), item("bolt"));
    expect(d).toEqual({ zone: "graveyard", cardID: "bolt", seat: "opp" });
    expect(destinationSelectors(d!)).toEqual([pileSelector("graveyard", "opp")]);
    expect(pileSelector("graveyard", "opp")).toBe('[data-pile="graveyard"][data-pile-owner="opp"]');
  });

  it("finds exile (flashback) and the command zone, card first then pile", () => {
    const ex = lingerDestination(view({ exile: [card("fb", "me")] }), item("fb"));
    expect(ex).toEqual({ zone: "exile", cardID: "fb", seat: "me" });
    expect(destinationSelectors(ex!)).toEqual([
      '[data-instance-id="fb"]',
      pileSelector("exile", "me"),
    ]);
    const cmd = lingerDestination(view({ command: [card("cmdr", "opp")] }), item("cmdr"));
    expect(cmd).toEqual({ zone: "command", cardID: "cmdr", seat: "opp" });
  });

  it("is null for an ability, a copy, and a card in a hidden zone", () => {
    const v = view({ grave: [card("bolt", "opp")] });
    expect(lingerDestination(v, item("bolt", { kind: "activated" }))).toBeNull();
    expect(lingerDestination(v, item("bolt", { kind: "triggered" }))).toBeNull();
    expect(lingerDestination(v, item("bolt", { isCopy: true }))).toBeNull();
    expect(lingerDestination(v, item("bounced"))).toBeNull();
  });
});

describe("flight gate", () => {
  const s = (enabled: boolean, cardPlay: boolean, reduceMotion: boolean) => ({
    animations: { enabled, cardPlay },
    accessibility: { reduceMotion },
  });
  it("flies only with the master switch and cardPlay on and reduced motion off", () => {
    expect(flightAllowed(s(true, true, false))).toBe(true);
    expect(flightAllowed(s(false, true, false))).toBe(false);
    expect(flightAllowed(s(true, false, false))).toBe(false);
    expect(flightAllowed(s(true, true, true))).toBe(false);
  });
});

describe("geometry", () => {
  const board = { width: 1400, height: 900 };

  it("stays where it was drawn when nothing live is there", () => {
    expect(placeLinger(box(12, 300), [], board)).toEqual(box(12, 300));
    expect(placeLinger(box(12, 300), [box(600, 300)], board)).toEqual(box(12, 300));
  });

  it("moves out to the right of the live stack so it never covers the new top card", () => {
    const placed = placeLinger(box(12, 300), [box(12, 300), box(4, 260, 120, 400)], board);
    expect(placed).toEqual(box(124 + LINGER_GAP, 300));
  });

  it("goes above the stack when the right has no room", () => {
    const placed = placeLinger(box(200, 400, 1000, 60), [box(180, 380, 1040, 200)], board);
    expect(placed.top).toBe(380 - LINGER_GAP - 60);
    expect(placed.left).toBe(200);
  });

  it("tells a card from a row", () => {
    expect(lingerShape(box(0, 0, 240, 335))).toBe("card");
    expect(lingerShape(box(0, 0, 300, 34))).toBe("row");
  });

  it("puts the copy's centre on the landing at the landing's size", () => {
    const t = flightTransform(box(0, 0, 200, 280), box(1000, 700, 50, 70));
    expect(t.scale).toBeCloseTo(0.25);
    expect(t.x).toBeCloseTo(1025 - 25);
    expect(t.y).toBeCloseTo(735 - 35);
  });
});
