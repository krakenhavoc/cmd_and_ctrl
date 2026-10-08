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
  //
  // ADR 0111 PR 6: a sheet in the action dock; Sacrifice and Cancel are
  // the dock's action bar (Enter / Escape through its one key handler).

  import type { AltCostPriceView, CardView, SacrificeGroupView } from "../../protocol";
  import {
    canConfirmSacrificeRange,
    canFillEachOf,
    chooseForMeState,
    chooseSacrificeForMe,
    chooseSacrificeSetForMe,
    fillsEachOf,
    keepAvailablePicks,
    sacrificeCeiling,
    toggleSacrificePickInRange,
  } from "../../sacrificeCost";
  import type { SacrificeRange } from "../../sacrificeCost";
  import { cancelAction, confirmAction, type DockAction } from "../../dock";
  import DockSheet from "./DockSheet.svelte";

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
    // ADR 0100 §3: the number picked is the spell's X ("sacrifice X
    // lands"), which the caller sends as the x_value.
    countIsX?: boolean;
    // #2526: the clause's set rule ("a Swamp and a Forest") — the picks
    // must fill every part with a different permanent. Confirm stays
    // shut until they do, and "Choose for me" fills a set that does.
    eachOf?: SacrificeGroupView[];
    // ADR 0135 §1: the spell this pays for, when the picker pays a
    // spell's alternative cost rather than an ability's cost. The hint
    // then reads "Tap an untapped creature you control to cast Orim's
    // Cure."
    castName?: string;
    // ADR 0135 §4: an emerge offer's price per candidate (its mana value
    // and what the spell costs with it sacrificed), from the server's
    // pricer. Shown beside each option, so the player sees what each
    // creature saves.
    prices?: Record<string, AltCostPriceView>;
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
    countIsX = false,
    eachOf,
    castName,
    prices,
    onConfirm,
    onCancel,
  }: Props = $props();

  // The bounds this picker enforces. A caller that passes only
  // `count` gets N..N, which is exactly what it got before #1213.
  const range = $derived<SacrificeRange>({ min: min ?? count, max: count });
  // ADR 0135 §1: a spell's alternative cost names the spell it pays for.
  const hint = $derived(
    `${verb} ${label} ${castName ? `to cast ${castName}.` : "to pay for this ability."}`,
  );
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

  const ready = $derived(
    canConfirmSacrificeRange(chosen, range, options.length) && fillsEachOf(chosen, eachOf),
  );
  const short = $derived(options.length < range.min || !canFillEachOf(eachOf));
  const chooseForMeButton = $derived(chooseForMeState(range.min, options.length));

  function pick(id: string): void {
    chosen = toggleSacrificePickInRange(chosen, id, ceiling, range.min);
  }

  function chooseForMe(): void {
    const ids = options.map((c) => c.instance_id);
    chosen =
      eachOf && eachOf.length > 0
        ? chooseSacrificeSetForMe(ids, eachOf)
        : chooseSacrificeForMe(ids, range.min);
  }

  function confirm(): void {
    if (!ready) return;
    onConfirm([...chosen]);
  }

  const secondary = $derived<DockAction[]>([
    ...(chooseForMeButton.shown
      ? [
          {
            id: "choose-for-me",
            label: "Choose for me",
            title: "Tokens first, then the lowest mana value. You still confirm.",
            disabled: chooseForMeButton.disabled,
            onPress: chooseForMe,
          },
        ]
      : []),
    cancelAction(onCancel),
  ]);
</script>

{#if source}
  <DockSheet
    label={source.name}
    src={castName ? "alternative cost" : "additional cost"}
    width={560}
    sheetKey={`sac:${source.instance_id}:${verb}:${label}`}
    count={chooseForMeButton.shown
      ? `${chosen.length} / ${range.max > 0 ? range.max : `${range.min}+`} picked`
      : undefined}
    primary={confirmAction(verb, confirm, { disabled: !ready })}
    {secondary}
  >
    <p class="prompt-hint">{hint}</p>
    {#if eachOf && eachOf.length > 0}
      <p class="prompt-hint">
        One permanent for each part: {eachOf.map((g) => g.label).join(", ")}.
      </p>
    {/if}
    {#if countIsX}
      <p class="prompt-hint">The number you pick is X.</p>
    {:else if range.min === 0 && options.length > 0}
      <p class="prompt-hint">Pick as many as you like, or none.</p>
    {/if}
    {#if options.length === 0}
      <p class="prompt-hint error">Nothing you control can pay this cost.</p>
    {:else}
      {#if short}
        <p class="prompt-hint error">
          {#if options.length < range.min}
            You control {options.length} of the {range.min} permanents this cost needs.
          {:else}
            You don't control a different permanent for every part of this cost.
          {/if}
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
              {#if prices?.[c.instance_id]}
                {@const p = prices[c.instance_id]}
                <span
                  class="note price"
                  title={`Mana value ${p.mana_value}: the spell costs ${p.price} with this sacrificed`}
                  >MV {p.mana_value} · costs {p.price}</span
                >
              {/if}
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </DockSheet>
{/if}

<style>
  .name {
    flex: 1 1 auto;
  }
  .pt,
  .price {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--fg-muted);
  }
</style>
