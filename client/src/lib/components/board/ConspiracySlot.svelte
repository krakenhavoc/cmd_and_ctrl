<script lang="ts">
  // ConspiracySlot — a seat's conspiracy cards in the command zone (CR 315,
  // ADR 0227). They sit beside that seat's commanders and companion: yours
  // next to the castable strip, everyone else's next to their command strip.
  //
  // A face-up conspiracy is public and shows face up, with the hover zoom
  // every Card has. A face-down one (hidden agenda, CR 702.106a) shows its
  // back to everyone but its owner (CR 315.7); the server sends the owner
  // the card and its `chosen_name`, which shows as a caption under it.
  //
  // Your own face-down conspiracy carries the reveal button: turn it face up
  // and reveal the name (CR 702.106c), free, any time you have priority. The
  // row comes from the server (`special_actions`, owner only) through the
  // same builder every special-action row uses (specialActionItems), so the
  // button is greyed exactly when the engine would refuse it.
  import type { CardView, PlayerView } from "../../protocol";
  import type { MenuAction } from "../../contextMenu.logic";
  import { specialActionItems } from "../../contextMenu.logic";
  import { NO_LEGAL_ACTIONS, type LegalActions } from "../../legalActions";
  import { L } from "../../labels";
  import Card from "./Card.svelte";

  interface Props {
    seat: Pick<PlayerView, "id" | "name" | "conspiracies">;
    // The viewer's own seat: sends the special action. Undefined for
    // everyone else's conspiracies, which are view only.
    onSpecialAction?: (action: MenuAction) => void;
    legal?: LegalActions;
    legalGate?: LegalActions;
    // A SeatSummary's row: smaller cards, no button.
    compact?: boolean;
  }

  const {
    seat,
    onSpecialAction,
    legal = NO_LEGAL_ACTIONS,
    legalGate = NO_LEGAL_ACTIONS,
    compact = false,
  }: Props = $props();

  const cards: CardView[] = $derived(seat.conspiracies ?? []);

  function revealRow(card: CardView) {
    if (!onSpecialAction || !card.face_down) return undefined;
    return specialActionItems(card, card.owner || seat.id, legal, legalGate).find(
      (row) => row.action?.params?.kind === "reveal_agenda",
    );
  }

  function caption(card: CardView): string {
    if (card.chosen_name) return card.chosen_name;
    return card.face_down ? "Hidden agenda" : "Conspiracy";
  }
</script>

{#if cards.length > 0}
  <div
    class="conspiracy-slot"
    class:compact
    role="group"
    aria-label={L.conspiracies(seat.name, cards.length)}
  >
    {#each cards as card (card.instance_id)}
      {@const reveal = revealRow(card)}
      <div class="entry">
        <div class="slot" title={card.name ? `${card.name} · conspiracy` : "Face-down conspiracy"}>
          <Card {card} />
          <span class="caption">{caption(card)}</span>
        </div>
        {#if reveal && !compact}
          <button
            type="button"
            class="reveal"
            class:ready={reveal.ready}
            disabled={reveal.disabled}
            title={reveal.hint ?? reveal.label}
            aria-label={L.revealAgenda(card.name)}
            onclick={() => {
              if (reveal.action && !reveal.disabled) onSpecialAction?.(reveal.action);
            }}
          >
            Reveal
          </button>
        {/if}
      </div>
    {/each}
  </div>
{/if}

<style>
  /* Sized like CompanionSlot: 42% of a creature card beside a hand, 56px
     in a seat summary. */
  .conspiracy-slot {
    --strip-h: calc(var(--card-h, 130px) * 0.42);
    flex: none;
    display: flex;
    gap: 4px;
    align-self: flex-end;
    padding: 0 2px 2px;
  }
  .conspiracy-slot.compact {
    --strip-h: 56px;
    align-self: center;
    padding: 0;
  }
  .entry {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
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
    max-width: calc(var(--card-w) + 24px);
    overflow: hidden;
    text-overflow: ellipsis;
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
  .reveal {
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
  .reveal.ready {
    border-color: var(--accent-strong);
    box-shadow: 0 0 0 1px var(--accent-strong);
  }
  .reveal:disabled {
    opacity: 0.45;
    cursor: default;
  }
  .reveal:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
</style>
