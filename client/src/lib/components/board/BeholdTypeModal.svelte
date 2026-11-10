<script lang="ts">
  // BeholdTypeModal — #5406: choose the creature type a behold that
  // prints none is paid with. Celestial Reunion's "you may choose a
  // creature type and behold two creatures of that type" (CR 601.2b).
  //
  // It opens after the optional cost is ticked and before the behold
  // pick, only when the server lists two or more `behold_types`; with one
  // the type is taken without asking. Each row is a type with the number
  // of your creatures (in hand and on the battlefield) that could pay it.
  // The server lists them alphabetically; the row with the most creatures
  // starts selected.
  //
  // A sheet in the action dock, like the permission type picker: the
  // confirm and Cancel are the dock's action bar.
  import { onDestroy } from "svelte";
  import type { BeholdTypeOption, CardView } from "../../protocol";
  import { cancelAction, confirmAction } from "../../dock";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // The spell being cast; null closes the modal.
    card: CardView | null;
    types: BeholdTypeOption[];
    count: number;
    onConfirm: (type: string) => void;
    onCancel: () => void;
  }

  const { card, types, count, onConfirm, onCancel }: Props = $props();

  // The row with the most creatures; the first on a tie.
  function suggested(list: BeholdTypeOption[]): number {
    let best = 0;
    list.forEach((t, i) => {
      if (t.cards.length > (list[best]?.cards.length ?? 0)) best = i;
    });
    return best;
  }

  let chosen = $state(0);
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = suggested(types);
    }
  });

  function confirm(): void {
    const t = types[chosen];
    if (t) onConfirm(t.type);
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
    src="choose a creature type"
    width={420}
    sheetKey={`behold-type:${card.instance_id}`}
    primary={confirmAction(`Behold ${types[chosen]?.type ?? ""}`, confirm)}
    secondary={[cancelAction(onCancel)]}
  >
    <p class="prompt-hint">
      Choose a creature type. You then behold {count} creatures of that type: creature cards in your hand
      or creatures you control.
    </p>
    <ul class="type-options">
      {#each types as t, i (t.type)}
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
            <span class="type-name">{t.type}</span>
            <span class="type-note">{t.cards.length} to choose from</span>
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
    max-height: 240px;
    overflow-y: auto;
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
