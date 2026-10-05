import { describe, expect, it } from "vitest";

import {
  coinTurns,
  coinWon,
  DICE_BURST_MIN_MS,
  DICE_EDGE,
  DICE_FADE_MS,
  DICE_GAP,
  DICE_HOLD_MS,
  DICE_QUEUE_MAX,
  DICE_SHOWN_MAX,
  DICE_SLOT_MS,
  DICE_TUMBLE_MS,
  DICE_TUMBLE_STEPS,
  diceLabel,
  diceMotion,
  dicePhase,
  diceSeed,
  diceShown,
  diceSlotMs,
  diceThrow,
  emptyDiceState,
  faceAt,
  isReleased,
  nextDiceChange,
  placeDice,
  primeDice,
  rollFromLog,
  rollsFromLogs,
  trackDice,
  tumbleFaces,
  tumbleIndex,
  visiblePlays,
  type DiceOptions,
  type DiceState,
} from "./dice";
import type { LogEvent } from "./protocol";

const MOTION: DiceOptions = { motion: true, speed: 1 };
const STILL: DiceOptions = { motion: false, speed: 1 };

function roll(seq: number, seat = 0, results: number[] = [14], sides = 20): LogEvent {
  return {
    seq,
    kind: "roll",
    seat,
    text: `P${seat} rolled ${results.length === 1 ? `a d${sides}` : `${results.length}d${sides}`}: ${results.join(", ")}`,
    sides,
    results,
  };
}

function flip(seq: number, seat = 0, faces = ["heads"], call?: "heads" | "tails"): LogEvent {
  return {
    seq,
    kind: "flip",
    seat,
    text: `P${seat} flipped: ${faces.join(", ")}`,
    faces,
    ...(call ? { call } : {}),
  };
}

function step(seq: number): LogEvent {
  return { seq, kind: "step", seat: -1, text: "Upkeep", step: "upkeep" };
}

// A primed, empty table, then one frame.
function after(logs: LogEvent[], now = 0, opts = MOTION, prev?: DiceState): DiceState {
  const start = prev ?? primeDice([]);
  return trackDice(start, rollsFromLogs(logs), now, opts);
}

describe("rollFromLog", () => {
  it("makes one animation per roll entry, every die of the batch in it", () => {
    const r = rollFromLog(roll(7, 2, [3, 11], 12));
    expect(r).toMatchObject({
      key: "seq:7",
      seq: 7,
      seat: 2,
      kind: "die",
      sides: 12,
      results: [3, 11],
    });
  });

  it("makes a coin animation from a flip, with its call", () => {
    const r = rollFromLog(flip(8, 1, ["tails", "heads"], "heads"));
    expect(r).toMatchObject({ kind: "coin", sides: 2, faces: ["tails", "heads"], call: "heads" });
  });

  it("ignores other kinds and entries with nothing to land on", () => {
    expect(rollFromLog(step(1))).toBeNull();
    expect(rollFromLog({ seq: 2, kind: "roll", seat: 0, text: "x", sides: 20 })).toBeNull();
    expect(rollFromLog({ seq: 3, kind: "flip", seat: 0, text: "x", faces: ["edge"] })).toBeNull();
  });

  it("keeps log order", () => {
    expect(rollsFromLogs([roll(3), step(4), flip(5), roll(6, 1)]).map((r) => r.seq)).toEqual([
      3, 5, 6,
    ]);
  });
});

