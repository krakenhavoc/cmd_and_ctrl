// The scene kit (ADR 0209 §4, #4397): the parts every signature scene is
// built from.
//
//   overlay      a full-board wash (a red sky, a sickly green, a dark tint)
//   canvasPart   a canvas of particles, painted each frame from one clock
//   rain         particle rain: round fireballs trailing flame that crash
//                and burn the ground, or ash and embers that drift
//   fill         a rising fill (tar, water) with a moving surface and
//                bubbles, over the whole board or over one card
//   ooze         sludge welling up out of the ground across the board
//   sink         a card going down into the ground
//   shadowPass   a dragon's shadow crossing the board
//   strike       a spear through the back of a card and out the front
//   flyLand      a dragon that swoops around the board, then lands as the card
//   breath       the dragon breathing fire on a seat as it flies
//
// Every part draws plain DOM (inline styles, inline SVG) or a canvas
// inside the scene's layer and moves it with the Web Animations API
// through motion.run (a canvas keeps its clock on such an animation), so
// a skip (unmounting the layer) cancels all of it. Times are the scene's own
// milliseconds at speed 1; ctx.t turns them into real ones (the cap, Short
// and animations.speed are already folded into it, fx/signatures.ts). A
// part ends by the time it was given: nothing outlives its scene.
//
// Every shape is drawn here, original to C2: no card art except the card's
// own image, which the board already shows (ADR 0209 §6).
//
// The geometry (where drops fall, the wave, the swoop) is pure and
// unit-tested in kit.test.ts.

import { seeded } from "../../combatStrikes";
import { EASE, run } from "../../motion";

// ---- The contract a scene module keeps ----

/** A card on the board, by its centre in board pixels, its size at rest and its rotation. */
export interface SceneCard {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
  rot: number;
  src: string | null;
}

export interface SceneContext {
  /** The scene's layer: a board-sized box, the parts' parent. */
  layer: HTMLElement;
  board: { w: number; h: number };
  /** Seeds the scene's randomness, so a scene always looks the same for one entry. */
  seed: string;
  /** Short (animations.effects): play the first beat only. */
  short: boolean;
  /** The scene's ms at speed 1 -> real ms. */
  t(ms: number): number;
  /** The permanents its batch took (resolve), where they stood. */
  cards: SceneCard[];
  /** Where it lands, as the card (enter), or null. */
  landing: SceneCard | null;
  /** The players' seats (their avatar circles), when the board has them. */
  seats?: SceneSeat[];
}

/** A player's seat: its avatar circle's centre and radius, and whether it is a foe of the card's controller. */
export interface SceneSeat {
  id: string;
  x: number;
  y: number;
  r: number;
  foe: boolean;
}

/** What a file under fx/scenes/ exports. */
export interface SceneModule {
  play(ctx: SceneContext): void;
}

// ---- Small helpers ----

type Style = Record<string, string>;

function div(parent: HTMLElement, style: Style, part: string): HTMLDivElement {
  const el = document.createElement("div");
  el.dataset.scenePart = part;
  Object.assign(el.style, { position: "absolute", left: "0px", top: "0px" }, style);
  parent.appendChild(el);
  return el;
}

function svg(parent: HTMLElement, viewBox: string, inner: string, style: Style, part: string) {
  const host = div(parent, style, part);
  host.innerHTML = `<svg viewBox="${viewBox}" width="100%" height="100%" preserveAspectRatio="none" style="display:block;overflow:visible">${inner}</svg>`;
  return host;
}

/** anim runs keyframes from `start` for `dur` scene ms; fill both, so a part waits unseen and ends gone. */
function anim(
  ctx: SceneContext,
  el: Element,
  frames: Keyframe[],
  start: number,
  dur: number,
  easing: string = EASE.linear,
  iterations = 1,
): void {
  const n = Math.max(1, Math.floor(iterations));
  run(el, frames, {
    delay: ctx.t(Math.max(0, start)),
    duration: Math.max(1, ctx.t(dur / n)),
    iterations: n,
    easing,
    fill: "both",
  });
}

function clamp(v: number, lo: number, hi: number): number {
  return Math.min(hi, Math.max(lo, v));
}

function r1(n: number): number {
  return Math.round(n * 10) / 10;
}

/** The on-screen box of a card (a tapped card is turned). */
export function screenBox(c: Pick<SceneCard, "width" | "height" | "rot">): {
  w: number;
  h: number;
} {
  const turned = Math.abs(Math.round(c.rot / 90)) % 2 === 1;
  return turned ? { w: c.height, h: c.width } : { w: c.width, h: c.height };
}

/** The board's scale against a 1200 px table, for sizing parts. */
export function boardScale(board: { w: number; h: number }): number {
  return clamp(Math.min(board.w / 1200, board.h / 760), 0.55, 1.4);
}

// ---- canvas: the particle layer ----
//
// Fire, sludge and breath are particles: hundreds of soft sprites a frame,
// more than DOM nodes should carry. A canvas part paints them on one
// board-sized canvas. Its clock is a Web Animation on the canvas itself
// (through motion.run, like every other part), so a skip that cancels the
// layer's animations stops it and the caps hold it; each frame paints the
// moment the scene has reached. Every particle is a function of time, planned up
// front from the scene's seed: a dropped frame skips ahead, never lags,
// and a scene looks the same each time it plays.

type G = CanvasRenderingContext2D;

function context2d(c: HTMLCanvasElement): G | null {
  try {
    return c.getContext("2d");
  } catch {
    return null;
  }
}

export interface CanvasPartOpts {
  part: string;
  start: number;
  dur: number;
  /** Paints the moment `ms` (scene ms, inside [start, start + dur]) from scratch. */
  draw(g: G, ms: number): void;
}

/** canvasPart paints `draw` on a board-sized canvas every frame from `start` for `dur`. */
export function canvasPart(ctx: SceneContext, o: CanvasPartOpts): HTMLCanvasElement {
  const { w, h } = ctx.board;
  const canvas = document.createElement("canvas");
  canvas.dataset.scenePart = o.part;
  Object.assign(canvas.style, {
    position: "absolute",
    left: "0px",
    top: "0px",
    width: `${w}px`,
    height: `${h}px`,
  });
  ctx.layer.appendChild(canvas);
  // The clock: on for the part's window, gone at its end.
  anim(
    ctx,
    canvas,
    [
      { offset: 0, opacity: 1 },
      { offset: 0.995, opacity: 1 },
      { offset: 1, opacity: 0 },
    ],
    o.start,
    o.dur,
  );
  const g = context2d(canvas);
  if (!g || typeof requestAnimationFrame !== "function") return canvas;
  const dpr = clamp(typeof window === "undefined" ? 1 : window.devicePixelRatio || 1, 1, 2);
  canvas.width = Math.max(1, Math.round(w * dpr));
  canvas.height = Math.max(1, Math.round(h * dpr));
  // The clock animation says when the part is over or skipped (cancelled:
  // idle). The moment to paint is read off performance.now(), from when
  // the scene started: it is the same clock to within a frame, and it is
  // still right where the page draws few frames (a background tab coming
  // back, a headless capture).
  const clock = canvas.getAnimations?.().at(-1) ?? null;
  const k = ctx.t(1) || 1;
  const born = performance.now();
  const end = o.start + o.dur;
  const tick = () => {
    if (!canvas.isConnected || clock?.playState === "idle") return;
    const ms = (performance.now() - born) / k;
    g.setTransform(dpr, 0, 0, dpr, 0, 0);
    g.globalAlpha = 1;
    g.globalCompositeOperation = "source-over";
    g.clearRect(0, 0, w, h);
    if (ms >= end || clock?.playState === "finished") return;
    if (ms >= o.start) {
      g.save();
      o.draw(g, ms);
      g.restore();
    }
    requestAnimationFrame(tick);
  };
  requestAnimationFrame(tick);
  return canvas;
}

// Soft round sprites, drawn once and stamped many times a frame.
type SpriteName =
  | "core"
  | "fire"
  | "ember"
  | "glow"
  | "smoke"
  | "scorch"
  | "fume"
  | "sheen"
  | "spray";
const SPRITE_STOPS: Record<SpriteName, [number, string][]> = {
  core: [
    [0, "rgba(255,255,246,1)"],
    [0.35, "rgba(255,240,170,0.95)"],
    [0.7, "rgba(255,190,80,0.5)"],
    [1, "rgba(255,150,40,0)"],
  ],
  fire: [
    [0, "rgba(255,246,200,1)"],
    [0.2, "rgba(255,214,110,0.95)"],
    [0.45, "rgba(255,130,30,0.65)"],
    [0.72, "rgba(220,50,0,0.25)"],
    [1, "rgba(160,20,0,0)"],
  ],
  ember: [
    [0, "rgba(255,170,80,0.9)"],
    [0.4, "rgba(230,80,10,0.5)"],
    [1, "rgba(140,20,0,0)"],
  ],
  glow: [
    [0, "rgba(255,180,80,0.6)"],
    [0.45, "rgba(255,100,20,0.22)"],
    [1, "rgba(255,60,0,0)"],
  ],
  smoke: [
    [0, "rgba(38,28,26,0.6)"],
    [0.55, "rgba(30,22,20,0.28)"],
    [1, "rgba(20,14,12,0)"],
  ],
  scorch: [
    [0, "rgba(14,6,3,0.9)"],
    [0.45, "rgba(36,12,4,0.6)"],
    [0.75, "rgba(60,20,6,0.22)"],
    [1, "rgba(0,0,0,0)"],
  ],
  fume: [
    [0, "rgba(22,40,20,0.5)"],
    [0.6, "rgba(14,28,12,0.22)"],
    [1, "rgba(8,16,8,0)"],
  ],
  sheen: [
    [0, "rgba(120,168,92,0.32)"],
    [0.5, "rgba(80,120,60,0.12)"],
    [1, "rgba(60,90,40,0)"],
  ],
  spray: [
    [0, "rgba(255,120,120,0.95)"],
    [0.35, "rgba(170,10,30,0.8)"],
    [1, "rgba(90,0,12,0)"],
  ],
};
const sprites = new Map<SpriteName, HTMLCanvasElement | null>();

function sprite(name: SpriteName): HTMLCanvasElement | null {
  if (sprites.has(name)) return sprites.get(name) ?? null;
  const c = document.createElement("canvas");
  c.width = c.height = 64;
  const g = context2d(c);
  if (!g) {
    sprites.set(name, null);
    return null;
  }
  const grad = g.createRadialGradient(32, 32, 0, 32, 32, 32);
  for (const [at, col] of SPRITE_STOPS[name]) grad.addColorStop(at, col);
  g.fillStyle = grad;
  g.fillRect(0, 0, 64, 64);
  sprites.set(name, c);
  return c;
}

