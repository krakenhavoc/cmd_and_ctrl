<script lang="ts">
  // DeckPowerBadge is a seated deck's estimated power level (#3671), on
  // the Commander Bracket scale (1 to 5, a half step for the strong end
  // of a bracket). It never shrinks: the seat line beside it is cut off
  // with an ellipsis on a narrow seat, and the rating must stay visible.
  let { power }: { power: number } = $props();
</script>

{#if power}
  <b
    class="pwr"
    title={`Estimated power level ${power} on the Commander Bracket scale (1–5; .5 is the strong end of a bracket)`}
    aria-label={`estimated power level ${power}`}>PWR {power}</b
  >
{/if}

<style>
  .pwr {
    flex: none;
    font-family: var(--font-mono);
    font-size: 9.5px;
    font-weight: 700;
    letter-spacing: 0.06em;
    line-height: 1;
    padding: 2px 4px;
    border-radius: 3px;
    color: var(--accent-strong);
    border: 1px solid color-mix(in srgb, var(--accent) 55%, transparent);
    background: color-mix(in srgb, var(--accent) 12%, transparent);
    white-space: nowrap;
  }
</style>
