// rowFit keeps a battlefield row on one line (#2336). A row whose cards
// are wider than the row first shrinks them (#2438), down to a size at
// which the name strip, power/toughness and counters still read, and
// only then overlaps them, each by the same amount, just enough to fit,
// the way the hand fan does, instead of wrapping onto a second line the
// panel has no room for and scrolling inside it.
//
// The overlap is capped: past it a card would show too little to read
// or click, so a board that big still scrolls, sideways.
//
// #3044: the exception is a row that would shrink a long way (below
// WRAP_BELOW of its size) while the panel holds the height for more
// lines at a larger card. That row wraps (fitWrap), within the height
// the panel shares out between its rows (allocateRowHeights).

/** The most of a card that may sit under its neighbour. */
export const MAX_OVERLAP = 0.78;

/**
 * The smallest card height, in px, a row shrinks its cards to (#2438):
 * the art tile's name strip, power/toughness and counter badges still
 * read at 56×40, and hover still zooms. A row whose cards are already
 * this small does not shrink at all.
 */
export const MIN_CARD_H = 56;

/**
 * fitScale is the share of their own size a row's cards are drawn at so
 * the row fits `available`: 1 when it fits already, never below
 * `minScale`. `natural` is the width of everything that scales with the
 * card (the cards, their attachments, the strip's overlaps) at full
 * size, and `fixed` what does not (the gaps between piles).
 */
export function fitScale(
  natural: number,
  fixed: number,
  available: number,
  minScale: number,
): number {
  if (natural <= 0 || available <= 0) return 1;
  const s = (available - fixed) / natural;
  return Math.max(Math.min(1, minScale), Math.min(1, s));
}

/**
 * fitOverlap is how many px each pile after the first moves left so
 * the row fits `available`, or 0 when it fits already. `widths` are the
 * piles' own widths, `gap` the flex gap between them, and `cardW` the
 * width of one card, which the cap is a share of.
 */
export function fitOverlap(
  widths: readonly number[],
  gap: number,
  available: number,
  cardW: number,
): number {
  const n = widths.length;
  if (n < 2 || available <= 0) return 0;
  const natural = widths.reduce((sum, w) => sum + w, 0) + gap * (n - 1);
  if (natural <= available) return 0;
  const need = (natural - available) / (n - 1);
  return Math.min(need, cardW * MAX_OVERLAP + gap);
}

/**
 * stripOpen is how much of the land strip's designed overlap it still
 * needs, from 0 (the piles sit edge to edge) to 1 (the full overlap
 * that leaves 40% of each card showing). `spread` is the strip's
 * content width with no overlap, `packed` its width with the full
 * overlap, `available` the room it has. The strip uses the room it has
 * before it overlaps at all (#2960), and overlaps only as much as the
 * rest requires. Widths are linear in the overlap, so this is exact.
 */
export function stripOpen(spread: number, packed: number, available: number): number {
  if (available <= 0 || spread <= available) return 0;
  const room = spread - packed;
  if (room <= 0.5) return 1;
  return Math.min(1, Math.max(0, (spread - available) / room));
}

/**
 * tappedRoom is the extra width, in px, a tapped card needs on EACH side
 * of its box (#2442). A card turns 90 degrees about its centre, so its
 * box keeps the upright width `cardW` while the painted card is `cardH`
 * wide: it overhangs by (cardH - cardW) / 2 per side. BattlefieldRow
 * gives a tapped pile that much padding (the same expression in CSS), so
 * the row's measured widths, fitScale and fitOverlap already include it.
 * A card wider than tall has no overhang.
 */
export function tappedRoom(cardW: number, cardH: number): number {
  return Math.max(0, (cardH - cardW) / 2);
}

/**
 * overrideStale is whether a shrunk row's px card size was worked out
 * for a different row size than `baseH`, the one it has now (#3028): a
 * resize moved it. `appliedH` is the size the override was applied to,
 * 0 with no override.
 */
export function overrideStale(scale: number, appliedH: number, baseH: number): boolean {
  return scale < 1 && appliedH > 0 && baseH > 0 && Math.abs(appliedH - baseH) > 0.5;
}

