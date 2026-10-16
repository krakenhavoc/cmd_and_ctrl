// sizeSink is Game.svelte's onDockSize in miniature: a size in $state,
// written only when it changes, so the write reads the stored value
// first. A component that calls it from a tracked context depends on
// that state, which is what dockSizeObserver.render.test.ts pins.

export interface SizeSink {
  report(w: number, h: number): void;
  readonly size: { w: number; h: number };
}

export function sizeSink(): SizeSink {
  let size = $state({ w: 0, h: 0 });
  return {
    report(w: number, h: number): void {
      if (w !== size.w || h !== size.h) size = { w, h };
    },
    get size() {
      return size;
    },
  };
}
