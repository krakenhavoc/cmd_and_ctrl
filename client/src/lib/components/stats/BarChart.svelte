<script lang="ts">
  // BarChart — horizontal bars, each with its own label and value text
  // (ADR 0217 §8), so nothing is told by colour alone. The labels are
  // HTML, so they stay readable at any width; each bar is a small SVG
  // track. An optional tick marks a comparison (an average). The chart
  // is one image named by its figure's caption.
  import { barShares, type ChartBar } from "../../statsCharts";

  interface Props {
    labelledby: string;
    bars: ChartBar[];
    /** The scale's top; the largest value (or tick) when unset. */
    max?: number;
  }
  const { labelledby, bars, max }: Props = $props();

  const top = $derived(max ?? Math.max(0, ...bars.map((b) => Math.max(b.value, b.tick ?? 0))));
  const shares = $derived(
    barShares(
      bars.map((b) => b.value),
      top,
    ),
  );
</script>

<div class="bars" role="img" aria-labelledby={labelledby}>
  {#each bars as b, i (b.label)}
    <div class="row" class:muted={b.muted}>
      <span class="label">
        {#if b.mark}<span class="mark" style:background={b.tone}>{b.mark}</span>{/if}
        {b.label}
      </span>
      <svg class="track" viewBox="0 0 100 10" preserveAspectRatio="none" aria-hidden="true">
        <rect class="bg" x="0" y="0" width="100" height="10" rx="1" />
        <rect
          class="fill"
          x="0"
          y="0"
          width={Math.max(shares[i] * 100, b.value > 0 ? 1 : 0)}
          height="10"
          rx="1"
          style:fill={b.tone ?? "var(--chart-1)"}
        />
        {#if b.tick != null && top > 0}
          <rect
            class="tick"
            x={Math.min(99.4, (b.tick / top) * 100)}
            y="-1"
            width="0.8"
            height="12"
          />
        {/if}
      </svg>
      <span class="value">{b.text}</span>
    </div>
  {/each}
</div>

<style>
  .bars {
    display: grid;
    gap: 8px;
  }
  .row {
    display: grid;
    grid-template-columns: minmax(6.5rem, 32%) minmax(60px, 1fr) auto;
    align-items: center;
    gap: 10px;
    font-size: 0.86rem;
  }
  .label {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    overflow-wrap: anywhere;
    color: var(--fg);
  }
  .mark {
    flex: none;
    display: inline-grid;
    place-items: center;
    width: 1.3em;
    height: 1.3em;
    border-radius: 50%;
    font: 700 0.7rem/1 var(--font-mono);
    color: #10161d;
  }
  .track {
    width: 100%;
    height: 14px;
    display: block;
  }
  .bg {
    fill: var(--overlay);
  }
  .tick {
    fill: var(--fg);
  }
  .value {
    color: var(--fg-muted);
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
    font-size: 0.8rem;
  }
  .muted .fill {
    opacity: 0.45;
  }
  .muted .label {
    color: var(--fg-muted);
  }
  @media (max-width: 819px) {
    .row {
      grid-template-columns: minmax(0, 1fr) auto;
      gap: 2px 10px;
    }
    .track {
      grid-column: 1 / -1;
      grid-row: 2;
    }
  }
</style>
