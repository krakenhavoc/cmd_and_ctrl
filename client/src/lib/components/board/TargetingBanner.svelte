<script lang="ts">
  // TargetingBanner is the "click a target for X" row of the
  // board's attention strip, shown while a cast-with-targets is in
  // flight. Mounted once by Game.svelte (inside Board's `attention`
  // snippet); visible only while the targeting store is non-null.

  import {
    targeting,
    cancel,
    confirm,
    legalTargetCount,
    isMultiPick,
    canConfirm,
  } from "../../targeting";
  import Icon from "../Icon.svelte";
  import { doubledTriggerLabel } from "../../triggerDoubling";

  const state = $derived($targeting);
  const count = $derived(state ? legalTargetCount(state) : -1);
  // S20 sub-PR 5: multi-target clauses show the pick tally and a
  // Done button instead of completing on the first click.
  const multi = $derived(state !== null && isMultiPick(state));
  const confirmable = $derived(state !== null && canConfirm(state));
  const tally = $derived.by(() => {
    if (!state || !multi) return "";
    const n = state.picked.length;
    if (state.max > 0) return `${n}/${state.max} picked`;
    return `${n} picked`;
  });
  const doubledLabel = $derived(doubledTriggerLabel(state?.doubledBy, state?.doubledByName));

  function modeHint(mode: string | undefined): string {
    switch (mode) {
      case "any":
        return "a player or creature";
      case "player":
        return "a player";
      case "creature":
        return "a creature";
      case "permanent":
        return "a permanent";
      case "stack_spell":
        return "a spell on the stack";
      case "card_in_graveyard":
        return "a card in a graveyard";
      default:
        return "a target";
    }
  }
</script>

{#if state}
  <div
    class="banner"
    role="dialog"
    aria-live="polite"
    aria-label={`Select target for ${state.card.name}`}
  >
    <span class="label"><Icon name="sword" size={12} /> target</span>
    <span class="prompt">
      {#if state.choiceID}
        <strong>{state.card.name}</strong> triggered — click {state.label || "a target"}
      {:else}
        Click {modeHint(state.mode)} to target
        <strong>{state.card.name}</strong>
      {/if}
      {#if count >= 0}
        <span class="count">· {count} legal</span>
      {/if}
      {#if doubledLabel}
        <span class="count">· {doubledLabel}</span>
      {/if}
      {#if multi}
        <span class="count">· {tally}</span>
      {/if}
    </span>
    {#if multi}
      <button
        type="button"
        class="primary done"
        disabled={!confirmable}
        onclick={confirm}
        title={state.min > 0 && state.picked.length < state.min
          ? `pick at least ${state.min}`
          : "confirm targets (Enter)"}
      >
        Done
      </button>
    {/if}
    {#if !state.choiceID}
      <button type="button" class="ghost cancel" onclick={cancel} title="cancel (Esc)">
        Cancel <kbd>Esc</kbd>
      </button>
    {/if}
  </div>
{/if}

<style>
  .banner {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 10px 9px 14px;
    background: color-mix(in srgb, var(--surface) 94%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid rgba(217, 180, 92, 0.45);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.35;
    box-sizing: border-box;
    animation: banner-in 200ms var(--ease);
  }
  @keyframes banner-in {
    from {
      opacity: 0;
      transform: translateY(-6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    font-weight: 700;
    color: var(--gold-strong);
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex: 0 0 auto;
  }
  .prompt {
    flex: 1;
    min-width: 0;
  }
  .prompt strong {
    color: var(--fg);
    font-weight: 700;
  }
  .count {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--fg-dim);
    margin-left: 4px;
    white-space: nowrap;
  }
  .done,
  .cancel {
    flex: 0 0 auto;
    height: 28px;
    padding: 0 10px;
    font-size: 11.5px;
    border-radius: 7px;
  }
  .cancel {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .cancel kbd {
    font-family: var(--font-mono);
    font-size: 9.5px;
    color: var(--fg-dim);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 0 4px;
    line-height: 16px;
  }
  .done:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
