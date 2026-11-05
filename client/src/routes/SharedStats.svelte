<script lang="ts">
  // SharedStats — a person's shared stats (#/stats/shared/<token>, ADR
  // 0217 §6 and §8). Read-only, and it needs no sign-in: the person who
  // made the link chose to post it. It shows their display name and
  // their own stats (the advanced overview while their advanced stats
  // are on, else their basic record), never another player's name, a
  // table name, a game or a replay. The viewer can change the filters.
  // A revoked or unknown link says so.
  import { untrack } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import AdvancedOverview from "../lib/components/stats/AdvancedOverview.svelte";
  import RecordTables from "../lib/components/stats/RecordTables.svelte";
  import StatsFilters from "../lib/components/stats/StatsFilters.svelte";
  import StatTile from "../lib/components/stats/StatTile.svelte";
  import {
    fetchSharedStats,
    type Makeup,
    type SharedStats,
    type StatsRange,
  } from "../lib/advancedStats";
  import { LobbyApiError } from "../lib/session";
  import { avgPlaceText, recordText, winRateText } from "../lib/myStats";

  interface Props {
    token: string;
  }
  const { token }: Props = $props();

  let data = $state<SharedStats | null>(null);
  let loading = $state(true);
  let error = $state("");
  let gone = $state(false);
  let makeup = $state<Makeup>("all");
  let range = $state<StatsRange>("all");

  async function load(): Promise<void> {
    loading = true;
    error = "";
    gone = false;
    try {
      data = await fetchSharedStats(token, makeup, range);
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 404) gone = true;
      else error = err instanceof LobbyApiError ? err.message : "could not load these stats";
    } finally {
      loading = false;
    }
  }

  // Only the token is tracked: a filter change reads on its own.
  $effect(() => {
    void token;
    untrack(() => void load());
  });

  function onFilters(m: Makeup, r: StatsRange): void {
    makeup = m;
    range = r;
    void load();
  }

  const title = $derived(data ? `${data.display_name}'s stats` : "Shared stats");
</script>

<Page
  eyebrow="Shared stats"
  {title}
  lede="A read-only copy of one player's own CMD & CTRL games, shared by them. Other players are never named."
>
  {#if gone}
    <p class="card note" role="status" data-testid="shared-stats-gone">
      This stats link doesn't work. It may have been revoked, or a newer link replaced it. Ask the
      person who shared it for a new one.
    </p>
  {:else if error}
    <p class="card note error" role="alert">{error}</p>
    <button type="button" class="btn" onclick={() => void load()}>Try again</button>
  {:else if !data}
    <p class="muted" role="status">Loading…</p>
  {:else}
    <div data-testid="shared-stats">
      {#if data.advanced && data.overview}
        <StatsFilters {makeup} {range} onchange={onFilters} disabled={loading} />
        <div class:stale={loading}>
          <AdvancedOverview overview={data.overview} shared />
        </div>
      {:else}
        <p class="muted small">
          {data.display_name} has advanced stats switched off, so this link shows their basic record.
        </p>
        <div class="tiles">
          <StatTile label="Games played" value={`${data.basic.games}`} />
          <StatTile
            label="Record (W-L)"
            value={recordText(data.basic)}
            sub={`${winRateText(data.basic)} won`}
          />
          <StatTile label="Average place" value={avgPlaceText(data.basic.avg_place)} />
        </div>
        <div class="stack">
          <RecordTables byDeck={data.basic.by_deck} byCommander={data.basic.by_commander} />
        </div>
      {/if}
    </div>
  {/if}
</Page>

<style>
  .muted {
    color: var(--fg-muted);
    margin: 0;
  }
  .small {
    font-size: 0.9rem;
    margin-bottom: 12px;
  }
  .note.error {
    color: var(--danger);
    border-color: var(--danger);
    margin-bottom: 12px;
  }
  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 12px;
  }
  .stack {
    display: grid;
    gap: 16px;
    margin-top: 16px;
  }
  .stale {
    opacity: 0.6;
  }
</style>
