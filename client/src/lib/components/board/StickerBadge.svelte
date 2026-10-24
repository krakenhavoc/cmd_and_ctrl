<script lang="ts">
  // StickerBadge — the stickers on a card (ADR 0208, CR 123): a sticker
  // glyph and how many, the stickers themselves on hover. Drawn only
  // while the card has one. The card's name, abilities and power and
  // toughness already show what the stickers changed; this says why.
  import Icon from "../Icon.svelte";
  import { L } from "../../labels";
  import type { StickerView } from "../../protocol";
  import { stickersTitle } from "../../stickers";

  let { stickers }: { stickers?: StickerView[] } = $props();
  const list = $derived(stickers ?? []);
</script>

{#if list.length > 0}
  <span class="sticker-badge" title={stickersTitle(list)} aria-label={L.stickersOnCard}>
    <Icon name="sticker" size={11} /><span class="n">{list.length}</span>
  </span>
{/if}

<style>
  /* Left edge, half way down: clear of the CMD / designation badges at
     the top, the token and phased marks at the bottom and the counter
     pips on the right. */
  .sticker-badge {
    position: absolute;
    left: 3px;
    top: 50%;
    transform: translateY(-50%);
    z-index: 4;
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px 5px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--bg) 88%, transparent);
    color: var(--accent);
    border: 1px solid color-mix(in srgb, var(--accent) 45%, transparent);
    font-size: 9px;
    font-weight: 800;
    line-height: 1.2;
  }
  .n {
    font-family: var(--font-mono, ui-monospace, monospace);
    font-variant-numeric: tabular-nums;
  }
</style>
