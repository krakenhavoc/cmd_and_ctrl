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
    onActivateManaAbility?: (card: CardView, abilityIndex: number) => void;
  }

  const {
    label,
    cards,
    selectedCombatCardID = null,
    onCardClick,
    onActivateManaAbility,
  }: Props = $props();

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
          onActivateManaAbility={onActivateManaAbility
            ? (idx) => onActivateManaAbility(c, idx)
            : undefined}
        />
      </div>
    {/each}
  </div>
</div>

<style>
  .col {
    position: relative;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.02) 0%, rgba(0, 0, 0, 0.15) 100%),
      var(--surface-sunken);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 18px 10px 10px;
    min-height: 0;
    min-width: 0;
    overflow: auto;
    box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.35);
  }
  .col-label {
    position: absolute;
    top: 5px;
    left: 10px;
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    pointer-events: none;
    font-weight: 700;
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
