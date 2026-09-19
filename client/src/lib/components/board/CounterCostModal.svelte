<script lang="ts">
  // CounterCostModal — #625, #789, #943: choose what pays a counter
  // activation cost.
  //
  // Two layouts, because the printed cards ask two different
  // questions:
  //
  //   pick one   Heart of Kiran's "remove a loyalty counter from a
  //              planeswalker you control" with more than one walker
  //              out; Fain, the Broker's "a counter from a creature
  //              you control", where the kind is part of the answer.
  //              One row per (permanent, kind), radio-style.
  //   how many   Iron Spider's "two +1/+1 counters from among
  //              artifacts you control" and Mage-Ring Network's "any
  //              number of storage counters" — the player says how
  //              many come off each permanent, with a running total
  //              against the printed number or the floor.
  //
  // #943's any-kind among cost (Tekuthal's "three counters from among
  // other artifacts, creatures, and planeswalkers you control") needs
  // no third layout and no kind dropdown: a row has always been a
  // (permanent, KIND) pair, so a creature with a +1/+1 and a shield
  // counter simply offers two rows and the stepper on each IS the
  // kind choice. What changes is only the label — with no printed
  // kind, each row says which kind it spends.
  //
  // One modal for both, and for both ability kinds: since #789 a MANA
  // ability carries the identical counter fields, so Vivid Creek's
  // charge counter and Heart of Kiran's loyalty counter open the same
  // prompt. That is what the structural `ability` prop buys.
  //
  // SacrificeCostModal's shape — a cost is not a target, so this is a
  // plain list of your own permanents rather than the board-click
  // targeting flow, and it opens before any target prompt. The rows
  // come straight off the server's counter_cost_options, most counters
  // first, and Board skips this modal entirely when there is only one
  // way to pay.

  import { onDestroy } from "svelte";
  import type { CardView } from "../../protocol";
  import {
    counterChoiceKey,
    counterChoices,
    counterPaymentReady,
    isMultiCounterCost,
    type CounterChoice,
    type CounterCostShape,
  } from "../../counterCost";
  import ModalLayer from "../ModalLayer.svelte";

  // The cost-shaped subset both ActivatedAbilityView and
  // ManaAbilityView satisfy, plus the index that identifies the row.
  type CounterCostAbility = CounterCostShape & { index: number };

  interface Props {
    // The ability's source, for the heading; null closes the modal.
    card: CardView | null;
    ability: CounterCostAbility | null;
    // The battlefield, to name the options.
    board: CardView[];
    onConfirm: (choices: CounterChoice[]) => void;
    onCancel: () => void;
  }

  const { card, ability, board, onConfirm, onCancel }: Props = $props();

  // Single-pick: the chosen row's key. Many-pick: how many counters
  // come off each row, keyed the same way.
  let chosen = $state<string | null>(null);
  let counts = $state<Record<string, number>>({});

  // Reset when a different activation opens the prompt.
  let lastKey: string | null = null;
  $effect(() => {
    const key = card && ability ? `${card.instance_id}:${ability.index}` : null;
    if (key !== lastKey) {
      lastKey = key;
      chosen = null;
      counts = {};
    }
  });

  const choices = $derived(ability ? counterChoices(ability) : []);
  const multi = $derived(!!ability && isMultiCounterCost(ability));
  const anyKind = $derived(!ability?.counter_cost_kind);
  const n = $derived(ability?.counter_cost_n ?? 1);
  const max = $derived(ability?.counter_cost_max ?? 0);

  // The picks as the payload builder wants them: one entry per
  // permanent actually paying, carrying its share.
  const picks = $derived.by<CounterChoice[]>(() => {
    if (!multi) {
      const c = choices.find((x) => counterChoiceKey(x) === chosen);
      return c ? [{ ...c, n }] : [];
    }
    return choices
      .map((c) => ({ ...c, n: counts[counterChoiceKey(c)] ?? 0 }))
      .filter((c) => (c.n ?? 0) > 0);
  });

  const paid = $derived(picks.reduce((sum, c) => sum + (c.n ?? 0), 0));
  const ready = $derived(!!ability && counterPaymentReady(ability, picks));

  const hint = $derived.by(() => {
    if (!ability) return "";
    const kind = ability.counter_cost_kind ? `${ability.counter_cost_kind} ` : "";
    const from = ability.counter_cost_self
      ? "this permanent"
      : (ability.counter_cost_label ?? "a permanent you control");
    if (ability.counter_cost_variable) {
      const floor = n > 0 ? ` At least ${n}.` : "";
      return `Remove any number of ${kind}counters from ${from} to pay for this ability.${floor}`;
    }
    const what = n === 1 ? `a ${kind}counter` : `${n} ${kind}counters`;
    if (ability.counter_cost_among) {
      const mix = ability.counter_cost_kind ? "" : " Any kinds.";
      return `Remove ${what} from among ${from} to pay for this ability. Split them however you like.${mix}`;
    }
    return `Remove ${what} from ${from} to pay for this ability.`;
  });

  // The many-pick footer's target: an exact number for an among cost,
  // a floor for a variable one.
  const goal = $derived(ability?.counter_cost_variable ? `${paid}` : `${paid} / ${n}`);

  function nameOf(id: string): string {
    return board.find((c) => c.instance_id === id)?.name ?? "a permanent";
  }

  // rowOf names one row for a screen reader. With no printed kind a
  // permanent can hold two rows, so the kind is part of the name or
  // the two buttons read identically (#943).
  function rowOf(c: CounterChoice): string {
    return anyKind ? `${nameOf(c.cardID)} (${c.kind})` : nameOf(c.cardID);
  }

  function ceilingFor(c: CounterChoice): number {
    if (!ability?.counter_cost_variable) return c.count;
    return max > 0 ? Math.min(c.count, max) : c.count;
  }

  function bump(c: CounterChoice, delta: number): void {
    const key = counterChoiceKey(c);
    const next = (counts[key] ?? 0) + delta;
    if (next < 0 || next > ceilingFor(c)) return;
    counts = { ...counts, [key]: next };
  }

  function confirm(): void {
    if (!ready) return;
    onConfirm(picks);
  }

  function handleKey(e: KeyboardEvent): void {
    if (!card) return;
    if (e.key === "Enter") {
      e.preventDefault();
      confirm();
    } else if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    }
  }
  $effect(() => {
    if (!card) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });
  onDestroy(() => document.removeEventListener("keydown", handleKey));
