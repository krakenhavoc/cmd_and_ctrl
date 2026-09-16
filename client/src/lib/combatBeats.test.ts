import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  BEAT_CUE_HOLD_MS,
  BEAT_EFFECT_MS,
  BEAT_PAUSE_MS,
  BeatSequencer,
  arrowGeometry,
  arrowIDsFor,
  arrowRender,
  beatMode,
  cueAnchor,
  cueDetail,
  emptyBeatTracker,
  ghostGeometry,
  keepArrowCache,
  planFrame,
  scaledMs,
  schedule,
  splitBeats,
  track,
  type BeatMode,
  type BeatTracker,
  type CachedArrow,
  type ScheduledCue,
} from "./combatBeats";
import type { LogEvent } from "./protocol";

// ---- Log builders ----

const TURN = 5;

function step(seq: number, name: string, turn = TURN): LogEvent {
  return { seq, kind: "step", step: name, turn, seat: 0, text: name };
}
function attack(seq: number, card: string, seat = 1): LogEvent {
  return { seq, kind: "attack", turn: TURN, seat: 0, card_id: card, target_seat: seat, text: "" };
}
function block(seq: number, blocker: string, attacker: string, turn = TURN): LogEvent {
  return { seq, kind: "block", turn, seat: 1, card_id: blocker, target: attacker, text: "" };
}
function dmgCard(
  seq: number,
  source: string,
  target: string,
  amount: number,
  combatStep?: "first_strike" | "regular",
): LogEvent {
  return {
    seq,
    kind: "damage",
    turn: TURN,
    seat: 0,
    card_id: source,
    target,
    amount,
    combat: true,
    combat_step: combatStep,
    text: `${source} dealt ${amount} combat damage to ${target}`,
  };
}
function dmgSeat(
  seq: number,
  source: string,
  seat: number,
  amount: number,
  combatStep?: "first_strike" | "regular",
): LogEvent {
  return {
    seq,
    kind: "damage",
    turn: TURN,
    seat: 0,
    card_id: source,
    target_seat: seat,
    amount,
    combat: true,
    combat_step: combatStep,
    text: `${source} dealt ${amount} combat damage to seat ${seat}`,
  };
}
function dies(seq: number, card: string): LogEvent {
  return {
    seq,
    kind: "zone",
    turn: TURN,
    seat: 0,
    card_id: card,
    old_zone: "battlefield",
    new_zone: "graveyard",
    text: `${card} died`,
  };
}
function life(seq: number, seat: number, amount: number): LogEvent {
  return { seq, kind: "life", turn: TURN, seat, amount, text: "" };
}

// declaredCombat is a turn up to declared blocks: Ace attacks seat 1,
// the Bears block it.
function declaredCombat(): LogEvent[] {
  return [
    step(100, "declare_attackers"),
    attack(101, "ace"),
    step(110, "declare_blockers"),
    block(111, "bears", "ace"),
  ];
}

// primedOn primes a tracker on a frame, as the first frame after join.
function primedOn(log: LogEvent[]): BeatTracker {
  return planFrame(emptyBeatTracker(), log, { mode: "full", speed: 1 }).tracker;
}

function plan(tracker: BeatTracker, log: LogEvent[], mode: BeatMode = "full", speed = 1) {
  return planFrame(tracker, log, { mode, speed });
}

// ---- Settings ----

describe("beatMode", () => {
  it("is full only when animations, damagePopups and motion are all allowed", () => {
    expect(beatMode(true, true, false)).toBe("full");
  });
  it.each([
    [false, true, false],
    [true, false, false],
    [true, true, true],
    [false, false, true],
  ])("is still for enabled=%s damagePopups=%s reduceMotion=%s", (a, d, r) => {
    expect(beatMode(a, d, r)).toBe("still");
  });
});

describe("scaledMs", () => {
  it("multiplies by animations.speed", () => {
    expect(scaledMs(400, 1)).toBe(400);
    expect(scaledMs(400, 2)).toBe(800);
    expect(scaledMs(400, 0.5)).toBe(200);
  });
  it("falls back to speed 1 for a nonsense speed", () => {
    expect(scaledMs(400, 0)).toBe(400);
    expect(scaledMs(400, Number.NaN)).toBe(400);
  });
});

// ---- Tracking ----

