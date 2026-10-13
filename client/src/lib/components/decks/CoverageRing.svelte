<script lang="ts">
  // CoverageRing — how much of a saved deck plays as printed, as a ring
  // and "n/total" (ADR 0158 §2). Green at 96 % and up, amber below. The
  // sentence the list used to print is the ring's accessible name.
  import type { DeckRing } from "../../deckGrid";

  let { ring }: { ring: DeckRing } = $props();

  const R = 14;
  const C = 2 * Math.PI * R;
  const label = $derived(`${ring.n} of ${ring.total} cards play as printed`);
</script>

<span class="ring {ring.tone}" role="img" aria-label={label} title={label}>
  <svg viewBox="0 0 36 36" width="34" height="34" aria-hidden="true">
    <circle class="track" cx="18" cy="18" r={R} />
    <circle
      class="arc"
      cx="18"
      cy="18"
      r={R}
      stroke-dasharray="{ring.share * C} {C}"
      transform="rotate(-90 18 18)"
    />
  </svg>
  <span class="num" aria-hidden="true"><b>{ring.n}</b>/{ring.total}</span>
</span>

<style>
  .ring {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    flex: 0 0 auto;
  }
  circle {
    fill: none;
    stroke-width: 3.5;
  }
  .track {
    stroke: var(--border-strong);
  }
  .arc {
    stroke-linecap: round;
  }
  .good .arc {
    stroke: var(--mint);
  }
  .warn .arc {
    stroke: var(--priority);
  }
  .num {
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--fg-dim);
    white-space: nowrap;
  }
  .num b {
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 800;
    color: var(--fg);
  }
</style>
