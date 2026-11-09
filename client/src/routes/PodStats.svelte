<script lang="ts">
  // PodStats — #/pods/<id>/stats, a pod's stats and leaderboard (ADR 0220
  // §7 and §11, slice 5 of #4880 = ADR 0217 slice 7). Members only.
  //
  // The leaderboard first: sharing members with at least 5 counted pod
  // games, by win rate, then average place, every row with its count.
  // Then the members (a member who is not sharing shows with no numbers),
  // then the pod's games: how they were won, the first-turn effect, and
  // head-to-head. It reuses the chart kit of #/stats. A member hidden by a
  // block is simply not here; nothing says so (§6).
  import { untrack } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import PodShareSwitch from "../lib/components/PodShareSwitch.svelte";
  import BarChart from "../lib/components/stats/BarChart.svelte";
  import ChartFigure from "../lib/components/stats/ChartFigure.svelte";
  import StatsFilters from "../lib/components/stats/StatsFilters.svelte";
  import StatTile from "../lib/components/stats/StatTile.svelte";
  import { kindText, pctText, rateText, type Makeup } from "../lib/advancedStats";
  import { signedInUserID } from "../lib/myGames";
  import { podErrorLine } from "../lib/pods";
  import {
    boardCaption,
    fetchPodLeaderboard,
    fetchPodStats,
    gamesLine,
    headToHeadLine,
    NOT_SHARING_LINE,
    placeLine,
    podGamesLine,
    podTurnOrderCaption,
    podWinKindsCaption,
    recordLine,
    winRateLine,
    type PodLeaderboard,
    type PodStats,
    type StatsRange,
  } from "../lib/podStats";
  import { LobbyApiError, session } from "../lib/session";
  import { loadStatsPrefs, statsPrefs } from "../lib/statsPrefs";
  import type { ChartBar } from "../lib/statsCharts";

  interface Props {
    id: string;
  }
  const { id }: Props = $props();

  const isPerson = $derived(
    signedInUserID($session) !== null && $session?.principal.role !== "admin",
  );

  let stats = $state<PodStats | null>(null);
  let board = $state<PodLeaderboard | null>(null);
  let range = $state<StatsRange>("all");
  let loading = $state(true);
  let gone = $state(false);
  let error = $state("");

  async function load(): Promise<void> {
    loading = true;
    error = "";
    try {
      [stats, board] = await Promise.all([
        fetchPodStats(id, range),
        fetchPodLeaderboard(id, range),
      ]);
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 404) gone = true;
      else error = podErrorLine(err);
    } finally {
      loading = false;
    }
  }

  // Only the pod id is tracked: a range change reads on its own.
  $effect(() => {
    void id;
    if (!isPerson) {
      loading = false;
      return;
    }
    untrack(() => {
      void load();
      void loadStatsPrefs();
    });
  });

  function onRange(_: Makeup, r: StatsRange): void {
    range = r;
    void load();
  }

  function initial(name: string): string {
    return name.trim().slice(0, 1).toUpperCase() || "?";
  }

  const kindBars = $derived<ChartBar[]>(
    (stats?.win_kinds ?? []).map((k) => ({
      label: kindText(k.kind),
      value: k.games,
      text: `${k.games}`,
      tone: "var(--chart-1)",
    })),
  );
  const positionBars = $derived<ChartBar[]>(
    (stats?.turn_order ?? []).map((r) => ({
      label:
        ["1st to act", "2nd to act", "3rd to act", "4th to act"][r.position - 1] ??
        `#${r.position}`,
      value: r.wins.rate ?? 0,
      text: rateText(r.wins),
      tone: "var(--chart-2)",
    })),
  );
  const sharingCount = $derived((stats?.members ?? []).filter((m) => m.sharing).length);
  const iAmSharing = $derived($statsPrefs.share_pods);
</script>