describe("timing", () => {
  it("tumbles 900 ms, holds 1.6 s, fades 200 ms", () => {
    const [p] = after([roll(1)], 1000).plays;
    expect(p.startAt).toBe(1000);
    expect(p.settleAt).toBe(1000 + DICE_TUMBLE_MS);
    expect(p.fadeAt).toBe(1000 + DICE_TUMBLE_MS + DICE_HOLD_MS);
    expect(p.endAt).toBe(1000 + DICE_TUMBLE_MS + DICE_HOLD_MS + DICE_FADE_MS);
    expect([DICE_TUMBLE_MS, DICE_HOLD_MS, DICE_FADE_MS]).toEqual([900, 1600, 200]);
  });

  it("scales the tumble with animation speed, never the hold", () => {
    const [p] = after([roll(1)], 0, { motion: true, speed: 2 }).plays;
    expect(p.settleAt - p.startAt).toBe(1800);
    expect(p.fadeAt - p.settleAt).toBe(DICE_HOLD_MS);
    const [q] = after([roll(1)], 0, { motion: true, speed: 0.5 }).plays;
    expect(q.settleAt - q.startAt).toBe(450);
    expect(q.fadeAt - q.settleAt).toBe(DICE_HOLD_MS);
  });

  it("reads the phases off the clock", () => {
    const [p] = after([roll(1)], 0).plays;
    expect(dicePhase(p, 0)).toBe("tumble");
    expect(dicePhase(p, 899)).toBe("tumble");
    expect(dicePhase(p, 900)).toBe("hold");
    expect(dicePhase(p, 2499)).toBe("hold");
    expect(dicePhase(p, 2500)).toBe("fade");
    expect(visiblePlays({ plays: [p] }, 2699)).toHaveLength(1);
    expect(visiblePlays({ plays: [p] }, 2700)).toHaveLength(0);
  });

  it("knows its next change, and nothing once it is idle", () => {
    const s = after([roll(1)], 0);
    expect(nextDiceChange(s, 0)).toBe(900);
    expect(nextDiceChange(s, 900)).toBe(2500);
    expect(nextDiceChange(s, 2500)).toBe(2700);
    expect(nextDiceChange(s, 2700)).toBeNull();
  });
});

describe("the strip cue waits for the die", () => {
  it("is released when the die settles, not when the frame arrives", () => {
    const s = after([roll(1)], 0);
    expect(isReleased(s, 1, 0)).toBe(false);
    expect(isReleased(s, 1, 899)).toBe(false);
    expect(isReleased(s, 1, 900)).toBe(true);
  });

  it("is released at once with motion off", () => {
    const s = after([roll(1)], 50, STILL);
    expect(isReleased(s, 1, 50)).toBe(true);
  });

  it("is released at once with motion off even behind a die still in the air", () => {
    let s = after([roll(1)], 0, MOTION);
    s = trackDice(s, rollsFromLogs([roll(1), roll(2)]), 100, STILL);
    expect(isReleased(s, 2, 100)).toBe(true);
  });

  it("does not release an entry it has never seen", () => {
    expect(isReleased(emptyDiceState(), 9, 1e12)).toBe(false);
  });
});

describe("motion off", () => {
  it("shows the result settled at once, for the same hold, then fades", () => {
    const [p] = after([roll(1, 0, [17])], 0, STILL).plays;
    expect(p.motion).toBe(false);
    expect(p.settleAt).toBe(p.startAt);
    expect(p.fadeAt - p.settleAt).toBe(DICE_HOLD_MS);
    expect(p.endAt - p.fadeAt).toBe(DICE_FADE_MS);
    expect(dicePhase(p, 0)).toBe("hold");
    expect(faceAt(p, 0, 0)).toBe(17);
  });

  it("is off with the master switch, the dice toggle or reduced motion", () => {
    expect(diceMotion({ enabled: true, dice: true, reduceMotion: false })).toBe(true);
    expect(diceMotion({ enabled: false, dice: true, reduceMotion: false })).toBe(false);
    expect(diceMotion({ enabled: true, dice: false, reduceMotion: false })).toBe(false);
    expect(diceMotion({ enabled: true, dice: true, reduceMotion: true })).toBe(false);
  });
});

describe("priming, undo and the window", () => {
  it("never animates what was in the window on the first frame", () => {
    const primed = primeDice(rollsFromLogs([roll(1), flip(2)]));
    const s = trackDice(primed, rollsFromLogs([roll(1), flip(2)]), 0, MOTION);
    expect(s.plays).toHaveLength(0);
    expect(isReleased(s, 1, 0)).toBe(true);
    expect(s.primed.has(2)).toBe(true);
  });

  it("animates each entry once across frames", () => {
    let s = after([roll(1)], 0);
    s = trackDice(s, rollsFromLogs([roll(1), step(2)]), 100, MOTION);
    expect(s.plays).toHaveLength(1);
  });

  it("animates a rewound roll again when it is rolled again", () => {
    let s = after([roll(5)], 0);
    s = trackDice(s, rollsFromLogs([]), 5000, MOTION); // the undo frame
    expect(s.seen.has("seq:5")).toBe(false);
    s = trackDice(s, rollsFromLogs([roll(5)]), 6000, MOTION); // the redo
    expect(s.plays.map((p) => p.startAt)).toEqual([6000]);
  });

  it("stops a die whose entry was rewound mid-tumble", () => {
    let s = after([roll(5)], 0);
    s = trackDice(s, [], 300, MOTION);
    expect(s.plays).toHaveLength(0);
  });
});

