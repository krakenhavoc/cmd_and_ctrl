<script lang="ts">
  // Stats — your own games (#/stats, ADR 0165 §7 and ADR 0217 §8).
  //
  // Basic stats are on for every signed-in person: record, by deck and by
  // commander, most-cast cards and recent games (GET /me/stats). With
  // advanced stats on (ADR 0217 §2), the page is the advanced overview
  // instead (GET /me/stats/advanced): a filter bar, stat tiles, then
  // Results, Play, Mana, Cards, Colours and Decks, with recent games
  // under them. While advanced stats are off the page says what they are
  // and offers the switch (the opt-in card), never an error. The switch
  // stays at the head of the page once they are on, and "Share my stats"
  // opens the share dialog (ADR 0217 §6).
  //
  // Only the caller's own seats are counted; the server never sends
  // anyone else's. Signed-out visitors are sent to the front page by the
  // router. A guest or admin session is not a person and has no seats of
  // its own, so it sees a line saying to sign in with Discord.
  import { onMount, untrack } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import ShopLinks from "../lib/components/ShopLinks.svelte";
  import ShopNote from "../lib/components/ShopNote.svelte";
  import AdvancedStatsOptIn from "../lib/components/AdvancedStatsOptIn.svelte";
  import AdvancedOverview from "../lib/components/stats/AdvancedOverview.svelte";
  import RecordTables from "../lib/components/stats/RecordTables.svelte";
  import StatsFilters from "../lib/components/stats/StatsFilters.svelte";
  import ShareStatsDialog from "../lib/components/stats/ShareStatsDialog.svelte";
  import { L } from "../lib/labels";
  import { loadStatsPrefs, setAdvancedStats, statsPrefs, statsPrefsNote } from "../lib/statsPrefs";
  import { visibleShops } from "../lib/shopLinksStore";
  import { fetchMyStats } from "../lib/api";
  import {
    fetchAdvancedOverview,
    type AdvancedOverview as Overview,
    type Makeup,
    type StatsRange,
  } from "../lib/advancedStats";
  import { LobbyApiError, session } from "../lib/session";
  import { signedInUserID } from "../lib/myGames";
  import { avgPlaceText, placeText, resultText, winRateText, type MyStats } from "../lib/myStats";

  const isPerson = $derived(signedInUserID($session) !== null);

  let stats = $state<MyStats | null>(null);
  let loading = $state(true);
  let error = $state("");

  let overview = $state<Overview | null>(null);
  let overviewLoading = $state(false);
  let overviewError = $state("");
  let makeup = $state<Makeup>("all");
  let range = $state<StatsRange>("all");
  let sharing = $state(false);
  let switching = $state(false);
  let switchError = $state("");

  const advanced = $derived($statsPrefs.available && $statsPrefs.advanced);

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

  async function loadOverview(): Promise<void> {
    overviewLoading = true;
    overviewError = "";
    try {
      overview = await fetchAdvancedOverview(makeup, range);
    } catch (err) {
      // A 403 here means the setting went off under us (another tab):
      // read it again, and the page falls back to the opt-in card.
      if (err instanceof LobbyApiError && err.status === 403) {
        overview = null;
        void loadStatsPrefs();
      } else {
        overviewError =
          err instanceof LobbyApiError ? err.message : "could not load your advanced stats";
      }
    } finally {
      overviewLoading = false;
    }
  }

  onMount(() => {
    if (isPerson) {
      void load();
      void loadStatsPrefs();
    } else loading = false;
  });

  // Advanced stats on: read the overview; off: drop it.
  // Only `advanced` is tracked: a filter change reads on its own.
  $effect(() => {
    if (advanced) untrack(() => void loadOverview());
    else overview = null;
  });

  function onFilters(m: Makeup, r: StatsRange): void {
    makeup = m;
    range = r;
    void loadOverview();
  }

  async function toggle(on: boolean): Promise<void> {
    switching = true;
    switchError = "";
    try {
      await setAdvancedStats(on);
    } catch (e) {
      switchError = e instanceof Error ? e.message : "could not save; try again";
    } finally {
      switching = false;
    }
  }

  function when(ms: number): string {
    return new Date(ms).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }
</script>

<Page
  eyebrow="Stats"
  title="Your games"
  lede="How your decks and commanders do at CMD & CTRL tables. Only your own seats are counted."
>
  {#snippet actions()}
    {#if isPerson}
      <a class="btn ghost" href="#/pods">{L.navPods}</a>
    {/if}
    {#if isPerson && $statsPrefs.available}
      <button type="button" class="btn" onclick={() => (sharing = true)} data-testid="share-stats">
        {L.shareStats}
      </button>
    {/if}
  {/snippet}

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
    {#if advanced}
      <div class="switch-bar">
        <label class="switch-row">
          <input
            type="checkbox"
            checked={$statsPrefs.advanced}
            disabled={switching}
            onchange={(e) => toggle(e.currentTarget.checked)}
            data-testid="advanced-stats-switch"
          />
          <span>
            {L.advancedStatsSwitch}
            <span class="help">{statsPrefsNote($statsPrefs)}</span>
          </span>
        </label>
        {#if switchError}<p class="error-line" role="alert">{switchError}</p>{/if}
      </div>
    {:else}
      <AdvancedStatsOptIn />
    {/if}

    {#if stats.games === 0}
      <p class="card note" role="status" data-testid="stats-empty">
        No finished games yet. Play a game to the end and your record starts here.
      </p>
    {:else}
      {#if advanced}
        <StatsFilters {makeup} {range} onchange={onFilters} disabled={overviewLoading} />
        {#if overviewError}
          <p class="card note error" role="alert">{overviewError}</p>
          <button type="button" class="btn" onclick={() => void loadOverview()}>Try again</button>
        {:else if overview}
          <div class:stale={overviewLoading} data-testid="advanced-overview">
            <AdvancedOverview {overview} />
          </div>
        {:else}
          <p class="muted" role="status">Loading your advanced stats…</p>
        {/if}
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
      {/if}
      {#if stats.unfinished > 0}
        <p class="muted small">
          {stats.unfinished}
          {stats.unfinished === 1 ? "table" : "tables"} ended with no result and
          {stats.unfinished === 1 ? "is" : "are"} not counted.
        </p>
      {/if}

      <div class="stack">
        {#if !advanced}
          <RecordTables byDeck={stats.by_deck} byCommander={stats.by_commander} />
        {/if}

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

{#if sharing}
  <ShareStatsDialog onclose={() => (sharing = false)} />
{/if}

<style>
  .stack {
    display: grid;
    gap: 16px;
    margin-top: 16px;
  }
  .muted {
    color: var(--fg-muted);
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
  .switch-bar {
    margin: 0 0 16px;
  }
  .switch-row {
    display: inline-flex;
    gap: 0.6rem;
    align-items: flex-start;
    font-weight: 600;
  }
  .switch-row input {
    margin-top: 3px;
  }
  .switch-row .help {
    display: block;
    font-weight: 400;
    font-size: 0.85rem;
    color: var(--fg-muted);
  }
  .error-line {
    margin: 6px 0 0;
    color: var(--danger);
  }
  .stale {
    opacity: 0.6;
    transition: opacity 120ms var(--ease);
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
  @media (max-width: 819px) {
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
