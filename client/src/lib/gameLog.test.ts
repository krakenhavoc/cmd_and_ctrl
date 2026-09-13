import { describe, expect, it } from "vitest";

import { entrySeats, filterBySeat, groupLog, logTone, seatName } from "./gameLog";
import type { GameView, LogEvent, LogKind } from "./protocol";

let seq = 0;
function entry(kind: LogKind, text: string, extra: Partial<LogEvent> = {}): LogEvent {
  seq += 1;
  return { seq, kind, seat: 0, text, ...extra };
}

function step(turn: number, seat: number, name: string): LogEvent {
  return entry("step", `Turn ${turn} — P${seat + 1} · ${name}`, { seat, turn, step: name });
}

describe("groupLog", () => {
  it("folds entries into the step they happened in", () => {
    const groups = groupLog([
      step(1, 0, "precombat main"),
      entry("cast", "P1 cast Llanowar Elves"),
      entry("resolve", "Llanowar Elves resolved"),
      step(1, 0, "declare attackers"),
      entry("attack", "Grizzly Bears attacks P2", { target_seat: 1 }),
    ]);
    expect(groups).toHaveLength(2);
    expect(groups[0].header).toContain("precombat main");
    expect(groups[0].entries.map((e) => e.kind)).toEqual(["cast", "resolve"]);
    expect(groups[1].entries.map((e) => e.kind)).toEqual(["attack"]);
  });

  it("keeps entries that precede the first step in an unheaded block", () => {
    // The log is a ring, so the oldest entries it holds routinely have
    // no step announcement in front of them.
    const groups = groupLog([
      entry("draw", "P1 drew 7 cards", { amount: 7 }),
      step(2, 1, "upkeep"),
      entry("life", "P2 lost 1 life", { seat: 1, amount: -1 }),
    ]);
    expect(groups[0].header).toBeNull();
    expect(groups[0].seat).toBe(-1);
    expect(groups[0].entries).toHaveLength(1);
    expect(groups[1].header).toContain("upkeep");
  });

  it("drops empty step blocks but keeps the most recent one", () => {
    const groups = groupLog([
      step(3, 0, "upkeep"),
      step(3, 0, "draw"),
      entry("draw", "P1 drew a card", { amount: 1 }),
      step(3, 0, "precombat main"),
    ]);
    expect(groups.map((g) => g.header)).toEqual([
      expect.stringContaining("draw"),
      expect.stringContaining("precombat main"),
    ]);
  });

  it("returns nothing for an absent or empty log", () => {
    expect(groupLog(undefined)).toEqual([]);
    expect(groupLog([])).toEqual([]);
  });
});

describe("filterBySeat", () => {
  const log = [
    step(4, 0, "precombat main"),
    entry("cast", "P1 cast Bolt", { seat: 0 }),
    entry("damage", "Bolt dealt 3 damage to P2", { seat: 0, target_seat: 1 }),
    entry("draw", "P3 drew a card", { seat: 2 }),
  ];

  it("keeps entries that involve the seat, as actor or as target", () => {
    const mine = filterBySeat(log, 1);
    expect(mine.map((e) => e.kind)).toEqual(["step", "damage"]);
  });

  it("always keeps step entries, which are the only timestamps", () => {
    expect(filterBySeat(log, 3).map((e) => e.kind)).toEqual(["step"]);
  });

  it("passes everything through when no seat is selected", () => {
    expect(filterBySeat(log, null)).toHaveLength(log.length);
  });
});

describe("entrySeats", () => {
  it("lists actor and target without duplicating a self-targeting entry", () => {
    expect(entrySeats(entry("damage", "x", { seat: 2, target_seat: 2 }))).toEqual([2]);
    expect(entrySeats(entry("damage", "x", { seat: 2, target_seat: 0 }))).toEqual([2, 0]);
    expect(entrySeats(entry("resolve", "x", { seat: -1 }))).toEqual([]);
  });
});

describe("logTone", () => {
  it("has a tone for every kind the server can send", () => {
    const kinds: LogKind[] = [
      "step",
      "cast",
      "resolve",
      "fizzle",
      "counter",
      "zone",
      "draw",
      "life",
      "damage",
      "attack",
      "block",
      "token",
      "sacrifice",
      "eliminated",
    ];
    for (const k of kinds) expect(logTone(k)).toMatch(/^tone-/);
  });
});

describe("seatName", () => {
  const view = {
    seats: [
      { id: "a", name: "Aang" },
      { id: "b", name: "Katara" },
    ],
  } as unknown as GameView;

  it("resolves a seat index, and refuses the no-seat sentinel", () => {
    expect(seatName(view, 1)).toBe("Katara");
    expect(seatName(view, -1)).toBeNull();
    expect(seatName(view, 9)).toBeNull();
    expect(seatName(null, 0)).toBeNull();
  });
});
