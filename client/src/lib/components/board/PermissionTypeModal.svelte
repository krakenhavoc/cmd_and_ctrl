<script lang="ts">
  // PermissionTypeModal — #2167: choose which card type a cast spends
  // under a permission that opens one of each type. Muldrotha, the
  // Gravetide's reminder text: "If a card has multiple permanent types,
  // choose one as you play it." Aminatou's Augury asks the same.
  //
  // It opens after the face picker (the type is a question about the
  // face being cast) and before the costs, only when the face carries
  // two or more `permission_types`. The server ranks them, so the first
  // row — the type the permission's other cards need least — starts
  // selected.
  //
  // A sheet in the action dock, like the face picker: the confirm and
  // Cancel are the dock's action bar.
  import { onDestroy } from "svelte";
  import type { CardView } from "../../protocol";
  import { cancelAction, confirmAction } from "../../dock";
  import {
    permissionTypeLabel,
    permissionTypeVerb,
    permissionTypesOf,
  } from "../../permissionTypes";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // The face being cast; null closes the modal.
    card: CardView | null;
    onConfirm: (permissionType: string) => void;
    onCancel: () => void;
  }

  const { card, onConfirm, onCancel }: Props = $props();

  const types = $derived(card ? permissionTypesOf(card) : []);

  let chosen = $state(0);
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = 0;
    }
  });

  function confirm(): void {
    const t = types[chosen];
    if (t) onConfirm(t);
  }

  // The arrows and 1-9 pick a row while the sheet is open; Enter and
  // Escape are the dock's.
  function handleKey(e: KeyboardEvent): void {
    if (!card) return;
    if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
      e.preventDefault();
      chosen = Math.max(0, chosen - 1);
    } else if (e.key === "ArrowRight" || e.key === "ArrowDown") {
      e.preventDefault();
      chosen = Math.min(types.length - 1, chosen + 1);
    } else if (e.key >= "1" && e.key <= "9") {
      const n = Number(e.key) - 1;
      if (n < types.length) {
        e.preventDefault();
        chosen = n;
      }
    }
  }
  $effect(() => {
    if (!card) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });
  onDestroy(() => document.removeEventListener("keydown", handleKey));
</script>

{#if card && types.length > 1}
  <DockSheet
    label={card.name}
    src="one of each card type"
    width={420}
    sheetKey={`permission-type:${card.instance_id}`}
    primary={confirmAction(permissionTypeVerb(card, types[chosen] ?? ""), confirm)}
    secondary={[cancelAction(onCancel)]}
  >
    <p class="prompt-hint">
      This card has more than one type. Which one does it use? You can cast one spell of each type.
    </p>
    <ul class="type-options">
      {#each types as t, i (t)}
        <li>
          <button
            type="button"
            class="type-opt"
            class:on={chosen === i}
            aria-pressed={chosen === i}
            onclick={() => (chosen = i)}
            ondblclick={() => {
              chosen = i;
              confirm();
            }}
          >
            <span class="type-name">{permissionTypeLabel(t)}</span>
            {#if i === 0}<span class="type-note">suggested</span>{/if}
          </button>
        </li>
      {/each}
    </ul>
  </DockSheet>
{/if}

<style>
  .prompt-hint {
    margin: 0 0 10px;
    font-size: 13px;
    opacity: 0.85;
  }

  .type-options {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .type-opt {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
    min-width: 120px;
    padding: 8px 12px;
    border-radius: 10px;
    border: 2px solid transparent;
    background: var(--surface-sunken);
    color: inherit;
    cursor: pointer;
    text-align: left;
  }

  .type-opt:hover {
    border-color: var(--border-strong, rgba(255, 255, 255, 0.25));
  }

  .type-opt.on {
    border-color: var(--accent, #d9a441);
  }

  .type-name {
    font-weight: 600;
  }

  .type-note {
    font-size: 11px;
    opacity: 0.7;
  }
</style>