/** blot stamps a sprite: centre, radius (or an ellipse's half-sizes) and alpha. */
function blot(
  g: G,
  s: HTMLCanvasElement | null,
  x: number,
  y: number,
  r: number,
  a: number,
  ry = r,
) {
  if (!s || a <= 0.004 || r <= 0.2) return;
  g.globalAlpha = Math.min(1, a);
  g.drawImage(s, x - r, y - ry, r * 2, ry * 2);
}

function smooth(e0: number, e1: number, x: number): number {
  const t = clamp((x - e0) / (e1 - e0), 0, 1);
  return t * t * (3 - 2 * t);
}
const easeOut3 = (u: number) => 1 - Math.pow(1 - clamp(u, 0, 1), 3);
const easeIn2 = (u: number) => clamp(u, 0, 1) ** 2;
const easeIn3 = (u: number) => clamp(u, 0, 1) ** 3;

/** A particle planned up front: born at `t`, gone after `life`, from (x, y) at (vx, vy) px per ms. */
interface Mote {
  t: number;
  life: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  r: number;
  ph: number;
}

/** Its age as a fraction of its life at `ms`, or -1 when it is not alive. */
function ageOf(p: Mote, ms: number): number {
  const a = (ms - p.t) / p.life;
  return a < 0 || a >= 1 ? -1 : a;
}

/** A spark that flies out and slows (no gravity: the table is seen from above). */
function flown(p: Mote, ms: number, tau = 160): { x: number; y: number } {
  const a = Math.max(0, ms - p.t);
  const d = tau * (1 - Math.exp(-a / tau));
  return { x: p.x + p.vx * d, y: p.y + p.vy * d };
}

/** sparks plans `n` sparks bursting from (x, y) at `t`. */
function burst(
  rnd: () => number,
  x: number,
  y: number,
  t: number,
  n: number,
  speed: number,
  life: [number, number],
): Mote[] {
  const out: Mote[] = [];
  for (let i = 0; i < n; i++) {
    const a = rnd() * Math.PI * 2;
    const v = speed * (0.45 + rnd() * 0.75);
    out.push({
      t: t + rnd() * 30,
      life: life[0] + rnd() * (life[1] - life[0]),
      x,
      y,
      vx: Math.cos(a) * v,
      vy: Math.sin(a) * v,
      r: 0.7 + rnd() * 0.8,
      ph: rnd() * 6.28,
    });
  }
  return out;
}

/** drawSparks strokes each live spark as a short bright streak along its flight. */
function drawSparks(g: G, sparks: readonly Mote[], ms: number, color: string, k: number) {
  g.strokeStyle = color;
  g.lineCap = "round";
  for (const p of sparks) {
    const u = ageOf(p, ms);
    if (u < 0) continue;
    const at = flown(p, ms);
    const back = flown(p, ms - 28);
    g.globalAlpha = (1 - u) * 0.95;
    g.lineWidth = p.r * 1.8 * k * (1 - u * 0.6);
    g.beginPath();
    g.moveTo(back.x, back.y);
    g.lineTo(at.x, at.y);
    g.stroke();
  }
}

// ---- overlay ----

export interface OverlayOpts {
  /** Any CSS background. */
  background: string;
  start: number;
  dur: number;
  /** Peak opacity, and how much of `dur` the fade in and out each take. */
  peak?: number;
  fade?: number;
}

/** overlay washes the whole board: in, held, out. */
export function overlay(ctx: SceneContext, o: OverlayOpts): HTMLElement {
  const el = div(
    ctx.layer,
    { width: "100%", height: "100%", background: o.background, opacity: "0" },
    "overlay",
  );
  const f = clamp(o.fade ?? 0.25, 0.01, 0.5);
  const peak = o.peak ?? 1;
  anim(
    ctx,
    el,
    [
      { offset: 0, opacity: 0 },
      { offset: f, opacity: peak },
      { offset: 1 - f, opacity: peak },
      { offset: 1, opacity: 0 },
    ],
    o.start,
    o.dur,
  );
  return el;
}

// ---- rain ----

export type RainKind = "fire" | "ash";

export interface Drop {
  /** Where it lands (fire) or ends (ash), in board px. */
  x: number;
  y: number;
  /** Where it starts, above the board. */
  fromX: number;
  fromY: number;
  size: number;
  /** When it starts, as a fraction of the window, and its fall as a fraction too. */
  at: number;
  fall: number;
  /** Its turn, in degrees: a fireball's head points the way it falls; a flake spins. */
  rot: number;
}

/**
 * rainDrops plans a rain of `n` drops over a board, seeded. Fire comes in
 * from the upper left at a slant and lands inside the board; a target
 * (a card's centre) gets one fireball of its own, aimed at it, early. Ash
 * drifts down from the top edge with a little sideways sway. Every drop
 * starts and ends within its window: at + fall <= 1.
 */
export function rainDrops(
  seed: string,
  n: number,
  board: { w: number; h: number },
  kind: RainKind,
  targets: readonly { x: number; y: number }[] = [],
): Drop[] {
  const rnd = seeded(`${seed}:${kind}`);
  const k = boardScale(board);
  const out: Drop[] = [];
  const total = Math.max(0, Math.floor(n)) + (kind === "fire" ? targets.length : 0);
  for (let i = 0; i < total; i++) {
    const target = kind === "fire" && i < targets.length ? targets[i] : null;
    if (kind === "fire") {
      const x = target ? target.x : board.w * (0.05 + rnd() * 0.9);
      const y = target ? target.y : board.h * (0.16 + rnd() * 0.78);
      const slant = ((22 + rnd() * 14) * Math.PI) / 180;
      const size = (22 + rnd() * 18) * k;
      const rise = y + 60 + size * 3;
      const fall = target ? 0.3 + rnd() * 0.12 : 0.22 + rnd() * 0.2;
      const at = target ? Math.min(1 - fall, 0.1 + rnd() * 0.25) : rnd() * (1 - fall);
      const dx = Math.tan(slant) * rise;
      out.push({
        x: r1(x),
        y: r1(y),
        fromX: r1(x - dx),
        fromY: r1(y - rise),
        size: r1(size),
        at: Math.round(at * 1000) / 1000,
        fall: Math.round(fall * 1000) / 1000,
        rot: r1((-Math.atan2(dx, rise) * 180) / Math.PI),
      });
    } else {
      const x = board.w * rnd();
      const fall = 0.55 + rnd() * 0.4;
      const at = rnd() * (1 - fall);
      const y = board.h * (0.35 + rnd() * 0.65);
      out.push({
        x: r1(x + (rnd() - 0.5) * 120 * k),
        y: r1(y),
        fromX: r1(x),
        fromY: r1(-20 - rnd() * 60),
        size: r1((3 + rnd() * 6) * k),
        at: Math.round(at * 1000) / 1000,
        fall: Math.round(fall * 1000) / 1000,
        rot: r1((rnd() * 2 - 1) * 540),
      });
    }
  }
  return out;
}

export interface RainOpts {
  kind: RainKind;
  count: number;
  start: number;
  /** The window every drop falls (and, for fire, burns) within. */
  dur: number;
  /** Fire only: each fireball bursts where it lands and leaves the ground burning until `burnUntil`. */
  impact?: boolean;
  burnUntil?: number;
  /** Fire only: a fireball aimed at each of these, first. */
  targets?: readonly { x: number; y: number }[];
  seed?: string;
}

/** A fireball planned: its drop, where it is at a time, and what it leaves behind. */
interface Fireball {
  d: Drop;
  t0: number;
  hit: number;
  R: number;
  trail: Mote[];
  smoke: Mote[];
  sparks: Mote[];
  flames: Mote[];
  ph: number;
}

/** Where a fireball is `ms` into its fall: it speeds up as it comes down. */
function ballAt(b: Fireball, ms: number): { x: number; y: number } {
  const e = easeIn2((ms - b.t0) / Math.max(1, b.hit - b.t0));
  return { x: b.d.fromX + (b.d.x - b.d.fromX) * e, y: b.d.fromY + (b.d.y - b.d.fromY) * e };
}

/**
 * planFireballs turns a rain's drops into fireballs: a round ball of fire
 * with a trail of flame and smoke behind it, and, where it lands, a flash,
 * a ring, sparks, a scorch and flames that burn until `burnUntil`.
 */
function planFireballs(
  seed: string,
  drops: readonly Drop[],
  start: number,
  dur: number,
  impact: boolean,
  burnUntil: number,
  k: number,
): Fireball[] {
  const rnd = seeded(`${seed}:fireballs`);
  return drops.map((d) => {
    const t0 = start + d.at * dur;
    const hit = t0 + d.fall * dur;
    const R = d.size * 0.58;
    const b: Fireball = {
      d,
      t0,
      hit,
      R,
      trail: [],
      smoke: [],
      sparks: [],
      flames: [],
      ph: rnd() * 6.28,
    };
    const dx = d.x - d.fromX;
    const dy = d.y - d.fromY;
    const len = Math.hypot(dx, dy) || 1;
    const ux = dx / len;
    const uy = dy / len;
    // The trail: a flame shed every few ms, drifting back and out.
    for (let t = t0; t < hit; t += 3.5) {
      const p = ballAt(b, t);
      const side = (rnd() - 0.5) * 0.08;
      b.trail.push({
        t,
        life: 90 + rnd() * 110,
        x: p.x + (rnd() - 0.5) * R * 0.8,
        y: p.y + (rnd() - 0.5) * R * 0.8,
        vx: -ux * 0.05 - uy * side * 1.6,
        vy: -uy * 0.05 + ux * side * 1.6 - 0.02,
        r: R * (0.8 + rnd() * 0.6),
        ph: rnd() * 6.28,
      });
    }
    for (let t = t0 + 20; t < hit; t += 22) {
      const p = ballAt(b, t);
      b.smoke.push({
        t,
        life: 260 + rnd() * 220,
        x: p.x,
        y: p.y,
        vx: (rnd() - 0.5) * 0.03,
        vy: -0.02 - rnd() * 0.02,
        r: R * (0.6 + rnd() * 0.4),
        ph: rnd() * 6.28,
      });
    }
    if (!impact) return b;
    b.sparks = burst(rnd, d.x, d.y, hit, 12, 0.55 * k, [240, 520]);
    // The ground burns: a flame licks up every 40 ms until the fire dies.
    for (let t = hit + 30; t < burnUntil - 120; t += 13 + rnd() * 8) {
      const life = Math.min(burnUntil - t, 340 + rnd() * 260);
      if (life < 120) break;
      const off = (rnd() - 0.5) * 2;
      b.flames.push({
        t,
        life,
        x: d.x + off * R * 1.3,
        y: d.y + (rnd() - 0.5) * R * 0.6,
        vx: off * 0.01,
        vy: -(0.06 + rnd() * 0.07) * k,
        // Taller in the middle of the fire, lower at its edges.
        r: R * (0.55 + rnd() * 0.45) * (1.15 - Math.abs(off) * 0.45),
        ph: rnd() * 6.28,
      });
    }
    return b;
  });
}

