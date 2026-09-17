// Is the board card's failed-art pip (#33) visible on a tapped card in
// a battlefield row, on a land in a pile, and on an attachment drawn
// behind its host?
//
// A tapped Card turns 90° clockwise, and the turn makes the tile its
// own stacking context, so the pip's z-index counts only inside it: a
// later tile that overlaps the pip paints over it. The land strip
// overlaps tapped piles by 35% of a width and stacks same-name lands
// 4px apart, a tapped tile in an ordinary row overhangs its
// neighbours, and a host covers all but a sliver of the Auras and
// Equipment tucked behind it. Where the pip sits is CSS in Card.svelte
// and BattlefieldRow.svelte, the overlaps are CSS in
// BattlefieldRow.svelte, and the client has no DOM to lay either out
// in — so this reads the numbers out of those files and does the
// geometry. It was measured against Chromium's hit-testing
// (elementFromPoint at the pip) and agrees: 22px, the old tapped
// position, is covered on every tapped land but the rightmost, and the
// old attachment position, 22px down the left edge, under a tapped
// host.
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
// a calc() over percentages, --card-w and --card-h.
function px(value: string, w: number, h: number, percentOf = 0): number {
  const expr = value
    .replace(/([\d.]+)%/g, (_, n: string) => `(${n} / 100 * ${percentOf})`)
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

// --- the strip ----------------------------------------------------------
//
// In the strip every land is inside a .pile: one pile per name among
// the untapped lands, one pile per tapped land. The between-pile
// margin-left rules, in source order, with a matcher and specificity
// each, so the cascade is resolved the way a browser does.
interface Item {
  tapped: boolean;
  first: boolean;
  prevTapped: boolean;
}
const STRIP_PILE = ".row.strip .pile";
const stripRules = [...rowSvelte.matchAll(/\n {2}(\.row\.strip \.pile[^{]*?) \{([^}]*)\}/g)]
  .filter(([, selector, body]) => /margin-left:/.test(body) && !/\[role=/.test(selector))
  .map(([, selector, body]) => {
    const tail = selector.slice(STRIP_PILE.length);
    const matches = (it: Item): boolean => {
      if (tail.startsWith(":not(.tapped) + .pile")) {
        const own = tail.slice(":not(.tapped) + .pile".length);
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

// Inside a pile: the rest-state step between copies, and the step
// once the pile is hovered and spread.
const WITHIN = '.row.strip .pile > [role="listitem"] + [role="listitem"]';
const WITHIN_HOVER = '.row.strip .pile.multi:hover > [role="listitem"] + [role="listitem"]';
const withinRest = (w: number, h: number) =>
  w + px(decl(block(rowSvelte, WITHIN), "margin-left"), w, h);
const withinHover = (w: number, h: number) =>
  w + px(decl(block(rowSvelte, WITHIN_HOVER), "margin-left"), w, h);

// A pile of `n` same-name lands (n is 1 for a tapped land), and the
// pip of each tile in it with the room it has: right of the row's
// clip edge and left of whatever paints over it next. At rest the top
// copy (the last in the DOM) is what must be readable; every copy
// under it is checked spread out, which is what hovering the pile
// does and the only state in which those copies can be clicked.
interface Pile {
  tapped: boolean;
  n: number;
}
type Check = { pip: [number, number]; room: [number, number]; state: "rest" | "spread" };

function pipsInStrip(piles: Pile[], w: number, h: number): Check[] {
  const out: Check[] = [];
  let x = 0;
  const top = px(tappedTop, w, h, h);
  const pipOf = (left: number, tapped: boolean): [number, number] =>
    tapped
      ? [left + w / 2 - (top + PIP - h / 2), left + w / 2 - (top - h / 2)]
      : [left + PIP_LEFT, left + PIP_LEFT + PIP];
  const visualLeft = (left: number, tapped: boolean) => (tapped ? left + w / 2 - h / 2 : left);
  piles.forEach((p, i) => {
    const margin = stripMargin(
      { tapped: p.tapped, first: i === 0, prevTapped: piles[i - 1]?.tapped ?? false },
      w,
      h,
    );
    const restWidth = w + withinRest(w, h) * (p.n - 1);
    x += margin;
    const pileLeft = x;
    // The top copy at rest, with the next pile's left edge as its room.
    const next = piles[i + 1];
    const nextMargin = next
      ? stripMargin({ tapped: next.tapped, first: false, prevTapped: p.tapped }, w, h)
      : Infinity;
    const nextLeft = next ? visualLeft(pileLeft + restWidth + nextMargin, next.tapped) : Infinity;
    const topLeft = pileLeft + withinRest(w, h) * (p.n - 1);
    out.push({ pip: pipOf(topLeft, p.tapped), room: [-ROW_PAD_LEFT, nextLeft], state: "rest" });
    // Every copy under it, spread: each one's room ends at the copy
    // after it.
    for (let k = 0; k + 1 < p.n; k++) {
      const left = pileLeft + withinHover(w, h) * k;
      out.push({
        pip: pipOf(left, p.tapped),
        room: [-ROW_PAD_LEFT, left + withinHover(w, h)],
        state: "spread",
      });
    }
    x = pileLeft + restWidth;
  });
  return out;
}

// pipsInRow lays out an ordinary (non-strip) row of tiles, 8px apart,
// and returns each pip's extent and room.
function pipsInRow(tapped: boolean[], w: number, h: number): Check[] {
  const out: Check[] = [];
  const visualLeft: number[] = [];
  let x = 0;
  const top = px(tappedTop, w, h, h);
  tapped.forEach((t, i) => {
    x += i === 0 ? 0 : w + 8;
    visualLeft.push(t ? x + w / 2 - h / 2 : x);
    const pip: [number, number] = t
      ? [x + w / 2 - (top + PIP - h / 2), x + w / 2 - (top - h / 2)]
      : [x + PIP_LEFT, x + PIP_LEFT + PIP];
    out.push({ pip, room: [-ROW_PAD_LEFT, Infinity], state: "rest" });
  });
  out.forEach((o, i) => {
    if (i + 1 < visualLeft.length) o.room[1] = visualLeft[i + 1];
  });
  return out;
}

// The land sizes: the floors of the self, opponent and across-table
// ramps (--card-h-sm), and the ceiling of the self ramp. The ramps
// scale between them and the fixed parts of the geometry — the pip,
// the 4px pile step — only get more room as the tiles grow.
const STRIP_SIZES: [number, number][] = [
  [88, 123],
  [64, 90],
  [48, 67],
  [111, 156],
];
const ROW_SIZES: [number, number][] = [
  [120, 168],
  [88, 123],
  [64, 90],
  [150, 210],
];
const ROW_LAYOUTS = [
  [true, true, true],
  [false, false, true, true],
  [true, true, true, true, true, true, true],
  [false, false, false],
];
const u = (n: number): Pile => ({ tapped: false, n });
const T = (): Pile => ({ tapped: true, n: 1 });
const STRIP_LAYOUTS: Pile[][] = [
  [T(), T(), T()],
  [u(1), u(1), T(), T()],
  [T(), T(), T(), T(), T(), T(), T()],
  [u(1), u(1), u(1)],
  [u(4)],
  [u(4), u(2), T(), T()],
  [u(3), u(1), u(2)],
  [u(2), T()],
];

function assertVisible(checks: Check[]) {
  checks.forEach(({ pip, room, state }, i) => {
    expect(pip[0], `tile ${i} (${state}) pip clipped`).toBeGreaterThanOrEqual(room[0]);
    expect(pip[1], `tile ${i} (${state}) pip covered by the next tile`).toBeLessThanOrEqual(
      room[1],
    );
  });
}

describe("failed-art pip on the battlefield", () => {
  for (const [w, h] of ROW_SIZES) {
    for (const layout of ROW_LAYOUTS) {
      const tiles = layout.map((t) => (t ? "T" : "u")).join("");
      it(`is uncovered and unclipped: row ${w}x${h} ${tiles}`, () => {
        assertVisible(pipsInRow(layout, w, h));
      });
    }
  }
  for (const [w, h] of STRIP_SIZES) {
    for (const layout of STRIP_LAYOUTS) {
      const tiles = layout.map((p) => (p.tapped ? "T" : `u${p.n}`)).join(" ");
      it(`is uncovered and unclipped: strip ${w}x${h} ${tiles}`, () => {
        assertVisible(pipsInStrip(layout, w, h));
      });
    }
  }
});

// --- attachments ------------------------------------------------------
//
// BattlefieldRow draws a host's Auras and Equipment first inside its
// .host-stack, each pulled under the next by a negative margin and
// dropped a few pixels, and the host last, on top. What stays visible
// of an attachment depends on what is tapped: a sliver down its left
// edge while its neighbour is upright, a band along its bottom once
// the host turns, and the left end of its own turned tile when the
// attachment itself is tapped. The pip has to sit where all of those
// agree, so this lays the stack out in two dimensions.

const BORDER = px(decl(block(cardSvelte, ".card"), "border").split(/\s+/)[0], 0, 0);
const attachmentRule = block(rowSvelte, ".host-stack .attachment");
const ATTACHMENT_DROP = px(
  /translateY\(([^)]+)\)/.exec(decl(attachmentRule, "transform"))?.[1] ?? "",
  0,
  0,
);

// pipVars resolves --art-error-top / -left for a Card, in cascade
// order: Card.svelte's .card, then .card.tapped, then BattlefieldRow's
// rule for a Card inside an attachment, which is more specific than
// either (it is scoped by two of the row's own classes).
const ATTACHED_CARD = ".host-stack .attachment :global(.card)";
function pipVars(tapped: boolean, attached: boolean): { top: string; left: string } {
  const bodies = [block(cardSvelte, ".card")];
  if (tapped) bodies.push(block(cardSvelte, ".card.tapped"));
  if (attached && rowSvelte.includes(`${ATTACHED_CARD} {`)) {
    bodies.push(block(rowSvelte, ATTACHED_CARD));
  }
  const pick = (prop: string) =>
    bodies.reduce<string | null>(
      (v, b) => (new RegExp(`(?:^|[;\\s])${prop}:`).test(b) ? decl(b, prop) : v),
      null,
    )!;
  return { top: pick("--art-error-top"), left: pick("--art-error-left") };
}

type Rect = { x1: number; y1: number; x2: number; y2: number };

// A tile's box, and its pip, as painted. Turning (dx, dy) from the
// tile's centre 90° clockwise gives (-dy, dx).
function paint(box: Rect, tapped: boolean, r: Rect): Rect {
  if (!tapped)
    return { x1: box.x1 + r.x1, y1: box.y1 + r.y1, x2: box.x1 + r.x2, y2: box.y1 + r.y2 };
  const cx = (box.x1 + box.x2) / 2;
  const cy = (box.y1 + box.y2) / 2;
  const w = box.x2 - box.x1;
  const h = box.y2 - box.y1;
  return {
    x1: cx - (r.y2 - h / 2),
    x2: cx - (r.y1 - h / 2),
    y1: cy + (r.x1 - w / 2),
    y2: cy + (r.x2 - w / 2),
  };
}

// visibleShare samples the pip on a grid and returns the fraction of
// it that no later tile paints over and the row does not clip.
function visibleShare(pip: Rect, later: Rect[], clipLeft: number): number {
  const N = 28;
  let seen = 0;
  for (let i = 0; i < N; i++) {
    for (let j = 0; j < N; j++) {
      const x = pip.x1 + ((i + 0.5) / N) * (pip.x2 - pip.x1);
      const y = pip.y1 + ((j + 0.5) / N) * (pip.y2 - pip.y1);
      const hidden =
        x < clipLeft || later.some((t) => x > t.x1 && x < t.x2 && y > t.y1 && y < t.y2);
      if (!hidden) seen++;
    }
  }
  return seen / (N * N);
}

// stackPips lays out one .host-stack — attachments, then the host —
// with the stack's left edge at 0, and returns each tile's pip with
// the share of it left visible.
function stackPips(attachments: boolean[], hostTapped: boolean, w: number, h: number) {
  const step = w + px(decl(attachmentRule, "margin-right"), w, h);
  const stackMargin = px(
    decl(block(rowSvelte, ".host-stack.has-attachments"), "margin-left"),
    w,
    h,
  );
  const tiles = [...attachments, hostTapped].map((tapped, i) => {
    const attached = i < attachments.length;
    const y = attached ? ATTACHMENT_DROP : 0;
    const box = { x1: i * step, y1: y, x2: i * step + w, y2: y + h };
    const vars = pipVars(tapped, attached);
    const left = BORDER + px(vars.left, w, h, w - 2 * BORDER);
    const top = BORDER + px(vars.top, w, h, h - 2 * BORDER);
    const pip = paint(box, tapped, { x1: left, y1: top, x2: left + PIP, y2: top + PIP });
    return { attached, tapped, face: paint(box, tapped, { x1: 0, y1: 0, x2: w, y2: h }), pip };
  });
  return tiles.map((t, i) => {
    // The pip is inside its own tile (the tile clips it otherwise)…
    const inside =
      t.pip.x1 >= t.face.x1 &&
      t.pip.x2 <= t.face.x2 &&
      t.pip.y1 >= t.face.y1 &&
      t.pip.y2 <= t.face.y2;
    const later = tiles.slice(i + 1).map((l) => l.face);
    // …and the row's padding is the most room the stack has to its
    // left, for a first land with no strip margin.
    return { ...t, inside, share: visibleShare(t.pip, later, -(stackMargin + ROW_PAD_LEFT)) };
  });
}

// Every board size a host-stack is drawn at: the viewer's creatures and
// lands, and the two opponent panel sizes.
const STACK_SIZES: [number, number][] = [
  [120, 168],
  [88, 123],
  [64, 90],
  [48, 67],
];
// What is attached, and whether each is tapped, then whether the host
// is. A tapped Aura or Equipment is unusual; a tapped host carrying
// one — an enchanted land tapped for mana, an equipped attacker — is
// every other turn. Not covered: an upright attachment in front of a
// tapped one on a tile under about 88px, where the turned tile's
// band reaches below the upright one's bottom edge. The pip is
// modelled as its square box, which covers more than the round pip.
const STACKS: [boolean[], boolean][] = [
  [[false], false],
  [[false], true],
  [[true], false],
  [[true], true],
  [[false, false], false],
  [[false, false], true],
  [[true, true], true],
];

describe("failed-art pip on an attachment", () => {
  for (const [w, h] of STACK_SIZES) {
    for (const [attachments, hostTapped] of STACKS) {
      const name = `${attachments.map((t) => (t ? "T" : "u")).join("")}+${hostTapped ? "T" : "u"}`;
      it(`is visible behind its host: ${w}x${h} ${name}`, () => {
        stackPips(attachments, hostTapped, w, h).forEach((t, i) => {
          expect(t.inside, `tile ${i} pip outside its tile`).toBe(true);
          // A 48px tile leaves 28% of its width, 13.4px, uncovered
          // beside the next one — less than the 14px pip — so there
          // most of the pip, not all of it, is the most there is room
          // for. Everywhere else, all of it.
          const need = w < 60 ? 0.75 : 1;
          expect(t.share, `tile ${i} pip covered`).toBeGreaterThanOrEqual(need);
        });
      });
    }
  }
});
