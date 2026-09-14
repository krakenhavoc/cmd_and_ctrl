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
    // S31: why the CR 307.1 sorcery-speed window is shut, or "" when
    // it is open. Pass-through to Card → ManaAbilityMenu, which greys
    // `sorcery_speed` abilities with it.
    sorcerySpeedBlocked?: string;
    // compact — one card size down (PlayerPanel's --card-w-sm). Used
    // for the middle band: non-creature permanents and lands.
    compact?: boolean;
    // strip — overlapped horizontal strip instead of a wrapping grid.
    // Lands: seven of them fit in half a panel, and tapped ones sort
    // to the right so the untapped count reads at a glance.
    strip?: boolean;
    // S24: the Equipment and Auras attached to each host, keyed by
    // the host's instance ID. Derived by PlayerPanel from the
    // one-directional `attached_to` on the wire. They are drawn
    // inside the host's own listitem, offset behind it — the same
    // negative-margin overlap the land strip uses — so a sword reads
    // as being ON the creature rather than as a separate permanent.
    attachmentsByHost?: Record<string, CardView[]>;
    // S24: the name of the player each Curse-style permanent enchants,
    // keyed by the permanent's instance ID. A card attached to a
    // PLAYER has no host to be drawn behind, so the relation would be
    // invisible without this. Derived by PlayerPanel, which is the
    // component that can see the seat list.
    curseTargets?: Record<string, string>;
  }

  const {
    label,
    cards,
    selectedCombatCardID = null,
    onCardClick,
    onActivateManaAbility,
    onActivateAbility,
    sorcerySpeedBlocked = "",
    compact = false,
    strip = false,
    attachmentsByHost = {},
    curseTargets = {},
  }: Props = $props();

  const sorted = $derived.by(() => {
    const byX = [...cards].sort((a, b) => (a.battle_x ?? 0) - (b.battle_x ?? 0));
    if (!strip) return byX;
    // Stable partition: untapped first, tapped after, order kept within each.
    return [...byX.filter((c) => !c.tapped), ...byX.filter((c) => c.tapped)];
  });
</script>

<div class="row" class:compact class:strip data-zone={label}>
  <span class="row-label" aria-hidden="true">
    {label}
    {#if cards.length > 0}<span class="row-count">{cards.length}</span>{/if}
  </span>
  <div class="row-cards" role="list" aria-label={label}>
    {#each sorted as c (c.instance_id)}
      <div role="listitem" class:tapped={!!c.tapped} use:etbPulse>
        <div
          class="host-stack"
          class:has-attachments={(attachmentsByHost[c.instance_id] ?? []).length > 0}
        >
          {#each attachmentsByHost[c.instance_id] ?? [] as a (a.instance_id)}
            <div class="attachment">
              <Card
                card={a}
                enchantedPlayer={curseTargets[a.instance_id]}
                onClick={onCardClick}
                onActivateManaAbility={onActivateManaAbility
                  ? (idx) => onActivateManaAbility(a, idx)
                  : undefined}
                onActivateAbility={onActivateAbility
                  ? (idx) => onActivateAbility(a, idx)
                  : undefined}
                {sorcerySpeedBlocked}
              />
            </div>
          {/each}
          <Card
            card={c}
            enchantedPlayer={curseTargets[c.instance_id]}
            selected={selectedCombatCardID === c.instance_id}
            attacking={!!c.attacking_target}
            blocking={!!c.blocking_target}
            onClick={onCardClick}
            onActivateManaAbility={onActivateManaAbility
              ? (idx) => onActivateManaAbility(c, idx)
              : undefined}
            onActivateAbility={onActivateAbility ? (idx) => onActivateAbility(c, idx) : undefined}
            {sorcerySpeedBlocked}
          />
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .row {
    position: relative;
    padding: 18px 6px 6px;
    min-height: 0;
    min-width: 0;
    overflow: auto;
  }
  .row.compact {
    --card-w: var(--card-w-sm, 88px);
    --card-h: var(--card-h-sm, 123px);
  }
  .row-label {
    position: absolute;
    top: 4px;
    left: 6px;
    display: inline-flex;
    align-items: baseline;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    pointer-events: none;
    font-weight: 600;
    white-space: nowrap;
  }
  .row-count {
    letter-spacing: 0;
    font-weight: 500;
  }
  /* A host and everything attached to it. The attachments are laid
     out first and overlapped leftwards behind the host, which is
     drawn last so it sits on top; the wrapper claims only the host's
     width, so a creature carrying two swords does not reflow the row
     out from under its neighbours. */
  .host-stack {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
  }
  .host-stack.has-attachments {
    margin-left: calc(var(--card-w, 88px) * 0.34);
  }
  .host-stack .attachment {
    flex: 0 0 auto;
    margin-right: calc(var(--card-w, 88px) * -0.72);
    transform: translateY(8px);
    filter: brightness(0.88);
  }
  .host-stack .attachment:hover {
    z-index: 7;
    filter: none;
  }
  .row-cards {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 8px;
    align-content: flex-start;
    align-items: flex-start;
    height: 100%;
  }
  /* Land strip: no wrap, each card overlaps the previous so the name
     band stays readable; tapped cards (rotated by Card.svelte) are
     sorted to the end and given room for their rotated width. */
  .row.strip .row-cards {
    flex-wrap: nowrap;
    gap: 0;
    padding-right: calc(var(--card-h, 123px) * 0.2);
  }
  .row.strip .row-cards > [role="listitem"] {
    flex: 0 0 auto;
    margin-left: calc(var(--card-w, 88px) * -0.6);
  }
  .row.strip .row-cards > [role="listitem"]:first-child {
    margin-left: 0;
  }
  .row.strip .row-cards > [role="listitem"].tapped {
    margin-left: calc(var(--card-w, 88px) * -0.35);
  }
  .row.strip .row-cards > [role="listitem"]:not(.tapped) + [role="listitem"].tapped {
    margin-left: calc(var(--card-h, 123px) * 0.2);
  }
  .row.strip .row-cards > [role="listitem"]:hover {
    z-index: 6;
  }
</style>
