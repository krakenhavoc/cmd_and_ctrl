<script lang="ts">
  // Stats — your own games (#/stats, ADR 0165 §7, #3366): games played,
  // record, average place, the same by deck and by commander, and your
  // recent games with turns and place. A free page for every signed-in
  // person, in ADR 0158's frame. It shows only the caller's own seats
  // (GET /me/stats); the server never sends anyone else's.
  //
  // Signed-out visitors are sent to the front page by the router. A
  // guest or admin session is not a person and has no seats of its own,
  // so it sees a line saying to sign in with Discord.
  import { onMount } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import ShopLinks from "../lib/components/ShopLinks.svelte";
  import ShopNote from "../lib/components/ShopNote.svelte";
  import { visibleShops } from "../lib/shopLinksStore";
  import { fetchMyStats } from "../lib/api";
  import { LobbyApiError, session } from "../lib/session";
  import { signedInUserID } from "../lib/myGames";
  import {
    avgPlaceText,
    placeText,
    recordText,
    resultText,
    winRateText,
    type MyStats,
    type StatsGroup,
  } from "../lib/myStats";

  const isPerson = $derived(signedInUserID($session) !== null);

  let stats = $state<MyStats | null>(null);
  let loading = $state(true);
  let error = $state("");

  async function load(): Promise<void> {
    loading = true;
    error = "";
    try {
      stats = await fetchMyStats();
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "could not load your statistics";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    if (isPerson) void load();
    else loading = false;
  });

  function when(ms: number): string {
    return new Date(ms).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }
</script>

