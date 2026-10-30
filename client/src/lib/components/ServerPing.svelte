<script lang="ts">
  // ServerPing — the site's live signal: a dot and "Online · 38 ms",
  // this browser's round trip to the server (lib/serverPing.ts). Hidden
  // until the first answer, so a page never shows a guess. Not a live
  // region: re-announcing a ping every 30 seconds would be noise.
  import { serverPing, pingLabel } from "../serverPing";

  const { compact = false }: { compact?: boolean } = $props();
  const label = $derived(pingLabel($serverPing));
</script>

{#if label}
  <span
    class="ping {$serverPing.state}"
    class:compact
    title="Your round trip to the CMD & CTRL server, checked every 30 seconds"
  >
    <span class="dot" aria-hidden="true"></span>
    <span class="text">{label}</span>
  </span>
{/if}

<style>
  .ping {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 30px;
    padding: 0 12px;
    border-radius: 999px;
    background: var(--overlay);
    color: var(--fg-muted);
    font: 600 13px/1 var(--font-ui);
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }
  .ping.compact {
    height: auto;
    padding: 0;
    background: none;
  }
  .dot {
    position: relative;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--fg-dim);
  }
  .online .dot {
    background: var(--mint);
  }
  .offline .dot {
    background: var(--danger);
  }
  .offline {
    color: var(--danger);
  }
  /* A slow ring off the dot while the server answers: the one thing on
     the page that says it is live. Off under reduced motion. */
  .online .dot::after {
    content: "";
    position: absolute;
    inset: 0;
    border-radius: 50%;
    background: var(--mint);
    opacity: 0;
  }
  @media (prefers-reduced-motion: no-preference) {
    :global(:root:not([data-reduce-motion="1"])) .online .dot::after {
      animation: ping-ring 2.4s var(--ease) infinite;
    }
  }
  @keyframes ping-ring {
    0% {
      transform: scale(1);
      opacity: 0.55;
    }
    70%,
    100% {
      transform: scale(2.6);
      opacity: 0;
    }
  }
</style>