function drawFireballs(g: G, balls: readonly Fireball[], ms: number, burnUntil: number, k: number) {
  const core = sprite("core");
  const fire = sprite("fire");
  const ember = sprite("ember");
  const glow = sprite("glow");
  const smoke = sprite("smoke");
  const scorch = sprite("scorch");
  const dying = 1 - smooth(burnUntil - 420, burnUntil, ms);

  // Under the fire: the scorch and the smoke, painted normally.
  g.globalCompositeOperation = "source-over";
  for (const b of balls) {
    if (ms >= b.hit) {
      const a = smooth(b.hit, b.hit + 140, ms) * (0.35 + 0.55 * dying);
      blot(g, scorch, b.d.x, b.d.y, b.R * 2.8, a, b.R * 1.5);
    }
    for (const p of b.smoke) {
      const u = ageOf(p, ms);
      if (u < 0) continue;
      const age = ms - p.t;
      blot(g, smoke, p.x + p.vx * age, p.y + p.vy * age, p.r * (1 + u * 1.6), 0.42 * (1 - u));
    }
    for (const p of b.flames) {
      const u = ageOf(p, ms);
      if (u < 0.55) continue;
      const age = ms - p.t;
      blot(g, smoke, p.x + p.vx * age, p.y + p.vy * age - 6 * k, p.r * (1 + u), 0.3 * (1 - u));
    }
  }

  // The fire itself, added up so where flames cross they burn brighter.
  g.globalCompositeOperation = "lighter";
  for (const b of balls) {
    if (ms >= b.hit) {
      const flick = 1 + 0.12 * Math.sin(ms / 45 + b.ph) + 0.06 * Math.sin(ms / 17 + b.ph * 2);
      blot(g, glow, b.d.x, b.d.y, b.R * 3.6 * flick, 0.75 * dying * smooth(b.hit, b.hit + 80, ms));
    }
    for (const p of b.flames) {
      const u = ageOf(p, ms);
      if (u < 0) continue;
      const age = ms - p.t;
      const x = p.x + p.vx * age + Math.sin(age / 70 + p.ph) * 2.5 * k;
      const y = p.y + p.vy * age;
      // A tongue of flame: tall, narrowing as it rises and cools.
      const r = p.r * (1 - u * 0.6) * (u < 0.12 ? u / 0.12 : 1);
      blot(g, u < 0.45 ? fire : ember, x, y, r * 0.8, (1 - u) * 0.6 * dying, r * 1.5);
    }
    for (const p of b.trail) {
      const u = ageOf(p, ms);
      if (u < 0) continue;
      const age = ms - p.t;
      // Each flame puffs out a little, then shrinks as it cools.
      const r = p.r * (0.8 + 0.5 * Math.sin(Math.PI * Math.min(1, u * 1.4))) * (1 - u * 0.6);
      blot(g, u < 0.3 ? fire : ember, p.x + p.vx * age, p.y + p.vy * age, r, (1 - u) * 0.5);
    }
    if (ms >= b.t0 && ms < b.hit) {
      const p = ballAt(b, ms);
      // The tail: a tapering tongue of flame back along its fall.
      const back = ballAt(b, Math.max(b.t0, ms - 45));
      const tx = back.x - p.x;
      const ty = back.y - p.y;
      const tl = Math.hypot(tx, ty);
      if (tl > 2) {
        const nx = -ty / tl;
        const ny = tx / tl;
        const wob = Math.sin(ms / 19 + b.ph) * b.R * 0.25;
        const tail = g.createLinearGradient(p.x, p.y, back.x, back.y);
        tail.addColorStop(0, "rgba(255,214,120,0.9)");
        tail.addColorStop(0.35, "rgba(255,120,24,0.55)");
        tail.addColorStop(1, "rgba(190,36,0,0)");
        g.globalAlpha = smooth(b.t0, b.t0 + 60, ms);
        g.fillStyle = tail;
        g.beginPath();
        g.moveTo(p.x + nx * b.R * 0.95, p.y + ny * b.R * 0.95);
        g.quadraticCurveTo(
          p.x + tx * 0.5 + nx * (b.R * 0.6 + wob),
          p.y + ty * 0.5 + ny * (b.R * 0.6 + wob),
          back.x + nx * wob,
          back.y + ny * wob,
        );
        g.quadraticCurveTo(
          p.x + tx * 0.5 - nx * (b.R * 0.6 - wob),
          p.y + ty * 0.5 - ny * (b.R * 0.6 - wob),
          p.x - nx * b.R * 0.95,
          p.y - ny * b.R * 0.95,
        );
        g.closePath();
        g.fill();
      }
      const f = 1 + 0.07 * Math.sin(ms / 23 + b.ph);
      const appear = smooth(b.t0, b.t0 + 60, ms);
      blot(g, glow, p.x, p.y, b.R * 3.4 * f, 0.95 * appear);
      blot(g, fire, p.x, p.y, b.R * 1.75 * f, appear);
      blot(g, core, p.x, p.y, b.R * 1.0 * f, appear);
    }
    if (ms >= b.hit && ms < b.hit + 300) {
      const u = (ms - b.hit) / 300;
      blot(g, core, b.d.x, b.d.y, b.R * (1.5 + 3.5 * easeOut3(u)), 1 - u);
      blot(g, glow, b.d.x, b.d.y, b.R * (2 + 5 * easeOut3(u)), (1 - u) * 0.9);
      g.globalAlpha = (1 - u) * (1 - u) * 0.45;
      g.strokeStyle = "rgba(255,170,90,1)";
      g.lineWidth = Math.max(0.5, 2 * (1 - u)) * k;
      g.beginPath();
      g.ellipse(
        b.d.x,
        b.d.y,
        b.R * (1 + 4.2 * easeOut3(u)),
        b.R * (0.6 + 2.4 * easeOut3(u)),
        0,
        0,
        Math.PI * 2,
      );
      g.stroke();
    }
    drawSparks(g, b.sparks, ms, "rgba(255,220,150,1)", k);
  }
}

/** Ash and embers: flakes drift down, sparks of the fire float up. */
interface Flake extends Drop {
  t0: number;
  tf: number;
  ember: boolean;
  ph: number;
}

function drawAsh(g: G, flakes: readonly Flake[], ms: number, k: number) {
  const ember = sprite("ember");
  for (const f of flakes) {
    const u = (ms - f.t0) / f.tf;
    if (u < 0 || u >= 1) continue;
    const fade = smooth(0, 0.15, u) * (1 - smooth(0.75, 1, u));
    if (f.ember) {
      g.globalCompositeOperation = "lighter";
      const x = f.x + Math.sin(u * 7 + f.ph) * 10 * k;
      const y = f.fromY + (f.y - f.fromY) * u;
      const tw = 0.6 + 0.4 * Math.sin(ms / 60 + f.ph);
      blot(g, ember, x, y, f.size * 1.2, fade * tw);
      continue;
    }
    g.globalCompositeOperation = "source-over";
    const e = u * u * (3 - 2 * u);
    const x = f.fromX + (f.x - f.fromX) * e + Math.sin(u * 9 + f.ph) * 6 * k;
    const y = f.fromY + (f.y - f.fromY) * e;
    g.globalAlpha = 0.85 * fade;
    g.save();
    g.translate(x, y);
    g.rotate(((f.rot * u) / 180) * Math.PI);
    g.scale(1, 0.55 + 0.45 * Math.abs(Math.sin(u * 8 + f.ph)));
    g.fillStyle = "rgba(64,58,56,1)";
    g.beginPath();
    g.ellipse(0, 0, f.size * 0.6, f.size * 0.42, 0, 0, Math.PI * 2);
    g.fill();
    g.strokeStyle = "rgba(255,120,50,0.5)";
    g.lineWidth = 0.8;
    g.stroke();
    g.restore();
  }
}

/** rain drops fire or ash on the board, on one canvas. */
export function rain(ctx: SceneContext, o: RainOpts): HTMLCanvasElement {
  const seed = o.seed ?? ctx.seed;
  const k = boardScale(ctx.board);
  const drops = rainDrops(seed, o.count, ctx.board, o.kind, o.targets ?? []);
  const end = o.start + o.dur;
  if (o.kind === "ash") {
    const rnd = seeded(`${seed}:flakes`);
    const flakes: Flake[] = drops.map((d, i) => ({
      ...d,
      t0: o.start + d.at * o.dur,
      tf: d.fall * o.dur,
      // One in three is an ember floating up from the fire instead.
      ember: i % 3 === 2,
      ph: rnd() * 6.28,
    }));
    for (const f of flakes) {
      if (!f.ember) continue;
      f.fromY = f.y + 40 * k;
      f.y = f.y - (90 + rnd() * 120) * k;
    }
    return canvasPart(ctx, {
      part: "ash",
      start: o.start,
      dur: o.dur,
      draw: (g, ms) => drawAsh(g, flakes, ms, k),
    });
  }
  const burnUntil = Math.max(end, o.burnUntil ?? end);
  const balls = planFireballs(seed, drops, o.start, o.dur, !!o.impact, burnUntil, k);
  return canvasPart(ctx, {
    part: "fire",
    start: o.start,
    dur: burnUntil - o.start,
    draw: (g, ms) => drawFireballs(g, balls, ms, burnUntil, k),
  });
}

/** fireballPlan is what a fire rain plans for each fireball: when it lands, its trail and its flames. */
export function fireballPlan(
  seed: string,
  board: { w: number; h: number },
  o: Pick<RainOpts, "count" | "start" | "dur" | "burnUntil" | "targets" | "impact">,
): { hit: number; trail: number; flames: number; lastFlameEnd: number }[] {
  const drops = rainDrops(seed, o.count, board, "fire", o.targets ?? []);
  const end = o.start + o.dur;
  const burnUntil = Math.max(end, o.burnUntil ?? end);
  return planFireballs(seed, drops, o.start, o.dur, !!o.impact, burnUntil, boardScale(board)).map(
    (b) => ({
      hit: b.hit,
      trail: b.trail.length,
      flames: b.flames.length,
      lastFlameEnd: Math.max(0, ...b.flames.map((f) => f.t + f.life)),
    }),
  );
}

// ---- fill ----

