import { describe, expect, it } from "vitest";

import {
  REVEAL_CUE_LIMIT,
  REVEAL_TTL_MS,
  RANDOM_CUE_LIMIT,
  RANDOM_CUE_TTL_MS,
  dismissReveal,
  emptyRandomState,
  emptyRevealState,
  hiddenRevealCount,
  primeRandomEvents,
  primeRevealState,
  revealHeadline,
  trackRandomEvents,
  trackReveals,
} from "./reveals";
import type { LogEvent, RevealView } from "./protocol";

function reveal(seq: number, extra: Partial<RevealView> = {}): RevealView {
  return {
    seq,
    seat: 0,
    turn: 4,
    cards: [{ name: `Card ${seq}`, type_line: "Instant" }],
    ...extra,
  };
}

function randomLog(seq: number, kind: "roll" | "flip" = "flip"): LogEvent {
  return { seq, kind, seat: 0, text: `${kind} result ${seq}` };
}

describe("trackReveals", () => {
  it("cues a reveal the first time its seq appears", () => {
    const s = trackReveals(emptyRevealState(), [reveal(10)], 0);
    expect(s.cues).toHaveLength(1);
    expect(s.cues[0].reveal.seq).toBe(10);
  });

  it("does not re-cue a reveal the window keeps repeating", () => {
    // The server window reports the same reveal on every frame for the
    // rest of the turn. Announcing it once is the whole job.
    let s = trackReveals(emptyRevealState(), [reveal(10)], 0);
    for (let t = 100; t < 1000; t += 100) {
      s = trackReveals(s, [reveal(10)], t);
    }
    expect(s.cues).toHaveLength(1);
    expect(s.cues[0].shownAt).toBe(0);
  });

  it("ages a cue out on the timer, not on the window", () => {
    // Still in the window, past its welcome.
    let s = trackReveals(emptyRevealState(), [reveal(10)], 0);
    s = trackReveals(s, [reveal(10)], REVEAL_TTL_MS - 1);
    expect(s.cues).toHaveLength(1);
    s = trackReveals(s, [reveal(10)], REVEAL_TTL_MS);
    expect(s.cues).toHaveLength(0);
  });

  it("expires a cue even on a frame that admits nothing", () => {
    let s = trackReveals(emptyRevealState(), [reveal(10)], 0);
    s = trackReveals(s, undefined, REVEAL_TTL_MS + 1);
    expect(s.cues).toHaveLength(0);
  });

  it("keeps the newest when a burst overflows the strip", () => {
    const s = trackReveals(emptyRevealState(), [reveal(1), reveal(2), reveal(3), reveal(4)], 0);
    expect(s.cues).toHaveLength(REVEAL_CUE_LIMIT);
    expect(s.cues.map((c) => c.reveal.seq)).toEqual([3, 4]);
    // Everything was still SEEN, so the two that lost the strip do not
    // come back on the next frame.
    const next = trackReveals(s, [reveal(1), reveal(2), reveal(3), reveal(4)], 10);
    expect(next.cues.map((c) => c.reveal.seq)).toEqual([3, 4]);
  });

  it("does not mutate the state it was given", () => {
    const first = trackReveals(emptyRevealState(), [reveal(1)], 0);
    const before = first.cues.length;
    trackReveals(first, [reveal(2)], 10);
    expect(first.cues).toHaveLength(before);
    expect(first.seen.has(2)).toBe(false);
  });
});

describe("trackRandomEvents", () => {
  it("drops undone outcomes and cues the replay again", () => {
    let s = trackRandomEvents(emptyRandomState(), [randomLog(10)], 0);
    s = trackRandomEvents(s, [], 100);
    expect(s.cues).toHaveLength(0);
    expect(s.seen.size).toBe(0);
    s = trackRandomEvents(s, [randomLog(10)], 200);
    expect(s.cues.map((c) => c.log.seq)).toEqual([10]);
  });
  it("cues each random batch once even when the log is replayed", () => {
    let s = trackRandomEvents(emptyRandomState(), [randomLog(10)], 0);
    s = trackRandomEvents(s, [randomLog(10), randomLog(11, "roll")], 100);
    expect(s.cues.map((c) => c.log.seq)).toEqual([10, 11]);
  });

  it("primes reconnect history and ages cues independently", () => {
    const primed = primeRandomEvents([randomLog(7)]);
    expect(trackRandomEvents(primed, [randomLog(7)], 0).cues).toHaveLength(0);
    let s = trackRandomEvents(emptyRandomState(), [randomLog(8)], 0);
    s = trackRandomEvents(s, undefined, RANDOM_CUE_TTL_MS);
    expect(s.cues).toHaveLength(0);
  });

  it("keeps only the newest random batches", () => {
    const s = trackRandomEvents(emptyRandomState(), [randomLog(1), randomLog(2), randomLog(3)], 0);
    expect(RANDOM_CUE_LIMIT).toBe(2);
    expect(s.cues.map((c) => c.log.seq)).toEqual([2, 3]);
  });
});

describe("primeRevealState", () => {
  it("swallows the window a reconnecting client arrives to", () => {
    // These reveals happened before this client was looking. Showing
    // them now would claim they are happening now.
    const primed = primeRevealState([reveal(7), reveal(8)]);
    expect(primed.cues).toHaveLength(0);
    const s = trackReveals(primed, [reveal(7), reveal(8)], 0);
    expect(s.cues).toHaveLength(0);
  });

  it("still cues what happens after the prime", () => {
    const primed = primeRevealState([reveal(7)]);
    const s = trackReveals(primed, [reveal(7), reveal(9)], 0);
    expect(s.cues.map((c) => c.reveal.seq)).toEqual([9]);
  });

  it("primes to empty when the frame carries no reveals", () => {
    expect(primeRevealState(undefined).seen.size).toBe(0);
  });
});

describe("dismissReveal", () => {
  it("drops the cue and does not let the window put it back", () => {
    let s = trackReveals(emptyRevealState(), [reveal(10)], 0);
    s = dismissReveal(s, 10);
    expect(s.cues).toHaveLength(0);
    s = trackReveals(s, [reveal(10)], 100);
    expect(s.cues).toHaveLength(0);
  });
});

describe("hiddenRevealCount", () => {
  it("is zero for a reveal the frame named in full", () => {
    expect(hiddenRevealCount(reveal(1))).toBe(0);
    expect(hiddenRevealCount(reveal(1, { count: 1 }))).toBe(0);
  });

  it("reports the remainder when the server capped the card list", () => {
    // Hermit Druid with no basic land reveals the whole library; the
    // wire ships eight of them and says how many there were.
    const r = reveal(1, {
      count: 63,
      cards: Array.from({ length: 8 }, (_, i) => ({ name: `Card ${i}` })),
    });
    expect(hiddenRevealCount(r)).toBe(55);
  });
});

describe("revealHeadline", () => {
  it("names the source card when the frame carries one", () => {
    expect(revealHeadline(reveal(1, { source: "Hermit Druid" }), "Alice")).toBe(
      "Alice revealed · Hermit Druid",
    );
  });

  it("falls back when the source is not in a zone the table can see", () => {
    expect(revealHeadline(reveal(1), "Alice")).toBe("Alice revealed");
    expect(revealHeadline(reveal(1), "")).toBe("A player revealed");
  });
});
