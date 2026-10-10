<script lang="ts">
  // DamageStepper — combat damage assigned on the card itself (#2956,
  // ADR 0147). While a damage_assignment sheet is open
  // (lib/damageAssignment.ts boardDamageAssign), each blocker carries
  // the damage assigned to it with − and +, marked once it is lethal,
  // and the attacker carries the running total: what is left to assign
  // and, with trample, what goes over to the player (with its own − and
  // +). The sheet in the dock reads the same store, so either surface
  // can be used and "Deal damage" sends what both show.
  //
  // The buttons stop the click, so a press never reaches the card's own
  // click (cast, tap, target).
  import {
    assignedTotal,
    boardDamageAssign,
    stepOnBoard,
    stepTrampleOnBoard,
  } from "../../damageAssignment";
  import { L } from "../../labels";

  interface Props {
    cardID: string;
    // The card's name, for the buttons' accessible names.
    name: string;
  }

  const { cardID, name }: Props = $props();

  const s = $derived($boardDamageAssign);
  const isBlocker = $derived(!!s && cardID in s.shares.amounts);
  const isAttacker = $derived(!!s && s.attackerID === cardID);
  const amount = $derived(s && isBlocker ? s.shares.amounts[cardID] : 0);
  const lethal = $derived(s && isBlocker ? (s.lethal[cardID] ?? null) : null);
  const isLethal = $derived(lethal !== null && amount >= lethal);
  const left = $derived(s ? s.power - assignedTotal(s.shares) : 0);

  function press(e: MouseEvent, delta: 1 | -1, trample = false): void {
    e.stopPropagation();
    e.preventDefault();
    if (trample) stepTrampleOnBoard(delta);
    else stepOnBoard(cardID, delta);
  }
</script>

{#if s && isBlocker}
  <div
    class="dmg blocker"
    class:lethal={isLethal}
    role="group"
    aria-label={`damage to ${name}`}
    data-testid="damage-stepper"
  >
    <button
      type="button"
      aria-label={L.removeDamage(name)}
      disabled={amount <= 0}
      onclick={(e) => press(e, -1)}>−</button
    >
    <span
      class="amount"
      aria-live="polite"
      title={lethal === null ? undefined : `lethal damage: ${lethal}`}
    >
      <strong>{amount}</strong>
      {#if lethal !== null}
        <small>{isLethal ? "✓" : `/${lethal}`}</small>
      {/if}
    </span>
    <button
      type="button"
      aria-label={L.addDamage(name)}
      disabled={left <= 0 && s.shares.trample <= 0}
      onclick={(e) => press(e, 1)}>+</button
    >
  </div>
{:else if s && isAttacker}
  <div class="dmg attacker" role="group" aria-label={`damage from ${name}`}>
    <span class="total" title="damage left to assign">
      {left === 0 ? "all assigned" : `${left} of ${s.power} left`}
    </span>
    {#if s.allowTrample}
      <span class="trample">
        <button
          type="button"
          aria-label={L.removeDamage("the defending player")}
          disabled={s.shares.trample <= 0}
          onclick={(e) => press(e, -1, true)}>−</button
        >
        <span title="trample damage to the defending player">{s.shares.trample} over</span>
        <button
          type="button"
          aria-label={L.addDamage("the defending player")}
          disabled={left <= 0}
          onclick={(e) => press(e, 1, true)}>+</button
        >
      </span>
    {/if}
  </div>
{/if}

<style>
  .dmg {
    position: absolute;
    left: 50%;
    top: 50%;
    /* Upright on a tapped card too: undo the card's own tap rotation. */
    transform: translate(-50%, -50%) rotate(calc(-1 * var(--tap-rot, 0deg)));
    z-index: 4;
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 2px;
    border-radius: 999px;
    background: var(--surface-sunken);
    border: 1px solid var(--attack);
    box-shadow: 0 2px 8px color-mix(in srgb, var(--shadow-ink) 50%, transparent);
    color: var(--fg);
    font-family: var(--font-ui);
    white-space: nowrap;
  }
  .dmg.lethal {
    border-color: var(--target);
  }
  .dmg.attacker {
    flex-direction: column;
    gap: 2px;
    border-radius: 8px;
    padding: 3px 5px;
    font-size: 10px;
    font-weight: 600;
  }
  .amount {
    display: inline-flex;
    align-items: baseline;
    gap: 1px;
    min-width: 1.8em;
    justify-content: center;
  }
  .amount strong {
    font-size: 14px;
    font-variant-numeric: tabular-nums;
  }
  .amount small {
    font-size: 9px;
    color: var(--fg-muted);
  }
  .dmg.lethal .amount small {
    color: var(--target);
  }
  .trample {
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
  button {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    color: var(--fg);
    font-size: 14px;
    line-height: 1;
    padding: 0;
    cursor: pointer;
  }
  button:hover:not(:disabled) {
    background: var(--surface-hover);
  }
  button:disabled {
    opacity: 0.4;
    cursor: default;
  }
  button:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
</style>
