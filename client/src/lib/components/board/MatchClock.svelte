<script lang="ts">
  // MatchClock — the time left between two games of a match (ADR 0222
  // §3), counted from when the frame that carried the deadline arrived,
  // so a device whose clock is wrong still shows the server's time.
  import { clockText, secondsLeft } from "../../match";
  import type { SideboardingView } from "../../protocol";

  const { sideboarding, label }: { sideboarding: SideboardingView; label: string } = $props();

  let receivedAt = $state(Date.now());
  let now = $state(Date.now());
  $effect(() => {
    void sideboarding.deadline;
    void sideboarding.seconds_left;
    receivedAt = Date.now();
    now = receivedAt;
  });
  $effect(() => {
    const t = setInterval(() => (now = Date.now()), 500);
    return () => clearInterval(t);
  });
  const left = $derived(secondsLeft(sideboarding, receivedAt, now));
</script>

<span class="clock" class:low={left <= 30} role="timer" aria-label={label}>{clockText(left)}</span>

<style>
  .clock {
    flex: none;
    padding: 2px 10px;
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    color: var(--fg);
    font-family: var(--font-mono);
    font-size: 14px;
    font-variant-numeric: tabular-nums;
  }
  .clock.low {
    border-color: var(--danger-line);
    color: var(--danger);
  }
</style>