describe("track", () => {
  it("primes on the first frame: nothing is fresh", () => {
    const log = [
      ...declaredCombat(),
      step(120, "combat_damage"),
      dmgCard(121, "ace", "bears", 1, "first_strike"),
    ];
    const r = track(emptyBeatTracker(), log);
    expect(r.primed).toBe(true);
    expect(r.fresh).toEqual([]);
    expect(r.tracker.maxSeq).toBe(121);
  });

  it("returns only entries above the watermark afterwards", () => {
    const t = primedOn(declaredCombat());
    const r = track(t, [...declaredCombat(), step(120, "combat_damage")]);
    expect(r.fresh.map((e) => e.seq)).toEqual([120]);
    expect(r.tracker.maxSeq).toBe(120);
  });

  it("re-primes when asked, even on a primed tracker", () => {
    const t = primedOn(declaredCombat());
    const log = [
      ...declaredCombat(),
      step(120, "combat_damage"),
      dmgCard(121, "ace", "bears", 1, "first_strike"),
    ];
    const r = track(t, log, true);
    expect(r.primed).toBe(true);
    expect(r.fresh).toEqual([]);
  });

  it("treats a lower highest seq as a rewind and cues nothing from it", () => {
    const t = primedOn([...declaredCombat(), step(120, "combat_damage")]);
    const r = track(t, declaredCombat());
    expect(r.rewound).toBe(true);
    expect(r.fresh).toEqual([]);
    expect(r.tracker.maxSeq).toBe(111);
  });
});

// ---- Acceptance criteria, data in / data out ----

describe("AC1: double strike, blocker survives beat 1", () => {
  const before = declaredCombat();
  const after = [
    ...before,
    step(120, "combat_damage"),
    dmgCard(121, "ace", "bears", 1, "first_strike"),
    dmgCard(122, "ace", "bears", 1, "regular"),
    dmgCard(123, "bears", "ace", 2, "regular"),
    dies(124, "bears"),
    dies(125, "ace"),
  ];

  it("splits into two beats, both deaths in beat 2", () => {
    const t = track(primedOn(before), after);
    const { beats } = splitBeats(t.tracker, after, t.fresh);
    expect(beats.map((b) => b.tag)).toEqual(["first_strike", "regular"]);
    expect(beats[0].entries.map((e) => e.seq)).toEqual([121]);
    expect(beats[1].damage.map((e) => e.seq)).toEqual([122, 123]);
    expect(beats[1].entries.map((e) => e.seq)).toEqual([122, 123, 124, 125]);
    expect(beats[1].pauseBefore).toBe(true);
    expect(beats[0].arrowIDs).toEqual(["blk-bears"]);
    expect(beats[1].arrowIDs).toEqual(["blk-bears"]);
  });

  it("schedules beat 2 one pause after beat 1, both labelled", () => {
    const { cues } = plan(primedOn(before), after);
    expect(cues.map((c) => [c.tag, c.atMs, c.label])).toEqual([
      ["first_strike", 0, "First strike"],
      ["regular", BEAT_PAUSE_MS, "Regular damage"],
    ]);
  });

  it("names arrows that come back ghost from the previous frame's cache (AC13)", () => {
    const { cues } = plan(primedOn(before), after);
    const cache = new Set(["atk-ace", "blk-bears"]); // measured at declare_blockers
    const live = new Set<string>(); // both creatures are gone in the end state
    for (const c of cues) {
      for (const id of c.arrowIDs) expect(arrowRender(id, live, cache)).toBe("ghost");
    }
  });
});

describe("AC2: double strike, blocker dies in beat 1 (log-driven, bug A fixed)", () => {
  it("is one beat with no pause", () => {
    const before = [
      step(100, "declare_attackers"),
      attack(101, "ace"),
      block(111, "squire", "ace"),
    ];
    const after = [
      ...before,
      step(120, "combat_damage"),
      dmgCard(121, "ace", "squire", 1, "first_strike"),
      dies(122, "squire"),
    ];
    const { cues } = plan(primedOn(before), after);
    expect(cues).toHaveLength(1);
    expect(cues[0]).toMatchObject({ tag: "first_strike", atMs: 0, label: "First strike" });
    expect(cues[0].entries.map((e) => e.seq)).toEqual([121, 122]);
    expect(arrowRender(cues[0].arrowIDs[0], new Set(), new Set(["blk-squire"]))).toBe("ghost");
  });
});

