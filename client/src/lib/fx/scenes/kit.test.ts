// @vitest-environment jsdom
//
// ADR 0209 §4 (#4397, #4416): the scene kit's geometry, and every scene
// held to its caps. A scene is drawn with motion.run recorded, and no
// animation it starts (a canvas part's clock included) may run past its
// scene's length (SIGNATURE_CAP_MS at most) in Full, or past its first
// beat (0.8 s at most) in Short. jsdom has no 2D canvas, so a canvas part
// plans and times itself but paints nothing here; the preview harness is
// where it is looked at.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../motion", async (importOriginal) => {
  const real = await importOriginal<typeof import("../../motion")>();
  return {
    ...real,
    run: vi.fn(() => ({ cancel() {}, done: Promise.resolve(true) })),
  };
});

import { run } from "../../motion";
import { SCENES, SIGNATURE_CAP_MS, signatureTiming, type SceneID } from "../signatures";
import {
  DRAGON_PATH,
  aimSeat,
  breathWindow,
  fireballPlan,
  flight,
  flightAt,
  headingDeg,
  mouthAt,
  oozeSprings,
  rainDrops,
  roughBladePath,
  screenBox,
  springRadius,
  swoopPath,
  tearPath,
  wavePath,
  type SceneCard,
  type SceneContext,
  type SceneSeat,
} from "./kit";

// jsdom logs "not implemented" for a 2D context; the kit treats null as no canvas.
HTMLCanvasElement.prototype.getContext = (() => null) as never;

const board = { w: 1200, h: 800 };

describe("rainDrops", () => {
  it("fire falls at a slant onto the board, inside its window, the same every time", () => {
    const drops = rainDrops("s", 20, board, "fire");
    expect(drops).toHaveLength(20);
    expect(rainDrops("s", 20, board, "fire")).toEqual(drops);
    expect(rainDrops("t", 20, board, "fire")).not.toEqual(drops);
    for (const d of drops) {
      expect(d.x).toBeGreaterThanOrEqual(0);
      expect(d.x).toBeLessThanOrEqual(board.w);
      expect(d.y).toBeGreaterThan(0);
      expect(d.y).toBeLessThanOrEqual(board.h);
      // From above the board, and from the left: it comes in at a slant.
      expect(d.fromY).toBeLessThan(0);
      expect(d.fromX).toBeLessThan(d.x);
      expect(d.at).toBeGreaterThanOrEqual(0);
      expect(d.at + d.fall).toBeLessThanOrEqual(1);
    }
  });

  it("a fireball is aimed at each target first, and early", () => {
    const targets = [
      { x: 300, y: 400 },
      { x: 700, y: 420 },
    ];
    const drops = rainDrops("s", 3, board, "fire", targets);
    expect(drops).toHaveLength(5);
    expect(drops.slice(0, 2).map((d) => ({ x: d.x, y: d.y }))).toEqual(targets);
    for (const d of drops.slice(0, 2)) expect(d.at).toBeLessThanOrEqual(0.36);
  });

  it("a fireball trails flame all the way down, and the ground burns until the fire dies", () => {
    const plan = fireballPlan("s", board, {
      count: 4,
      start: 100,
      dur: 900,
      impact: true,
      burnUntil: 2800,
      targets: [{ x: 300, y: 400 }],
    });
    expect(plan).toHaveLength(5);
    for (const b of plan) {
      expect(b.hit).toBeGreaterThan(100);
      expect(b.hit).toBeLessThanOrEqual(1000);
      // A flame shed every few ms of its fall: a tail, not a dot.
      expect(b.trail).toBeGreaterThan(20);
      expect(b.flames).toBeGreaterThan(10);
      expect(b.lastFlameEnd).toBeLessThanOrEqual(2800);
    }
    // No impact, no fire on the ground.
    for (const b of fireballPlan("s", board, { count: 2, start: 0, dur: 600 })) {
      expect(b.flames).toBe(0);
    }
  });

  it("ash drifts down from the top edge, slowly", () => {
    for (const d of rainDrops("s", 12, board, "ash")) {
      expect(d.fromY).toBeLessThan(0);
      expect(d.fall).toBeGreaterThanOrEqual(0.55);
      expect(d.at + d.fall).toBeLessThanOrEqual(1);
    }
  });
});