describe("seats", () => {
  it("plays different seats at once, each at its own place", () => {
    const s = after([roll(1, 0), roll(2, 1), flip(3, 2), roll(4, 3)], 0);
    expect(s.plays.map((p) => p.startAt)).toEqual([0, 0, 0, 0]);
    expect(new Set(s.plays.map((p) => p.roll.seat))).toEqual(new Set([0, 1, 2, 3]));
  });

  it("queues a seat's next roll behind the one on screen, in log order", () => {
    let s = after([roll(1)], 0);
    s = trackDice(s, rollsFromLogs([roll(1), roll(2)]), 300, MOTION);
    const [a, b] = [...s.plays].sort((x, y) => x.startAt - y.startAt);
    expect(a.roll.seq).toBe(1);
    expect(b.roll.seq).toBe(2);
    expect(b.startAt).toBe(a.endAt);
    expect(b.settleAt - b.startAt).toBe(DICE_TUMBLE_MS);
  });
});

describe("bursts", () => {
  it("gives each entry of a burst max(400 ms, 2.5 s / n)", () => {
    expect(diceSlotMs(1)).toBe(DICE_SLOT_MS);
    expect(diceSlotMs(2)).toBe(1250);
    expect(diceSlotMs(5)).toBe(500);
    expect(diceSlotMs(10)).toBe(DICE_BURST_MIN_MS);
  });

  it("plays a burst in log order, back to back, each in its share", () => {
    const s = after([roll(1), roll(2)], 0);
    const [a, b] = [...s.plays].sort((x, y) => x.startAt - y.startAt);
    expect([a.roll.seq, b.roll.seq]).toEqual([1, 2]);
    expect(a.startAt).toBe(0);
    expect(a.fadeAt - a.startAt).toBe(1250);
    expect(b.startAt).toBe(a.endAt);
  });

  it("keeps at most three, dropping the oldest, and releases the dropped at once", () => {
    const s = after([roll(1), roll(2), roll(3), roll(4), roll(5)], 0);
    expect(s.plays.map((p) => p.roll.seq).sort()).toEqual([3, 4, 5]);
    expect(s.plays).toHaveLength(DICE_QUEUE_MAX);
    expect(isReleased(s, 1, 0)).toBe(true);
    expect(isReleased(s, 2, 0)).toBe(true);
    expect(isReleased(s, 3, 0)).toBe(false);
  });

  it("never cuts short the one on screen; it counts toward the three", () => {
    let s = after([roll(1)], 0);
    s = trackDice(s, rollsFromLogs([roll(1), roll(2), roll(3), roll(4)]), 200, MOTION);
    const seqs = [...s.plays].sort((x, y) => x.startAt - y.startAt).map((p) => p.roll.seq);
    expect(seqs).toEqual([1, 3, 4]);
    const first = s.plays.find((p) => p.roll.seq === 1)!;
    expect(first.endAt).toBe(DICE_SLOT_MS + DICE_FADE_MS);
  });

  it("counts a burst per seat: another seat's rolls do not shorten mine", () => {
    const s = after([roll(1, 0), roll(2, 1), roll(3, 1)], 0);
    const mine = s.plays.find((p) => p.roll.seat === 0)!;
    expect(mine.fadeAt - mine.startAt).toBe(DICE_SLOT_MS);
  });
});

