<script lang="ts">
  // StatTile — one number with its label and the count it stands on
  // (ADR 0217 §8), plus an optional small chart under it (a trend).
  import type { Snippet } from "svelte";

  interface Props {
    label: string;
    value: string;
    sub?: string;
    testid?: string;
    /** A value that is words (a deck name), set smaller so it wraps by word. */
    words?: boolean;
    children?: Snippet;
  }
  const { label, value, sub, testid, words = false, children }: Props = $props();
</script>

<div class="tile" data-testid={testid}>
  <span class="tile-label">{label}</span>
  <span class="tile-value" class:words>{value}</span>
  {#if sub}<span class="tile-sub">{sub}</span>{/if}
  {@render children?.()}
</div>

<style>
  .tile {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 16px 18px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    min-width: 0;
  }
  .tile-label {
    font: 600 0.72rem/1.2 var(--font-mono);
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .tile-value {
    font: 800 1.75rem/1.1 var(--font-display);
    font-variant-numeric: tabular-nums;
    overflow-wrap: break-word;
  }
  .tile-value.words {
    font-size: 1.25rem;
    line-height: 1.2;
  }
  .tile-sub {
    color: var(--fg-muted);
    font-size: 0.85rem;
  }
</style>
