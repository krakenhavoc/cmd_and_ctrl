import { describe, expect, it } from "vitest";
import { createCueTracker, isRealPriority, turnBeginCue, type CueFrame } from "./interfaceCues";

function frame(over: Partial<CueFrame> = {}): CueFrame {
  return {
    playing: true,
    hasPriority: false,
    tableBusy: false,
    hasPendingChoice: false,
    verdict: "hold",
    promptIDs: [],
    ...over,
  };
}

describe("isRealPriority", () => {
  it("is true when the viewer holds priority and the autopass holds", () => {
    expect(isRealPriority(frame({ hasPriority: true }))).toBe(true);
    expect(isRealPriority(frame({ hasPriority: true, verdict: "clear-toggle" }))).toBe(true);
  });

  it("is false on an auto-pass or a timed bluff", () => {
    expect(isRealPriority(frame({ hasPriority: true, verdict: "pass" }))).toBe(false);
    expect(isRealPriority(frame({ hasPriority: true, verdict: { kind: "bluff" } }))).toBe(false);
  });

  it("is false without priority, on a busy table, or with a prompt open", () => {
    expect(isRealPriority(frame())).toBe(false);
    expect(isRealPriority(frame({ hasPriority: true, tableBusy: true }))).toBe(false);
    expect(isRealPriority(frame({ hasPriority: true, hasPendingChoice: true }))).toBe(false);
    expect(isRealPriority(frame({ hasPriority: true, playing: false }))).toBe(false);
  });
});

describe("createCueTracker", () => {
  it("plays nothing on the first frame, even if it is already waiting on you", () => {
    const t = createCueTracker();
    expect(
      t.observe(frame({ hasPriority: true, promptIDs: ["c1"], hasPendingChoice: true })),
    ).toEqual([]);
  });

  it("cues priority on the edge, once per run of holding it", () => {
    const t = createCueTracker();
    t.observe(frame());
    expect(t.observe(frame({ hasPriority: true }))).toEqual(["priority"]);
    expect(t.observe(frame({ hasPriority: true }))).toEqual([]);
    expect(t.observe(frame())).toEqual([]);
    expect(t.observe(frame({ hasPriority: true }))).toEqual(["priority"]);
  });

  it("does not cue priority the viewer auto-passes", () => {
    const t = createCueTracker();
    t.observe(frame());
    expect(t.observe(frame({ hasPriority: true, verdict: "pass" }))).toEqual([]);
    expect(t.observe(frame({ hasPriority: true, verdict: { kind: "bluff" } }))).toEqual([]);
    // The cue comes when the table truly stops, not when priority first arrived.
    expect(t.observe(frame({ hasPriority: true }))).toEqual(["priority"]);
  });

  it("cues a prompt once per new choice, and again for the next one", () => {
    const t = createCueTracker();
    t.observe(frame());
    const open = { hasPriority: true, hasPendingChoice: true };
    expect(t.observe(frame({ ...open, promptIDs: ["c1"] }))).toEqual(["prompt"]);
    expect(t.observe(frame({ ...open, promptIDs: ["c1"] }))).toEqual([]);
    expect(t.observe(frame({ ...open, promptIDs: ["c1", "c2"] }))).toEqual(["prompt"]);
    expect(t.observe(frame({ ...open, promptIDs: ["c2"] }))).toEqual([]);
    expect(t.observe(frame({ ...open, promptIDs: ["c3", "c4"] }))).toEqual(["prompt"]);
  });

  it("answers a prompt, then cues priority when the decision is real", () => {
    const t = createCueTracker();
    t.observe(frame());
    t.observe(frame({ hasPriority: true, hasPendingChoice: true, promptIDs: ["c1"] }));
    expect(t.observe(frame({ hasPriority: true }))).toEqual(["priority"]);
  });

  it("stays silent when the game is not being played", () => {
    const t = createCueTracker();
    t.observe(frame({ playing: false }));
    expect(t.observe(frame({ playing: false, promptIDs: ["c1"] }))).toEqual([]);
  });

  it("baselines again after reset", () => {
    const t = createCueTracker();
    t.observe(frame());
    t.reset();
    expect(t.observe(frame({ hasPriority: true }))).toEqual([]);
  });
});

describe("turnBeginCue", () => {
  it("is your_turn when untap begins on the viewer's turn", () => {
    expect(turnBeginCue("cleanup", "untap", true)).toBe("your_turn");
  });

  it("is turn_change when untap begins on someone else's turn", () => {
    expect(turnBeginCue("cleanup", "untap", false)).toBe("turn_change");
  });

  it("is nothing for the baseline frame or any other step", () => {
    expect(turnBeginCue(null, "untap", true)).toBeNull();
    expect(turnBeginCue("untap", "upkeep", true)).toBeNull();
    expect(turnBeginCue("untap", "untap", true)).toBeNull();
  });
});
