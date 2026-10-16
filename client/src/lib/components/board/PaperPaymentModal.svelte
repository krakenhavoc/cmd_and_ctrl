<script lang="ts">
  // PaperPaymentModal — "What did you pay?" (#3855, ADR 0068 amendment
  // 2026-10-15).
  //
  // A payment the engine waived (strict mana off with too little in the
  // pool, or Cast anyway) that something on the table reads — a converge
  // or sunburst spell, Vexing Bauble, Coin of Mastery — comes back from
  // the server as `paper_payment_required`, with what the pool left
  // unpaid one slot at a time. Nothing has been paid. This asks what the
  // player paid outside the pool, prefilled with the cost minus what the
  // pool paid, or lets them say an effect on paper let them cast it
  // without paying its mana cost. The answer re-sends the same action.
  //
  // Each slot offers only the symbols that may pay it (a {R} slot only
  // R, a {W/U} slot W or U, a generic slot any), so whatever the player
  // picks is a legal declaration; the server checks it anyway. When a
  // reader counts mana from artifact sources, each slot also asks
  // whether an artifact made it.
  //
  // A sheet in the action dock (ADR 0111 PR 6): the confirm, the free
  // answer and Cancel are the dock's action bar.

  import type { PaperPaymentPrompt } from "../../protocol";
  import { cancelAction, confirmAction, type DockAction } from "../../dock";
  import { L } from "../../labels";
  import {
    FREE_PAYMENT,
    declaredPayment,
    paperPaymentSymbols,
    paperPaymentValid,
    prefillPaperPayment,
    slotCount,
    type PaperPayment,
    type PaperSlotChoice,
  } from "../../paperPayment";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // The card cast, or the permanent whose ability is activated. Null
    // closes the sheet.
    cardName: string | null;
    prompt: PaperPaymentPrompt | null;
    // "Cast" for a cast, "Activate" for an ability. Only a cast offers
    // the free answer.
    verb?: string;
    onConfirm: (payment: PaperPayment) => void;
    onCancel: () => void;
  }

  const { cardName, prompt, verb = "Cast", onConfirm, onCancel }: Props = $props();

  let choices = $state<PaperSlotChoice[]>([]);
  // A new prompt is a new question: prefill it again.
  let lastPrompt: PaperPaymentPrompt | null = null;
  $effect(() => {
    if (prompt !== lastPrompt) {
      lastPrompt = prompt;
      choices = prompt ? prefillPaperPayment(prompt) : [];
    }
  });

  const valid = $derived(prompt !== null && paperPaymentValid(prompt, choices));
  const symbols = $derived(paperPaymentSymbols(choices));

  function pick(i: number, color: string): void {
    choices[i] = { ...choices[i]!, color };
  }
  function toggleArtifact(i: number): void {
    choices[i] = { ...choices[i]!, artifact: !choices[i]!.artifact };
  }
  function confirm(): void {
    if (!prompt || !valid) return;
    onConfirm(declaredPayment(prompt, choices));
  }

  const secondary = $derived.by((): DockAction[] => {
    const out: DockAction[] = [];
    if (verb === "Cast") {
      out.push({
        id: "free",
        label: L.castFree,
        title:
          "an effect let you cast it without paying its mana cost; anything that reads whether mana was spent sees that",
        onPress: () => onConfirm(FREE_PAYMENT),
      });
    }
    out.push(cancelAction(onCancel));
    return out;
  });
</script>

{#if cardName && prompt}
  <DockSheet
    label={L.paperPayment(cardName)}
    title={L.paperPayment(cardName)}
    width={480}
    sheetKey={`paper:${cardName}:${prompt.slots.map((s) => s.symbol).join("")}`}
    primary={confirmAction(`${verb}, ${symbols} paid on paper`, confirm, { disabled: !valid })}
    {secondary}
  >
    <p class="prompt-hint">
      Your mana pool couldn't pay all of it, and something on the table reads what was paid. Say
      what you paid outside the pool. The game log shows the table.
    </p>
    <ol class="slots">
      {#each prompt.slots as slot, i (i)}
        <li class="slot">
          <span class="symbol"
            >{slot.symbol}{#if slotCount(slot) > 1}<span class="count">×{slotCount(slot)}</span
              >{/if}</span
          >
          {#if slot.x}<span class="xtag">{L.paperSlotOnX}</span>{/if}
          <span class="options" role="group" aria-label={`paid for ${slot.symbol}`}>
            {#each slot.options as opt (opt)}
              <button
                type="button"
                class="ghost opt"
                aria-pressed={choices[i]?.color === opt}
                onclick={() => pick(i, opt)}>{`{${opt}}`}</button
              >
            {/each}
          </span>
          {#if prompt.artifact_sources}
            <label class="artifact">
              <input
                type="checkbox"
                checked={choices[i]?.artifact ?? false}
                onchange={() => toggleArtifact(i)}
              />
              from an artifact
            </label>
          {/if}
        </li>
      {/each}
    </ol>
  </DockSheet>
{/if}

<style>
  .slots {
    list-style: none;
    margin: 8px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .slot {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }
  .symbol {
    font-family: var(--font-mono);
    min-width: 3.5rem;
    color: var(--fg);
  }
  .count {
    color: var(--fg-muted);
  }
  /* A slot of the cost's {X} part: the mana a spend-only-on-X clause
     restricts, shown apart so the player can see which mana pays X
     (#3941). */
  .xtag {
    font-size: 12px;
    color: var(--fg-muted);
    border: 1px solid var(--border, currentColor);
    border-radius: 4px;
    padding: 0 4px;
  }
  .options {
    display: flex;
    gap: 4px;
  }
  .opt {
    font-family: var(--font-mono);
    padding: 2px 6px;
  }
  .opt[aria-pressed="true"] {
    outline: 2px solid var(--accent, var(--mint));
  }
  .artifact {
    font-size: 12px;
    color: var(--fg-muted);
    display: flex;
    align-items: center;
    gap: 4px;
  }
</style>
