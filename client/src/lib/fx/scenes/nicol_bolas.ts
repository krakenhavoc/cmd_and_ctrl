// Nicol Bolas (ADR 0209 §4, #4397), shared by every catalog printing named
// Nicol Bolas: a dragon's shadow passes over the table, then the dragon
// swoops around the battlefield and lands as the card, on the stack when
// he is cast or on the battlefield when he enters without being cast.
//
//   beat 1 (0-800 ms)      the light goes; his shadow crosses the board
//   beat 2 (800-2450 ms)   he swoops around the table and lands as the card
//
// Timings are SCENES.nicol_bolas's (fx/signatures.ts). With no card to
// land on (it could not be found on the board), beat 2 is left out.

import { flyLand, overlay, shadowPass, type SceneContext } from "./kit";

const DUSK =
  "radial-gradient(ellipse at 50% 40%, rgba(30,6,40,0.15) 0%, rgba(20,4,30,0.42) 70%, rgba(10,2,16,0.55) 100%)";

export function play(ctx: SceneContext): void {
  const { w, h } = ctx.board;
  const pass = {
    from: { x: -0.3 * w, y: 0.25 * h },
    to: { x: 1.3 * w, y: 0.7 * h },
    width: Math.max(260, w * 0.6),
  };
  if (ctx.short || !ctx.landing) {
    overlay(ctx, { background: DUSK, start: 0, dur: 800, fade: 0.3 });
    shadowPass(ctx, { start: 0, dur: 800, ...pass, opacity: 0.55 });
    return;
  }
  overlay(ctx, { background: DUSK, start: 0, dur: 2450, fade: 0.18 });
  shadowPass(ctx, { start: 0, dur: 800, ...pass, opacity: 0.55 });
  flyLand(ctx, { to: ctx.landing, start: 800, dur: 1650 });
}
