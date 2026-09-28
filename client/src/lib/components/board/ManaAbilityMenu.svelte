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
  import {
    abilityBlocked as sharedAbilityBlocked,
    chargedManaCostLabel,
    chargedManaCostNote,
    type AbilityCost,
  } from "../../contextMenu.logic";
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
    // #1438: a left-click on a mana source taps it FOR mana now, so
    // turning it sideways WITHOUT making mana lives here, labelled so
    // nobody mistakes it for the mana row. Undefined hides it.
    onRawTap?: () => void;
    onClose?: () => void;
    // #1695: the paying player's current life — the card's
    // controller, same as the right-click context menu's rule. A
    // battlefield row only ever holds one seat's own permanents, so
    // PlayerPanel hands down that seat's life once per panel, the same
    // way it hands down `sorcerySpeedBlocked`. Undefined leaves an
    // unpayable life cost unblocked, like every other advisory check
    // here with nothing to judge against — the server's CR 119.4
    // refusal is still the real gate.
    payerLife?: number;
  }

  const {
    abilities,
    tapped,
    onActivate,
    activated = [],
    onActivateAbility,
    summoningSick = false,
    sorcerySpeedBlocked = "",
    onRawTap,
    onClose,
    payerLife,
  }: Props = $props();

  function activate(index: number): void {
    onActivate(index);
    onClose?.();
  }

  // An ability is unavailable when its tap cost can't be paid, its
  // life cost is more than the payer has, a sacrifice / return /
  // tap-other cost has nothing to pay it with, or any of the other
  // reasons the right-click context menu already judges through
  // `abilityBlocked` in contextMenu.logic.ts (#1695 — this used to be
  // a second, hand-maintained copy of that whole predicate, and the
  // copy never grew a life-cost check at all, which is what let an
  // unaffordable "Pay N life" ability stay clickable here after #1690
  // fixed the context menu's copy).
  //
  // The one check kept local is the sorcery-speed override: PlayerPanel
  // computes `sorcerySpeedBlocked` once per panel and threads it down
  // as a plain string — Card, Hand, PileBar, CommandZone and
  // ZoneBrowserModal all take the same prop — rather than the
  // LoyaltyContext + GameView the context menu has directly in scope.
  // Checked first, exactly where it sat before; everything else
  // (including the shared function's own `timing_closed` fallback,
  // for a row this prop doesn't cover) comes from the shared
  // predicate now.
  function abilityBlocked(a: AbilityCost): string {
    if (a.tap_cost && tapped) return "already tapped";
    if (a.tap_cost && summoningSick) return "summoning sickness";
    if (a.sorcery_speed && sorcerySpeedBlocked) return sorcerySpeedBlocked;
    return sharedAbilityBlocked(a, tapped, summoningSick, undefined, payerLife);
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
  {#if onRawTap}
    <div class="divider" role="separator"></div>
    <button
      type="button"
      class="menu-item"
      role="menuitem"
      title="Turn it sideways without adding mana"
      data-raw-tap
      onclick={(ev) => {
        ev.stopPropagation();
        onRawTap?.();
        onClose?.();
      }}
    >
      <span class="label">Tap (no mana)</span>
      <span class="cost" aria-hidden="true">↻</span>
    </button>
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
