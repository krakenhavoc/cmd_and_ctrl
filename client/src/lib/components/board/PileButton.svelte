<script lang="ts">
  // PileButton is the small stacked-corner control for the four
  // pile-style zones each player owns: EXILE, GRAVEYARD, DECK
  // (library), CMD ZONE. The face is a label + count plus an
  // optional thumbnail of the top card (used by GRAVEYARD where
  // the top card is public). DECK uses a stylised back; EXILE and
  // CMD ZONE show counts only at v1.
  //
  // Click is dispatched up to the parent so PileBar can route it:
  // the library opens its draw / shuffle menu (#2962), GRAVEYARD and
  // EXILE open the zone browser.

  import type { CardView, ZoneView } from "../../protocol";
  import Card from "./Card.svelte";

  interface Props {
    label: string;
    zone: ZoneView;
    // faceDown: render the pile as a card back (LIBRARY) regardless
    // of whether the top card is technically known. GRAVEYARD / EXILE
    // are face-up; CMD ZONE is face-up too (commanders are public).
    faceDown?: boolean;
    disabled?: boolean;
    onClick?: () => void;
    // ADR 0105 (#1789): how many cards in this pile the viewer can do
    // something with right now (legalActions.ts readyCount). Zero —
    // the default, and what every pile shows while highlights are off
    // — draws nothing.
    readyCount?: number;
    // ADR 0119 §3: which of the seat's piles this is ("library",
    // "graveyard", "exile") and whose, as `data-pile` and
    // `data-pile-owner`, so a card leaving the stack can fly to it.
    pile?: string;
    owner?: string;
    // #2559: a short line under the count for what the pile holds that
    // is not the viewer's to act on — "7 theirs to play" on a seat whose
    // own exiled cards it may play (Memory Vessel). Also read out.
    note?: string;
    // #2962: the click opens a menu (the library's draw / shuffle).
    haspopup?: boolean;
    expanded?: boolean;
  }

  const {
    label,
    zone,
    faceDown = false,
    disabled = false,
    onClick,
    readyCount = 0,
    pile,
    owner,
    note,
    haspopup = false,
    expanded = false,
  }: Props = $props();

  const topCard = $derived(zone.cards.length > 0 ? zone.cards[zone.cards.length - 1] : null);

  // Synthetic CardView for the face-down render path. Card.svelte
  // ignores name / scryfall_id / etc. when faceDown=true, so the
  // placeholder is purely structural. Derived because `label` is a
  // reactive prop — wrapping it in $derived keeps the instance_id
  // stable across renders for the same pile.
  const backPlaceholder = $derived<CardView>({
    instance_id: `${label}-back`,
    name: "",
    owner: "",
    controller: "",
  });
</script>

<button
  type="button"
  class="pile"
  class:disabled
  class:has-cards={zone.count > 0}
  class:ready={readyCount > 0}
  {disabled}
  data-pile={pile}
  data-pile-owner={owner}
  onclick={onClick}
  aria-haspopup={haspopup ? "menu" : undefined}
  aria-expanded={haspopup ? expanded : undefined}
  aria-label={`${label}: ${zone.count} card${zone.count === 1 ? "" : "s"}${readyCount > 0 ? `, ${readyCount} ready` : ""}${note ? `, ${note}` : ""}`}
  title={`${label} · ${zone.count}${readyCount > 0 ? ` · ${readyCount} ready` : ""}${note ? ` · ${note}` : ""}`}
>
  <span class="thumb">
    {#if faceDown && zone.count > 0}
      <Card card={backPlaceholder} faceDown />
    {:else if topCard && !faceDown}
      <Card card={topCard} />
    {:else}
      <span class="empty" aria-hidden="true"></span>
    {/if}
  </span>
  <span class="meta">
    <span class="label">{label}</span>
    <span class="count">{zone.count}</span>
    {#if readyCount > 0}
      <span class="ready-count" aria-hidden="true">{readyCount} ready</span>
    {/if}
    {#if note}
      <span class="note" aria-hidden="true">{note}</span>
    {/if}
  </span>
</button>

<style>
  .pile {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    gap: 4px;
    padding: 5px 3px 4px;
    background: var(--surface-raised);
    border: 1px solid transparent;
    border-radius: var(--radius);
    color: inherit;
    font: inherit;
    cursor: pointer;
    width: 100%;
    box-sizing: border-box;
    box-shadow: none;
    transition:
      border-color 140ms var(--ease),
      background 140ms var(--ease);
  }
  .pile:disabled,
  .pile.disabled {
    cursor: default;
    opacity: 0.55;
  }
  .pile:not(:disabled):hover {
    border-color: color-mix(in srgb, var(--accent) 45%, transparent);
    background: var(--surface-hover);
    transform: none;
  }
  .pile:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .thumb {
    width: var(--thumb-w, 50px);
    height: var(--thumb-h, 70px);
    display: flex;
    align-items: center;
    justify-content: center;
    --card-w: var(--thumb-w, 50px);
    --card-h: var(--thumb-h, 70px);
  }
  .empty {
    width: 100%;
    height: 100%;
    border-radius: 4px;
    border: 1px dashed var(--border-strong);
    background: rgba(0, 0, 0, 0.25);
  }
  .meta {
    display: flex;
    flex-direction: column;
    align-items: center;
    line-height: 1.2;
    gap: 2px;
  }
  /* Sentence case from the lower-case pile names, which stay in the
     button's accessible name. */
  .label {
    display: inline-block;
    font-family: var(--font-ui);
    font-size: 10px;
    color: var(--fg-dim);
    font-weight: 600;
  }
  .label::first-letter {
    text-transform: uppercase;
  }
  /* ADR 0105: "N ready" in the one ready colour. A count, not a ring:
     the thumb shows one card and the ready ones may be under it. */
  .pile.ready {
    border-color: var(--ready-soft);
  }
  .ready-count {
    font-family: var(--font-mono);
    font-size: 8px;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    line-height: 1;
    padding: 1px 4px;
    border-radius: 999px;
    color: var(--ready-ink);
    background: var(--ready);
    white-space: nowrap;
  }
  /* #2559: what the pile holds for its own seat, not the viewer — a
     quiet outline rather than the ready pill. */
  .note {
    font-family: var(--font-mono);
    font-size: 8px;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    line-height: 1;
    padding: 1px 4px;
    border-radius: 999px;
    color: var(--fg-muted);
    border: 1px solid var(--border-strong);
    white-space: nowrap;
  }
  .count {
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 700;
    color: var(--fg);
    font-variant-numeric: tabular-nums;
  }
</style>
