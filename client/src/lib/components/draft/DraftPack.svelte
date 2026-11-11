<script lang="ts">
  // DraftPack — the pack the viewer holds, and the pick (ADR 0222 §5).
  //
  // Each card is a real toggle button named for the card. A click
  // chooses it; the bar's primary sends the pick once exactly `take`
  // cards are chosen. With one card to take, a double-click picks at
  // once, and Enter on the chosen card confirms it.
  //
  // The component reads `pack` and nothing else about cards: there is
  // nothing else on the wire (CR 905.1c).
  import type { Snippet } from "svelte";
  import { untrack } from "svelte";

  import {
    canConfirm,
    cardAccessibleName,
    pickCount,
    pruneSelection,
    toggleSelection,
    type DraftCard,
  } from "../../draft";
  import { L } from "../../labels";
  import Icon from "../Icon.svelte";
  import DraftCardArt from "./DraftCardArt.svelte";

  interface Props {
    pack: DraftCard[];
    /** How many cards to take: 1 or 2. */
    take: number;
    /** A pick is on its way, or the socket is down. */
    busy?: boolean;
    /** Why the primary cannot be pressed right now, shown in the bar. */
    blocked?: string;
    onpick: (cards: string[]) => void;
    /** The pick timer, drawn at the start of the bar. */
    timer?: Snippet;
  }
  const { pack, take, busy = false, blocked = "", onpick, timer }: Props = $props();

  let selected = $state<string[]>([]);

  // A new pack, or a smaller `take`, drops whatever no longer applies.
  $effect(() => {
    const kept = pruneSelection(
      untrack(() => selected),
      pack,
      take,
    );
    if (kept.length !== untrack(() => selected.length)) selected = kept;
  });

  const ready = $derived(canConfirm(selected, pack, take));
  const chosen = $derived(
    selected.map((id) => pack.find((c) => c.id === id)?.name ?? "").filter(Boolean),
  );
  const prompt = $derived.by(() => {
    if (blocked) return blocked;
    if (ready) return chosen.join(" and ");
    const more = take - selected.length;
    if (selected.length === 0) {
      return take === 1 ? "Choose a card." : `Choose ${take} cards.`;
    }
    return `${chosen.join(" and ")}. Choose ${more} more.`;
  });

  function toggle(id: string): void {
    if (busy) return;
    selected = toggleSelection(selected, id, take);
  }

  function confirm(): void {
    if (busy || !ready) return;
    onpick([...selected]);
  }

  // With one card to take: a double-click picks it outright.
  function pickNow(id: string): void {
    if (busy || take !== 1) return;
    selected = [id];
    onpick([id]);
  }

  // Enter on the card already chosen confirms it. Enter on any other
  // card is the button's own click, which chooses it.
  function onKey(e: KeyboardEvent, id: string): void {
    if (e.key !== "Enter" || take !== 1 || busy) return;
    if (selected.length === 1 && selected[0] === id) {
      e.preventDefault();
      confirm();
    }
  }
</script>

<div class="pack-wrap">
  <ul class="pack" role="group" aria-label={L.draftPack}>
    {#each pack as card (card.id)}
      {@const on = selected.includes(card.id)}
      <li>
        <button
          type="button"
          class="pick-card"
          class:on
          aria-pressed={on}
          aria-label={cardAccessibleName(card)}
          disabled={busy}
          onclick={() => toggle(card.id)}
          ondblclick={() => pickNow(card.id)}
          onkeydown={(e) => onKey(e, card.id)}
        >
          <DraftCardArt {card} />
          {#if on}
            <span class="mark" aria-hidden="true"><Icon name="check" size={14} /></span>
          {/if}
        </button>
      </li>
    {/each}
  </ul>

  <!-- Sticks to the foot of the window while the pack scrolls, so the
       timer and the pick are always in reach. z 5: above the cards,
       under the site header (30) and everything that floats. -->
  <div class="bar">
    {@render timer?.()}
    <p class="prompt" class:warn={blocked !== ""}>{prompt}</p>
    <button type="button" class="btn primary" disabled={busy || !ready} onclick={confirm}>
      {pickCount(take)}
    </button>
  </div>
  {#if take === 1}
    <p class="tip">Double-click a card to pick it straight away.</p>
  {/if}
</div>

<style>
  .pack-wrap {
    min-width: 0;
  }
  .pack {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    /* Big enough to read a card: five across at the 1200px page, three at 820px. */
    grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
    gap: 14px;
  }
  .pick-card {
    position: relative;
    display: block;
    width: 100%;
    height: auto;
    padding: 0;
    border: 0;
    border-radius: 4.75% / 3.5%;
    background: none;
    cursor: pointer;
    transition:
      transform 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  @media (hover: hover) {
    .pick-card:hover:not(:disabled) {
      transform: translateY(-3px);
      box-shadow: 0 14px 28px -14px var(--shadow-ink);
    }
  }
  .pick-card:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 3px;
  }
  /* Chosen: the pick signal as a ring, and a check in the corner, so it
     is never colour alone. The ring is inset from the focus outline. */
  .pick-card.on {
    box-shadow: 0 0 0 3px var(--pick);
    transform: translateY(-3px);
  }
  .pick-card:disabled {
    cursor: default;
    opacity: 0.6;
  }
  .mark {
    position: absolute;
    top: 8px;
    left: 8px;
    display: grid;
    place-items: center;
    width: 24px;
    height: 24px;
    border-radius: 999px;
    background: var(--pick);
    color: var(--bg);
    box-shadow: var(--shadow-sm);
  }
  .bar {
    position: sticky;
    bottom: 12px;
    z-index: 5;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px 16px;
    margin-top: 16px;
    padding: 12px 16px;
    border-radius: var(--radius-lg);
    background: var(--surface-raised);
    border: 1px solid var(--border-strong);
    box-shadow: var(--shadow);
  }
  .prompt {
    flex: 1 1 20ch;
    min-width: 0;
    margin: 0;
    font-size: 0.875rem;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .prompt.warn {
    color: var(--fg-muted);
  }
  .tip {
    margin: 10px 0 0;
    font-size: 0.8125rem;
    color: var(--fg-muted);
  }
</style>
