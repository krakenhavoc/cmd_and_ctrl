<script lang="ts">
  // DiscardCostModal — S21 sub-PR 5: pick the cards that pay a
  // spell's additional cost ("As an additional cost to cast this
  // spell, discard a card").
  //
  // Opens first in the cast flow, before X / modes / targeting.
  // Not because the rules pay costs first — CR 601.2h comes after
  // targets at 601.2c — but because the whole cast rides one
  // cast_spell message, so collection order is a UI choice, and
  // this is the choice most likely to make a player back out.
  //
  // A cost is not a target: it can't be responded to, and nothing
  // can make a card in your own hand an illegal choice. So this is
  // a plain list rather than the board-click targeting flow, and
  // the same shape as SacrificeCostModal.
  import { onDestroy } from "svelte";
  import type { CardView } from "../../protocol";

  interface Props {
    // The spell being cast; null closes the modal.
    card: CardView | null;
    // Cards in the caster's hand that can pay — everything except
    // the spell itself (CR 601.2a puts it on the stack first).
    options: CardView[];
    onConfirm: (instanceIDs: string[]) => void;
    onCancel: () => void;
  }

  const { card, options, onConfirm, onCancel }: Props = $props();

  const need = $derived(card?.additional_cost?.discard_cards ?? 0);
  const label = $derived(card?.additional_cost?.label ?? "Discard a card");

  let chosen = $state<string[]>([]);

  // Reset when a different cast opens the prompt.
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = [];
    }
  });

  const ready = $derived(need > 0 && chosen.length === need);
  const short = $derived(options.length < need);

  function toggle(id: string): void {
    if (chosen.includes(id)) {
      chosen = chosen.filter((c) => c !== id);
      return;
    }
    if (chosen.length >= need) return;
    chosen = [...chosen, id];
  }

  function confirm(): void {
    if (!ready) return;
    onConfirm(chosen);
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

{#if card}
  <div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="discard-cost-title">
    <div class="modal">
      <h2 id="discard-cost-title">{card.name}</h2>
      <p class="prompt">
        {label}
        {#if need > 1}<span class="count">({chosen.length}/{need})</span>{/if}
      </p>
      {#if short}
        <p class="empty">
          You need {need} card{need === 1 ? "" : "s"} in hand to pay this cost.
        </p>
      {:else}
        <ul class="options">
          {#each options as c (c.instance_id)}
            <li>
              <button
                type="button"
                class="option"
                class:selected={chosen.includes(c.instance_id)}
                aria-pressed={chosen.includes(c.instance_id)}
                onclick={() => toggle(c.instance_id)}
              >
                <span class="mark" aria-hidden="true"
                  >{chosen.includes(c.instance_id) ? "●" : "○"}</span
                >
                <span class="name">{c.name}</span>
                {#if c.mana_cost}
                  <span class="cost">{c.mana_cost}</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
      <div class="actions">
        <button type="button" class="cancel" onclick={onCancel}>Cancel</button>
        <button type="button" class="confirm" disabled={!ready} onclick={confirm}>Discard</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.65);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1100;
  }
  .modal {
    background: #1c1f2a;
    color: #e6e8ee;
    border: 1px solid #3a4055;
    border-radius: 6px;
    padding: 1.25rem 1.5rem;
    min-width: 320px;
    max-width: 460px;
    box-shadow: 0 12px 28px rgba(0, 0, 0, 0.4);
  }
  h2 {
    margin: 0 0 0.25rem 0;
    font-size: 1.05rem;
    color: #f4ead5;
  }
  .prompt {
    margin: 0 0 0.75rem 0;
    color: #aab2c8;
    font-size: 0.9rem;
  }
  .count {
    color: #f4ead5;
  }
  .empty {
    margin: 0 0 0.9rem 0;
    color: #ff9a9a;
    font-size: 0.9rem;
  }
  .options {
    list-style: none;
    margin: 0 0 1rem 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    max-height: 40vh;
    overflow-y: auto;
  }
  .option {
    width: 100%;
    display: flex;
    align-items: baseline;
    gap: 0.6rem;
    text-align: left;
    padding: 0.45rem 0.65rem;
    border-radius: 4px;
    border: 1px solid #3a4055;
    background: #262a38;
    color: #e6e8ee;
    cursor: pointer;
    font-size: 0.92rem;
  }
  .option.selected {
    border-color: #ff9a9a;
    background: #3a2630;
  }
  .mark {
    flex: 0 0 auto;
    color: #ff9a9a;
  }
  .name {
    flex: 1 1 auto;
  }
  .cost {
    flex: 0 0 auto;
    color: #aab2c8;
    font-size: 0.85rem;
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }
  .actions button {
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    border: 1px solid #3a4055;
    background: #262a38;
    color: #e6e8ee;
    cursor: pointer;
  }
  .actions .confirm {
    background: #6b2f3a;
    border-color: #8c4351;
  }
  .actions .confirm:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
