<script lang="ts">
  // IgnoreEffects — the CR 116.2d special action on your own panel
  // (#1986): "ignore this effect until end of turn" for the price of
  // sacrificing a permanent of YOUR choice. Damping Engine is the card.
  //
  // The server decides everything: `seat.ignorable_effects` lists, for the
  // viewer's own seat only, each static ability that bears on it right now
  // and the permanents it may sacrifice (payment order). A button here opens
  // the one sacrifice picker every sacrifice cost uses, over those
  // permanents, and a confirm sends
  //   special_action { card_id: <the Engine>, kind: "ignore_effect",
  //                    sacrifice_id: <the pick> }
  // The action uses no stack and takes no mana, so there is no auto-tap, no
  // timing and no strict flag to send; the server refuses a pick that is
  // not on its list, so the picker cannot offer a wrong one.
  //
  // Renders nothing when no effect bears on the seat, which is almost
  // always.
  import type {
    ActionPayload,
    ActionType,
    CardView,
    GameView,
    IgnorableEffectView,
  } from "../../protocol";
  import { orderSacrificeOptions } from "../../sacrificeCost";
  import { L } from "../../labels";
  import SacrificeCostModal from "./SacrificeCostModal.svelte";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seatID: string;
    effects: IgnorableEffectView[] | undefined;
    view: GameView;
    sendAction: ActionSender;
  }

  const { seatID, effects, view, sendAction }: Props = $props();

  let picking = $state<IgnorableEffectView | null>(null);

  const source: CardView | null = $derived(
    picking
      ? (view.battlefield.cards.find((c) => c.instance_id === picking!.source_id) ?? null)
      : null,
  );
  const options = $derived(
    picking ? orderSacrificeOptions(view.battlefield.cards, picking.sacrifice_ids) : [],
  );

  function confirm(ids: string[]): void {
    const row = picking;
    picking = null;
    if (!row || ids.length !== 1) return;
    sendAction(
      "special_action",
      { card_id: row.source_id, kind: "ignore_effect", sacrifice_id: ids[0] },
      seatID,
    );
  }
</script>

{#if effects && effects.length > 0}
  <div class="ignore-effects" role="group" aria-label="Effects you may ignore">
    {#each effects as row (row.source_id)}
      <button
        type="button"
        class="ignore"
        title={row.label}
        aria-label={L.ignoreEffect(row.source_name)}
        onclick={() => (picking = row)}
      >
        Ignore {row.source_name}
      </button>
    {/each}
  </div>
{/if}

<SacrificeCostModal
  {source}
  label="a permanent"
  {options}
  count={1}
  onConfirm={confirm}
  onCancel={() => (picking = null)}
/>

<style>
  .ignore-effects {
    flex: none;
    display: flex;
    flex-direction: column;
    gap: 3px;
    align-self: flex-end;
    padding: 0 2px 2px;
  }
  .ignore {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 2px 8px;
    border-radius: 999px;
    border: 1px solid var(--accent-strong);
    background: rgba(8, 7, 6, 0.85);
    color: var(--accent-strong);
    cursor: pointer;
    white-space: nowrap;
  }
  .ignore:hover,
  .ignore:focus-visible {
    box-shadow: 0 0 0 1px var(--accent-strong);
  }
</style>
