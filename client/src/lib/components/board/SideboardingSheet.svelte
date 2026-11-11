<script lang="ts">
  // SideboardingSheet — the screen between two games of a match (ADR
  // 0222 §3). It is not something the game asks during play, so it is a
  // full-screen dialog (DESIGN.md §6), like the zone browser: the last
  // game is over, and the board behind it is the finished one, which
  // "Look at the board" shows.
  //
  // The lists are the player's own pool, by name, in the shape the Decks
  // page lists a sideboard (decks/SideboardList.svelte). A move is kept
  // here until the deck is one the format's check takes (sixty cards,
  // fifteen in the sideboard), then sent at once: a one-for-one swap is
  // two clicks, and the deck the timer finds is always the last legal
  // one. The server's check is the one that counts.
  import { untrack } from "svelte";
  import { seatColor } from "../../colors";
  import { L } from "../../labels";
  import {
    deckProblem,
    gameOfText,
    lastGameText,
    mainOf,
    moveCard,
    poolLines,
    sameDeck,
    scoreLine,
    type PoolLine,
  } from "../../match";
  import type { MatchPoolView, MatchView, PlayerView, SideboardingView } from "../../protocol";
  import ModalLayer from "../ModalLayer.svelte";
  import MatchClock from "./MatchClock.svelte";

  const {
    match,
    sideboarding,
    pool,
    seats,
    viewerSeat,
    error = "",
    onSetDeck,
    onChoosePlayFirst,
    onReady,
    onConcedeMatch,
    onHide,
  }: {
    match: MatchView;
    sideboarding: SideboardingView;
    pool: MatchPoolView;
    seats: PlayerView[];
    viewerSeat: number;
    /** The server's refusal of the last deck sent, if it refused one. */
    error?: string;
    onSetDeck: (main: string[]) => void;
    onChoosePlayFirst: (playFirst: boolean) => void;
    onReady: (ready: boolean) => void;
    onConcedeMatch: () => void;
    onHide: () => void;
  } = $props();

  // The deck as the player has it on this screen. It follows the
  // server's whenever the server's changes (a swap that landed, or the
  // same seat sideboarding in another tab).
  const serverMain = $derived(mainOf(pool));
  const serverKey = $derived([...serverMain].sort().join(","));
  let main = $state<string[]>([]);
  $effect(() => {
    void serverKey;
    main = untrack(() => serverMain);
  });

  const mainSet = $derived(new Set(main));
  const deckLines = $derived(poolLines(pool, mainSet, true));
  const sideLines = $derived(poolLines(pool, mainSet, false));
  const total = $derived(pool.deck.length + pool.sideboard.length);
  const deckCount = $derived(main.length);
  const sideCount = $derived(total - main.length);
  const problem = $derived(deckProblem(deckCount, sideCount, sideboarding));
  const unsent = $derived(!sameDeck(main, serverMain));
  const changed = $derived(!sameDeck(main, pool.registered));

  const ready = $derived(sideboarding.ready.includes(viewerSeat));
  const chooser = $derived(sideboarding.chooser_seat === viewerSeat);
  const chooserName = $derived(
    seats.find((s) => s.seat === sideboarding.chooser_seat)?.name ?? "The other player",
  );
  const playFirst = $derived(sideboarding.play_first !== false);
  const nextGame = $derived(match.game + 1);

  function move(line: PoolLine): void {
    if (ready) return;
    // The last copy listed, so the first click on a line of four takes
    // the same card a second click would not put back by accident.
    const id = line.ids[line.ids.length - 1];
    apply(moveCard(main, id));
  }

  function apply(next: string[]): void {
    main = next;
    const side = total - next.length;
    if (deckProblem(next.length, side, sideboarding) === "" && !sameDeck(next, serverMain)) {
      onSetDeck(next);
    }
  }

  let confirmConcede = $state(false);
</script>

