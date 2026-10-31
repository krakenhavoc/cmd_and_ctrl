<script lang="ts">
  // IdentityPips — a deck's colour identity as small lettered discs, in
  // WUBRG order (ADR 0212 §1: a seat tile's status row; the picker grid's
  // tiles next). Filled from the fixed identity tokens (app.css, ADR 0158
  // §2), which no skin overrides, with the letter on each so the colour is
  // never the only signal. One image with one name for a screen reader.
  import { orderedColors } from "../prebuiltDecks";

  const { colors }: { colors: readonly string[] | undefined } = $props();

  const NAMES: Record<string, string> = {
    W: "white",
    U: "blue",
    B: "black",
    R: "red",
    G: "green",
  };

  const letters = $derived(orderedColors(colors ? [...colors] : undefined));
</script>

{#if letters.length > 0}
  <span
    class="pips"
    role="img"
    aria-label={`colour identity: ${letters.map((l) => NAMES[l]).join(", ")}`}
  >
    {#each letters as l (l)}
      <i class="pip" style:--pip={`var(--identity-${l.toLowerCase()})`} aria-hidden="true">{l}</i>
    {/each}
  </span>
{/if}

<style>
  .pips {
    display: inline-flex;
    gap: 3px;
    flex: none;
  }
  .pip {
    display: grid;
    place-items: center;
    width: 15px;
    height: 15px;
    border-radius: 50%;
    background: var(--pip);
    /* Dark ink on the light identity fills, in every skin. */
    color: #0b0f14;
    font: 700 8.5px/1 var(--font-mono);
    font-style: normal;
  }
</style>
