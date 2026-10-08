// ADR 0134: what moves when combat damage lands, as plain functions.
// The rendered half is combatStrikes.render.test.ts.

import { describe, expect, it } from "vitest";

import {
  LUNGE_MAX_FRACTION,
  LUNGE_MIN_FRACTION,
  STRIKE_CONTACT_OVERLAP_PX,
  STRIKE_COPY_CAP,
  STRIKE_LATE_MS,
  CRUMBLE_COLS,
  CRUMBLE_MAX_DELAY,
  CRUMBLE_ROWS,
  STREAK_MIN_PX,
  aimBox,
  combatMotion,
  crumbleShards,
  crumbleTiming,
  faceVector,
  lethalHit,
  lungeVector,
  streakSegment,
  strikeLate,
  strikeTimeline,
  strikesFor,
  stepChangePlaysCombatSound,
  strikesScheduled,
  usableTile,
  type CachedTile,
  type StrikeBox,
} from "./combatStrikes";
import type { LogEvent } from "./protocol";

// Seat 0 attacks seat 1 throughout: the step entries carry the active
// seat (CR 506.2).
const TURN = 7;
const ATK = 0;
const DEF = 1;

function step(seq: number, name: string, seat = ATK): LogEvent {
  return { seq, kind: "step", step: name, turn: TURN, seat, text: name };
}
function block(seq: number, blocker: string, attacker: string): LogEvent {
  return {
    seq,
    kind: "block",
    turn: TURN,
    seat: DEF,
    card_id: blocker,
    target: attacker,
    text: "",
  };
}
function hitCard(seq: number, seat: number, source: string, target: string, amount = 2): LogEvent {
  return {
    seq,
    kind: "damage",
    turn: TURN,
    seat,
    card_id: source,
    target,
    amount,
    combat: true,
    text: "",
  };
}
function hitSeat(seq: number, seat: number, source: string, targetSeat: number): LogEvent {
  return {
    seq,
    kind: "damage",
    turn: TURN,
    seat,
    card_id: source,
    target_seat: targetSeat,
    amount: 3,
    combat: true,
    text: "",
  };
}
function dies(seq: number, card: string): LogEvent {
  return {
    seq,
    kind: "zone",
    turn: TURN,
    seat: -1,
    card_id: card,
    old_zone: "battlefield",
    new_zone: "graveyard",
    text: "",
  };
}

// The beat's members are what follows the damage step entry.
function beat(log: LogEvent[]): { entries: LogEvent[] } {
  const at = log.findIndex((e) => e.kind === "step" && e.step === "combat_damage");
  return { entries: log.slice(at + 1) };
}

