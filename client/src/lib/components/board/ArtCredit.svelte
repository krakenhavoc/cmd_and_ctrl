<script lang="ts">
  // ArtCredit — "Illus. <artist>" over the foot of an art crop (#2982).
  // For a surface that needs no layout of its own around the line (the
  // stack tile); the board's art tile reads the artist itself because
  // its top-anchored marks have to step down around the credit.
  // Renders nothing until the printing's metadata names an artist.
  import { creditText, useArtist } from "../../artistCredit.svelte";

  let { scryfallId, face = 0 }: { scryfallId: string | null | undefined; face?: number } = $props();

  const artist = useArtist(() => ({ id: scryfallId, face }));
</script>

{#if artist.value}
  <span class="art-credit" aria-hidden="true">{creditText(artist.value)}</span>
{/if}

<style>
  /* Over the crop, never over a full card: near-white on a dark band
     reads over any art in every theme, like the board tile's name strip. */
  .art-credit {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 1;
    box-sizing: border-box;
    padding: 1px 4px;
    font-size: 8px;
    line-height: 1.3;
    font-style: italic;
    color: #e6e0cf;
    background: rgba(6, 9, 18, 0.62);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    pointer-events: none;
  }
</style>