// --- wrapping into the panel's height (#3044) --------------------------
//
// A row that would shrink its cards a long way to stay on one line may
// wrap onto more lines instead, when the height the panel gives it
// (its budget) holds the lines at a larger card than one line can.
// The budget is the height the row takes on one line as drawn, plus a
// share of whatever the panel has spare (allocateRowHeights), so a
// wrapped row never takes height another row was going to use.

/**
 * A row only considers wrapping once one line would draw its cards
 * below this share of their own size. Above it a single line, which
 * reads a board at a glance, is worth the slight shrink.
 */
export const WRAP_BELOW = 0.75;

/** A wrapped row's cards must come out at least this much larger. */
export const WRAP_GAIN = 1.1;

/** The most lines a row wraps onto. */
export const MAX_LINES = 4;

/**
 * lineCount is how many lines a wrapping flex row lays `widths` out
 * on, the way the browser does: each pile goes on the current line if
 * it fits, else starts the next one.
 */
export function lineCount(widths: readonly number[], gap: number, available: number): number {
  let lines = 0;
  let x = 0;
  for (const w of widths) {
    if (lines > 0 && x + gap + w <= available) {
      x += gap + w;
      continue;
    }
    lines++;
    x = w;
  }
  return lines;
}

/**
 * scaleForLines is the largest share (at most 1) of their own widths
 * at which `widths` fit on `lines` lines of `available`. Shrinking
 * every pile never adds a line, so a bisection finds it.
 */
export function scaleForLines(
  widths: readonly number[],
  gap: number,
  available: number,
  lines: number,
): number {
  if (widths.length === 0 || lineCount(widths, gap, available) <= lines) return 1;
  let lo = 0;
  let hi = 1;
  for (let i = 0; i < 30; i++) {
    const mid = (lo + hi) / 2;
    const scaled = widths.map((w) => w * mid);
    if (lineCount(scaled, gap, available) <= lines) lo = mid;
    else hi = mid;
  }
  return lo;
}

/**
 * What a row's piles measure at full size, for fitWrap and rowDemand.
 * `widths` are the piles' own widths and `lineH` the tallest pile;
 * `gap` is the flex gap, both across and between lines; `pad` is the
 * row's own vertical padding (the label sits in it).
 */
export interface RowShape {
  widths: readonly number[];
  gap: number;
  available: number;
  lineH: number;
  pad: number;
}

/** The share one line draws the cards at, before any floor. */
function oneLineScale(r: RowShape): number {
  const natural = r.widths.reduce((sum, w) => sum + w, 0);
  if (natural <= 0 || r.available <= 0) return 1;
  return Math.min(1, (r.available - r.gap * Math.max(0, r.widths.length - 1)) / natural);
}

/** The height of `lines` lines of cards drawn at `scale`. */
function linesHeight(r: RowShape, lines: number, scale: number): number {
  return r.pad + lines * r.lineH * scale + (lines - 1) * r.gap;
}

/**
 * rowDemand is what a row asks of the panel's height: `base`, the
 * height of its one line as drawn (shrunk to fit its width, never
 * under `minScale`), and `want`, as many lines at full size as its
 * piles need (up to MAX_LINES), or `base` when one line already draws
 * them at WRAP_BELOW or more. A row shrunk on one line leaves the
 * height it does not use to the panel's spare, for a row that wraps.
 */
export function rowDemand(r: RowShape, minScale = 0): { base: number; want: number } {
  const one = oneLineScale(r);
  const drawn = Math.max(one, Math.min(1, minScale));
  const base = r.lineH > 0 ? r.pad + r.lineH * drawn : r.pad;
  if (r.lineH <= 0 || r.widths.length < 2 || one >= WRAP_BELOW) {
    return { base, want: base };
  }
  const lines = Math.min(MAX_LINES, lineCount(r.widths, r.gap, r.available));
  return { base, want: Math.max(base, linesHeight(r, lines, 1)) };
}

/**
 * fitWrap is the lines and card share a row draws at when wrapping
 * draws its cards larger than one line does, within `budget` px of
 * height, or null to stay on one line. `minScale` is the floor a
 * single line shrinks to (MIN_CARD_H) before it overlaps; a wrapped
 * row never goes under it, and must beat what one line needs (its
 * cards' share before the floor and the overlap) by WRAP_GAIN.
 */