describe("strikesFor: sides", () => {
  it("an unblocked attacker lunges at the defending player, whose avatar shakes", () => {
    const log = [
      step(100, "declare_attackers"),
      step(120, "combat_damage"),
      hitSeat(121, ATK, "ogre", DEF),
    ];
    const plan = strikesFor(beat(log), log);
    expect(plan.movers).toEqual([
      {
        cardID: "ogre",
        side: "attacker",
        aim: [{ kind: "seat", seat: DEF }],
        dies: false,
        flash: false,
        lethal: false,
        through: [],
      },
    ]);
    expect(plan.impacts).toEqual([
      { target: { kind: "seat", seat: DEF }, dies: false, lethal: false },
    ]);
  });

  it("reads the sides from the step entry's seat, with the block entries in the window", () => {
    const log = [
      step(100, "declare_attackers"),
      step(110, "declare_blockers"),
      block(111, "wall", "ogre"),
      step(120, "combat_damage"),
      hitCard(121, ATK, "ogre", "wall", 5),
      hitCard(122, DEF, "wall", "ogre", 2),
    ];
    const plan = strikesFor(beat(log), log);
    // The attacker lunges and flashes (its blocker hit it back); the
    // blocker braces and shakes.
    expect(plan.movers.map((m) => [m.cardID, m.side, m.flash])).toEqual([
      ["ogre", "attacker", true],
    ]);
    expect(plan.impacts).toEqual([
      { target: { kind: "card", cardID: "wall" }, dies: false, lethal: false },
    ]);
  });

  it("reads the same sides from the step entry's seat when the block entries have left the window", () => {
    const log = [
      step(120, "combat_damage"),
      hitCard(121, ATK, "ogre", "wall", 5),
      hitCard(122, DEF, "wall", "ogre", 2),
    ];
    const plan = strikesFor(beat(log), log);
    expect(plan.movers.map((m) => [m.cardID, m.side])).toEqual([["ogre", "attacker"]]);
    expect(plan.impacts.map((i) => i.target)).toEqual([{ kind: "card", cardID: "wall" }]);
  });

  it("falls back to the block entries when no step entry is in the window", () => {
    const log = [
      block(111, "wall", "ogre"),
      hitCard(121, ATK, "ogre", "wall", 5),
      hitCard(122, DEF, "wall", "ogre", 2),
    ];
    const plan = strikesFor({ entries: log.slice(1) }, log);
    expect(plan.movers.map((m) => [m.cardID, m.side])).toEqual([["ogre", "attacker"]]);
  });
});

describe("strikesFor: several blockers (CR 510.1c)", () => {
  it("gives one mover aimed at every blocker it damaged, and they all shake", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "bear1", "titan"),
      block(112, "bear2", "titan"),
      step(120, "combat_damage"),
      hitCard(121, ATK, "titan", "bear1", 3),
      hitCard(122, ATK, "titan", "bear2", 3),
    ];
    const plan = strikesFor(beat(log), log);
    expect(plan.movers).toHaveLength(1);
    expect(plan.movers[0].aim).toEqual([
      { kind: "card", cardID: "bear1" },
      { kind: "card", cardID: "bear2" },
    ]);
    expect(plan.impacts.map((i) => i.target)).toEqual([
      { kind: "card", cardID: "bear1" },
      { kind: "card", cardID: "bear2" },
    ]);
  });

  it("aims at the centroid of the blockers' centres, with their average size", () => {
    const a: StrikeBox = { x: 100, y: 300, w: 80, h: 112 };
    const b: StrikeBox = { x: 300, y: 300, w: 60, h: 100 };
    expect(aimBox([a, b])).toEqual({ x: 200, y: 300, w: 70, h: 106 });
    expect(aimBox([])).toBeNull();
  });
});

describe("strikesFor: trample (CR 702.19b)", () => {
  it("lunges at the blocker, and the player's avatar shakes at the same contact", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "bear", "wurm"),
      step(120, "combat_damage"),
      hitCard(121, ATK, "wurm", "bear", 2),
      hitSeat(122, ATK, "wurm", DEF),
    ];
    const plan = strikesFor(beat(log), log);
    expect(plan.movers[0].aim).toEqual([{ kind: "card", cardID: "bear" }]);
    expect(plan.impacts.map((i) => i.target)).toEqual([
      { kind: "card", cardID: "bear" },
      { kind: "seat", seat: DEF },
    ]);
  });

  it("names the player its excess hit as the streak's end (ADR 0134 question 5)", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "bear1", "wurm"),
      block(112, "bear2", "wurm"),
      step(120, "combat_damage"),
      hitCard(121, ATK, "wurm", "bear1", 2),
      hitCard(122, ATK, "wurm", "bear2", 2),
      hitSeat(123, ATK, "wurm", DEF),
    ];
    const [wurm] = strikesFor(beat(log), log).movers;
    expect(wurm.aim).toHaveLength(2);
    expect(wurm.through).toEqual([{ kind: "seat", seat: DEF }]);
  });

  it("draws no streak for an unblocked attacker, or with no block entries to tell", () => {
    const unblocked = [step(120, "combat_damage"), hitSeat(121, ATK, "wurm", DEF)];
    expect(strikesFor(beat(unblocked), unblocked).movers[0].through).toEqual([]);
    const noBlocks = [
      step(120, "combat_damage"),
      hitCard(121, ATK, "wurm", "bear", 2),
      hitSeat(122, ATK, "wurm", DEF),
    ];
    expect(strikesFor(beat(noBlocks), noBlocks).movers[0].through).toEqual([]);
  });

  it("does not aim at a planeswalker its excess hit, which only shakes", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "bear", "wurm"),
      step(120, "combat_damage"),
      hitCard(121, ATK, "wurm", "bear", 2),
      hitCard(122, ATK, "wurm", "jace", 4),
    ];
    const plan = strikesFor(beat(log), log);
    expect(plan.movers[0].aim).toEqual([{ kind: "card", cardID: "bear" }]);
    expect(plan.movers[0].through).toEqual([{ kind: "card", cardID: "jace" }]);
    expect(plan.impacts.map((i) => i.target)).toContainEqual({ kind: "card", cardID: "jace" });
  });
});

