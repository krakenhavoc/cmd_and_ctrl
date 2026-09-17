<script lang="ts">
  // SacrificeCostModal — S21 sub-PR 2: pick the permanents that pay
  // a sacrifice cost ("Sacrifice a creature:", and since #747
  // "Sacrifice two artifacts:").
  //
  // A cost is not a target: it can't be responded to, and hexproof
  // doesn't apply. So this is a plain list of your own permanents
  // rather than the board-click targeting flow — and it opens
  // BEFORE any target prompt, matching the order costs are paid in
  // (CR 601.2h).
  //
  // #747: `count` is the clause's N (sacrifice_options.max). At 1 the
  // picker behaves as it always did — a click picks, Enter confirms.
  // At N it is a multi-select that confirms at exactly N, plus a
  // "Choose for me" button that fills the selection with the first N
  // options. The options arrive in the server's payment order (tokens
  // first, then lower mana value, then the ability's own source), so
  // the button picks what the bots would. It never confirms for the
  // player: they can change the picks before pressing Sacrifice.

  import { onDestroy } from "svelte";
  import type { CardView } from "../../protocol";
  import {
    canConfirmSacrifice,
    chooseForMeState,
    chooseSacrificeForMe,
    toggleSacrificePick,
  } from "../../sacrificeCost";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    // The ability's source, for the heading.
    source: CardView | null;
    // The clause ("a creature", "three Foods") and the permanents
    // that can pay it, in payment order.
    label: string;
    options: CardView[];
    // How many permanents the clause sacrifices. 1 unless the view
    // said otherwise.
    count?: number;
    onConfirm: (instanceIDs: string[]) => void;
    onCancel: () => void;
  }

  const { source, label, options, count = 1, onConfirm, onCancel }: Props = $props();

  let chosen = $state<string[]>([]);

  // Reset when a different activation opens the prompt.
  let lastSourceID: string | null = null;
  $effect(() => {
    const id = source?.instance_id ?? null;
    if (id !== lastSourceID) {
      lastSourceID = id;
      chosen = [];
    }
  });

  const ready = $derived(canConfirmSacrifice(chosen, count));
  const short = $derived(options.length < count);
  const chooseForMeButton = $derived(chooseForMeState(count, options.length));

  function pick(id: string): void {
    chosen = toggleSacrificePick(chosen, id, count);
  }

  function chooseForMe(): void {
    chosen = chooseSacrificeForMe(
      options.map((c) => c.instance_id),
      count,
    );
  }

  function confirm(): void {
    if (!ready) return;
    onConfirm([...chosen]);
  }

  function handleKey(e: KeyboardEvent): void {
    if (!source) return;
    if (e.key === "Enter") {
      e.preventDefault();
      confirm();
    } else if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    }
  }
  $effect(() => {
    if (!source) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });
  onDestroy(() => document.removeEventListener("keydown", handleKey));
</script>

{#if source}
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="sac-title">
    <div class="prompt-modal sac-modal">
      <h2 id="sac-title">
        {source.name}
        <span class="prompt-src" aria-hidden="true">additional cost</span>
      </h2>
      <p class="prompt-hint">Sacrifice {label} to pay for this ability.</p>
      {#if options.length === 0}
        <p class="prompt-hint error">Nothing you control can pay this cost.</p>
      {:else}
        {#if short}
          <p class="prompt-hint error">
            You control {options.length} of the {count} permanents this cost needs.
          </p>
        {/if}
        <ul class="prompt-options">
          {#each options as c (c.instance_id)}
            {@const on = chosen.includes(c.instance_id)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on
                aria-pressed={on}
                disabled={!on && count > 1 && chosen.length >= count}
                onclick={() => pick(c.instance_id)}
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
        {#if chooseForMeButton.shown}
          <span class="prompt-count" aria-live="polite">{chosen.length} / {count} picked</span>
          <button
            type="button"
            class="ghost"
            disabled={chooseForMeButton.disabled}
            title="Tokens first, then the lowest mana value. You still confirm."
            onclick={chooseForMe}>Choose for me</button
          >
        {/if}
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" disabled={!ready} onclick={confirm}>Sacrifice</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .sac-modal {
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
</style>