</script>

{#if card && ability}
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="counter-cost-title">
    <div class="prompt-modal counter-modal">
      <h2 id="counter-cost-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">remove counters</span>
      </h2>
      <p class="prompt-hint">{hint}</p>
      {#if choices.length === 0}
        <p class="prompt-hint error">Nothing you control can pay this cost.</p>
      {:else}
        <ul class="prompt-options">
          {#each choices as c (counterChoiceKey(c))}
            <li>
              {#if multi}
                <div class="prompt-opt stepper-row">
                  <span class="name">{nameOf(c.cardID)}</span>
                  <span class="note count">
                    {#if anyKind}{c.kind} ×{c.count}{:else}{c.count} {c.kind}{/if}
                  </span>
                  <button
                    type="button"
                    class="step"
                    aria-label="Remove one fewer from {rowOf(c)}"
                    disabled={(counts[counterChoiceKey(c)] ?? 0) === 0}
                    onclick={() => bump(c, -1)}>−</button
                  >
                  <span class="prompt-num">{counts[counterChoiceKey(c)] ?? 0}</span>
                  <button
                    type="button"
                    class="step"
                    aria-label="Remove one more from {rowOf(c)}"
                    disabled={(counts[counterChoiceKey(c)] ?? 0) >= ceilingFor(c)}
                    onclick={() => bump(c, 1)}>+</button
                  >
                </div>
              {:else}
                <button
                  type="button"
                  class="prompt-opt"
                  class:on={chosen === counterChoiceKey(c)}
                  aria-pressed={chosen === counterChoiceKey(c)}
                  onclick={() => (chosen = counterChoiceKey(c))}
                >
                  <span class="prompt-radio" aria-hidden="true"></span>
                  <span class="name">{nameOf(c.cardID)}</span>
                  <span class="note count">
                    {#if anyKind}{c.kind} ×{c.count}{:else}{c.count} {c.kind}{/if}
                  </span>
                </button>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
      <div class="prompt-foot">
        {#if multi}
          <span class="prompt-count" class:enough={ready}>{goal}</span>
        {/if}
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" disabled={!ready} onclick={confirm}
          >Remove <span class="kbd">↵</span></button
        >
      </div>
    </div>
  </div>
{/if}

<style>
  .counter-modal {
    width: min(440px, calc(100vw - 32px));
  }
  .name {
    flex: 1 1 auto;
  }
  .count {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--fg-muted);
  }
  .stepper-row {
    gap: 8px;
  }
  .step {
    min-width: 28px;
    padding: 2px 6px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: transparent;
    color: inherit;
    cursor: pointer;
  }
  .step:disabled {
    opacity: 0.4;
    cursor: default;
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
