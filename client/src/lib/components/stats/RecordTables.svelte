<script lang="ts">
  // RecordTables — ADR 0165 §7's basic record by deck and by commander,
  // for #/stats with advanced stats off and for a shared view.
  import { avgPlaceText, recordText, winRateText, type StatsGroup } from "../../myStats";

  interface Props {
    byDeck: StatsGroup[];
    byCommander: StatsGroup[];
  }
  const { byDeck, byCommander }: Props = $props();
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

{@render breakdown("By deck", "by-deck", byDeck)}
{@render breakdown("By commander", "by-commander", byCommander)}

<style>
  h2 {
    margin-top: 0;
  }
  .muted {
    color: var(--fg-muted);
    margin: 0;
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
    color: var(--fg-muted);
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
</style>
