<script lang="ts">
  // CompanionSlot — a seat's revealed companion, waiting outside the game
  // (CR 702.139a, ADR 0172 §8). It sits beside that seat's commanders:
  // yours next to the castable strip, everyone else's next to their
  // command strip. A revealed companion is public, so every seat shows
  // it face up, with the hover zoom every Card has and a "Companion"
  // caption.
  //
  // Your own carries the companion special action (CR 116.2g): pay {3}
  // and put it into your hand, once per game, in your main phase with an
  // empty stack. The row comes from the server (`special_actions` on the
  // seat's `companion`, owner-only), through the same builder every
  // special-action row uses (specialActionItems), so the button is greyed
  // exactly when the engine would refuse it, and sends what the popover
  // row sends. A right-click opens the card's popover with the same row.
  //
  // Renders nothing once the companion has been taken: the server drops
  // `companion` and the card is in the hand like any other.
  import type { CardView, PlayerView } from "../../protocol";
  import type { MenuAction } from "../../contextMenu.logic";
  import { specialActionItems } from "../../contextMenu.logic";
  import { NO_LEGAL_ACTIONS, type LegalActions } from "../../legalActions";
  import { L } from "../../labels";
  import Card from "./Card.svelte";

  interface Props {
    seat: Pick<PlayerView, "id" | "name" | "companion">;
    // The viewer's own seat: sends the special action. Undefined for
    // everyone else's companion, which is view only.
    onSpecialAction?: (action: MenuAction) => void;
    legal?: LegalActions;
    legalGate?: LegalActions;
    sorcerySpeedBlocked?: string;
    // A SeatSummary's row: smaller card, no button.
    compact?: boolean;
  }

  const {
    seat,
    onSpecialAction,
    legal = NO_LEGAL_ACTIONS,
    legalGate = NO_LEGAL_ACTIONS,
    sorcerySpeedBlocked = "",
    compact = false,
  }: Props = $props();

  const card: CardView | undefined = $derived(seat.companion);
  const take = $derived(
    card && onSpecialAction
      ? specialActionItems(card, card.owner || seat.id, legal, legalGate).find(
          (row) => row.action?.params?.kind === "companion",
        )
      : undefined,
  );
  const cost = $derived(
    card?.special_actions?.find((sa) => sa.kind === "companion")?.cost ?? "{3}",
  );

  function send(): void {
    if (take?.action && !take.disabled) onSpecialAction?.(take.action);
  }
</script>

{#if card}
  <div
    class="companion-slot"
    class:compact
    role="group"
    aria-label={L.companion(seat.name, card.name)}
  >
    <div class="slot" title={`${card.name} · companion, outside the game`}>
      <Card {card} {onSpecialAction} {legal} {legalGate} {sorcerySpeedBlocked} />
      <span class="caption">Companion</span>
    </div>
    {#if take && !compact}
      <button
        type="button"
        class="take"
        class:ready={take.ready}
        disabled={take.disabled}
        title={take.hint ?? take.label}
        aria-label={L.takeCompanion(cost, card.name)}
        onclick={send}
      >
        {cost} · to hand
      </button>
    {/if}
  </div>
{/if}

<style>
  /* Sized like CommandStrip's commander: 42% of a creature card beside
     a hand, 56px in a seat summary. The size is handed to the Card one
     element down through --strip-h, as CommandStrip does. */
  .companion-slot {
    --strip-h: calc(var(--card-h, 130px) * 0.42);
    flex: none;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
    align-self: flex-end;
    padding: 0 2px 2px;
  }
  .companion-slot.compact {
    --strip-h: 56px;
    align-self: center;
    padding: 0;
  }
  .slot {
    --card-h: var(--strip-h);
    --card-w: calc(var(--strip-h) * 5 / 7);
    position: relative;
    width: var(--card-w);
    height: var(--card-h);
    border-radius: 6px;
    box-shadow: 0 0 0 1px var(--accent-line);
  }
  .caption {
    position: absolute;
    left: 50%;
    bottom: 3px;
    transform: translateX(-50%);
    background: rgba(8, 7, 6, 0.9);
    color: var(--accent-strong);
    font-family: var(--font-mono);
    font-size: 8px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    padding: 1px 5px;
    border-radius: 999px;
    border: 1px solid var(--accent-line);
    pointer-events: none;
    white-space: nowrap;
  }
  .take {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 2px 8px;
    border-radius: 999px;
    border: 1px solid var(--accent-line);
    background: rgba(8, 7, 6, 0.85);
    color: var(--accent-strong);
    cursor: pointer;
    white-space: nowrap;
  }
  .take.ready {
    border-color: var(--accent-strong);
    box-shadow: 0 0 0 1px var(--accent-strong);
  }
  .take:disabled {
    opacity: 0.45;
    cursor: default;
  }
</style>
