<script lang="ts">
  // ModePickerModal — S20 sub-PR 4: choose the mode(s) of a modal
  // spell ("Choose one —", "Choose two —") before targeting / cast.
  // Radio buttons for choose-one, checkboxes for choose-N. An option
  // that targets and has no legal target right now is disabled with
  // a "no legal target" note; the Confirm button enables once the
  // count is within min..max.
  //
  // Confirm hands the chosen indexes back to Board, which walks every
  // chosen bullet's target clauses and then fires cast_spell.
  //
  // #764 changed three things here, and each was a rule the picker
  // was getting wrong:
  //
  //   - It refused a SECOND targeted option. Kolaghan's Command
  //     ("choose two", every bullet targets) could not be cast at
  //     all, because the server could not carry two target groups.
  //   - It SORTED the chosen indexes ascending, which threw away the
  //     order the modes resolve in (CR 700.2c).
  //   - It had no notion of choosing the same bullet twice
  //     (CR 700.2d, Mystic Confluence), which is a count per option
  //     rather than a toggle.

  import Icon from "../Icon.svelte";
  import { onDestroy } from "svelte";
  import type { CardView, ModeOptionView } from "../../protocol";
  import { modeOptionCastable } from "../../targeting";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    card: CardView | null;
    onConfirm: (modes: number[]) => void;
    onCancel: () => void;
  }

  const { card, onConfirm, onCancel }: Props = $props();

  const spec = $derived(card?.modes ?? null);
  const single = $derived((spec?.max ?? 1) === 1 && !(spec?.repeatable ?? false));
  const repeatable = $derived(spec?.repeatable ?? false);

  // chosen is the multiset of option indexes IN THE ORDER CHOSEN —
  // which is the order they resolve in (CR 700.2c) and the order
  // their targets are asked for.
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

  // Spree (CR 702.172a, ADR 0065's 2026-09-23 amendment): the chosen
  // modes' own costs, in announce order, for the "what am I paying"
  // summary beside Confirm. Plain concatenation — the server prices
  // the actual sum (#544); this is a preview, not a computation.
  const extraCosts = $derived(
    chosen.map((i) => spec?.options[i]?.cost).filter((c): c is string => !!c),
  );

  function targetedCount(indexes: number[]): number {
    if (!spec) return 0;
    return indexes.filter((i) => spec.options[i]?.legal_targets !== undefined).length;
  }

  // countOf is how many times an option has been chosen — 0 or 1 for
  // an ordinary modal card, more only under CR 700.2d.
  function countOf(i: number): number {
    return chosen.filter((x) => x === i).length;
  }

  function toggle(i: number, option: ModeOptionView): void {
    if (!spec || !modeOptionCastable(option)) return;
    if (single) {
      chosen = [i];
      return;
    }
    if (!repeatable && chosen.includes(i)) {
      chosen = chosen.filter((x) => x !== i);
      return;
    }
    if (spec.max > 0 && chosen.length >= spec.max) return;
    // Appended, never sorted: the order the player clicks in is the
    // order CR 700.2c resolves in and the order the targets are
    // asked for (#764).
    chosen = [...chosen, i];
  }

  // remove takes one occurrence of a repeated bullet back off.
  function remove(i: number): void {
    const at = chosen.lastIndexOf(i);
    if (at < 0) return;
    chosen = [...chosen.slice(0, at), ...chosen.slice(at + 1)];
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
  <ModalLayer />
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
        {#if repeatable}
          You may choose the same mode more than once.
        {/if}
        {#if targetedCount(chosen) > 0}
          Each chosen mode with a target is picked next, in this order.
        {/if}
      </p>
      <ul class="prompt-options" role={single ? "radiogroup" : "group"}>
        {#each spec.options as option, i (i)}
          {@const castable = modeOptionCastable(option)}
          {@const times = countOf(i)}
          {@const selected = times > 0}
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
              {#if option.cost}
                <span class="cost" title={`additional cost ${option.cost}`}>+{option.cost}</span>
              {/if}
              {#if repeatable && times > 0}
                <span class="times">&times;{times}</span>
              {/if}
              {#if !castable}
                <span class="note">no legal target</span>
              {/if}
            </button>
            {#if repeatable && times > 0}
              <button
                type="button"
                class="ghost minus"
                aria-label={`Take back one ${option.label}`}
                onclick={() => remove(i)}>&minus;</button
              >
            {/if}
          </li>
        {/each}
      </ul>
      <div class="prompt-foot">
        {#if !single}
          <span class="prompt-count">{count} / {spec.max} modes</span>
        {/if}
        {#if extraCosts.length > 0}
          <span class="cost extra-cost" title="additional cost of the chosen modes"
            >+{extraCosts.join(" ")}</span
          >
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
  .times {
    font-variant-numeric: tabular-nums;
    opacity: 0.8;
  }
  .minus {
    padding: 0 0.5rem;
  }
  .cost {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--fg-dim);
    white-space: nowrap;
  }
  .extra-cost {
    margin-right: auto;
  }
</style>