/**
 * wavePath is a fill's surface: a wave across `w` (two periods' worth, so
 * sliding it half its width loops), `amp` high, closed down to `depth`.
 * Path data in px.
 */
export function wavePath(seed: string, w: number, amp: number, depth: number): string {
  const rnd = seeded(`${seed}:wave`);
  const waves = 6;
  const step = w / waves;
  const a = Math.max(1, amp);
  let d = `M0 ${r1(a)}`;
  for (let i = 0; i < waves; i++) {
    const x0 = i * step;
    const up = a * (0.4 + rnd() * 0.6);
    d += ` Q${r1(x0 + step / 4)} ${r1(a - up)} ${r1(x0 + step / 2)} ${r1(a)}`;
    d += ` Q${r1(x0 + (3 * step) / 4)} ${r1(a + up * 0.6)} ${r1(x0 + step)} ${r1(a)}`;
  }
  d += ` L${r1(w)} ${r1(depth)} L0 ${r1(depth)} Z`;
  return d;
}

export interface FillLevel {
  /** Scene ms. */
  at: number;
  /** 0 (empty) to 1 (the box's full height). */
  level: number;
}

export interface FillOpts {
  /** The box it fills: top-left and size, in board px. */
  box: { x: number; y: number; w: number; h: number };
  /** The level over time; the first and last should be 0. */
  levels: readonly FillLevel[];
  /** Colours: the surface, the body top and the body bottom. */
  surface: string;
  top: string;
  bottom: string;
  bubbles?: number;
  bubble?: string;
  seed?: string;
}

/** fill rises in a box (tar, water) and drains, with a moving surface and bubbles. */
export function fill(ctx: SceneContext, o: FillOpts): void {
  const levels = [...o.levels].sort((a, b) => a.at - b.at);
  if (levels.length < 2) return;
  const start = levels[0].at;
  const end = levels[levels.length - 1].at;
  const dur = Math.max(1, end - start);
  const peak = Math.max(...levels.map((l) => clamp(l.level, 0, 1.2)));
  const amp = clamp(o.box.h * 0.03, 4, 14);
  const bodyH = Math.max(1, o.box.h * peak + amp);
  const clip = div(
    ctx.layer,
    {
      transform: `translate(${o.box.x}px, ${o.box.y}px)`,
      width: `${o.box.w}px`,
      height: `${o.box.h}px`,
      overflow: "hidden",
    },
    "fill",
  );
  const body = div(
    clip,
    {
      top: "auto",
      bottom: "0px",
      width: "100%",
      height: `${bodyH}px`,
      background: `linear-gradient(to bottom, ${o.top}, ${o.bottom})`,
      transform: `translateY(${bodyH}px)`,
    },
    "fill-body",
  );
  // How far down the body sits at a level: a full body shows `peak`.
  const offsetAt = (level: number) => r1(bodyH - (o.box.h * clamp(level, 0, 1.2) + amp));
  anim(
    ctx,
    body,
    levels.map((l) => ({
      offset: (l.at - start) / dur,
      transform: `translateY(${l.level <= 0 ? bodyH : offsetAt(l.level)}px)`,
    })),
    start,
    dur,
    EASE.sineInOut,
  );
  const surface = svg(
    body,
    `0 0 ${r1(o.box.w * 2)} ${r1(amp * 3)}`,
    `<path d="${wavePath(o.seed ?? ctx.seed, o.box.w * 2, amp, amp * 3)}" fill="${o.surface}"/>`,
    { top: `${-amp}px`, width: "200%", height: `${amp * 3}px` },
    "fill-surface",
  );
  anim(
    ctx,
    surface,
    [{ transform: "translateX(0px)" }, { transform: `translateX(${r1(-o.box.w)}px)` }],
    start,
    dur,
  );
  const rnd = seeded(`${o.seed ?? ctx.seed}:bubbles`);
  const n = Math.max(0, Math.floor(o.bubbles ?? 0));
  for (let i = 0; i < n; i++) {
    const size = clamp(o.box.w * 0.012, 5, 16) * (0.6 + rnd() * 0.9);
    const x = rnd() * o.box.w;
    const y = amp + rnd() * Math.max(1, bodyH * 0.5);
    const life = dur * (0.12 + rnd() * 0.16);
    const at = start + dur * 0.12 + rnd() * Math.max(0, dur * 0.76 - life);
    const b = div(
      body,
      {
        width: `${r1(size)}px`,
        height: `${r1(size)}px`,
        borderRadius: "50%",
        border: `1.5px solid ${o.bubble ?? o.surface}`,
        background:
          "radial-gradient(circle at 35% 35%, rgba(255,255,255,0.35), rgba(255,255,255,0) 60%)",
        boxSizing: "border-box",
        opacity: "0",
      },
      "bubble",
    );
    const p = `translate(${r1(x)}px, ${r1(y)}px) translate(-50%, -50%)`;
    anim(
      ctx,
      b,
      [
        { offset: 0, opacity: 0, transform: `${p} translateY(10px) scale(0.3)` },
        { offset: 0.25, opacity: 1, transform: `${p} translateY(0px) scale(0.8)` },
        { offset: 0.85, opacity: 1, transform: `${p} translateY(-12px) scale(1.05)` },
        { offset: 1, opacity: 0, transform: `${p} translateY(-14px) scale(1.6)` },
      ],
      at,
      life,
      EASE.power1Out,
    );
  }
}

// ---- ooze ----

/**
 * A spring of sludge: where it wells up out of the ground, how wide it
 * grows (its radius at the surface, board px) and its life in scene ms:
 * it starts at `at`, is full by `full`, starts draining at `drain` and is
 * gone by `gone`.
 */
export interface Spring {
  x: number;
  y: number;
  r: number;
  at: number;
  full: number;
  drain: number;
  gone: number;
  ph: number;
  /** A card's own spring: it wells up under the card first. */
  card: boolean;
}

export interface OozeTiming {
  start: number;
  /** How long a spring takes to well up to full. */
  rise: number;
  /** When the springs begin to drain, and when the last is gone. */
  drain: number;
  end: number;
  /** How much of the board the sludge covers at its height, about (0..1). */
  cover?: number;
}

/**
 * oozeSprings plans where sludge comes up across the board: a jittered
 * grid of springs, each with a smaller one beside it so the pools are not
 * round, and one under each card the scene takes, earliest. Seeded; every
 * spring is gone by `end`.
 */
export function oozeSprings(
  seed: string,
  board: { w: number; h: number },
  cards: readonly Pick<SceneCard, "x" | "y" | "width" | "height" | "rot">[],
  o: OozeTiming,
): Spring[] {
  const rnd = seeded(`${seed}:ooze`);
  const { w, h } = board;
  const cols = clamp(Math.round(w / 230), 3, 7);
  const rows = clamp(Math.round(h / 190), 2, 4);
  const n = cols * rows;
  const cover = clamp(o.cover ?? 0.75, 0.1, 1.2);
  const base = Math.sqrt((cover * w * h) / (n * Math.PI));
  const span = Math.max(1, o.end - o.drain);
  const life = (at: number, rise: number) => {
    const full = Math.min(o.drain - 1, at + rise);
    const drain = Math.max(full + 1, o.drain + rnd() * span * 0.25);
    const gone = clamp(o.end - rnd() * span * 0.15, drain + 1, o.end);
    return { at, full, drain, gone };
  };
  const out: Spring[] = [];
  cards.forEach((c, i) => {
    const b = screenBox(c);
    out.push({
      x: c.x,
      y: c.y,
      r: Math.max(b.w, b.h) * 0.62,
      ...life(o.start + i * 30, o.rise * 0.6),
      ph: rnd() * 6.28,
      card: true,
    });
  });
  for (let j = 0; j < rows; j++) {
    for (let i = 0; i < cols; i++) {
      const x = ((i + 0.5 + (rnd() - 0.5) * 0.6) / cols) * w;
      const y = ((j + 0.5 + (rnd() - 0.5) * 0.6) / rows) * h;
      const r = base * (0.75 + rnd() * 0.45);
      const t = life(o.start + rnd() * o.rise * 0.5, o.rise * (0.6 + rnd() * 0.4));
      out.push({ x, y, r, ...t, ph: rnd() * 6.28, card: false });
      const a = rnd() * Math.PI * 2;
      out.push({
        x: x + Math.cos(a) * r * 0.6,
        y: y + Math.sin(a) * r * 0.6,
        r: r * (0.45 + rnd() * 0.2),
        ...t,
        at: Math.min(t.full - 1, t.at + 80),
        ph: rnd() * 6.28,
        card: false,
      });
    }
  }
  return out.map((s) => ({ ...s, x: r1(s.x), y: r1(s.y), r: r1(s.r) }));
}

/** springRadius is a spring's radius at `ms`: it wells up, breathes, and drains back into the ground. */
export function springRadius(s: Spring, ms: number): number {
  if (ms <= s.at || ms >= s.gone) return 0;
  const k =
    ms < s.full
      ? easeOut3((ms - s.at) / (s.full - s.at))
      : ms > s.drain
        ? 1 - easeIn3((ms - s.drain) / (s.gone - s.drain))
        : 1;
  return s.r * k * (1 + 0.05 * Math.sin(ms / 230 + s.ph));
}

export interface OozeOpts extends OozeTiming {
  /** The cards it takes: a spring wells up under each. */
  cards: readonly SceneCard[];
  bubbles?: number;
  seed?: string;
}

// Very dark green, nearly black, with a sheen where the light catches it.
const OOZE_RIM = [22, 44, 20];
const OOZE_DEEP = [6, 15, 8];
const OOZE_SHINE = [110, 150, 80];
// The metaball kernel: (1 - d^2 / reach^2)^2, with reach = REACH r, is
// AT_R at d = r; dividing by it puts the surface (field 1) at r.
const REACH = 1.7;
const AT_R = (1 - 1 / (REACH * REACH)) ** 2;

/**
 * ooze wells sludge up out of the ground across the board, merging into
 * one tar pit that covers parts of the cards, bubbling and fuming, then
 * drains back down. Drawn as a metaball field (the springs' r^2 / d^2
 * summed on a coarse grid), shaded by its depth and lit from the upper
 * left, then scaled up smooth.
 */
