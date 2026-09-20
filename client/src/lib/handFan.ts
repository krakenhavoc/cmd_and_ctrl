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