describe("ooze", () => {
  const cards = [card("a", 300), card("b", 700)];
  const timing = { start: 60, rise: 1200, drain: 2400, end: 3300 };

  it("wells up from springs across the whole board, and under each card first", () => {
    const springs = oozeSprings("s", board, cards, timing);
    expect(oozeSprings("s", board, cards, timing)).toEqual(springs);
    const under = springs.filter((s) => s.card);
    expect(under.map((s) => ({ x: s.x, y: s.y }))).toEqual([
      { x: 300, y: 400 },
      { x: 700, y: 400 },
    ]);
    const field = springs.filter((s) => !s.card);
    // Across the board, not along an edge: every quarter has a spring.
    for (const [qx, qy] of [
      [0, 0],
      [1, 0],
      [0, 1],
      [1, 1],
    ]) {
      expect(
        field.some(
          (s) =>
            s.x >= (qx * board.w) / 2 &&
            s.x < ((qx + 1) * board.w) / 2 &&
            s.y >= (qy * board.h) / 2 &&
            s.y < ((qy + 1) * board.h) / 2,
        ),
      ).toBe(true);
    }
    for (const s of springs) {
      expect(s.at).toBeGreaterThanOrEqual(timing.start);
      expect(s.full).toBeGreaterThan(s.at);
      expect(s.drain).toBeGreaterThan(s.full);
      expect(s.gone).toBeGreaterThan(s.drain);
      expect(s.gone).toBeLessThanOrEqual(timing.end);
    }
    // The cards' springs are full before the board's.
    const firstFull = Math.min(...field.map((s) => s.full));
    for (const s of under) expect(s.full).toBeLessThanOrEqual(firstFull + 1);
  });

  it("a spring rises, holds and drains back into the ground", () => {
    const [s] = oozeSprings("s", board, cards, timing);
    expect(springRadius(s, s.at)).toBe(0);
    expect(springRadius(s, (s.full + s.drain) / 2)).toBeGreaterThan(s.r * 0.9);
    expect(springRadius(s, s.gone)).toBe(0);
    expect(springRadius(s, s.gone + 100)).toBe(0);
  });
});

describe("the dragon's breath", () => {
  const to = card("bolas", 800);
  const f = flight({ seed: "s", board }, { to, start: 900, dur: 2400 });
  const foe: SceneSeat = { id: "opp", x: 600, y: 60, r: 26, foe: true };

  it("reads back where the dragon is along its swoop", () => {
    expect(flightAt(f, f.start)).toMatchObject({ x: f.path[0].x, y: f.path[0].y });
    const end = flightAt(f, f.land);
    expect(end.x).toBeCloseTo(to.x);
    expect(end.y).toBeCloseTo(to.y);
    // The mouth is ahead of the body, the way it flies.
    const mid = (f.start + f.land) / 2;
    const p = flightAt(f, mid);
    const m = mouthAt(f, mid);
    expect(Math.hypot(m.x - p.x, m.y - p.y)).toBeCloseTo(f.span * 0.47);
  });

  it("breathes while it loops, facing the seat, well before it lands", () => {
    const w = breathWindow(f, foe, 560)!;
    expect(w).toBeTruthy();
    expect(w.start).toBeGreaterThanOrEqual(f.start);
    expect(w.start + w.dur).toBeLessThan(f.land - 200);
    const mid = w.start + w.dur / 2;
    const p = flightAt(f, mid);
    const m = mouthAt(f, mid);
    const a = (p.rot * Math.PI) / 180;
    const facing = Math.cos(a) * (foe.x - m.x) + Math.sin(a) * (foe.y - m.y);
    expect(facing).toBeGreaterThan(0);
  });

  it("aims at a foe of the card's controller, else at any seat", () => {
    const me: SceneSeat = { id: "me", x: 600, y: 760, r: 26, foe: false };
    const ctx = (seats: SceneSeat[]) => ({ seed: "s", seats });
    expect(aimSeat(ctx([me, foe]))).toBe(foe);
    expect(aimSeat(ctx([me]))).toBe(me);
    expect(aimSeat(ctx([]))).toBeNull();
  });
});

