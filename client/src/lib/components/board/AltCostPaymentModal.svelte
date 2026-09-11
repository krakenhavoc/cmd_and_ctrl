<script lang="ts">
  // AltCostPaymentModal — S28: pick the card that pays the non-mana
  // half of a chosen alternative cost. Force of Will's "exile a blue
  // card from your hand", Daze's "return an Island you control",
  // Solitude's evoke pitch.
  //
  // SacrificeCostModal's plain option list, for the same reason: a
  // cost is not a target, so this is a list of your own cards rather
  // than the board-click targeting flow. It opens immediately after
  // the alternative-cost picker and before every other cost prompt,
  // matching the order the server validates them in.
  //
  // The life half of the same cost (Force of Will's 1) has no picker
  // — there is nothing to choose — so it is shown as a line of copy
  // and charged server-side.
  import { onDestroy } from "svelte";
  import type { AlternativeCostView, CardView } from "../../protocol";

  interface Props {
    // The spell being cast; null closes the modal.
    card: CardView | null;
    // The offer being paid, for its label and its life component.
    offer: AlternativeCostView | null;
    // The cards that can pay, already filtered by the server.
    options: CardView[];
    onConfirm: (instanceID: string) => void;
    onCancel: () => void;
  }

  const { card, offer, options, onConfirm, onCancel }: Props = $props();

  let chosen = $state<string | null>(null);

  // Reset when a different cast opens the prompt.
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = null;
    }
  });

  function confirm(): void {
    if (!chosen) return;
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

{#if card && offer}
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="alt-pay-title">
    <div class="prompt-modal ap-modal">
      <h2 id="alt-pay-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">alternative cost · CR 118.9</span>
      </h2>
      <p class="prompt-hint">
        {offer.label ?? offer.key}. Choose {offer.pay_label ?? "a card"}.
        {#if offer.life}
          You also pay {offer.life} life.
        {/if}
      </p>
      {#if options.length === 0}
        <p class="prompt-hint error">You have nothing that can pay this cost.</p>
      {:else}
        <ul class="prompt-options">
          {#each options as c (c.instance_id)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on={chosen === c.instance_id}
                aria-pressed={chosen === c.instance_id}
                onclick={() => (chosen = c.instance_id)}
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
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" disabled={!chosen} onclick={confirm}>
          Pay <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .ap-modal {
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
