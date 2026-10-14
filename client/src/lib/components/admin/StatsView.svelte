<script lang="ts">
  // Stats (ADR 0165 §8, #3425): what the game statistics hold, for
  // operators who have no shell on the database. Totals first (and, when
  // finished games have no record yet, a note that the backfill fills
  // them in), then recent games, per-person records, the top commanders
  // by games and by wins, and the most-cast cards. Every list is capped
  // by the server (25 rows). Only a win, a loss or a draw counts toward
  // a record; average place is over the games that have a place.
  //
  // Admin only, like every view under this tab strip: a 403 means the
  // admin status is stale (onforbidden), never a request to retry.
  import { onMount } from "svelte";
  import { fetchAdminStats } from "../../api";
  import { gameHash, type AdminStatsResponse, type AdminStatsTally } from "../../adminViews";
  import { avgPlaceText, recordText, winRateText } from "../../myStats";
  import { LobbyApiError } from "../../session";
  import RelTime from "./RelTime.svelte";

  interface Props {
    onforbidden: () => void;
  }
  const { onforbidden }: Props = $props();

  let data = $state<AdminStatsResponse | null>(null);
  let loading = $state(false);
  let error = $state("");

  async function load(): Promise<void> {
    loading = true;
    error = "";
    try {
      data = await fetchAdminStats();
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 403) {
        onforbidden();
        return;
      }
      error = err instanceof LobbyApiError ? err.message : "couldn't load the statistics";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
  });

  const now = $derived(Date.now());
  const empty = $derived(data !== null && data.totals.games === 0 && data.totals.unrecorded === 0);
</script>

{#snippet tallyCells(t: AdminStatsTally)}
  <td class="num">{t.games}</td>
  <td class="num">{recordText(t)}</td>
  <td class="num wide">{winRateText(t)}</td>
  <td class="num">{avgPlaceText(t.avg_place)}</td>
{/snippet}

{#snippet tallyHead()}
  <th class="num">Games</th>
  <th class="num">Record</th>
  <th class="num wide">Win rate</th>
  <th class="num">Avg place</th>
{/snippet}

<div class="view">
  <div class="thead">
    <p class="scope">
      Counted from each finished game's record. Only wins, losses and draws count toward a record.
    </p>
    <button type="button" class="ghost sm" disabled={loading} onclick={() => void load()}>
      {loading ? "Refreshing…" : "Refresh"}
    </button>
  </div>

  {#if error}
    <p class="notice err" role="alert">{error}</p>
  {/if}

  {#if data}
    <section class="card" aria-label="statistics totals">
      <dl class="facts">
        <div>
          <dt>Games with a record</dt>
          <dd>{data.totals.games}</dd>
        </div>
        <div>
          <dt>Live</dt>
          <dd>{data.totals.live}</dd>
        </div>
        <div>
          <dt>Replay</dt>
          <dd>{data.totals.replay}</dd>
        </div>
        <div>
          <dt>Log complete</dt>
          <dd>{data.totals.log_complete}</dd>
        </div>
        <div>
          <dt>Log incomplete</dt>
          <dd>{data.totals.log_incomplete}</dd>
        </div>
        <div>
          <dt>Ended, no record</dt>
          <dd>{data.totals.unrecorded}</dd>
        </div>
      </dl>
      {#if data.totals.unrecorded > 0}
        <p class="notice backfill">
          {data.totals.unrecorded} finished {data.totals.unrecorded === 1
            ? "game has"
            : "games have"}
          no record yet. Running the backfill (<code>statsbackfill</code>) computes them from their
          replays.
        </p>
      {/if}
    </section>

    {#if empty}
      <p class="empty">No game statistics yet. A record is made when a game ends.</p>
    {:else}
      <h3>Recent games</h3>
      {#if data.recent.length === 0}
        <p class="empty">No games with a record.</p>
      {:else}
        <div class="scroll">
          <table aria-label="recent games">
            <thead>
              <tr>
                <th>Game</th>
                <th>Ended</th>
                <th class="num">Turns</th>
                <th class="num">Rounds</th>
                <th class="wide">Source</th>
              </tr>
            </thead>
            <tbody>
              {#each data.recent as g (g.game_id)}
                <tr>
                  <td
                    ><a class="tname" href={gameHash(g.game_id)}>{g.name || "untitled table"}</a
                    ></td
                  >
                  <td><RelTime ms={g.ended_at} {now} /></td>
                  <td class="num">{g.turns}</td>
                  <td class="num">{g.rounds}</td>
                  <td class="wide dim">{g.source}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}

      <h3>Records by person</h3>
      {#if data.people.length === 0}
        <p class="empty">No finished games yet.</p>
      {:else}
        <div class="scroll">
          <table aria-label="records by person">
            <thead>
              <tr><th>Player</th>{@render tallyHead()}</tr>
            </thead>
            <tbody>
              {#each data.people as p (p.kind + ":" + p.name)}
                <tr>
                  <td
                    >{p.name}{#if p.kind !== "account"}
                      <span class="chip dim">{p.kind}</span>{/if}</td
                  >
                  {@render tallyCells(p)}
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}

      <h3>Top commanders by games</h3>
      {#if data.commanders_by_games.length === 0}
        <p class="empty">No commanders recorded.</p>
      {:else}
        <div class="scroll">
          <table aria-label="commanders by games">
            <thead>
              <tr><th>Commander</th>{@render tallyHead()}</tr>
            </thead>
            <tbody>
              {#each data.commanders_by_games as c (c.name)}
                <tr><td>{c.name}</td>{@render tallyCells(c)}</tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}

      <h3>Top commanders by wins</h3>
      {#if data.commanders_by_wins.length > 0}
        <div class="scroll">
          <table aria-label="commanders by wins">
            <thead>
              <tr><th>Commander</th>{@render tallyHead()}</tr>
            </thead>
            <tbody>
              {#each data.commanders_by_wins as c (c.name)}
                <tr><td>{c.name}</td>{@render tallyCells(c)}</tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <p class="empty">No commanders recorded.</p>
      {/if}

      <h3>Most-cast cards</h3>
      {#if data.cards.length === 0}
        <p class="empty">No casts recorded.</p>
      {:else}
        <div class="scroll">
          <table aria-label="most-cast cards">
            <thead>
              <tr><th>Card</th><th class="num">Casts</th><th class="num">Games</th></tr>
            </thead>
            <tbody>
              {#each data.cards as c (c.name)}
                <tr>
                  <td>{c.name}</td>
                  <td class="num">{c.casts}</td>
                  <td class="num">{c.games}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    {/if}
  {:else if loading}
    <p class="empty" aria-live="polite">Loading statistics…</p>
  {/if}
</div>

<style>
  .scope {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
    flex: 1 1 auto;
  }
  h3 {
    margin: 8px 0 0;
    font-size: 14px;
  }
  .backfill {
    margin-top: 12px;
    color: var(--fg-muted);
  }
</style>