export function fitWrap(
  r: RowShape,
  budget: number,
  minScale: number,
): { lines: number; scale: number } | null {
  if (budget <= 0 || r.lineH <= 0 || r.available <= 0 || r.widths.length < 2) return null;
  const one = oneLineScale(r);
  if (one >= WRAP_BELOW) return null;

  let best: { lines: number; scale: number } | null = null;
  for (let k = 2; k <= MAX_LINES; k++) {
    const byWidth = scaleForLines(r.widths, r.gap, r.available, k);
    const byHeight = (budget - r.pad - (k - 1) * r.gap) / (k * r.lineH);
    const s = Math.min(1, byWidth, byHeight);
    if (s < minScale || s < one * WRAP_GAIN) continue;
    if (!best || s > best.scale + 1e-3) best = { lines: k, scale: s };
  }
  if (!best) return null;
  const lines = lineCount(
    r.widths.map((w) => w * best.scale),
    r.gap,
    r.available,
  );
  return { lines, scale: best.scale };
}

/**
 * allocateRowHeights shares `total` px of panel height between rows
 * stacked one above the other, from each one's rowDemand. Every row
 * gets its `base`; what is left over goes to the rows that want more,
 * in proportion to how much more, never past what each wants. A panel
 * with nothing left over gives each row its base, so a row wraps only
 * inside the height its one line takes.
 */
export function allocateRowHeights(
  total: number,
  demands: readonly { base: number; want: number }[],
): number[] {
  const bases = demands.map((d) => d.base);
  const spare = total - bases.reduce((sum, b) => sum + b, 0);
  const extra = demands.map((d) => Math.max(0, d.want - d.base));
  const asked = extra.reduce((sum, e) => sum + e, 0);
  if (spare <= 0 || asked <= 0) return bases;
  const share = Math.min(1, spare / asked);
  return demands.map((d, i) => d.base + extra[i] * share);
}

/**
 * lineHeight is the full-size height of a row's one line: its tallest
 * pile, measured at `scale`. #3893: an empty row that holds its line
 * (the creature row, BattlefieldRow `hold`) asks for one card of its
 * own height, `cardH`, which is what its CSS keeps open; one label high
 * instead, it left that height to the back row's budget, and the back
 * row moved up into the creatures' place.
 */
export function lineHeight(
  pileHeights: readonly number[],
  scale: number,
  cardH: number,
  hold = false,
): number {
  if (pileHeights.length === 0) return hold ? Math.max(0, cardH) : 0;
  return Math.max(0, ...pileHeights) / (scale || 1);
}

/**
 * roomOf is the width a row's card list has with its cards at full
 * size and no overlap. A back-row zone is as wide as its cards (up to
 * its share, #3028), so once shrunk and overlapped it measured only the
 * room its overlapped cards took, and kept them that way: an artifact
 * row crushed into a third of the width it was allowed (#3044). The
 * override is lifted for the reading and put back in the same task, so
 * nothing is painted or observed in between.
 */
function roomOf(node: HTMLElement): number {
  const w = node.style.getPropertyValue("--card-w");
  const h = node.style.getPropertyValue("--card-h");
  const o = node.style.getPropertyValue("--fit-overlap");
  if (!w && !h && (!o || parseFloat(o) === 0)) return node.clientWidth;
  node.style.removeProperty("--card-w");
  node.style.removeProperty("--card-h");
  node.style.setProperty("--fit-overlap", "0px");
  const room = node.clientWidth;
  if (w) node.style.setProperty("--card-w", w);
  if (h) node.style.setProperty("--card-h", h);
  if (o) node.style.setProperty("--fit-overlap", o);
  else node.style.removeProperty("--fit-overlap");
  return room;
}

/** The options BattlefieldRow passes fitRow. */
export interface FitRowOptions {
  /** The land strip: piles overlap by design and never wrap. */
  strip?: boolean;
  /** #4087: never overlap the piles (planeswalkers): shrink, wrap, then scroll. */
  spread?: boolean;
  /** #3044: the height the panel gives this row, 0 when unknown. */
  budget?: number;
  /** #3893: an empty row keeps one card's line (lineHeight). */
  hold?: boolean;
  /** #3044: called with the row's demand whenever it moves. */
  onDemand?: (d: RowReport) => void;
}

/**
 * What fitRow reports to the panel: the row's rowDemand, and `used`,
 * the height of its lines while it is wrapped (0 on one line). A grid
 * track sized `auto` beside a `1fr` one is not promised its content's
 * height, so the panel holds the creature track open to `used`.
 */
