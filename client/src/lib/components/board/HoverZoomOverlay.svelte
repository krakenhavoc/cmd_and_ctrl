<script lang="ts">
  // HoverZoomOverlay is the large floating card preview anchored at
  // (0.8, 0.2) of the board area — top-right quadrant. Driven by the
  // module-scope `hoveredCard` store written from Card.svelte's
  // pointerenter / pointerleave. Replaces the Pixi ZoomPreview from
  // client/src/lib/zoom-preview.ts.
  //
  // Image fetch is the browser's HTTP cache; the second hover of the
  // same card paints instantly as long as the server emits cache
  // headers on /cards/:id/image. The overlay is pointer-events:none
  // so it never eats clicks meant for the table behind it.

  import { hoveredCard } from "../../cardTypes";

  const card = $derived($hoveredCard);
  const imgSrc = $derived(
    card?.scryfall_id ? `/cards/${card.scryfall_id}/image?size=normal` : null,
  );
</script>

{#if card}
  <div class="overlay" aria-hidden="true">
    {#if imgSrc}
      <img src={imgSrc} alt="" />
    {:else}
      <span class="name">{card.name}</span>
    {/if}
  </div>
{/if}

<style>
  .overlay {
    position: absolute;
    /* Anchor centre at (0.8, 0.2) of the board area. translate(-50%,
       -50%) re-centres around the chosen anchor so the visual centre,
       not the top-left, sits at the percentage. */
    left: 80%;
    top: 20%;
    transform: translate(-50%, -50%);
    width: clamp(180px, 22vw, 360px);
    aspect-ratio: 63 / 88;
    background: #0b1220;
    border: 1px solid #4a5270;
    border-radius: 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
    pointer-events: none;
    z-index: 40;
    overflow: hidden;
    padding: 6px;
    box-sizing: border-box;
  }
  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
    border-radius: 8px;
  }
  .name {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    color: #e0e6f5;
    font-size: 14px;
    text-align: center;
    padding: 12px;
    box-sizing: border-box;
  }
</style>
