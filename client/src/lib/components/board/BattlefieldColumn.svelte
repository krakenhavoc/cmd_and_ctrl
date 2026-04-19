<script lang="ts">
  // BattlefieldColumn is the vertical sibling of BattlefieldRow,
  // used for the tall ENCHANTMENTS / ARTIFACTS column on the right
  // side of the player panel. Same data + click contract as the row;
  // only the layout direction differs. Kept as a separate component
  // (rather than a `direction` prop on a single component) because the
  // CSS for vertical wrapping vs horizontal wrapping diverges enough
  // that a single template would need branchy class logic that's
  // harder to read than two short files.

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

<div class="col" data-zone={label}>
  <span class="col-label" aria-hidden="true">{label}</span>
  <div class="col-cards" role="list" aria-label={label}>
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
  .col {
    position: relative;
    background: #111a2b;
    border: 1px solid #2e3a55;
    border-radius: 6px;
    padding: 18px 8px 8px;
    min-height: 0;
    min-width: 0;
    overflow: auto;
  }
  .col-label {
    position: absolute;
    top: 4px;
    left: 8px;
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: #6c7a99;
    pointer-events: none;
  }
  .col-cards {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 8px;
    align-content: flex-start;
    height: 100%;
  }
</style>
