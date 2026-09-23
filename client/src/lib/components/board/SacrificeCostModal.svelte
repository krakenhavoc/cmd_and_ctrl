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
    canConfirmSacrificeRange,
    chooseForMeState,
    chooseSacrificeForMe,
    keepAvailablePicks,
    sacrificeCeiling,
    toggleSacrificePickInRange,
  } from "../../sacrificeCost";
  import type { SacrificeRange } from "../../sacrificeCost";
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
    // #1213: the clause's FLOOR, when it differs from `count` — an
    // open count ("sacrifice one or more artifacts") is min 1 with no
    // ceiling, and a clause whose count is the announced X has no
    // printed bounds at all. Absent means the fixed clause every
    // caller had before, where the floor and the ceiling are `count`.
    min?: number;
    // #1213: the verb, for the hint and the confirm button. A
    // return-to-hand cost is the same picker one verb over, so it
    // says "Return" rather than "Sacrifice".
    verb?: string;
    onConfirm: (instanceIDs: string[]) => void;
    onCancel: () => void;
  }

  const {
    source,
    label,
    options,
    count = 1,
    min,
    verb = "Sacrifice",
    onConfirm,
    onCancel,
  }: Props = $props();

  // The bounds this picker enforces. A caller that passes only
  // `count` gets N..N, which is exactly what it got before #1213.
  const range = $derived<SacrificeRange>({ min: min ?? count, max: count });
  const ceiling = $derived(sacrificeCeiling(range, options.length));

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

  // A picked permanent that leaves the battlefield while the picker is
  // open (a chosen Treasure destroyed in response) drops out of the
  // selection, so its slot frees up and a confirm never sends it.
  $effect(() => {
    const kept = keepAvailablePicks(
      chosen,
      options.map((c) => c.instance_id),
    );
    if (kept !== chosen) chosen = kept;
  });

  const ready = $derived(canConfirmSacrificeRange(chosen, range, options.length));
  const short = $derived(options.length < range.min);
  const chooseForMeButton = $derived(chooseForMeState(range.min, options.length));

  function pick(id: string): void {
    chosen = toggleSacrificePickInRange(chosen, id, ceiling);
  }

  function chooseForMe(): void {
    chosen = chooseSacrificeForMe(
      options.map((c) => c.instance_id),
      range.min,
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
      <p class="prompt-hint">{verb} {label} to pay for this ability.</p>
      {#if options.length === 0}
        <p class="prompt-hint error">Nothing you control can pay this cost.</p>
      {:else}
        {#if short}
          <p class="prompt-hint error">
            You control {options.length} of the {range.min} permanents this cost needs.
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
                disabled={!on && ceiling > 1 && chosen.length >= ceiling}
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
          <span class="prompt-count" aria-live="polite"
            >{chosen.length} / {range.max > 0 ? range.max : `${range.min}+`} picked</span
          >
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
        <button type="button" class="primary" disabled={!ready} onclick={confirm}>{verb}</button>
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
