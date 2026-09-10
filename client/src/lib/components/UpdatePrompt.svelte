<script lang="ts">
  // UpdatePrompt — the only way a new client build ever takes over.
  //
  // A service worker that called skipWaiting() on install would reload the
  // page under whoever happened to be holding priority. So the new build
  // sits in `waiting` and this toast appears instead: non-modal, ignorable,
  // and anchored out of the way of the board. The player reloads between
  // turns, or dismisses and gets asked again at the next update check.
  //
  // See lib/pwa.ts and docs/decisions/0031-progressive-web-app.md.

  import Icon from "./Icon.svelte";
  import { updateReady, applyUpdate, dismissUpdate } from "../pwa";
</script>

{#if $updateReady}
  <div class="update" role="status" aria-live="polite">
    <span class="ico" aria-hidden="true"><Icon name="spark" size={14} /></span>
    <div class="copy">
      <strong>New version available</strong>
      <span>Reload when you're not mid-turn.</span>
    </div>
    <button type="button" class="primary" onclick={applyUpdate}>Reload</button>
    <button type="button" class="dismiss" aria-label="dismiss" onclick={dismissUpdate}>
      <Icon name="x" size={13} />
    </button>
  </div>
{/if}

<style>
  .update {
    position: fixed;
    left: max(1rem, env(safe-area-inset-left));
    bottom: max(1rem, env(safe-area-inset-bottom));
    z-index: 900;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    max-width: min(24rem, calc(100vw - 2rem));
    padding: 0.6rem 0.7rem 0.6rem 0.85rem;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    background: var(--surface-raised);
    box-shadow: var(--shadow-lg);
    color: var(--fg);
  }
  .ico {
    display: inline-flex;
    color: var(--gold);
  }
  .copy {
    display: flex;
    flex-direction: column;
    line-height: 1.3;
  }
  .copy strong {
    font-size: 0.82rem;
    font-weight: 600;
  }
  .copy span {
    font-size: 0.72rem;
    color: var(--fg-muted);
  }
  .update button {
    padding: 0.35rem 0.7rem;
    font-size: 0.75rem;
  }
  .update button.primary {
    background: var(--accent);
    color: var(--accent-fg);
    border-color: var(--accent);
  }
  .update button.dismiss {
    padding: 0.35rem;
    background: transparent;
    border-color: transparent;
    color: var(--fg-dim);
  }
  .update button.dismiss:hover {
    color: var(--fg);
  }
</style>
