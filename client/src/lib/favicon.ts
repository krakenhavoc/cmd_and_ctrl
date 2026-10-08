// The tab icon, drawn in the selected skin's colours. The geometry is
// public/icons/favicon.svg's, which stays as the first-paint and no-JS
// fallback; favicon.test.ts holds the two together.

export interface FaviconColours {
  /** The tile: the skin's --bg. */
  bg: string;
  /** The four Cs: --accent, so a custom accent is followed too. */
  accent: string;
  /** The diamond: the skin's primary text colour, --fg. */
  mark: string;
}

const C_PATH =
  "M-10-40V-80L-30-100H-52A48 48 0 0 0-100-52V-30L-80-10H-40L-33-17L-48-32H-71A7 7 0 0 1-78-39V-52A26 26 0 0 1-52-78H-39A7 7 0 0 1-32-71V-48L-17-33Z";
const DIAMOND_PATH = "M0-38L38 0L0 38L-38 0Z";

/** A colour that is safe to put in an SVG attribute inside a data: URL. */
export function usableColour(v: string): boolean {
  const s = v.trim();
  // color-mix() and var() do not resolve inside a standalone favicon image.
  return s !== "" && !/color-mix|var\(|["'<>&]/.test(s);
}

export function faviconSvg({ bg, accent, mark }: FaviconColours): string {
  return (
    `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="64" height="64" viewBox="0 0 64 64">` +
    `<defs><path id="c" d="${C_PATH}"/></defs>` +
    `<rect width="64" height="64" rx="12" fill="${bg}"/>` +
    `<g transform="translate(32 32) scale(0.27)" fill="${accent}">` +
    `<use href="#c" xlink:href="#c"/>` +
    `<use href="#c" xlink:href="#c" transform="scale(-1 1)"/>` +
    `<use href="#c" xlink:href="#c" transform="scale(1 -1)"/>` +
    `<use href="#c" xlink:href="#c" transform="scale(-1 -1)"/>` +
    `<path d="${DIAMOND_PATH}" fill="${mark}"/>` +
    `</g></svg>`
  );
}

export function faviconDataUrl(c: FaviconColours): string {
  return "data:image/svg+xml," + encodeURIComponent(faviconSvg(c));
}

/** Point link[rel=icon] at the skin's icon; leaves the static one if any colour is unusable. */
export function applyFavicon(doc: Document, c: FaviconColours): void {
  if (![c.bg, c.accent, c.mark].every(usableColour)) return;
  const href = faviconDataUrl({ bg: c.bg.trim(), accent: c.accent.trim(), mark: c.mark.trim() });
  doc.querySelector('link[rel="icon"]')?.setAttribute("href", href);
}
