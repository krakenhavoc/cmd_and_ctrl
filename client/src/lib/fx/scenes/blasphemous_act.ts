// Blasphemous Act (ADR 0209 §4, #4397): the sky goes red, fireballs crash
// onto the battlefield, one onto each creature it kills, and the ground
// burns while ash comes down.
//
//   beat 1 (0-800 ms)     the sky reddens; the first volley lands, one
//                         fireball on each creature (they crumble at 650)
//   beat 2 (800-2400 ms)  a heavier fall; the ground burns; ash drifts down
//
// Timings are SCENES.blasphemous_act's (fx/signatures.ts).

import { overlay, rain, type SceneContext } from "./kit";

const SKY =
  "linear-gradient(to bottom, rgba(150,24,0,0.55) 0%, rgba(210,70,10,0.28) 38%, rgba(120,20,0,0) 72%)";
const GROUND =
  "linear-gradient(to top, rgba(255,110,20,0.32) 0%, rgba(200,50,0,0.16) 30%, rgba(120,20,0,0) 60%)";

export function play(ctx: SceneContext): void {
  const targets = ctx.cards.slice(0, 10).map((c) => ({ x: c.x, y: c.y }));
  if (ctx.short) {
    overlay(ctx, { background: SKY, start: 0, dur: 800, peak: 1, fade: 0.3 });
    rain(ctx, { kind: "fire", count: 6, start: 0, dur: 800, impact: true, targets });
    return;
  }
  overlay(ctx, { background: SKY, start: 0, dur: 2400, peak: 1, fade: 0.2 });
  rain(ctx, {
    kind: "fire",
    count: 8,
    start: 60,
    dur: 740,
    impact: true,
    burnUntil: 2400,
    targets,
  });
  rain(ctx, {
    kind: "fire",
    count: 14,
    start: 800,
    dur: 900,
    impact: true,
    burnUntil: 2400,
    seed: `${ctx.seed}:2`,
  });
  overlay(ctx, { background: GROUND, start: 700, dur: 1700, peak: 1, fade: 0.25 });
  rain(ctx, { kind: "ash", count: 22, start: 1250, dur: 1150, seed: `${ctx.seed}:ash` });
}