describe("strikesFor: blockers (ADR 0134 question 6)", () => {
  it("a first-strike blocker lunges in a beat where its attacker dealt it nothing", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "knight", "ogre"),
      step(120, "first_strike_damage"),
      hitCard(121, DEF, "knight", "ogre", 2),
    ];
    const plan = strikesFor({ entries: log.slice(3) }, log);
    expect(plan.movers).toEqual([
      {
        cardID: "knight",
        side: "blocker",
        aim: [{ kind: "card", cardID: "ogre" }],
        dies: false,
        flash: false,
        lethal: false,
        through: [],
      },
    ]);
    // The attacker is not moving in this beat, so it shakes.
    expect(plan.impacts).toEqual([
      { target: { kind: "card", cardID: "ogre" }, dies: false, lethal: false },
    ]);
  });

  it("a blocker facing a lunging attacker braces: no lunge, it shakes", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "wall", "ogre"),
      step(120, "combat_damage"),
      hitCard(121, ATK, "ogre", "wall", 5),
      hitCard(122, DEF, "wall", "ogre", 2),
    ];
    const plan = strikesFor(beat(log), log);
    expect(plan.movers.some((m) => m.cardID === "wall")).toBe(false);
    expect(plan.impacts.map((i) => i.target)).toEqual([{ kind: "card", cardID: "wall" }]);
  });

  it("a blocker facing a 0-power attacker lunges at it", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "bear", "wall"),
      step(120, "combat_damage"),
      hitCard(121, DEF, "bear", "wall", 2),
    ];
    const plan = strikesFor(beat(log), log);
    expect(plan.movers.map((m) => [m.cardID, m.side])).toEqual([["bear", "blocker"]]);
  });
});

describe("strikesFor: planeswalkers and battles", () => {
  it("an unblocked attacker on a planeswalker aims at its tile, a card target", () => {
    const log = [step(120, "combat_damage"), hitCard(121, ATK, "ogre", "jace", 3)];
    const plan = strikesFor(beat(log), log);
    expect(plan.movers[0].aim).toEqual([{ kind: "card", cardID: "jace" }]);
    expect(plan.impacts).toEqual([
      { target: { kind: "card", cardID: "jace" }, dies: false, lethal: false },
    ]);
  });
});

describe("strikesFor: deaths (CR 704.3)", () => {
  it("come from the beat's own zone entries", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "bear", "ogre"),
      step(120, "combat_damage"),
      hitCard(121, ATK, "ogre", "bear", 3),
      hitCard(122, DEF, "bear", "ogre", 2),
      dies(123, "bear"),
      dies(124, "ogre"),
    ];
    const plan = strikesFor(beat(log), log);
    expect(plan.movers).toEqual([
      expect.objectContaining({ cardID: "ogre", dies: true, flash: true, lethal: true }),
    ]);
    expect(plan.impacts).toEqual([
      { target: { kind: "card", cardID: "bear" }, dies: true, lethal: true },
    ]);
  });

  it("does not count a zone change from somewhere else", () => {
    const log = [
      step(120, "combat_damage"),
      hitCard(121, ATK, "ogre", "jace", 3),
      { ...dies(122, "jace"), old_zone: "hand" },
    ];
    expect(strikesFor(beat(log), log).impacts[0].dies).toBe(false);
  });
});