<ModalLayer />
<div class="prompt-backdrop" role="dialog" aria-modal="true" aria-label={L.sideboarding}>
  <div class="prompt-modal sideboarding" data-testid="sideboarding">
    <h2>
      Sideboarding
      <span class="prompt-src">{gameOfText({ ...match, game: nextGame })} is next</span>
    </h2>

    <div class="status">
      <span class="result">{lastGameText(match, seats)}</span>
      <span class="score">{scoreLine(match, seats)}</span>
      <span class="clock-slot"><MatchClock {sideboarding} label="time left to sideboard" /></span>
    </div>

    <p class="prompt-hint">
      Move cards between your deck and your sideboard. The next game starts when both players are
      ready, or when the timer runs out, with your deck as it stands here. Your saved deck does not
      change: this lasts for the match.
    </p>

    <div class="lists">
      {#each [{ title: "Deck", count: deckCount, lines: deckLines, to: "sideboard", verb: "Take out", mark: "in" }, { title: "Sideboard", count: sideCount, lines: sideLines, to: "deck", verb: "Bring in", mark: "out" }] as side (side.title)}
        <section class="list" aria-label={`your ${side.title.toLowerCase()}`}>
          <h3>{side.title}<span class="count">{side.count}</span></h3>
          {#if side.lines.length === 0}
            <p class="empty">No cards.</p>
          {:else}
            <ul>
              {#each side.lines as line (line.name)}
                <li class:moved={line.moved > 0}>
                  <span class="n">{line.ids.length}×</span>
                  <span class="name">
                    {line.name}
                    {#if line.card.type_line}<span class="type">{line.card.type_line}</span>{/if}
                  </span>
                  {#if line.moved > 0}
                    <span class="tag mark">{line.moved} {side.mark}</span>
                  {/if}
                  <button
                    type="button"
                    class="ghost move"
                    disabled={ready}
                    aria-label={`move one ${line.name} to your ${side.to}`}
                    onclick={() => move(line)}>{side.verb}</button
                  >
                </li>
              {/each}
            </ul>
          {/if}
        </section>
      {/each}
    </div>

    <p class="deck-state" class:warn={problem !== "" || error !== ""} role="status">
      {#if problem}
        {problem} Until then the next game uses your last complete deck.
      {:else if error}
        {error}
      {:else if unsent}
        Saving your deck…
      {:else if changed}
        Your deck is saved for the next game.
      {:else}
        This is your registered deck.
      {/if}
    </p>

    <fieldset class="choice" disabled={!chooser || ready}>
      <legend>Who plays first</legend>
      {#if chooser}
        <p class="choice-note">You lost the last game, so you choose.</p>
        <label>
          <input
            type="radio"
            name="play-first"
            checked={playFirst}
            onchange={() => onChoosePlayFirst(true)}
          />
          Play first
        </label>
        <label>
          <input
            type="radio"
            name="play-first"
            checked={!playFirst}
            onchange={() => onChoosePlayFirst(false)}
          />
          Draw first
        </label>
      {:else}
        <p class="choice-note">
          {chooserName} chooses, and you will see the choice when the game starts.
        </p>
      {/if}
    </fieldset>

    <ul class="ready-list" aria-label="who is ready">
      {#each seats as seat (seat.id)}
        <li>
          <span class="seat-dot" style="background:{seatColor(seat.seat)}"></span>
          <b>{seat.name}</b>
          <span class="muted"
            >{sideboarding.ready.includes(seat.seat) ? "ready" : "sideboarding…"}</span
          >
        </li>
      {/each}
    </ul>

    <div class="actions">
      {#if confirmConcede}
        <span class="confirm-text">Concede the whole match?</span>
        <button type="button" class="danger" onclick={onConcedeMatch}>{L.concedeMatch}</button>
        <button type="button" class="ghost" onclick={() => (confirmConcede = false)}
          >Keep playing</button
        >
      {:else}
        <button type="button" class="ghost" onclick={() => (confirmConcede = true)}>
          {L.concedeMatch}…
        </button>
        <button type="button" class="ghost" onclick={onHide}>Look at the board</button>
        <button
          type="button"
          class="ghost"
          disabled={ready || !changed}
          onclick={() => apply([...pool.registered])}>Back to registered deck</button
        >
        <span class="spacer"></span>
        {#if ready}
          <button type="button" onclick={() => onReady(false)}>Not ready</button>
        {:else}
          <button
            type="button"
            class="primary"
            disabled={problem !== "" || unsent}
            onclick={() => onReady(true)}>{L.sideboardReady}</button
          >
        {/if}
      {/if}
    </div>
  </div>
</div>

<style>
  .sideboarding {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-height: min(92vh, 860px);
  }
  .status {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 6px 16px;
    font-size: 14px;
  }
  .score {
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
  }
  .clock-slot {
    margin-inline-start: auto;
  }
  .lists {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    min-height: 0;
    overflow: auto;
  }
  .list {
    min-width: 0;
  }
  h3 {
    margin: 0 0 6px;
    font-size: 14px;
    color: var(--fg);
  }
  .count {
    margin-inline-start: 0.5em;
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-weight: 400;
    font-variant-numeric: tabular-nums;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .list li {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 2px 0;
    font-size: 13px;
    border-bottom: 1px solid var(--border, transparent);
  }
  .list li.moved .name {
    color: var(--accent-strong);
  }
  .n {
    flex: none;
    width: 2.4em;
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .type {
    display: block;
    color: var(--fg-muted);
    font-size: 11px;
  }
  .mark {
    flex: none;
  }
  .move {
    flex: none;
    font-size: 12px;
    padding: 4px 10px;
  }
  .empty,
  .choice-note {
    margin: 0;
    color: var(--fg-muted);
    font-size: 13px;
  }
  .deck-state {
    margin: 0;
    font-size: 13px;
    color: var(--fg-muted);
  }
  .deck-state.warn {
    color: var(--danger);
  }
  .choice {
    margin: 0;
    padding: 8px 12px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px 18px;
    font-size: 13px;
  }
  .choice legend {
    padding: 0 6px;
    color: var(--fg-muted);
    font-size: 12px;
  }
  .choice label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .ready-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 18px;
    font-size: 13px;
  }
  .ready-list li {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .seat-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
  .muted {
    color: var(--fg-muted);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }
  .spacer {
    flex: 1;
  }
  .confirm-text {
    font-size: 13px;
    margin-inline-end: auto;
  }
</style>
