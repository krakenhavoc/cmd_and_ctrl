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
  // It also marks the other end (#3380): each stack item's caster
  // avatar, and an ability's source permanent (CR 113.7), carry
  // `data-stack-source` and glow in the caster's seat colour, pulsing
  // for the top item, so an item is tied to a seat on the table and
  // not only to a name in the stack panel (planSourceMarks /
  // findSourceMarks / syncSourceMarks).
  //
  // Renders nothing. Re-syncs on every snapshot and on board resize
  // (a seat collapsing or expanding remounts its cards), and clears
  // every mark it set when it is destroyed.

  import { onDestroy, untrack } from "svelte";
  import type { GameView } from "../../protocol";
  import { buildStackLane } from "../../stackLane";
  import {
    findSourceMarks,
    findTargetMarks,
    syncSourceMarks,
    syncTargetMarks,
  } from "../../stackArrows";

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
  let sources = new Set<HTMLElement>();

  function sync(): void {
    const board = boardEl;
    marked = syncTargetMarks(marked, board ? findTargetMarks(board, items) : []);
    sources = syncSourceMarks(sources, board ? findSourceMarks(board, items) : []);
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
    sources = syncSourceMarks(sources, []);
  });
</script>

<style>
  /* The ring. Outline and a drop-shadow rather than box-shadow, so it
     composes with a card's own selection or target ring instead of
     replacing it. */
  :global([data-stack-lane-target]) {
    outline: 2px solid var(--stack-lane-ring, var(--accent));
    outline-offset: 3px;
    filter: drop-shadow(
      0 0 8px color-mix(in srgb, var(--stack-lane-ring, var(--accent)) 55%, transparent)
    );
  }

  /* #3380: who put it there. A glow, not a ring, in the caster's seat
     colour, so it never reads as "targeted". The top item's pulses.
     The targeting prompt's gold source glow wins while it is open. */
  :global([data-stack-source]:not([data-targeting-source])) {
    filter: drop-shadow(
        0 0 4px color-mix(in srgb, var(--stack-source-glow, var(--accent)) 90%, transparent)
      )
      drop-shadow(
        0 0 12px color-mix(in srgb, var(--stack-source-glow, var(--accent)) 55%, transparent)
      );
  }
  :global([data-stack-source="next"]:not([data-targeting-source])) {
    animation: -global-stack-source-pulse 1.6s ease-in-out infinite alternate;
  }
  @keyframes -global-stack-source-pulse {
    from {
      filter: drop-shadow(
          0 0 3px color-mix(in srgb, var(--stack-source-glow, var(--accent)) 75%, transparent)
        )
        drop-shadow(
          0 0 8px color-mix(in srgb, var(--stack-source-glow, var(--accent)) 35%, transparent)
        );
    }
    to {
      filter: drop-shadow(0 0 6px var(--stack-source-glow, var(--accent)))
        drop-shadow(
          0 0 16px color-mix(in srgb, var(--stack-source-glow, var(--accent)) 70%, transparent)
        );
    }
  }
  /* Both a target and a source (a player who targets themself): the
     ring's outline stays, and the two glows stack. */
  :global([data-stack-lane-target][data-stack-source]:not([data-targeting-source])) {
    filter: drop-shadow(
        0 0 8px color-mix(in srgb, var(--stack-lane-ring, var(--accent)) 55%, transparent)
      )
      drop-shadow(
        0 0 6px color-mix(in srgb, var(--stack-source-glow, var(--accent)) 90%, transparent)
      );
    animation: none;
  }
  @media (prefers-reduced-motion: reduce) {
    :global([data-stack-source]) {
      animation: none;
    }
  }
  :global(:root[data-reduce-motion="1"] [data-stack-source]) {
    animation: none;
  }
</style>
