<script lang="ts">
  // PageTipSlot — where a site page draws its first-use tip (ADR 0156 §3
  // item 5): in the page frame, under the lede, instead of over the
  // page's controls. Page.svelte mounts one; HintLayer sees its
  // attribute and publishes the page's tips as "inline", and this draws
  // them. The registry, one tip per visit and Hide tips are unchanged
  // (ADR 0125). Renders an empty slot when no tip is up.

  import { activeTip, TIP_SLOT_ATTR } from "../../hints/runtime";
  import TipCard from "./TipCard.svelte";

  const attrs = { [TIP_SLOT_ATTR]: "" };
</script>

<div class="page-tip" {...attrs}>
  {#if $activeTip && $activeTip.placement.kind === "inline"}
    <TipCard tip={$activeTip} />
  {/if}
</div>

<style>
  .page-tip {
    max-width: 880px;
  }
  .page-tip :global(.hint-card) {
    margin-top: 16px;
  }
</style>
