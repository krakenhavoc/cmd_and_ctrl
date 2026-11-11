<script lang="ts">
  // DraftCardArt — one draft card as its image, the way the catalogue
  // draws a card: the Scryfall scan at the card's own shape, loaded only
  // as it nears the screen, with the `cardArt` retry marker (#33).
  //
  // Under the image sits the card as words (name, cost, type line). It
  // shows while the scan loads, when the card has no printing to draw,
  // and when the image route refuses, so a card can always be read and
  // picked.
  import { cardArt } from "../../cardArt";
  import { draftImageURL, type DraftCard } from "../../draft";
  import { cardEdge } from "../../draftEdge";
  import ManaCost from "../board/ManaCost.svelte";

  interface Props {
    card: DraftCard;
    /** The image's alt text. "" inside a control that names the card itself. */
    alt?: string;
  }
  const { card, alt = "" }: Props = $props();

  const src = $derived(draftImageURL(card, "normal"));
  // The colour edge (#5704): a slight band round the card, from its
  // colour identity. Decoration only: the cost, the art and the words
  // say the same thing.
  const edge = $derived(cardEdge(card));
  // The image is hidden once it has failed, so the words under it show;
  // a retry that loads brings it back.
  let failed = $state(false);
</script>

<span class="art" data-card={card.id} data-identity={edge.key} style:--edge={edge.background}>
  <span class="words" aria-hidden={src && !failed ? "true" : undefined}>
    <span class="w-name">{card.name}</span>
    {#if card.mana_cost}<ManaCost cost={card.mana_cost} size={13} />{/if}
    {#if card.type_line}<span class="w-type">{card.type_line}</span>{/if}
  </span>
  {#if src}
    <img
      {src}
      {alt}
      class:failed
      loading="lazy"
      decoding="async"
      width="488"
      height="680"
      onerror={() => (failed = true)}
      onload={() => (failed = false)}
      use:cardArt={src}
    />
  {/if}
</span>

<style>
  .art {
    /* positioned for the failed-art pip (#33) */
    position: relative;
    --art-error-top: 8px;
    --art-error-right: 8px;
    display: block;
    aspect-ratio: 488 / 680;
    border-radius: 4.75% / 3.5%;
    overflow: hidden;
    /* The edge is the border: the gradient shows through a transparent
       border and the card keeps its own ground inside it. A hairline
       outside keeps a light edge (white) readable on a light page. */
    border: 3px solid transparent;
    background:
      linear-gradient(var(--surface-sunken), var(--surface-sunken)) padding-box,
      var(--edge, none) border-box;
    box-shadow: 0 0 0 1px var(--border);
    box-sizing: border-box;
  }
  .words {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 10px;
    text-align: center;
  }
  .w-name {
    font-family: var(--font-display);
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .w-type {
    font-size: 0.6875rem;
    line-height: 1.4;
    color: var(--fg-muted);
    overflow-wrap: anywhere;
  }
  img {
    position: relative;
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  img.failed {
    visibility: hidden;
  }
</style>
