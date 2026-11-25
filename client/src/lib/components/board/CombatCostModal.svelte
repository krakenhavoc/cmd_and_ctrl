<script lang="ts">
  // CombatCostModal — #6190: the permanents a NON-MANA cost to attack or
  // block is paid with (CR 508.1h, 509.1d).
  //
  // Exalted Dragon "can't attack unless you sacrifice a land", Hollow
  // Warrior "can't attack or block unless you tap an untapped creature
  // you control", Floodtide Serpent "can't attack unless you return an
  // enchantment you control to its owner's hand". The cost is part of
  // the declaration (never a prompt afterwards), so the sheet opens
  // BEFORE the action goes out and its answer rides it as
  // `sacrifice_ids` / `tap_ids` / `return_ids`.
  //
  // A cost is not a target: hexproof does not apply and nothing can be
  // responded to, so this is a plain list of the candidates the server
  // offered, like SacrificeCostModal's — one list per clause, because a
  // declaration can owe two different ones ("a land" and "two Islands").
  //
  // Cancelling sends nothing, which is CR 508.1d and CR 509.1c: a player
  // is never required to pay a cost to attack or block.
  //
  // ADR 0111 PR 6: a sheet in the action dock; the confirm and Cancel are
  // the dock's action bar.

  import { cancelAction, confirmAction, type DockAction } from "../../dock";
  import {
    canConfirmCombatCost,
    chooseCombatCostForMe,
    combatCostParams,
    combatCostUnpayable,
    combatCostVerb,
    emptyPicks,
    pickedElsewhere,
    togglePick,
    type CombatCostParams,
    type CombatCostPrompt,
  } from "../../combatCost";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // Null closes the sheet.
    prompt: CombatCostPrompt | null;
    onConfirm: (params: CombatCostParams) => void;
    onCancel: () => void;
  }

  const { prompt, onConfirm, onCancel }: Props = $props();

  let picks = $state<string[][]>([]);
  let lastKey: string | null = null;
  // A new question starts empty: which land to sacrifice is never
  // guessed for the player.
  $effect(() => {
    const key = prompt
      ? `${prompt.verb}:${prompt.asks.map((a) => `${a.creature}:${a.demand.kind}:${a.demand.clause ?? ""}`).join("|")}`
      : null;
    if (key !== lastKey) {
      lastKey = key;
      picks = prompt ? emptyPicks(prompt.asks) : [];
    }
  });

  const asks = $derived(prompt?.asks ?? []);
  const short = $derived(prompt ? combatCostUnpayable(asks) : null);
  const ready = $derived(canConfirmCombatCost(picks, asks));

  function pick(index: number, id: string): void {
    picks = togglePick(picks, asks, index, id);
  }

  function chooseForMe(): void {
    picks = chooseCombatCostForMe(asks);
  }

  function confirm(): void {
    if (!ready) return;
    onConfirm(combatCostParams(picks, asks));
  }

  const secondary = $derived<DockAction[]>([
    ...(short
      ? []
      : [
          {
            id: "choose-for-me",
            label: "Choose for me",
            title: "The first candidates the server listed. You still confirm.",
            disabled: false,
            onPress: chooseForMe,
          },
        ]),
    cancelAction(onCancel),
  ]);
  const picked = $derived(picks.reduce((n, row) => n + row.length, 0));
  const needed = $derived(asks.reduce((n, a) => n + a.demand.count, 0));
</script>

{#if prompt}
  <DockSheet
    label="Pay to attack"
    src="cost to attack or block"
    width={560}
    sheetKey={`combat-cost:${prompt.verb}:${asks.length}`}
    count={`${picked} / ${needed} picked`}
    primary={confirmAction(prompt.verb, confirm, { disabled: !ready })}
    {secondary}
  >
    {#if short}
      <p class="prompt-hint error">
        {combatCostVerb(short.demand)} to attack or block with {short.creatureName}: you control
        {short.options.length} of the {short.demand.count} permanents it needs.
      </p>
    {/if}
    {#each asks as ask, i (`${ask.creature}:${ask.demand.kind}:${i}`)}
      <p class="prompt-hint">
        {combatCostVerb(ask.demand)} for {ask.creatureName}{ask.demand.source &&
        ask.demand.source !== ask.creatureName
          ? ` (${ask.demand.source})`
          : ""}.
      </p>
      {#if ask.options.length === 0}
        <p class="prompt-hint error">Nothing you control can pay this.</p>
      {:else}
        <ul class="prompt-options">
          {#each ask.options as c (c.instance_id)}
            {@const on = (picks[i] ?? []).includes(c.instance_id)}
            {@const elsewhere = pickedElsewhere(picks, i, c.instance_id)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on
                aria-pressed={on}
                onclick={() => pick(i, c.instance_id)}
              >
                <span class="prompt-radio" aria-hidden="true"></span>
                <span class="name">{c.name}</span>
                {#if c.power !== undefined && c.toughness !== undefined}
                  <span class="note pt">{c.power}/{c.toughness}</span>
                {/if}
                {#if elsewhere}
                  <span class="note">paying another cost</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    {/each}
  </DockSheet>
{/if}

<style>
  .name {
    flex: 1 1 auto;
  }
  .pt {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--fg-muted);
  }
  .note {
    font-size: 12px;
    color: var(--fg-muted);
  }
</style>
