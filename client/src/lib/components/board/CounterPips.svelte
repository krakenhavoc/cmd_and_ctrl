<script lang="ts">
  // CounterPips renders a stack of pip chips at the top-right of a
  // battlefield card showing each named counter and its count. Pip
  // colour comes from the shared counter-type registry (see
  // client/src/lib/counterTypes.ts) so +1/+1 is green, -1/-1 is
  // red, loyalty is purple, etc. Unknown counter names render with
  // a neutral colour and the first 3 characters of the name as the
  // chip text.

  import { counterStyle } from "../../counterTypes";

  interface Props {
    counters?: Record<string, number>;
  }

  const { counters }: Props = $props();

  // Stable iteration order — sorts by name so a given creature's
  // pips render the same way each snapshot regardless of map
  // iteration order. Drops zero / negative entries so a sparse map
  // stays sparse on the visible UI.
  const entries = $derived(
    counters
      ? Object.entries(counters)
          .filter(([, count]) => count > 0)
          .sort(([a], [b]) => a.localeCompare(b))
      : [],
  );
</script>

{#if entries.length > 0}
  <div class="pips" aria-hidden="true">
    {#each entries as [name, count] (name)}
      {@const style = counterStyle(name)}
      <span class="pip" style:background={style.color} title={`${count}× ${name}`}>
        {style.abbr}{count > 1 ? "·" + count : ""}
      </span>
    {/each}
  </div>
{/if}

<style>
  .pips {
    position: absolute;
    top: 4px;
    right: 4px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    pointer-events: none;
    z-index: 4;
  }
  .pip {
    display: inline-block;
    padding: 2px 7px;
    border-radius: 999px;
    color: #0c1426;
    font-size: 10px;
    font-weight: 800;
    text-shadow: 0 1px 0 rgba(255, 255, 255, 0.5);
    box-shadow:
      0 2px 4px rgba(0, 0, 0, 0.55),
      inset 0 1px 0 rgba(255, 255, 255, 0.35),
      inset 0 -1px 0 rgba(0, 0, 0, 0.2);
    line-height: 1.2;
    letter-spacing: 0.01em;
  }
</style>
