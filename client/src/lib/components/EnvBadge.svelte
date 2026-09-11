<script lang="ts">
  // A persistent marker that this is NOT production.
  //
  // It renders on every route including login, because the failure it
  // guards against is a human one: filing a bug against dev thinking
  // it was prod, or worse, assuming prod is dev and spawning a card
  // mid-game. Corner-anchored, click-through (pointer-events: none)
  // so it can never swallow a click on the board beneath it, and
  // rendered only when the server says env === "dev".
  //
  // Deliberately not dismissible. A badge you can hide is a badge
  // that is hidden exactly when it matters.
  import { isDev } from "../env";
</script>

{#if $isDev}
  <div class="env-badge" role="status" aria-label="Develop environment">
    <span class="dot" aria-hidden="true"></span>
    DEV
  </div>
{/if}

<style>
  .env-badge {
    position: fixed;
    right: 0;
    bottom: 0;
    z-index: 9999;
    display: flex;
    align-items: center;
    gap: 0.4em;
    padding: 0.28rem 0.7rem 0.28rem 0.6rem;
    border-top-left-radius: var(--radius, 8px);
    border-top: 1px solid var(--gold, #ffd07a);
    border-left: 1px solid var(--gold, #ffd07a);
    background: var(--bg-1, #0b1220);
    color: var(--gold, #ffd07a);
    font-family: var(--font-mono, ui-monospace, "SF Mono", Menlo, monospace);
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.12em;
    line-height: 1;
    box-shadow: var(--shadow-sm, 0 1px 2px rgba(0, 0, 0, 0.35));
    /* Never intercept input — this sits above the board. */
    pointer-events: none;
    user-select: none;
  }

  .dot {
    width: 0.42em;
    height: 0.42em;
    border-radius: 50%;
    background: var(--gold, #ffd07a);
  }

  /* The badge is an always-on indicator, so it must not animate. */
  @media (prefers-reduced-motion: reduce) {
    .env-badge {
      transition: none;
    }
  }
</style>
