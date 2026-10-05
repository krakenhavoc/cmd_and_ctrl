<script lang="ts">
  // Accounts (ADR 0124 §3.1, §7): every account, one row each, with
  // first seen, last sign-in, games played and last played. `played`
  // (#/admin/accounts?played=7d, the Accounts-that-played tile's link)
  // is the server's filter and keeps exactly the accounts the tile
  // counts. `sort` is in the hash too but never sent: the whole list is
  // on the page, so it is sorted here, and the search box filters it as
  // you type. Loads once; Refresh asks again.
  import { onMount } from "svelte";
  import { fetchAdminAccounts } from "../../api";
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

  async function load(): Promise<void> {
    loading = true;
    error = "";
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
        <table aria-label="accounts">
          <thead>
            <tr>
              <th scope="col">Account</th>
              <th scope="col" class="wide">First seen</th>
              <th scope="col">Last sign-in</th>
              <th scope="col" class="num">Games</th>
              <th scope="col">Last played</th>
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
