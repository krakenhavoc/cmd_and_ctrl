<script lang="ts">
  // AttackLifeStepper — #4198: how many of the attack tax's Phyrexian
  // symbols are paid with 2 life each (CR 107.4f), the rest with mana.
  //
  // PhyrexianCostModal's stepper, shaped for a declaration: the symbols
  // are the attackers' tax (`attack_targets[].tax_phyrexian` times the
  // attackers sent), and the ceiling is the symbols, or what the life
  // total can buy (CR 119.4), whichever is smaller. Both sheets that
  // ask the question (the attackers picker and AttackLifeModal) draw
  // this one control so they cannot drift apart.
  //
  // It owns no state: the sheet holds the count, so it can send it.

  import { PhyrexianLifePerSymbol, phyrexianLifeCost } from "../../phyrexianLife";
  import { attackLifeCeiling, attackLifeSentence } from "../../attackLife";
  import { L } from "../../labels";

  interface Props {
    // Phyrexian symbols in the whole declaration's tax.
    symbols: number;
    // The payer's life total, for the CR 119.4 cap.
    life: number;
    // The tax one creature pays, shown beside the count ("{W/P}").
    each?: string;
    value: number;
    onChange: (n: number) => void;
  }

  const { symbols, life, each = "", value, onChange }: Props = $props();

  const max = $derived(attackLifeCeiling(symbols, life));
  const n = $derived(Math.max(0, Math.min(value, max)));
  const lifeCost = $derived(phyrexianLifeCost(n));
</script>

<div class="pay-row" role="group" aria-label={L.attackPayWithLife}>
  <span class="pay-label">Pay with life</span>
  <button
    type="button"
    class="ghost step"
    disabled={n <= 0}
    onclick={() => onChange(n - 1)}
    aria-label="one fewer symbol">−</button
  >
  <output class="count" aria-live="polite">{n} of {symbols}</output>
  <button
    type="button"
    class="ghost step"
    disabled={n >= max}
    onclick={() => onChange(n + 1)}
    aria-label="one more symbol">+</button
  >
  <span class="price" class:spending={lifeCost > 0}>
    {lifeCost === 0 ? "no life" : `${lifeCost} life`}
  </span>
</div>
<p class="status">
  {#if each}{each} per creature, each symbol its mana or {PhyrexianLifePerSymbol} life:
  {/if}{attackLifeSentence(n, symbols)}.
  {#if max < symbols}You have {life} life, so you can buy at most {max}.{/if}
</p>

<style>
  .pay-row {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
    flex-wrap: wrap;
  }
  .pay-label {
    font-weight: 600;
    color: var(--fg);
  }
  .step {
    min-width: 2rem;
    font-family: var(--font-mono);
    font-size: 16px;
    line-height: 1;
    padding: 4px 8px;
  }
  .count {
    font-family: var(--font-mono);
    font-size: 15px;
    min-width: 5.5rem;
    text-align: center;
  }
  .price {
    flex: 1;
    text-align: right;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .price.spending {
    color: var(--danger);
  }
  .status {
    font-size: 12px;
    color: var(--fg-muted);
    line-height: 1.35;
    margin: 8px 0 0;
  }
</style>
