<script lang="ts">
  // DraftDeckColumns — the deckbuilding layout of the picks (#5848 §5):
  // Main in columns by mana value, the cards in each column stacked so
  // every name shows, and the Sideboard in a grid below. Deck view, in
  // the middle of a draft, is this layout.
  //
  // The component owns no state: `side` is the ids in the Sideboard and
  // `onmove` says where a card goes. A card is a button named for what a
  // click does ("Move to sideboard: Shoreline Looter"), and a move is
  // said once for a screen reader, as in DraftSplit. A card also drags to
  // the other area with the pointer (a held finger on touch) and shows
  // its large preview, with the same helpers as DraftSplit (#5848 §1, §2).
  import { tick, type Snippet } from "svelte";

  import type { DraftCard } from "../../draft";
  import { draftCard } from "../../draftCardPointer";
  import { dragState } from "../../draftDrag.svelte";
  import { byManaValue, columnLabel } from "../../draftColumns";
  import { ratedCardName } from "../../draftRatings.svelte";
  import { splitCards, type Zone } from "../../draftSplit";
  import { L } from "../../labels";
  import DraftCardArt from "./DraftCardArt.svelte";
  import DraftPreview from "./DraftPreview.svelte";

  interface Props {
    cards: DraftCard[];
    /** The ids of the cards in the Sideboard; every other card is in Main. */
    side: string[];
    onmove: (id: string, to: Zone) => void;
    /** Drawn under each card's large preview (the reasons for its rating). */
    previewExtra?: Snippet<[DraftCard]>;
  }
  const { cards, side, onmove, previewExtra }: Props = $props();

  const zones = $derived(splitCards(cards, side));
  const columns = $derived(byManaValue(zones.main));

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
  const scope = `deck-${Math.random().toString(36).slice(2, 8)}`;
  let said = $state("");

  // A drop on the area the card is already in is not a move: the zones
  // never report one, so the drag's own `ondrop` is only ever the other area.
  function move(card: DraftCard, to: Zone): void {
    onmove(card.id, to);
    said = `Moved ${card.name} to ${to === "main" ? "main" : "sideboard"}.`;
    // The button that was pressed is gone from its area; focus follows
    // the card to the other one instead of to nowhere.
    void tick().then(() => {
      for (const el of root?.querySelectorAll<HTMLElement>("[data-move-id]") ?? []) {
        if (el.dataset.moveId === card.id) el.focus();
      }
    });
  }
</script>

{#snippet tile(card: DraftCard, to: Zone)}
  <button
    type="button"
    class="tile"
    aria-label={`${to === "side" ? "Move to sideboard" : "Move to main"}: ${ratedCardName(card)}`}
    title={`${to === "side" ? "Move to sideboard" : "Move to main"}: ${card.name}`}
    data-move-id={card.id}
    use:draftCard={{
      card,
      scope,
      drag: { from: to === "side" ? "main" : "side", ondrop: (z) => move(card, z) },
    }}
    onclick={() => move(card, to)}
  >
    <DraftCardArt {card} />
  </button>
{/snippet}

<div class="deck" bind:this={root} data-testid="draft-deck-columns">
  <section
    class="card zone"
    class:drop={dropOn === "main"}
    aria-labelledby="draft-deckview-main-h"
    data-testid="draft-deckview-main"
    data-drop-zone="main"
    bind:this={mainEl}
  >
    <h2 id="draft-deckview-main-h">
      {L.draftDeckZone} <span class="n">{zones.main.length}</span>
    </h2>
    {#if zones.main.length === 0}
      <p class="none">Nothing in your main yet. Click a card in your sideboard to play it.</p>
    {:else}
      <div class="cols">
        {#each columns as col, i (i)}
          <div class="col" data-mv={i}>
            <div class="colh"><span class="sr-only">mana value </span>{columnLabel(i)}</div>
            <div class="stack">
              {#each col as card (card.id)}
                {@render tile(card, "side")}
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  <section
    class="card zone"
    class:drop={dropOn === "side"}
    aria-labelledby="draft-deckview-side-h"
    data-testid="draft-deckview-side"
    data-drop-zone="side"
    bind:this={sideEl}
  >
    <h2 id="draft-deckview-side-h">
      {L.draftSideboardZone} <span class="n">{zones.side.length}</span>
    </h2>
    {#if zones.side.length === 0}
      <p class="none">Nothing set aside. Click a card in your main to set it aside.</p>
    {:else}
      <ul class="sidegrid">
        {#each zones.side as card (card.id)}
          <li>{@render tile(card, "main")}</li>
        {/each}
      </ul>
    {/if}
  </section>

  <p class="sr-only" role="status" data-testid="draft-deckview-said">{said}</p>
  <DraftPreview {scope} children={previewExtra} />
</div>

<style>
  .deck {
    display: flex;
    flex-direction: column;
    gap: 22px;
    min-width: 0;
  }
  .zone {
    min-width: 0;
  }
  /* A zone a card is being dragged over, as in DraftSplit. */
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
  .none {
    margin: 0;
    padding: 14px 8px;
    border: 1px dashed var(--border-strong);
    border-radius: var(--radius);
    font-size: 0.8125rem;
    text-align: center;
    color: var(--fg-muted);
  }
  .cols {
    display: grid;
    grid-template-columns: repeat(7, minmax(0, 1fr));
    gap: 10px;
  }
  .col {
    min-width: 0;
  }
  .colh {
    margin-bottom: 4px;
    text-align: center;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--fg-muted);
  }
  /* Each card but the first is pulled up over the one above it by most of
     its own height, so only the name strip of each shows. A percentage
     margin is of the column's width, and a card is 1.393 times as tall as
     it is wide. */
  .stack {
    display: flex;
    flex-direction: column;
  }
  .stack > .tile + .tile {
    margin-top: -113%;
  }
  .stack > .tile:hover,
  .stack > .tile:focus-visible {
    z-index: 2;
  }
  .tile {
    position: relative;
    display: block;
    width: 100%;
    padding: 0;
    border: 0;
    background: none;
    border-radius: 4.75% / 3.5%;
    cursor: pointer;
  }
  .tile:hover {
    box-shadow: 0 0 0 2px var(--accent);
  }
  .tile:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  .sidegrid {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 8px;
    grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
  }
  @media (max-width: 1099px) {
    .cols {
      grid-template-columns: repeat(4, minmax(0, 1fr));
    }
  }
  @media (max-width: 719px) {
    .cols {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
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
