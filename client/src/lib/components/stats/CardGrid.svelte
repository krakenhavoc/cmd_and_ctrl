<script lang="ts" module>
  export interface GridCard {
    name: string;
    scryfallId?: string;
    /** The numbers under the name: "9 casts in 6 games". */
    meta: string;
    /** A word tag under that, in words (never colour alone). */
    tag?: string;
  }
</script>

<script lang="ts">
  // CardGrid — cards as images with a name and a line of numbers under
  // each (ADR 0217 §8: most-cast, dead cards). Card images need a
  // session, so a shared view (`names`) lists the names alone. An image
  // whose card has no printing is left out and the name stands alone.
  import { scryfallImageURL } from "../../cardImage";

  interface Props {
    cards: GridCard[];
    names?: boolean;
  }
  const { cards, names = false }: Props = $props();
</script>

<ul class="card-grid" class:names-only={names}>
  {#each cards as c (c.name)}
    <li>
      {#if !names && c.scryfallId}
        <img
          src={scryfallImageURL(c.scryfallId, "normal")}
          alt={c.name}
          loading="lazy"
          width="146"
          height="204"
        />
      {/if}
      <span class="cname">{c.name}</span>
      <span class="cmeta">{c.meta}</span>
      {#if c.tag}<span class="ctag">{c.tag}</span>{/if}
    </li>
  {/each}
</ul>

<style>
  .card-grid {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(118px, 1fr));
    gap: 12px;
  }
  .card-grid.names-only {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  }
  li {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }
  img {
    width: 100%;
    height: auto;
    aspect-ratio: 488 / 680;
    border-radius: 5%;
    background: var(--surface-sunken);
    margin-bottom: 4px;
  }
  .cname {
    font-weight: 600;
    font-size: 0.88rem;
    overflow-wrap: anywhere;
  }
  .cmeta {
    color: var(--fg-muted);
    font-size: 0.8rem;
    margin: 0;
  }
  .ctag {
    align-self: flex-start;
    margin-top: 2px;
    padding: 1px 7px;
    border: 1px solid var(--border);
    border-radius: 999px;
    font-size: 0.74rem;
    color: var(--fg);
  }
</style>
