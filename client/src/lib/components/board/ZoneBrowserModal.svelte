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
  import { openZoneBrowser, type BrowsableZone } from "../../zoneBrowser";
  import Icon from "../Icon.svelte";
  import {
    buildMovePayload,
    canManageZone,
    cardsForZone,
    castableFromZone,
    grantedFace,
    grantedFaceIndex,
    impulseActionLabel,
    impulseGrantFor,
  } from "../../zoneBrowser.logic";
  import { printedCostClaimable, type CastSourceZone } from "../../targeting";
  import ModalLayer from "../ModalLayer.svelte";

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
    // S29: a flashback / escape card in the viewer's own graveyard
    // casts through the Board's ordinary prompt chain — cost picker,
    // additional cost, X, modes, targeting — rather than firing a
    // bare payload. The modal hands the card up with the zone it came
    // out of and closes; everything after that is the same code path
    // a hand cast takes.
    //
    // #874: the exile impulse button goes through it too, and passes
    // `face` when the grant names one. It used to dispatch cast_spell
    // itself, which meant an {X} spell granted for its PRINTED cost
    // could only ever be cast for X = 0 and a targeted one got no
    // targets — the server was being asked to validate an
    // announcement the player was never given the chance to make.
    onCastCard?: (card: CardView, fromZone: CastSourceZone, face?: number) => void;
  }

  const {
    view,
    viewerID,
    zoneKind,
    ownerSeat,
    sendAction,
    onClose,
    onTargetCard,
    onCastCard,
  }: Props = $props();

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
  //
  // S29 warp added a floor to the window: a warped creature's grant
  // is stamped the moment the end step exiles it and stays dark
  // until the next turn, so the turn number rides both derivations.
  const grantFor = (card: CardView) => impulseGrantFor(card, zoneKind, viewerID, view.turn.number);
  // #874: the impulse button is gated on `onCastCard` for the reason
  // the graveyard one is — the click is a hand-off to the Board's cast
  // chain now, not a dispatch of its own, so without a handler there
  // is nothing behind the button.
  const labelFor = (card: CardView) =>
    onCastCard === undefined
      ? null
      : impulseActionLabel(card, zoneKind, viewerID, view.turn.number);
  // S32: the name of the half the grant actually casts. Same as the
  // card's own name for every grant that names no face.
  const grantedName = (card: CardView) => grantedFace(card, grantFor(card)).name ?? "card";

  // S29: "cast from here" for the zones whose permission is printed
  // on the card rather than granted to an instance. Only the
  // graveyard today; the gate lives in zoneBrowser.logic.ts so
  // vitest can exercise it without a renderer.
  const castableFor = (card: CardView) =>
    castableFromZone(card, zoneKind, viewerID, ownerSeat.id) && onCastCard !== undefined;

  // The label is the printed clause when the card offers exactly one
  // way in ("Flashback {2}{R}"), so the button reads like the card.
  // Two or more ways in, or none, fall back to the verb — the Board's
  // picker is about to ask anyway.
  //
  // #1012: "one way in" counts the PRINTED cost too. A Gravecrawler
  // under an Underworld Breach has one offer and the printed cost
  // beside it, and labelling that button "Escape" would name a price
  // the player has not chosen yet.
  const castLabelFor = (card: CardView) =>
    card.alternative_costs?.length === 1 && !printedCostClaimable(card)
      ? card.alternative_costs[0].label || "cast"
      : "cast";

  function castFromZone(card: CardView): void {
    if (!onCastCard || zoneKind !== "graveyard") return;
    onCastCard(card, "graveyard");
    onClose();
  }

  // playFromExile hands an impulse cast to the Board's one cast entry
  // point, exactly as the graveyard button does (#874). It used to
  // dispatch `cast_spell` here, which skipped every announce-time
  // prompt: a grant that lets you cast for the PRINTED cost was
  // therefore always announced with X = 0, and a spell with targets
  // or modes with none at all.
  //
  // The grant's face rides along. `grantedFace` is not used for the
  // hand-off, because the chain wants the raw card plus the index —
  // `cardAsFace` is applied once, by the Board, for the face picker
  // and for this on the same line.
  function playFromExile(card: CardView): void {
    const grant = grantFor(card);
    if (!onCastCard || !grant) return;
    onCastCard(card, "exile", grantedFaceIndex(grant));
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

  // Tabs switch zones for the same owner without closing (the stack
  // is a shared zone with no owner, so it stays a single view).
  const tabs: { kind: BrowsableZone; label: string }[] = [
    { kind: "graveyard", label: "graveyard" },
    { kind: "exile", label: "exile" },
    { kind: "command", label: "command" },
  ];
  const showTabs = $derived(zoneKind !== "stack");
  function switchZone(kind: BrowsableZone): void {
    if (kind === zoneKind) return;
    openZoneBrowser({ zoneKind: kind, ownerID: ownerSeat.id, ownerName: ownerSeat.name });
  }

  let query = $state("");
  const shown = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) return zoneCards;
    return zoneCards.filter((c) => (c.name ?? "").toLowerCase().includes(q));
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

<ModalLayer />

<div
  class="prompt-backdrop"
  role="dialog"
  aria-modal="true"
  aria-labelledby="zone-browser-title"
  tabindex="-1"
  onclick={backdropClick}
  onkeydown={onKey}
