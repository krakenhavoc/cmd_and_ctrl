// Toxic Deluge (ADR 0209 §4, #4397, #4416): a near-black green sludge
// wells up out of the ground all across the battlefield, the pools
// spreading and running together until the board is a tar pit lapping
// over the cards. The creatures it kills sink into it and are gone; the
// pit bubbles and fumes, then drains back into the ground. The creatures
// that survive keep a faint stain until the turn ends (fx/tints.ts, drawn
// on their tiles, not here).
//
//   beat 1 (0-800 ms)      the ground darkens; sludge wells up under each
//                          creature it takes and from springs across the
//                          board
//   beat 2 (800-3400 ms)   the pools merge and cover the board; the dying
//                          sink (gone by about 2 s); it bubbles; it drains
//
// The scene draws the cards' exits itself (SCENES.toxic_deluge.ownsCards):
// each one sinks, so the generic death cue does not play over it.
// Timings are SCENES.toxic_deluge's (fx/signatures.ts).

import { ooze, overlay, sink, type SceneContext } from "./kit";

const MIASMA =
  "radial-gradient(ellipse at 50% 55%, rgba(16,34,12,0.16) 0%, rgba(8,18,6,0.3) 70%, rgba(4,10,4,0.45) 100%)";

export function play(ctx: SceneContext): void {
  const cards = ctx.cards.slice(0, 16);
  if (ctx.short) {
    overlay(ctx, { background: MIASMA, start: 0, dur: 800, fade: 0.3 });
    ooze(ctx, { cards, start: 0, rise: 330, drain: 520, end: 800, cover: 0.32, bubbles: 10 });
    cards.forEach((c, i) => sink(ctx, { card: c, start: 120 + i * 20, dur: 560 - i * 20 }));
    return;
  }
  overlay(ctx, { background: MIASMA, start: 0, dur: 3400, fade: 0.16 });
  ooze(ctx, { cards, start: 60, rise: 1250, drain: 2450, end: 3350, cover: 0.5, bubbles: 34 });
  // Each card it takes sinks into the pool that wells up under it.
  cards.forEach((c, i) => sink(ctx, { card: c, start: 480 + i * 70, dur: 1300 }));
}
