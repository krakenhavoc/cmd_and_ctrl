// rowFit keeps a battlefield row on one line (#2336). A row whose cards
// are wider than the row first shrinks them (#2438), down to a size at
// which the name strip, power/toughness and counters still read, and
// only then overlaps them, each by the same amount, just enough to fit,
// the way the hand fan does, instead of wrapping onto a second line the
// panel has no room for and scrolling inside it.
//
// The overlap is capped: past it a card would show too little to read
// or click, so a board that big still scrolls, sideways.

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
 * fitRow is the action on a row's card list. Whenever the row, a pile,
 * the set of piles or the row's own card size changes, it shrinks the
 * cards to fit (publishing `--card-w` / `--card-h` on the list) and then
 * publishes the overlap that is still needed as `--fit-overlap`.
 *
 * The row's own card size is read from a `.fit-probe-card` beside the list,
 * which is sized by the inherited `--card-w` / `--card-h` and so never
 * sees the list's override. With `strip`, the piles overlap each other
 * by design (the land strip), so the row's natural width is its
 * scrollWidth and there is no `--fit-overlap`.
 */
export function fitRow(
  node: HTMLElement,
  strip = false,
): { update(s: boolean): void; destroy(): void } {
  // No layout to measure (jsdom): the row keeps its natural spacing.
  if (typeof ResizeObserver === "undefined") return { update() {}, destroy() {} };
  let frame = 0;
  let scale = 1;
  const probe =
    node.parentElement?.querySelector<HTMLElement>(":scope > .fit-probe > .fit-probe-card") ?? null;
  const measure = (): void => {
    frame = 0;
    const piles = [...node.children].filter((c): c is HTMLElement => c instanceof HTMLElement);
    const style = getComputedStyle(node);
    const gap = parseFloat(style.columnGap) || 0;
    const available = node.clientWidth;
    const baseW = probe?.offsetWidth ?? 0;
    const baseH = probe?.offsetHeight ?? 0;
    // What the row's cards measure now, at `scale`, and at full size.
    const widths = piles.map((p) => p.offsetWidth);
    const fixed = strip ? 0 : gap * Math.max(0, piles.length - 1);
    const now = strip ? node.scrollWidth : widths.reduce((sum, w) => sum + w, 0);
    const natural = now / scale;
    const minScale = baseH > 0 ? MIN_CARD_H / baseH : 1;
    const next = baseW > 0 && baseH > 0 ? fitScale(natural, fixed, available, minScale) : 1;
    const was = scale;
    if (Math.abs(next - scale) > 0.005) {
      scale = next;
      if (scale < 1) {
        node.style.setProperty("--card-w", `${baseW * scale}px`);
        node.style.setProperty("--card-h", `${baseH * scale}px`);
      } else {
        node.style.removeProperty("--card-w");
        node.style.removeProperty("--card-h");
      }
    }
    // The piles were measured at `was`; the overlap is for the new size.
    const ratio = scale / was;
    const px = strip
      ? 0
      : fitOverlap(
          widths.map((w) => w * ratio),
          gap,
          available,
          (baseW || piles[0]?.querySelector<HTMLElement>(".card")?.offsetWidth || 0) * scale,
        );
    node.style.setProperty("--fit-overlap", `${px}px`);
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
    update(s: boolean) {
      strip = s;
      schedule();
    },
    destroy() {
      if (frame) cancelAnimationFrame(frame);
      resize.disconnect();
      mutate.disconnect();
    },
  };
}
