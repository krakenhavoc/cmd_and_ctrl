<script lang="ts">
  // CurvePanels — the mana curve over the land-drop curve, two panels in
  // one SVG on one x axis (ADR 0217 §4 and §8): x is a number 0 … 7+, the
  // top panel counts the deck's non-land cards at that mana value, the
  // bottom one draws the share of games that had hit land drop N by own
  // turn N (solid) and had N lands controlled by then, ramp included
  // (dashed, dimmer). A deck whose 6- and 7-drops sit over a falling line
  // shows the problem without arithmetic. One y axis per panel, never a
  // dual axis. Every point is labelled by its position and a <title>.
  import { bandCentre, columns, linePath, linePoints, plotBox, ticks } from "../../statsCharts";
  import { pctText, type LandCurvePoint } from "../../advancedStats";

  interface Props {
    labelledby: string;
    /** Non-land cards at mana value 0 … 6 and 7+ (8 numbers). */
    curve: number[];
    /** The land-drop reliability at N = 1 … 7 (7 is 7 and up). */
    points: LandCurvePoint[];
  }
  const { labelledby, curve, points }: Props = $props();

  const W = 360;
  const H = 336;
  const LEFT = 40;
  const RIGHT = 10;
  const top = plotBox(W, H, { top: 26, right: RIGHT, bottom: H - 130, left: LEFT });
  const low = plotBox(W, H, { top: 186, right: RIGHT, bottom: 56, left: LEFT });

  const cats = $derived(curve.map((_, i) => (i >= curve.length - 1 ? `${i}+` : `${i}`)));
  const bars = $derived(columns(top, curve));
  // The land curve starts at N = 1, so category 0 has no point: a gap.
  const hit = $derived<(number | null)[]>([null, ...points.map((p) => p.hit.rate)]);
  const ctl = $derived<(number | null)[]>([null, ...points.map((p) => p.controlled.rate)]);
  const yTicks = ticks(1, 2);
  const yOf = (v: number): number => low.bottom - v * (low.bottom - low.top);
  const n = $derived(cats.length);
</script>

<div class="curves">
  <svg viewBox="0 0 {W} {H}" role="img" aria-labelledby={labelledby}>
    <text class="panel" x={LEFT} y="12">Non-land cards at each mana value</text>
    <line class="base" x1={top.left} x2={top.right} y1={top.bottom} y2={top.bottom} />
    {#each bars as b, i (i)}
      <rect x={b.x} y={b.y} width={b.width} height={b.height} rx="2" style:fill="var(--chart-1)" />
      {#if b.value > 0}
        <text class="count" x={b.x + b.width / 2} y={b.y - 4} text-anchor="middle">{b.value}</text>
      {/if}
    {/each}

    <text class="panel" x={LEFT} y="172">Games with land drop N hit by turn N</text>
    {#each yTicks as t (t)}
      <line class="grid" x1={low.left} x2={low.right} y1={yOf(t)} y2={yOf(t)} />
      <text class="axis" x={low.left - 6} y={yOf(t) + 3.5} text-anchor="end">{pctText(t)}</text>
    {/each}
    <path class="series dim" d={linePath(low, ctl, 1)} style:stroke="var(--chart-3)" />
    {#each linePoints(low, ctl, 1) as p (p.index)}
      <circle class="dim" cx={p.x} cy={p.y} r="2.5" style:fill="var(--chart-3)">
        <title>Lands controlled by turn {cats[p.index]}: {pctText(p.value)}</title>
      </circle>
    {/each}
    <path class="series" d={linePath(low, hit, 1)} style:stroke="var(--chart-2)" />
    {#each linePoints(low, hit, 1) as p (p.index)}
      <circle cx={p.x} cy={p.y} r="3.2" style:fill="var(--chart-2)">
        <title>Land drop {cats[p.index]} by turn {cats[p.index]}: {pctText(p.value)}</title>
      </circle>
    {/each}

    {#each cats as c, i (i)}
      <text class="axis" x={bandCentre(low, i, n)} y={low.bottom + 16} text-anchor="middle"
        >{c}</text
      >
    {/each}
    <text class="axis name" x={(low.left + low.right) / 2} y={H - 28} text-anchor="middle"
      >Mana value, and land drop number</text
    >
  </svg>
  <ul class="legend" aria-hidden="true">
    <li><span class="swatch" style:border-color="var(--chart-2)"></span>Land drop hit</li>
    <li>
      <span class="swatch dim" style:border-color="var(--chart-3)"></span>Lands controlled (ramp
      counts)
    </li>
  </ul>
</div>

<style>
  .curves {
    max-width: 480px;
  }
  svg {
    width: 100%;
    height: auto;
    display: block;
    overflow: visible;
  }
  .base,
  .grid {
    stroke: var(--border);
    stroke-width: 1;
  }
  .axis,
  .count,
  .panel {
    fill: var(--fg-muted);
    font: 500 12px var(--font-mono);
  }
  .count {
    fill: var(--fg);
  }
  .panel {
    font-family: var(--font-ui);
    font-weight: 600;
    font-size: 12px;
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
    margin: 4px 0 0;
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