describe("the shapes", () => {
  it("a fill's surface is a closed wave across its width", () => {
    const d = wavePath("s", 400, 8, 24);
    expect(d.startsWith("M0 8")).toBe(true);
    expect(d).toContain("L400 24 L0 24 Z");
  });

  it("a torn hole is a closed star", () => {
    const d = tearPath("s");
    expect(d.startsWith("M")).toBe(true);
    expect(d.endsWith("Z")).toBe(true);
    expect(d.split(" L")).toHaveLength(18);
  });

  it("the spear's blade is a closed, rough leaf, the same every time", () => {
    const d = roughBladePath("s");
    expect(d.startsWith("M20 0")).toBe(true);
    expect(d.endsWith("Z")).toBe(true);
    expect(roughBladePath("s")).toBe(d);
    expect(roughBladePath("t")).not.toBe(d);
    // Hammered rough: the edge does not run smooth, it chips in and out.
    const xs = [...d.matchAll(/L([\d.]+) /g)].map((m) => Number(m[1])).filter((x) => x > 20);
    let turns = 0;
    for (let i = 2; i < xs.length; i++) {
      if (Math.sign(xs[i] - xs[i - 1]) !== Math.sign(xs[i - 1] - xs[i - 2])) turns++;
    }
    expect(turns).toBeGreaterThanOrEqual(4);
  });

  it("the dragon is our own path", () => {
    expect(DRAGON_PATH.startsWith("M")).toBe(true);
    expect(DRAGON_PATH.trim().endsWith("Z")).toBe(true);
  });

  it("a tapped card's box is turned", () => {
    expect(screenBox({ width: 80, height: 112, rot: 0 })).toEqual({ w: 80, h: 112 });
    expect(screenBox({ width: 80, height: 112, rot: 90 })).toEqual({ w: 112, h: 80 });
  });

  it("a swoop comes in from off the board, loops it and ends on the card", () => {
    const to = { x: 640, y: 520 };
    const path = swoopPath("s", board, to);
    expect(path[0].offset).toBe(0);
    expect(path.at(-1)).toMatchObject({ x: to.x, y: to.y, offset: 1 });
    expect(Math.min(path[0].x, board.w - path[0].x)).toBeLessThan(0);
    for (let i = 1; i < path.length; i++) {
      expect(path[i].offset).toBeGreaterThanOrEqual(path[i - 1].offset);
    }
    // It goes around: it reaches near three of the board's edges.
    const xs = path.map((p) => p.x);
    const ys = path.map((p) => p.y);
    expect(Math.max(...xs) - Math.min(...xs)).toBeGreaterThan(board.w * 0.7);
    expect(Math.max(...ys) - Math.min(...ys)).toBeGreaterThan(board.h * 0.6);
    expect(headingDeg({ x: 0, y: 0 }, { x: 0, y: 10 })).toBe(90);
  });
});

// ---- every scene, against its caps ----

function card(id: string, x: number): SceneCard {
  return {
    id,
    x,
    y: 400,
    width: 80,
    height: 112,
    rot: 0,
    src: `/${id}.jpg`,
  };
}

function context(scene: SceneID, short: boolean, speed = 1): SceneContext {
  const layer = document.createElement("div");
  document.body.appendChild(layer);
  const t = signatureTiming(SCENES[scene], short ? "short" : "full", speed)!;
  return {
    layer,
    board,
    seed: `sig-${scene}`,
    short,
    t: (ms) => ms * t.k,
    cards: [card("a", 300), card("b", 500), card("c", 700)],
    landing: card("bolas", 600),
    seats: [
      { id: "opp", x: 600, y: 60, r: 26, foe: true },
      { id: "me", x: 600, y: 760, r: 26, foe: false },
    ],
  };
}

/** The latest any recorded animation ends, in real ms. */
function lastEnd(): number {
  let end = 0;
  for (const [, , opts] of vi.mocked(run).mock.calls) {
    const o = opts as KeyframeAnimationOptions;
    const n = typeof o.iterations === "number" ? o.iterations : 1;
    end = Math.max(end, Number(o.delay ?? 0) + Number(o.duration ?? 0) * n);
  }
  return end;
}

const IDS = Object.keys(SCENES) as SceneID[];

beforeEach(() => vi.mocked(run).mockClear());
afterEach(() => {
  document.body.innerHTML = "";
});

