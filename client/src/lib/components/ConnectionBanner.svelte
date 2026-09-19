<script lang="ts">
  // The board admitting it is stale (#519, ADR 0044).
  //
  // Before this, a dropped socket looked exactly like the engine
  // freezing: the pre-disconnect snapshot stayed rendered and
  // interactive, clicks went nowhere, and the only tell was a coloured
  // dot in the command bar. Players read that as the board freezing
  // and reported it as one, which is the single most expensive kind of
  // bug report — a real defect filed against the wrong subsystem.
  //
  // Two deliberate non-goals:
  //
  //   * It does not blank the board. The stale snapshot surviving a
  //     reconnect is the design (ws.ts resetSnapshotTracking is NOT
  //     called on the retry path), because a board that is ten seconds
  //     old is far more useful mid-game than an empty one. The fix is
  //     admitting it is stale, not erasing it.
  //
  //   * It does not claim the engine is broken. The copy lives in
  //     connectionBanner.ts and is tested against #266's vocabulary,
  //     so a genuine freeze is still reportable as a freeze.
  //
  // Positioned `fixed` rather than in flow: this has to be unmissable
  // from wherever it is mounted, and a banner that scrolls off the top
  // of a long board is a banner that is not doing its job.
  import { connectionAnnouncement, connectionBanner } from "../connectionBanner";
  import { settings } from "../settings";
  import type { ConnectionStatus } from "../ws";

  interface Props {
    /** The GameClient `status` store's current value. */
    status: ConnectionStatus;
    /** The GameClient `reconnectAttempt` store's current value (#518). */
    attempt?: number;
    /** Dial now instead of waiting out the backoff — GameClient.retryNow. */
    onRetry?: () => void;
  }

  let { status, attempt = 0, onRetry }: Props = $props();

  const banner = $derived(connectionBanner(status, attempt));
  const announcement = $derived(connectionAnnouncement(status, attempt));

  // S11.5 animations toggle, same gate the rest of the board uses: the
  // master switch AND the accessibility preference. When either is off
  // the dot still renders in its "attention" colour — the information
  // survives, the motion does not. The @media rule below is the belt
  // for a browser-level preference set after load.
  const animate = $derived($settings.animations.enabled && !$settings.accessibility.reduceMotion);
</script>

<!--
  The live region is mounted unconditionally, outside the {#if}. A
  region that appears at the same moment its content does may not be
  announced at all, and — more importantly — recovery has to be
  announced too. The banner tells a sighted player the table is back by
  vanishing; vanishing is silent, so this is the only thing that says
  it. The banner itself carries no live role, so nothing is read twice.
-->
<div class="sr-only" role="status" aria-live="polite" aria-atomic="true">{announcement}</div>

{#if banner}
  <!-- Tone as `class:` directives rather than an interpolated class
       name, so Svelte can see both rules are reachable and does not
       prune them as unused CSS. -->
  <div
    class="connection-banner"
    class:tone-retrying={banner.tone === "retrying"}
    class:tone-lost={banner.tone === "lost"}
  >
    <span class="dot" class:animate aria-hidden="true"></span>
    <div class="copy">
      <strong class="headline">{banner.headline}</strong>
      <span class="detail">{banner.detail}</span>
    </div>
    {#if onRetry}
      <button type="button" class="retry" onclick={() => onRetry?.()}>{banner.retryLabel}</button>
    {/if}
  </div>
{/if}

<style>
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    margin: -1px;
    padding: 0;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .connection-banner {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    /* Above the board and its overlays, below the DEV badge (9999) —
       knowing you are on dev still outranks everything. */
    z-index: 900;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.55rem 1rem;
    background: var(--bg-1, #0b1220);
    color: var(--fg, #e8eef9);
    border-bottom: 2px solid var(--tone-color);
    box-shadow: var(--shadow-sm, 0 2px 8px rgba(0, 0, 0, 0.45));
    font-size: 0.86rem;
    line-height: 1.3;
  }

  .tone-retrying {
    --tone-color: var(--gold, #ffd07a);
  }

  .tone-lost {
    --tone-color: var(--danger, #ff6b6b);
  }

  .dot {
    flex: none;
    width: 0.6rem;
    height: 0.6rem;
    border-radius: 50%;
    background: var(--tone-color);
  }

  /* Gated in markup by the S11.5 animations toggle, as the command-bar
     pill is. The class is only applied when motion is allowed. */
  .dot.animate {
    animation: connection-pulse 1.2s var(--ease, ease-in-out) infinite;
  }

  @keyframes connection-pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.25;
    }
  }

  .copy {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.15rem 0.5rem;
    min-width: 0;
  }

  .headline {
    color: var(--tone-color);
    font-weight: 700;
  }

  .detail {
    opacity: 0.9;
  }

  .retry {
    flex: none;
    margin-left: auto;
    padding: 0.3rem 0.8rem;
    border: 1px solid var(--tone-color);
    border-radius: var(--radius, 8px);
    background: transparent;
    color: var(--tone-color);
    font: inherit;
    font-weight: 600;
    cursor: pointer;
  }

  .retry:hover,
  .retry:focus-visible {
    background: color-mix(in srgb, var(--tone-color) 18%, transparent);
  }

  /* Belt to the settings toggle's braces, for a preference the browser
     learns about after this component mounted. */
  @media (prefers-reduced-motion: reduce) {
    .dot.animate {
      animation: none;
    }
  }
</style>