describe("batches", () => {
  it("draws at most six dice and counts the rest", () => {
    const big = rollFromLog(roll(1, 0, [1, 2, 3, 4, 5, 6, 7, 8, 9], 6))!;
    expect(diceShown(big)).toEqual({ count: DICE_SHOWN_MAX, more: 3 });
    expect(diceShown(rollFromLog(roll(2, 0, [4, 9], 12))!)).toEqual({ count: 2, more: 0 });
    expect(
      diceShown(rollFromLog(flip(3, 0, ["heads", "tails", "heads", "heads", "tails"]))!),
    ).toEqual({
      count: 5,
      more: 0,
    });
  });

  it("labels the group", () => {
    expect(diceLabel(rollFromLog(roll(1))!)).toBe("d20");
    expect(diceLabel(rollFromLog(roll(1, 0, [3, 4], 12))!)).toBe("2d12");
    expect(diceLabel(rollFromLog(flip(1))!)).toBe("coin");
    expect(diceLabel(rollFromLog(flip(1, 0, ["heads", "tails"]))!)).toBe("2 coins");
  });

  it("says which coins of a called flip won", () => {
    const r = rollFromLog(flip(1, 0, ["tails", "heads"], "heads"))!;
    expect([coinWon(r, 0), coinWon(r, 1)]).toEqual([false, true]);
    expect(coinWon(rollFromLog(flip(2))!, 0)).toBeUndefined();
  });
});

describe("the tumble is deterministic", () => {
  it("is the same sequence for the same entry, and lands on the result", () => {
    const a = tumbleFaces(diceSeed("seq:12"), 20, 14);
    const b = tumbleFaces(diceSeed("seq:12"), 20, 14);
    expect(a).toEqual(b);
    expect(a).toHaveLength(DICE_TUMBLE_STEPS);
    expect(a[a.length - 1]).toBe(14);
  });

  it("differs between entries and between dice of one entry", () => {
    expect(tumbleFaces(diceSeed("seq:12"), 20, 14)).not.toEqual(
      tumbleFaces(diceSeed("seq:13"), 20, 14),
    );
    expect(tumbleFaces(diceSeed("seq:12", 0), 20, 14)).not.toEqual(
      tumbleFaces(diceSeed("seq:12", 1), 20, 14),
    );
  });

  it("stays on the die, never repeats a face back to back, and never shows the result early", () => {
    for (let seq = 0; seq < 200; seq++) {
      for (const sides of [4, 6, 12, 20]) {
        const result = 1 + (seq % sides);
        const f = tumbleFaces(diceSeed(`seq:${seq}`), sides, result);
        for (let i = 0; i < f.length; i++) {
          expect(f[i]).toBeGreaterThanOrEqual(1);
          expect(f[i]).toBeLessThanOrEqual(sides);
          if (i > 0) expect(f[i]).not.toBe(f[i - 1]);
          if (i < f.length - 1) expect(f[i]).not.toBe(result);
        }
      }
    }
  });

  it("never calls Math.random", () => {
    const real = Math.random;
    Math.random = () => {
      throw new Error("Math.random called");
    };
    try {
      const s = after([roll(1, 0, [3, 5], 6), flip(2, 1, ["tails"])], 0);
      for (const p of s.plays) {
        if (p.roll.kind === "die") faceAt(p, 0, 100);
        diceThrow(p.roll.key, 0);
        if (p.roll.kind === "coin") coinTurns(p.roll.key, 0, "tails");
      }
    } finally {
      Math.random = real;
    }
  });

  it("changes faces fast, then slow, and shows the result only once settled", () => {
    expect(tumbleIndex(0, 900)).toBe(0);
    expect(tumbleIndex(899, 900)).toBe(DICE_TUMBLE_STEPS - 2);
    expect(tumbleIndex(900, 900)).toBe(DICE_TUMBLE_STEPS - 1);
    expect(tumbleIndex(0, 0)).toBe(DICE_TUMBLE_STEPS - 1);
    const firstHalf = tumbleIndex(450, 900);
    expect(firstHalf).toBeGreaterThan((DICE_TUMBLE_STEPS - 1) / 2);
  });

  it("reads the face off the clock, landing on the server's number", () => {
    const [p] = after([roll(9, 0, [7], 20)], 0).plays;
    const seen = new Set<number>();
    for (let t = 0; t < 900; t += 30) seen.add(faceAt(p, 0, t));
    expect(seen.size).toBeGreaterThan(3);
    expect(seen.has(7)).toBe(false);
    expect(faceAt(p, 0, 900)).toBe(7);
  });

  it("lands a coin on its face: heads on an even count of half turns, tails on an odd one", () => {
    for (let i = 0; i < 50; i++) {
      expect(coinTurns(`seq:${i}`, 0, "heads") % 2).toBe(0);
      expect(coinTurns(`seq:${i}`, 0, "tails") % 2).toBe(1);
      expect(coinTurns(`seq:${i}`, 0, "heads")).toBeGreaterThanOrEqual(7);
    }
  });

  it("throws every die a whole number of turns, so it lands upright", () => {
    for (let i = 0; i < 50; i++) expect(Math.abs(diceThrow(`seq:${i}`, 0).spin % 360)).toBe(0);
  });
});