export function ooze(ctx: SceneContext, o: OozeOpts): HTMLCanvasElement {
  const seed = o.seed ?? ctx.seed;
  const { w, h } = ctx.board;
  const k = boardScale(ctx.board);
  const springs = oozeSprings(seed, ctx.board, o.cards, o);
  const rnd = seeded(`${seed}:ooze-bubbles`);
  const bubbles: (Mote & { s: Spring })[] = [];
  const n = Math.max(0, Math.floor(o.bubbles ?? 30));
  for (let i = 0; i < n && springs.length > 0; i++) {
    const s = springs[Math.floor(rnd() * springs.length)];
    const life = 420 + rnd() * 380;
    const lo = s.full - 100;
    const hi = s.drain - life;
    if (hi <= lo) continue;
    const a = rnd() * Math.PI * 2;
    const d = rnd() * s.r * 0.6;
    bubbles.push({
      t: lo + rnd() * (hi - lo),
      life,
      x: s.x + Math.cos(a) * d,
      y: s.y + Math.sin(a) * d,
      vx: 0,
      vy: 0,
      r: (3 + rnd() * 5) * k,
      ph: rnd() * 6.28,
      s,
    });
  }
  const fumes: Mote[] = springs
    .filter((s, i) => !s.card && i % 2 === 0)
    .map((s) => ({
      t: s.full - 200 + rnd() * 300,
      life: Math.min(900 + rnd() * 500, Math.max(200, s.gone - s.full)),
      x: s.x + (rnd() - 0.5) * s.r * 0.5,
      y: s.y + (rnd() - 0.5) * s.r * 0.4,
      vx: (rnd() - 0.5) * 0.02,
      vy: -(0.025 + rnd() * 0.03) * k,
      r: s.r * 0.5,
      ph: rnd() * 6.28,
    }));

  const cell = Math.max(2.5, Math.sqrt((w * h) / 60000));
  const gw = Math.max(2, Math.ceil(w / cell));
  const gh = Math.max(2, Math.ceil(h / cell));
  const field = new Float32Array(gw * gh);
  let off: HTMLCanvasElement | null = null;
  let og: G | null = null;
  let img: ImageData | null = null;

  const paintField = (g: G, ms: number) => {
    if (!off) {
      off = document.createElement("canvas");
      off.width = gw;
      off.height = gh;
      og = context2d(off);
      img = og?.createImageData(gw, gh) ?? null;
    }
    if (!og || !img) return;
    field.fill(0);
    let any = false;
    for (const s of springs) {
      const r = springRadius(s, ms);
      if (r < 1) continue;
      any = true;
      // A soft kernel that falls to nothing at `reach` (so no seams), scaled
      // to be exactly 1 at the spring's own radius: lone, a pool is r wide;
      // near another, they reach for each other and run together.
      const reach = r * REACH;
      const inv = 1 / (reach * reach);
      const x0 = Math.max(0, Math.floor((s.x - reach) / cell));
      const x1 = Math.min(gw - 1, Math.ceil((s.x + reach) / cell));
      const y0 = Math.max(0, Math.floor((s.y - reach) / cell));
      const y1 = Math.min(gh - 1, Math.ceil((s.y + reach) / cell));
      for (let gy = y0; gy <= y1; gy++) {
        const dy = (gy + 0.5) * cell - s.y;
        const row = gy * gw;
        for (let gx = x0; gx <= x1; gx++) {
          const dx = (gx + 0.5) * cell - s.x;
          const q = 1 - (dx * dx + dy * dy) * inv;
          if (q > 0) field[row + gx] += (q * q) / AT_R;
        }
      }
    }
    if (!any) return;
    const data = img.data;
    for (let gy = 0; gy < gh; gy++) {
      for (let gx = 0; gx < gw; gx++) {
        const i = gy * gw + gx;
        const f = field[i];
        const p = i * 4;
        if (f < 0.8) {
          data[p + 3] = 0;
          continue;
        }
        const a = smooth(0.82, 1.0, f);
        const depth = clamp((f - 1) / 1.1, 0, 1);
        // The field rises toward a pool's middle; where it rises toward
        // the lower right, the surface faces the light (upper left).
        const fx = field[gy * gw + Math.min(gw - 1, gx + 1)] - field[gy * gw + Math.max(0, gx - 1)];
        const fy = field[Math.min(gh - 1, gy + 1) * gw + gx] - field[Math.max(0, gy - 1) * gw + gx];
        const shine = clamp((fx + fy) * 2.2, 0, 1) * (1 - depth) ** 2;
        const shade = clamp(-(fx + fy) * 1.2, 0, 1) * (1 - depth);
        for (let c = 0; c < 3; c++) {
          const base = OOZE_RIM[c] + (OOZE_DEEP[c] - OOZE_RIM[c]) * depth;
          data[p + c] = base * (1 - shade * 0.5) + (OOZE_SHINE[c] - base) * shine * 0.55;
        }
        data[p + 3] = a * 246;
      }
    }
    og.putImageData(img, 0, 0);
    g.imageSmoothingEnabled = true;
    g.globalAlpha = 1;
    g.drawImage(off, 0, 0, gw * cell, gh * cell);
  };

  // Slow glints of light sliding over the surface, drawn only on the sludge.
  const glints = springs
    .filter((s) => !s.card)
    .slice(0, 10)
    .map((s) => ({ s, a: rnd() * 6.28, d: rnd() * 0.5, sp: 0.0004 + rnd() * 0.0005 }));
  const fume = sprite("fume");
  const sheen = sprite("sheen");
  return canvasPart(ctx, {
    part: "ooze",
    start: o.start,
    dur: o.end - o.start,
    draw: (g, ms) => {
      paintField(g, ms);
      g.globalCompositeOperation = "source-atop";
      for (const gl of glints) {
        const r = springRadius(gl.s, ms);
        if (r < 8) continue;
        const a = gl.a + ms * gl.sp;
        const x = gl.s.x + Math.cos(a) * r * gl.d;
        const y = gl.s.y + Math.sin(a) * r * gl.d * 0.6;
        blot(g, sheen, x, y, r * 0.7, 0.5, r * 0.32);
      }
      // Bubbles swell on the surface and pop.
      for (const b of bubbles) {
        const u = ageOf(b, ms);
        if (u < 0) continue;
        const sr = springRadius(b.s, ms);
        if (Math.hypot(b.x - b.s.x, b.y - b.s.y) + b.r > sr * 0.9) continue;
        g.globalCompositeOperation = "source-over";
        if (u < 0.8) {
          const r = b.r * (0.35 + 0.65 * easeOut3(u / 0.8));
          g.globalAlpha = 0.95;
          g.fillStyle = "rgba(10,24,10,1)";
          g.strokeStyle = "rgba(96,140,62,0.85)";
          g.lineWidth = 1.1;
          g.beginPath();
          g.arc(b.x, b.y, r, 0, Math.PI * 2);
          g.fill();
          g.stroke();
          g.fillStyle = "rgba(200,235,170,0.55)";
          g.beginPath();
          g.arc(b.x - r * 0.35, b.y - r * 0.38, r * 0.22, 0, Math.PI * 2);
          g.fill();
        } else {
          const v = (u - 0.8) / 0.2;
          g.globalAlpha = 1 - v;
          g.strokeStyle = "rgba(110,156,72,0.9)";
          g.lineWidth = 1.2 * (1 - v) + 0.3;
          g.beginPath();
          g.arc(b.x, b.y, b.r * (1 + v * 1.4), 0, Math.PI * 2);
          g.stroke();
          g.fillStyle = "rgba(40,70,30,1)";
          for (let j = 0; j < 4; j++) {
            const a = b.ph + (j * Math.PI) / 2;
            g.beginPath();
            g.arc(
              b.x + Math.cos(a) * b.r * (1 + v * 2),
              b.y + Math.sin(a) * b.r * (1 + v * 2) - v * 3,
              1.1,
              0,
              Math.PI * 2,
            );
            g.fill();
          }
        }
      }
      // Rings run out from where a card goes down.
      g.globalCompositeOperation = "source-over";
      for (const s of springs) {
        if (!s.card) continue;
        for (const lag of [150, 520]) {
          const u = (ms - (s.full - 250 + lag)) / 700;
          if (u < 0 || u >= 1) continue;
          g.globalAlpha = (1 - u) * 0.6;
          g.strokeStyle = "rgba(92,134,62,1)";
          g.lineWidth = 1.4;
          g.beginPath();
          g.ellipse(
            s.x,
            s.y,
            s.r * (0.45 + 0.5 * easeOut3(u)),
            s.r * (0.4 + 0.45 * easeOut3(u)),
            0,
            0,
            Math.PI * 2,
          );
          g.stroke();
        }
      }
      // Fumes rise off the pit.
      g.globalCompositeOperation = "source-over";
      for (const p of fumes) {
        const u = ageOf(p, ms);
        if (u < 0) continue;
        const age = ms - p.t;
        const x = p.x + p.vx * age + Math.sin(age / 300 + p.ph) * 8 * k;
        blot(g, fume, x, p.y + p.vy * age, p.r * (0.6 + u), 0.55 * Math.sin(Math.PI * u));
      }
    },
  });
}

// ---- sink ----

export interface SinkOpts {
  card: SceneCard;
  start: number;
  dur: number;
}

/**
 * sink draws the card where it stood and lets it go down into the ground:
 * it tilts, shrinks away from the viewer and darkens to the sludge's
 * colour, into the pool welling up under it (drawn before it), then is
 * gone. It shows as it stood until `start`.
 */
export function sink(ctx: SceneContext, o: SinkOpts): HTMLElement {
  const c = o.card;
  const host = div(
    ctx.layer,
    {
      width: `${r1(c.width)}px`,
      height: `${r1(c.height)}px`,
      borderRadius: "8px",
      overflow: "hidden",
      background: "#0d1220",
      boxShadow: "0 4px 10px rgba(0,0,0,0.55)",
    },
    "sink",
  );
  if (c.src) {
    const img = document.createElement("img");
    img.src = c.src;
    img.alt = "";
    img.decoding = "async";
    img.draggable = false;
    Object.assign(img.style, {
      width: "100%",
      height: "100%",
      objectFit: "cover",
      display: "block",
    });
    host.appendChild(img);
  }
  const murk = div(
    host,
    {
      width: "100%",
      height: "100%",
      background:
        "linear-gradient(to top, rgba(4,12,5,1) 0%, rgba(8,22,9,0.92) 55%, rgba(14,34,14,0.75) 100%)",
      opacity: "0",
    },
    "sink-murk",
  );
  const rnd = seeded(`${ctx.seed}:${c.id}:sink`);
  const lean = (rnd() < 0.5 ? -1 : 1) * (3 + rnd() * 4);
  const at = (s: number, r: number, dy: number) =>
    `translate(${r1(c.x)}px, ${r1(c.y + dy)}px) translate(-50%, -50%) rotate(${r1(c.rot + r)}deg) scale(${s})`;
  anim(
    ctx,
    host,
    [
      { offset: 0, opacity: 1, transform: at(1, 0, 0) },
      { offset: 0.2, opacity: 1, transform: at(1.02, lean * 0.3, -2) },
      { offset: 0.65, opacity: 1, transform: at(0.86, lean, 6) },
      { offset: 1, opacity: 0, transform: at(0.62, lean * 1.4, 12) },
    ],
    o.start,
    o.dur,
    EASE.power1In,
  );
  anim(
    ctx,
    murk,
    [
      { offset: 0, opacity: 0 },
      { offset: 0.6, opacity: 0.75 },
      { offset: 1, opacity: 1 },
    ],
    o.start,
    o.dur,
    EASE.sineInOut,
  );
  return host;
}

