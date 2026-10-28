// Toxic Deluge (ADR 0209 §4, #4397): poison wells up from the bottom of
// the table like a tar pit, bubbling, rises over the creatures it takes,
// and drains away.
//
//   beat 1 (0-800 ms)     a green-black tar wells up along the bottom edge
//   beat 2 (800-2400 ms)  it rises over the board; tar closes over each
//                         creature (they go at 900); it drains
//
// Timings are SCENES.toxic_deluge's (fx/signatures.ts).

import { fill, overlay, screenBox, type SceneContext } from "./kit";

const TAR = {
  surface: "rgba(96,140,34,0.92)",
  top: "rgba(40,62,16,0.94)",
  bottom: "rgba(8,12,4,0.97)",
  bubble: "rgba(170,230,80,0.85)",
};
const MIASMA =
  "linear-gradient(to top, rgba(90,140,20,0.3) 0%, rgba(50,80,10,0.18) 40%, rgba(20,30,0,0) 80%)";

export function play(ctx: SceneContext): void {
  const board = { x: 0, y: 0, w: ctx.board.w, h: ctx.board.h };
  if (ctx.short) {
    overlay(ctx, { background: MIASMA, start: 0, dur: 800, fade: 0.3 });
    fill(ctx, {
      box: board,
      levels: [
        { at: 0, level: 0 },
        { at: 450, level: 0.26 },
        { at: 800, level: 0 },
      ],
      ...TAR,
      bubbles: 10,
    });
    return;
  }
  overlay(ctx, { background: MIASMA, start: 0, dur: 2400, fade: 0.2 });
  fill(ctx, {
    box: board,
    levels: [
      { at: 0, level: 0 },
      { at: 800, level: 0.26 },
      { at: 1500, level: 0.58 },
      { at: 1850, level: 0.6 },
      { at: 2400, level: 0 },
    ],
    ...TAR,
    bubbles: 26,
  });
  // The pit takes each creature: tar closes over its card.
  ctx.cards.slice(0, 12).forEach((c, i) => {
    const b = screenBox(c);
    const pad = 6;
    const at = 650 + i * 40;
    fill(ctx, {
      box: { x: c.x - b.w / 2 - pad, y: c.y - b.h / 2 - pad, w: b.w + pad * 2, h: b.h + pad * 2 },
      levels: [
        { at, level: 0 },
        { at: at + 450, level: 1.05 },
        { at: 1900, level: 1.05 },
        { at: 2350, level: 0 },
      ],
      ...TAR,
      bubbles: 3,
      seed: `${ctx.seed}:${c.id}`,
    });
  });
}
