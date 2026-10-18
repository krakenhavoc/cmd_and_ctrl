// stackIdentityEdge — the pile's top item in its colours (ADR 0119
// amendment 2026-10-16 §C, call 21).
//
// The top card of the pile gets a 3px edge in the item's colours, in the
// language of the site's W U B R G mark (SiteHeader / SiteFooter): hard
// bands of the fixed `--identity-*` tokens, never a blend. A mono item
// is one colour all round; a multicolour item splits the edge into
// equal shares of its perimeter, in W U B R G order, clockwise from the
// top-left corner; a colourless item is `--identity-c`. It is still: the
// motion on the stack is the arrival trace's (§D).
//
// The colours are the card's own (CardView.colors, layer 5 included),
// for an ability its source's: the card view the item already carries
// for its preview. A source the viewer may not see has none, and keeps
// the edge it had before (the accent ring): the edge never guesses.
//
// Plain functions over plain data, pinned by stackIdentityEdge.test.ts;
// StackLanePile.svelte paints what comes back.

import type { CardView } from "./protocol";

const WUBRG = ["W", "U", "B", "R", "G"] as const;

/** The edge's width, in px. */
export const IDENTITY_EDGE_W = 3;

/**
 * The item's colours as WUBRG letters, in that order: [] for a
 * colourless card, null when there is no card the viewer may see (a
 * face-down spell, a hidden or vanished source), which draws no edge.
 */
export function identityLetters(
  card: Pick<CardView, "colors"> | null | undefined,
): string[] | null {
  if (!card) return null;
  const have = new Set((card.colors ?? []).map((c) => String(c).toUpperCase()));
  return WUBRG.filter((l) => have.has(l));
}

/** The fixed tokens the edge runs through: one per colour, steel for colourless. */
export function identityTokens(letters: readonly string[]): string[] {
  if (letters.length === 0) return ["var(--identity-c)"];
  return letters.map((l) => `var(--identity-${l.toLowerCase()})`);
}

/** Degrees, rounded to a hundredth, so the CSS is stable to compare. */
function deg(n: number): number {
  return Math.round(n * 100) / 100;
}

/**
 * The bearing of a point on the box's rim from its centre, in the
 * conic gradient's terms: degrees clockwise from straight up.
 */
function bearing(x: number, y: number): number {
  const a = (Math.atan2(x, -y) * 180) / Math.PI;
  return a < 0 ? a + 360 : a;
}

/**
 * The point `p` along the rim of a w×h box centred on 0,0, measured
 * clockwise from its top-left corner (along the top edge first).
 */
function rimPoint(p: number, w: number, h: number): { x: number; y: number } {
  const per = 2 * (w + h);
  let d = ((p % per) + per) % per;
  if (d <= w) return { x: -w / 2 + d, y: -h / 2 };
  d -= w;
  if (d <= h) return { x: w / 2, y: -h / 2 + d };
  d -= h;
  if (d <= w) return { x: w / 2 - d, y: h / 2 };
  d -= w;
  return { x: -w / 2, y: h / 2 - d };
}

/**
 * The CSS background for the edge: a flat token for a mono or colourless
 * item, else a hard-stop conic gradient whose bands each take an equal
 * share of the rim, W U B R G clockwise from the top-left corner.
 * `aspect` is the box's width over its height (a card is 488 / 680).
 *
 * A conic gradient splits by angle, and equal angles are not equal runs
 * of a rectangle's rim; so each boundary is placed at the bearing of the
 * rim point that sits k/n of the way round. On a 3px edge the band
 * boundaries are then where the eye measures them.
 *
 * Null when the item has no card the viewer may see.
 */
export function identityEdgeBackground(
  card: Pick<CardView, "colors"> | null | undefined,
  aspect: number,
): string | null {
  const letters = identityLetters(card);
  if (letters === null) return null;
  const tokens = identityTokens(letters);
  if (tokens.length === 1) return tokens[0];
  const w = aspect > 0 && Number.isFinite(aspect) ? aspect : 488 / 680;
  const h = 1;
  const per = 2 * (w + h);
  const start = bearing(-w / 2, -h / 2);
  const stops: string[] = [];
  let from = 0;
  tokens.forEach((t, i) => {
    let to = 360;
    if (i < tokens.length - 1) {
      const p = rimPoint(((i + 1) * per) / tokens.length, w, h);
      to = bearing(p.x, p.y) - start;
      if (to <= 0) to += 360;
    }
    stops.push(`${t} ${deg(from)}deg ${deg(to)}deg`);
    from = to;
  });
  return `conic-gradient(from ${deg(start)}deg, ${stops.join(", ")})`;
}
