<script lang="ts">
  // DraftCurve — the mana curve of Main (#5702), lands left out: a bar
  // for each mana value from 0 to 6 and more, and the average mana value.
  // The same chart is the compact one under the picks and the larger one
  // in the full stats (#5848 §6, `large`: the count sits on each bar).
  //
  // The chart is for the eye; a screen reader gets the whole curve as one
  // sentence (lib/draftStats.ts curveSummary), and every bar has its
  // number beside it.
  import type { DraftCard } from "../../draft";
  import { curveStats } from "../../draftDeckStats";
  import { curveSummary } from "../../draftStats";
  import { L } from "../../labels";

  interface Props {
    main: DraftCard[];
    /** The overlay's taller chart, with the count on each bar. */
    large?: boolean;
  }
  const { main, large = false }: Props = $props();

  const stats = $derived(curveStats(main));
  const tallest = $derived(Math.max(1, ...stats.columns.map((c) => c.count)));
</script>

<div class="curvebox" class:large data-testid="draft-curve">
  <div class="curve" role="img" aria-label={`${L.draftCurve}: ${curveSummary(stats.columns)}`}>
    {#each stats.columns as col (col.value)}
      <div class="col" aria-hidden="true">
        <span class="count">{col.count}</span>
        <span class="bar" style:height={`${Math.round((col.count / tallest) * 100)}%`}></span>
        <span class="mv">{col.label}</span>
      </div>
    {/each}
  </div>
  <p class="avg" data-testid="draft-curve-average">
    {L.draftStatsAverageMv(stats.average.toFixed(2))}
  </p>
</div>

<style>
  .curvebox {
    --h: 64px;
    min-width: 0;
  }
  .curvebox.large {
    --h: 140px;
  }
  .curve {
    display: grid;
    grid-template-columns: repeat(7, minmax(0, 1fr));
    gap: 6px;
    height: var(--h);
  }
  .col {
    display: grid;
    grid-template-rows: auto minmax(0, 1fr) auto;
    justify-items: center;
    gap: 3px;
    min-height: 0;
  }
  .count,
  .mv {
    font-family: var(--font-mono);
    font-size: 10px;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .large .count,
  .large .mv {
    font-size: 12px;
  }
  .bar {
    align-self: end;
    width: 100%;
    min-height: 2px;
    border-radius: var(--radius-sm) var(--radius-sm) 0 0;
    background: var(--accent);
  }
  .avg {
    margin: 6px 0 0;
    font-size: 12px;
    color: var(--fg-muted);
  }
</style>
