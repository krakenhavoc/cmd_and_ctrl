<script lang="ts">
  // ShopLinks — one group of links to buy the real cards (ADR 0179): a
  // card search per shop (`card`), or a pasted-list cart per shop
  // (`list`). Renders nothing unless the operator configured a shop and
  // the viewer has not hidden them.
  //
  // NEVER place this on the game table while a game is running (ADR
  // 0179 §3). The placements are the Decks page, the Catalog, History's
  // ended games (GameCardsPanel; the server refuses the read until the
  // game has ended), the pre-built decks and the stats page (§7).
  //
  // `note` (default on) shows the disclosure under the links. A surface
  // with many small groups (a card per row) shows one ShopNote above
  // them instead and passes note={false}.
  import { cardShopLinks, listShopLinks, type CardLine } from "../shopLinks";
  import { visibleShops } from "../shopLinksStore";
  import ShopNote from "./ShopNote.svelte";

  interface Props {
    /** A card's name: one search link per shop. */
    card?: string;
    /** A list of cards: one cart link per shop. Wins over `card`. */
    list?: readonly CardLine[];
    /** The words before the links. */
    label?: string;
    /** Show the disclosure under the links. */
    note?: boolean;
    /** A single small line, for a tile or a row. */
    compact?: boolean;
  }

  const { card, list, label = "Buy", note = true, compact = false }: Props = $props();

  const links = $derived(
    list ? listShopLinks($visibleShops, list) : card ? cardShopLinks($visibleShops, card) : [],
  );
  const what = $derived(list ? "" : card ? ` ${card}` : "");
</script>

{#if links.length > 0}
  <div class="shop" class:compact>
    <p class="links">
      <span class="lead">{label}:</span>
      {#each links as l, i (l.id)}
        {#if i > 0}<span class="sep" aria-hidden="true">·</span>{/if}
        <a
          href={l.href}
          target="_blank"
          rel="noopener noreferrer sponsored"
          aria-label="{label}{what} at {l.name} (opens in a new tab)">{l.name}</a
        >
      {/each}
    </p>
    {#if note}<ShopNote />{/if}
  </div>
{/if}

<style>
  .shop {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }
  .links {
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 4px 8px;
    font-size: 13px;
  }
  .lead {
    color: var(--fg-muted);
    font-weight: 600;
  }
  .sep {
    color: var(--fg-dim);
  }
  a {
    color: var(--accent-strong);
    font-weight: 700;
    text-decoration: none;
  }
  a:hover {
    color: var(--accent);
    text-decoration: underline;
  }
  a:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
    border-radius: 2px;
  }
  .compact .links {
    font-size: 11.5px;
    gap: 2px 6px;
  }
  .compact .lead {
    font-weight: 500;
    color: var(--fg-dim);
  }
  .compact a {
    font-weight: 600;
    color: var(--fg-muted);
  }
  .compact a:hover {
    color: var(--accent-strong);
  }
</style>
