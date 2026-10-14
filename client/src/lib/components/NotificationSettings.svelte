<script lang="ts">
  // The Settings "Notifications" tab (ADR 0162 §1): three switches for
  // Discord direct messages, every one off until the person turns it
  // on. Settings shows the tab only when the server says the feature is
  // available here (a signed-in person, on a server with the bot
  // token), so this never draws a form that cannot work.

  import {
    DM_FAILED_HINT,
    NOTIFICATION_SWITCHES,
    notificationPrefs,
    setNotification,
    type NotificationKind,
  } from "../notifications";

  let busy = $state(false);
  let error = $state("");

  async function toggle(kind: NotificationKind, on: boolean): Promise<void> {
    busy = true;
    error = "";
    try {
      await setNotification(kind, on);
    } catch (e) {
      error = e instanceof Error ? e.message : "could not save; try again";
    } finally {
      busy = false;
    }
  }
</script>

<h3>Notifications</h3>
<p class="help">
  CMD &amp; CTRL can send you a Discord direct message when a game needs you and you are away from
  it: no tab open on the table, or the tab is hidden. At most one message per table every ten
  minutes.
</p>

{#if $notificationPrefs.dm_failed}
  <p class="refused" role="status">{DM_FAILED_HINT}</p>
{/if}

{#each NOTIFICATION_SWITCHES as sw (sw.kind)}
  <label class="switch-row">
    <input
      type="checkbox"
      checked={$notificationPrefs[sw.kind]}
      disabled={busy}
      onchange={(e) => toggle(sw.kind, e.currentTarget.checked)}
    />
    <span>
      {sw.label}
      <span class="help">{sw.help}</span>
    </span>
  </label>
{/each}

{#if error}
  <p class="error" role="alert">{error}</p>
{/if}

<style>
  .help {
    color: var(--fg-dim);
    font-size: 11.5px;
    line-height: 1.45;
    margin: 4px 0 8px;
  }
  .switch-row {
    display: flex;
    gap: 0.6rem;
    align-items: flex-start;
    margin: 0.6rem 0;
  }
  .switch-row .help {
    display: block;
    margin: 2px 0 0;
  }
  .refused {
    padding: 0.5rem 0.7rem;
    border-left: 3px solid var(--danger-line, var(--danger));
    background: var(--danger-soft);
    font-size: 12px;
  }
  .error {
    color: var(--danger);
  }
</style>
