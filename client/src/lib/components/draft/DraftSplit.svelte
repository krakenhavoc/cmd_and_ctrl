<script lang="ts">
  // DraftSplit — the viewer's picks in two areas, Main and Sideboard
  // (#5699). It is the one component of the pick screen's panel beside
  // the pack and of the deckbuilding screen's zones.
  //
  // A card moves with a drag onto the other area (a pointer-event drag,
  // so a mouse, a pen, or a finger held for a moment, #5848), or with a
  // click or Enter on the card, which is a button named for what it does
  // ("Move to sideboard: Shoreline Looter"). A card dragged here from the
  // pack is a pick: DraftPack does that, and these areas are only the
  // drop targets (`data-drop-zone`). The component owns no state: `side`
  // is the ids in the Sideboard and `onmove` says where a card goes; the
  // page keeps the split.
  import { tick, type Snippet } from "svelte";

  import type { DraftCard } from "../../draft";
  import { dragState } from "../../draftDrag.svelte";
  import { splitCards, type Zone } from "../../draftSplit";
  import { L } from "../../labels";
  import DraftPool from "./DraftPool.svelte";
  import DraftPreview from "./DraftPreview.svelte";

  interface Props {
    cards: DraftCard[];
    /** The ids of the cards in the Sideboard; every other card is in Main. */
    side: string[];
    onmove: (id: string, to: Zone) => void;
    /** Draw the cards as images (the deckbuilding screen) instead of rows. */
    images?: boolean;
    /** The two areas' ids and test ids: `${main}`, `${side}`, with `-h` on the headings. */
    ids?: { main: string; side: string };
    emptyMain?: string;
    emptySide?: string;
    /** Drawn under each card's large preview (the reasons for its rating). */
    previewExtra?: Snippet<[DraftCard]>;
  }
  const {
    cards,
    side,
    onmove,
    images = false,
    ids = { main: "draft-deck", side: "draft-side" },
    emptyMain = "No cards yet.",
    emptySide = "Nothing set aside.",
    previewExtra,
  }: Props = $props();

  const zones = $derived(splitCards(cards, side));

  let root = $state<HTMLElement | null>(null);
  let mainEl = $state<HTMLElement | null>(null);
  let sideEl = $state<HTMLElement | null>(null);
  // The area a card is being dragged over, for its highlight.
  const dropOn = $derived<Zone | null>(
    dragState.over && dragState.over === mainEl
      ? "main"
      : dragState.over && dragState.over === sideEl
        ? "side"
        : null,
  );
  // Both areas share one preview host, so a swap from one to the other is instant.
  const scope = `split-${Math.random().toString(36).slice(2, 8)}`;
  // Said once per move, for a screen reader (DESIGN.md §8).
  let said = $state("");

  function nameOf(id: string): string {
    return cards.find((c) => c.id === id)?.name ?? "The card";
  }

  function move(id: string, to: Zone, keepFocus: boolean): void {
    onmove(id, to);
    said = `Moved ${nameOf(id)} to ${to === "main" ? "main" : "sideboard"}.`;
    // The button that was pressed is gone from its area; focus follows
    // the card to the other one instead of to nowhere, and the view stays
    // where it is (#5921).
    if (keepFocus) {
      void tick().then(() => {
        const all = root?.querySelectorAll<HTMLElement>("[data-move-id]") ?? [];
        for (const el of all) if (el.dataset.moveId === id) el.focus({ preventScroll: true });
      });
    }
  }

  // A card dropped on the other area; a drop on its own is not a move.
  const dragged = (card: DraftCard, to: Zone): void => move(card.id, to, false);
  const toggle = (card: DraftCard, from: Zone): void =>
    move(card.id, from === "main" ? "side" : "main", true);
</script>

<div class="split" bind:this={root} class:images>
  <section
    class="card zone"
    class:drop={dropOn === "main"}
    aria-labelledby={`${ids.main}-h`}
    data-testid={ids.main}
    data-drop-zone="main"
    bind:this={mainEl}
  >
    <h2 id={`${ids.main}-h`}>
      {L.draftDeckZone} <span class="n">{zones.main.length}</span>
    </h2>
    <DraftPool
      cards={zones.main}
      {images}
      verb="Move to sideboard"
      empty={emptyMain}
      onpick={(card) => toggle(card, "main")}
      zone="main"
      ondrag={dragged}
      previewScope={scope}
    />
  </section>

  <section
    class="card zone"
    class:drop={dropOn === "side"}
    aria-labelledby={`${ids.side}-h`}
    data-testid={ids.side}
    data-drop-zone="side"
    bind:this={sideEl}
  >
    <h2 id={`${ids.side}-h`}>
      {L.draftSideboardZone} <span class="n">{zones.side.length}</span>
    </h2>
    <DraftPool
      cards={zones.side}
      {images}
      verb="Move to main"
      empty={emptySide}
      onpick={(card) => toggle(card, "side")}
      zone="side"
      ondrag={dragged}
      previewScope={scope}
    />
  </section>

  <p class="sr-only" role="status" data-testid="draft-split-said">{said}</p>
  <DraftPreview {scope} children={previewExtra} />
</div>

<style>
  .split {
    display: flex;
    flex-direction: column;
    gap: 14px;
    min-width: 0;
  }
  /* The deckbuilding screen's zones are the subject of the page, so
     they are a step more room apart. */
  .split.images {
    gap: 22px;
    /* Two zones share the page with the panel, so their tiles are a
       step smaller than a finished pool's: three across at 820px. */
    --pool-tile: 100px;
  }
  .zone {
    min-width: 0;
  }
  /* A zone a card is being dragged over: the hairline thickens to the
     accent, and nothing moves. */
  .zone.drop {
    border-color: var(--accent);
    box-shadow: 0 0 0 1px var(--accent);
    background: color-mix(in srgb, var(--accent) 6%, var(--surface));
  }
  h2 {
    margin: 0 0 12px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    /* a .card's h2 is a mono eyebrow by default; these are headings */
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  .n {
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>
