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

  import Icon from "../Icon.svelte";
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
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="mode-picker-title">
    <div class="prompt-modal">
      <h2 id="mode-picker-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">
          {single ? "choose one" : `choose up to ${spec.max}`}
        </span>
      </h2>
      <p class="prompt-hint">
        {spec.prompt}
        {#if targetedCount(chosen) > 0}
          Each chosen mode with a target is picked next, in this order.
        {/if}
      </p>
      <ul class="prompt-options" role={single ? "radiogroup" : "group"}>
        {#each spec.options as option, i (i)}
          {@const castable = modeOptionCastable(option)}
          {@const selected = chosen.includes(i)}
          <li>
            <button
              type="button"
              class="prompt-opt"
              class:on={selected}
              class:off={!castable}
              role={single ? "radio" : "checkbox"}
              aria-checked={selected}
              aria-disabled={!castable}
              onclick={() => toggle(i, option)}
            >
              <span class="prompt-radio" aria-hidden="true"></span>
              <span class="label">{option.label}</span>
              {#if !castable}
                <span class="note">no legal target</span>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
      <div class="prompt-foot">
        {#if !single}
          <span class="prompt-count">{count} / {spec.max} modes</span>
        {/if}
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" disabled={!canConfirm} onclick={confirm}>
          {targetedCount(chosen) > 0 ? "Choose targets" : "Cast"}
          <Icon name="chevronRight" size={13} />
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .off {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .label {
    flex: 1 1 auto;
  }
</style>