describe("the lethal hit (ADR 0134 PR 2)", () => {
  function eliminated(seq: number, seat: number, cause: string): LogEvent {
    return { seq, kind: "eliminated", turn: TURN, seat, cause, text: "" };
  }

  it("is a card target the beat killed, and only that one", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "bear1", "titan"),
      block(112, "bear2", "titan"),
      step(120, "combat_damage"),
      hitCard(121, ATK, "titan", "bear1", 2),
      hitCard(122, ATK, "titan", "bear2", 1),
      dies(123, "bear1"),
    ];
    const plan = strikesFor(beat(log), log);
    expect(plan.impacts.map((i) => [i.target, i.lethal])).toEqual([
      [{ kind: "card", cardID: "bear1" }, true],
      [{ kind: "card", cardID: "bear2" }, false],
    ]);
  });

  it("is a player the beat eliminated: by life, by commander damage, by poison", () => {
    for (const cause of ["life", "commander_damage", "poison"]) {
      const log = [
        step(120, "combat_damage"),
        hitSeat(121, ATK, "ogre", DEF),
        eliminated(122, DEF, cause),
      ];
      expect(strikesFor(beat(log), log).impacts[0].lethal, cause).toBe(true);
    }
  });

  it("is not a player who survived the hit, or one who conceded after it", () => {
    const hitOnly = [step(120, "combat_damage"), hitSeat(121, ATK, "ogre", DEF)];
    expect(strikesFor(beat(hitOnly), hitOnly).impacts[0].lethal).toBe(false);
    const conceded = [...hitOnly, eliminated(122, DEF, "concede")];
    expect(strikesFor(beat(conceded), conceded).impacts[0].lethal).toBe(false);
  });

  it("is not a player someone else's hit eliminated", () => {
    const log = [
      step(120, "combat_damage"),
      hitSeat(121, ATK, "ogre", DEF),
      eliminated(122, 2, "life"),
    ];
    expect(strikesFor(beat(log), log).impacts[0].lethal).toBe(false);
  });

  it("is the flash on a mover only when a hit back killed it", () => {
    const log = [
      step(110, "declare_blockers"),
      block(111, "bear", "ogre"),
      step(120, "combat_damage"),
      hitCard(121, ATK, "ogre", "bear", 3),
      hitCard(122, DEF, "bear", "ogre", 2),
    ];
    expect(strikesFor(beat(log), log).movers[0].lethal).toBe(false);
    const killed = [...log, dies(123, "ogre")];
    expect(strikesFor(beat(killed), killed).movers[0]).toMatchObject({
      dies: true,
      flash: true,
      lethal: true,
    });
  });

  it("lethalHit reads the beat's deaths for a card and its eliminations for a seat", () => {
    const died = new Set(["bear"]);
    const out = new Set([1]);
    expect(lethalHit({ kind: "card", cardID: "bear" }, died, out)).toBe(true);
    expect(lethalHit({ kind: "card", cardID: "ogre" }, died, out)).toBe(false);
    expect(lethalHit({ kind: "seat", seat: 1 }, died, out)).toBe(true);
    expect(lethalHit({ kind: "seat", seat: 2 }, died, out)).toBe(false);
  });
});

