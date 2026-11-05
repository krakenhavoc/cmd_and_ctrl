<script lang="ts">
  // Histogram — vertical bars over category labels (ADR 0217 §8: missed
  // drops by land number, game length). Each bar's count sits above it,
  // so the chart is read without the axis. A muted bar (land 6+) is
  // dimmer but still labelled; an optional marker draws
  // a labelled line at a value (the median).
  import { columns, plotBox, bandCentre, valueX } from "../../statsCharts";

  interface Props {
    labelledby: string;
    categories: string[];
    values: number[];
    /** Which bars are muted (true) rather than emphasised. */
    muted?: boolean[];
    /** A labelled vertical line at a category value. */
    marker?: { at: number; label: string; categories: number[] };
    tone?: string;
    xLabel?: string;
  }
  const {
    labelledby,
    categories,
    values,
    muted = [],
    marker,
    tone = "var(--chart-1)",
    xLabel,
  }: Props = $props();

  const W = 360;
  const H = 160;
  const plot = $derived(plotBox(W, H, { top: 30, right: 8, bottom: xLabel ? 40 : 24, left: 8 }));
  const rects = $derived(columns(plot, values));
  const mx = $derived(marker ? valueX(plot, marker.categories, marker.at) : null);
</script>

<div class="histogram">
  <svg viewBox="0 0 {W} {H}" role="img" aria-labelledby={labelledby}>
    <line class="base" x1={plot.left} x2={plot.right} y1={plot.bottom} y2={plot.bottom} />
    {#each rects as r, i (i)}
      <rect
        x={r.x}
        y={r.y}
        width={r.width}
        height={r.height}
        rx="2"
        style:fill={tone}
        class:muted={muted[i]}
      />
      {#if r.value > 0}
        <text class="count" x={r.x + r.width / 2} y={r.y - 4} text-anchor="middle">{r.value}</text>
      {/if}
      <text
        class="axis"
        class:muted={muted[i]}
        x={bandCentre(plot, i, values.length)}
        y={plot.bottom + 16}
        text-anchor="middle">{categories[i]}</text
      >
    {/each}
    {#if mx != null && marker}
      <line class="marker" x1={mx} x2={mx} y1={14} y2={plot.bottom} />
      <text class="marker-label" x={mx + 4} y={11}>{marker.label}</text>
    {/if}
    {#if xLabel}
      <text class="axis name" x={(plot.left + plot.right) / 2} y={H - 4} text-anchor="middle"
        >{xLabel}</text
      >
    {/if}
  </svg>
</div>

<style>
  .histogram {
    max-width: 480px;
  }
  svg {
    width: 100%;
    height: auto;
    display: block;
    overflow: visible;
  }
  .base {
    stroke: var(--border-strong);
  }
  rect.muted {
    opacity: 0.4;
  }
  .count {
    fill: var(--fg);
    font: 600 12px var(--font-mono);
  }
  .axis {
    fill: var(--fg-muted);
    font: 500 12px var(--font-mono);
  }
  .axis.muted {
    opacity: 0.8;
  }
  .axis.name {
    font-family: var(--font-ui);
  }
  .marker {
    stroke: var(--fg);
    stroke-width: 1.25;
    stroke-dasharray: 3 3;
  }
  .marker-label {
    fill: var(--fg);
    font: 600 12px var(--font-ui);
  }
</style>
