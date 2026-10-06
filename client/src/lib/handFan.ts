// Hand fan geometry. Pure, so it can be tested without rendering the
// component: Hand.svelte does nothing with these but interpolate them
// into inline styles.

// Per-card fan angle in degrees. Caps the total fan spread so very
// large hands don't tip cards past sideways. Matches the Pixi math
// from drawHandFan in the old table.ts.
export function fanAngle(i: number, n: number): number {
  if (n <= 1) return 0;
  const maxStepDeg = 7; // ~0.12 rad
  const maxTotalDeg = 60;
  const step = Math.min(maxStepDeg, maxTotalDeg / (n - 1));
  const totalDeg = step * (n - 1);
  return -totalDeg / 2 + i * step;
}

export function fanLift(i: number, n: number): number {
  if (n <= 1) return 0;
  const center = (n - 1) / 2;
  const distFromCenter = Math.abs(i - center);
  return Math.round(distFromCenter * 3);
}

// #956 — how far each slot after the first pulls back over its
// predecessor, as a fraction of a card's width.
//
// The overlap used to be a constant, so the fan's width grew linearly
// with the card count: at the self hand's 0.5 that is 1 + (n-1)/2
// card-widths, i.e. 4 wide at seven cards and 7.5 at fourteen. Any
// panel narrower than that clipped the fan, because .panel sets
// overflow: hidden and --card-h has a 168px floor that stops the
// cards shrinking to fit.
//
// Tightening past a seven-card hand holds the fan inside about 4.5
// card-widths up to roughly twenty cards, which covers every hand
// size that occurs in play — #956 is a seven-to-fourteen card hand in
// a half-width panel. It is NOT a bound for all n: once CAP binds at
// about fifteen cards the overlap stops tightening and the width
// grows again at 0.15 card-widths per card. Holding 4.5 at sixty
// cards would need an overlap of 0.94, which leaves a 6% sliver of
// each card and stops reading as a fan at all, so the cap is the
// right trade and the limit is recorded in handFan.test.ts rather
// than papered over.
//
// CAP is the point past which the cards stop reading as separate
// cards; base is the layout's resting overlap (0.5 self, 0.62
// opponents' face-down fans, 0.85 stacked).
export function handOverlap(n: number, base: number): number {
  const CAP = 0.85;
  const LOOSE_UP_TO = 7;
  const PER_CARD = 0.045;
  if (n <= LOOSE_UP_TO) return base;
  return Math.min(CAP, base + (n - LOOSE_UP_TO) * PER_CARD);
}

// #2395 — how far the outermost card of a fan reaches past its own
// upright box, sideways, in px. A slot turns about its bottom centre,
// so its top corner swings out by h·sin θ and comes in by
// (w/2)(1 − cos θ); `tilt` scales every card's angle (fitFan below).
// The hand is clipped to its row at rest, but it lifts over the board
// on hover with nothing clipping it, and then this is how far the end
// cards stick out past the fan's upright width.
export function fanOverhang(n: number, cardW: number, cardH: number, tilt = 1): number {
  const rad = (Math.abs(fanAngle(0, n)) * tilt * Math.PI) / 180;
  return Math.max(0, cardH * Math.sin(rad) - (cardW / 2) * (1 - Math.cos(rad)));
}

// The overlap fitFan settles for before it gives up the fan's tilt
// altogether: at 0.7 every card shows 30% of its width, its name and
// some art. A row with room keeps the resting overlap and the full
// tilt; a narrower one tightens and flattens towards this.
export const FAN_SNUG = 0.7;
// The narrowest sliver of a card the fan will show before it scrolls
// instead, as a share of the card's width: 12% is 20px of a 168px
// card, still a target the pointer can rest on.
export const FAN_MIN_SLIVER = 0.12;

export interface FanFit {
  // The overlap to publish as --hand-overlap.
  overlap: number;
  // The share of fanAngle and fanLift to apply, 0 (flat) to 1.
  tilt: number;
  // True when even a flat fan at FAN_MIN_SLIVER is wider than the row:
  // the hand scrolls sideways instead of running under what is beside it.
  scroll: boolean;
}

// #2395 — fit a hand fan to the row it sits in. handOverlap sizes the
// fan by its card count alone, so a row narrower than that fan (the
// self hand beside the tutorial's coach card at 1440×900: seven
// 168px cards in 460px) let the end cards run out of the row, where
// the coach card, the piles above the hand's left end and the
// commander beside its right end cover them, and the pointer could not
// reach them.
//
// fitFan keeps the resting look wherever it fits. Where it does not, it
// tightens the overlap and, past FAN_SNUG, eases off the tilt a quarter
// at a time (a tilted fan is wider than an upright one by twice
// fanOverhang), then goes flat and tightens to FAN_MIN_SLIVER. Past
// that it asks for a scroll. The fan, end cards' corners included,
// then never reaches past `room`.
//
// `room` is the row's inner width in px; `cardW` and `cardH` one card's
// box. A zero room or card (not laid out yet, or jsdom) keeps the
// resting fan. `rotates` is false for the stacked layout, which has no
// tilt to give up.
export function fitFan(
  n: number,
  base: number,
  room: number,
  cardW: number,
  cardH: number,
  rotates = true,
): FanFit {
  const loose = handOverlap(n, base);
  if (n <= 1 || room <= 0 || cardW <= 0) return { overlap: loose, tilt: 1, scroll: false };
  const need = (tilt: number): number =>
    1 - (room - 2 * fanOverhang(n, cardW, cardH, tilt) - cardW) / ((n - 1) * cardW);
  const snug = Math.max(loose, FAN_SNUG);
  if (rotates) {
    for (const tilt of [1, 0.75, 0.5, 0.25]) {
      const overlap = Math.max(loose, need(tilt));
      if (overlap <= snug) return { overlap, tilt, scroll: false };
    }
  }
  const flat = Math.max(loose, need(0));
  const tightest = 1 - FAN_MIN_SLIVER;
  if (flat <= tightest) return { overlap: flat, tilt: 0, scroll: false };
  return { overlap: Math.max(loose, tightest), tilt: 0, scroll: true };
}
