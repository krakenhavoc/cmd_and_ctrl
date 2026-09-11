import { describe, it, expect } from "vitest";
import { parseReplay, frameLabel, describeDelta, type ReplayFrame } from "./replay";
import type { GameView } from "./protocol";

function view(over: Record<string, unknown> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [],
    battlefield: { cards: [], count: 0 },
    stack: { cards: [], count: 0 },
    exile: { cards: [], count: 0 },
    turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "beginning", step: "untap" },
    mulligans_open: false,
    ...over,
  } as unknown as GameView;
}

function line(seq: number, over: Record<string, unknown> = {}): string {
  return JSON.stringify({ seq, game: view(over) });
}

describe("parseReplay", () => {
  it("reads one frame per line, in order", () => {
    const frames = parseReplay([line(1), line(2), line(3)].join("\n"));
    expect(frames.map((f) => f.seq)).toEqual([1, 2, 3]);
  });

  it("ignores blank lines and a trailing newline", () => {
    expect(parseReplay(`${line(1)}\n\n${line(2)}\n`).length).toBe(2);
  });

  // The log is appended to while the game runs, so a download taken
  // mid-write can end in a half-written line. Dropping it must not
  // cost the frames that came before.
  it("keeps every complete frame when the last line is torn", () => {
    const torn = `${line(1)}\n${line(2)}\n{"seq":3,"ga`;
    const frames = parseReplay(torn);
    expect(frames.map((f) => f.seq)).toEqual([1, 2]);
  });

  it("skips lines that are not snapshot payloads", () => {
    const junk = [
      "null",
      "42",
      '"a string"',
      '{"seq":"nope","game":{}}',
      '{"game":{}}',
      '{"seq":9}',
    ].join("\n");
    expect(parseReplay(`${junk}\n${line(7)}`).map((f) => f.seq)).toEqual([7]);
  });

  it("returns nothing for an empty body", () => {
    expect(parseReplay("")).toEqual([]);
    expect(parseReplay("\n\n  \n")).toEqual([]);
  });
});

describe("frameLabel", () => {
  it("names the turn, active seat and step", () => {
    const f: ReplayFrame = {
      seq: 12,
      game: view({
        seats: [
          { id: "a", name: "Alice" },
          { id: "b", name: "Bob" },
        ],
        turn: {
          number: 3,
          active_seat: 1,
          priority_holder: 1,
          phase: "combat",
          step: "declare_attackers",
        },
      }),
    };
    expect(frameLabel(f)).toBe("T3 · Bob · declare_attackers · seq 12");
  });

  it("degrades rather than throwing on a sparse frame", () => {
    expect(frameLabel(null)).toBe("—");
    expect(frameLabel({ seq: 4, game: {} as GameView })).toBe("seq 4");
    // Turn cursor present but no seat by that index.
    const f = {
      seq: 5,
      game: view({
        seats: [],
        turn: { number: 1, active_seat: 2, priority_holder: 2, phase: "", step: "" },
      }),
    };
    expect(frameLabel(f as ReplayFrame)).toBe("T1 · seat 2 · seq 5");
  });
});

describe("describeDelta", () => {
  const base: ReplayFrame = {
    seq: 1,
    game: view({ seats: [{ id: "a", name: "Alice", life: 40, hand: { count: 7 } }] }),
  };

  it("calls the first frame out as the start", () => {
    expect(describeDelta(null, base)).toBe("first recorded state");
  });

  it("reports life and hand changes by seat", () => {
    const next: ReplayFrame = {
      seq: 2,
      game: view({ seats: [{ id: "a", name: "Alice", life: 37, hand: { count: 6 } }] }),
    };
    const d = describeDelta(base, next);
    expect(d).toContain("Alice 40→37");
    expect(d).toContain("Alice hand -1");
  });

  it("reports battlefield and stack movement", () => {
    const next: ReplayFrame = {
      seq: 2,
      game: view({
        seats: [{ id: "a", name: "Alice", life: 40, hand: { count: 7 } }],
        battlefield: { cards: [{}, {}], count: 2 },
        stack: { cards: [{}], count: 1 },
      }),
    };
    const d = describeDelta(base, next);
    expect(d).toContain("+2 battlefield");
    expect(d).toContain("+1 stack");
  });

  it("says so when nothing visible moved", () => {
    expect(describeDelta(base, { ...base, seq: 2 })).toBe("no visible change");
  });
});
