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

  import type { ManaAbilityView } from "../../protocol";

  interface Props {
    abilities: ManaAbilityView[];
    tapped: boolean;
    onActivate: (abilityIndex: number) => void;
    onClose?: () => void;
  }

  const { abilities, tapped, onActivate, onClose }: Props = $props();

  function activate(index: number): void {
    onActivate(index);
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

<div class="mana-menu" role="menu" aria-label="mana abilities">
  {#each abilities as a (a.index)}
    {@const disabled = !!a.tap_cost && tapped}
    <button
      type="button"
      class="menu-item"
      role="menuitem"
      {disabled}
      title={disabled ? "already tapped" : a.produced ? `produces ${a.produced}` : a.label}
      onclick={(ev) => {
        ev.stopPropagation();
        if (!disabled) activate(a.index);
      }}
    >
      <span class="label">{a.label || a.produced || "activate"}</span>
      {#if a.tap_cost}
        <span class="cost" aria-label="tap cost">↻</span>
      {/if}
    </button>
  {/each}
</div>

<style>
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
