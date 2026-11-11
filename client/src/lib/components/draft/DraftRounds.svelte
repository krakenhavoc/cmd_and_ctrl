<script lang="ts">
  // DraftRounds — a draft's rounds and its end (ADR 0222 §6, #5440):
  // who the viewer plays and the way to that table, every match of
  // every round, and the win-loss table.
  //
  // A match is played at a limited table the server opens. This page
  // only reads results; seats are named by number on the wire, and no
  // deck or table of another seat is ever sent.
  import { cardCount, DRAFT_NEW_HASH, type DraftView } from "../../draft";
  import {
    matchLengthName,
    matchLine,
    recordText,
    standingRows,
    yourMatchLine,
    yourMatchState,
  } from "../../draftRounds";
  import { L } from "../../labels";
  import Icon from "../Icon.svelte";

  interface Props {
    view: DraftView;
    live: boolean;
    /** The seat at the table is being taken. */
    busy: boolean;
    error: string;
    /** The draft has closed under the page. */
    closed: boolean;
    /** The viewer is signed in, so the draft is kept in My drafts. */
    kept: boolean;
    onplay: () => void;
  }
  const { view, live, busy, error, closed, kept, onplay }: Props = $props();

  const play = $derived(view.play);
  const finished = $derived(view.phase === "finished");
  const state = $derived(yourMatchState(view));
  const rows = $derived(standingRows(view));
  const mine = $derived(rows.find((r) => r.you));
</script>

<section class="card now" data-testid="draft-rounds">
  {#if finished}
    <h2 class="headline">{L.draftFinished}</h2>
    <p class="help">
      {#if mine && mine.wins + mine.losses + mine.draws > 0}
        You finished {recordText(mine)} with {cardCount((view.pool ?? []).length)} from
        {view.set_name ?? view.set.toUpperCase()}.
      {:else}
        You drafted {cardCount((view.pool ?? []).length)} from
        {view.set_name ?? view.set.toUpperCase()}.
      {/if}
      {#if kept}
        Your picks, your deck and your results are kept in {L.myDrafts}.
      {:else}
        This draft is not kept: it is here while the page stays open, and gone once it closes.
      {/if}
    </p>
    {#if closed}
      <p class="help">This draft has closed. What is on this page stays until you leave it.</p>
    {/if}
    <div class="frow">
      <a class="link-arrow" href={DRAFT_NEW_HASH}>{L.draftNewDraft} →</a>
      {#if kept}<a class="link-arrow" href="#/drafts">{L.myDrafts} →</a>{/if}
    </div>
  {:else if play}
    <div class="head">
      <h2 class="headline">{L.draftRound(play.round, play.rounds)}</h2>
      <span class="chip">{matchLengthName(view.best_of)}</span>
    </div>
    <p class="help strong" data-testid="draft-your-match">{yourMatchLine(view)}</p>
    {#if error}
      <p class="notice err" role="alert">
        <Icon name="x" size={14} />
        <span>{error}</span>
      </p>
    {/if}
    {#if state === "ready" || state === "opening"}
      <div class="frow">
        <button
          type="button"
          class="btn primary"
          disabled={state !== "ready" || busy || !live}
          onclick={onplay}
        >
          {busy ? "Taking your seat…" : L.draftPlayMatch}
        </button>
      </div>
      <p class="help">
        The match is played at its own table. When it ends, come back here for the next round.
      </p>
    {/if}
  {/if}
</section>

{#if play}
  <div class="cols">
    <section class="card" aria-labelledby="draft-matches-h">
      <h2 id="draft-matches-h">Matches</h2>
      {#each [...play.matches].reverse() as round, i (play.matches.length - i)}
        {@const n = play.matches.length - i}
        <h3>Round {n}</h3>
        <ul class="matches" data-round={n}>
          {#each round as m (`${m.a}-${m.b ?? "bye"}`)}
            <li class:mine={m.a === view.you || m.b === view.you} class:open={!m.done}>
              {matchLine(view, m)}
            </li>
          {/each}
        </ul>
      {/each}
      <p class="help foot">A match between two bots is not played: its result is a coin toss.</p>
    </section>

    <section class="card" aria-labelledby="draft-records-h">
      <h2 id="draft-records-h">{L.draftRecords}</h2>
      <table data-testid="draft-records">
        <thead>
          <tr>
            <th scope="col">Seat</th>
            <th scope="col" class="num">Won</th>
            <th scope="col" class="num">Lost</th>
            <th scope="col" class="num">Drawn</th>
          </tr>
        </thead>
        <tbody>
          {#each rows as r (r.seat)}
            <tr class:you={r.you}>
              <th scope="row">
                {r.name}{#if r.you}{" "}<span class="tagline">you</span>{:else if r.bot}{" "}<span
                    class="tagline">bot</span
                  >{/if}
              </th>
              <td class="num">{r.wins}</td>
              <td class="num">{r.losses}</td>
              <td class="num">{r.draws}</td>
            </tr>
          {/each}
        </tbody>
      </table>
      <p class="help foot">There are no prizes and no standings beyond this table.</p>
    </section>
  </div>
{/if}

<style>
  .now {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px 12px;
  }
  .cols {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(280px, 380px);
    gap: 22px;
    align-items: start;
  }
  @media (max-width: 819px) {
    .cols {
      grid-template-columns: minmax(0, 1fr);
    }
  }
  h2 {
    margin: 0 0 12px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    color: var(--fg);
  }
  h2.headline {
    margin: 0;
    font-size: 24px;
  }
  h3 {
    margin: 14px 0 6px;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  h2 + h3 {
    margin-top: 0;
  }
  .help {
    margin: 0;
    font-size: 0.875rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
  .help.strong {
    font-weight: 700;
    color: var(--fg);
  }
  .help.foot {
    margin-top: 12px;
    font-size: 0.75rem;
  }
  .frow {
    display: flex;
    flex-wrap: wrap;
    gap: 10px 18px;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    padding: 3px 10px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.04em;
    color: var(--fg-muted);
  }
  .matches {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .matches li {
    padding: 6px 0;
    font-size: 0.875rem;
    color: var(--fg);
  }
  .matches li + li {
    border-top: 1px solid var(--border);
  }
  .matches li.open {
    color: var(--fg-muted);
  }
  /* The viewer's own match: heavier, and marked at the edge. */
  .matches li.mine {
    font-weight: 700;
    padding-left: 10px;
    border-left: 2px solid var(--accent);
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }
  th,
  td {
    padding: 6px 4px;
    text-align: left;
    font-weight: 400;
    color: var(--fg);
    border-top: 1px solid var(--border);
  }
  thead th {
    border-top: 0;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .num {
    text-align: right;
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
  }
  tr.you th,
  tr.you td {
    font-weight: 700;
  }
  .tagline {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .notice {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 0;
    padding: 10px 14px;
    border-radius: var(--radius);
    font-size: 0.8125rem;
    line-height: 1.5;
    color: var(--fg);
  }
  .notice.err {
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 35%, transparent);
  }
</style>
