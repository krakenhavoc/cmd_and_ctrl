<script lang="ts">
  // RelTime: a time as "4 min ago", with the exact local time in the
  // tooltip (ADR 0124 §7). `now` is the answer's generated_at, so a
  // clock skew between this browser and the server never shows a time
  // in the future.
  import { exactTime, isoTime, relativeTime } from "../../adminViews";

  interface Props {
    ms: number | undefined;
    now: number;
    /** shown before the relative time, e.g. "since" */
    prefix?: string;
  }
  const { ms, now, prefix = "" }: Props = $props();
</script>

{#if ms}
  <time datetime={isoTime(ms)} title={exactTime(ms)}
    >{prefix ? `${prefix} ` : ""}{relativeTime(ms, now)}</time
  >
{:else}
  <span class="dim">—</span>
{/if}

<style>
  time {
    white-space: nowrap;
  }
  .dim {
    color: var(--fg-dim);
  }
</style>
