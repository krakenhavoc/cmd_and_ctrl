// Bitter Triumph (ADR 0209 §4, #4397): a spear stabs through the back of
// the target card and out of its face.
//
//   beat 1 (0-650 ms)     the face tears and the point bursts out toward
//                         you, with a red flash (the card goes at 550)
//   beat 2 (650-1500 ms)  the spear holds and is drawn back
//
// Timings are SCENES.bitter_triumph's (fx/signatures.ts). The scene needs
// a card: a Bitter Triumph whose target did not die plays the generic cues.

import { overlay, strike, type SceneContext } from "./kit";

const VIGNETTE = "radial-gradient(ellipse at center, rgba(60,0,10,0) 40%, rgba(60,0,12,0.45) 100%)";

export function play(ctx: SceneContext): void {
  const target = ctx.cards[0];
  if (!target) return;
  const dur = ctx.short ? 650 : 1500;
  overlay(ctx, { background: VIGNETTE, start: 0, dur, fade: 0.25 });
  strike(ctx, { card: target, start: 0, dur });
}
