<script lang="ts">
  // TargetingBanner is the floating "click a target for X" prompt
  // that appears while a cast-with-targets is in flight. Mounted
  // once by Game.svelte; visible only while the targeting store is
  // non-null.

  import { targeting, cancel } from "../../targeting";

  const state = $derived($targeting);

  function modeHint(mode: string | undefined): string {
    switch (mode) {
      case "any":
        return "a player or creature";
      case "player":
        return "a player";
      case "creature":
        return "a creature";
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
    <span class="prompt">
      Click {modeHint(state.mode)} to target
      <strong>{state.card.name}</strong>
    </span>
    <button type="button" class="cancel" onclick={cancel} title="cancel (Esc)"> Cancel </button>
  </div>
{/if}

<style>
  .banner {
    position: absolute;
    top: 12px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 18px;
    border-radius: 999px;
    background: linear-gradient(180deg, rgba(80, 60, 0, 0.94) 0%, rgba(50, 38, 0, 0.94) 100%);
    color: var(--gold);
    border: 1px solid rgba(200, 168, 106, 0.7);
    font-size: 13px;
    z-index: 60;
    box-shadow:
      0 10px 30px rgba(0, 0, 0, 0.55),
      0 0 24px rgba(255, 208, 122, 0.18),
      inset 0 1px 0 rgba(255, 255, 255, 0.08);
    backdrop-filter: blur(6px);
    animation: banner-in 200ms var(--ease);
  }
  @keyframes banner-in {
    from {
      opacity: 0;
      transform: translate(-50%, -8px);
    }
    to {
      opacity: 1;
      transform: translate(-50%, 0);
    }
  }
  .prompt strong {
    color: #ffe69a;
    font-weight: 700;
  }
  .cancel {
    padding: 4px 12px;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 700;
    background: rgba(42, 32, 16, 0.9);
    color: var(--gold);
    border: 1px solid rgba(200, 168, 106, 0.6);
    border-radius: 999px;
    cursor: pointer;
    box-shadow: none;
    transition:
      background 120ms var(--ease),
      border-color 120ms var(--ease);
  }
  .cancel:hover {
    background: rgba(58, 46, 20, 0.95);
    border-color: var(--gold);
  }
</style>