export interface RowReport {
  base: number;
  want: number;
  used: number;
}

/**
 * fitRow is the action on a row's card list. Whenever the row, a pile,
 * the set of piles or the row's own card size changes, it shrinks the
 * cards to fit (publishing `--card-w` / `--card-h` on the list) and then
 * publishes the overlap that is still needed as `--fit-overlap`.
 *
 * #3044: before it shrinks a row a long way, it tries wrapping it
 * (fitWrap) within the height the panel gives it (`budget`); a row that
 * wraps carries `data-wrap` and no overlap. It reports the row's
 * rowDemand through `onDemand`, from full-size measures that do not
 * depend on how the row is drawn, so the panel's budgets settle.
 *
 * The row's own card size is read from a `.fit-probe-card` beside the list,
 * which is sized by the inherited `--card-w` / `--card-h` and so never
 * sees the list's override. With `strip`, the piles overlap each other
 * by design (the land strip), so there is no `--fit-overlap`; it
 * publishes `--strip-k` instead, the share of that overlap in use
 * (stripOpen), which is 0 while the strip has room to spare. The strip
 * never wraps.
 */
export function fitRow(
  node: HTMLElement,
  options: FitRowOptions = {},
): { update(o: FitRowOptions): void; destroy(): void } {
  // No layout to measure (jsdom): the row keeps its natural spacing.
  if (typeof ResizeObserver === "undefined") return { update() {}, destroy() {} };
  let opts = options;
  let frame = 0;
  let scale = 1;
  // The row's own card height `scale` was applied to, while it is
  // below 1. The override is in px, so it does not follow that size
  // when the panel resizes (#3028).
  let appliedH = 0;
  let reported: RowReport = { base: -1, want: -1, used: -1 };
  const probe =
    node.parentElement?.querySelector<HTMLElement>(":scope > .fit-probe > .fit-probe-card") ?? null;
  const measure = (): void => {
    frame = 0;
    const strip = !!opts.strip;
    const piles = [...node.children].filter((c): c is HTMLElement => c instanceof HTMLElement);
    const style = getComputedStyle(node);
    const gap = parseFloat(style.columnGap) || 0;
    const available = strip ? node.clientWidth : roomOf(node);
    // The land strip (#2960): the content width with the overlap fully
    // off (spread) and fully on (packed), read straight from layout.
    // The piles' own widths do not depend on it, only their margins do.
    let spread = 0;
    let packed = 0;
    if (strip && piles.length > 0) {
      const first = piles[0];
      const last = piles[piles.length - 1];
      const padR = parseFloat(style.paddingRight) || 0;
      const content = (): number => last.offsetLeft + last.offsetWidth - first.offsetLeft + padR;
      node.style.setProperty("--strip-k", "0");
      spread = content();
      node.style.setProperty("--strip-k", "1");
      packed = content();
    }
    const baseW = probe?.offsetWidth ?? 0;
    const baseH = probe?.offsetHeight ?? 0;
    // #3028: the row's own size moved under a px override (the window
    // was resized). The cards still have the old size, and a row that
    // is only as wide as its cards cannot tell from them how much room
    // it has now. Drop the override and measure again next frame, from
    // the full size, as a freshly drawn row does.
    if (overrideStale(scale, appliedH, baseH)) {
      scale = 1;
      appliedH = 0;
      node.style.removeProperty("--card-w");
      node.style.removeProperty("--card-h");
      schedule();
      return;
    }
    // What the row's cards measure now, at `scale`, and at full size.
    const widths = piles.map((p) => p.offsetWidth);
    const lineH = lineHeight(
      piles.map((p) => p.offsetHeight),
      scale,
      baseH,
      !!opts.hold,
    );
    const rowStyle = node.parentElement ? getComputedStyle(node.parentElement) : null;
    const pad = rowStyle
      ? (parseFloat(rowStyle.paddingTop) || 0) + (parseFloat(rowStyle.paddingBottom) || 0)
      : 0;
    const shape: RowShape = {
      // Fractional widths: offsetWidth rounds, and twenty rounded piles
      // can be 10px off, a line's worth of difference at the break.
      widths: piles.map((p) => p.getBoundingClientRect().width / scale),
      gap,
      // A few px short of the measured width, so the browser's own line
      // breaking never needs a line fitWrap did not.
      available: Math.max(0, available - 3),
      lineH,
      pad,
    };
    const fixed = strip ? 0 : gap * Math.max(0, piles.length - 1);
    const now = strip ? packed : widths.reduce((sum, w) => sum + w, 0);
    const natural = now / scale;
    const minScale = baseH > 0 ? MIN_CARD_H / baseH : 1;
    const wrap =
      !strip && baseW > 0 && baseH > 0 ? fitWrap(shape, opts.budget ?? 0, minScale) : null;
    node.toggleAttribute("data-wrap", !!wrap);
    const next = wrap
      ? wrap.scale
      : baseW > 0 && baseH > 0
        ? fitScale(natural, fixed, available, minScale)
        : 1;
    // The strip never wraps: it asks for its one line, as drawn.
    const stripH = pad + lineH * next;
    const demand: RowReport = {
      ...(strip ? { base: stripH, want: stripH } : rowDemand(shape, minScale)),
      used: wrap
        ? Math.ceil(pad + wrap.lines * lineH * wrap.scale + (wrap.lines - 1) * gap) + 1
        : 0,
    };
    if (
      Math.abs(demand.base - reported.base) > 0.5 ||
      Math.abs(demand.want - reported.want) > 0.5 ||
      Math.abs(demand.used - reported.used) > 0.5
    ) {
      reported = demand;
      opts.onDemand?.(demand);
    }
    const was = scale;
    if (Math.abs(next - scale) > 0.005) {
      scale = next;
      if (scale < 1) {
        node.style.setProperty("--card-w", `${baseW * scale}px`);
        node.style.setProperty("--card-h", `${baseH * scale}px`);
        appliedH = baseH;
      } else {
        node.style.removeProperty("--card-w");
        node.style.removeProperty("--card-h");
        appliedH = 0;
      }
    }
    // The piles were measured at `was`; the overlap is for the new size.
    const ratio = scale / was;
    const px =
      strip || wrap || opts.spread
        ? 0
        : fitOverlap(
            widths.map((w) => w * ratio),
            gap,
            available,
            (baseW || piles[0]?.querySelector<HTMLElement>(".card")?.offsetWidth || 0) * scale,
          );
    node.style.setProperty("--fit-overlap", `${px}px`);
    if (strip) {
      // Cards resized this frame: the widths above are the old size's,
      // so hold the full overlap until the resize observer re-measures.
      const k = scale !== was || scale < 1 ? 1 : stripOpen(spread, packed, available);
      node.style.setProperty("--strip-k", k.toFixed(3));
      // #4090: the strip takes the room the other zones leave, but no
      // more than its piles need laid edge to edge, so the back row's
      // zones sit together instead of the lands sitting at one end of a
      // wide empty strip. Only at full card size: a shrunk row has
      // already used all its room, and its spread width is the shrunk
      // one.
      const row = node.parentElement;
      if (row) {
        if (scale >= 1 && was >= 1 && spread > 0) {
          const rowPad =
            (parseFloat(rowStyle?.paddingLeft ?? "0") || 0) +
            (parseFloat(rowStyle?.paddingRight ?? "0") || 0);
          row.style.setProperty("--strip-spread", `${Math.ceil(spread + rowPad)}px`);
        } else {
          row.style.removeProperty("--strip-spread");
        }
      }
    }
  };
  const schedule = (): void => {
    if (!frame) frame = requestAnimationFrame(measure);
  };
  const resize = new ResizeObserver(schedule);
  const observeAll = (): void => {
    resize.disconnect();
    resize.observe(node);
    if (probe) resize.observe(probe);
    for (const c of node.children) resize.observe(c);
    schedule();
  };
  const mutate = new MutationObserver(observeAll);
  mutate.observe(node, { childList: true });
  observeAll();
  return {
    update(o: FitRowOptions) {
      // A new options object comes with every render of the row; only a
      // new strip or hold flag or budget needs the row measured again.
      const moved =
        !!o.strip !== !!opts.strip ||
        !!o.hold !== !!opts.hold ||
        Math.abs((o.budget ?? 0) - (opts.budget ?? 0)) > 0.5;
      opts = o;
      if (moved) schedule();
    },
    destroy() {
      if (frame) cancelAnimationFrame(frame);
      resize.disconnect();
      mutate.disconnect();
    },
  };
}
