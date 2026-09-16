// Is the board card's failed-art pip (#33) visible on a tapped card in
// a battlefield row?
//
// A tapped Card turns 90° clockwise, and the turn makes the tile its
// own stacking context, so the pip's z-index counts only inside it: a
// later tile that overlaps the pip paints over it. The land strip
// overlaps tapped lands by 35% of a width, and a tapped tile in an
// ordinary row overhangs its neighbours. Where the pip sits is CSS in
// Card.svelte, the overlaps are CSS in BattlefieldRow.svelte, and the
// client has no DOM to lay either out in — so this reads the numbers
// out of those files and does the geometry. It was measured against
// Chromium's hit-testing (elementFromPoint at the pip) and agrees:
// 22px, the old tapped position, is covered on every tapped land but
// the rightmost.
//
// If a regex here stops matching, the CSS moved: update the pattern,
// and keep the assertions.

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const source = (rel: string) => readFileSync(fileURLToPath(new URL(rel, import.meta.url)), "utf8");
const cardSvelte = source("./components/board/Card.svelte");
const rowSvelte = source("./components/board/BattlefieldRow.svelte");
const appCss = source("../app.css");

// block returns the declarations of the rule whose selector is exactly
// `selector`.
function block(css: string, selector: string): string {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const m = new RegExp(`(?:^|\\n)\\s*${escaped} \\{([^}]*)\\}`).exec(css);
  expect(m, `no rule for ${selector}`).not.toBeNull();
  return m![1];
}

function decl(body: string, prop: string): string {
  const m = new RegExp(`(?:^|[;\\s])${prop}:\\s*([^;]+);`).exec(body);
  expect(m, `no ${prop}`).not.toBeNull();
  return m![1].trim();
}

// px evaluates a length: `12px`, `0`, a percentage of `percentOf`, or
// a calc() over --card-w / --card-h.
function px(value: string, w: number, h: number, percentOf = 0): number {
  const pct = /^([\d.]+)%$/.exec(value);
  if (pct) return (Number(pct[1]) / 100) * percentOf;
  const expr = value
    .replace(/var\(--card-w[^)]*\)/g, String(w))
    .replace(/var\(--card-h[^)]*\)/g, String(h))
    .replace(/calc/g, "")
    .replace(/px/g, "");
  expect(expr, `cannot evaluate ${value}`).toMatch(/^[\d.\s()+\-*/]+$/);
  return Number(new Function(`return (${expr});`)());
}

const PIP = px(decl(block(appCss, ".card-art-error"), "width"), 0, 0);
const PIP_LEFT = px(decl(block(cardSvelte, ".card"), "--art-error-left"), 0, 0);
const tappedTop = decl(block(cardSvelte, ".card.tapped"), "--art-error-top");
const ROW_PAD_LEFT = px(decl(block(rowSvelte, ".row"), "padding").split(/\s+/)[1], 0, 0);

// The strip's margin-left rules, in source order, with a matcher and
// specificity each, so the cascade is resolved the way a browser does.
interface Item {
  tapped: boolean;
  first: boolean;
  prevTapped: boolean;
}
const STRIP_ITEM = '.row.strip .row-cards > [role="listitem"]';
const stripRules = [...rowSvelte.matchAll(/\n {2}(\.row\.strip \.row-cards > [^{]+?) \{([^}]*)\}/g)]
  .filter(([, , body]) => /margin-left:/.test(body))
  .map(([, selector, body]) => {
    const tail = selector.slice(STRIP_ITEM.length);
    const matches = (it: Item): boolean => {
      if (tail.startsWith(':not(.tapped) + [role="listitem"]')) {
        const own = tail.slice(':not(.tapped) + [role="listitem"]'.length);
        return !it.first && !it.prevTapped && ownMatches(own, it);
      }
      return ownMatches(tail, it);
    };
    return {
      selector,
      matches,
      specificity: (selector.replace(/:not\(/g, "(").match(/\.[\w-]+|\[[^\]]+\]|:[\w-]+/g) ?? [])
        .length,
      marginLeft: decl(body, "margin-left"),
    };
  });

function ownMatches(tail: string, it: Item): boolean {
  const parts: string[] = tail.match(/\.[\w-]+|:[\w-]+/g) ?? [];
  expect(tail.replace(/\.tapped|:first-child|:hover/g, ""), `unmodelled selector ${tail}`).toBe("");
  if (parts.includes(":hover")) return false;
  if (parts.includes(".tapped") && !it.tapped) return false;
  if (parts.includes(":first-child") && !it.first) return false;
  return true;
}

function stripMargin(it: Item, w: number, h: number): number {
  let best: (typeof stripRules)[number] | null = null;
  for (const r of stripRules) {
    if (r.matches(it) && (!best || r.specificity >= best.specificity)) best = r;
  }
  expect(best, "no margin rule matched").not.toBeNull();
  return px(best!.marginLeft, w, h);
}

// pipsInRow lays out a row of tiles and returns, for each, the pip's
// horizontal extent and the part of the row it may occupy unhidden:
// right of the row's clip edge and left of the next tile, which paints
// over it. Later tiles never reach further left than the next one, and
// every tile spans the band a tapped pip sits in, so horizontal is
// enough.
function pipsInRow(tapped: boolean[], w: number, h: number, strip: boolean) {
  const out: { pip: [number, number]; room: [number, number] }[] = [];
  const visualLeft: number[] = [];
  let x = 0;
  tapped.forEach((t, i) => {
    const margin = strip
      ? stripMargin({ tapped: t, first: i === 0, prevTapped: tapped[i - 1] ?? false }, w, h)
      : 0;
    x += i === 0 ? margin : w + (strip ? 0 : 8) + margin;
    visualLeft.push(t ? x + w / 2 - h / 2 : x);
    // Turning (dx, dy) from the tile's centre 90° clockwise gives
    // (-dy, dx): the pip's distance down the tile runs right to left.
    const top = px(tappedTop, w, h, h);
    const pip: [number, number] = t
      ? [x + w / 2 - (top + PIP - h / 2), x + w / 2 - (top - h / 2)]
      : [x + PIP_LEFT, x + PIP_LEFT + PIP];
    out.push({ pip, room: [-ROW_PAD_LEFT, Infinity] });
  });
  out.forEach((o, i) => {
    if (i + 1 < visualLeft.length) o.room[1] = visualLeft[i + 1];
  });
  return out;
}

// Self lands (--card-w-sm / --card-h-sm) and the two opponent sizes.
const STRIP_SIZES: [number, number][] = [
  [88, 123],
  [64, 90],
  [48, 67],
];
const ROW_SIZES: [number, number][] = [
  [120, 168],
  [88, 123],
  [64, 90],
];
const LAYOUTS = [
  [true, true, true],
  [false, false, true, true],
  [true, true, true, true, true, true, true],
  [false, false, false],
];

describe("failed-art pip on the battlefield", () => {
  for (const strip of [true, false]) {
    for (const [w, h] of strip ? STRIP_SIZES : ROW_SIZES) {
      for (const layout of LAYOUTS) {
        const tiles = layout.map((t) => (t ? "T" : "u")).join("");
        it(`is uncovered and unclipped: ${strip ? "strip" : "row"} ${w}x${h} ${tiles}`, () => {
          pipsInRow(layout, w, h, strip).forEach(({ pip, room }, i) => {
            expect(pip[0], `tile ${i} pip clipped`).toBeGreaterThanOrEqual(room[0]);
            expect(pip[1], `tile ${i} pip covered by the next tile`).toBeLessThanOrEqual(room[1]);
          });
        });
      }
    }
  }
});
