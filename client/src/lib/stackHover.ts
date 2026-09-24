// stackHover — the #322 hover preview for the stack, shared by the
// docked StackOverlay and the floating stack lane (#1467).
//
// A stack row draws its own thumb rather than mounting Card.svelte,
// and Card.svelte used to be the only writer of the shared
// `hoveredCard` store, so HoverZoomOverlay never heard about a stack
// row. The fix is to write the store from the row on pointerenter,
// honouring the same settings.display.hoverDelayMs dwell as the table.
// Two surfaces doing that each with their own timer is two chances to
// get the bookkeeping wrong, so it lives here once.
//
// The bookkeeping that matters:
//   - only ever clear OUR write — a battlefield card hovered after us
//     owns the slot and must survive (`clear`);
//   - a row that resolves out from under the cursor never fires
//     pointerleave, so the owner calls `sync` with the live ids and a
//     vanished row drops its preview;
//   - unmount is the other way a row vanishes mid-hover (`destroy`).

import type { Writable } from "svelte/store";
import type { CardView } from "./protocol";

export interface StackHover {
  /** Pointer or focus entered row `itemID`, whose previewable card is `card` (null: nothing to zoom). */
  enter(itemID: string, card: CardView | null, delayMs: number): void;
  /** Pointer or focus left row `itemID`. */
  leave(itemID: string): void;
  /** The rows still on screen; a hovered row not among them drops its preview. */
  sync(liveIDs: Iterable<string>): void;
  /** Cancel any dwell and drop our preview. */
  destroy(): void;
}

export function createStackHover(
  store: Writable<CardView | null>,
  timers: {
    set: (fn: () => void, ms: number) => ReturnType<typeof setTimeout>;
    clear: (t: ReturnType<typeof setTimeout>) => void;
  } = {
    set: (fn, ms) => setTimeout(fn, ms),
    clear: (t) => clearTimeout(t),
  },
): StackHover {
  let timer: ReturnType<typeof setTimeout> | null = null;
  let hoveredItemID: string | null = null;
  let previewedInstanceID: string | null = null;

  function cancel(): void {
    if (timer !== null) {
      timers.clear(timer);
      timer = null;
    }
  }

  function clear(): void {
    const inst = previewedInstanceID;
    hoveredItemID = null;
    previewedInstanceID = null;
    if (!inst) return;
    store.update((c) => (c?.instance_id === inst ? null : c));
  }

  function show(card: CardView): void {
    previewedInstanceID = card.instance_id;
    store.set(card);
  }

  return {
    enter(itemID, card, delayMs) {
      if (!card) return;
      cancel();
      hoveredItemID = itemID;
      if (delayMs <= 0) {
        show(card);
        return;
      }
      timer = timers.set(() => {
        timer = null;
        show(card);
      }, delayMs);
    },
    leave(itemID) {
      cancel();
      if (hoveredItemID !== itemID) return;
      clear();
    },
    sync(liveIDs) {
      if (hoveredItemID === null) return;
      for (const id of liveIDs) if (id === hoveredItemID) return;
      cancel();
      clear();
    },
    destroy() {
      cancel();
      clear();
    },
  };
}
