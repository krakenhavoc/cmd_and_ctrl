<script lang="ts">
  // CostConfirmModal — ADR 0109 §7 (#1902) and owner decision 3: the
  // confirm for an activation cost the player chooses nothing for —
  // "Discard a card at random" (Pyromancy) and "Exile the top N cards of
  // your library" (Seasoned Tactician, Arc-Slogger). There is nothing to
  // pick, so it is not a picker: it says what the payment will do, and
  // says why it can't be paid when the hand or the library is short
  // (CR 118.3), with the confirm disabled.
  //
  // A sheet in the action dock (ADR 0111 PR 6), like every cost picker:
  // the confirm and Cancel are the dock's action bar, Enter / Escape
  // through its one key handler.
  import type { CardView } from "../../protocol";
  import type { CostConfirmLine } from "../../libraryCost";
  import { cancelAction, confirmAction } from "../../dock";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // The permanent whose ability is being activated; null closes it.
    card: CardView | null;
    lines: CostConfirmLine[];
    note?: string;
    onConfirm: () => void;
    onCancel: () => void;
  }

  const {
    card,
    lines,
    note = "activation cost · CR 602.2b",
    onConfirm,
    onCancel,
  }: Props = $props();

  const payable = $derived(lines.length > 0 && lines.every((l) => !l.short));

  function confirm(): void {
    if (!payable) return;
    onConfirm();
  }
</script>

{#if card}
  <DockSheet
    label={card.name}
    src={note}
    width={480}
    sheetKey={`costconfirm:${card.instance_id}`}
    primary={confirmAction("Pay and activate", confirm, { disabled: !payable })}
    secondary={[cancelAction(onCancel)]}
  >
    <p class="prompt-hint">To activate this ability you will:</p>
    <ul class="cost-lines">
      {#each lines as line (line.text)}
        <li class:short={!!line.short}>
          <span class="text">{line.text}</span>
          {#if line.short}
            <span class="prompt-hint error">{line.short}</span>
          {/if}
        </li>
      {/each}
    </ul>
  </DockSheet>
{/if}

<style>
  .cost-lines {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }
  .cost-lines li {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    padding: 0.5rem 0.65rem;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    background: var(--surface-sunken);
  }
  .cost-lines li.short {
    border-color: var(--danger);
  }
  .text {
    font-weight: 600;
  }
</style>
