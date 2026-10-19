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
  // #4114: a target of a divided effect (Armament Dragon's counters,
  // Fireball's damage) also carries a small badge with its share, in the
  // ring's colour (lib/stackShares.ts). The badges are fixed-position
  // elements measured from the target's own rect, so they sit above any
  // card without touching the card's markup.
  //
  // Re-syncs on every snapshot and on board resize (a seat collapsing
  // or expanding remounts its cards), and clears every mark it set when
  // it is destroyed.

  import { onDestroy, untrack } from "svelte";
  import { metaFor } from "../../cardMetaCache";
  import type { CardView, GameView } from "../../protocol";
  import { buildStackLane } from "../../stackLane";
  import {
    findSourceMarks,
    findTargetMarks,
    syncSourceMarks,
    syncTargetMarks,
  } from "../../stackArrows";
  import { measureShareBadges, type ShareBadge } from "../../stackShares";

  interface Props {
    view: GameView;
    viewerID: string | null;
    boardEl: HTMLElement | null;
  }

  const { view, viewerID, boardEl }: Props = $props();

  // #4114: a spell's share badge says "dmg" or "+N" from its card's
  // oracle text, which is not on the wire. Only a spell that divides
  // something needs it, so only those printings are asked for (an
  // ability carries its own text).
  let oracle = $state<Record<string, string>>({});
  const dividedPrintings = $derived(
    (view.stack_items ?? [])
      .filter((it) => it.kind === "spell" && it.distribution)
      .map((it) => view.stack.cards.find((c) => c.instance_id === it.id)?.scryfall_id)
      .filter((id): id is string => !!id)
      .sort()
      .join(","),
  );
  $effect(() => {
    const ids = dividedPrintings === "" ? [] : dividedPrintings.split(",");
    const unsubs = ids.map((id) =>
      metaFor(id).subscribe((meta) => {
        const text = meta?.oracle_text;
        if (!text) return;
        untrack(() => {
          if (oracle[id] !== text) oracle = { ...oracle, [id]: text };
        });
      }),
    );
    return () => unsubs.forEach((u) => u());
  });
  function oracleTextFor(card: CardView): string | undefined {
    return card.scryfall_id ? oracle[card.scryfall_id] : undefined;
  }

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
      oracleTextFor,
    }).stackItems,
  );

  let marked = new Set<HTMLElement>();
  let sources = new Set<HTMLElement>();
  let badges = $state.raw<ShareBadge[]>([]);

  function sync(): void {
    const board = boardEl;
    marked = syncTargetMarks(marked, board ? findTargetMarks(board, items) : []);
    sources = syncSourceMarks(sources, board ? findSourceMarks(board, items) : []);
    badges = board ? measureShareBadges(board, items) : [];
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
    // A scrolled seat or lane moves its cards without resizing them.
    window.addEventListener("scroll", schedule, true);
    return () => {
      ro.disconnect();
      window.removeEventListener("scroll", schedule, true);
    };
  });

  onDestroy(() => {
    cancelRaf(pending);
    marked = syncTargetMarks(marked, []);
    sources = syncSourceMarks(sources, []);
  });
</script>

{#each badges as b (b.key)}
  <span
    class="share-badge"
    data-stack-share={b.text}
    title={b.title}
    style:left="{b.right}px"
    style:top="{b.top}px"
    style:--share-color={b.color}>{b.text}</span
  >
{/each}

<style>
  /* #4114: what this target gets of a divided effect. Pinned to the
     target's top-right corner and nudged out over its edge so it never
     hides the card's name or art; pointer-events off so the card under
     it stays clickable. */
  .share-badge {
    position: fixed;
    z-index: 40;
    transform: translate(-70%, -35%);
    pointer-events: none;
    padding: 1px 6px;
    border-radius: 999px;
    border: 1.5px solid var(--share-color, var(--accent));
    background: var(--surface-raised, #1b1b22);
    color: var(--text, #fff);
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 700;
    line-height: 1.3;
    white-space: nowrap;
    box-shadow: 0 2px 8px rgb(0 0 0 / 0.5);
  }

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
