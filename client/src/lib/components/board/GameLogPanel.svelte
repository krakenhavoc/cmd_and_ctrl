<script lang="ts">
  // GameLogPanel — the public game log (S31 sub-PR 0, ADR 0033 §4).
  //
  // A right-side drawer over the table. Reads GameView.log, which the
  // server has already rendered and already redacted for this viewer:
  // an entry that says "a card" is one this seat is not entitled to
  // see named, and this component deliberately does NOT try to
  // recover the name from the board. Rendering is grouping, colour,
  // and scroll position — nothing else.
  //
  // Newest at the bottom, the way a table log reads. The drawer
  // auto-follows the tail unless the reader has scrolled up, which is
  // the one piece of state worth keeping: a player scrolling back to
  // check what happened last turn must not be yanked forward by the
  // next frame.

  import type { GameView } from "../../protocol";
  import { filterBySeat, groupLog, logTone, seatName } from "../../gameLog";
  import { seatColor } from "../../colors";
  import Icon from "../Icon.svelte";

  interface Props {
    view: GameView;
    viewerID: string | null;
    onClose: () => void;
  }

  const { view, viewerID, onClose }: Props = $props();

  const viewerSeat = $derived(view.seats.findIndex((s) => s.id === viewerID));
  let mineOnly = $state(false);
  const filterSeat = $derived(mineOnly && viewerSeat >= 0 ? viewerSeat : null);
  const groups = $derived(groupLog(filterBySeat(view.log ?? [], filterSeat)));
  const total = $derived(view.log?.length ?? 0);

  // Tail-following. `atBottom` is recomputed on every scroll; the
  // effect below only scrolls when it is true, so a reader parked
  // mid-history stays parked.
  let body = $state<HTMLDivElement | null>(null);
  let atBottom = $state(true);

  function onScroll(): void {
    if (!body) return;
    atBottom = body.scrollHeight - body.scrollTop - body.clientHeight < 24;
  }

  $effect(() => {
    // Touch the entry count so this reruns on every new frame.
    void total;
    void groups.length;
    if (body && atBottom) body.scrollTop = body.scrollHeight;
  });
</script>

<aside class="log-panel" aria-label="game log">
  <header class="log-head">
    <span class="log-title"><Icon name="scroll" size={14} /> game log</span>
    <span class="muted log-count">{total} entries</span>
    {#if viewerSeat >= 0}
      <label class="log-filter" title="show only entries involving your seat">
        <input type="checkbox" bind:checked={mineOnly} />
        mine
      </label>
    {/if}
    <button type="button" class="log-close" onclick={onClose} aria-label="close game log">
      <Icon name="x" size={14} />
    </button>
  </header>

  <div class="log-body" bind:this={body} onscroll={onScroll}>
    {#if groups.length === 0}
      <p class="muted log-empty">Nothing has happened yet.</p>
    {/if}
    {#each groups as group (group.key)}
      <section class="log-group">
        {#if group.header}
          <h4 class="log-header" style="--seat:{group.seat >= 0 ? seatColor(group.seat) : '#888'}">
            {group.header}
          </h4>
        {:else}
          <h4 class="log-header earlier" style="--seat:#888">earlier</h4>
        {/if}
        <ol class="log-entries">
          {#each group.entries as entry (entry.seq)}
            <li class="log-entry {logTone(entry.kind)}">
              <span
                class="log-dot"
                style="background:{entry.seat >= 0 ? seatColor(entry.seat) : 'transparent'}"
                aria-hidden="true"
              ></span>
              <span class="log-text">{entry.text}</span>
              {#if entry.target_seat !== undefined && seatName(view, entry.target_seat)}
                <span
                  class="log-target"
                  style="color:{seatColor(entry.target_seat)}"
                  aria-hidden="true"
                >
                  ●
                </span>
              {/if}
            </li>
          {/each}
        </ol>
      </section>
    {/each}
  </div>
</aside>

<style>
  .log-panel {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    width: min(380px, 92vw);
    display: flex;
    flex-direction: column;
    background: #14161c;
    border-left: 1px solid #2a2f3a;
    box-shadow: -12px 0 32px rgb(0 0 0 / 45%);
    z-index: 60;
  }

  .log-head {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.6rem 0.7rem;
    border-bottom: 1px solid #2a2f3a;
    font-size: 0.85rem;
  }

  .log-title {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  .log-count {
    margin-left: auto;
    font-size: 0.75rem;
  }

  .log-filter {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.75rem;
    color: #9aa3b2;
    cursor: pointer;
  }

  .log-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.2rem;
    background: none;
    border: none;
    color: #9aa3b2;
    cursor: pointer;
  }

  .log-close:hover {
    color: #e6e9ef;
  }

  .log-body {
    flex: 1;
    overflow-y: auto;
    padding: 0.4rem 0.6rem 1.2rem;
  }

  .log-empty {
    padding: 1rem 0.2rem;
    font-size: 0.8rem;
  }

  .log-header {
    position: sticky;
    top: 0;
    margin: 0.7rem 0 0.25rem;
    padding: 0.2rem 0.35rem;
    background: #14161c;
    border-left: 3px solid var(--seat);
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: lowercase;
    color: #c7cedb;
  }

  .log-header.earlier {
    color: #7b8494;
    font-style: italic;
  }

  .log-entries {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .log-entry {
    display: flex;
    align-items: baseline;
    gap: 0.4rem;
    padding: 0.12rem 0.35rem;
    font-size: 0.78rem;
    line-height: 1.35;
    color: #c3cad6;
    border-left: 2px solid transparent;
  }

  .log-entry:hover {
    background: #1b1e26;
  }

  .log-dot {
    flex: none;
    width: 5px;
    height: 5px;
    border-radius: 50%;
    transform: translateY(-1px);
  }

  .log-text {
    flex: 1;
  }

  .log-target {
    flex: none;
    font-size: 0.6rem;
  }

  /* Tone accents — one per LOG_TONE value in gameLog.ts. */
  .tone-step {
    color: #c7cedb;
  }

  .tone-cast {
    color: #93c5fd;
  }

  .tone-resolve {
    color: #a5b4fc;
  }

  .tone-zone {
    color: #c3cad6;
  }

  .tone-quiet {
    color: #8b94a3;
  }

  .tone-life {
    color: #86efac;
  }

  .tone-damage {
    color: #fca5a5;
  }

  .tone-combat {
    color: #fcd34d;
  }

  .tone-bad {
    color: #f0abfc;
  }

  .muted {
    color: #7b8494;
  }
</style>
