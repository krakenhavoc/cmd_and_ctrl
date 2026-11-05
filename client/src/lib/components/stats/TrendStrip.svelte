<script lang="ts">
  // TrendStrip — the last games, oldest on the left, one cell each (ADR
  // 0217 §8: the screw, flood and colour-screw tiles' small trend). A game
  // with the trouble is a full cell; one without is a short muted tick, so
  // height carries it as well as colour.
  interface Props {
    /** Newest first, as the wire sends them. */
    marks: boolean[];
    label: string;
    tone: string;
  }
  const { marks, label, tone }: Props = $props();
  const cells = $derived([...marks].reverse());
</script>

{#if cells.length > 0}
  <svg class="trend" viewBox="0 0 100 16" preserveAspectRatio="none" role="img" aria-label={label}>
    {#each cells as on, i (i)}
      <rect
        x={(i * 100) / 20 + 0.6}
        y={on ? 0 : 12}
        width={100 / 20 - 1.2}
        height={on ? 16 : 4}
        style:fill={on ? tone : "var(--fg-muted)"}
        opacity={on ? 1 : 0.45}
      />
    {/each}
  </svg>
{/if}

<style>
  .trend {
    width: 100%;
    max-width: 180px;
    height: 16px;
    display: block;
    margin-top: 6px;
  }
</style>