describe("placeDice", () => {
  const board = { width: 1000, height: 600 };
  const size = { width: 60, height: 70 };

  it("puts the dice on the avatar's side toward the centre", () => {
    // The bottom seat, near the middle: above it.
    const low = placeDice(board, { left: 400, top: 540, width: 40, height: 40 }, size, null);
    expect(low.side).toBe("above");
    expect(low.top).toBe(540 - DICE_GAP - size.height);
    // A seat on the left edge, halfway down: to its right.
    const left = placeDice(board, { left: 10, top: 280, width: 40, height: 40 }, size, null);
    expect(left).toMatchObject({ side: "right", left: 10 + 40 + DICE_GAP });
    // Top-right: below it.
    const top = placeDice(board, { left: 700, top: 10, width: 40, height: 40 }, size, null);
    expect(top.side).toBe("below");
  });

  it("keeps the dice inside the board", () => {
    const p = placeDice(board, { left: 970, top: 0, width: 30, height: 30 }, size, null);
    expect(p.left).toBeLessThanOrEqual(board.width - size.width - DICE_EDGE);
    expect(p.top).toBeGreaterThanOrEqual(DICE_EDGE);
  });

  it("falls back to the attention strip when the avatar is not on screen", () => {
    const p = placeDice(board, null, size, { left: 12, top: 120, width: 400, height: 0 });
    expect(p).toEqual({ side: "strip", left: 12, top: 120 + DICE_GAP });
  });
});

describe("placeDice avoiding overlays (#2260)", () => {
  const board = { width: 1200, height: 700 };
  const size = { width: 80, height: 90 };
  // The drawer: 380px wide down the right edge.
  const log = { left: 820, top: 0, width: 380, height: 700 };
  const hits = (p: { left: number; top: number }, r: typeof log) =>
    p.left < r.left + r.width &&
    p.left + size.width > r.left &&
    p.top < r.top + r.height &&
    p.top + size.height > r.top;

  it("is unchanged with nothing to avoid", () => {
    const a = { left: 1000, top: 300, width: 40, height: 40 };
    expect(placeDice(board, a, size, null, [])).toEqual(placeDice(board, a, size, null));
  });

  it("falls back to the strip when the avatar is under the open log", () => {
    const strip = { left: 12, top: 120, width: 400, height: 0 };
    const under = { left: 1000, top: 300, width: 40, height: 40 };
    const open = placeDice(board, under, size, strip, [log]);
    expect(open.side).toBe("strip");
    expect(hits(open, log)).toBe(false);
    // With the log closed the same avatar keeps its die.
    expect(placeDice(board, under, size, strip, []).side).not.toBe("strip");
  });

  it("keeps the strip fallback clear of the log", () => {
    const strip = { left: 700, top: 20, width: 400, height: 0 };
    const p = placeDice(board, null, size, strip, [log]);
    expect(hits(p, log)).toBe(false);
  });

  it("moves to another side when the preferred side is covered, for every avatar position", () => {
    const spots = [
      { left: 700, top: 10, width: 40, height: 40 },
      { left: 770, top: 300, width: 40, height: 40 },
      { left: 400, top: 640, width: 40, height: 40 },
      { left: 10, top: 300, width: 40, height: 40 },
    ];
    for (const a of spots) {
      const p = placeDice(board, a, size, null, [log]);
      expect(hits(p, log), JSON.stringify(a)).toBe(false);
    }
  });

  it("keeps the die off a seat's text read-out", () => {
    // Opponent at the top, centre: its counts line sits just below the
    // avatar, which is where "toward the centre" would put the die.
    const a = { left: 560, top: 10, width: 60, height: 60 };
    const counts = { left: 480, top: 74, width: 240, height: 20 };
    const plain = placeDice(board, a, size, null, []);
    expect(plain.side).toBe("below");
    expect(hits(plain, counts)).toBe(true);
    const p = placeDice(board, a, size, null, [counts]);
    expect(hits(p, counts)).toBe(false);
  });
});

