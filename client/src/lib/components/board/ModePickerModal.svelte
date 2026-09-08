<script lang="ts">
  // ModePickerModal — S20 sub-PR 4: choose the mode(s) of a modal
  // spell ("Choose one —", "Choose two —") before targeting / cast.
  // Radio buttons for choose-one, checkboxes for choose-N. An option
  // that targets and has no legal target right now is disabled with
  // a "no legal target" note; the Confirm button enables once the
  // count is within min..max.
  //
  // Confirm hands the chosen indexes back to Board, which either
  // enters targeting for the (single) targeted option or fires
  // cast_spell straight away.

  import { onDestroy } from "svelte";
  import type { CardView, ModeOptionView } from "../../protocol";
  import { modeOptionCastable } from "../../targeting";

  interface Props {
    card: CardView | null;
    onConfirm: (modes: number[]) => void;
    onCancel: () => void;
  }

  const { card, onConfirm, onCancel }: Props = $props();

  const spec = $derived(card?.modes ?? null);
  const single = $derived((spec?.max ?? 1) === 1);

  let chosen = $state<number[]>([]);

  // Reset when a different card opens the prompt.
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = [];
    }
  });

  const count = $derived(chosen.length);
  const canConfirm = $derived(
    spec !== null && count >= spec.min && count <= (spec.max > 0 ? spec.max : count),
  );

  function targetedCount(indexes: number[]): number {
    if (!spec) return 0;
    return indexes.filter((i) => spec.options[i]?.legal_targets !== undefined).length;
  }

  function toggle(i: number, option: ModeOptionView): void {
    if (!spec || !modeOptionCastable(option)) return;
    if (single) {
      chosen = [i];
      return;
    }
    if (chosen.includes(i)) {
      chosen = chosen.filter((x) => x !== i);
      return;
    }
    if (spec.max > 0 && chosen.length >= spec.max) return;
    // Sub-PR 4: one targeted option per cast.
    if (targetedCount([...chosen, i]) > 1) return;
    chosen = [...chosen, i].sort((a, b) => a - b);
  }

  function confirm(): void {
    if (!card || !canConfirm) return;
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

{#if card && spec}
  <div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="mode-picker-title">
    <div class="modal">
      <h2 id="mode-picker-title">{card.name}</h2>
      <p class="prompt">
        {spec.prompt}
        {#if !single}
          <span class="count">({count}/{spec.max})</span>
        {/if}
      </p>
      <ul class="options" role={single ? "radiogroup" : "group"}>
        {#each spec.options as option, i (i)}
          {@const castable = modeOptionCastable(option)}
          {@const selected = chosen.includes(i)}
          <li>
            <button
              type="button"
              class="option"
              class:selected
              class:disabled={!castable}
              role={single ? "radio" : "checkbox"}
              aria-checked={selected}
              aria-disabled={!castable}
              onclick={() => toggle(i, option)}
            >
              <span class="mark" aria-hidden="true">{selected ? "●" : "○"}</span>
              <span class="label">{option.label}</span>
              {#if !castable}
                <span class="note">no legal target</span>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
      <div class="actions">
        <button type="button" class="cancel" onclick={onCancel}>Cancel</button>
        <button type="button" class="confirm" disabled={!canConfirm} onclick={confirm}>
          {targetedCount(chosen) > 0 ? "Choose target" : "Cast"}
        </button>
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
    min-width: 340px;
    max-width: 520px;
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
    color: #7f8aa8;
  }
  .options {
    list-style: none;
    margin: 0 0 1rem 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }
  .option {
    width: 100%;
    display: flex;
    align-items: baseline;
    gap: 0.6rem;
    text-align: left;
    padding: 0.5rem 0.65rem;
    border-radius: 4px;
    border: 1px solid #3a4055;
    background: #262a38;
    color: #e6e8ee;
    cursor: pointer;
    font-size: 0.92rem;
    line-height: 1.3;
  }
  .option.selected {
    border-color: #6fe3a4;
    background: #22392c;
  }
  .option.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .mark {
    flex: 0 0 auto;
    color: #6fe3a4;
  }
  .label {
    flex: 1 1 auto;
  }
  .note {
    flex: 0 0 auto;
    font-size: 0.75rem;
    color: #ff9a9a;
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
    background: #2d5a3f;
    border-color: #3f7a55;
  }
  .actions .confirm:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
