<script lang="ts">
  // DraftPool — the viewer's own cards, grouped by colour (W, U, B, R, G,
  // multicolour, colourless, lands apart) and by mana value inside each
  // group (lib/draft.ts groupPool).
  //
  // Two looks. During the picks it is a compact list: mana value, name,
  // cost, so the pack keeps the screen. Once the draft is done the pool
  // is the subject, so it is drawn as card images.
  //
  // `cards` is the viewer's `picks` or `pool`. No other seat's cards
  // exist on the wire (CR 905.1c).
  import {
    cardAccessibleName,
    DRAG_TYPE,
    groupPool,
    stackCopies,
    type DraftCard,
    type PoolGroupKey,
  } from "../../draft";
  import ManaCost from "../board/ManaCost.svelte";
  import DraftCardArt from "./DraftCardArt.svelte";

  interface Props {
    cards: DraftCard[];
    /** Draw the cards as images (a finished draft) instead of a list. */
    images?: boolean;
    /**
     * Makes every card a button (the deckbuilding screen, #5440): a
     * click calls it with the card, and the card can be dragged. `verb`
     * begins each button's name ("Add to deck: Lightning Bolt").
     */
    onpick?: (card: DraftCard) => void;
    verb?: string;
    /** What an empty pool says. */
    empty?: string;
  }
  const { cards, images = false, onpick, verb = "", empty = "No picks yet." }: Props = $props();

  // A dragged card carries its draft id; DraftBuild's zones read it.
  function dragStart(ev: DragEvent, card: DraftCard): void {
    ev.dataTransfer?.setData(DRAG_TYPE, card.id);
    if (ev.dataTransfer) ev.dataTransfer.effectAllowed = "move";
  }

  const groups = $derived(groupPool(cards));

  // The colour-identity tokens (DESIGN.md §2), which no skin changes.
  // Many colours and lands have no colour of their own: a neutral ink.
  const DOT: Record<PoolGroupKey, string> = {
    W: "var(--identity-w)",
    U: "var(--identity-u)",
    B: "var(--identity-b)",
    R: "var(--identity-r)",
    G: "var(--identity-g)",
    multi: "var(--fg)",
    colorless: "var(--identity-c)",
    land: "var(--fg-muted)",
  };
</script>

{#if groups.length === 0}
  <p class="none">{empty}</p>
{:else}
  <div class="groups" class:images data-testid="draft-pool">
    {#each groups as g (g.key)}
      <section class="group" data-group={g.key}>
        <h3>
          <span class="dot" style:--dot={DOT[g.key]} aria-hidden="true"></span>
          {g.name}
          <span class="n">{g.count}</span>
        </h3>
        {#if images}
          <ul class="imgs">
            {#each g.cards as card (card.id)}
              <li>
                {#if onpick}
                  <button
                    type="button"
                    class="move"
                    aria-label={`${verb}: ${cardAccessibleName(card)}`}
                    title={`${verb}: ${card.name}`}
                    draggable="true"
                    ondragstart={(ev) => dragStart(ev, card)}
                    onclick={() => onpick(card)}
                  >
                    <DraftCardArt {card} />
                  </button>
                {:else}
                  <DraftCardArt {card} alt={card.name} />
                {/if}
              </li>
            {/each}
          </ul>
        {:else}
          <ul class="rows">
            {#each g.byValue as row (row.manaValue)}
              {#each stackCopies(row.cards) as s (s.card.id)}
                <li>
                  <span class="mv"><span class="sr-only">{"mana value "}</span>{row.manaValue}</span
                  >
                  <span class="name">
                    {#if s.count > 1}<span class="x">{s.count}×</span>{" "}{/if}{s.card.name}
                  </span>
                  {#if s.card.mana_cost}<ManaCost cost={s.card.mana_cost} size={12} />{/if}
                </li>
              {/each}
            {/each}
          </ul>
        {/if}
      </section>
    {/each}
  </div>
{/if}

<style>
  .none {
    margin: 0;
    font-size: 0.8125rem;
    color: var(--fg-muted);
  }
  .groups {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
    gap: 16px 24px;
    min-width: 0;
  }
  .groups.images {
    grid-template-columns: minmax(0, 1fr);
    gap: 24px;
  }
  .group {
    min-width: 0;
  }
  h3 {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 8px;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .dot {
    width: 9px;
    height: 9px;
    border-radius: 999px;
    background: var(--dot);
    border: 1px solid var(--border-strong);
    box-sizing: border-box;
  }
  .n {
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  .rows,
  .imgs {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .rows li {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    padding: 3px 0;
    font-size: 0.8125rem;
    color: var(--fg);
  }
  .rows li + li {
    border-top: 1px solid var(--border);
  }
  .mv {
    flex: none;
    width: 2ch;
    text-align: right;
    font-family: var(--font-mono);
    font-size: 11px;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
  .name {
    flex: 1 1 auto;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .x {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--fg-muted);
  }
  /* A card that moves between the deck and the sideboard. The button
     is the card: no chrome of its own, a ring on hover and focus. */
  .move {
    display: block;
    width: 100%;
    padding: 0;
    border: 0;
    background: none;
    border-radius: 4.75% / 3.5%;
    cursor: pointer;
  }
  .move:hover {
    box-shadow: 0 0 0 2px var(--border-strong);
  }
  .move:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  .imgs {
    display: grid;
    /* the catalogue's tile width: six across at the 1200px page */
    grid-template-columns: repeat(auto-fill, minmax(128px, 1fr));
    gap: 14px;
  }
</style>
