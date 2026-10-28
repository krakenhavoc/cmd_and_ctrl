// Blasphemous Act (ADR 0209 §4, #4397, #4416): the sky goes red and balls
// of fire come down on the battlefield, one onto each creature it kills,
// each a round, white-hot ball trailing flame and smoke. Where one lands
// it bursts, throws sparks, scorches the ground and leaves it burning
// while ash and embers drift.
//
//   beat 1 (0-800 ms)      the sky reddens; the first volley comes in, one
//                          fireball at each creature (they crumble at 850)
//   beat 2 (800-3200 ms)   a heavier fall; the ground burns; ash drifts
//
// Timings are SCENES.blasphemous_act's (fx/signatures.ts).

import { overlay, rain, type SceneContext } from "./kit";

const SKY =
  "linear-gradient(to bottom, rgba(140,20,0,0.6) 0%, rgba(200,64,10,0.3) 38%, rgba(120,20,0,0) 72%)";
const GROUND =
  "linear-gradient(to top, rgba(255,110,20,0.3) 0%, rgba(200,50,0,0.14) 30%, rgba(120,20,0,0) 60%)";
const SMOKE = "radial-gradient(ellipse at 50% 30%, rgba(30,10,6,0) 30%, rgba(24,8,4,0.4) 100%)";

export function play(ctx: SceneContext): void {
  const targets = ctx.cards.slice(0, 10).map((c) => ({ x: c.x, y: c.y }));
  if (ctx.short) {
    overlay(ctx, { background: SKY, start: 0, dur: 800, peak: 1, fade: 0.3 });
    rain(ctx, {
      kind: "fire",
      count: 5,
      start: 0,
      dur: 640,
      impact: true,
      burnUntil: 800,
      targets,
    });
    return;
  }
  overlay(ctx, { background: SKY, start: 0, dur: 3200, peak: 1, fade: 0.18 });
  overlay(ctx, { background: SMOKE, start: 900, dur: 2300, peak: 1, fade: 0.3 });
  rain(ctx, {
    kind: "fire",
    count: 6,
    start: 80,
    dur: 1000,
    impact: true,
    burnUntil: 3150,
    targets,
  });
  rain(ctx, {
    kind: "fire",
    count: 12,
    start: 950,
    dur: 1150,
    impact: true,
    burnUntil: 3150,
    seed: `${ctx.seed}:2`,
  });
  overlay(ctx, { background: GROUND, start: 900, dur: 2300, peak: 1, fade: 0.25 });
  rain(ctx, { kind: "ash", count: 30, start: 1500, dur: 1650, seed: `${ctx.seed}:ash` });
}