// ---- The dragon ----

/**
 * A dragon in flight seen from above, facing right (+x), drawn for C2:
 * a 200 x 160 box, body on the middle line, the wings spread up and down.
 */
export const DRAGON_PATH =
  "M196 80 L178 72 L160 70 L138 69 L120 66 L108 38 L84 4 L88 22 L74 26 L78 42 L62 44 L68 58 L52 60 L90 70 " +
  "L62 76 L30 75 L8 82 L30 83 L62 84 L90 90 L52 100 L68 102 L62 116 L78 118 L74 134 L88 138 L84 156 " +
  "L108 122 L120 94 L138 91 L160 90 L178 88 Z";

function dragon(parent: HTMLElement, w: number, fill: string, part: string, glow?: string) {
  const host = div(parent, { width: `${r1(w)}px`, height: `${r1(w * 0.8)}px` }, part);
  const flap = svg(
    host,
    "0 0 200 160",
    `<path d="${DRAGON_PATH}" fill="${fill}"${glow ? ` style="filter:drop-shadow(0 0 6px ${glow})"` : ""}/>`,
    { width: "100%", height: "100%", transformOrigin: "50% 50%" },
    `${part}-wings`,
  );
  return { host, flap };
}

/** The wing beat: the silhouette narrows across its wings and opens again. */
function beat(ctx: SceneContext, flap: HTMLElement, start: number, dur: number): void {
  const beats = Math.max(1, Math.round(dur / 420));
  anim(
    ctx,
    flap,
    [
      { offset: 0, transform: "scale(1, 1)" },
      { offset: 0.5, transform: "scale(0.94, 0.55)" },
      { offset: 1, transform: "scale(1, 1)" },
    ],
    start,
    dur,
    EASE.sineInOut,
    beats,
  );
}

// ---- shadowPass ----

export interface ShadowOpts {
  start: number;
  dur: number;
  from: { x: number; y: number };
  to: { x: number; y: number };
  /** The shadow's wingspan, in board px. */
  width: number;
  opacity?: number;
}

/** headingDeg is the screen angle of travel from a to b, in degrees (0 = right, clockwise). */
export function headingDeg(a: { x: number; y: number }, b: { x: number; y: number }): number {
  return r1((Math.atan2(b.y - a.y, b.x - a.x) * 180) / Math.PI);
}

/** shadowPass sends a dragon's shadow across the board, soft-edged, growing a little as it goes. */
export function shadowPass(ctx: SceneContext, o: ShadowOpts): void {
  const { host, flap } = dragon(ctx.layer, o.width, "rgba(4,2,8,1)", "shadow");
  host.style.filter = `blur(${r1(clamp(o.width * 0.025, 3, 12))}px)`;
  host.style.opacity = "0";
  const rot = headingDeg(o.from, o.to);
  const op = o.opacity ?? 0.5;
  const at = (p: { x: number; y: number }, s: number) =>
    `translate(${r1(p.x)}px, ${r1(p.y)}px) translate(-50%, -50%) rotate(${rot}deg) scale(${s})`;
  anim(
    ctx,
    host,
    [
      { offset: 0, opacity: 0, transform: at(o.from, 0.9) },
      { offset: 0.15, opacity: op },
      { offset: 0.85, opacity: op },
      { offset: 1, opacity: 0, transform: at(o.to, 1.1) },
    ],
    o.start,
    o.dur,
    EASE.sineInOut,
  );
  beat(ctx, flap, o.start, o.dur);
}

// ---- strike ----

/**
 * tearPath is the torn hole a spear leaves in a card: a jagged star over a
 * 0..100 box, seeded.
 */
export function tearPath(seed: string, points = 9): string {
  const rnd = seeded(`${seed}:tear`);
  let d = "";
  for (let i = 0; i < points * 2; i++) {
    const a = (i / (points * 2)) * Math.PI * 2;
    const r = i % 2 === 0 ? 34 + rnd() * 16 : 10 + rnd() * 8;
    const x = 50 + Math.cos(a) * r;
    const y = 50 + Math.sin(a) * r;
    d += `${i === 0 ? "M" : " L"}${r1(x)} ${r1(y)}`;
  }
  return `${d} Z`;
}

/**
 * roughBladePath is a spearhead's blade, point up, over a 40 x 62 box:
 * a leaf shape whose edges are hammered rough and chipped, seeded.
 */
export function roughBladePath(seed: string, steps = 13): string {
  const rnd = seeded(`${seed}:blade`);
  const side = (sign: number) => {
    const pts: string[] = [];
    for (let i = 1; i <= steps; i++) {
      const t = i / steps;
      const hw = 3.6 + 9.4 * Math.sin(Math.PI * Math.min(t, 0.9) ** 0.8);
      const chip = i % 3 === 1 && i < steps ? 0.66 + rnd() * 0.12 : 0.86 + rnd() * 0.16;
      pts.push(`L${r1(20 + sign * hw * chip)} ${r1(t * 62 - rnd() * 1.6)}`);
    }
    return pts;
  };
  const right = side(1);
  const left = side(-1).reverse();
  return `M20 0 ${right.join(" ")} L24 62 L16 62 ${left.join(" ")} Z`;
}

// A spear seen coming at you, point up over a 40 x 140 box: a rough,
// dark iron head with a blooded point, a bound socket and a black shaft.
function spearSvg(seed: string): string {
  const blade = roughBladePath(seed);
  const binding = [0, 1, 2, 3, 4]
    .map(
      (i) =>
        `<path d="M14.5 ${75 + i * 2.6} L25.5 ${77.4 + i * 2.6}" stroke="#3a2a1c" stroke-width="1.5"/>`,
    )
    .join("");
  return (
    '<defs><linearGradient id="c2-spear-iron" x1="0" x2="1">' +
    '<stop offset="0" stop-color="#16171b"/><stop offset="0.46" stop-color="#5d616a"/>' +
    '<stop offset="0.54" stop-color="#2c2f35"/><stop offset="1" stop-color="#0e0f12"/></linearGradient>' +
    '<linearGradient id="c2-spear-blood" x1="0" x2="0" y1="0" y2="1">' +
    '<stop offset="0" stop-color="rgba(150,6,20,0.75)"/><stop offset="0.5" stop-color="rgba(120,0,16,0)"/></linearGradient>' +
    '<linearGradient id="c2-spear-shaft" x1="0" x2="1">' +
    '<stop offset="0" stop-color="#000"/><stop offset="0.38" stop-color="#202026"/>' +
    '<stop offset="0.6" stop-color="#060607"/><stop offset="1" stop-color="#000"/></linearGradient></defs>' +
    '<rect x="16.6" y="84" width="6.8" height="56" fill="url(#c2-spear-shaft)"/>' +
    '<rect x="15" y="72" width="10" height="16" rx="1.5" fill="#0b0b0d"/>' +
    binding +
    '<rect x="15.4" y="60" width="9.2" height="13" rx="2" fill="#141519" stroke="rgba(255,255,255,0.14)" stroke-width="0.6"/>' +
    `<path d="${blade}" fill="url(#c2-spear-iron)" stroke="rgba(214,218,228,0.6)" stroke-width="0.7" stroke-linejoin="bevel"/>` +
    `<path d="${blade}" fill="url(#c2-spear-blood)"/>` +
    '<path d="M20 4 L19.4 22 L20.6 38 L20 57" fill="none" stroke="rgba(190,196,208,0.4)" stroke-width="0.8"/>'
  );
}

export interface StrikeOpts {
  card: SceneCard;
  start: number;
  dur: number;
  /** The way the point comes out, in degrees (0 = up, clockwise). */
  angle?: number;
}

/**
 * strike drives a spear through the back of a card and out of its face:
 * the face tears open, the point bursts out toward the viewer with a red
 * flash and splinters, holds, and is drawn back.
 */
