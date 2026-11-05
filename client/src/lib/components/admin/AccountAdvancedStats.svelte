<script lang="ts">
  // An admin's advanced-stats switch for one person (ADR 0217 §2, owner
  // answer 1): PUT /admin/users/{id}/advanced-stats. It is stored as set
  // by an admin, and the person can still switch it off. The line beside
  // it says the state: on, off, or following the default, and who set it.
  // It reads no stats. Used by the Accounts list and one account.
  import { L } from "../../labels";
  import { LobbyApiError } from "../../session";
  import {
    accountStatsText,
    setAccountAdvancedStats,
    type AdminAccountStats,
  } from "../../statsPrefs";

  interface Props {
    name: string;
    row: AdminAccountStats;
    onchange: (row: AdminAccountStats) => void;
    onforbidden: () => void;
  }
  const { name, row, onchange, onforbidden }: Props = $props();

  let busy = $state(false);
  let error = $state("");

  async function toggle(on: boolean): Promise<void> {
    busy = true;
    error = "";
    try {
      onchange(await setAccountAdvancedStats(row.user_id, on));
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? err.message : "couldn't save";
    } finally {
      busy = false;
    }
  }
</script>

<label class="adv">
  <input
    type="checkbox"
    aria-label={L.accountAdvancedStats(name)}
    checked={row.advanced}
    disabled={busy}
    onchange={(e) => void toggle(e.currentTarget.checked)}
  />
  <span class="state" class:dim={row.choice === "default"}>{accountStatsText(row)}</span>
</label>
{#if error}
  <span class="err" role="alert">{error}</span>
{/if}

<style>
  .adv {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: 32px;
    white-space: nowrap;
  }
  .state {
    font-size: 12.5px;
  }
  .state.dim {
    color: var(--fg-muted);
  }
  .err {
    display: block;
    color: var(--danger);
    font-size: 12px;
  }
</style>
