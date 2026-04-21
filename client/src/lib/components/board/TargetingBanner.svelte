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
    padding: 8px 14px;
    border-radius: 8px;
    background: rgba(80, 60, 0, 0.92);
    color: #ffd07a;
    border: 1px solid #c8a86a;
    font-size: 13px;
    z-index: 60;
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.5);
  }
  .prompt strong {
    color: #ffe69a;
    font-weight: 700;
  }
  .cancel {
    padding: 4px 10px;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    background: #2a2010;
    color: #ffd07a;
    border: 1px solid #c8a86a;
    border-radius: 4px;
    cursor: pointer;
  }
  .cancel:hover {
    background: #3a2e14;
  }
</style>
