<script lang="ts">
  // LineChart — one or two series over category x labels (ADR 0217 §8:
  // aggression over turns). Inline SVG on a fixed viewBox that scales
  // with its box, capped so its labels never grow past a readable size.
  // A null value is a gap, never a zero. Each series is named in a legend
  // line with its swatch and its line style, so colour is not the only
  // cue; points carry their values in a <title>.
  import { linePath, linePoints, plotBox, bandCentre, ticks } from "../../statsCharts";

  interface Series {
    name: string;
    values: (number | null)[];
    tone: string;
    /** Drawn dashed and dimmer: the secondary line. */
    dim?: boolean;
  }
  interface Props {
    labelledby: string;
    categories: string[];
    series: Series[];
    /** The y axis's top; the series' max when unset. */
    max?: number;
    format?: (v: number) => string;
    /** The x axis's name, under it. */
    xLabel?: string;
  }
  const {
    labelledby,
    categories,
    series,
    max,
    format = (v: number) => (Math.round(v * 10) / 10).toString(),
    xLabel,
  }: Props = $props();

  const W = 360;
  const H = 170;
  const plot = $derived(plotBox(W, H, { top: 10, right: 10, bottom: xLabel ? 42 : 26, left: 38 }));
  const dataMax = $derived(
    Math.max(0, ...series.flatMap((s) => s.values.filter((v): v is number => v != null))),
  );
  const yTicks = $derived(ticks(max ?? dataMax, 4));
  const top = $derived(yTicks[yTicks.length - 1]);
  const yOf = (v: number): number => plot.bottom - (v / (top || 1)) * (plot.bottom - plot.top);
</script>

<div class="line-chart">
  <svg viewBox="0 0 {W} {H}" role="img" aria-labelledby={labelledby}>
    {#each yTicks as t (t)}
      <line class="grid" x1={plot.left} x2={plot.right} y1={yOf(t)} y2={yOf(t)} />
      <text class="axis" x={plot.left - 6} y={yOf(t) + 3.5} text-anchor="end">{format(t)}</text>
    {/each}
    {#each categories as c, i (i)}
      <text
        class="axis"
        x={bandCentre(plot, i, categories.length)}
        y={plot.bottom + 16}
        text-anchor="middle">{c}</text
      >
    {/each}
    {#if xLabel}
      <text class="axis name" x={(plot.left + plot.right) / 2} y={H - 4} text-anchor="middle"
        >{xLabel}</text
      >
    {/if}
    {#each series as s (s.name)}
      <path
        class="series"
        class:dim={s.dim}
        d={linePath(plot, s.values, top)}
        style:stroke={s.tone}
      />
      {#each linePoints(plot, s.values, top) as p (p.index)}
        <circle cx={p.x} cy={p.y} r={s.dim ? 2.5 : 3.2} style:fill={s.tone} class:dim={s.dim}>
          <title>{s.name}, {categories[p.index]}: {format(p.value)}</title>
        </circle>
      {/each}
    {/each}
  </svg>
  {#if series.length > 1}
    <ul class="legend" aria-hidden="true">
      {#each series as s (s.name)}
        <li><span class="swatch" class:dim={s.dim} style:border-color={s.tone}></span>{s.name}</li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .line-chart {
    max-width: 480px;
  }
  svg {
    width: 100%;
    height: auto;
    display: block;
    overflow: visible;
  }
  .grid {
    stroke: var(--border);
    stroke-width: 1;
  }
  .axis {
    fill: var(--fg-muted);
    font: 500 12px var(--font-mono);
  }
  .axis.name {
    font-family: var(--font-ui);
  }
  .series {
    fill: none;
    stroke-width: 2.25;
    stroke-linejoin: round;
    stroke-linecap: round;
  }
  .series.dim {
    stroke-width: 1.75;
    stroke-dasharray: 5 4;
    opacity: 0.75;
  }
  circle.dim {
    opacity: 0.75;
  }
  .legend {
    list-style: none;
    margin: 6px 0 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 4px 16px;
    font-size: 0.82rem;
    color: var(--fg-muted);
  }
  .legend li {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .swatch {
    width: 18px;
    height: 0;
    border-top: 2.5px solid;
  }
  .swatch.dim {
    border-top-style: dashed;
    opacity: 0.75;
  }
</style>
