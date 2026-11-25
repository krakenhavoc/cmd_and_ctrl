<script lang="ts">
  // PayToCancel — "… unless they pay {N} before that step" on your own
  // panel (#6220; Nafs Asp, Sabertooth Cobra). One button per payment
  // waiting on you, live while you have priority and the mana; pressing it
  // sends the server's own move (payToCancel.ts). Renders nothing when no
  // payment waits, which is almost always.
  import type { ActionPayload, ActionType, GameView, PayToCancelView } from "../../protocol";
  import { payToCancelButtons } from "../../payToCancel";
  import { L } from "../../labels";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seatID: string;
    rows: PayToCancelView[] | undefined;
    view: GameView;
    sendAction: ActionSender;
  }

  const { seatID, rows, view, sendAction }: Props = $props();

  const buttons = $derived(payToCancelButtons(view, seatID, rows));
</script>

{#if buttons.length > 0}
  <div class="pay-to-cancel" role="group" aria-label="Payments you may make">
    {#each buttons as b (b.id)}
      <button
        type="button"
        class="pay"
        title={b.title}
        aria-label={L.payToCancel(b.what)}
        disabled={!b.enabled}
        onclick={() => b.params && sendAction("special_action", b.params, seatID)}
      >
        {b.text}
      </button>
    {/each}
  </div>
{/if}

<style>
  .pay-to-cancel {
    flex: none;
    display: flex;
    flex-direction: column;
    gap: 3px;
    align-self: flex-end;
    padding: 0 2px 2px;
  }
  .pay {
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
  .pay:disabled {
    opacity: 0.5;
    cursor: default;
  }
  .pay:hover:not(:disabled),
  .pay:focus-visible {
    box-shadow: 0 0 0 1px var(--accent-strong);
  }
</style>
