<script lang="ts">
  // SacrificeCostModal — S21 sub-PR 2: pick the permanent that pays
  // an activated ability's sacrifice cost ("Sacrifice a creature:").
  //
  // A cost is not a target: it can't be responded to, and hexproof
  // doesn't apply. So this is a plain list of your own permanents
  // rather than the board-click targeting flow — and it opens
  // BEFORE any target prompt, matching the order costs are paid in
  // (CR 601.2h).

  import { onDestroy } from "svelte";
  import type { CardView } from "../../protocol";

  interface Props {
    // The ability's source, for the heading.
    source: CardView | null;
    // The clause ("a creature") and the permanents that can pay it.
    label: string;
    options: CardView[];
    onConfirm: (instanceID: string) => void;
    onCancel: () => void;
  }

  const { source, label, options, onConfirm, onCancel }: Props = $props();

  let chosen = $state<string | null>(null);

  // Reset when a different activation opens the prompt.
  let lastSourceID: string | null = null;
  $effect(() => {
    const id = source?.instance_id ?? null;
    if (id !== lastSourceID) {
      lastSourceID = id;
      chosen = null;
    }
  });

  function confirm(): void {
    if (!chosen) return;
    onConfirm(chosen);
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
  <div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="sac-title">
    <div class="modal">
      <h2 id="sac-title">{source.name}</h2>
      <p class="prompt">Sacrifice {label}</p>
      {#if options.length === 0}
        <p class="empty">Nothing you control can pay this cost.</p>
      {:else}
        <ul class="options">
          {#each options as c (c.instance_id)}
            <li>
              <button
                type="button"
                class="option"
                class:selected={chosen === c.instance_id}
                aria-pressed={chosen === c.instance_id}
                onclick={() => (chosen = c.instance_id)}
              >
                <span class="mark" aria-hidden="true">{chosen === c.instance_id ? "●" : "○"}</span>
                <span class="name">{c.name}</span>
                {#if c.power !== undefined && c.toughness !== undefined}
                  <span class="pt">{c.power}/{c.toughness}</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
      <div class="actions">
        <button type="button" class="cancel" onclick={onCancel}>Cancel</button>
        <button type="button" class="confirm" disabled={!chosen} onclick={confirm}>Sacrifice</button
        >
      </div>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.65);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1100;
  }
  .modal {
    background: #1c1f2a;
    color: #e6e8ee;
    border: 1px solid #3a4055;
    border-radius: 6px;
    padding: 1.25rem 1.5rem;
    min-width: 320px;
    max-width: 460px;
    box-shadow: 0 12px 28px rgba(0, 0, 0, 0.4);
  }
  h2 {
    margin: 0 0 0.25rem 0;
    font-size: 1.05rem;
    color: #f4ead5;
  }
  .prompt {
    margin: 0 0 0.75rem 0;
    color: #aab2c8;
    font-size: 0.9rem;
  }
  .empty {
    margin: 0 0 0.9rem 0;
    color: #ff9a9a;
    font-size: 0.9rem;
  }
  .options {
    list-style: none;
    margin: 0 0 1rem 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    max-height: 40vh;
    overflow-y: auto;
  }
  .option {
    width: 100%;
    display: flex;
    align-items: baseline;
    gap: 0.6rem;
    text-align: left;
    padding: 0.45rem 0.65rem;
    border-radius: 4px;
    border: 1px solid #3a4055;
    background: #262a38;
    color: #e6e8ee;
    cursor: pointer;
    font-size: 0.92rem;
  }
  .option.selected {
    border-color: #ff9a9a;
    background: #3a2630;
  }
  .mark {
    flex: 0 0 auto;
    color: #ff9a9a;
  }
  .name {
    flex: 1 1 auto;
  }
  .pt {
    flex: 0 0 auto;
    color: #aab2c8;
    font-size: 0.85rem;
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }
  .actions button {
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    border: 1px solid #3a4055;
    background: #262a38;
    color: #e6e8ee;
    cursor: pointer;
  }
  .actions .confirm {
    background: #6b2f3a;
    border-color: #8c4351;
  }
  .actions .confirm:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
