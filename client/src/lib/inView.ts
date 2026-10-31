// inView.ts — a Svelte action that calls `onenter` once, the first time
// its node comes near the view of its scrolling ancestor (the nearest
// `[data-scroll-root]`, else the viewport), then lets go.
//
// The deck picker's grid holds up to 200 saved decks (ADR 0212
// consequences: "a library of 200 decks loads only what scrolls into
// view"). An <img loading="lazy"> already waits for the image, but the
// artist's credit is a GET /cards/{id} per crop, so a tile mounts its
// art only once it is about to be seen. Where IntersectionObserver is
// missing (jsdom), everything counts as seen at once.

export interface InViewOptions {
  onenter: () => void;
  /** How far outside the view still counts, as a CSS margin. */
  margin?: string;
}

export function inView(node: HTMLElement, opts: InViewOptions) {
  let current = opts;
  if (typeof IntersectionObserver === "undefined") {
    current.onenter();
    return {};
  }
  const root = node.parentElement?.closest<HTMLElement>("[data-scroll-root]") ?? null;
  const io = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) {
        io.disconnect();
        current.onenter();
      }
    },
    { root, rootMargin: current.margin ?? "160px" },
  );
  io.observe(node);
  return {
    update(next: InViewOptions) {
      current = next;
    },
    destroy() {
      io.disconnect();
    },
  };
}
