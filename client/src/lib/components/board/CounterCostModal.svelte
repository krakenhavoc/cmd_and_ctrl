<script lang="ts">
  // CounterCostModal — #625: choose what pays a "remove N counters"
  // activation cost. Heart of Kiran's "remove a loyalty counter from a
  // planeswalker you control" with more than one walker out; Fain, the
  // Broker's "remove a counter from a creature you control", where the
  // kind is a choice too.
  //
  // SacrificeCostModal's shape — a cost is not a target, so this is a
  // plain list of your own permanents rather than the board-click
  // targeting flow, and it opens before any target prompt — with one
  // row per (permanent, kind) instead of per permanent, because for
  // "a counter" of any kind the kind is part of the answer. The rows
  // come straight off the server's counter_cost_options, most counters
  // first, and Board skips this modal entirely when there is only one.

  import { onDestroy } from "svelte";
  import type { ActivatedAbilityView, CardView } from "../../protocol";
  import { counterChoiceKey, counterChoices, type CounterChoice } from "../../counterCost";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    // The ability's source, for the heading; null closes the modal.
    card: CardView | null;
    ability: ActivatedAbilityView | null;
    // The battlefield, to name the options.
    board: CardView[];
    onConfirm: (choice: CounterChoice) => void;
    onCancel: () => void;
  }

  const { card, ability, board, onConfirm, onCancel }: Props = $props();

  let chosen = $state<string | null>(null);

  // Reset when a different activation opens the prompt.
  let lastKey: string | null = null;
  $effect(() => {
    const key = card && ability ? `${card.instance_id}:${ability.index}` : null;
    if (key !== lastKey) {
      lastKey = key;
      chosen = null;
    }
  });

  const choices = $derived(ability ? counterChoices(ability) : []);
  const anyKind = $derived(!ability?.counter_cost_kind);
  const n = $derived(ability?.counter_cost_n ?? 1);

  const hint = $derived.by(() => {
    if (!ability) return "";
    const kind = ability.counter_cost_kind ? `${ability.counter_cost_kind} ` : "";
    const what = n === 1 ? `a ${kind}counter` : `${n} ${kind}counters`;
    const from = ability.counter_cost_self
      ? "this permanent"
      : (ability.counter_cost_label ?? "a permanent you control");
    return `Remove ${what} from ${from} to pay for this ability.`;
  });

  function nameOf(id: string): string {
    return board.find((c) => c.instance_id === id)?.name ?? "a permanent";
  }

  function confirm(): void {
    const c = choices.find((x) => counterChoiceKey(x) === chosen);
    if (!c) return;
    onConfirm(c);
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
            </li>
          {/each}
        </ul>
      {/if}
      <div class="prompt-foot">
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" disabled={!chosen} onclick={confirm}>Remove</button>
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
</style>