describe("strikesFor: many attackers", () => {
  it(`moves at most ${STRIKE_COPY_CAP} copies, in log order, and every target still shakes`, () => {
    const log: LogEvent[] = [step(120, "combat_damage")];
    for (let i = 0; i < 15; i++) log.push(hitCard(121 + i, ATK, `tok${i}`, `pw${i}`, 1));
    const plan = strikesFor(beat(log), log);
    expect(plan.movers).toHaveLength(STRIKE_COPY_CAP);
    expect(plan.movers.map((m) => m.cardID)).toEqual(
      Array.from({ length: STRIKE_COPY_CAP }, (_, i) => `tok${i}`),
    );
    expect(plan.impacts).toHaveLength(15);
  });

  it("plans nothing for a beat with no combat damage", () => {
    const log = [step(120, "combat_damage"), dies(121, "bear")];
    expect(strikesFor(beat(log), log)).toEqual({ movers: [], impacts: [] });
  });
});

describe("lungeVector", () => {
  const card = (x: number, y: number): StrikeBox => ({ x, y, w: 80, h: 112 });

  it("stops at contact plus the overlap, between the clamps", () => {
    // Straight down 300 px: each card reaches 56 px along the way.
    const v = lungeVector(card(0, 0), card(0, 300));
    const contact = 300 - 56 - 56 + STRIKE_CONTACT_OVERLAP_PX;
    expect(v.travel).toBeCloseTo(contact);
    expect(v.x).toBeCloseTo(0);
    expect(v.y).toBeCloseTo(contact);
  });

  it(`never stops short of ${LUNGE_MIN_FRACTION} of the distance`, () => {
    // Far apart, contact would be most of the way, so it is not the floor;
    // close together, contact is under half way and the floor wins.
    const near = lungeVector(card(0, 0), card(0, 130));
    expect(near.travel).toBeCloseTo(LUNGE_MIN_FRACTION * 130);
  });

  it(`never goes past ${LUNGE_MAX_FRACTION} of the distance`, () => {
    const far = lungeVector({ x: 0, y: 0, w: 4, h: 4 }, { x: 1000, y: 0, w: 4, h: 4 });
    expect(far.travel).toBeCloseTo(LUNGE_MAX_FRACTION * 1000);
  });

  it("uses each box's extent along the diagonal", () => {
    const from = card(0, 0);
    const to = card(300, 400); // d = 500, ux 0.6, uy 0.8
    const half = 0.6 * 40 + 0.8 * 56;
    const v = lungeVector(from, to);
    expect(v.travel).toBeCloseTo(500 - 2 * half + STRIKE_CONTACT_OVERLAP_PX);
    expect(v.x / v.y).toBeCloseTo(0.75);
  });

  it("is zero for two boxes on one centre", () => {
    expect(lungeVector(card(5, 5), card(5, 5))).toEqual({ x: 0, y: 0, travel: 0 });
  });
});

describe("strikeTimeline", () => {
  it("is 180 / 60 / 260 ms at speed 1, contact at the end of the out phase", () => {
    expect(strikeTimeline(1)).toEqual({
      outMs: 180,
      holdMs: 60,
      backMs: 260,
      contactMs: 180,
      totalMs: 500,
      shakeMs: 160,
      lethalShakeMs: 240,
      crumbleMs: 420,
      streakDrawMs: 60,
      streakFadeMs: 220,
    });
  });

  it("scales every phase by speed", () => {
    const t = strikeTimeline(2);
    expect([
      t.outMs,
      t.holdMs,
      t.backMs,
      t.totalMs,
      t.shakeMs,
      t.lethalShakeMs,
      t.crumbleMs,
      t.streakDrawMs,
      t.streakFadeMs,
    ]).toEqual([360, 120, 520, 1000, 320, 480, 840, 120, 440]);
    expect(strikeTimeline(0.5).totalMs).toBe(250);
  });

  it("draws trample's streak inside the contact hold", () => {
    const t = strikeTimeline(1);
    expect(t.streakDrawMs).toBeLessThanOrEqual(t.holdMs);
  });

  it("treats a nonsense speed as 1", () => {
    expect(strikeTimeline(0).totalMs).toBe(500);
  });
});

