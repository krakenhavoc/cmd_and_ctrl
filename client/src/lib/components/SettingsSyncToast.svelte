<script lang="ts">
  // SettingsSyncToast — ADR 0110 §4 owner answer 6. A sign-in applied
  // the account's settings over values this browser had. The account's
  // copy wins; this offers the other choice for the rest of the visit.
  // Non-modal, like UpdatePrompt, and stacked above it when both show.

  import Icon from "./Icon.svelte";
  import { updateReady } from "../pwa";
  import {
    dismissSettingsSyncToast,
    keepBrowserSettings,
    settingsSyncToast,
  } from "../settingsSync";
</script>

{#if $settingsSyncToast}
  <div class="sync" class:raised={$updateReady} role="status" aria-live="polite">
    <span class="ico" aria-hidden="true"><Icon name="spark" size={14} /></span>
    <div class="copy">
      <strong>Using your account's settings</strong>
      <span>This browser had different ones.</span>
    </div>
    <button type="button" class="primary" onclick={keepBrowserSettings}>
      Keep this browser's instead
    </button>
    <button
      type="button"
      class="dismiss"
      aria-label="keep the account's settings"
      onclick={dismissSettingsSyncToast}
    >
      <Icon name="x" size={13} />
    </button>
  </div>
{/if}

<style>
  .sync {
    position: fixed;
    left: max(1rem, env(safe-area-inset-left));
    bottom: max(1rem, env(safe-area-inset-bottom));
    z-index: 900;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    max-width: min(28rem, calc(100vw - 2rem));
    padding: 0.6rem 0.7rem 0.6rem 0.85rem;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    background: var(--surface-raised);
    box-shadow: var(--shadow-lg);
    color: var(--fg);
  }
  .sync.raised {
    bottom: calc(max(1rem, env(safe-area-inset-bottom)) + 4.25rem);
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
  .sync button {
    padding: 0.35rem 0.7rem;
    font-size: 0.75rem;
  }
  .sync button.primary {
    background: var(--accent);
    color: var(--accent-fg);
    border-color: var(--accent);
  }
  .sync button.dismiss {
    padding: 0.35rem;
    background: transparent;
    border-color: transparent;
    color: var(--fg-dim);
  }
  .sync button.dismiss:hover {
    color: var(--fg);
  }
</style>
