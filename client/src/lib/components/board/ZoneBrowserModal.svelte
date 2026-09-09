<script lang="ts">
  // ZoneBrowserModal is the S18.5 follow-up to the S06 promise ("clickable
  // modal browser" for graveyard / exile / command). Clicking a pile in
  // the bottom-left chip row (PileBar) pops this up for the selected
  // zone so every seated player can inspect the cards — graveyard and
  // exile are public, command is public, and the stack is public too.
  //
  // Owner affordances: when the viewer owns the zone (owner ==
  // controller of every card in it per MTG zone semantics), each card
  // gets a small action button cluster (hand / battlefield / library).
  // Non-owners see a view-only grid. All moves go through the existing
  // `move_card` action — the server gates on controller match, so we
  // don't need a new action kind.
  //
  // Hover-zoom works inside the modal because Card.svelte already
  // writes to the shared hoveredCard store on mouseenter; the existing
  // HoverZoomOverlay mounted by Board.svelte picks it up unchanged.

  import type { ActionPayload, ActionType, CardView, GameView } from "../../protocol";
  import Card from "./Card.svelte";
  import type { BrowsableZone } from "../../zoneBrowser";
  import {
    buildMovePayload,
    canManageZone,
    cardsForZone,
    impulseActionLabel,
    impulseGrantFor,
  } from "../../zoneBrowser.logic";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    view: GameView;
    viewerID: string | null;
    zoneKind: BrowsableZone;
    ownerSeat: { id: string; name: string };
    sendAction: ActionSender;
    onClose: () => void;
    // S20: while a targeting prompt is live, clicking a browsed card
    // (a graveyard card for Eternal Witness) offers it as the
    // target. Returns true when the click was consumed.
    onTargetCard?: (card: CardView) => boolean;
  }

  const { view, viewerID, zoneKind, ownerSeat, sendAction, onClose, onTargetCard }: Props =
    $props();

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

  // canManage is true when the viewer is the owner of the zone;
  // stack is always false. Drives whether the per-card action
  // cluster renders. Kept as a $derived so the modal reacts if the
  // viewer's identity ever changes mid-modal (defence in depth;
  // unlikely in practice).
  const canManage = $derived(canManageZone(zoneKind, viewerID, ownerSeat.id));

  // S21 sub-PR 6: impulse exile. The button is keyed on the grant
  // rather than on zone ownership — the stolen card sits in the
  // victim's exile slice, and the thief is the one who may play it.
  // Both derivations live in zoneBrowser.logic.ts so vitest can
  // exercise them without a renderer.
  const grantFor = (card: CardView) => impulseGrantFor(card, zoneKind, viewerID);
  const labelFor = (card: CardView) => impulseActionLabel(card, zoneKind, viewerID);

  function playFromExile(card: CardView): void {
    sendAction(
      "cast_spell",
      { instance_id: card.instance_id, from_zone: "exile" },
      viewerID ?? undefined,
    );
    onClose();
  }

  // move fires move_card from the source zone to the chosen
  // destination. buildMovePayload returns null when the client-side
  // guard rejects the action (non-owner, stack, controller mismatch);
  // the server gates independently, so the check here is for UX
  // cleanliness (no speculative action frames).
  function move(card: CardView, dest: "hand" | "battlefield" | "library"): void {
    if (!viewerID) return;
    const payload = buildMovePayload(zoneKind, ownerSeat.id, viewerID, card, dest);
    if (!payload) return;
    sendAction("move_card", payload, viewerID);
  }

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
            <Card {card} onClick={onTargetCard ? () => void onTargetCard?.(card) : undefined} />
            {#if labelFor(card)}
              <div class="actions" aria-label="play from exile">
                <button
                  type="button"
                  class="act impulse"
                  title={grantFor(card)?.any_color
                    ? "spend mana as though it were any colour"
                    : "playable until end of turn"}
                  aria-label={`${labelFor(card)} ${card.name || "card"} from exile`}
                  onclick={() => playFromExile(card)}
                >
                  {labelFor(card)}
                </button>
              </div>
            {/if}
            {#if canManage}
              <div class="actions" aria-label="move card">
                <button
                  type="button"
                  class="act"
                  title="to hand"
                  aria-label={`move ${card.name || "card"} to hand`}
                  onclick={() => move(card, "hand")}
                >
                  hand
                </button>
                <button
                  type="button"
                  class="act"
                  title="to battlefield"
                  aria-label={`move ${card.name || "card"} to battlefield`}
                  onclick={() => move(card, "battlefield")}
                >
                  field
                </button>
                <button
                  type="button"
                  class="act"
                  title="to top of library"
                  aria-label={`move ${card.name || "card"} to library`}
                  onclick={() => move(card, "library")}
                >
                  lib
                </button>
              </div>
            {/if}
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
  .act.impulse {
    border-color: var(--gold, #c9a227);
    color: var(--gold, #c9a227);
  }
  .actions {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
    justify-content: center;
  }
  .act {
    background: rgba(255, 255, 255, 0.04);
    color: var(--fg);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 999px;
    padding: 3px 9px;
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    cursor: pointer;
    font-family: inherit;
    font-weight: 600;
    transition:
      background 120ms var(--ease),
      border-color 120ms var(--ease),
      color 120ms var(--ease);
  }
  .act:hover {
    background: rgba(122, 167, 255, 0.12);
    border-color: var(--accent);
    color: var(--accent);
  }
  .act:focus-visible {
    outline: 1px solid var(--accent);
    outline-offset: 1px;
  }
</style>
