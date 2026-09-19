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
  // overload cost" leaves the printed cost available, so it is listed
  // first as an ordinary option and is the default the keyboard
  // confirms.
  //
  // #1012: "leaves the printed cost available" is not true of every
  // cast, and the modal used to assume it was. A flashback cast out
  // of the graveyard may not be announced at the cost in the card's
  // corner (CR 702.34b, rule 3 of validateCastPathLocked), so "Its
  // mana cost" was a preselected default the server would refuse with
  // ErrCastCostRequired. The server now says so —
  // `alternative_cost_required` — and the row is dropped rather than
  // disabled: an option that cannot ever be taken for this cast is
  // not a choice the player declined, it is one the card does not
  // offer here.
  //
  // It opens before every other cast prompt, and that is not a UI
  // preference the way DiscardCostModal's position is: overload and
  // cleave rewrite the target clause, so the answer here decides what
  // the targeting prompt after it is even allowed to offer.
  //
  // ADR 0073 (#664) puts the OPTIONAL ADDITIONAL costs in the same
  // prompt rather than in a modal of their own — kicker, multikicker,
  // buyback. They are the same question asked at the same moment
  // ("what am I paying for this?"), CR 601.2b announces them
  // together, and a second modal would be one more click for no extra
  // decision. They also COMPOSE with the radio list above them: an
  // alternative cost REPLACES the mana cost and an optional one ADDS
  // to whichever cost is being paid, so the toggles stay live
  // whichever radio is selected.
  //
  // A card with optional costs and no alternative costs opens this
  // same modal with only the add-ons showing, which is why the
  // heading and the hint are written for both.
  import { onDestroy } from "svelte";
  import type { CardView } from "../../protocol";
  import {
    alternativeCostsOf,
    optionalCostMaxTimes,
    optionalCostPayOptions,
    optionalCostSelection,
    optionalCostsOf,
    printedCostClaimable,
  } from "../../targeting";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    // The spell being cast; null closes the modal.
    card: CardView | null;
    // Fires with the chosen cost's key (or undefined for "pay the
    // printed mana cost") and the optional costs being paid, as
    // repeated indices.
    onConfirm: (key: string | undefined, optional: number[]) => void;
    onCancel: () => void;
  }

  const { card, onConfirm, onCancel }: Props = $props();

  const offers = $derived(card ? alternativeCostsOf(card) : []);
  const addOns = $derived(card ? optionalCostsOf(card) : []);
  // #1012: whether "Its mana cost" is a row at all.
  const printedOK = $derived(card ? printedCostClaimable(card) : true);

  // How many times each optional cost is being paid, by index. A Map
  // rather than an array so the wire shape is built in exactly one
  // place (optionalCostSelection) and the picker never has to think
  // about repeated indices.
  let paying = $state(new Map<number, number>());

  // An offer whose sacrifice list is present-and-empty cannot be
  // taken right now — a Constant Mists with no land. Disabled rather
  // than hidden: the card prints the cost, and hiding it would look
  // like a bug.
  function unpayable(index: number): boolean {
    const offer = addOns[index];
    if (!offer) return true;
    const options = optionalCostPayOptions(offer);
    return options !== undefined && options.length === 0;
  }

  function timesPaid(index: number): number {
    return paying.get(index) ?? 0;
  }

  function setTimes(index: number, n: number): void {
    const offer = addOns[index];
    if (!offer) return;
    const capped = Math.max(0, Math.min(n, optionalCostMaxTimes(offer)));
    const next = new Map(paying);
    if (capped === 0) next.delete(index);
    else next.set(index, capped);
    paying = next;
  }

  function toggle(index: number): void {
    setTimes(index, timesPaid(index) > 0 ? 0 : 1);
  }

  // undefined = the printed mana cost. Reset whenever a different
  // cast opens the prompt, so last turn's overload isn't preselected.
  //
  // #1012: when the printed cost is not claimable from this zone the
  // default is the FIRST offer instead, because `undefined` there is
  // an announcement the server refuses and the confirm button must
  // never start on one.
  let chosen = $state<string | undefined>(undefined);
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = printedOK ? undefined : offers[0]?.key;
      paying = new Map();
    }
  });

  function confirm(): void {
    onConfirm(chosen, optionalCostSelection(paying));
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
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="alt-cost-title">
    <div class="prompt-modal ac-modal">
      <h2 id="alt-cost-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">alternative cost · CR 118.9</span>
      </h2>
      <p class="prompt-hint">
        {offers.length === 0
          ? "Pay any additional costs?"
          : printedOK
            ? "Cast this for which cost?"
            : "Cast this for which cost? Its mana cost can't be paid from here."}
      </p>
      {#if offers.length > 0}
        <ul class="prompt-options">
          {#if printedOK}
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
          {/if}
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
      {/if}
      {#if addOns.length > 0}
        <p class="prompt-hint add-on-hint">
          Additional costs <span class="prompt-src" aria-hidden="true">CR 601.2b</span>
        </p>
        <ul class="prompt-options">
          {#each addOns as offer (offer.index)}
            <li>
              {#if optionalCostMaxTimes(offer) > 1}
                <div class="prompt-opt stepper" class:on={timesPaid(offer.index) > 0}>
                  <span class="name">{offer.label ?? offer.key}</span>
                  {#if offer.mana_cost}
                    <span class="note cost">{offer.mana_cost}</span>
                  {/if}
                  <button
                    type="button"
                    class="step"
                    aria-label={`Pay ${offer.label ?? offer.key} one fewer time`}
                    disabled={timesPaid(offer.index) === 0}
                    onclick={() => setTimes(offer.index, timesPaid(offer.index) - 1)}>-</button
                  >
                  <span class="times" aria-live="polite">x{timesPaid(offer.index)}</span>
                  <button
                    type="button"
                    class="step"
                    aria-label={`Pay ${offer.label ?? offer.key} one more time`}
                    disabled={unpayable(offer.index) ||
                      timesPaid(offer.index) >= optionalCostMaxTimes(offer)}
                    onclick={() => setTimes(offer.index, timesPaid(offer.index) + 1)}>+</button
                  >
                </div>
              {:else}
                <button
                  type="button"
                  class="prompt-opt"
                  class:on={timesPaid(offer.index) > 0}
                  aria-pressed={timesPaid(offer.index) > 0}
                  disabled={unpayable(offer.index)}
                  onclick={() => toggle(offer.index)}
                >
                  <span class="prompt-radio" aria-hidden="true"></span>
                  <span class="name">{offer.label ?? offer.key}</span>
                  {#if offer.mana_cost}
                    <span class="note cost">{offer.mana_cost}</span>
                  {/if}
                </button>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
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
  .add-on-hint {
    margin-top: 12px;
  }
  .stepper {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
  }
  .step {
    min-width: 28px;
  }
  .times {
    font-family: var(--font-mono);
    min-width: 2.5em;
    text-align: center;
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
