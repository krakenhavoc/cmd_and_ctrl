<script lang="ts">
  // DraftStatsStrip — Colours and Curve of Main, always in view and
  // compact (#5848 §6). It sits under the picks on the pick screen and
  // runs across the top of Deck view and the deckbuilding screen
  // (`wide`: the two side by side). Main only; the rest is the full
  // stats overlay's, when it is asked for.
  import type { DraftCard } from "../../draft";
  import { L } from "../../labels";
  import DraftColours from "./DraftColours.svelte";
  import DraftCurve from "./DraftCurve.svelte";

  interface Props {
    /** The cards in Main. */
    main: DraftCard[];
    /** Across the top of a page: the colours and the curve side by side. */
    wide?: boolean;
  }
  const { main, wide = false }: Props = $props();
</script>

<section class="card strip" class:wide aria-label={L.draftStatsStripName} data-testid="draft-stats">
  <div>
    <h3>{L.draftStatColours}</h3>
    <DraftColours {main} compact />
  </div>
  <div>
    <h3>{L.draftStatCurve}</h3>
    <DraftCurve {main} />
  </div>
</section>

<style>
  .strip {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .strip.wide {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 14px 28px;
  }
  .strip > div {
    min-width: 0;
  }
  h3 {
    margin: 0 0 6px;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
</style>
