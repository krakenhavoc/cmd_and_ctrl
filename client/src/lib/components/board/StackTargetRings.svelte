<script lang="ts">
  // StackTargetRings — rings what the stack targets, in every stack
  // style (ADR 0119 §4).
  //
  // While an item is on the stack, each permanent and player header it
  // targets carries `data-stack-lane-target` and a `--stack-lane-ring`
  // colour: gold for the top item, the caster's seat colour for the
  // rest (lib/stackArrows.ts, planTargetMarks / findTargetMarks /
  // syncTargetMarks). The fan and the pile used to do this themselves,
  // from their own arrows, so compact (now the phone style), spotlight
  // and ribbon rang nothing. Compact is drawn in the attention strip,
  // not inside StackLaneHost, so the one owner of the rings is this
  // component, which the board mounts once for every style.
  //
  // Renders nothing. Re-syncs on every snapshot and on board resize
  // (a seat collapsing or expanding remounts its cards), and clears
  // every mark it set when it is destroyed.

  import { onDestroy, untrack } from "svelte";
  import type { GameView } from "../../protocol";
  import { buildStackLane } from "../../stackLane";
  import { findTargetMarks, syncTargetMarks } from "../../stackArrows";

  interface Props {
    view: GameView;
    viewerID: string | null;
    boardEl: HTMLElement | null;
  }

  const { view, viewerID, boardEl }: Props = $props();

  // Only the stack's items, their targets and their casters' colours
  // are read, so the model is built without the summary's inputs.
  const items = $derived(
    buildStackLane({
      stack: view.stack,
      stackItems: view.stack_items,
      pendingTriggers: [],
      seats: view.seats,
      battlefield: view.battlefield,
      exile: view.exile,
      viewerID,
      priorityHolder: null,
      splitSecondActive: false,
    }).stackItems,
  );

  let marked = new Set<HTMLElement>();

  function sync(): void {
    const board = boardEl;
    marked = syncTargetMarks(marked, board ? findTargetMarks(board, items) : []);
  }

  const raf: (cb: () => void) => number =
    typeof requestAnimationFrame === "function"
      ? (cb) => requestAnimationFrame(cb)
      : (cb) => setTimeout(cb, 16) as unknown as number;
  const cancelRaf: (h: number) => void =
    typeof cancelAnimationFrame === "function"
      ? (h) => cancelAnimationFrame(h)
      : (h) => clearTimeout(h);
  let pending = 0;
  function schedule(): void {
    cancelRaf(pending);
    pending = raf(() => untrack(sync));
  }

  $effect(() => {
    // Now (the DOM is current when effects run) and again next frame,
    // for a card still mounting.
    void items;
    const board = boardEl;
    untrack(sync);
    if (!board) return;
    schedule();
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(schedule);
    ro.observe(board);
    return () => ro.disconnect();
  });

  onDestroy(() => {
    cancelRaf(pending);
    marked = syncTargetMarks(marked, []);
  });
</script>

<style>
  /* The ring. Outline and a drop-shadow rather than box-shadow, so it
     composes with a card's own selection or target ring instead of
     replacing it. */
  :global([data-stack-lane-target]) {
    outline: 2px solid var(--stack-lane-ring, var(--gold));
    outline-offset: 3px;
    filter: drop-shadow(
      0 0 8px color-mix(in srgb, var(--stack-lane-ring, var(--gold)) 55%, transparent)
    );
  }
</style>
