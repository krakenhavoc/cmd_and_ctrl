<script lang="ts">
  // DraftSeats — who is at the draft (ADR 0222 §5).
  //
  // One row per seat: its name, whether it is you, the host or a bot,
  // what it is doing (picking, picked, away) and how many cards it has
  // drafted. That is everything the wire says about another seat: a
  // name, a status and a count, never a card (CR 905.1c).
  //
  // A draft has up to eight seats, so the table's four seat colours are
  // not used here: they mean a seat at a table (DESIGN.md §2).
  import {
    cardCount,
    openSeats,
    seatStatus,
    seatStatusWord,
    type DraftPlayer,
    type DraftView,
  } from "../../draft";
  import { L } from "../../labels";
  import Icon from "../Icon.svelte";

  interface Props {
    view: DraftView;
  }
  const { view }: Props = $props();

  const players = $derived([...view.players].sort((a, b) => a.seat - b.seat));
  const open = $derived(openSeats(view));
  const showCounts = $derived(view.state !== "lobby");

  function roles(p: DraftPlayer): string[] {
    const out: string[] = [];
    if (p.seat === view.you) out.push("you");
    if (p.host) out.push("host");
    return out;
  }
</script>

<ul class="seats" aria-label={L.draftSeats}>
  {#each players as p (p.seat)}
    {@const status = seatStatus(view, p)}
    <li class="seat" class:me={p.seat === view.you} data-seat={p.seat} data-status={status}>
      <span class="num">{p.seat + 1}</span>
      <span class="who">
        <span class="name">{p.name}</span>
        {#each roles(p) as r (r)}<span class="tag">{r}</span>{/each}
        {#if p.bot}
          <span class="tag bot"><Icon name="robot" size={11} /> bot</span>
        {/if}
      </span>
      {#if status !== "seated"}
        <span class="status {status}">
          {#if status === "picked"}<Icon name="check" size={12} />{/if}
          {seatStatusWord(status)}
        </span>
      {/if}
      {#if showCounts}
        <span class="count">{cardCount(p.picked ?? 0)}</span>
      {/if}
    </li>
  {/each}
  {#each { length: open } as _, i (i)}
    <li class="seat open">
      <span class="num">{players.length + i + 1}</span>
      <span class="who"><span class="name dim">Open seat</span></span>
      <span class="status">a bot will sit here</span>
    </li>
  {/each}
</ul>

<style>
  .seats {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .seat {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px 10px;
    min-width: 0;
    padding: 8px 0;
    font-size: 0.875rem;
  }
  .seat + .seat {
    border-top: 1px solid var(--border);
  }
  .num {
    flex: none;
    width: 2ch;
    font-family: var(--font-mono);
    font-size: 11px;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .who {
    flex: 1 1 12ch;
    display: inline-flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px 8px;
    min-width: 0;
  }
  .name {
    font-weight: 600;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .name.dim {
    font-weight: 400;
    color: var(--fg-muted);
  }
  .seat.me .name {
    font-weight: 700;
  }
  .tag.bot {
    border-style: dashed;
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--fg-muted);
  }
  /* The word carries the status; the ink only backs it up. */
  .status.picking {
    color: var(--accent-strong);
  }
  .status.picked {
    color: var(--mint);
  }
  .count {
    flex: none;
    min-width: 8ch;
    text-align: right;
    font-family: var(--font-mono);
    font-size: 11px;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .seat.open {
    border-top-style: dashed;
  }
</style>
