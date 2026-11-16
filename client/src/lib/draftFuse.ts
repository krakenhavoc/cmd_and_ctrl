// draftFuse.ts — the pick clock as a fuse (#5848 §7): how much of the
// fuse is left, what colour it burns, and when it flashes.
//
// Pure functions. DraftFuse.svelte runs them every animation frame from
// the server's deadline; the text clock (draft.ts, secondsLeft) says the
// same thing in words.

/** The fuse flashes for this long before the pick is made for the viewer. */
export const FUSE_FLASH_MS = 8000;

type RGB = readonly [number, number, number];

// Where the fuse is when it takes each colour, as the fraction of the
// time that is left: blue down to 55 %, fully orange by 35 %, red by
// 15 %, and a deeper red as it goes out.
const BLUE: RGB = [61, 139, 255];
const ORANGE: RGB = [255, 140, 26];
const RED: RGB = [255, 60, 40];
const EMBER: RGB = [225, 20, 20];
const STOPS: readonly (readonly [number, RGB])[] = [
  [1, BLUE],
  [0.55, BLUE],
  [0.35, ORANGE],
  [0.15, RED],
  [0, EMBER],
];

/**
 * The part of the fuse still burning, 0 to 1: `msLeft` of `totalMs`.
 * More time than the total (a clock that started early) is a full fuse,
 * and no total is an empty one.
 */
export function fuseFraction(msLeft: number, totalMs: number): number {
  if (!(totalMs > 0) || !Number.isFinite(msLeft)) return 0;
  return Math.min(1, Math.max(0, msLeft / totalMs));
}

/** The fuse's colour as red, green and blue, blended between the stops. */
export function fuseRGB(fraction: number): [number, number, number] {
  const f = Math.min(1, Math.max(0, Number.isFinite(fraction) ? fraction : 0));
  for (let i = 0; i < STOPS.length - 1; i++) {
    const [f0, c0] = STOPS[i];
    const [f1, c1] = STOPS[i + 1];
    if (f <= f0 && f >= f1) {
      const t = (f0 - f) / (f0 - f1);
      return [0, 1, 2].map((j) => Math.round(c0[j] + (c1[j] - c0[j]) * t)) as [
        number,
        number,
        number,
      ];
    }
  }
  return [...EMBER];
}

/** The fuse's colour as a CSS value. */
export function fuseColour(fraction: number): string {
  const [r, g, b] = fuseRGB(fraction);
  return `rgb(${r}, ${g}, ${b})`;
}

/** True in the last eight seconds, until the clock is out. */
export function fuseFlashing(msLeft: number | null): boolean {
  return msLeft !== null && msLeft > 0 && msLeft <= FUSE_FLASH_MS;
}
