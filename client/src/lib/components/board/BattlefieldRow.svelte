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
    // onActivateManaAbility — fires `activate_mana_ability` for a
    // battlefield permanent the viewer controls. Routed down to
    // Card.svelte so the right-click / context menu can hit it.
    // Undefined suppresses the menu entirely (opponent panels).
    onActivateManaAbility?: (card: CardView, abilityIndex: number) => void;
    // S21 sub-PR 2: CR 602 activated abilities, same menu.
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
  }

  const {
    label,
    cards,
    selectedCombatCardID = null,
    onCardClick,
    onActivateManaAbility,
    onActivateAbility,
  }: Props = $props();

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
          onActivateManaAbility={onActivateManaAbility
            ? (idx) => onActivateManaAbility(c, idx)
            : undefined}
          onActivateAbility={onActivateAbility ? (idx) => onActivateAbility(c, idx) : undefined}
        />
      </div>
    {/each}
  </div>
</div>

<style>
  .row {
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
  .row-label {
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
  .row-cards {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 8px;
    align-content: flex-start;
    height: 100%;
  }
</style>
