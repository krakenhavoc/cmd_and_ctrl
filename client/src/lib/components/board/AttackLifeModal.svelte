<script lang="ts">
  // AttackLifeModal — #4198: the question a Phyrexian attack tax asks
  // before a declaration goes out. Norn's Annex prints {W/P} for each
  // attacking creature; each is paid with {W} or 2 life (CR 107.4f).
  //
  // The attackers picker has the stepper inline, with its own checked
  // set. This sheet is for the declarations that have no picker: the
  // dock's "Attack with N" for the staged set, the chosen-attackers
  // confirm, a re-pointed attacker. A cast with a printed Phyrexian
  // symbol always asks (shouldAskPhyrexianLife), so this does too, and
  // the answer rides the one declare_attackers as `phyrexian_life`.
  //
  // ADR 0111 PR 6: a sheet in the action dock; the confirm and Cancel
  // are the dock's action bar.

  import { cancelAction, confirmAction } from "../../dock";
  import { clampAttackLife, type AttackLifePrompt } from "../../attackLife";
  import { phyrexianLifeCost } from "../../phyrexianLife";
  import DockSheet from "./DockSheet.svelte";
  import AttackLifeStepper from "./AttackLifeStepper.svelte";

  interface Props {
    // Null closes the sheet.
    prompt: AttackLifePrompt | null;
    // The payer's life total, read live so the cap follows it.
    life: number;
    onConfirm: (n: number) => void;
    onCancel: () => void;
  }

  const { prompt, life, onConfirm, onCancel }: Props = $props();

  let n = $state(0);
  let lastKey: string | null = null;
  // A new question starts at 0: paying the colour is what most
  // declarations mean to do.
  $effect(() => {
    const key = prompt ? `${prompt.verb}:${prompt.symbols}` : null;
    if (key !== lastKey) {
      lastKey = key;
      n = 0;
    }
  });

  const live = $derived(prompt ? clampAttackLife(n, prompt.symbols, life) : 0);
  const lifeCost = $derived(phyrexianLifeCost(live));
</script>

{#if prompt}
  <DockSheet
    label="Pay the attack tax"
    src={prompt.each}
    width={440}
    sheetKey={`attack-life:${prompt.verb}:${prompt.symbols}`}
    primary={confirmAction(`${prompt.verb}${lifeCost > 0 ? ` for ${lifeCost} life` : ""}`, () =>
      onConfirm(live),
    )}
    secondary={[cancelAction(onCancel)]}
  >
    <p class="prompt-hint">
      {prompt.attackers === 1
        ? "Attacking costs one Phyrexian symbol."
        : `Attacking with ${prompt.attackers} creatures costs ${prompt.symbols} Phyrexian symbols.`}
      Pay each with its colour of mana, or with 2 life.
    </p>
    <AttackLifeStepper
      symbols={prompt.symbols}
      {life}
      each={prompt.each}
      value={n}
      onChange={(v) => (n = clampAttackLife(v, prompt.symbols, life))}
    />
  </DockSheet>
{/if}
