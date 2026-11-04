// deckedOutMotion — how the decked-out cue moves (#4879, ADR 0209).
//
// TableEffects.svelte mounts the cue's parts (the library's red pulse,
// the wisps, the skull clipped to the player's mat, and copies of the
// hand's cards) and calls deckedMotion once they are in the DOM. Every
// part runs on the Web Animations API from the cue's own clock: `ms`,
// already capped, scaled by animations.speed and compressed on a busy
// table (tableEffects.ts deckedTiming, busyScale), split by DECKED_BEATS.
// Every animation fills both ways, so a hand card holds at rest until it
// is flung and a part that has finished stays gone until the layer
// removes the cue. Nothing waits on it, and it touches no game state.

import { EASE, run, type Run } from "./motion";
import { DECKED_BEATS, type DeckedFling, type DeckedWisp } from "./tableEffects";

/** A box in px: its top-left corner and its size. */
export interface Box {
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface DeckedMotion {
  ms: number;
  /** The library pile, in the mat's own px. */
  lib: Box;
  /** The mat's size, px. */
  mat: { w: number; h: number };
  /** The skull's width at full size, px; its height is 1.2 times that. */
  skull: number;
  /** The board's diagonal, px: how far a flung card goes. */
  diag: number;
  wisps: readonly DeckedWisp[];
  flings: readonly DeckedFling[];
}

/** The part of the cue from share `a` to share `b`, as one animation's timing. */
function span(ms: number, a: number, b: number, easing: string = EASE.linear) {
  return {
    delay: a * ms,
    duration: Math.max(1, (b - a) * ms),
    easing,
    fill: "both" as const,
  };
}

/**
 * The skull's keyframes: it rises out of the pile at the pile's size,
 * swells and drifts until it fills the mat (DECKED_BEATS.fill), sways
 * once, then dissolves upward with a blur. Offsets are shares of the
 * skull's own span, from the rise to the end of the cue. Exported for
 * the test that holds its start to the pile and its middle to the mat.
 */
export function skullKeyframes(d: Pick<DeckedMotion, "lib" | "mat" | "skull">): Keyframe[] {
  const [r0] = DECKED_BEATS.rise;
  const o = (share: number) => Math.min(1, Math.max(0, (share - r0) / (1 - r0)));
  const h = d.skull * 1.2;
  // From the mat's centre (where the skull sits at rest) to the pile's.
  const dx = d.lib.x + d.lib.w / 2 - d.mat.w / 2;
  const dy = d.lib.y + d.lib.h / 2 - d.mat.h / 2;
  const s0 = Math.max(0.02, d.lib.h / Math.max(1, h));
  const at = (x: number, y: number, s: number, rot = 0) =>
    `translate(${round(x)}px, ${round(y)}px) scale(${round(s)}) rotate(${rot}deg)`;
  const sway = DECKED_BEATS.sway;
  const fill = DECKED_BEATS.fill;
  return [
    { offset: 0, opacity: 0, transform: at(dx, dy, s0), filter: "blur(0px)" },
    {
      offset: o(r0 + 0.05),
      opacity: 0.85,
      transform: at(dx, dy - d.lib.h * 0.25, s0 * 1.1),
      filter: "blur(0px)",
      easing: EASE.power2InOut,
    },
    { offset: o(fill), opacity: 0.82, transform: at(0, 0, 1), filter: "blur(0px)" },
    { offset: o(fill + (sway - fill) / 3), opacity: 0.82, transform: at(0, 0, 1, -5) },
    { offset: o(fill + ((sway - fill) * 2) / 3), opacity: 0.82, transform: at(0, 0, 1, 4) },
    { offset: o(sway), opacity: 0.8, transform: at(0, 0, 1), filter: "blur(0px)" },
    { offset: 1, opacity: 0, transform: at(0, -h * 0.22, 1.06), filter: "blur(12px)" },
  ];
}

function round(n: number): number {
  return Math.round(n * 100) / 100;
}

/** Starts every part of one decked-out cue; the returned runs cancel it. */
export function deckedMotion(node: HTMLElement, d: DeckedMotion): Run[] {
  const runs: Run[] = [];
  const B = DECKED_BEATS;

  const pulse = node.querySelector<HTMLElement>(".dk-pulse");
  if (pulse) {
    runs.push(
      run(
        pulse,
        [
          { opacity: 0, boxShadow: "0 0 0 0 rgba(255, 40, 40, 0)" },
          { offset: 0.35, opacity: 1, boxShadow: "0 0 22px 8px rgba(255, 46, 46, 0.85)" },
          { opacity: 0, boxShadow: "0 0 30px 12px rgba(255, 40, 40, 0)" },
        ],
        span(d.ms, B.pulse[0], B.pulse[1], EASE.power1Out),
      ),
    );
  }

  const [w0, w1] = B.wisps;
  node.querySelectorAll<HTMLElement>(".dk-wisp").forEach((el, i) => {
    const w = d.wisps[i];
    if (!w) return;
    const start = w0 + w.at * (w1 - w0);
    const end = Math.min(w1 + 0.15, start + 0.42);
    const lh = d.lib.h;
    runs.push(
      run(
        el,
        [
          { opacity: 0, transform: "translate(-50%, 0) scale(0.6)" },
          { offset: 0.2, opacity: 0.85 },
          {
            offset: 0.6,
            opacity: 0.6,
            transform: `translate(calc(-50% + ${round(w.sway * lh * 0.5)}px), ${round(-w.rise * lh * 0.6)}px) scale(1.1)`,
          },
          {
            opacity: 0,
            transform: `translate(calc(-50% + ${round(-w.sway * lh * 0.3)}px), ${round(-w.rise * lh)}px) scale(1.5)`,
          },
        ],
        span(d.ms, start, end, EASE.sineInOut),
      ),
    );
  });

  const skull = node.querySelector<HTMLElement>(".dk-skull");
  if (skull) runs.push(run(skull, skullKeyframes(d), span(d.ms, B.rise[0], 1)));

  const [f0, f1] = B.fling;
  node.querySelectorAll<HTMLElement>(".dk-card").forEach((el, i) => {
    const f = d.flings[i];
    if (!f) return;
    const reach = f.reach * d.diag;
    // Outward, and falling a little as it goes.
    const x = round(f.dx * reach);
    const y = round(f.dy * reach + d.diag * 0.2);
    runs.push(
      run(
        el,
        [
          { opacity: 1, transform: "translate(-50%, -50%) rotate(0deg)" },
          { offset: 0.75, opacity: 1 },
          {
            opacity: 0,
            transform: `translate(calc(-50% + ${x}px), calc(-50% + ${y}px)) rotate(${f.spin}deg)`,
          },
        ],
        span(d.ms, f0, f1, EASE.power1In),
      ),
    );
  });

  return runs;
}
