// stackPlacement.svelte.ts — #3429, ADR 0176: the drag, save and reset of a
// movable stack panel, shared by the docked card (StackOverlay) and the
// floating lane (StackLaneHost) so the two behave the same way.
//
// The caller owns the geometry (the docked card is placed against the
// strip, the lane against the board) and the markup. This owns the state
// and the pointer protocol:
//
//   - `local` is what this tab has put down that the server has not echoed
//     yet (a spot, or null for a reset); undefined means none. A new saved
//     `position` from the server replaces it. A spectator has nothing
//     saved, so theirs stays.
//   - the header captures the pointer, so the release can never click an
//     item under it; a move under DRAG_SLOP is a click, not a drag.
//   - `onSave` / `onReset` are undefined for a spectator: the panel moves
//     for them and nothing is kept.

import type { StackPositionView } from "./protocol";
import { stackFractions } from "./stackPosition";

/** A drag below this many pixels is a click on the header, not a move. */
export const DRAG_SLOP = 3;

export interface StackPlacementOptions {
  /** The seat's saved spot from the server. */
  position: () => StackPositionView | null | undefined;
  onSave: () => ((pos: StackPositionView) => void | Promise<void>) | undefined;
  onReset: () => (() => void | Promise<void>) | undefined;
  /** The panel being dragged, and the board its spot is a fraction of. */
  panel: () => HTMLElement | null | undefined;
  board: () => HTMLElement | null | undefined;
  /** Called as a drag starts, to bring the caller's geometry up to date. */
  before?: () => void;
}

export function createStackPlacement(o: StackPlacementOptions) {
  let local = $state<StackPositionView | null | undefined>(undefined);
  let dragging = $state<StackPositionView | null>(null);

  // Keyed on the values, not the object: every state frame hands in a new
  // one, and only a changed spot is the server's answer.
  const key = $derived.by(() => {
    const p = o.position();
    return p ? `${p.x},${p.y}` : "";
  });
  $effect(() => {
    // Read for its dependency; a bare `void key` can be dropped by the
    // TypeScript transform before Svelte sees it in a .svelte.ts module.
    if (key.length >= 0) local = undefined;
  });

  const placed = $derived(dragging ?? (local !== undefined ? local : (o.position() ?? null)));

  let drag: { id: number; dx: number; dy: number; x0: number; y0: number; moved: boolean } | null =
    null;

  function spotAt(e: PointerEvent): StackPositionView | null {
    const panel = o.panel();
    const board = o.board();
    if (!drag || !panel || !board) return null;
    const b = board.getBoundingClientRect();
    return stackFractions(
      e.clientX - drag.dx - b.left,
      e.clientY - drag.dy - b.top,
      { width: b.width, height: b.height },
      { width: panel.offsetWidth, height: panel.offsetHeight },
    );
  }

  function reset(): void {
    local = null;
    const fn = o.onReset();
    if (fn) {
      Promise.resolve(fn()).catch(() => {
        local = undefined;
      });
    }
  }

  return {
    /** Where the panel is: mid-drag, else this tab's, else the server's. */
    get placed(): StackPositionView | null {
      return placed;
    },
    get dragging(): boolean {
      return dragging !== null;
    },
    reset,
    onPointerDown(e: PointerEvent): void {
      const panel = o.panel();
      if (e.button !== 0 || !panel || !o.board()) return;
      if ((e.target as Element | null)?.closest("button")) return;
      o.before?.();
      const r = panel.getBoundingClientRect();
      drag = {
        id: e.pointerId,
        dx: e.clientX - r.left,
        dy: e.clientY - r.top,
        x0: e.clientX,
        y0: e.clientY,
        moved: false,
      };
      (e.currentTarget as Element).setPointerCapture?.(e.pointerId);
      e.preventDefault();
    },
    onPointerMove(e: PointerEvent): void {
      if (!drag || e.pointerId !== drag.id) return;
      if (!drag.moved && Math.hypot(e.clientX - drag.x0, e.clientY - drag.y0) < DRAG_SLOP) return;
      drag.moved = true;
      dragging = spotAt(e);
    },
    onPointerUp(e: PointerEvent): void {
      if (!drag || e.pointerId !== drag.id) return;
      const spot = drag.moved ? spotAt(e) : null;
      drag = null;
      dragging = null;
      if (!spot) return;
      local = spot;
      const save = o.onSave();
      if (save) {
        Promise.resolve(save(spot)).catch(() => {
          local = undefined;
        });
      }
    },
    onPointerCancel(): void {
      drag = null;
      dragging = null;
    },
  };
}