<Page eyebrow="Pod stats" title={stats?.name ?? "Pod stats"}>
  {#snippet lede()}
    {#if stats}{podGamesLine(stats)}{:else}A pod's record, from the members who share theirs.{/if}
  {/snippet}
  {#snippet actions()}
    <a class="btn ghost" href={`#/pods/${id}`}>Back to the pod</a>
  {/snippet}

  {#if !isPerson}
    <p class="card note" role="status">Sign in with Discord to see your pods.</p>
  {:else if gone}
    <div class="card note" role="alert">
      <p>That pod isn’t there. It may have been deleted, or you may have left it.</p>
      <a class="btn" href="#/pods">Your pods</a>
    </div>
  {:else if error}
    <div class="card note" role="alert">
      <p>{error}</p>
      <button type="button" class="btn sm" onclick={() => void load()}>Try again</button>
    </div>
  {:else if !stats || !board}
    <p class="muted" role="status">Loading…</p>
  {:else}
    <div data-testid="pod-stats">
      <StatsFilters makeup="all" {range} onchange={onRange} disabled={loading} showMakeup={false} />

      {#if !iAmSharing}
        <section class="card nudge" aria-labelledby="ps-you" data-testid="pod-stats-not-sharing">
          <h2 id="ps-you">You’re not on the board</h2>
          <p>
            You aren’t sharing your stats with your pods, so the others don’t see your record and
            you have no row here. You can still look at theirs.
          </p>
          <PodShareSwitch onchange={() => void load()} />
        </section>
      {/if}

      <div class="tiles">
        <StatTile label="Pod games" value={`${stats.games}`} sub="two or more sharing members" />
        <StatTile
          label="Sharing"
          value={`${sharingCount} of ${stats.members.length}`}
          sub="members share their stats"
        />
        <StatTile
          label="Top win rate"
          value={board.rows[0] ? pctText(board.rows[0].win_rate) : "-"}
          sub={board.rows[0]
            ? `${board.rows[0].display_name}, ${gamesLine(board.rows[0].games)}`
            : `needs ${board.min_games} games`}
        />
      </div>

      <div class:stale={loading}>
        <section class="card" aria-labelledby="ps-board" data-testid="pod-leaderboard">
          <h2 id="ps-board">Leaderboard</h2>
          <p class="caption">{boardCaption(board)}</p>
          {#if board.rows.length > 0}
            <div class="scroll">
              <table>
                <thead>
                  <tr>
                    <th scope="col">Rank</th>
                    <th scope="col">Player</th>
                    <th scope="col" class="num">Win rate</th>
                    <th scope="col" class="num">Record</th>
                    <th scope="col" class="num">Avg place</th>
                    <th scope="col" class="num">Games</th>
                  </tr>
                </thead>
                <tbody>
                  {#each board.rows as r (r.user_id)}
                    <tr class:me={r.you}>
                      <td>{r.rank}</td>
                      <th scope="row">{r.display_name}{r.you ? " (you)" : ""}</th>
                      <td class="num">{winRateLine(r)}</td>
                      <td class="num">{recordLine(r)}</td>
                      <td class="num">{placeLine(r.avg_place)}</td>
                      <td class="num">{r.games}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
          {#if board.unranked.length > 0}
            <p class="hint">
              Not ranked yet (a rank needs {board.min_games} pod games):
              {board.unranked
                .map((u) => `${u.display_name} has ${u.games}, ${u.needs} to go`)
                .join("; ")}.
            </p>
          {/if}
        </section>

        <section class="card" aria-labelledby="ps-members">
          <h2 id="ps-members">Members</h2>
          <ul class="rows">
            {#each stats.members as m (m.user_id)}
              <li>
                <span class="who">
                  {#if m.avatar_url}
                    <img class="avatar" src={m.avatar_url} alt="" width="32" height="32" />
                  {:else}
                    <span class="avatar fallback" aria-hidden="true">{initial(m.display_name)}</span
                    >
                  {/if}
                  <span class="text">
                    <span class="name">{m.display_name}{m.you ? " (you)" : ""}</span>
                    {#if m.sharing && m.pod_games && m.overall}
                      <span class="sub">
                        Pod games {recordLine(m.pod_games)} over {gamesLine(m.pod_games.games)} · all
                        games {recordLine(m.overall)}, {winRateLine(m.overall)} won, average place
                        {placeLine(m.overall.avg_place)}
                      </span>
                    {:else}
                      <span class="sub">{NOT_SHARING_LINE}</span>
                    {/if}
                  </span>
                </span>
                <span class="acts">
                  {#if m.owner}<span class="tag">Owner</span>{/if}
                  <span class="tag" class:on={m.sharing}
                    >{m.sharing ? "Sharing" : "Not sharing"}</span
                  >
                </span>
              </li>
            {/each}
          </ul>
        </section>

        {#if stats.games > 0}
          <div class="grid">
            <ChartFigure
              id="cap-pod-kinds"
              title="How pod games are won"
              caption={podWinKindsCaption(stats.win_kinds)}
              columns={["Kind", "Games"]}
              rows={stats.win_kinds.map((k) => [kindText(k.kind), k.games])}
            >
              {#snippet children(cid)}
                {#if kindBars.length > 0}
                  <BarChart labelledby={cid} bars={kindBars} />
                {:else}
                  <p class="muted">No pod game here has a known win kind yet.</p>
                {/if}
              {/snippet}
            </ChartFigure>

            <ChartFigure
              id="cap-pod-order"
              title="Turn order in pod games"
              caption={podTurnOrderCaption(stats.turn_order)}
              columns={["Position", "Win rate"]}
              rows={stats.turn_order.map((r) => [`${r.position}`, rateText(r.wins)])}
            >
              {#snippet children(cid)}
                {#if positionBars.length > 0}
                  <BarChart labelledby={cid} bars={positionBars} max={1} />
                {:else}
                  <p class="muted">Turn order needs four-seat pod games.</p>
                {/if}
              {/snippet}
            </ChartFigure>
          </div>

          <section class="card" aria-labelledby="ps-h2h" data-testid="pod-head-to-head">
            <h2 id="ps-h2h">Head to head</h2>
            {#if stats.head_to_head.length === 0}
              <p class="muted">Two sharing members need to have sat at a table together.</p>
            {:else}
              <ul class="rows">
                {#each stats.head_to_head as h (`${h.a.user_id}-${h.b.user_id}`)}
                  <li><span class="text">{headToHeadLine(h)}</span></li>
                {/each}
              </ul>
              <p class="hint">
                Wins in the games both sat in. A third seat’s win counts for neither.
              </p>
            {/if}
          </section>
        {/if}
      </div>
    </div>
  {/if}
</Page>

<style>
  .muted {
    color: var(--fg-muted);
    margin: 0;
  }
  .note {
    padding: 16px;
  }
  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
    gap: 12px;
    margin: 0 0 16px;
  }
  section.card {
    padding: 16px;
    margin: 0 0 16px;
  }
  h2 {
    margin: 0 0 8px;
    font-size: 1rem;
  }
  .nudge p {
    margin: 0 0 8px;
    color: var(--fg-muted);
  }
  .caption {
    margin: 0 0 12px;
    color: var(--fg-muted);
  }
  .hint {
    margin: 12px 0 0;
    color: var(--fg-muted);
    font-size: 0.85rem;
  }
  .stale {
    opacity: 0.6;
  }
  .scroll {
    overflow-x: auto;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }
  th,
  td {
    padding: 8px 10px;
    text-align: left;
    border-bottom: 1px solid var(--border);
  }
  thead th {
    font: 600 0.7rem/1.2 var(--font-mono);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  tbody th {
    font-weight: 600;
  }
  .num {
    text-align: right;
    font-variant-numeric: tabular-nums;
  }
  tr.me {
    background: var(--surface-raised, var(--surface));
  }
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 8px;
  }
  .rows li {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .who {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .name {
    font-weight: 600;
  }
  .sub {
    color: var(--fg-muted);
    font-size: 0.85rem;
  }
  .avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    flex: none;
  }
  .avatar.fallback {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--surface-raised, var(--surface));
    border: 1px solid var(--border-strong);
    font-weight: 600;
  }
  .acts {
    display: flex;
    gap: 6px;
  }
  .grid {
    display: grid;
    gap: 16px;
    margin: 0 0 16px;
  }
  @media (min-width: 820px) {
    .grid {
      grid-template-columns: 1fr 1fr;
    }
  }
</style>
