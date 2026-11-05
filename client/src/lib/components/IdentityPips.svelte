<script lang="ts">
  // IdentityPips — a deck's colour identity as its mana symbols, {W} {U}
  // {B} {R} {G} in WUBRG order, the way a card prints them (#4847; ADR
  // 0212 §1: a seat tile's status row, and the picker grid's tiles). The
  // symbols are ManaSymbol's, the same ones a cost and the mana pool draw,
  // so the shape (sun, drop, skull, flame, tree) carries the colour as well
  // as the disc does. A deck known to be colourless shows {C}; one whose
  // identity is unknown shows nothing. One image with one name for a
  // screen reader.
  import { orderedColors } from "../prebuiltDecks";
  import ManaSymbol from "./board/ManaSymbol.svelte";

  interface Props {
    /** WUBRG letters, any order or case. */
    colors: readonly string[] | undefined;
    /** The commanders are known and have no colour: draw {C} for no colours. */
    colorless?: boolean;
    /** Each symbol's diameter in CSS px. */
    size?: number;
  }

  const { colors, colorless = false, size = 16 }: Props = $props();

  const NAMES: Record<string, string> = {
    W: "white",
    U: "blue",
    B: "black",
    R: "red",
    G: "green",
    C: "colourless",
  };

  const ordered = $derived(orderedColors(colors ? [...colors] : undefined));
  const symbols = $derived(ordered.length > 0 ? ordered : colorless ? ["C"] : []);
</script>

{#if symbols.length > 0}
  <span
    class="pips"
    role="img"
    aria-label={`colour identity: ${symbols.map((s) => NAMES[s]).join(", ")}`}
  >
    {#each symbols as s (s)}
      <ManaSymbol symbol={s} {size} />
    {/each}
  </span>
{/if}

<style>
  .pips {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    flex: none;
    /* Room for each symbol's offset shadow disc, which overflows. */
    padding-bottom: 1px;
  }
</style>