// ADR 0121 §5: a die or a coin rolled at the table. Keyed on its
// roll_id, because an undo of an earlier action writes the same line
// again under a new seq.
function tableRoll(seq: number, rollID: number, seat = 1, result: number | "heads" = 14): LogEvent {
  const coin = result === "heads";
  return {
    seq,
    kind: "table_roll",
    seat,
    text: coin
      ? `P${seat} flipped a coin at the table: heads`
      : `P${seat} rolled a d20 at the table: ${result}`,
    roll_id: rollID,
    ...(coin ? { faces: ["heads"] } : { sides: 20, results: [result] }),
  };
}

describe("table rolls", () => {
  it("keys a table roll on its roll_id, a die or a coin", () => {
    expect(rollFromLog(tableRoll(9, 3))).toMatchObject({
      key: "table:3",
      seq: 9,
      seat: 1,
      kind: "die",
      sides: 20,
      results: [14],
      source: "table",
    });
    expect(rollFromLog(tableRoll(10, 4, 0, "heads"))).toMatchObject({
      key: "table:4",
      kind: "coin",
      sides: 2,
      faces: ["heads"],
      source: "table",
    });
    // No roll_id (an older server) or nothing to land on: no animation.
    expect(rollFromLog({ ...tableRoll(11, 5), roll_id: undefined })).toBeNull();
    expect(rollFromLog({ seq: 12, kind: "table_roll", seat: 0, text: "x", roll_id: 6 })).toBeNull();
  });

  it("tumbles at the roller's seat like any roll", () => {
    const [p] = after([tableRoll(9, 3)], 1000).plays;
    expect(p.roll.seat).toBe(1);
    expect(p.settleAt).toBe(1000 + DICE_TUMBLE_MS);
  });

  it("does not play again when an undo writes its line under a new seq", () => {
    const first = after([roll(5, 0), tableRoll(9, 3)], 1000);
    expect(first.plays.map((p) => p.roll.key).sort()).toEqual(["seq:5", "table:3"]);
    // The undo rewound seq 5 and wrote the table roll again at seq 5.
    const moved = trackDice(first, rollsFromLogs([tableRoll(5, 3)]), 1100, MOTION);
    expect(moved.plays.map((p) => p.roll.key)).toEqual(["table:3"]);
    const play = moved.plays[0];
    expect(play.roll.seq).toBe(5);
    expect(play.startAt).toBe(1000);
    // Its strip cue follows it to the new seq, not released early.
    expect(isReleased(moved, 5, 1100)).toBe(false);
    expect(isReleased(moved, 5, play.settleAt)).toBe(true);
  });

  it("stays primed when the line of a roll from before this client moves", () => {
    const primed = primeDice(rollsFromLogs([tableRoll(9, 3)]));
    const moved = trackDice(primed, rollsFromLogs([tableRoll(4, 3)]), 1000, MOTION);
    expect(moved.plays).toEqual([]);
    expect(moved.primed.has(4)).toBe(true);
    expect(isReleased(moved, 4, 1000)).toBe(true);
  });

  it("plays a new roll with a new roll_id", () => {
    const first = after([tableRoll(9, 3)], 1000);
    const next = trackDice(first, rollsFromLogs([tableRoll(9, 3), tableRoll(10, 4)]), 1100, MOTION);
    expect(next.plays.map((p) => p.roll.key)).toEqual(["table:3", "table:4"]);
  });
});