describe("crumbleTiming (ADR 0134 question 7, PR 2)", () => {
  it("crumbles a dead target after its lethal shake at contact", () => {
    // 180 contact + 240 lethal shake, then 420 of crumble.
    expect(crumbleTiming(strikeTimeline(1), "target")).toEqual({ startMs: 420, endMs: 840 });
  });

  it("crumbles a dead mover once its copy is home", () => {
    expect(crumbleTiming(strikeTimeline(1), "mover")).toEqual({ startMs: 500, endMs: 920 });
  });

  it("never starts before the hit has landed (CR 704.3), at any speed", () => {
    for (const speed of [0.5, 1, 1.5, 2]) {
      const t = strikeTimeline(speed);
      expect(crumbleTiming(t, "target").startMs).toBeGreaterThan(t.contactMs);
      expect(crumbleTiming(t, "mover").startMs).toBe(t.totalMs);
      expect(crumbleTiming(t, "mover").endMs - crumbleTiming(t, "mover").startMs).toBe(t.crumbleMs);
    }
  });
});

describe("crumbleShards", () => {
  // The area of a polygon, by the shoelace formula, in percent².
  const area = (pts: { x: number; y: number }[]) =>
    Math.abs(
      pts.reduce((a, p, i) => {
        const q = pts[(i + 1) % pts.length];
        return a + p.x * q.y - q.x * p.y;
      }, 0),
    ) / 2;

  it(`breaks a card into ${CRUMBLE_COLS} × ${CRUMBLE_ROWS} pieces that tile it exactly`, () => {
    const shards = crumbleShards("bear", 80, 112);
    expect(shards).toHaveLength(CRUMBLE_COLS * CRUMBLE_ROWS);
    const total = shards.reduce((a, s) => a + area(s.points), 0);
    expect(total).toBeCloseTo(100 * 100, 6);
    for (const s of shards) {
      for (const p of s.points) {
        expect(p.x).toBeGreaterThanOrEqual(0);
        expect(p.x).toBeLessThanOrEqual(100);
        expect(p.y).toBeGreaterThanOrEqual(0);
        expect(p.y).toBeLessThanOrEqual(100);
      }
      expect(s.clip.startsWith("polygon(")).toBe(true);
    }
  });

  it("is the same for the same card, and different for another", () => {
    expect(crumbleShards("bear", 80, 112)).toEqual(crumbleShards("bear", 80, 112));
    expect(crumbleShards("bear", 80, 112)).not.toEqual(crumbleShards("ogre", 80, 112));
  });

  it("falls down the screen, and the lowest row lets go first", () => {
    const shards = crumbleShards("bear", 80, 112);
    for (const s of shards) {
      expect(s.dy).toBeGreaterThan(0);
      expect(s.delay).toBeGreaterThanOrEqual(0);
      expect(s.delay).toBeLessThanOrEqual(CRUMBLE_MAX_DELAY);
    }
    const bottom = shards.slice(-CRUMBLE_COLS);
    expect(bottom.every((s) => s.delay === 0)).toBe(true);
  });

  it("still falls down the screen from a tapped face", () => {
    // The face is turned 90° clockwise. Turned back onto the screen,
    // every shard still falls down it.
    const shards = crumbleShards("ogre", 80, 112, 90);
    for (const s of shards) {
      const screen = faceVector({ x: s.dx, y: s.dy }, -90);
      expect(screen.y).toBeGreaterThan(0);
    }
  });

  it("faceVector undoes the face's turn", () => {
    expect(faceVector({ x: 0, y: 10 }, 0)).toEqual({ x: 0, y: 10 });
    const v = faceVector({ x: 0, y: 10 }, 90);
    expect(v.x).toBeCloseTo(10);
    expect(v.y).toBeCloseTo(0);
  });
});

