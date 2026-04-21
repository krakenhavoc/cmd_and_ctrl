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

  interface Props {
    pool: string[] | undefined;
  }

  const { pool }: Props = $props();

  const COLOR_ORDER = ["W", "U", "B", "R", "G", "C"] as const;
  type ManaColor = (typeof COLOR_ORDER)[number];

  const COLOR_STYLE: Record<ManaColor, { fill: string; glyph: string; label: string }> = {
    W: { fill: "#f4ead5", glyph: "☀", label: "white" },
    U: { fill: "#aad4ff", glyph: "💧", label: "blue" },
    B: { fill: "#2b2b3d", glyph: "☠", label: "black" },
    R: { fill: "#ff9a85", glyph: "🔥", label: "red" },
    G: { fill: "#92c493", glyph: "🌿", label: "green" },
    C: { fill: "#c6cfdd", glyph: "◇", label: "colorless" },
  };

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
        {@const style = COLOR_STYLE[color]}
        <span
          class="pip"
          style:--fill={style.fill}
          title={`${counts[color]} ${style.label}`}
          aria-label={`${counts[color]} ${style.label} mana`}
        >
          <span class="glyph">{style.glyph}</span>
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
    padding: 2px 6px 2px 4px;
    background: var(--fill);
    color: #0a0e1a;
    border: 1px solid rgba(0, 0, 0, 0.35);
    border-radius: 999px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.35);
    min-width: 22px;
    text-align: center;
  }
  .glyph {
    display: inline-block;
    line-height: 1;
    filter: drop-shadow(0 1px 0 rgba(255, 255, 255, 0.25));
  }
  .count {
    font-size: 10px;
    font-weight: 800;
    letter-spacing: 0;
  }
</style>