describe("AC3: first strike", () => {
  it("is one beat; the regular step dealt nothing", () => {
    const before = [
      step(100, "declare_attackers"),
      attack(101, "knight"),
      block(111, "bears", "knight"),
    ];
    const after = [
      ...before,
      step(120, "combat_damage"),
      dmgCard(121, "knight", "bears", 2, "first_strike"),
      dies(122, "bears"),
    ];
    const { cues } = plan(primedOn(before), after);
    expect(cues.map((c) => c.tag)).toEqual(["first_strike"]);
    expect(cues[0].atMs).toBe(0);
    expect(arrowRender(cues[0].arrowIDs[0], new Set(), new Set(["blk-bears"]))).toBe("ghost");
  });
});

describe("AC4: double strike to a player", () => {
  it("is two beats on the live attack arrow, including seat 0", () => {
    const before = [step(100, "declare_attackers"), attack(101, "ace", 0)];
    const after = [
      ...before,
      step(120, "combat_damage"),
      dmgSeat(121, "ace", 0, 1, "first_strike"),
      life(122, 0, -1),
      dmgSeat(123, "ace", 0, 1, "regular"),
      life(124, 0, -1),
    ];
    const { cues } = plan(primedOn(before), after);
    expect(cues.map((c) => [c.tag, c.atMs])).toEqual([
      ["first_strike", 0],
      ["regular", BEAT_PAUSE_MS],
    ]);
    for (const c of cues) {
      expect(c.arrowIDs).toEqual(["atk-ace"]);
      expect(arrowRender("atk-ace", new Set(["atk-ace"]), new Set(["atk-ace"]))).toBe("live");
    }
    expect(cues[0].entries.map((e) => e.seq)).toEqual([121, 122]);
  });
});

describe("AC5: mixed combat", () => {
  it("plays every first-strike entry in beat 1 and every regular entry in beat 2", () => {
    const before = [
      step(100, "declare_attackers"),
      attack(101, "ace"),
      attack(102, "knight"),
      attack(103, "ogre"),
      step(110, "declare_blockers"),
      block(111, "wall", "knight"),
    ];
    const after = [
      ...before,
      step(120, "combat_damage"),
      dmgSeat(121, "ace", 1, 1, "first_strike"),
      dmgCard(122, "knight", "wall", 2, "first_strike"),
      dmgSeat(123, "ace", 1, 1, "regular"),
      dmgSeat(124, "ogre", 1, 3, "regular"),
      dmgCard(125, "wall", "knight", 1, "regular"),
    ];
    const { cues } = plan(primedOn(before), after);
    expect(cues.map((c) => c.tag)).toEqual(["first_strike", "regular"]);
    expect(cues[0].arrowIDs).toEqual(["atk-ace", "blk-wall"]);
    expect(cues[1].arrowIDs).toEqual(["atk-ace", "atk-ogre", "blk-wall"]);
    const lastFirst = Math.max(...cues[0].entries.map((e) => e.seq));
    const firstRegular = Math.min(...cues[1].entries.map((e) => e.seq));
    expect(lastFirst).toBeLessThan(firstRegular);
  });
});

describe("AC6: no first strike anywhere", () => {
  it("schedules nothing for untagged combat damage", () => {
    const before = [
      step(100, "declare_attackers"),
      attack(101, "ogre"),
      block(111, "wall", "ogre"),
    ];
    const after = [
      ...before,
      step(120, "combat_damage"),
      dmgCard(121, "ogre", "wall", 5),
      dmgCard(122, "wall", "ogre", 2),
      dies(123, "wall"),
    ];
    expect(plan(primedOn(before), after).cues).toEqual([]);
  });

  it("schedules nothing when the untagged damage lands from a later prompt frame", () => {
    const frameN = [...declaredCombat(), block(112, "bears2", "ace"), step(120, "combat_damage")];
    const frameN1 = [...frameN, dmgCard(121, "ace", "bears", 3), dmgCard(122, "ace", "bears2", 2)];
    const t = plan(primedOn(declaredCombat()), frameN).tracker;
    expect(plan(t, frameN1).cues).toEqual([]);
  });
});

