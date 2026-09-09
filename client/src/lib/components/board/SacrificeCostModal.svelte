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
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="sac-title">
    <div class="prompt-modal sac-modal">
      <h2 id="sac-title">
        {source.name}
        <span class="prompt-src" aria-hidden="true">additional cost</span>
      </h2>
      <p class="prompt-hint">Sacrifice {label} to pay for this ability.</p>
      {#if options.length === 0}
        <p class="prompt-hint error">Nothing you control can pay this cost.</p>
      {:else}
        <ul class="prompt-options">
          {#each options as c (c.instance_id)}
            <li>
              <button
                type="button"
                class="prompt-opt"
                class:on={chosen === c.instance_id}
                aria-pressed={chosen === c.instance_id}
                onclick={() => (chosen = c.instance_id)}
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
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" disabled={!chosen} onclick={confirm}>Sacrifice</button
        >
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
