<script lang="ts">
  // ManaCost — a brace-notation cost drawn as mana pips (#2231).
  //
  // "{U}{U}{R}" is blue, blue, red; every generic number sums into one
  // grey circle ("{1}{1}{U}" is a 2 and a blue pip), {C} keeps its own
  // diamond pip, and X, hybrid, Phyrexian, snow and {0} are drawn too.
  // The pips are decorative: the wrapper carries the readable name
  // (pass `label` to say more than the cost, e.g. "costs … to cast").
  import { costPips, costWords } from "../../manaSymbol";
  import ManaSymbol from "./ManaSymbol.svelte";

  interface Props {
    /** The cost in brace notation: "{2}{U}". */
    cost: string;
    /** Pip diameter in CSS px. */
    size?: number;
    /** Accessible name. Defaults to "mana cost <words>". */
    label?: string;
  }

  const { cost, size = 14, label }: Props = $props();

  const pips = $derived(costPips(cost));
  const name = $derived(label ?? `mana cost ${costWords(cost)}`.trim());
</script>

{#if pips.length > 0}
  <span class="mana-cost" role="img" aria-label={name} data-cost={cost}>
    {#each pips as p, i (i)}
      <ManaSymbol symbol={p.symbol} {size} />
    {/each}
  </span>
{:else if cost}
  <!-- Not brace notation (a plain word): show it as it came. -->
  <span class="mana-cost-text">{cost}</span>
{/if}

<style>
  .mana-cost {
    display: inline-flex;
    align-items: center;
    gap: 1px;
    flex: 0 0 auto;
    vertical-align: middle;
  }
</style>
