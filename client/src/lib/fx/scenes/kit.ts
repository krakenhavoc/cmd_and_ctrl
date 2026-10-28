// The scene kit (ADR 0209 §4, #4397): the parts every signature scene is
// built from.
//
//   overlay      a full-board wash (a red sky, a sickly green, a dark tint)
//   rain         particle rain: fireballs that crash and burn the ground,
//                or ash that drifts down
//   fill         a rising fill (tar, water) with a moving surface and
//                bubbles, over the whole board or over one card
//   shadowPass   a dragon's shadow crossing the board
//   strike       a spear through the back of a card and out the front
//   flyLand      a dragon that swoops around the board, then lands as the card
//
// Every part draws plain DOM (inline styles, inline SVG) inside the scene's
// layer and moves it with the Web Animations API through motion.run, so a
// skip (unmounting the layer) cancels all of it. Times are the scene's own
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
  /** Where it lands, as the card (arrive), or null. */
  landing: SceneCard | null;
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
  /** Fire only: each fireball bursts where it lands and leaves a burning scorch until `burnUntil`. */
  impact?: boolean;
  burnUntil?: number;
  /** Fire only: a fireball aimed at each of these, first. */
  targets?: readonly { x: number; y: number }[];
  seed?: string;
}

const FIREBALL =
  "radial-gradient(circle at 50% 84%, #fffbe6 0 9%, #ffd86e 17%, #ff8f1f 30%, rgba(255,70,0,0.75) 42%, rgba(255,40,0,0) 56%)," +
  "linear-gradient(to bottom, rgba(255,90,0,0) 0%, rgba(255,110,20,0.25) 40%, rgba(255,150,40,0.7) 74%, rgba(255,200,90,0) 100%)";
const FLASH =
  "radial-gradient(circle, rgba(255,252,230,0.95) 0%, rgba(255,190,80,0.8) 24%, rgba(255,90,10,0.45) 48%, rgba(255,60,0,0) 70%)";
const SCORCH =
  "radial-gradient(ellipse, rgba(255,150,50,0.65) 0%, rgba(150,40,0,0.6) 30%, rgba(30,8,2,0.55) 52%, rgba(0,0,0,0) 72%)";
const FLAME =
  "radial-gradient(ellipse at 50% 88%, rgba(255,250,210,0.95) 0%, rgba(255,190,60,0.9) 26%, rgba(255,90,10,0.7) 52%, rgba(255,40,0,0) 74%)";

