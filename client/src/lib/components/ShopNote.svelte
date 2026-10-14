<script lang="ts">
  // ShopNote — the disclosure every group of shop links carries (ADR
  // 0179 §4), with the way to hide them. Renders nothing when the viewer
  // sees no shop links (none configured, or hidden), so a caller can put
  // it beside a group without its own check.
  import { SHOP_DISCLOSURE } from "../shopLinks";
  import { setShopLinksShown, visibleShops } from "../shopLinksStore";
</script>

{#if $visibleShops.length > 0}
  <p class="shop-note">
    {SHOP_DISCLOSURE}
    <button
      type="button"
      class="hide"
      title="Show them again in Settings, Display"
      onclick={() => setShopLinksShown(false)}
    >
      Hide shop links
    </button>
  </p>
{/if}

<style>
  .shop-note {
    margin: 0;
    font-size: 11.5px;
    line-height: 1.45;
    color: var(--fg-dim);
  }
  .hide {
    all: unset;
    margin-left: 4px;
    color: var(--fg-muted);
    text-decoration: underline;
    text-underline-offset: 2px;
    cursor: pointer;
  }
  .hide:hover {
    color: var(--fg);
  }
  .hide:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
    border-radius: 2px;
  }
</style>
