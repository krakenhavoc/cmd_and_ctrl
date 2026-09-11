<script lang="ts">
  // ManaAbilityMenu — small popover anchored to a battlefield permanent
  // that lists the card's activated mana abilities (S15). One button
  // per entry in card.mana_abilities; click fires the supplied
  // onActivate callback with the ability's index. Tap-cost greys the
  // button when the card is already tapped. Dismisses on Escape or
  // outside click (handled by the parent that mounts it).
  //
  // The menu is position-neutral — the parent wraps it in an
  // absolutely-positioned container anchored near the card. Keeping
  // layout concerns out of this component lets the battlefield
  // figure out overflow / flip-to-top-if-near-edge itself.

  import type { ActivatedAbilityView, ManaAbilityView } from "../../protocol";

  interface Props {
    abilities: ManaAbilityView[];
    tapped: boolean;
    onActivate: (abilityIndex: number) => void;
    // S21 sub-PR 2: CR 602 activated abilities, listed below the
    // mana abilities in the same popover. Costs that need a further
    // choice (sacrifice, target) are collected by the parent after
    // the click.
    activated?: ActivatedAbilityView[];
    onActivateAbility?: (abilityIndex: number) => void;
    summoningSick?: boolean;
    onClose?: () => void;
  }

  const {
    abilities,
    tapped,
    onActivate,
    activated = [],
    onActivateAbility,
    summoningSick = false,
    onClose,
  }: Props = $props();

  function activate(index: number): void {
    onActivate(index);
    onClose?.();
  }

  // An ability is unavailable when its tap cost can't be paid, or
  // when a sacrifice cost has nothing to pay it with. The server
  // re-checks everything; this is just the affordance.
  //
  // Mana and activated abilities share the cost-shaped fields, so
  // one predicate covers both — the activated-only clauses (targets)
  // simply don't appear on a ManaAbilityView.
  type CostShaped = {
    tap_cost?: boolean;
    sacrifice_label?: string;
    sacrifice_options?: { players?: string[]; cards?: string[] };
    legal_targets?: { players?: string[]; cards?: string[] };
  };

  function abilityBlocked(a: CostShaped): string {
    if (a.tap_cost && tapped) return "already tapped";
    if (a.tap_cost && summoningSick) return "summoning sickness";
    if (a.sacrifice_options) {
      const n = a.sacrifice_options.cards?.length ?? 0;
      if (n === 0) return `nothing to sacrifice (${a.sacrifice_label ?? "a permanent"})`;
    }
    if (a.legal_targets) {
      const n = (a.legal_targets.players?.length ?? 0) + (a.legal_targets.cards?.length ?? 0);
      if (n === 0) return "no legal target";
    }
    return "";
  }

  function activateAbility(index: number): void {
    onActivateAbility?.(index);
    onClose?.();
  }

  function onKey(ev: KeyboardEvent): void {
    if (ev.key === "Escape") {
      ev.preventDefault();
      onClose?.();
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="mana-menu" role="menu" aria-label="abilities">
  {#each abilities as a (a.index)}
    {@const blocked = abilityBlocked(a)}
    <button
      type="button"
      class="menu-item"
      role="menuitem"
      disabled={!!blocked}
      title={blocked || (a.produced ? `produces ${a.produced}` : a.label)}
      onclick={(ev) => {
        ev.stopPropagation();
        if (!blocked) activate(a.index);
      }}
    >
      <span class="label">{a.label || a.produced || "activate"}</span>
      {#if a.tap_cost}
        <span class="cost" aria-label="tap cost">↻</span>
      {/if}
      {#if a.sacrifice_cost || a.sacrifice_options}
        <span class="cost" aria-label="sacrifice cost">†</span>
      {/if}
      {#if a.life_cost}
        <!-- S22: a "Pay N life" cost component (Mana Confluence).
             Advisory — the server does the CR 118.8 check. The
             painlands' "deals 1 damage to you" is a RIDER, not a
             cost, so it shows up in the label instead of here. -->
        <span class="cost" aria-label={`pay ${a.life_cost} life`}>♥{a.life_cost}</span>
      {/if}
    </button>
  {/each}
  {#if activated.length > 0}
    {#if abilities.length > 0}
      <div class="divider" role="separator"></div>
    {/if}
    {#each activated as a (a.index)}
      {@const blocked = abilityBlocked(a)}
      <button
        type="button"
        class="menu-item"
        role="menuitem"
        disabled={!!blocked}
        title={blocked || a.label}
        onclick={(ev) => {
          ev.stopPropagation();
          if (!blocked) activateAbility(a.index);
        }}
      >
        <span class="label">{a.label || "activate"}</span>
        {#if a.tap_cost}
          <span class="cost" aria-label="tap cost">↻</span>
        {/if}
      </button>
    {/each}
  {/if}
</div>

<style>
  .divider {
    height: 1px;
    margin: 4px 2px;
    background: rgba(200, 168, 106, 0.35);
  }

  .mana-menu {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 4px;
    background: rgba(12, 16, 30, 0.96);
    color: var(--gold);
    border: 1px solid rgba(200, 168, 106, 0.55);
    border-radius: 6px;
    box-shadow: 0 10px 24px rgba(0, 0, 0, 0.6);
    min-width: 180px;
    font-size: 11px;
    z-index: 60;
  }
  .menu-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    background: transparent;
    color: inherit;
    border: 1px solid transparent;
    border-radius: 4px;
    text-align: left;
    cursor: pointer;
    font-size: 11px;
    line-height: 1.2;
  }
  .menu-item:hover:not(:disabled),
  .menu-item:focus-visible:not(:disabled) {
    background: rgba(200, 168, 106, 0.15);
    border-color: rgba(200, 168, 106, 0.4);
    outline: none;
  }
  .menu-item:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .label {
    flex: 1;
  }
  .cost {
    font-weight: 700;
    opacity: 0.75;
  }
</style>
