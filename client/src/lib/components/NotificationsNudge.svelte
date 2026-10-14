<script lang="ts">
  // The one-time nudge at the table (ADR 0162 §1): the first time a
  // signed-in player sits at a table while Discord notifications are
  // available and every switch is off, one line offers "Your turn".
  // Either answer is remembered in the synced settings, so it never
  // shows again, on this browser or any other.
  //
  // Game.svelte mounts it only for a seated player at a real table (not
  // a spectator, not the practice table). Everything else is decided
  // here: no session with a user, a server with no bot token, or a
  // switch already on, and it draws nothing.

  import { onMount } from "svelte";
  import { settings } from "../settings";
  import {
    dismissNudge,
    loadNotifications,
    notificationPrefs,
    setNotification,
    shouldNudge,
  } from "../notifications";

  let busy = $state(false);
  let error = $state("");
  // Asked once per mount, so the line never flashes up on stale data.
  let loaded = $state(false);

  const show = $derived(loaded && shouldNudge($notificationPrefs, $settings.help.seen));

  onMount(() => {
    void loadNotifications().then(() => (loaded = true));
  });

  async function turnOn(): Promise<void> {
    busy = true;
    error = "";
    try {
      await setNotification("turn", true);
      dismissNudge();
    } catch (e) {
      error = e instanceof Error ? e.message : "could not turn it on; try again";
    } finally {
      busy = false;
    }
  }
</script>

{#if show}
  <div class="dm-nudge" role="status">
    <span class="text">Get a Discord DM when it's your turn?</span>
    <button class="primary" disabled={busy} onclick={turnOn}>Turn on</button>
    <button disabled={busy} onclick={dismissNudge}>Not now</button>
    {#if error}<span class="error">{error}</span>{/if}
  </div>
{/if}

<style>
  .dm-nudge {
    position: fixed;
    top: 56px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 40;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    max-width: calc(100vw - 32px);
    padding: 8px 12px;
    border: 1px solid var(--accent-line);
    border-radius: 8px;
    background: var(--surface-raised);
    box-shadow: 0 4px 16px rgb(0 0 0 / 0.35);
    font-size: 13px;
  }
  .text {
    margin-right: 4px;
  }
  .error {
    color: var(--danger);
    flex-basis: 100%;
  }
</style>
