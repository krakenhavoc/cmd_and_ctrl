// stickColumn.ts — a column that sticks without scrolling on its own
// (#5848 §8): the draft room's picks column and colour-pair panel.
//
// A column shorter than the window sticks just under the site header
// (and the fuse). A taller one scrolls with the page until its end is
// in view, then sticks there, so the page keeps the one scroll bar.
// `position: sticky` is the room's CSS; this sets `--stick`, the `top`
// it uses, from the column's own height.

/** The gap kept under the header, and below a column that sticks by its end, in pixels. */
export const STICK_GAP = 16;

/**
 * The `top` for a sticky column: `offset` (the space the header and fuse
 * take, and a gap) when it fits the window, else the negative value that
 * leaves its end `STICK_GAP` above the window's foot.
 */
export function stickTop(offset: number, height: number, viewport: number): number {
  return Math.min(offset, viewport - height - STICK_GAP);
}

/** The space above a sticky column: the site header, the fuse when there is one, and a gap. */
function offsetAbove(): number {
  const head = document.querySelector<HTMLElement>(".site-header");
  const fuse = document.querySelector<HTMLElement>('[data-testid="draft-fuse"]');
  return (head?.offsetHeight ?? 61) + (fuse?.offsetHeight ?? 0) + STICK_GAP;
}

/** A Svelte action: keeps `--stick` on the node current as it, or the window, resizes. */
export function stickColumn(node: HTMLElement): { destroy: () => void } {
  const apply = (): void => {
    node.style.setProperty(
      "--stick",
      `${stickTop(offsetAbove(), node.offsetHeight, window.innerHeight)}px`,
    );
  };
  apply();
  const ro = typeof ResizeObserver === "undefined" ? null : new ResizeObserver(apply);
  ro?.observe(node);
  window.addEventListener("resize", apply);
  return {
    destroy() {
      ro?.disconnect();
      window.removeEventListener("resize", apply);
    },
  };
}
