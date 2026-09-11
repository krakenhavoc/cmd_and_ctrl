<script lang="ts">
  // CrewCostModal — S27: pick the creatures you tap to crew a
  // Vehicle (CR 702.122a).
  //
  // Shaped like TapCostModal and deliberately not the same component,
  // because the arithmetic is the opposite way round. Convoke counts
  // PERMANENTS against a ceiling — tap at most N, tapping none is
  // fine. Crew counts POWER against a floor — tap any number, but the
  // total has to reach the crew number or the ability cannot be
  // activated at all. So this modal tracks a running power total, the
  // confirm button is disabled until it clears the bar, and there is
  // no "crew with nothing" escape hatch.
  //
  // Two things the option list gets right that are easy to get wrong:
  //
  //   - Summoning-sick creatures ARE offered. Tapping a creature to
  //     crew is not paying a {T} cost, so a creature cast this turn
  //     may crew (CR 702.122b). The server agrees; this list comes
  //     straight off its crew_options.
  //   - Crewing is a cost, not a target, so hexproof and protection
  //     never apply and the list is a plain roster of your own
  //     untapped creatures rather than the board-click targeting
  //     flow — the same shape as SacrificeCostModal.

  import { onDestroy } from "svelte";
  import type { ActivatedAbilityView, CardView } from "../../protocol";
  import { crewAvailablePower, crewPower, crewSatisfied } from "../../crew";

  interface Props {
    // The Vehicle being crewed; null closes the modal.
    card: CardView | null;
    // The crew ability, for its number and its label.
    ability: ActivatedAbilityView | null;
    // The creatures the server says could pay it right now.
    options: CardView[];
    onConfirm: (instanceIDs: string[]) => void;
    onCancel: () => void;
  }

  const { card, ability, options, onConfirm, onCancel }: Props = $props();

  let chosen = $state<string[]>([]);

  // Reset when a different activation opens the prompt.
  let lastKey: string | null = null;
  $effect(() => {
    const key = card && ability ? `${card.instance_id}:${ability.index}` : null;
    if (key !== lastKey) {
      lastKey = key;
      chosen = [];
    }
  });

  const need = $derived(ability?.crew_cost ?? 0);

  // Power is read off the CardView, which already carries the
  // post-layer effective value the server will re-check against — so
  // the running total shown here and the total the server computes
  // agree without a second round trip. A card with no power (it
  // should not appear here) contributes nothing rather than NaN.
  const total = $derived(crewPower(options, chosen));
  const enough = $derived(crewSatisfied(options, chosen, need));

  // The best the whole board could do. When it falls short the modal
  // says so plainly instead of letting a player click around looking
  // for the creature that would finish the total.
  const available = $derived(crewAvailablePower(options));

  function toggle(id: string): void {
    chosen = chosen.includes(id) ? chosen.filter((c) => c !== id) : [...chosen, id];
  }

  function confirm(): void {
    if (!enough) return;
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

{#if card && ability}
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="crew-cost-title">
    <div class="prompt-modal crew-modal">
      <h2 id="crew-cost-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">Crew {need} · CR 702.122</span>
      </h2>
      <p class="prompt-hint">
        Tap any number of untapped creatures you control with total power {need} or more.
      </p>
      {#if options.length === 0}
        <p class="prompt-hint error">You control no untapped creatures to crew with.</p>
      {:else if available < need}
        <p class="prompt-hint error">
          Your untapped creatures total {available} power — not enough to crew {need}.
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
                {#if c.power !== undefined && c.toughness !== undefined}
                  <span class="note pt">{c.power}/{c.toughness}</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
      <div class="prompt-foot">
        <span class="prompt-count" class:enough>{total} / {need} power</span>
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" disabled={!enough} onclick={confirm}>
          Crew
          <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .crew-modal {
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
  .prompt-count.enough {
    color: var(--accent);
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
