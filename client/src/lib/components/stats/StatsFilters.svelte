<script lang="ts">
  // StatsFilters — the filter bar of #/stats and a shared view (ADR 0217
  // §3 and §8): table makeup and date range, each a labelled radio group.
  // A change is announced once through a polite live region.
  import { L } from "../../labels";
  import {
    filterAnnouncement,
    MAKEUP_TEXT,
    MAKEUPS,
    RANGE_TEXT,
    RANGES,
    type Makeup,
    type StatsRange,
  } from "../../advancedStats";

  interface Props {
    makeup: Makeup;
    range: StatsRange;
    onchange: (makeup: Makeup, range: StatsRange) => void;
    disabled?: boolean;
    /** False for a page with only a date range (a pod's stats, ADR 0220 §7). */
    showMakeup?: boolean;
  }
  const { makeup, range, onchange, disabled = false, showMakeup = true }: Props = $props();

  let announcement = $state("");

  function pick(m: Makeup, r: StatsRange): void {
    if (m === makeup && r === range) return;
    announcement = filterAnnouncement(m, r);
    onchange(m, r);
  }
</script>

<div class="filters" data-testid="stats-filters">
  {#if showMakeup}
    <fieldset {disabled}>
      <legend>{L.statsMakeupFilter}</legend>
      <div class="pills">
        {#each MAKEUPS as m (m)}
          <label class="pill" class:on={m === makeup}>
            <input
              type="radio"
              name="stats-makeup"
              value={m}
              checked={m === makeup}
              onchange={() => pick(m, range)}
            />
            {MAKEUP_TEXT[m]}
          </label>
        {/each}
      </div>
    </fieldset>
  {/if}
  <fieldset {disabled}>
    <legend>{L.statsRangeFilter}</legend>
    <div class="pills">
      {#each RANGES as r (r)}
        <label class="pill" class:on={r === range}>
          <input
            type="radio"
            name="stats-range"
            value={r}
            checked={r === range}
            onchange={() => pick(makeup, r)}
          />
          {RANGE_TEXT[r]}
        </label>
      {/each}
    </div>
  </fieldset>
  <p class="visually-hidden" role="status" aria-live="polite">{announcement}</p>
</div>

<style>
  .filters {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 28px;
    margin: 0 0 16px;
  }
  fieldset {
    border: 0;
    margin: 0;
    padding: 0;
    min-width: 0;
  }
  legend {
    padding: 0;
    margin-bottom: 6px;
    font: 600 0.7rem/1.2 var(--font-mono);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .pills {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .pill {
    position: relative;
    display: inline-flex;
    align-items: center;
    min-height: 32px;
    padding: 0 12px;
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    background: var(--surface);
    color: var(--fg-muted);
    font: 600 0.85rem/1 var(--font-ui);
    cursor: pointer;
  }
  .pill:hover {
    color: var(--fg);
    background: var(--surface-hover);
  }
  .pill.on {
    border-color: var(--accent-line);
    background: var(--accent-soft);
    color: var(--fg);
  }
  .pill input {
    position: absolute;
    opacity: 0;
    inset: 0;
    margin: 0;
    cursor: pointer;
  }
  .pill:has(input:focus-visible) {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  fieldset:disabled .pill {
    opacity: 0.6;
    cursor: progress;
  }
  .visually-hidden {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
  }
</style>