{#snippet breakdown(title: string, id: string, rows: StatsGroup[])}
  <section class="card" aria-labelledby={id}>
    <h2 {id}>{title}</h2>
    {#if rows.length === 0}
      <p class="muted">Nothing yet.</p>
    {:else}
      <div class="scroll">
        <table>
          <thead>
            <tr>
              <th scope="col">Name</th>
              <th scope="col" class="num">Games</th>
              <th scope="col" class="num">Record</th>
              <th scope="col" class="num">Win rate</th>
              <th scope="col" class="num">Avg. place</th>
            </tr>
          </thead>
          <tbody>
            {#each rows as row (row.name)}
              <tr>
                <th scope="row" class="name">{row.name}</th>
                <td class="num">{row.games}</td>
                <td class="num">{recordText(row)}</td>
                <td class="num">{winRateText(row)}</td>
                <td class="num">{avgPlaceText(row.avg_place)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </section>
{/snippet}

<Page
  eyebrow="Stats"
  title="Your games"
  lede="How your decks and commanders do at CMD & CTRL tables. Only your own seats are counted."
>
  {#if !isPerson}
    <p class="card note" role="status">
      Statistics belong to a signed-in account. Sign in with Discord to see how your games go.
    </p>
  {:else if loading}
    <p class="muted" role="status">Loading your statistics…</p>
  {:else if error}
    <p class="card note error" role="alert">{error}</p>
    <button type="button" class="btn" onclick={() => void load()}>Try again</button>
  {:else if stats}
    {#if stats.games === 0}
      <p class="card note" role="status" data-testid="stats-empty">
        No finished games yet. Play a game to the end and your record starts here.
      </p>
    {:else}
      <div class="tiles" data-testid="stats-tiles">
        <div class="tile">
          <span class="tile-label">Games played</span>
          <span class="tile-value">{stats.games}</span>
        </div>
        <div class="tile">
          <span class="tile-label">Record (W-L-D)</span>
          <span class="tile-value">{stats.wins}-{stats.losses}-{stats.draws}</span>
          <span class="tile-sub">{winRateText(stats)} won</span>
        </div>
        <div class="tile">
          <span class="tile-label">Average place</span>
          <span class="tile-value">{avgPlaceText(stats.avg_place)}</span>
        </div>
      </div>
      {#if stats.unfinished > 0}
        <p class="muted small">
          {stats.unfinished}
          {stats.unfinished === 1 ? "table" : "tables"} ended with no result and
          {stats.unfinished === 1 ? "is" : "are"} not counted.
        </p>
      {/if}

      <div class="stack">
        {@render breakdown("By deck", "by-deck", stats.by_deck)}
        {@render breakdown("By commander", "by-commander", stats.by_commander)}

        {#if (stats.top_cards ?? []).length > 0 && $visibleShops.length > 0}
          <section class="card" aria-labelledby="top-cards" data-testid="top-cards">
            <h2 id="top-cards">Your most-cast cards</h2>
            <ShopNote />
            <ul class="top">
              {#each stats.top_cards ?? [] as c (c.name)}
                <li>
                  <span class="cname">{c.name}</span>
                  <span class="casts">{c.casts} {c.casts === 1 ? "cast" : "casts"}</span>
                  <ShopLinks card={c.name} compact note={false} />
                </li>
              {/each}
            </ul>
          </section>
        {/if}

        <section class="card" aria-labelledby="recent">
          <h2 id="recent">Recent games</h2>
          <ul class="recent">
            {#each stats.recent as g (g.game_id)}
              <li class="game">
                <span class="chip {g.result}">{resultText(g.result)}</span>
                <span class="game-main">
                  <span class="game-name">{g.name}</span>
                  <span class="game-sub">
                    {g.commanders.length > 0 ? g.commanders.join(" / ") : (g.deck ?? "")}
                  </span>
                </span>
                <span class="game-meta">
                  {placeText(g.place)} of {g.players} · {g.turns}
                  {g.turns === 1 ? "turn" : "turns"} · {when(g.ended_at)}
                </span>
              </li>
            {/each}
          </ul>
        </section>
      </div>
    {/if}
  {/if}
</Page>

<style>
  .stack {
    display: grid;
    gap: 16px;
    margin-top: 16px;
  }
  .muted {
    color: var(--fg-dim);
    margin: 0;
  }
  .small {
    font-size: 0.85rem;
    margin-top: 8px;
  }
  .note.error {
    color: var(--danger, var(--fg));
    border-color: var(--danger, var(--border));
    margin-bottom: 12px;
  }
  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 12px;
  }
  .tile {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 16px 18px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    min-width: 0;
  }
  .tile-label {
    font: 600 0.72rem/1 var(--font-mono);
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .tile-value {
    font: 800 2rem/1.1 var(--font-display);
    font-variant-numeric: tabular-nums;
  }
  .tile-sub {
    color: var(--fg-dim);
    font-size: 0.85rem;
  }
  h2 {
    margin-top: 0;
  }
  .scroll {
    overflow-x: auto;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.92rem;
  }
  th,
  td {
    padding: 8px 10px;
    text-align: left;
    border-bottom: 1px solid var(--border);
  }
  thead th {
    font: 600 0.7rem/1 var(--font-mono);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  tbody tr:last-child > * {
    border-bottom: 0;
  }
  th.name {
    font-weight: 600;
    min-width: 12ch;
  }
  .num {
    text-align: right;
    white-space: nowrap;
    font-variant-numeric: tabular-nums;
  }
  .top {
    list-style: none;
    margin: 8px 0 0;
    padding: 0;
    display: grid;
  }
  .top li {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto auto;
    gap: 4px 12px;
    align-items: baseline;
    padding: 8px 0;
    border-bottom: 1px solid var(--border);
  }
  .top li:last-child {
    border-bottom: 0;
  }
  .cname {
    font-weight: 600;
    overflow-wrap: anywhere;
  }
  .casts {
    color: var(--fg-dim);
    font-size: 0.85rem;
    white-space: nowrap;
  }
  .recent {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
  }
  .game {
    display: grid;
    grid-template-columns: 5.5rem 1fr auto;
    gap: 12px;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid var(--border);
  }
  .game:last-child {
    border-bottom: 0;
  }
  .game-main {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .game-name,
  .game-sub {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .game-name {
    font-weight: 600;
  }
  .game-sub,
  .game-meta {
    color: var(--fg-dim);
    font-size: 0.85rem;
  }
  .chip {
    justify-self: start;
    padding: 3px 10px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    font: 700 0.75rem/1.2 var(--font-ui);
  }
  .chip.win {
    border-color: var(--accent);
    color: var(--accent);
  }
  @media (max-width: 640px) {
    .top li {
      grid-template-columns: minmax(0, 1fr) auto;
    }
    .top li :global(.shop) {
      grid-column: 1 / -1;
    }
    .game {
      grid-template-columns: 1fr;
      gap: 4px;
    }
  }
</style>
