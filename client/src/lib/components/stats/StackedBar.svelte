<script lang="ts">
  // StackedBar — one 100% bar split into labelled segments (ADR 0217 §8:
  // place distribution, the spread across opponents). Each segment
  // carries its name and share in a legend line under the bar, so a
  // narrow segment is still read, and colour is never the only cue.
  import { stackSegments } from "../../statsCharts";
  import { pctText } from "../../advancedStats";

  interface Props {
    /** The bar's own name, read before its legend ("4 players, 16 games"). */
    label: string;
    parts: number[];
    names: string[];
    tones: string[];
  }
  const { label, parts, names, tones }: Props = $props();
  const segs = $derived(stackSegments(parts));
</script>

<div class="stacked">
  <span class="label">{label}</span>
  <svg viewBox="0 0 100 10" preserveAspectRatio="none" aria-hidden="true">
    <rect class="bg" x="0" y="0" width="100" height="10" rx="1" />
    {#each segs as s (s.index)}
      <rect
        x={s.start * 100}
        y="0"
        width={s.share * 100}
        height="10"
        style:fill={tones[s.index % tones.length]}
        class="seg"
      />
    {/each}
  </svg>
  <ul class="legend">
    {#each segs as s (s.index)}
      <li>
        <span class="swatch" style:background={tones[s.index % tones.length]}></span>
        {names[s.index]}
        <span class="pct">{pctText(s.share)}</span>
      </li>
    {/each}
  </ul>
</div>

<style>
  .stacked {
    display: grid;
    gap: 6px;
  }
  .label {
    font-size: 0.82rem;
    color: var(--fg-muted);
  }
  svg {
    width: 100%;
    height: 18px;
    display: block;
  }
  .bg {
    fill: var(--overlay);
  }
  .seg {
    stroke: var(--surface);
    stroke-width: 0.4;
    vector-effect: non-scaling-stroke;
  }
  .legend {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 4px 14px;
    font-size: 0.82rem;
  }
  .legend li {
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .swatch {
    width: 10px;
    height: 10px;
    border-radius: 2px;
  }
  .pct {
    color: var(--fg-muted);
    font-variant-numeric: tabular-nums;
  }
</style>