describe("streakSegment (trample's streak)", () => {
  const card = (x: number, y: number): StrikeBox => ({ x, y, w: 80, h: 112 });
  const avatar = (x: number, y: number): StrikeBox => ({ x, y, w: 60, h: 60 });

  it("runs from the blockers' edge to the avatar's edge along the centres", () => {
    // Straight up 400 px: the card reaches 56 px, the avatar 30.
    const seg = streakSegment(card(500, 500), avatar(500, 100))!;
    expect(seg.x).toBeCloseTo(500);
    expect(seg.y).toBeCloseTo(444);
    expect(seg.length).toBeCloseTo(400 - 56 - 30);
    expect(seg.angle).toBeCloseTo(-90);
  });

  it("starts at the centroid's edge for several blockers", () => {
    const from = aimBox([card(400, 500), card(600, 500)])!;
    const seg = streakSegment(from, avatar(800, 500))!;
    // Along +x: the centroid at 500 reaches 40 px, the avatar 30.
    expect(seg.x).toBeCloseTo(540);
    expect(seg.length).toBeCloseTo(300 - 40 - 30);
    expect(seg.angle).toBeCloseTo(0);
  });

  it(`draws nothing for a gap under ${STREAK_MIN_PX} px, or one centre`, () => {
    expect(streakSegment(card(0, 0), avatar(0, 100))).toBeNull();
    expect(streakSegment(card(0, 0), card(0, 0))).toBeNull();
  });
});

describe("gates", () => {
  it("combatMotion needs the master switch, the combat toggle and no reduced motion", () => {
    expect(combatMotion({ enabled: true, combat: true, reduceMotion: false })).toBe(true);
    expect(combatMotion({ enabled: false, combat: true, reduceMotion: false })).toBe(false);
    expect(combatMotion({ enabled: true, combat: false, reduceMotion: false })).toBe(false);
    expect(combatMotion({ enabled: true, combat: true, reduceMotion: true })).toBe(false);
  });

  it("a frame that arrives while the page is hidden schedules no strikes", () => {
    expect(strikesScheduled(true, "visible")).toBe(true);
    expect(strikesScheduled(true, "hidden")).toBe(false);
    expect(strikesScheduled(false, "visible")).toBe(false);
    expect(strikesScheduled(true, undefined)).toBe(true);
  });

  it("the step change plays combat_resolve only when no strike will (ADR 0134 §7)", () => {
    expect(stepChangePlaysCombatSound(true, "visible")).toBe(false);
    expect(stepChangePlaysCombatSound(false, "visible")).toBe(true);
    expect(stepChangePlaysCombatSound(true, "hidden")).toBe(true);
    expect(stepChangePlaysCombatSound(true, undefined)).toBe(false);
  });

  it(`drops a strike that would start more than ${STRIKE_LATE_MS} ms late`, () => {
    expect(strikeLate(1000, 0, 1000 + STRIKE_LATE_MS)).toBe(false);
    expect(strikeLate(1000, 0, 1000 + STRIKE_LATE_MS + 1)).toBe(true);
    // A beat scheduled 1120 ms out (speed 2's pause) is on time when it fires then.
    expect(strikeLate(1000, 1120, 2120)).toBe(false);
  });
});

describe("usableTile (the reflow rule)", () => {
  const tile: CachedTile = {
    box: { x: 10, y: 10, w: 80, h: 112 },
    width: 80,
    height: 112,
    rot: 0,
    board: { w: 1200, h: 800 },
  };
  it("uses a tile cached on a board of this size", () => {
    expect(usableTile(tile, { w: 1200.5, h: 800 })).toBe(tile);
  });
  it("never uses one cached on a board of another size", () => {
    expect(usableTile(tile, { w: 1000, h: 800 })).toBeNull();
    expect(usableTile(undefined, { w: 1200, h: 800 })).toBeNull();
  });
});
