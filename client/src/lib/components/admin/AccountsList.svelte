<script lang="ts">
  // Accounts (ADR 0124 §3.1, §7): every account, one row each, with
  // first seen, last sign-in, games played and last played. `played`
  // (#/admin/accounts?played=7d, the Accounts-that-played tile's link)
  // is the server's filter and keeps exactly the accounts the tile
  // counts. `sort` is in the hash too but never sent: the whole list is
  // on the page, so it is sorted here, and the search box filters it as
  // you type. Loads once; Refresh asks again.
  //
  // Advanced stats (ADR 0217 owner answer 1): the server's default above
  // the table, and an Advanced stats column with each person's switch and
  // its state (on, off, or following the default) and who set it, from
  // GET /admin/stats/advanced. Neither reads anyone's stats.
  import { onMount } from "svelte";
  import { fetchAdminAccounts } from "../../api";
  import { L } from "../../labels";
  import {
    accountsHash,
    filterAccounts,
    sortAccounts,
    type AccountSort,
    type AccountsFilter,
    type AdminAccountsResponse,
    type PlayedWindow,
  } from "../../adminViews";
  import { navigate } from "../../router";
  import { LobbyApiError } from "../../session";
  import AccountLink from "./AccountLink.svelte";
  import AccountAdvancedStats from "./AccountAdvancedStats.svelte";
  import AdvancedStatsDefault from "./AdvancedStatsDefault.svelte";
  import {
    accountStatsOf,
    fetchAdminAdvancedStats,
    withAccount,
    withDefault,
    type AdminAdvancedStats,
  } from "../../statsPrefs";
  import RelTime from "./RelTime.svelte";

  interface Props {
    filter: AccountsFilter;
    onforbidden: () => void;
  }
  const { filter, onforbidden }: Props = $props();

  let data = $state<AdminAccountsResponse | null>(null);
  let loading = $state(false);
  let error = $state("");
  let query = $state("");
  // null until loaded, and left null when this server keeps no setting
  // (503 with no database): the column and the default are not drawn.
  let adv = $state<AdminAdvancedStats | null>(null);

  async function loadAdvanced(): Promise<void> {
    try {
      adv = await fetchAdminAdvancedStats();
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) onforbidden();
      adv = null;
    }
  }

  async function load(): Promise<void> {
    loading = true;
    error = "";
    void loadAdvanced();
    try {
      data = await fetchAdminAccounts(filter);
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? err.message : "couldn't load the accounts";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
  });

  function setFilter(patch: Partial<AccountsFilter>): void {
    navigate(accountsHash({ ...filter, ...patch }));
  }

  function pick<T extends string>(e: Event): T | undefined {
    const v = (e.currentTarget as HTMLSelectElement).value;
    return v === "" ? undefined : (v as T);
  }

  const rows = $derived(
    data ? filterAccounts(sortAccounts(data.accounts, filter.sort), query) : [],
  );
  const now = $derived(data?.generated_at ?? Date.now());
</script>

<div class="view">
  <div class="filters" role="group" aria-label="filter accounts">
    <label>
      <span>Played</span>
      <select
        value={filter.played ?? ""}
        onchange={(e) => setFilter({ played: pick<PlayedWindow>(e) })}
      >
        <option value="">any time, or never</option>
        <option value="1d">in the last day</option>
        <option value="7d">in the last 7 days</option>
        <option value="30d">in the last 30 days</option>
      </select>
    </label>
    <label>
      <span>Sort</span>
      <select value={filter.sort ?? ""} onchange={(e) => setFilter({ sort: pick<AccountSort>(e) })}>
        <option value="">playing now, then recent</option>
        <option value="last_played">last played</option>
        <option value="first_seen">newest accounts</option>
        <option value="name">name</option>
        <option value="games">games played</option>
      </select>
    </label>
    <label class="search">
      <span>Find</span>
      <input type="search" placeholder="name" bind:value={query} aria-label="find an account" />
    </label>
    <button type="button" class="ghost sm" disabled={loading} onclick={() => void load()}>
      {loading ? "Refreshing…" : "Refresh"}
    </button>
  </div>

  {#if error}
    <p class="notice err" role="alert">{error}</p>
  {/if}

  {#if adv}
    <AdvancedStatsDefault
      value={adv.default}
      onchange={(d) => adv && (adv = withDefault(adv, d))}
      {onforbidden}
    />
  {/if}

  {#if data}
    <p class="count">
      {data.accounts.length}
      {data.accounts.length === 1 ? "account" : "accounts"}{filter.played
        ? ` played in the last ${filter.played}`
        : ""}{query.trim() ? ` · ${rows.length} shown` : ""}{data.truncated
        ? " · the first 1,000 only"
        : ""}
    </p>
    {#if rows.length === 0}
      <p class="empty">No accounts match.</p>
    {:else}
      <div class="scroll">
        <table aria-label={L.accountsTable}>
          <thead>
            <tr>
              <th scope="col">Account</th>
              <th scope="col" class="wide">First seen</th>
              <th scope="col">Last sign-in</th>
              <th scope="col" class="num">Games</th>
              <th scope="col">Last played</th>
              {#if adv}<th scope="col">Advanced stats</th>{/if}
            </tr>
          </thead>
          <tbody>
            {#each rows as a (a.id)}
              <tr>
                <td class="tcell">
                  <div>
                    <AccountLink account={a} />
                    {#if a.playing_now}<span class="chip live">playing now</span>{/if}
                  </div>
                </td>
                <td class="wide"><RelTime ms={a.first_seen_at} {now} /></td>
                <td><RelTime ms={a.last_sign_in_at} {now} /></td>
                <td class="num">{a.games_played}</td>
                <td><RelTime ms={a.last_played_at} {now} /></td>
                {#if adv}
                  <td>
                    <AccountAdvancedStats
                      name={a.name || "this account"}
                      row={accountStatsOf(adv, a.id)}
                      onchange={(r) => adv && (adv = withAccount(adv, r))}
                      {onforbidden}
                    />
                  </td>
                {/if}
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {:else if loading}
    <p class="empty" aria-live="polite">Loading accounts…</p>
  {/if}
</div>

<style>
  .count {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .search input {
    width: 12rem;
    max-width: 100%;
  }
</style>
