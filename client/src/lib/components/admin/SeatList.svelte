<script lang="ts">
  // SeatList: a table's seats as Live now and a table's detail show
  // them (ADR 0124 §3.3): who holds each seat (an account opens its
  // detail), what kind of holder it is, the host, the deck, and, at a
  // loaded table, how many sockets are on it and since when.
  import { seatKind, seatName, seatNumber, type AdminSeat } from "../../adminViews";
  import AccountLink from "./AccountLink.svelte";
  import RelTime from "./RelTime.svelte";

  interface Props {
    seats: AdminSeat[];
    now: number;
    winnerSeat?: number;
    label: string;
  }
  const { seats, now, winnerSeat, label }: Props = $props();
</script>

{#if seats.length === 0}
  <p class="none">No seats yet.</p>
{:else}
  <ul class="seats" aria-label={label}>
    {#each seats as s (s.seat)}
      <li class="seat" class:off={s.connected === 0 && s.kind !== "bot"}>
        <span class="no">{seatNumber(s.seat)}</span>
        <span class="who">
          {#if s.account}
            <AccountLink account={s.account} fallback={seatName(s)} />
          {:else}
            <span class="name">{seatName(s)}</span>
          {/if}
          <span class="kind" class:bot={s.kind === "bot"} class:agent={s.kind === "agent"}
            >{seatKind(s)}</span
          >
          {#if s.host}<span class="tag">host</span>{/if}
          {#if winnerSeat === s.seat}<span class="tag won">winner</span>{/if}
        </span>
        <span class="extra">
          {#if s.deck_name}<span class="deck" title="deck">{s.deck_name}</span>{/if}
          {#if s.connected !== undefined && s.kind !== "bot"}
            {#if s.connected > 0}
              <span class="conn on">
                connected{s.connected > 1 ? ` ×${s.connected}` : ""}
                {#if s.since}· <RelTime ms={s.since} {now} prefix="since" />{/if}
              </span>
            {:else}
              <span class="conn">not connected</span>
            {/if}
          {/if}
        </span>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .seats {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
  }
  .seat {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px 10px;
    padding: 7px 0;
    border-top: 1px solid var(--border);
    font-size: 13px;
  }
  .seat:first-child {
    border-top: none;
  }
  .no {
    flex: none;
    width: 4.2em;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.06em;
    color: var(--fg-dim);
  }
  .who {
    display: inline-flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    min-width: 0;
    flex: 1 1 12rem;
  }
  .name {
    font-weight: 600;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .kind,
  .tag {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.06em;
    padding: 1px 6px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    color: var(--fg-muted);
    white-space: nowrap;
  }
  .kind.agent {
    color: var(--agent);
    border-color: var(--agent-border);
    background: var(--agent-soft);
  }
  .kind.bot {
    color: var(--block);
  }
  .tag.won {
    color: var(--accent-strong);
    border-color: var(--accent);
  }
  .extra {
    display: inline-flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px 10px;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .deck {
    overflow-wrap: anywhere;
  }
  .conn.on {
    color: var(--mint);
  }
  .seat.off .name {
    color: var(--fg-muted);
  }
  .none {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
</style>