describe("AC7: still mode (helper half)", () => {
  const before = declaredCombat();
  const after = [
    ...before,
    step(120, "combat_damage"),
    dmgCard(121, "ace", "bears", 1, "first_strike"),
    dmgCard(122, "ace", "bears", 1, "regular"),
  ];

  it("has no motion, keeps the labels, and holds beat 1 for the scaled pause", () => {
    const { cues } = plan(primedOn(before), after, "still", 1);
    expect(cues.every((c) => c.motion === false)).toBe(true);
    expect(cues.map((c) => [c.label, c.atMs])).toEqual([
      ["First strike", 0],
      ["Regular damage", BEAT_PAUSE_MS],
    ]);
    // Beat 1's label is still up when beat 2's joins it.
    expect(cues[0].hideAtMs).toBeGreaterThan(cues[1].atMs);
  });

  it("full mode has motion with a speed-scaled effect duration", () => {
    const { cues } = plan(primedOn(before), after, "full", 2);
    expect(cues.every((c) => c.motion)).toBe(true);
    expect(cues[0].effectMs).toBe(BEAT_EFFECT_MS * 2);
  });
});

describe("AC8: reconnect primes", () => {
  it("replays no beats from a first frame that lands mid combat_damage", () => {
    const log = [
      ...declaredCombat(),
      step(120, "combat_damage"),
      dmgCard(121, "ace", "bears", 1, "first_strike"),
      dmgCard(122, "ace", "bears", 1, "regular"),
    ];
    let p = planFrame(emptyBeatTracker(), log, { mode: "full", speed: 1 });
    expect(p.primed).toBe(true);
    expect(p.cues).toEqual([]);
    // The same frame again, or one that only adds a death, cues nothing.
    p = plan(p.tracker, log);
    expect(p.cues).toEqual([]);
    p = plan(p.tracker, [...log, dies(123, "bears")]);
    expect(p.cues).toEqual([]);
  });

  it("replays no beats after a requested re-prime (reconnect with the board mounted)", () => {
    const t = primedOn(declaredCombat());
    const missed = [
      ...declaredCombat(),
      step(120, "combat_damage"),
      dmgCard(121, "ace", "bears", 1, "first_strike"),
    ];
    const p = planFrame(t, missed, { mode: "full", speed: 1, reprime: true });
    expect(p.primed).toBe(true);
    expect(p.cues).toEqual([]);
  });
});

describe("AC9: prevention skips beat 1", () => {
  it("is one regular beat with no pause", () => {
    const before = declaredCombat();
    const after = [
      ...before,
      step(120, "combat_damage"),
      // Every first-strike point prevented: no first_strike entry.
      dmgCard(122, "bears", "ace", 2, "regular"),
    ];
    const { cues } = plan(primedOn(before), after);
    expect(cues.map((c) => [c.tag, c.atMs, c.label])).toEqual([["regular", 0, "Regular damage"]]);
  });
});

