<script lang="ts">
  // The Settings "Stats" tab (ADR 0217 §2): one switch for advanced
  // stats, off unless the person, an admin or the server's default turned
  // it on. Settings shows the tab only to a signed-in person on a server
  // that can keep the setting (statsPrefs.available), so this never draws
  // a switch that cannot work. Guests get no stats.

  import { L } from "../labels";
  import PodShareSwitch from "./PodShareSwitch.svelte";
  import { setAdvancedStats, statsPrefs, statsPrefsNote } from "../statsPrefs";

  let busy = $state(false);
  let error = $state("");

  async function toggle(on: boolean): Promise<void> {
    busy = true;
    error = "";
    try {
      await setAdvancedStats(on);
    } catch (e) {
      error = e instanceof Error ? e.message : "could not save; try again";
    } finally {
      busy = false;
    }
  }
</script>

<h3>Stats</h3>
<p class="help">
  Your Stats page always shows your record, by deck and by commander. Advanced stats add more about
  how your games and decks play. They are free, and only you see them.
</p>

<label class="switch-row">
  <input
    type="checkbox"
    checked={$statsPrefs.advanced}
    disabled={busy}
    onchange={(e) => toggle(e.currentTarget.checked)}
  />
  <span>
    {L.advancedStatsSwitch}
    <span class="help" data-testid="advanced-stats-note">{statsPrefsNote($statsPrefs)}</span>
  </span>
</label>

{#if error}
  <p class="error" role="alert">{error}</p>
{/if}

<PodShareSwitch />

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
    color: var(--fg-muted);
  }
  .error {
    color: var(--danger);
  }
</style>
