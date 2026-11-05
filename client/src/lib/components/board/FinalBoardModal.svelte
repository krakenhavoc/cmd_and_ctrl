<script lang="ts">
  // FinalBoardModal — #4912: an eliminated player's final board, what they
  // had when they left the game, read only.
  //
  // Built on the zone browser (ZoneBrowserModal.svelte): the same full-
  // screen dialog, the same tab strip and the same card grid, with every
  // action taken out. A tab per zone the record holds (battlefield, hand,
  // graveyard, exile, command zone, never the library), each with its
  // count. On the battlefield an attachment follows its host and says what
  // it was on; a card that was tapped says so instead of turning in the
  // grid. Hover zoom works as in the zone browser: Card writes the shared
  // hovered card and Board's overlay draws it.
  //
  // The record is the seat's own `final_board`, on every frame, so this
  // opens the same after a reload or for a spectator who came late.

  import { onMount, tick } from "svelte";
  import type { PlayerView } from "../../protocol";
  import Card from "./Card.svelte";
  import Icon from "../Icon.svelte";
  import ModalLayer from "../ModalLayer.svelte";
  import { L } from "../../labels";
  import {
    FINAL_ZONES,
    finalBattlefieldOrder,
    finalCardAsView,
    finalCardNotes,
    finalZoneCards,
    firstFinalZone,
    type FinalZone,
  } from "../../finalBoard";

  interface Props {
    seat: PlayerView;
    onClose: () => void;
  }

  const { seat, onClose }: Props = $props();

  const name = $derived(seat.display_name || seat.name);
  const board = $derived(seat.final_board);

  let zone = $state<FinalZone>("battlefield");
  let chosen = false;
  $effect(() => {
    // Open on the battlefield, or the first zone with a card in it. Once
    // the viewer has picked a tab it stays theirs.
    if (!chosen) zone = firstFinalZone(board);
  });

  const cards = $derived.by(() => {
    const list = finalZoneCards(board, zone);
    return zone === "battlefield" ? finalBattlefieldOrder(list) : list;
  });

  function pick(kind: FinalZone): void {
    chosen = true;
    zone = kind;
  }

  // Arrow keys move along the tabs, as a tab strip's do.
  function tabKey(ev: KeyboardEvent, i: number): void {
    if (ev.key !== "ArrowRight" && ev.key !== "ArrowLeft") return;
    ev.preventDefault();
    const n = FINAL_ZONES.length;
    const next = FINAL_ZONES[(i + (ev.key === "ArrowRight" ? 1 : n - 1)) % n];
    pick(next.kind);
    void tick().then(() =>
      dialog?.querySelector<HTMLElement>(`[data-final-zone="${next.kind}"]`)?.focus(),
    );
  }

  let dialog = $state<HTMLElement | null>(null);
  let closeBtn = $state<HTMLButtonElement | null>(null);
  onMount(() => {
    const back = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    closeBtn?.focus();
    return () => back?.focus?.();
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

<div class="prompt-backdrop" role="presentation" onclick={backdropClick}>
  <div
    class="prompt-modal fb-modal"
    role="dialog"
    aria-modal="true"
    aria-label={L.finalBoardOf(name)}
    tabindex="-1"
    bind:this={dialog}
    data-final-board={seat.id}
  >
    <header class="fb-head">
      <h2>
        {name}<span class="sep">·</span><span class="zone">{L.finalBoard}</span>
      </h2>
      <button
        type="button"
        class="ghost close"
        onclick={onClose}
        aria-label="close"
        bind:this={closeBtn}
      >
        <Icon name="x" size={14} />
      </button>
    </header>
    <p class="fb-note">
      What {name} had when they left the game. Read only.
    </p>
    <div class="tabs" role="tablist" aria-label="zone">
      {#each FINAL_ZONES as t, i (t.kind)}
        {@const n = finalZoneCards(board, t.kind).length}
        <button
          type="button"
          class="tab"
          class:on={t.kind === zone}
          role="tab"
          aria-selected={t.kind === zone}
          tabindex={t.kind === zone ? 0 : -1}
          data-final-zone={t.kind}
          onclick={() => pick(t.kind)}
          onkeydown={(ev) => tabKey(ev, i)}
        >
          {t.label}<span class="count">{n}</span>
        </button>
      {/each}
    </div>
    <div class="panel" role="tabpanel" aria-label={FINAL_ZONES.find((z) => z.kind === zone)?.label}>
      {#if cards.length === 0}
        <p class="empty">Nothing here when they left.</p>
      {:else}
        <ul class="grid">
          {#each cards as card (card.instance_id)}
            {@const notes = finalCardNotes(card)}
            <li class="cell" data-final-card={card.instance_id}>
              <Card card={finalCardAsView(card, seat.id)} faceDown={card.face_down === true} />
              {#if notes.length > 0}
                <p class="note">{notes.join(" · ")}</p>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>
</div>

<style>
  .fb-modal {
    min-width: min(560px, calc(100vw - 32px));
    max-width: min(920px, calc(100vw - 32px));
    overflow: hidden;
    gap: 10px;
  }
  .fb-head {
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
    color: var(--accent-strong);
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
  }
  .close {
    width: 28px;
    height: 28px;
    padding: 0;
    border-radius: 50%;
    flex: 0 0 auto;
  }
  .fb-note {
    color: var(--fg-muted);
    font-size: 0.8rem;
    margin: 0;
    padding: 4px 8px;
    border-left: 2px solid var(--fg-dim);
  }
  .tabs {
    display: inline-flex;
    flex-wrap: wrap;
    align-self: flex-start;
    gap: 2px;
    padding: 3px;
    border-radius: 9px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
  }
  .tab {
    display: inline-flex;
    align-items: center;
    gap: 6px;
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
    transition:
      background 120ms var(--ease),
      color 120ms var(--ease);
  }
  .tab:hover {
    background: var(--surface-raised);
    color: var(--fg);
  }
  .tab.on {
    background: var(--surface-raised);
    border-color: var(--border-strong);
    color: var(--accent-strong);
  }
  .count {
    font-variant-numeric: tabular-nums;
    color: var(--fg-dim);
    font-weight: 500;
  }
  .tab.on .count {
    color: var(--fg-muted);
  }
  .empty {
    color: var(--fg-muted);
    font-size: 0.8rem;
    margin: 4px 0 0;
  }
  .panel {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
  .grid {
    list-style: none;
    padding: 4px 4px 8px;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 12px 10px;
    overflow: auto;
    --card-w: 110px;
    --card-h: 154px;
  }
  .cell {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
  }
  .note {
    margin: 0;
    max-width: 110px;
    font-size: 0.72rem;
    line-height: 1.25;
    color: var(--fg-muted);
    text-align: center;
    overflow-wrap: anywhere;
  }
</style>
