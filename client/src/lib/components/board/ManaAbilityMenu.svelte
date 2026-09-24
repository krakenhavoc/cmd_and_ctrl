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
  import { counterCostBlocked, type CounterCostShape } from "../../counterCost";
  import {
    ACTIVATION_CONDITION_UNMET,
    NO_COMMANDER_IDENTITY,
    chargedManaCostLabel,
    chargedManaCostNote,
    returnShortfall,
    tapOthersShortfall,
    type ReturnOptionsShape,
  } from "../../contextMenu.logic";
  import { sacrificeShortfall } from "../../sacrificeCost";
  import { hasSatisfiableTargets } from "../../timing";
  import ModalLayer from "../ModalLayer.svelte";

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
    // S31: why the CR 307.1 sorcery-speed window is shut right now,
    // or "" when it is open. Computed once per panel by PlayerPanel
    // (which has the snapshot) rather than per card, and consulted
    // only for abilities that carry `sorcery_speed`.
    //
    // Until S31 this popover greyed on tap / sacrifice / target and
    // nothing else, so an "activate only as a sorcery" ability stayed
    // clickable all through combat and an opponent's turn and came
    // back rejected. The flag had been on the wire since S21.
    sorcerySpeedBlocked?: string;
    onClose?: () => void;
  }

  const {
    abilities,
    tapped,
    onActivate,
    activated = [],
    onActivateAbility,
    summoningSick = false,
    sorcerySpeedBlocked = "",
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
    sacrifice_options?: { players?: string[]; cards?: string[]; min?: number; max?: number };
    // #1213 / #1227: a return-to-hand cost. Ninjutsu is the row this
    // matters most for — it is payable only in the declare-blockers
    // window, with an unblocked attacker on the board, so a hand card
    // that never greyed would be clickable and refused nearly always.
    return_label?: string;
    return_options?: ReturnOptionsShape;
    // #759: a tap-another cost (station), greyed the same way.
    tap_others_label?: string;
    tap_others_options?: ReturnOptionsShape;
    // #1157: `min` carries the clause's count, and an "up to N" clause
    // (min 0) is satisfied by an empty candidate list.
    legal_targets?: { players?: string[]; cards?: string[]; min?: number };
    // Never set on a ManaAbilityView — mana abilities don't use the
    // stack and have no timing restriction (CR 605.1a) — so the arm
    // below is inert for the first list and live for the second.
    sorcery_speed?: boolean;
    // #743: the server says the ability's "Activate only if …"
    // condition is false. Both lists carry it — Temple of the False
    // God's mana row as much as Tectonic Edge's destroy.
    condition_unmet?: boolean;
    // #844, CR 903.4f: a "in your commander's color identity" mana
    // ability with no identity to narrow to adds nothing. Mana
    // abilities only; the arm below is inert for the activated list.
    adds_no_mana?: boolean;
  } & CounterCostShape;

  function abilityBlocked(a: CostShaped): string {
    if (a.tap_cost && tapped) return "already tapped";
    if (a.tap_cost && summoningSick) return "summoning sickness";
    if (a.sorcery_speed && sorcerySpeedBlocked) return sorcerySpeedBlocked;
    if (a.condition_unmet) return ACTIVATION_CONDITION_UNMET;
    // #844: Command Tower with no commander, or a colourless one.
    if (a.adds_no_mana) return NO_COMMANDER_IDENTITY;
    // #747: count-aware — "needs three Foods (you have 2)".
    const sacrifice = sacrificeShortfall(a.sacrifice_options, a.sacrifice_label ?? "a permanent");
    if (sacrifice) return sacrifice;
    // #1213 / #1227: the same question one verb over, off the one
    // shared predicate the right-click menu asks.
    const returned = returnShortfall(a.return_options, a.return_label);
    if (returned) return returned;
    const tapOthers = tapOthersShortfall(a.tap_others_options, a.tap_others_label);
    if (tapOthers) return tapOthers;
    // #625: a "remove N counters" cost with nothing that can pay it.
    const counters = counterCostBlocked(a);
    if (counters) return counters;
    // #1157: CR 601.2c through the shared predicate — a clause needs
    // `min` candidates, and "up to N" needs none. The same one-line
    // copy of this test lived here and in contextMenu.logic.ts, and
    // both stopped at "the list is empty".
    if (!hasSatisfiableTargets(a.legal_targets)) return "no legal target";
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

<ModalLayer />

<div class="mana-menu" role="menu" aria-label="abilities">
  {#each abilities as a (a.index)}
    {@const blocked = abilityBlocked(a)}
    {@const costNote = chargedManaCostNote(a)}
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
      {#if a.mana_cost}
        <!-- #1190: the ability's OWN mana component (the Signet
             cycle's "{1}", Loot's exhaust "{G}"). Shows what the
             engine actually charges (charged_mana_cost); the tooltip
             names the printed cost only when a discount made the two
             differ. -->
        <span class="cost" aria-label="mana cost" title={costNote || `mana cost ${a.mana_cost}`}>
          {chargedManaCostLabel(a)}
        </span>
      {/if}
      {#if a.sacrifice_cost || a.sacrifice_options}
        <span class="cost" aria-label="sacrifice cost">†</span>
      {/if}
      {#if a.life_cost}
        <!-- S22: a "Pay N life" cost component (Mana Confluence).
             Advisory — the server does the CR 119.4 check. The
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
      {@const costNote = chargedManaCostNote(a)}
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
        {#if a.mana_cost}
          <!-- #1190: same chip as the mana list above — the row's
               Label text already prints the ability's cost baked in
               by hand, so this is the ONE place a discount that made
               the printed text stale is visible. -->
          <span class="cost" aria-label="mana cost" title={costNote || `mana cost ${a.mana_cost}`}>
            {chargedManaCostLabel(a)}
          </span>
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
