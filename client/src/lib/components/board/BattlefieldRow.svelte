<script lang="ts">
  // BattlefieldRow renders a horizontal strip of battlefield cards
  // for one type bucket (CREATURES at the top, LANDS on the middle
  // left). The row owns layout (flex with wrap) and per-card visual
  // state derivation; the parent supplies the click handler so combat
  // / tap routing stays centralised.
  //
  // Cards are sorted by battle_x so a future row-reorder UX has a
  // stable handle. battle_y is intentionally ignored: typed rows make
  // a 2-D placement meaningless, and the server already defaults
  // both axes to 0 for cards that have never been positioned.

  import type { CardView } from "../../protocol";
  import Card from "./Card.svelte";
  import { etbPulse } from "../../animations";

  interface Props {
    label: string;
    cards: CardView[];
    viewerID: string | null;
    selectedCombatCardID?: string | null;
    onCardClick?: (card: CardView, ev: MouseEvent) => void;
  }

  const { label, cards, selectedCombatCardID = null, onCardClick }: Props = $props();

  const sorted = $derived([...cards].sort((a, b) => (a.battle_x ?? 0) - (b.battle_x ?? 0)));
</script>

<div class="row" data-zone={label}>
  <span class="row-label" aria-hidden="true">{label}</span>
  <div class="row-cards" role="list" aria-label={label}>
    {#each sorted as c (c.instance_id)}
      <div role="listitem" use:etbPulse>
        <Card
          card={c}
          selected={selectedCombatCardID === c.instance_id}
          attacking={!!c.attacking_target}
          blocking={!!c.blocking_target}
          onClick={onCardClick}
        />
      </div>
    {/each}
  </div>
</div>

<style>
  .row {
    position: relative;
    background: #111a2b;
    border: 1px solid #2e3a55;
    border-radius: 6px;
    padding: 18px 8px 8px;
    min-height: 0;
    min-width: 0;
    overflow: auto;
  }
  .row-label {
    position: absolute;
    top: 4px;
    left: 8px;
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: #6c7a99;
    pointer-events: none;
  }
  .row-cards {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 8px;
    align-content: flex-start;
    height: 100%;
  }
</style>
