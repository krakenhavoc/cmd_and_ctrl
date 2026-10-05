<script lang="ts">
  // Games (ADR 0124 §3.3, §7): every table, newest first, with its
  // seats and outcome. The filters live in the hash (#/admin/games?
  // state=active&archived=false, the Grafana tiles' links), so a filter
  // change is a navigation and the list is asked again. Pages are the
  // server's opaque cursor, also in the hash. Practice tables are left
  // out unless `practice` says otherwise, so the unfiltered list counts
  // what the tiles count. Loads once; Refresh asks again.
  import { onMount } from "svelte";
  import { fetchAdminGames } from "../../api";
  import {
    accountHash,
    gameHash,
    gamesHash,
    outcomeLine,
    seatLabel,
    stateLabel,
    type AdminGamesResponse,
    type ArchivedFilter,
    type GameState,
    type GamesFilter,
    type PracticeFilter,
  } from "../../adminViews";
  import { navigate } from "../../router";
  import { LobbyApiError } from "../../session";
  import RelTime from "./RelTime.svelte";

  interface Props {
    filter: GamesFilter;
    onforbidden: () => void;
  }
  const { filter, onforbidden }: Props = $props();

  let data = $state<AdminGamesResponse | null>(null);
  let loading = $state(false);
  let error = $state("");

  async function load(): Promise<void> {
    loading = true;
    error = "";
    try {
      data = await fetchAdminGames(filter);
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? err.message : "couldn't load the tables";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
  });

  // A filter change starts again from the newest table.
  function setFilter(patch: Partial<GamesFilter>): void {
    navigate(gamesHash({ ...filter, ...patch, cursor: undefined }));
  }

  function pick<T extends string>(e: Event): T | undefined {
    const v = (e.currentTarget as HTMLSelectElement).value;
    return v === "" ? undefined : (v as T);
  }

  const now = $derived(data?.generated_at ?? Date.now());
</script>

<div class="view">
  <div class="filters" role="group" aria-label="filter tables">
    <label>
      <span>State</span>
      <select value={filter.state ?? ""} onchange={(e) => setFilter({ state: pick<GameState>(e) })}>
        <option value="">any</option>
        <option value="lobby">waiting</option>
        <option value="active">running</option>
        <option value="ended">ended</option>
      </select>
    </label>
    <label>
      <span>Archived</span>
      <select
        value={filter.archived ?? ""}
        onchange={(e) => setFilter({ archived: pick<ArchivedFilter>(e) })}
      >
        <option value="">any</option>
        <option value="false">not archived</option>
        <option value="true">archived</option>
      </select>
    </label>
    <label>
      <span>Practice</span>
      <select
        value={filter.practice ?? ""}
        onchange={(e) => setFilter({ practice: pick<PracticeFilter>(e) })}
      >
        <option value="">left out</option>
        <option value="include">included</option>
        <option value="only">only practice</option>
      </select>
    </label>
    <button type="button" class="ghost sm" disabled={loading} onclick={() => void load()}>
      {loading ? "Refreshing…" : "Refresh"}
    </button>
  </div>

  {#if filter.user}
    <p class="scope">
      Tables of <a href={accountHash(filter.user)}>one account</a> ·
      <a href={gamesHash({ ...filter, user: undefined, cursor: undefined })}>show every table</a>
    </p>
  {/if}

  {#if error}
    <p class="notice err" role="alert">{error}</p>
  {/if}

  {#if data}
    {#if data.games.length === 0}
      <p class="empty">No tables match.</p>
    {:else}
      <div class="scroll">
        <table aria-label="games">
          <thead>
            <tr>
              <th scope="col">Table</th>
              <th scope="col">State</th>
              <th scope="col" class="wide">Created</th>
              <th scope="col">Ended</th>
              <th scope="col">Outcome</th>
              <th scope="col" class="wide">Seats</th>
            </tr>
          </thead>
          <tbody>
            {#each data.games as g (g.id)}
              <tr>
                <td class="tcell">
                  <div>
                    <a class="tname" href={gameHash(g.id)}>{g.name || "untitled table"}</a>
                    {#if g.practice}<span class="chip">practice</span>{/if}
                    {#if g.archived_at}<span class="chip dim">archived</span>{/if}
                  </div>
                </td>
                <td
                  ><span class="chip" class:live={g.state === "active"}>{stateLabel(g.state)}</span
                  ></td
                >
                <td class="wide"><RelTime ms={g.created_at} {now} /></td>
                <td><RelTime ms={g.ended_at ?? g.archived_at} {now} /></td>
                <td>{outcomeLine(g)}</td>
                <td class="wide seats">
                  {g.seats.length === 0 ? "—" : g.seats.map(seatLabel).join(", ")}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
    <nav class="pages" aria-label="pages of tables">
      {#if filter.cursor}
        <a href={gamesHash({ ...filter, cursor: undefined })}>Newest tables</a>
      {/if}
      {#if data.next_cursor}
        <a href={gamesHash({ ...filter, cursor: data.next_cursor })}>Older tables</a>
      {/if}
    </nav>
  {:else if loading}
    <p class="empty" aria-live="polite">Loading tables…</p>
  {/if}
</div>

<style>
  .scope {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .seats {
    max-width: 28rem;
    color: var(--fg-muted);
    font-size: 12px;
  }
</style>