export function strike(ctx: SceneContext, o: StrikeOpts): void {
  const c = o.card;
  const box = screenBox(c);
  const pierce = o.start + o.dur * 0.22;
  const after = o.start + o.dur - pierce;
  const size = Math.min(box.w, box.h);
  const centre = `translate(${r1(c.x)}px, ${r1(c.y)}px) translate(-50%, -50%)`;

  const hole = svg(
    ctx.layer,
    "0 0 100 100",
    `<path d="${tearPath(ctx.seed)}" fill="#0b0507" stroke="rgba(255,236,226,0.9)" stroke-width="2" stroke-linejoin="round"/>`,
    { width: `${r1(size * 0.5)}px`, height: `${r1(size * 0.5)}px`, opacity: "0" },
    "tear",
  );
  anim(
    ctx,
    hole,
    [
      { offset: 0, opacity: 0, transform: `${centre} scale(0.1)` },
      { offset: 0.08, opacity: 1, transform: `${centre} scale(1)` },
      { offset: 0.8, opacity: 1, transform: `${centre} scale(1)` },
      { offset: 1, opacity: 0, transform: `${centre} scale(0.9)` },
    ],
    pierce,
    after,
    EASE.power2Out,
  );

  const flash = div(
    ctx.layer,
    {
      width: `${r1(size * 1.4)}px`,
      height: `${r1(size * 1.4)}px`,
      borderRadius: "50%",
      background:
        "radial-gradient(circle, rgba(255,245,240,0.95) 0%, rgba(230,40,50,0.7) 30%, rgba(120,0,20,0) 70%)",
      mixBlendMode: "screen",
      opacity: "0",
    },
    "strike-flash",
  );
  anim(
    ctx,
    flash,
    [
      { offset: 0, opacity: 1, transform: `${centre} scale(0.3)` },
      { offset: 1, opacity: 0, transform: `${centre} scale(1.5)` },
    ],
    pierce,
    Math.min(260, after),
    EASE.power2Out,
  );

  const rnd = seeded(`${ctx.seed}:splinters`);
  for (let i = 0; i < 7; i++) {
    const a = rnd() * Math.PI * 2;
    const dist = size * (0.35 + rnd() * 0.4);
    const s = 4 + rnd() * 6;
    const sp = div(
      ctx.layer,
      {
        width: `${r1(s)}px`,
        height: `${r1(s * 1.6)}px`,
        background: "linear-gradient(135deg, #efe6d6, #9c8f7a)",
        clipPath: "polygon(0 100%, 50% 0, 100% 70%)",
        opacity: "0",
      },
      "splinter",
    );
    anim(
      ctx,
      sp,
      [
        { offset: 0, opacity: 1, transform: `${centre} rotate(0deg) scale(0.6)` },
        {
          offset: 1,
          opacity: 0,
          transform: `${centre} translate(${r1(Math.cos(a) * dist)}px, ${r1(Math.sin(a) * dist)}px) rotate(${r1((rnd() * 2 - 1) * 300)}deg) scale(1.6)`,
        },
      ],
      pierce,
      Math.min(420, after),
      EASE.power2Out,
    );
  }

  // A spray of blood and grit where it tears through, on a canvas.
  const rs = seeded(`${ctx.seed}:spray`);
  const k = boardScale(ctx.board);
  const drops = burst(rs, c.x, c.y, pierce, 22, 0.5 * k, [260, 560]);
  const grit = burst(rs, c.x, c.y, pierce, 10, 0.32 * k, [300, 520]);
  const spray = sprite("spray");
  const glow = sprite("glow");
  const sprayEnd = Math.min(o.start + o.dur, pierce + 620);
  canvasPart(ctx, {
    part: "strike-spray",
    start: pierce,
    dur: sprayEnd - pierce,
    draw: (g, ms) => {
      g.globalCompositeOperation = "lighter";
      const u = (ms - pierce) / 220;
      if (u < 1) blot(g, glow, c.x, c.y, size * (0.4 + 0.5 * easeOut3(u)), (1 - u) * 0.6);
      g.globalCompositeOperation = "source-over";
      for (const p of drops) {
        const v = ageOf(p, ms);
        if (v < 0) continue;
        const at = flown(p, ms, 140);
        blot(g, spray, at.x, at.y, (2 + p.r * 2.4) * k * (1 - v * 0.5), (1 - v) * 0.95);
      }
      g.fillStyle = "rgba(20,14,12,1)";
      for (const p of grit) {
        const v = ageOf(p, ms);
        if (v < 0) continue;
        const at = flown(p, ms, 180);
        g.globalAlpha = 1 - v;
        g.fillRect(at.x - 1.2 * k, at.y - 1.2 * k, 2.4 * k * p.r, 2.4 * k * p.r);
      }
    },
  });

  // The spear, from the hole toward the viewer: its rough point and black
  // shaft grow (nearer) as they come out, hold, and are drawn back in.
  const len = box.h * 1.15;
  const head = svg(
    ctx.layer,
    "0 0 40 140",
    spearSvg(ctx.seed),
    {
      width: `${r1((len * 40) / 140)}px`,
      height: `${r1(len)}px`,
      transformOrigin: "50% 100%",
      opacity: "0",
      filter:
        "drop-shadow(0 0 1.5px rgba(255,214,204,0.45)) drop-shadow(0 8px 10px rgba(0,0,0,0.6))",
    },
    "spear",
  );
  const turn = o.angle ?? 28;
  const from = `translate(${r1(c.x)}px, ${r1(c.y)}px) translate(-50%, -100%) rotate(${turn}deg)`;
  anim(
    ctx,
    head,
    [
      { offset: 0, opacity: 0, transform: `${from} scale(0.5, 0)` },
      { offset: 0.05, opacity: 1, transform: `${from} scale(0.6, 0.2)` },
      { offset: 0.35, opacity: 1, transform: `${from} scale(1.1, 1.05)` },
      { offset: 0.7, opacity: 1, transform: `${from} scale(1.12, 1)` },
      { offset: 1, opacity: 0, transform: `${from} scale(0.6, 0.1)` },
    ],
    pierce,
    after,
    EASE.power3Out,
  );
}

// ---- flyLand ----

export interface PathPoint {
  x: number;
  y: number;
  /** Fraction of the path's length at this point. */
  offset: number;
  /** Heading here, degrees. */
  rot: number;
}

/**
 * swoopPath is a dragon's flight: in from off the board, one loop around
 * it, then a dive onto `to`. Seeded: the loop runs clockwise or the other
 * way. Smoothed (Catmull-Rom), with each point's share of the length
 * (for keyframe offsets) and its heading.
 */
export function swoopPath(
  seed: string,
  board: { w: number; h: number },
  to: { x: number; y: number },
  samples = 6,
): PathPoint[] {
  const rnd = seeded(`${seed}:swoop`);
  const { w, h } = board;
  let ctrl = [
    { x: -0.18 * w, y: 0.22 * h },
    { x: 0.35 * w, y: 0.08 * h },
    { x: 0.82 * w, y: 0.14 * h },
    { x: 0.92 * w, y: 0.55 * h },
    { x: 0.6 * w, y: 0.84 * h },
    { x: 0.2 * w, y: 0.7 * h },
    { x: 0.16 * w, y: 0.36 * h },
  ];
  if (rnd() < 0.5) ctrl = ctrl.map((p) => ({ x: w - p.x, y: p.y }));
  // The dive: from the loop's end, up and over, onto the card.
  const last = ctrl[ctrl.length - 1];
  const dive = { x: (last.x + to.x) / 2, y: Math.min(last.y, to.y) - 0.12 * h };
  const pts = [...ctrl, dive, { x: to.x, y: to.y }];

  const sampled: { x: number; y: number }[] = [];
  const at = (i: number) => pts[clamp(i, 0, pts.length - 1)];
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = at(i - 1);
    const p1 = at(i);
    const p2 = at(i + 1);
    const p3 = at(i + 2);
    for (let s = 0; s < samples; s++) {
      const t = s / samples;
      const t2 = t * t;
      const t3 = t2 * t;
      sampled.push({
        x:
          0.5 *
          (2 * p1.x +
            (-p0.x + p2.x) * t +
            (2 * p0.x - 5 * p1.x + 4 * p2.x - p3.x) * t2 +
            (-p0.x + 3 * p1.x - 3 * p2.x + p3.x) * t3),
        y:
          0.5 *
          (2 * p1.y +
            (-p0.y + p2.y) * t +
            (2 * p0.y - 5 * p1.y + 4 * p2.y - p3.y) * t2 +
            (-p0.y + 3 * p1.y - 3 * p2.y + p3.y) * t3),
      });
    }
  }
  sampled.push({ x: to.x, y: to.y });

  const lens = [0];
  for (let i = 1; i < sampled.length; i++) {
    const a = sampled[i - 1];
    const b = sampled[i];
    lens.push(lens[i - 1] + Math.hypot(b.x - a.x, b.y - a.y));
  }
  const total = lens[lens.length - 1] || 1;
  return sampled.map((p, i) => {
    const next = sampled[Math.min(i + 1, sampled.length - 1)];
    const prev = sampled[Math.max(0, i - 1)];
    const ref = i === sampled.length - 1 ? { a: prev, b: p } : { a: p, b: next };
    return {
      x: r1(p.x),
      y: r1(p.y),
      offset: Math.round((lens[i] / total) * 10000) / 10000,
      rot: headingDeg(ref.a, ref.b),
    };
  });
}

export interface FlyLandOpts {
  to: SceneCard;
  start: number;
  dur: number;
  /** Wingspan in flight, in board px. */
  width?: number;
  /** The colour it glows. */
  glow?: string;
}

/** A dragon's flight, planned: its path, when it flies and lands, and its size. */
export interface Flight {
  path: PathPoint[];
  /** Each point's heading, unwound so a turn never spins the long way. */
  rots: number[];
  start: number;
  flyDur: number;
  land: number;
  span: number;
  /** Its scale as it lands, shrunk onto the card. */
  landScale: number;
  board: { w: number; h: number };
}

/** The share of a flyLand's time the flight takes; the landing has the rest. */
const FLY_END = 0.82;
/** Where along its flight the dragon starts shrinking onto the card. */
const SHRINK_FROM = 0.78;

/** flight plans the swoop flyLand draws, so a scene can time other parts to it. */
export function flight(ctx: Pick<SceneContext, "seed" | "board">, o: FlyLandOpts): Flight {
  const box = screenBox(o.to);
  const span = o.width ?? clamp(ctx.board.w * 0.22, 140, 320);
  const path = swoopPath(ctx.seed, ctx.board, o.to);
  const flyDur = o.dur * FLY_END;
  let prev = path[0]?.rot ?? 0;
  const rots = path.map((p) => {
    let r = p.rot;
    while (r - prev > 180) r -= 360;
    while (r - prev < -180) r += 360;
    prev = r;
    return r;
  });
  return {
    path,
    rots,
    start: o.start,
    flyDur,
    land: o.start + flyDur,
    span,
    landScale: r1((box.w / span) * 1.6 * 100) / 100,
    board: ctx.board,
  };
}

/** flightAt is where the dragon is at scene ms `ms`, and its heading: its keyframes, read back. */
export function flightAt(f: Flight, ms: number): { x: number; y: number; rot: number } {
  const u = clamp((ms - f.start) / Math.max(1, f.flyDur), 0, 1);
  const p = f.path;
  let i = 1;
  while (i < p.length - 1 && p[i].offset < u) i++;
  const a = p[i - 1];
  const b = p[i];
  const s = b.offset > a.offset ? (u - a.offset) / (b.offset - a.offset) : 0;
  return {
    x: a.x + (b.x - a.x) * s,
    y: a.y + (b.y - a.y) * s,
    rot: f.rots[i - 1] + (f.rots[i] - f.rots[i - 1]) * s,
  };
}

/** mouthAt is the tip of the dragon's snout at `ms` (its nose is at the front of its box). */
export function mouthAt(f: Flight, ms: number): { x: number; y: number } {
  const p = flightAt(f, ms);
  const a = (p.rot * Math.PI) / 180;
  const reach = f.span * 0.47;
  return { x: p.x + Math.cos(a) * reach, y: p.y + Math.sin(a) * reach };
}

/**
 * breathWindow is when, during its loop, the dragon is best placed to
 * breathe on `target`: over the board, facing it, neither on top of it
 * nor across the board, and well before it dives to land. Null when there
 * is no such moment.
 */
