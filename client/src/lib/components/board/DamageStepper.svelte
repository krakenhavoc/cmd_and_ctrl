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
    stepGroupOnBoard,
    stepOnBoard,
    stepTrampleOnBoard,
  } from "../../damageAssignment";
  import { L } from "../../labels";

  interface Props {
    cardID: string;
    // The card's name, for the buttons' accessible names.
    name: string;
    // A token group's card (#2958): every member it stands for. The
    // card carries one stepper for the group. A press moves one point
    // on one member, lethal-first in the group's order, through the
    // same store the sheet reads (stepGroup).
    memberIDs?: readonly string[];
  }

  const { cardID, name, memberIDs }: Props = $props();

  const s = $derived($boardDamageAssign);
  const ids = $derived(memberIDs ?? [cardID]);
  const grouped = $derived(!!memberIDs && memberIDs.length > 1);
  // The members that are in this assignment.
  const blockers = $derived(s ? ids.filter((id) => id in s.shares.amounts) : []);
  const isBlocker = $derived(blockers.length > 0);
  const isAttacker = $derived(!!s && ids.includes(s.attackerID));
  const amount = $derived(s ? blockers.reduce((n, id) => n + s.shares.amounts[id], 0) : 0);
  // The group's lethal is the sum of its members'; null when the server
  // sent no figure for one of them.
  const lethal = $derived.by(() => {
    if (!s || !isBlocker) return null;
    let n = 0;
    for (const id of blockers) {
      const l = s.lethal[id];
      if (l === undefined) return null;
      n += l;
    }
    return n;
  });
  // Every member at its own lethal, not just the total: 3 + 0 on two
  // 2-toughness tokens is as many points as 2 + 2 and kills one.
  const isLethal = $derived(
    !!s &&
      isBlocker &&
      blockers.every((id) => s.lethal[id] !== undefined && s.shares.amounts[id] >= s.lethal[id]),
  );
  const left = $derived(s ? s.power - assignedTotal(s.shares) : 0);
  const label = $derived(grouped ? `${name} (group of ${blockers.length})` : name);

  function press(e: MouseEvent, delta: 1 | -1, trample = false): void {
    e.stopPropagation();
    e.preventDefault();
    if (trample) stepTrampleOnBoard(delta);
    else if (grouped) stepGroupOnBoard(ids, delta);
    else stepOnBoard(cardID, delta);
  }
</script>

{#if s && isBlocker}
  <div
    class="dmg blocker"
    class:lethal={isLethal}
    role="group"
    aria-label={`damage to ${label}`}
    data-testid="damage-stepper"
  >
    <button
      type="button"
      aria-label={L.removeDamage(label)}
      disabled={amount <= 0}
      onclick={(e) => press(e, -1)}>−</button
    >
    <span
      class="amount"
      aria-live="polite"
      title={lethal === null
        ? undefined
        : grouped
          ? `lethal damage: ${lethal} across ${blockers.length} tokens`
          : `lethal damage: ${lethal}`}
    >
      <strong>{amount}</strong>
      {#if lethal !== null}
        <small>{isLethal ? "✓" : `/${lethal}`}</small>
      {/if}
      {#if grouped}
        <small class="members">×{blockers.length}</small>
      {/if}
    </span>
    <button
      type="button"
      aria-label={L.addDamage(label)}
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