/** rain drops fire or ash on the board. */
export function rain(ctx: SceneContext, o: RainOpts): void {
  const drops = rainDrops(o.seed ?? ctx.seed, o.count, ctx.board, o.kind, o.targets ?? []);
  const end = o.start + o.dur;
  const burnUntil = Math.max(end, o.burnUntil ?? end);
  for (const d of drops) {
    const t0 = o.start + d.at * o.dur;
    const tf = d.fall * o.dur;
    if (o.kind === "ash") {
      const flake = div(
        ctx.layer,
        {
          width: `${d.size}px`,
          height: `${d.size * 0.7}px`,
          borderRadius: "40%",
          background: "rgba(70,64,62,0.85)",
          boxShadow: "0 0 3px rgba(255,120,40,0.35)",
          opacity: "0",
        },
        "ash",
      );
      anim(
        ctx,
        flake,
        [
          {
            offset: 0,
            opacity: 0,
            transform: `translate(${d.fromX}px, ${d.fromY}px) rotate(0deg)`,
          },
          { offset: 0.15, opacity: 0.9 },
          { offset: 0.8, opacity: 0.7 },
          { offset: 1, opacity: 0, transform: `translate(${d.x}px, ${d.y}px) rotate(${d.rot}deg)` },
        ],
        t0,
        tf,
        EASE.sineInOut,
      );
      continue;
    }
    const ball = div(
      ctx.layer,
      {
        width: `${d.size}px`,
        height: `${d.size * 2.6}px`,
        background: FIREBALL,
        borderRadius: "50% 50% 50% 50% / 30% 30% 70% 70%",
        mixBlendMode: "screen",
        opacity: "0",
      },
      "fireball",
    );
    const at = (x: number, y: number) =>
      `translate(${x}px, ${y}px) translate(-50%, -84%) rotate(${d.rot}deg)`;
    anim(
      ctx,
      ball,
      [
        { offset: 0, opacity: 0, transform: at(d.fromX, d.fromY) },
        { offset: 0.12, opacity: 1 },
        { offset: 0.94, opacity: 1 },
        { offset: 1, opacity: 0, transform: at(d.x, d.y) },
      ],
      t0,
      tf,
      EASE.power1In,
    );
    if (!o.impact) continue;
    const hit = t0 + tf;
    const flash = div(
      ctx.layer,
      {
        width: `${d.size * 3.4}px`,
        height: `${d.size * 3.4}px`,
        borderRadius: "50%",
        background: FLASH,
        mixBlendMode: "screen",
        opacity: "0",
      },
      "impact",
    );
    const flashMs = Math.min(280, Math.max(1, burnUntil - hit));
    anim(
      ctx,
      flash,
      [
        {
          offset: 0,
          opacity: 1,
          transform: `translate(${d.x}px, ${d.y}px) translate(-50%, -50%) scale(0.2)`,
        },
        {
          offset: 1,
          opacity: 0,
          transform: `translate(${d.x}px, ${d.y}px) translate(-50%, -50%) scale(1.5)`,
        },
      ],
      hit,
      flashMs,
      EASE.power2Out,
    );
    const burn = Math.max(1, burnUntil - hit);
    const scorch = div(
      ctx.layer,
      {
        width: `${d.size * 2.6}px`,
        height: `${d.size * 1.3}px`,
        borderRadius: "50%",
        background: SCORCH,
        transform: `translate(${d.x}px, ${d.y}px) translate(-50%, -50%)`,
        opacity: "0",
      },
      "scorch",
    );
    anim(
      ctx,
      scorch,
      [
        { offset: 0, opacity: 0 },
        { offset: 0.06, opacity: 1 },
        { offset: 0.7, opacity: 0.85 },
        { offset: 1, opacity: 0 },
      ],
      hit,
      burn,
    );
    const flame = div(
      ctx.layer,
      {
        width: `${d.size * 0.9}px`,
        height: `${d.size * 1.5}px`,
        borderRadius: "50% 50% 45% 45% / 65% 65% 35% 35%",
        background: FLAME,
        mixBlendMode: "screen",
        transformOrigin: "50% 100%",
        opacity: "0",
      },
      "flame",
    );
    const base = `translate(${d.x}px, ${d.y}px) translate(-50%, -100%)`;
    anim(
      ctx,
      flame,
      [
        { offset: 0, opacity: 0, transform: `${base} scale(0.6, 0.3)` },
        { offset: 0.1, opacity: 1, transform: `${base} scale(1, 1.1)` },
        { offset: 0.3, opacity: 0.85, transform: `${base} scale(0.85, 0.75)` },
        { offset: 0.5, opacity: 1, transform: `${base} scale(1.05, 1.15)` },
        { offset: 0.72, opacity: 0.8, transform: `${base} scale(0.9, 0.8)` },
        { offset: 1, opacity: 0, transform: `${base} scale(0.7, 0.4)` },
      ],
      hit,
      burn,
    );
  }
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

// A spearhead seen coming at you: a leaf blade on a short socket, its
// point at the top of a 40 x 100 box.
const SPEARHEAD =
  '<path d="M20 0 C29 18 33 40 27 62 L24 70 L16 70 L13 62 C7 40 11 18 20 0 Z" fill="url(#c2-blade)"/>' +
  '<path d="M20 4 L20 66" stroke="rgba(255,255,255,0.75)" stroke-width="1.2"/>' +
  '<rect x="14" y="70" width="12" height="8" rx="2" fill="#6b5a3a"/>' +
  '<rect x="16" y="78" width="8" height="22" fill="#4a3826"/>' +
  '<defs><linearGradient id="c2-blade" x1="0" x2="1"><stop offset="0" stop-color="#8e99a8"/>' +
  '<stop offset="0.5" stop-color="#f4f7fb"/><stop offset="1" stop-color="#6d7787"/></linearGradient></defs>';

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

  // The point, from the hole toward the viewer: it grows (nearer) as it
  // comes out, holds, and is drawn back in.
  const len = box.h * 0.85;
  const head = svg(
    ctx.layer,
    "0 0 40 100",
    SPEARHEAD,
    {
      width: `${r1(len * 0.4)}px`,
      height: `${r1(len)}px`,
      transformOrigin: "50% 100%",
      opacity: "0",
      filter: "drop-shadow(0 8px 10px rgba(0,0,0,0.6))",
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

/**
 * flyLand flies a dragon around the board and lands it as the card: it
 * shrinks onto the card's box as it dives, and the card's own image takes
 * its place with a shockwave, then gives way to the real card under it.
 */
export function flyLand(ctx: SceneContext, o: FlyLandOpts): void {
  const to = o.to;
  const box = screenBox(to);
  const span = o.width ?? clamp(ctx.board.w * 0.22, 140, 320);
  const path = swoopPath(ctx.seed, ctx.board, to);
  const flyEnd = 0.82;
  const flyDur = o.dur * flyEnd;
  const land = o.start + flyDur;
  const landScale = r1((box.w / span) * 1.6 * 100) / 100;

  // Unwind the headings so a turn through 180 degrees never spins the long way.
  let prev = path[0]?.rot ?? 0;
  const rots = path.map((p) => {
    let r = p.rot;
    while (r - prev > 180) r -= 360;
    while (r - prev < -180) r += 360;
    prev = r;
    return r;
  });
  const shrinkFrom = 0.78;
  const frames = (dx: number, dy: number, op: number): Keyframe[] =>
    path.map((p, i) => {
      const s =
        p.offset < shrinkFrom
          ? 1
          : 1 + (landScale - 1) * ((p.offset - shrinkFrom) / (1 - shrinkFrom));
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
      { offset: 0, opacity: 1, transform: `${at} scale(0.6)` },
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
