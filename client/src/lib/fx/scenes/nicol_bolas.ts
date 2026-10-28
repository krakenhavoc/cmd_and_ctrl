// Nicol Bolas (ADR 0209 §4, #4397, #4416), shared by every catalog
// printing named Nicol Bolas, when he enters the battlefield (cast and
// resolved, or put there): a dragon's shadow passes over the table, then
// the dragon swoops around the battlefield, breathes a burst of fire on
// an opponent's seat as he passes it, and lands as the card on his tile.
//
//   beat 1 (0-800 ms)      the light goes; his shadow crosses the board
//   beat 2 (900-3300 ms)   he swoops around the table, breathes fire on a
//                          seat, and lands as the card
//
// Timings are SCENES.nicol_bolas's (fx/signatures.ts). With no card to
// land on (it could not be found on the board), beat 2 is left out; with
// no seat on the board, he flies without breathing.

import {
  aimSeat,
  breath,
  breathWindow,
  flight,
  flyLand,
  overlay,
  shadowPass,
  type SceneContext,
} from "./kit";

const DUSK =
  "radial-gradient(ellipse at 50% 40%, rgba(30,6,40,0.15) 0%, rgba(20,4,30,0.42) 70%, rgba(10,2,16,0.55) 100%)";

/** The breath, in scene ms, and how long the seat burns after it. */
const BREATH_MS = 560;
const BURN_MS = 520;

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
  overlay(ctx, { background: DUSK, start: 0, dur: 3300, fade: 0.15 });
  shadowPass(ctx, { start: 0, dur: 950, ...pass, opacity: 0.55 });
  const fly = { to: ctx.landing, start: 900, dur: 2400 };
  const seat = aimSeat(ctx);
  if (seat) {
    const f = flight(ctx, fly);
    const win = breathWindow(f, seat, BREATH_MS);
    if (win) {
      breath(ctx, {
        flight: f,
        target: seat,
        start: win.start,
        dur: win.dur,
        until: Math.min(3260, win.start + win.dur + BURN_MS),
      });
    }
  }
  flyLand(ctx, fly);
}
