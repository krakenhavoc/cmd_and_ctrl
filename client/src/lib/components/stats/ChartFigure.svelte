<script lang="ts">
  // ChartFigure — the frame every stats chart sits in (ADR 0217 §8,
  // DESIGN.md §8): a <figure> whose <figcaption> says the finding in one
  // sentence, the chart (which names that caption as its label), and a
  // "Show as table" disclosure with the same numbers in a real table.
  import type { Snippet } from "svelte";
  import { L } from "../../labels";

  interface Props {
    /** The caption's id; the chart's role="img" points at it. */
    id: string;
    title: string;
    caption: string;
    /** The same numbers as the chart, for the table. */
    columns: string[];
    rows: (string | number)[][];
    /** A definition shown under the chart ("a spell you could pay for…"). */
    note?: string;
    children: Snippet<[string]>;
  }
  const { id, title, caption, columns, rows, note, children }: Props = $props();
</script>

<div class="chart" data-testid="stats-chart">
  <h3 class="chart-title">{title}</h3>
  <figure>
    <figcaption {id}>{caption}</figcaption>
    <div class="visual">{@render children(id)}</div>
    {#if note}<p class="note">{note}</p>{/if}
    {#if rows.length > 0}
      <details class="as-table">
        <summary>{L.showAsTable}</summary>
        <div class="scroll">
          <table>
            <caption class="visually-hidden">{title}</caption>
            <thead>
              <tr>
                {#each columns as c, i (i)}<th scope="col" class:num={i > 0}>{c}</th>{/each}
              </tr>
            </thead>
            <tbody>
              {#each rows as r, ri (ri)}
                <tr>
                  {#each r as cell, i (i)}
                    {#if i === 0}<th scope="row">{cell}</th>{:else}<td class="num">{cell}</td>{/if}
                  {/each}
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </details>
    {/if}
  </figure>
</div>

<style>
  .chart,
  figure {
    margin: 0;
    display: grid;
    /* One shrinkable column: an opened table scrolls inside its own box,
       never widens the page. */
    grid-template-columns: minmax(0, 1fr);
    gap: 6px;
    min-width: 0;
  }
  .chart-title {
    margin: 0;
    font: 600 0.72rem/1.2 var(--font-mono);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  figcaption {
    color: var(--fg);
    font-size: 0.95rem;
    line-height: 1.4;
  }
  .visual {
    min-width: 0;
    margin-top: 4px;
  }
  .note {
    margin: 0;
    color: var(--fg-muted);
    font-size: 0.8rem;
  }
  .as-table summary {
    cursor: pointer;
    color: var(--fg-muted);
    font-size: 0.82rem;
    width: fit-content;
  }
  .as-table summary:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  .scroll {
    overflow-x: auto;
    margin-top: 6px;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }
  th,
  td {
    padding: 5px 8px;
    text-align: left;
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
  }
  thead th {
    font: 600 0.68rem/1 var(--font-mono);
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
  .visually-hidden {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
  }
</style>
