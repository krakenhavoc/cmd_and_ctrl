<script lang="ts">
  // TapCostModal — S22: pick the untapped permanents you tap to help
  // pay for a spell. Convoke ("your creatures can help cast this
  // spell") and waterbend ("you can tap your artifacts and creatures
  // to help") are the same prompt under two names, so they share one
  // picker; the label and the colour hint come off the wire.
  //
  // Opens after the X prompt and before targeting. After X because a
  // waterbend {X} cost has no size until X is announced — there is
  // nothing to cap the picker at before then. Before targeting
  // because paying is the decision that makes the spell castable at
  // all, and a player who finds they can't afford it should back out
  // before picking victims.
  //
  // A cost is not a target: it can't be responded to and hexproof
  // doesn't apply, so this is a plain list of your own permanents
  // rather than the board-click targeting flow — the same shape as
  // DiscardCostModal and SacrificeCostModal.
  //
  // Unlike those two this prompt is OPTIONAL in both directions: you
  // may tap none (and pay the whole cost with mana), and Confirm is
  // always enabled. There is no count to satisfy.

  import { onDestroy } from "svelte";
  import type { CardView, TapCostView } from "../../protocol";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    // The spell being cast; null closes the modal.
    card: CardView | null;
    // The clause, off the card's tap_cost.
    cost: TapCostView | null;
    // The caster's untapped permanents that may pay it.
    options: CardView[];
    // How many may be tapped. 0 means the cost can't take any — a
    // waterbend {0} — and the picker says so rather than offering
    // buttons that do nothing.
    limit: number;
    onConfirm: (instanceIDs: string[]) => void;
    onCancel: () => void;
  }

  const { card, cost, options, limit, onConfirm, onCancel }: Props = $props();

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

  const label = $derived(cost?.label ?? "Tap permanents to help pay");
  const hint = $derived(
    cost?.color_clause
      ? "Each creature you tap pays for {1} or one mana of that creature's colour."
      : "Each permanent you tap pays for {1}.",
  );
  const full = $derived(chosen.length >= limit);

  function toggle(id: string): void {
    if (chosen.includes(id)) {
      chosen = chosen.filter((c) => c !== id);
      return;
    }
    if (full) return;
    chosen = [...chosen, id];
  }

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

{#if card && cost}
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="tap-cost-title">
    <div class="prompt-modal tc-modal">
      <h2 id="tap-cost-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">{cost.key} · CR 601.2h</span>
      </h2>
      <p class="prompt-hint">{label}</p>
      <p class="prompt-hint sub">{hint}</p>
      {#if limit === 0}
        <p class="prompt-hint error">This cost has nothing to pay — tap nothing and continue.</p>
      {:else if options.length === 0}
        <p class="prompt-hint error">
          You control no untapped permanents that can help. Pay the cost with mana instead.
        </p>
      {:else}
        <ul class="prompt-options">
          {#each options as c (c.instance_id)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on={chosen.includes(c.instance_id)}
                disabled={full && !chosen.includes(c.instance_id)}
                aria-pressed={chosen.includes(c.instance_id)}
                onclick={() => toggle(c.instance_id)}
              >
                <span class="prompt-radio" aria-hidden="true"></span>
                <span class="name">{c.name}</span>
                {#if c.power !== undefined && c.toughness !== undefined}
                  <span class="note pt">{c.power}/{c.toughness}</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
      <div class="prompt-foot">
        <span class="prompt-count">{chosen.length} / {limit} tapped</span>
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" onclick={confirm}>
          {chosen.length === 0 ? "Tap nothing" : "Tap"}
          <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .tc-modal {
    width: min(440px, calc(100vw - 32px));
  }
  .name {
    flex: 1 1 auto;
  }
  .pt {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--fg-muted);
  }
  .sub {
    font-size: 12px;
    color: var(--fg-muted);
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