describe("AC10: beats across frames", () => {
  const frameN = [
    ...declaredCombat(),
    block(112, "bears2", "ace"),
    step(120, "combat_damage"),
    dmgCard(121, "bears", "ace", 1, "first_strike"),
  ];

  it("continues beat 1 with no pause and no label, then pauses before beat 2", () => {
    let p = plan(primedOn(declaredCombat()), frameN);
    expect(p.cues.map((c) => [c.tag, c.atMs, c.label])).toEqual([
      ["first_strike", 0, "First strike"],
    ]);
    const frameN1 = [
      ...frameN,
      dmgCard(122, "ace", "bears", 2, "first_strike"),
      dmgCard(123, "ace", "bears2", 1, "first_strike"),
      dies(124, "bears"),
      dmgCard(125, "bears2", "ace", 1, "regular"),
    ];
    p = plan(p.tracker, frameN1);
    expect(p.cues.map((c) => [c.tag, c.atMs, c.label])).toEqual([
      ["first_strike", 0, null],
      ["regular", BEAT_PAUSE_MS, "Regular damage"],
    ]);
    expect(p.cues[0].entries.map((e) => e.seq)).toEqual([122, 123, 124]);
  });

  it("plays a later frame with only regular entries at once", () => {
    let p = plan(primedOn(declaredCombat()), frameN);
    p = plan(p.tracker, [...frameN, dmgCard(125, "bears2", "ace", 1, "regular")]);
    expect(p.cues.map((c) => [c.tag, c.atMs, c.label])).toEqual([["regular", 0, "Regular damage"]]);
  });

  it("shows a first-strike entry that arrives after beat 2 started at once, not reordered", () => {
    // #702 today: regular damage lands before the first-strike prompt's.
    const regularFirst = [
      ...declaredCombat(),
      step(120, "combat_damage"),
      dmgCard(121, "bears", "ace", 2, "regular"),
      dmgCard(122, "ace", "bears", 3, "first_strike"),
    ];
    const same = plan(primedOn(declaredCombat()), regularFirst);
    expect(same.cues.map((c) => [c.tag, c.atMs])).toEqual([
      ["regular", 0],
      ["first_strike", 0],
    ]);

    const frameA = [
      ...declaredCombat(),
      step(120, "combat_damage"),
      dmgCard(121, "bears", "ace", 2, "regular"),
    ];
    let p = plan(primedOn(declaredCombat()), frameA);
    p = plan(p.tracker, [...frameA, dmgCard(122, "ace", "bears", 3, "first_strike")]);
    expect(p.cues.map((c) => [c.tag, c.atMs, c.label])).toEqual([
      ["first_strike", 0, "First strike"],
    ]);
  });

  it("gives entries before the frame's first tagged damage no beat", () => {
    const frameA = [...declaredCombat(), step(120, "combat_damage")];
    let p = plan(primedOn(declaredCombat()), frameA);
    p = plan(p.tracker, [...frameA, dies(121, "bears"), life(122, 1, 3)]);
    expect(p.cues).toEqual([]);
  });

  it("never mixes the beats of two combat_damage steps in one turn", () => {
    const log = [
      ...declaredCombat(),
      step(120, "combat_damage"),
      dmgSeat(121, "ace", 1, 1, "first_strike"),
      step(130, "end_combat"),
      step(140, "declare_attackers"),
      attack(141, "ace"),
      step(150, "combat_damage"),
      dmgSeat(151, "ace", 1, 1, "first_strike"),
      dmgSeat(152, "ace", 1, 1, "regular"),
    ];
    const { cues } = plan(primedOn(declaredCombat()), log);
    expect(cues.map((c) => [c.stepSeq, c.tag, c.atMs, c.label])).toEqual([
      [120, "first_strike", 0, "First strike"],
      [150, "first_strike", 0, "First strike"],
      [150, "regular", BEAT_PAUSE_MS, "Regular damage"],
    ]);
  });

  it("gives entries whose step entry fell out of the log window no beat", () => {
    const t = primedOn([block(111, "bears", "ace")]);
    const { cues } = plan(t, [
      block(111, "bears", "ace"),
      dmgCard(121, "ace", "bears", 1, "first_strike"),
    ]);
    expect(cues).toEqual([]);
  });
});

describe("AC11: undo", () => {
  it("cues entries replayed at reused seq numbers after a rewind", () => {
    const before = declaredCombat();
    const dealt = [
      ...before,
      step(120, "combat_damage"),
      dmgCard(121, "ace", "bears", 1, "first_strike"),
    ];
    let p = plan(primedOn(before), dealt);
    expect(p.cues).toHaveLength(1);

    // Undo back to declared blocks: a rewind, which cues nothing.
    p = plan(p.tracker, before);
    expect(p.cues).toEqual([]);

    // The replay reuses 120 and 121 and is cued again, with its label.
    p = plan(p.tracker, dealt);
    expect(p.cues.map((c) => [c.tag, c.label])).toEqual([["first_strike", "First strike"]]);
  });

  it("forgets a cued beat inside the step when the undo lands mid-step", () => {
    const frameA = [
      ...declaredCombat(),
      step(120, "combat_damage"),
      dmgCard(121, "ace", "bears", 1, "first_strike"),
    ];
    let p = plan(primedOn(declaredCombat()), frameA);
    const mid = [...declaredCombat(), step(120, "combat_damage")];
    p = plan(p.tracker, mid);
    p = plan(p.tracker, frameA);
    expect(p.cues[0].label).toBe("First strike");
  });
});

// ---- Timing on real (fake) timers ----

