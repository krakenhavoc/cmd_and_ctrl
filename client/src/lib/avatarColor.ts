// Extracts a dominant accent colour from a player's avatar image so
// per-seat UI (phase indicator, priority pills, turn marker) can be
// tinted with a colour the player actually recognises as *theirs*.
// Default avatars fall back to the seatColor() palette; Discord-
// imported avatars are sampled.
//
// Algorithm:
//   1. Draw the avatar into a 16×16 offscreen canvas.
//   2. Bucket each pixel into one of 12 hue bins, ignoring pixels
//      that are transparent, near-black, near-white, or too
//      desaturated to read as a theme colour.
//   3. The modal bin wins; the bin's average RGB is lifted to a
//      consistent HSL target (high saturation, mid-light lightness)
//      so the colour pops against the dark chrome regardless of
//      whether the original avatar was muted pastel or garish neon.
//
// Caching:
//   Results are memoised per URL. Extraction is async (image load),
//   so getAvatarColor() returns the fallback immediately and fires
//   an optional `onResolved` callback once the real colour is
//   available. Consumers bridge that into Svelte $state so the UI
//   swaps in-place without a layout shift.
//
// CORS:
//   The /avatars/* endpoint is same-origin (proxied through the app
//   server), so no crossOrigin dance is required. We still set
//   crossOrigin="anonymous" defensively in case a future deployment
//   serves avatars from a separate CDN origin — browsers will honour
//   whatever the server returns for Access-Control-Allow-Origin.

import { avatarURL } from "./api";
import { seatColor } from "./colors";
import type { PlayerView } from "./protocol";

const CACHE = new Map<string, string>();
const PENDING = new Map<string, Promise<string>>();

function hasDocument(): boolean {
  return typeof document !== "undefined";
}

function extractColor(url: string): Promise<string | null> {
  if (!hasDocument()) return Promise.resolve(null);
  return new Promise<string | null>((resolve) => {
    const img = new Image();
    img.crossOrigin = "anonymous";
    img.onload = () => {
      try {
        const canvas = document.createElement("canvas");
        canvas.width = 16;
        canvas.height = 16;
        const ctx = canvas.getContext("2d");
        if (!ctx) return resolve(null);
        ctx.drawImage(img, 0, 0, 16, 16);
        const data = ctx.getImageData(0, 0, 16, 16).data;
        // 12 hue bins (30° each). Sum r/g/b in each bin so we can
        // average the winner rather than emit the bin's centre hue —
        // averaging keeps some of the original avatar's character
        // without letting one stray pixel drive the result.
        const bins: Array<{ r: number; g: number; b: number; n: number }> = Array.from(
          { length: 12 },
          () => ({ r: 0, g: 0, b: 0, n: 0 }),
        );
        for (let i = 0; i < data.length; i += 4) {
          const r = data[i];
          const g = data[i + 1];
          const b = data[i + 2];
          const a = data[i + 3];
          if (a < 128) continue;
          const { h, s, l } = rgbToHsl(r, g, b);
          // Filter grey / black / white: they're background on most
          // avatars and carry no identity. Thresholds are tuned on
          // the Discord default-colour grid + a handful of real
          // portraits — loosen if grayscale avatars need to land on
          // something other than the seat-palette fallback.
          if (s < 0.18 || l < 0.1 || l > 0.92) continue;
          const bin = Math.floor(h * 12) % 12;
          bins[bin].r += r;
          bins[bin].g += g;
          bins[bin].b += b;
          bins[bin].n += 1;
        }
        let bestIdx = -1;
        let bestN = 0;
        for (let i = 0; i < bins.length; i++) {
          if (bins[i].n > bestN) {
            bestN = bins[i].n;
            bestIdx = i;
          }
        }
        if (bestIdx === -1) return resolve(null);
        const bin = bins[bestIdx];
        const r = bin.r / bin.n;
        const g = bin.g / bin.n;
        const b = bin.b / bin.n;
        const { h, s, l } = rgbToHsl(r, g, b);
        // Theme-accent target: S≥0.65, L∈[0.55, 0.68]. Keeps the
        // colour legible on the dark surface and visually in the
        // same ballpark as the seat palette.
        const tuned = hslToHex(h, Math.max(s, 0.65), Math.min(Math.max(l, 0.55), 0.68));
        resolve(tuned);
      } catch {
        resolve(null);
      }
    };
    img.onerror = () => resolve(null);
    img.src = url;
  });
}

// getAvatarColor returns the best accent colour we have for this URL,
// or `fallback` when extraction hasn't completed / the image has no
// extractable colour. `onResolved` fires exactly once when the real
// colour is available; it's skipped entirely when `url` is null (no
// avatar to sample).
export function getAvatarColor(
  url: string | null,
  fallback: string,
  onResolved?: (color: string) => void,
): string {
  if (!url) return fallback;
  const cached = CACHE.get(url);
  if (cached) {
    // Deliver the cached colour via the same callback contract as
    // the async path so consumers don't need two code paths.
    if (onResolved) queueMicrotask(() => onResolved(cached));
    return cached;
  }
  let pending = PENDING.get(url);
  if (!pending) {
    pending = extractColor(url).then((c) => {
      const color = c ?? fallback;
      CACHE.set(url, color);
      return color;
    });
    PENDING.set(url, pending);
  }
  if (onResolved) void pending.then((c) => onResolved(c));
  return fallback;
}

// playerColor resolves a PlayerView to an accent colour, preferring
// the Discord-avatar sample when present and falling back to the
// deterministic seat palette when the seat hasn't claimed a Discord
// identity. Thin wrapper over getAvatarColor — most call sites want
// this form.
export function playerColor(seat: PlayerView, onResolved?: (color: string) => void): string {
  const url = avatarURL(seat.discord_id, seat.discord_avatar_hash);
  return getAvatarColor(url, seatColor(seat.seat), onResolved);
}

function rgbToHsl(r: number, g: number, b: number): { h: number; s: number; l: number } {
  r /= 255;
  g /= 255;
  b /= 255;
  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  const l = (max + min) / 2;
  if (max === min) return { h: 0, s: 0, l };
  const d = max - min;
  const s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
  let h: number;
  if (max === r) h = (g - b) / d + (g < b ? 6 : 0);
  else if (max === g) h = (b - r) / d + 2;
  else h = (r - g) / d + 4;
  h /= 6;
  return { h, s, l };
}

function hslToHex(h: number, s: number, l: number): string {
  if (s === 0) {
    const v = Math.round(l * 255);
    const hex = v.toString(16).padStart(2, "0");
    return `#${hex}${hex}${hex}`;
  }
  const hue2rgb = (p: number, q: number, t: number): number => {
    if (t < 0) t += 1;
    if (t > 1) t -= 1;
    if (t < 1 / 6) return p + (q - p) * 6 * t;
    if (t < 1 / 2) return q;
    if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
    return p;
  };
  const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
  const p = 2 * l - q;
  const r = Math.round(hue2rgb(p, q, h + 1 / 3) * 255);
  const g = Math.round(hue2rgb(p, q, h) * 255);
  const b = Math.round(hue2rgb(p, q, h - 1 / 3) * 255);
  return `#${[r, g, b].map((v) => v.toString(16).padStart(2, "0")).join("")}`;
}
