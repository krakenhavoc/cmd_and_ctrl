<script lang="ts">
  // The server's advanced-stats default (ADR 0217 owner answer 1): PUT
  // /admin/stats/advanced-default. It reaches everyone whose switch is
  // unset; anyone who chose for themselves, on or off, keeps their
  // answer. Until an admin stores one it is off (gamestats.AdvancedDefault).
  import { L } from "../../labels";
  import { LobbyApiError } from "../../session";
  import { setAdminAdvancedDefault, type AdminStatsDefault } from "../../statsPrefs";

  interface Props {
    value: AdminStatsDefault;
    onchange: (value: AdminStatsDefault) => void;
    onforbidden: () => void;
  }
  const { value, onchange, onforbidden }: Props = $props();

  let busy = $state(false);
  let error = $state("");

  async function toggle(on: boolean): Promise<void> {
    busy = true;
    error = "";
    try {
      onchange(await setAdminAdvancedDefault(on));
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? err.message : "couldn't save the default";
    } finally {
      busy = false;
    }
  }
</script>

<div class="card default" data-testid="advanced-stats-default">
  <label class="row">
    <input
      type="checkbox"
      checked={value.on}
      disabled={busy}
      onchange={(e) => void toggle(e.currentTarget.checked)}
    />
    <span>
      <span class="name">{L.advancedStatsDefault}</span>
      <span class="help">
        {value.on ? "On" : "Off"}{value.stored
          ? ""
          : " (never set; off until an admin turns it on)"}. Anyone who switched advanced stats on
        or off for themselves keeps their answer.
      </span>
    </span>
  </label>
  {#if error}
    <p class="err" role="alert">{error}</p>
  {/if}
</div>

<style>
  .default {
    padding: 12px 14px;
  }
  .row {
    display: flex;
    gap: 10px;
    align-items: flex-start;
  }
  .row input {
    margin-top: 3px;
  }
  .name {
    font-weight: 600;
  }
  .help {
    display: block;
    color: var(--fg-muted);
    font-size: 12.5px;
    line-height: 1.45;
  }
  .err {
    margin: 6px 0 0;
    color: var(--danger);
    font-size: 12.5px;
  }
</style>