describe("the scenes", () => {
  it.each(IDS)("%s draws within its whole length in Full", async (id) => {
    const mod = await SCENES[id].load();
    const ctx = context(id, false);
    mod.play(ctx);
    expect(run).toHaveBeenCalled();
    expect(ctx.layer.querySelectorAll("[data-scene-part]").length).toBeGreaterThan(0);
    const t = signatureTiming(SCENES[id], "full", 1)!;
    expect(lastEnd()).toBeLessThanOrEqual(t.ms + 0.001);
    expect(lastEnd()).toBeLessThanOrEqual(SIGNATURE_CAP_MS + 0.001);
    // Every part waits unseen and ends gone, never left on the board.
    for (const [, , opts] of vi.mocked(run).mock.calls) {
      expect((opts as KeyframeAnimationOptions).fill).toBe("both");
    }
  });

  it.each(IDS)("%s plays only its first beat in Short, within 0.8 s", async (id) => {
    const mod = await SCENES[id].load();
    mod.play(context(id, true));
    expect(run).toHaveBeenCalled();
    expect(lastEnd()).toBeLessThanOrEqual(800.001);
    expect(lastEnd()).toBeLessThanOrEqual(signatureTiming(SCENES[id], "short", 1)!.ms + 0.001);
  });

  it.each(IDS)("%s scales with animations.speed", async (id) => {
    const mod = await SCENES[id].load();
    mod.play(context(id, false, 2));
    expect(lastEnd()).toBeLessThanOrEqual(2 * SIGNATURE_CAP_MS + 0.001);
    expect(lastEnd()).toBeGreaterThan(SCENES[id].ms);
  });

  it("Blasphemous Act rains fire in two volleys, then ash", async () => {
    const mod = await SCENES.blasphemous_act.load();
    const ctx = context("blasphemous_act", false);
    mod.play(ctx);
    expect(ctx.layer.querySelectorAll('canvas[data-scene-part="fire"]')).toHaveLength(2);
    expect(ctx.layer.querySelectorAll('canvas[data-scene-part="ash"]')).toHaveLength(1);
  });

  it("Toxic Deluge wells up across the board and sinks each creature it takes", async () => {
    const mod = await SCENES.toxic_deluge.load();
    for (const short of [false, true]) {
      const ctx = context("toxic_deluge", short);
      mod.play(ctx);
      expect(ctx.layer.querySelectorAll('canvas[data-scene-part="ooze"]')).toHaveLength(1);
      // Each card it took, drawn where it stood, over the pool, going down.
      const sinking = [...ctx.layer.querySelectorAll<HTMLElement>('[data-scene-part="sink"]')];
      expect(sinking.map((el) => el.querySelector("img")?.getAttribute("src"))).toEqual([
        "/a.jpg",
        "/b.jpg",
        "/c.jpg",
      ]);
      const ooze = ctx.layer.querySelector('[data-scene-part="ooze"]')!;
      for (const el of sinking) {
        expect(ooze.compareDocumentPosition(el) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
      }
    }
  });

  it("Bitter Triumph tears the card and drives the point out of it", async () => {
    const mod = await SCENES.bitter_triumph.load();
    const ctx = context("bitter_triumph", false);
    mod.play(ctx);
    expect(ctx.layer.querySelector('[data-scene-part="tear"]')).toBeTruthy();
    const spear = ctx.layer.querySelector('[data-scene-part="spear"]')!;
    // A black shaft under a rough iron point.
    expect(spear.innerHTML).toContain("c2-spear-shaft");
    expect(spear.innerHTML).toContain(roughBladePath(ctx.seed));
    expect(ctx.layer.querySelector('canvas[data-scene-part="strike-spray"]')).toBeTruthy();
    // With no card it draws nothing (the generic cues stand).
    vi.mocked(run).mockClear();
    const none = { ...context("bitter_triumph", false), cards: [] };
    mod.play(none);
    expect(run).not.toHaveBeenCalled();
  });

  it("Nicol Bolas casts his shadow, swoops, breathes fire on a seat and lands as the card", async () => {
    const mod = await SCENES.nicol_bolas.load();
    const ctx = context("nicol_bolas", false);
    mod.play(ctx);
    expect(ctx.layer.querySelector('[data-scene-part="shadow"]')).toBeTruthy();
    expect(ctx.layer.querySelector('[data-scene-part="dragon"]')).toBeTruthy();
    expect(ctx.layer.querySelector('canvas[data-scene-part="breath"]')).toBeTruthy();
    const landed = ctx.layer.querySelector<HTMLElement>('[data-scene-part="landing-card"] img');
    expect(landed?.getAttribute("src")).toBe("/bolas.jpg");
    // No seat on the board: he flies and lands without breathing.
    const quiet = { ...context("nicol_bolas", false), seats: [] };
    mod.play(quiet);
    expect(quiet.layer.querySelector('[data-scene-part="dragon"]')).toBeTruthy();
    expect(quiet.layer.querySelector('[data-scene-part="breath"]')).toBeNull();
    // Short, or nowhere to land: the shadow alone.
    for (const c of [
      context("nicol_bolas", true),
      { ...context("nicol_bolas", false), landing: null },
    ]) {
      mod.play(c);
      expect(c.layer.querySelector('[data-scene-part="shadow"]')).toBeTruthy();
      expect(c.layer.querySelector('[data-scene-part="dragon"]')).toBeNull();
      expect(c.layer.querySelector('[data-scene-part="breath"]')).toBeNull();
    }
  });

  it("the landing ring waits unseen until he lands", async () => {
    const mod = await SCENES.nicol_bolas.load();
    const ctx = context("nicol_bolas", false);
    mod.play(ctx);
    const ring = ctx.layer.querySelector('[data-scene-part="landing-ring"]')!;
    const call = vi.mocked(run).mock.calls.find(([el]) => el === ring)!;
    expect((call[1] as Keyframe[])[0].opacity).toBe(0);
  });

  it("Dragon's Approach sends two shadows over the board, one in Short", async () => {
    const mod = await SCENES.dragons_approach.load();
    const full = context("dragons_approach", false);
    mod.play(full);
    expect(full.layer.querySelectorAll('[data-scene-part="shadow"]')).toHaveLength(2);
    const short = context("dragons_approach", true);
    mod.play(short);
    expect(short.layer.querySelectorAll('[data-scene-part="shadow"]')).toHaveLength(1);
  });
});
