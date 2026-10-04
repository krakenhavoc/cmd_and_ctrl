<script lang="ts">
  import ManaCost from "./ManaCost.svelte";
  // AltCostPaymentModal — S28: pick the card(s) that pay the
  // non-mana half of a chosen alternative cost. Force of Will's
  // "exile a blue card from your hand", Daze's "return an Island you
  // control", Solitude's evoke pitch — and, since S29, escape's
  // "exile five other cards from your graveyard".
  //
  // SacrificeCostModal's plain option list, for the same reason: a
  // cost is not a target, so this is a list of your own cards rather
  // than the board-click targeting flow. It opens immediately after
  // the alternative-cost picker and before every other cost prompt,
  // matching the order the server validates them in.
  //
  // S29 made the count a variable rather than a constant one. The
  // number comes off `pay_options.min` and nothing here knows which
  // keyword is being paid: at one the list behaves as it always did,
  // and above one it becomes a checklist that will not confirm until
  // exactly that many are ticked. "Exactly" is the server's rule too
  // — an escape cost that took four cards for a five-card price
  // would be a cheaper card than the one printed.
  //
  // The life half of the same cost (Force of Will's 1) has no picker
  // — there is nothing to choose — so it is shown as a line of copy
  // and charged server-side.
  //
  // ADR 0111 PR 6: a sheet in the action dock; the confirm and Cancel
  // are the dock's action bar (Enter / Escape through its one key
  // handler).
  import type { AlternativeCostView, CardView } from "../../protocol";
  import { altCostPayCount } from "../../targeting";
  import { cancelAction, confirmAction } from "../../dock";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // The spell being cast; null closes the modal.
    card: CardView | null;
    // The offer being paid, for its label, its life component and
    // how many cards it demands.
    offer: AlternativeCostView | null;
    // The cards that can pay, already filtered by the server.
    options: CardView[];
    onConfirm: (instanceIDs: string[]) => void;
    onCancel: () => void;
  }

  const { card, offer, options, onConfirm, onCancel }: Props = $props();

  const need = $derived(altCostPayCount(offer ?? undefined));

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

  // At a count of one the list stays a radio group — picking a
  // second card replaces the first rather than erroring, which is
  // what every single-card cost did before S29.
  function toggle(instanceID: string): void {
    if (need === 1) {
      chosen = chosen[0] === instanceID ? [] : [instanceID];
      return;
    }
    if (chosen.includes(instanceID)) {
      chosen = chosen.filter((id) => id !== instanceID);
      return;
    }
    if (chosen.length >= need) return;
    chosen = [...chosen, instanceID];
  }

  function confirm(): void {
    if (chosen.length !== need) return;
    onConfirm(chosen);
  }
</script>

{#if card && offer}
  <DockSheet
    label={card.name}
    src="alternative cost · CR 118.9"
    width={560}
    sheetKey={`altpay:${card.instance_id}:${offer.key}`}
    primary={confirmAction("Pay", confirm, { disabled: chosen.length !== need })}
    secondary={[cancelAction(onCancel)]}
  >
    <p class="prompt-hint">
      {offer.label ?? offer.key}. Choose {offer.pay_label ?? "a card"}.
      {#if need > 1}
        <span class="tally">{chosen.length} of {need}</span>
      {/if}
      {#if offer.life}
        You also pay {offer.life} life.
      {/if}
    </p>
    {#if options.length < need}
      <p class="prompt-hint error">You have nothing that can pay this cost.</p>
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
              {#if c.mana_cost}
                <span class="note cost"><ManaCost cost={c.mana_cost} size={14} /></span>
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
  .cost {
    font-family: var(--font-mono);
  }
  .tally {
    font-family: var(--font-mono);
    opacity: 0.8;
  }
</style>
