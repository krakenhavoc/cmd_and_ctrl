<script lang="ts">
  // AlternativeCostModal — S22: choose the cost a spell is cast for
  // when the card offers one paid INSTEAD of its mana cost. Overload,
  // evoke, cleave.
  //
  // The shape is DiscardCostModal's plain option list rather than the
  // board-click targeting flow, for the same reason: a cost is not a
  // target. What differs is that this prompt is optional. An
  // additional cost is a demand — a card that says "discard a card"
  // gives you no way out — whereas "you MAY cast this spell for its
  // overload cost" always leaves the printed cost available, so the
  // printed cost is listed first as an ordinary option and is the
  // default the keyboard confirms.
  //
  // It opens before every other cast prompt, and that is not a UI
  // preference the way DiscardCostModal's position is: overload and
  // cleave rewrite the target clause, so the answer here decides what
  // the targeting prompt after it is even allowed to offer.
  import { onDestroy } from "svelte";
  import type { CardView } from "../../protocol";
  import { alternativeCostsOf } from "../../targeting";

  interface Props {
    // The spell being cast; null closes the modal.
    card: CardView | null;
    // Fires with the chosen cost's key, or undefined for "pay the
    // printed mana cost".
    onConfirm: (key: string | undefined) => void;
    onCancel: () => void;
  }

  const { card, onConfirm, onCancel }: Props = $props();

  const offers = $derived(card ? alternativeCostsOf(card) : []);

  // undefined = the printed mana cost. Reset whenever a different
  // cast opens the prompt, so last turn's overload isn't preselected.
  let chosen = $state<string | undefined>(undefined);
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = undefined;
    }
  });

  function confirm(): void {
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
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="alt-cost-title">
    <div class="prompt-modal ac-modal">
      <h2 id="alt-cost-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">alternative cost · CR 118.9</span>
      </h2>
      <p class="prompt-hint">Cast this for which cost?</p>
      <ul class="prompt-options">
        <li>
          <button
            type="button"
            class="prompt-opt"
            class:on={chosen === undefined}
            aria-pressed={chosen === undefined}
            onclick={() => (chosen = undefined)}
          >
            <span class="prompt-radio" aria-hidden="true"></span>
            <span class="name">Its mana cost</span>
            {#if card.mana_cost}
              <span class="note cost">{card.mana_cost}</span>
            {/if}
          </button>
        </li>
        {#each offers as offer (offer.key)}
          <li>
            <button
              type="button"
              class="prompt-opt"
              class:on={chosen === offer.key}
              aria-pressed={chosen === offer.key}
              onclick={() => (chosen = offer.key)}
            >
              <span class="prompt-radio" aria-hidden="true"></span>
              <span class="name">{offer.label ?? offer.key}</span>
              {#if offer.mana_cost}
                <span class="note cost">{offer.mana_cost}</span>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
      <div class="prompt-foot">
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" onclick={confirm}>
          Cast <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .ac-modal {
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
