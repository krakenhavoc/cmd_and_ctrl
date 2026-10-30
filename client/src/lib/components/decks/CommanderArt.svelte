<script lang="ts">
  // CommanderArt — one commander's art crop leading a saved deck's card
  // (ADR 0158 §2), with the artist credited on the art (#2982): an art
  // crop drops Scryfall's artist line, so we print it. An image that
  // fails reports it and draws nothing, so the card falls back to its
  // plain band rather than a broken image. Spans only: the deck picker's
  // tiles and your seat's tile draw it inside a label or a button.
  import { creditText, useArtist } from "../../artistCredit.svelte";
  import { scryfallImageURL } from "../../cardImage";

  let { scryfallId, name, onfail }: { scryfallId: string; name: string; onfail: () => void } =
    $props();

  const artist = useArtist(() => ({ id: scryfallId }));
  let failed = $state(false);
</script>

{#if !failed}
  <span class="crop" title={artist.value ? `${name} (${creditText(artist.value)})` : name}>
    <img
      src={scryfallImageURL(scryfallId, "art_crop")}
      alt=""
      loading="lazy"
      decoding="async"
      onerror={() => {
        failed = true;
        onfail();
      }}
    />
    {#if artist.value}
      <span class="art-credit" aria-hidden="true">{creditText(artist.value)}</span>
    {/if}
  </span>
{/if}

<style>
  .crop {
    position: relative;
    display: block;
    flex: 1 1 0;
    min-width: 0;
    height: 100%;
    overflow: hidden;
  }
  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center 28%;
  }
  /* Top right, clear of the fade at the foot: near-white on a dark chip
     reads over any art in every skin, like the board's credit. */
  .art-credit {
    position: absolute;
    top: 6px;
    right: 6px;
    max-width: calc(100% - 12px);
    box-sizing: border-box;
    padding: 1px 6px;
    border-radius: 999px;
    font-size: 9.5px;
    line-height: 1.5;
    font-style: italic;
    color: #e6e0cf;
    background: rgba(6, 9, 18, 0.6);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    pointer-events: none;
  }
</style>
