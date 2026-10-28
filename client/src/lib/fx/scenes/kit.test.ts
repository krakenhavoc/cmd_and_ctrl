// @vitest-environment jsdom
//
// ADR 0209 §4 (#4397): the scene kit's geometry, and every scene held to
// its caps. A scene is drawn with motion.run recorded, and no animation
// it starts may run past its scene's length (2.5 s at most) in Full, or
// past its first beat (0.8 s at most) in Short.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../motion", async (importOriginal) => {
  const real = await importOriginal<typeof import("../../motion")>();
  return {
    ...real,
    run: vi.fn(() => ({ cancel() {}, done: Promise.resolve(true) })),
  };
});

import { run } from "../../motion";
import { SCENES, signatureTiming, type SceneID } from "../signatures";
import {
  DRAGON_PATH,
  headingDeg,
  rainDrops,
  screenBox,
  swoopPath,
  tearPath,
  wavePath,
  type SceneCard,
  type SceneContext,
} from "./kit";

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

  it("ash drifts down from the top edge, slowly", () => {
    for (const d of rainDrops("s", 12, board, "ash")) {
      expect(d.fromY).toBeLessThan(0);
      expect(d.fall).toBeGreaterThanOrEqual(0.55);
      expect(d.at + d.fall).toBeLessThanOrEqual(1);
    }
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

const card = (id: string, x: number): SceneCard => ({
  id,
  x,
  y: 400,
  width: 80,
  height: 112,
  rot: 0,
  src: `/${id}.jpg`,
});

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
    expect(lastEnd()).toBeLessThanOrEqual(2500.001);
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
    expect(lastEnd()).toBeLessThanOrEqual(2 * 2500.001);
    expect(lastEnd()).toBeGreaterThan(2500 * 0.5);
  });

  it("Blasphemous Act sends a fireball at each creature it takes, and burns the ground", async () => {
    const mod = await SCENES.blasphemous_act.load();
    const ctx = context("blasphemous_act", false);
    mod.play(ctx);
    expect(
      ctx.layer.querySelectorAll('[data-scene-part="fireball"]').length,
    ).toBeGreaterThanOrEqual(3);
    expect(ctx.layer.querySelectorAll('[data-scene-part="scorch"]').length).toBeGreaterThan(0);
    expect(ctx.layer.querySelectorAll('[data-scene-part="ash"]').length).toBeGreaterThan(0);
  });

  it("Toxic Deluge fills the board and closes over each creature", async () => {
    const mod = await SCENES.toxic_deluge.load();
    const ctx = context("toxic_deluge", false);
    mod.play(ctx);
    // The board's pit, and one over each of the three cards.
    expect(ctx.layer.querySelectorAll('[data-scene-part="fill"]')).toHaveLength(4);
    expect(ctx.layer.querySelectorAll('[data-scene-part="bubble"]').length).toBeGreaterThan(0);
  });

  it("Bitter Triumph tears the card and drives the point out of it", async () => {
    const mod = await SCENES.bitter_triumph.load();
    const ctx = context("bitter_triumph", false);
    mod.play(ctx);
    expect(ctx.layer.querySelector('[data-scene-part="tear"]')).toBeTruthy();
    expect(ctx.layer.querySelector('[data-scene-part="spear"]')).toBeTruthy();
    // With no card it draws nothing (the generic cues stand).
    vi.mocked(run).mockClear();
    const none = { ...context("bitter_triumph", false), cards: [] };
    mod.play(none);
    expect(run).not.toHaveBeenCalled();
  });

  it("Nicol Bolas casts his shadow, then swoops and lands as the card", async () => {
    const mod = await SCENES.nicol_bolas.load();
    const ctx = context("nicol_bolas", false);
    mod.play(ctx);
    expect(ctx.layer.querySelector('[data-scene-part="shadow"]')).toBeTruthy();
    expect(ctx.layer.querySelector('[data-scene-part="dragon"]')).toBeTruthy();
    const landed = ctx.layer.querySelector<HTMLElement>('[data-scene-part="landing-card"] img');
    expect(landed?.getAttribute("src")).toBe("/bolas.jpg");
    // Short, or nowhere to land: the shadow alone.
    for (const c of [
      context("nicol_bolas", true),
      { ...context("nicol_bolas", false), landing: null },
    ]) {
      mod.play(c);
      expect(c.layer.querySelector('[data-scene-part="shadow"]')).toBeTruthy();
      expect(c.layer.querySelector('[data-scene-part="dragon"]')).toBeNull();
    }
  });
});
