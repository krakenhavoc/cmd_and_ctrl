// Dragon's Approach (ADR 0209 §4, #4414): a dragon's shadow passes over the
// board as the spell resolves, and a second, lower one follows it.
//
//   beat 1 (0-800 ms)      the light goes; the shadow crosses the board
//   beat 2 (500-1900 ms)   a second, smaller shadow follows, lower down
//
// Timings are SCENES.dragons_approach's (fx/signatures.ts). The scene needs
// no card: it is drawn over the board, so it plays whatever the spell took.

import { overlay, shadowPass, type SceneContext } from "./kit";

const DUSK =
  "radial-gradient(ellipse at 50% 40%, rgba(40,8,6,0.12) 0%, rgba(26,6,4,0.4) 70%, rgba(12,3,2,0.55) 100%)";

export function play(ctx: SceneContext): void {
  const { w, h } = ctx.board;
  const first = {
    from: { x: -0.3 * w, y: 0.2 * h },
    to: { x: 1.3 * w, y: 0.6 * h },
    width: Math.max(260, w * 0.6),
  };
  if (ctx.short) {
    overlay(ctx, { background: DUSK, start: 0, dur: 800, fade: 0.3 });
    shadowPass(ctx, { start: 0, dur: 800, ...first, opacity: 0.55 });
    return;
  }
  overlay(ctx, { background: DUSK, start: 0, dur: 1900, fade: 0.2 });
  shadowPass(ctx, { start: 0, dur: 800, ...first, opacity: 0.55 });
  shadowPass(ctx, {
    start: 500,
    dur: 1400,
    from: { x: 1.3 * w, y: 0.55 * h },
    to: { x: -0.3 * w, y: 0.85 * h },
    width: Math.max(180, w * 0.4),
    opacity: 0.4,
  });
}
