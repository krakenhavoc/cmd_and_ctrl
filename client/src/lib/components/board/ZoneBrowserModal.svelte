<script lang="ts">
  // ZoneBrowserModal is the S18.5 follow-up to the S06 promise ("clickable
  // modal browser" for graveyard / exile / command). Clicking a pile in
  // the bottom-left chip row (PileBar) pops this up for the selected
  // zone so every seated player can inspect the cards — graveyard and
  // exile are public, command is public, and the stack is public too.
  //
  // Milestone 1: view-only browsable grid with hover-zoom. Owner
  // action affordances (move card to hand / battlefield / library)
  // land in milestone 2.
  //
  // Hover-zoom works inside the modal because Card.svelte already
  // writes to the shared hoveredCard store on mouseenter; the existing
  // HoverZoomOverlay mounted by Board.svelte picks it up unchanged.

  import type { CardView, GameView } from "../../protocol";
  import Card from "./Card.svelte";
  import type { BrowsableZone } from "../../zoneBrowser";
  import { cardsForZone } from "../../zoneBrowser.logic";

  interface Props {
    view: GameView;
    viewerID: string | null;
    zoneKind: BrowsableZone;
    ownerSeat: { id: string; name: string };
    onClose: () => void;
  }

  // viewerID is accepted for forward-compatibility with the M2 owner
  // action affordances; it's unused in the view-only M1 path.
  const { view, viewerID: _viewerID, zoneKind, ownerSeat, onClose }: Props = $props();

  // Source zone lookup — server broadcasts exile + stack as shared
  // top-level zones with per-card owner/controller, while graveyard
  // and command live on the PlayerView. The modal filters the shared
  // zones down to the chosen owner so the viewer sees the same slice
  // they clicked on. The filter lives in zoneBrowser.logic.ts so
  // vitest can exercise it without rendering the component.
  const zoneCards = $derived<CardView[]>(cardsForZone(view, zoneKind, ownerSeat.id));

  // zoneLabel and zoneDescription drive the header copy.
  const zoneLabel = $derived.by(() => {
    switch (zoneKind) {
      case "graveyard":
        return "graveyard";
      case "exile":
        return "exile";
      case "command":
        return "command zone";
      case "stack":
        return "stack";
    }
  });

  function backdropClick(ev: MouseEvent): void {
    if (ev.target === ev.currentTarget) onClose();
  }

  function onKey(ev: KeyboardEvent): void {
    if (ev.key === "Escape") {
      ev.stopPropagation();
      onClose();
    }
  }
</script>

<svelte:window on:keydown={onKey} />

<div
  class="backdrop"
  role="dialog"
  aria-modal="true"
  aria-labelledby="zone-browser-title"
  tabindex="-1"
  onclick={backdropClick}
  onkeydown={onKey}
>
  <div class="modal">
    <header>
      <h2 id="zone-browser-title">
        {ownerSeat.name}<span class="sep">·</span><span class="zone">{zoneLabel}</span>
        <span class="count">({zoneCards.length})</span>
      </h2>
      <button type="button" class="close" onclick={onClose} aria-label="close">×</button>
    </header>
    {#if zoneCards.length === 0}
      <p class="empty">No cards in this zone.</p>
    {:else}
      <ul class="grid">
        {#each zoneCards as card (card.instance_id)}
          <li class="cell">
            <Card {card} />
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(4, 8, 16, 0.7);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 190;
    animation: fade-in 160ms var(--ease);
  }
  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  .modal {
    background: linear-gradient(180deg, var(--surface) 0%, var(--bg-2) 100%);
    border: 1px solid rgba(122, 167, 255, 0.22);
    border-radius: var(--radius-xl, 14px);
    padding: 18px 22px 22px;
    min-width: 480px;
    max-width: 90vw;
    max-height: 86vh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    box-shadow:
      0 30px 80px rgba(0, 0, 0, 0.7),
      0 0 0 1px rgba(0, 0, 0, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
    animation: modal-in 220ms var(--ease);
  }
  @keyframes modal-in {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    gap: 12px;
  }
  h2 {
    margin: 0;
    font-size: 16px;
    letter-spacing: 0.02em;
    color: var(--fg);
    text-transform: none;
    font-weight: 700;
    display: flex;
    align-items: baseline;
    gap: 8px;
  }
  .sep {
    color: var(--fg-muted);
    font-weight: 400;
    opacity: 0.6;
  }
  .zone {
    color: var(--gold);
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-size: 12px;
    font-weight: 700;
  }
  .count {
    color: var(--fg-muted);
    font-weight: 500;
    font-variant-numeric: tabular-nums;
    font-size: 13px;
  }
  .close {
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: var(--fg);
    width: 28px;
    height: 28px;
    border-radius: 50%;
    font-size: 18px;
    line-height: 1;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
  }
  .close:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.24);
  }
  .empty {
    color: var(--fg-muted);
    font-size: 13px;
    margin: 8px 0 0;
  }
  .grid {
    list-style: none;
    padding: 4px;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 14px;
    overflow: auto;
    /* The Card inside uses --card-w/--card-h; we bump them up from
       the pile thumb size so the modal grid feels like a browsable
       catalogue rather than a list of chips. */
    --card-w: 110px;
    --card-h: 154px;
  }
  .cell {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
  }
</style>