describe("BeatSequencer", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  const before = declaredCombat();
  const both = [
    ...before,
    step(120, "combat_damage"),
    dmgCard(121, "ace", "bears", 1, "first_strike"),
    dmgCard(122, "ace", "bears", 1, "regular"),
  ];

  function recorder() {
    const fired: { at: number; cue: ScheduledCue }[] = [];
    const hidden: { at: number; stepSeq: number }[] = [];
    const t0 = Date.now();
    const seq = new BeatSequencer({
      onCue: (cue) => fired.push({ at: Date.now() - t0, cue }),
      onHide: (stepSeq) => hidden.push({ at: Date.now() - t0, stepSeq }),
    });
    return { fired, hidden, seq };
  }

  describe("AC12: beat 2 is one scaled pause after beat 1", () => {
    it.each([
      ["full", 1, 400],
      ["full", 2, 800],
      ["full", 0.5, 200],
      ["still", 1, 400],
      ["still", 2, 800],
      ["still", 0.5, 200],
    ] as const)("%s mode at speed %s → %s ms", (mode, speed, expected) => {
      const { fired, seq } = recorder();
      seq.play(plan(primedOn(before), both, mode, speed).cues);
      vi.advanceTimersByTime(expected - 1);
      expect(fired.map((f) => f.cue.tag)).toEqual(["first_strike"]);
      vi.advanceTimersByTime(1);
      expect(fired.map((f) => [f.cue.tag, f.at])).toEqual([
        ["first_strike", 0],
        ["regular", expected],
      ]);
    });
  });

  it("schedules nothing and fires nothing for untagged combat (AC6)", () => {
    const { fired, hidden, seq } = recorder();
    const untagged = [...before, step(120, "combat_damage"), dmgCard(121, "ace", "bears", 2)];
    seq.play(plan(primedOn(before), untagged).cues);
    expect(seq.pending).toBe(0);
    vi.runAllTimers();
    expect(fired).toEqual([]);
    expect(hidden).toEqual([]);
  });

  it("AC14: an end_combat frame before beat 2 does not remove beat 2", () => {
    const { fired, seq } = recorder();
    let p = plan(primedOn(before), both);
    seq.play(p.cues);
    vi.advanceTimersByTime(100);
    expect(fired).toHaveLength(1);

    // A bot passes: the game is in end_combat 100 ms later.
    p = plan(p.tracker, [...both, step(130, "end_combat")]);
    seq.play(p.cues);
    expect(seq.pending).toBe(1);
    expect(keepArrowCache("end_combat", seq.pending)).toBe(true);

    vi.advanceTimersByTime(BEAT_PAUSE_MS - 100);
    expect(fired.map((f) => f.cue.tag)).toEqual(["first_strike", "regular"]);
    expect(seq.pending).toBe(0);
    expect(keepArrowCache("end_combat", seq.pending)).toBe(false);
  });

  it("brings the step's text cue down once, after the last beat plus the hold", () => {
    const { hidden, seq } = recorder();
    seq.play(plan(primedOn(before), both, "still", 1).cues);
    vi.advanceTimersByTime(BEAT_PAUSE_MS + BEAT_CUE_HOLD_MS - 1);
    expect(hidden).toEqual([]);
    vi.advanceTimersByTime(1);
    expect(hidden).toEqual([{ at: BEAT_PAUSE_MS + BEAT_CUE_HOLD_MS, stepSeq: 120 }]);
    vi.runAllTimers();
    expect(hidden).toHaveLength(1);
  });

  it("pushes the hide out when a later frame adds a labelled beat to the step", () => {
    const { hidden, seq } = recorder();
    const frameA = [
      ...before,
      step(120, "combat_damage"),
      dmgCard(121, "ace", "bears", 1, "first_strike"),
    ];
    let p = plan(primedOn(before), frameA);
    seq.play(p.cues);
    vi.advanceTimersByTime(1000);
    p = plan(p.tracker, [...frameA, dmgCard(122, "bears", "ace", 2, "regular")]);
    seq.play(p.cues);
    vi.advanceTimersByTime(BEAT_CUE_HOLD_MS - 1000);
    expect(hidden).toEqual([]);
    vi.runAllTimers();
    expect(hidden).toEqual([{ at: 1000 + BEAT_CUE_HOLD_MS, stepSeq: 120 }]);
  });

  it("cancels everything on dispose", () => {
    const { fired, hidden, seq } = recorder();
    seq.play(plan(primedOn(before), both).cues);
    seq.dispose();
    vi.runAllTimers();
    expect(fired).toEqual([]);
    expect(hidden).toEqual([]);
    expect(seq.pending).toBe(0);
  });
});

describe("schedule", () => {
  it("is empty for no beats", () => {
    expect(schedule([], "full", 1)).toEqual([]);
  });
});

// ---- Arrows ----