>
  <div class="prompt-modal zb-modal">
    <header class="zb-head">
      <h2 id="zone-browser-title">
        {ownerSeat.name}<span class="sep">·</span><span class="zone">{zoneLabel}</span>
        <span class="count">({zoneCards.length})</span>
        {#if !canManage && zoneKind !== "stack"}
          <span class="prompt-src" aria-hidden="true">read only</span>
        {/if}
      </h2>
      <button type="button" class="ghost close" onclick={onClose} aria-label="close">
        <Icon name="x" size={14} />
      </button>
    </header>
    <div class="zb-tools">
      {#if showTabs}
        <div class="tabs" role="tablist" aria-label="zone">
          {#each tabs as t (t.kind)}
            <button
              type="button"
              class="tab"
              class:on={t.kind === zoneKind}
              role="tab"
              aria-selected={t.kind === zoneKind}
              onclick={() => switchZone(t.kind)}
            >
              {t.label}
            </button>
          {/each}
        </div>
      {/if}
      <input
        class="search"
        type="text"
        placeholder="search this zone"
        aria-label="search cards"
        bind:value={query}
      />
    </div>
    {#if zoneCards.length === 0}
      <p class="empty">No cards in this zone.</p>
    {:else if shown.length === 0}
      <p class="empty">Nothing matches “{query}”.</p>
    {:else}
      <ul class="grid">
        {#each shown as card (card.instance_id)}
          <li class="cell">
            <Card {card} onClick={onTargetCard ? () => void onTargetCard?.(card) : undefined} />
            {#if labelFor(card)}
              <!-- The impulse grant is the one action that shouldn't wait
                   for a hover: the thief needs to see that the card is
                   theirs to play. -->
              <div class="actions always" aria-label="play from exile">
                <button
                  type="button"
                  class="act impulse"
                  title={grantFor(card)?.any_color
                    ? "spend mana as though it were any colour"
                    : "playable until end of turn"}
                  aria-label={`${labelFor(card)} ${grantedName(card)} from exile`}
                  onclick={() => playFromExile(card)}
                >
                  {labelFor(card)}
                  {#if grantedName(card) !== card.name}
                    <!-- S32: a defeated Siege sits in exile showing
                         the battle, and the grant casts its back
                         face. Naming it on the button is the only
                         place a player learns which card the click
                         actually produces. -->
                    <span class="granted-face">{grantedName(card)}</span>
                  {/if}
                </button>
              </div>
            {/if}
            {#if castableFor(card)}
              <!-- S29: same always-visible treatment as the impulse
                   button, and for the same reason — a flashback card
                   in the graveyard is a resource, and a player who
                   has to hover to discover that will not discover
                   it. -->
              <div class="actions always" aria-label="cast from graveyard">
                <button
                  type="button"
                  class="act impulse"
                  title="cast from your graveyard"
                  aria-label={`cast ${card.name || "card"} from graveyard`}
                  onclick={() => castFromZone(card)}
                >
                  {castLabelFor(card)}
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
  .zb-modal {
    min-width: min(560px, calc(100vw - 32px));
    max-width: min(920px, calc(100vw - 32px));
    overflow: hidden;
    gap: 10px;
  }
  .zb-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }
  .sep {
    color: var(--fg-dim);
    font-weight: 400;
  }
  .zone {
    color: var(--gold-strong);
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
  }
  .count {
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-weight: 500;
    font-size: 12px;
  }
  .close {
    width: 28px;
    height: 28px;
    padding: 0;
    border-radius: 50%;
    flex: 0 0 auto;
  }
  .zb-tools {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }
  .tabs {
    display: inline-flex;
    gap: 2px;
    padding: 3px;
    border-radius: 9px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
  }
  .tab {
    height: 26px;
    padding: 0 12px;
    border: 1px solid transparent;
    border-radius: 7px;
    background: transparent;
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    font-weight: 700;
  }
  .tab:hover {
    background: var(--surface-raised);
    color: var(--fg);
  }
  .tab.on {
    background: var(--surface-raised);
    border-color: var(--border-strong);
    color: var(--gold-strong);
  }
  .search {
    flex: 1 1 160px;
    min-width: 140px;
    margin: 0;
    padding: 5px 10px;
    font-size: 12.5px;
  }
  .empty {
    color: var(--fg-muted);
    font-size: 13px;
    margin: 4px 0 0;
  }
  .grid {
    list-style: none;
    padding: 4px 4px 8px;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 12px 10px;
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
    position: relative;
  }
  .act.impulse {
    border-color: rgba(217, 180, 92, 0.6);
    color: var(--gold-strong);
  }
  /* The back-face name on a "cast it transformed" grant. Rendered
     inside the pill rather than under it so the button stays one
     hit target, and truncated rather than wrapped so a long name
     cannot reflow the grid. */
  .granted-face {
    margin-left: 4px;
    text-transform: none;
    letter-spacing: 0;
    font-weight: 500;
    opacity: 0.85;
    max-width: 10ch;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: inline-block;
    vertical-align: bottom;
  }
  /* Moves ride under the hovered / focused card; other cells keep a
     spacer so the grid doesn't reflow. */
  .actions {
    display: flex;
    gap: 4px;
    justify-content: center;
    opacity: 0;
    transition: opacity 120ms var(--ease);
  }
  .cell:hover .actions,
  .cell:focus-within .actions,
  .actions.always {
    opacity: 1;
  }
  .act {
    height: 22px;
    padding: 0 8px;
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 9.5px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 600;
  }
  .act:hover {
    color: var(--gold-strong);
    border-color: rgba(217, 180, 92, 0.5);
  }
</style>
