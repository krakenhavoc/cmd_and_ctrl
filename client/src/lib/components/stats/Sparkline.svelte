<script lang="ts">
  // Sparkline — a deck's last results as a row of small bars, oldest on
  // the left (ADR 0217 §8). A win is a full bar in the accent of wins, a
  // draw half, a loss a quarter: height carries the result as well as
  // colour. It is one image whose name says the count.
  import { sparkCells } from "../../statsCharts";

  interface Props {
    results: string[];
    label: string;
  }
  const { results, label }: Props = $props();
  const cells = $derived(sparkCells(results));
  const W = 100;
  const H = 20;
</script>

{#if cells.length > 0}
  <svg class="spark" viewBox="0 0 {W} {H}" preserveAspectRatio="none" role="img" aria-label={label}>
    {#each cells as c, i (i)}
      <rect
        x={(i * W) / 20 + 0.6}
        y={H - c.height * H}
        width={W / 20 - 1.2}
        height={c.height * H}
        class={c.result}
      />
    {/each}
  </svg>
{/if}

<style>
  .spark {
    width: 100%;
    max-width: 160px;
    height: 20px;
    display: block;
  }
  .win {
    fill: var(--chart-1);
  }
  .draw {
    fill: var(--chart-6);
  }
  .loss,
  .none {
    fill: var(--fg-muted);
    opacity: 0.6;
  }
</style>