describe("arrowIDsFor", () => {
  it("names the block arrow from either side of the block", () => {
    const log = declaredCombat();
    expect(arrowIDsFor([dmgCard(121, "ace", "bears", 1)], log)).toEqual(["blk-bears"]);
    expect(arrowIDsFor([dmgCard(122, "bears", "ace", 2)], log)).toEqual(["blk-bears"]);
  });

  it("uses the latest block entry per blocker in the same turn", () => {
    const log = [
      block(111, "bears", "ace", TURN - 1),
      block(112, "bears", "ogre"),
      block(113, "bears", "ace"),
    ];
    expect(arrowIDsFor([dmgCard(121, "ace", "bears", 1)], log)).toEqual(["blk-bears"]);
    const stale = [block(111, "bears", "ace", TURN - 1)];
    expect(arrowIDsFor([dmgCard(121, "ace", "bears", 1)], stale)).toEqual([]);
  });

  it("names no arrow for an attack on a planeswalker or a redacted source", () => {
    const log = [attack(101, "ace")];
    expect(arrowIDsFor([dmgCard(121, "ace", "walker", 2)], log)).toEqual([]);
    const redacted: LogEvent = { ...dmgSeat(121, "ace", 1, 2), card_id: undefined };
    expect(arrowIDsFor([redacted], log)).toEqual([]);
  });

  it("dedupes", () => {
    expect(arrowIDsFor([dmgSeat(121, "ace", 1, 1), dmgSeat(122, "ace", 1, 1)], [])).toEqual([
      "atk-ace",
    ]);
  });
});

describe("arrowRender (AC13 helper half)", () => {
  it("is live when live, ghost when only cached, none otherwise", () => {
    const live = new Set(["atk-a"]);
    const cached = new Map<string, unknown>([
      ["atk-a", {}],
      ["blk-b", {}],
    ]);
    expect(arrowRender("atk-a", live, cached)).toBe("live");
    expect(arrowRender("blk-b", live, cached)).toBe("ghost");
    expect(arrowRender("blk-c", live, cached)).toBe("none");
  });
});

describe("ghostGeometry", () => {
  const cached: CachedArrow = {
    kind: "block",
    fromCardID: "bears",
    toCardID: "ace",
    geo: arrowGeometry({ x: 10, y: 200 }, { x: 110, y: 100 }),
    board: { w: 800, h: 600 },
  };

  it("draws from the cache when both endpoints are gone", () => {
    expect(ghostGeometry(cached, { w: 800, h: 600 }, null, null)).toEqual(cached.geo);
  });

  it("re-measures an endpoint still on the board", () => {
    const g = ghostGeometry(cached, { w: 800.5, h: 600 }, null, { x: 300, y: 50 });
    expect(g).toEqual(arrowGeometry({ x: 10, y: 200 }, { x: 300, y: 50 }));
  });

  it("drops the ghost when the board changed size", () => {
    expect(ghostGeometry(cached, { w: 820, h: 600 }, null, null)).toBeNull();
    expect(ghostGeometry(cached, { w: 800, h: 560 }, null, null)).toBeNull();
  });
});

describe("cueAnchor", () => {
  it("sits at the board centre when no arrow could be drawn", () => {
    expect(cueAnchor([], { w: 800, h: 600 })).toEqual({ x: 400, y: 300 });
  });

  it("sits at the curve midpoint of the beat's arrows", () => {
    const g = { x1: 100, y1: 300, x2: 300, y2: 300, cx: 200, cy: 200 };
    expect(cueAnchor([g], { w: 800, h: 600 })).toEqual({ x: 200, y: 250 });
  });

  it("stays inside the board", () => {
    const g = { x1: 0, y1: 0, x2: 0, y2: 0, cx: 0, cy: 0 };
    const p = cueAnchor([g], { w: 800, h: 600 });
    expect(p.x).toBeGreaterThan(0);
    expect(p.y).toBeGreaterThan(0);
  });
});

describe("keepArrowCache", () => {
  it("keeps geometry through combat and while cues are pending", () => {
    expect(keepArrowCache("declare_attackers", 0)).toBe(true);
    expect(keepArrowCache("combat_damage", 0)).toBe(true);
    expect(keepArrowCache("end_combat", 1)).toBe(true);
    expect(keepArrowCache("end_combat", 0)).toBe(false);
    expect(keepArrowCache(undefined, 0)).toBe(false);
  });
});

describe("cueDetail", () => {
  it("joins the beat's log lines", () => {
    expect(cueDetail({ entries: [dmgSeat(1, "ace", 1, 1), dies(2, "bears")] })).toBe(
      "ace dealt 1 combat damage to seat 1; bears died",
    );
  });
});
