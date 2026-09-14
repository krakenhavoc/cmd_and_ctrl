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
  import ModalLayer from "../ModalLayer.svelte";

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
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="discard-cost-title">
    <div class="prompt-modal dc-modal">
      <h2 id="discard-cost-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">additional cost · CR 601.2f</span>
      </h2>
      <p class="prompt-hint">{label}</p>
      {#if short}
        <p class="prompt-hint error">
          You need {need} card{need === 1 ? "" : "s"} in hand to pay this cost.
        </p>
      {:else}
        <ul class="prompt-options">
          {#each options as c (c.instance_id)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on={chosen.includes(c.instance_id)}
                aria-pressed={chosen.includes(c.instance_id)}
                onclick={() => toggle(c.instance_id)}
              >
                <span class="prompt-radio" aria-hidden="true"></span>
                <span class="name">{c.name}</span>
                {#if c.mana_cost}
                  <span class="note cost">{c.mana_cost}</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
      <div class="prompt-foot">
        {#if need > 1}
          <span class="prompt-count">{chosen.length} / {need} picked</span>
        {/if}
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" disabled={!ready} onclick={confirm}>
          Discard <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .dc-modal {
    width: min(440px, calc(100vw - 32px));
  }
  .name {
    flex: 1 1 auto;
  }
  .cost {
    font-family: var(--font-mono);
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
