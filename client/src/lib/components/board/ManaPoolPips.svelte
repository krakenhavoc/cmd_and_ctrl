<script lang="ts">
  // ManaPoolPips renders a horizontal row of WUBRGC pips showing
  // how much of each color is currently floating in a player's
  // mana pool. Driven by PlayerView.mana_pool (S15). Zero-count
  // colors are hidden so an empty pool collapses to zero DOM;
  // any non-empty pool fans out one pill per represented color
  // with a count badge when multiples stack.
  //
  // Mana pools empty at every step boundary (CR 106.4), so this
  // bar is transient — usually empty, populated only during an
  // active cast flow. That's fine: the pips pop in when something
  // fires an ability and vanish on the next step-advance snapshot.
  //
  // #1438: each pip draws the same ManaSymbol the click-for-mana
  // picker does, so the {G} a player just picked is the {G} that
  // lands here.
  import { manaSymbolMeta } from "../../manaSymbol";
  import ManaSymbol from "./ManaSymbol.svelte";

  interface Props {
    pool: string[] | undefined;
  }

  const { pool }: Props = $props();

  const COLOR_ORDER = ["W", "U", "B", "R", "G", "C"] as const;

  // Histogram of (color → count). Empty keys dropped.
  const counts = $derived.by(() => {
    const out: Record<string, number> = {};
    if (!pool || pool.length === 0) return out;
    for (const token of pool) {
      const normalized = token.toUpperCase();
      out[normalized] = (out[normalized] ?? 0) + 1;
    }
    return out;
  });
</script>

{#if pool && pool.length > 0}
  <div class="mana-pool" aria-label="mana pool">
    {#each COLOR_ORDER as color (color)}
      {#if (counts[color] ?? 0) > 0}
        {@const name = manaSymbolMeta(color).name.toLowerCase()}
        <span
          class="pip"
          data-color={color}
          title={`${counts[color]} ${name}`}
          aria-label={`${counts[color]} ${name} mana`}
        >
          <ManaSymbol symbol={color} size={16} />
          {#if counts[color] > 1}
            <span class="count">×{counts[color]}</span>
          {/if}
        </span>
      {/if}
    {/each}
  </div>
{/if}

<style>
  .mana-pool {
    display: inline-flex;
    flex-wrap: nowrap;
    gap: 4px;
    align-items: center;
    font-size: 11px;
    font-weight: 700;
    line-height: 1;
  }
  .pip {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 1px;
    background: rgba(10, 14, 26, 0.85);
    color: #e8ecf6;
    border: 1px solid rgba(255, 255, 255, 0.18);
    border-radius: 999px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.35);
  }
  .count {
    padding-right: 5px;
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0;
  }
</style>
