// rowFit keeps a battlefield row on one line (#2336). A row whose cards
// are wider than the row overlaps them, each by the same amount, just
// enough to fit, the way the hand fan does, instead of wrapping onto a
// second line the panel has no room for and scrolling inside it.
//
// The overlap is capped: past it a card would show too little to read
// or click, so a board that big still scrolls, sideways.

/** The most of a card that may sit under its neighbour. */
export const MAX_OVERLAP = 0.78;

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
 * fitRow is the action: it measures the row's piles whenever the row,
 * a pile, or the set of piles changes size, and publishes the overlap
 * as `--fit-overlap` on the row's card list.
 */
export function fitRow(node: HTMLElement): { destroy(): void } {
  // No layout to measure (jsdom): the row keeps its natural spacing.
  if (typeof ResizeObserver === "undefined") return { destroy() {} };
  let frame = 0;
  const measure = (): void => {
    frame = 0;
    const piles = [...node.children].filter((c): c is HTMLElement => c instanceof HTMLElement);
    const style = getComputedStyle(node);
    const gap = parseFloat(style.columnGap) || 0;
    const first = piles[0]?.querySelector<HTMLElement>(".card");
    const cardW = first?.offsetWidth ?? 0;
    const px = fitOverlap(
      piles.map((p) => p.offsetWidth),
      gap,
      node.clientWidth,
      cardW,
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
    for (const c of node.children) resize.observe(c);
    schedule();
  };
  const mutate = new MutationObserver(observeAll);
  mutate.observe(node, { childList: true });
  observeAll();
  return {
    destroy() {
      if (frame) cancelAnimationFrame(frame);
      resize.disconnect();
      mutate.disconnect();
    },
  };
}