export function breathWindow(
  f: Flight,
  target: { x: number; y: number; r: number },
  dur: number,
): { start: number; dur: number } | null {
  const lo = f.start;
  const hi = f.land - 300 - dur;
  if (hi <= lo) return null;
  let best: { at: number; score: number } | null = null;
  for (let at = lo; at <= hi; at += 10) {
    const mid = at + dur / 2;
    const m = mouthAt(f, mid);
    const p = flightAt(f, mid);
    const dx = target.x - m.x;
    const dy = target.y - m.y;
    const d = Math.hypot(dx, dy);
    if (d < target.r + 40) continue;
    if (m.x < 0 || m.y < 0 || m.x > f.board.w || m.y > f.board.h) continue;
    const a = (p.rot * Math.PI) / 180;
    const facing = (Math.cos(a) * dx + Math.sin(a) * dy) / d;
    if (facing < 0.2) continue;
    const ideal = f.span * 1.4;
    const score = facing - Math.abs(d - ideal) / (ideal * 2.5);
    if (!best || score > best.score) best = { at, score };
  }
  return best ? { start: Math.round(best.at), dur } : null;
}

/** aimSeat picks the seat a scene aims at: a foe of the card's controller, seeded, else any seat. */
export function aimSeat(ctx: Pick<SceneContext, "seed" | "seats">): SceneSeat | null {
  const seats = ctx.seats ?? [];
  const foes = seats.filter((s) => s.foe);
  const pool = foes.length > 0 ? foes : seats;
  if (pool.length === 0) return null;
  return pool[Math.floor(seeded(`${ctx.seed}:seat`)() * pool.length)];
}

export interface BreathOpts {
  flight: Flight;
  target: SceneSeat;
  /** The breath, in scene ms: when fire leaves his mouth, and for how long. */
  start: number;
  dur: number;
  /** When the last flame on the target is out. */
  until: number;
}

/**
 * breath has the dragon breathe a stream of fire from its mouth, as it
 * flies, onto a seat's circle, which flares and burns for a moment.
 */
export function breath(ctx: SceneContext, o: BreathOpts): HTMLCanvasElement {
  const k = boardScale(ctx.board);
  const rnd = seeded(`${ctx.seed}:breath`);
  const T = o.target;
  const until = Math.max(o.start + o.dur, o.until);
  const stream: (Mote & { tx: number; ty: number; travel: number; reach: number })[] = [];
  let firstHit = Infinity;
  for (let t = o.start; t < o.start + o.dur; t += 2.5) {
    const m = mouthAt(o.flight, t);
    const aim = Math.atan2(T.y - m.y, T.x - m.x) + (rnd() - 0.5) * 0.2;
    const d = Math.hypot(T.x - m.x, T.y - m.y);
    const travel = clamp(d / (1.15 * k), 180, 420);
    const life = Math.min(until - t, travel * (1.05 + rnd() * 0.25));
    if (life < 60) break;
    firstHit = Math.min(firstHit, t + travel);
    stream.push({
      t,
      life,
      x: m.x,
      y: m.y,
      vx: Math.cos(aim),
      vy: Math.sin(aim),
      r: (4 + rnd() * 3) * k,
      ph: rnd() * 6.28,
      tx: T.x,
      ty: T.y,
      travel,
      reach: d * (0.9 + rnd() * 0.14),
    });
  }
  const ring: Mote[] = [];
  for (let t = firstHit; t < until - 160; t += 7) {
    const a = rnd() * Math.PI * 2;
    const rr = T.r * (0.8 + rnd() * 0.35);
    ring.push({
      t,
      life: Math.min(until - t, 260 + rnd() * 200),
      x: T.x + Math.cos(a) * rr,
      y: T.y + Math.sin(a) * rr,
      vx: Math.cos(a) * 0.01,
      vy: -(0.05 + rnd() * 0.05) * k,
      r: Math.max(6, T.r * (0.3 + rnd() * 0.18)),
      ph: rnd() * 6.28,
    });
  }
  const sparks = Number.isFinite(firstHit)
    ? burst(rnd, T.x, T.y, firstHit, 14, 0.5 * k, [240, 460])
    : [];
  const core = sprite("core");
  const fire = sprite("fire");
  const ember = sprite("ember");
  const glow = sprite("glow");
  const smoke = sprite("smoke");
  return canvasPart(ctx, {
    part: "breath",
    start: o.start,
    dur: until - o.start,
    draw: (g, ms) => {
      const pos = (p: (typeof stream)[number], age: number) => {
        const u = age / p.travel;
        const go = p.reach * easeOut3(Math.min(1, u)) + Math.max(0, age - p.travel) * 0.04;
        const wob = Math.sin(age / 36 + p.ph) * 5 * k * Math.min(1, u);
        return { x: p.x + p.vx * go - p.vy * wob, y: p.y + p.vy * go + p.vx * wob, u };
      };
      g.globalCompositeOperation = "source-over";
      stream.forEach((p, i) => {
        const v = ageOf(p, ms);
        if (v < 0.45 || i % 3 !== 0) return;
        const at = pos(p, ms - p.t);
        blot(g, smoke, at.x, at.y - 8 * k * v, p.r * (3 + v * 3), 0.32 * (1 - v));
      });
      g.globalCompositeOperation = "lighter";
      // The seat takes it: a glow, a flash, flames around its circle.
      if (ms >= firstHit) {
        const on = smooth(firstHit, firstHit + 90, ms) * (1 - smooth(until - 260, until, ms));
        const flick = 1 + 0.1 * Math.sin(ms / 40) + 0.05 * Math.sin(ms / 13);
        blot(g, glow, T.x, T.y, T.r * 2.6 * flick, 0.85 * on);
        if (ms < firstHit + 260) {
          const u = (ms - firstHit) / 260;
          blot(g, core, T.x, T.y, T.r * (1 + 1.6 * easeOut3(u)), (1 - u) * 0.9);
        }
      }
      for (const p of ring) {
        const u = ageOf(p, ms);
        if (u < 0) continue;
        const age = ms - p.t;
        const x = p.x + p.vx * age + Math.sin(age / 50 + p.ph) * 2 * k;
        const r = p.r * (1 - u * 0.6) * (u < 0.1 ? u / 0.1 : 1);
        blot(g, u < 0.4 ? fire : ember, x, p.y + p.vy * age, r * 0.8, (1 - u) * 0.6, r * 1.5);
      }
      // The stream: hottest as it leaves his mouth, reddening as it goes.
      for (const p of stream) {
        const v = ageOf(p, ms);
        if (v < 0) continue;
        const at = pos(p, ms - p.t);
        const r = p.r * (1 + 4.5 * Math.min(1, at.u));
        const a = (1 - v) * (v < 0.06 ? v / 0.06 : 1) * 0.5;
        blot(g, at.u < 0.55 ? fire : ember, at.x, at.y, r, a);
        if (at.u < 0.3) blot(g, core, at.x, at.y, r * 0.55, a);
      }
      // His mouth glows while he breathes.
      if (ms < o.start + o.dur) {
        const m = mouthAt(o.flight, ms);
        blot(g, glow, m.x, m.y, 26 * k, 0.9 * smooth(o.start, o.start + 60, ms));
      }
      drawSparks(g, sparks, ms, "rgba(255,210,140,1)", k);
    },
  });
}

/**
 * flyLand flies a dragon around the board and lands it as the card: it
 * shrinks onto the card's box as it dives, and the card's own image takes
 * its place with a shockwave, then gives way to the real card under it.
 */
export function flyLand(ctx: SceneContext, o: FlyLandOpts): void {
  const to = o.to;
  const box = screenBox(to);
  const f = flight(ctx, o);
  const { path, rots, span, flyDur, land, landScale } = f;
  const frames = (dx: number, dy: number, op: number): Keyframe[] =>
    path.map((p, i) => {
      const s =
        p.offset < SHRINK_FROM
          ? 1
          : 1 + (landScale - 1) * ((p.offset - SHRINK_FROM) / (1 - SHRINK_FROM));
      return {
        offset: p.offset,
        opacity: i === 0 ? 0 : i === path.length - 1 ? 0 : op,
        transform: `translate(${r1(p.x + dx)}px, ${r1(p.y + dy)}px) translate(-50%, -50%) rotate(${r1(rots[i])}deg) scale(${r1(s * 100) / 100})`,
      };
    });

  // Its shadow on the table, offset as if lit from the upper left.
  const shade = dragon(ctx.layer, span, "rgba(4,2,8,1)", "flight-shadow");
  shade.host.style.filter = `blur(${r1(clamp(span * 0.03, 3, 10))}px)`;
  shade.host.style.opacity = "0";
  anim(ctx, shade.host, frames(span * 0.18, span * 0.32, 0.35), o.start, flyDur, EASE.linear);
  beat(ctx, shade.flap, o.start, flyDur);

  const glow = o.glow ?? "rgba(180,40,90,0.85)";
  const body = dragon(ctx.layer, span, "#140b1c", "dragon", glow);
  body.host.style.opacity = "0";
  anim(ctx, body.host, frames(0, 0, 1), o.start, flyDur, EASE.linear);
  beat(ctx, body.flap, o.start, flyDur);

  const after = Math.max(1, o.start + o.dur - land);
  const at = `translate(${r1(to.x)}px, ${r1(to.y)}px) translate(-50%, -50%)`;
  const ring = div(
    ctx.layer,
    {
      width: `${r1(box.w * 1.3)}px`,
      height: `${r1(box.h * 1.1)}px`,
      borderRadius: "50%",
      border: "3px solid rgba(255,214,140,0.9)",
      boxShadow: `0 0 24px 6px ${glow}`,
      boxSizing: "border-box",
      opacity: "0",
    },
    "landing-ring",
  );
  anim(
    ctx,
    ring,
    [
      { offset: 0, opacity: 0, transform: `${at} scale(0.6)` },
      { offset: 0.04, opacity: 1, transform: `${at} scale(0.65)` },
      { offset: 1, opacity: 0, transform: `${at} scale(2.2)` },
    ],
    land,
    after,
    EASE.power2Out,
  );
  if (to.src) {
    const card = div(
      ctx.layer,
      {
        width: `${r1(to.width)}px`,
        height: `${r1(to.height)}px`,
        borderRadius: "8px",
        overflow: "hidden",
        boxShadow: `0 0 26px 4px ${glow}`,
        opacity: "0",
      },
      "landing-card",
    );
    const img = document.createElement("img");
    img.src = to.src;
    img.alt = "";
    img.decoding = "async";
    img.draggable = false;
    Object.assign(img.style, {
      width: "100%",
      height: "100%",
      objectFit: "cover",
      display: "block",
    });
    card.appendChild(img);
    const cardAt = (s: number) => `${at} rotate(${to.rot}deg) scale(${s})`;
    anim(
      ctx,
      card,
      [
        { offset: 0, opacity: 0, transform: cardAt(1.3) },
        { offset: 0.35, opacity: 1, transform: cardAt(1) },
        { offset: 0.7, opacity: 1, transform: cardAt(1) },
        { offset: 1, opacity: 0, transform: cardAt(1) },
      ],
      land - after * 0.15,
      after * 1.15,
      EASE.power2Out,
    );
  }
}
