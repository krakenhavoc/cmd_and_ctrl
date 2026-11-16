<script lang="ts">
  // DraftReasons — the box under a card's large preview that says why it
  // is rated what it is (#5848 §3): "Rating 58.1 (B+)", a few short lines,
  // and the numbers behind them with their source. The words are built by
  // lib/draftReasons.ts from the card's own data. Nothing is drawn for a
  // card with no rating, or when the viewer has hidden ratings.
  import type { DraftCard } from "../../draft";
  import { ratingFooter, ratingHeading, ratingReasons } from "../../draftReasons";
  import { ratingsPref } from "../../draftRatings.svelte";

  interface Props {
    card: DraftCard;
    /** The set's average win rate when drawn, in percent (the view's `set_avg_win_rate`). */
    setAvg?: number;
  }
  const { card, setAvg }: Props = $props();

  const lines = $derived(ratingReasons(card, setAvg));
  const footer = $derived(ratingFooter(card));
</script>

{#if ratingsPref.show && card.rating}
  <div class="why" data-testid="draft-reasons">
    <h3 class="head">{ratingHeading(card)}</h3>
    {#if lines.length > 0}
      <ul>
        {#each lines as line (line)}
          <li>{line}</li>
        {/each}
      </ul>
    {/if}
    <p class="foot">
      {#each footer as part (part)}<span>{part}</span>{" "}{/each}
    </p>
  </div>
{/if}

<style>
  .why {
    padding: 10px 12px;
    border-radius: var(--radius);
    background: var(--surface-raised);
    border: 1px solid var(--border-strong);
    font-size: 12.5px;
    line-height: 1.45;
    color: var(--fg-muted);
  }
  .head {
    margin: 0 0 4px;
    font-family: var(--font-display);
    font-size: 14px;
    font-weight: 800;
    letter-spacing: 0;
    text-transform: none;
    color: var(--fg);
  }
  ul {
    margin: 0;
    padding-left: 16px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .foot {
    margin: 6px 0 0;
    display: flex;
    flex-wrap: wrap;
    gap: 2px 12px;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--fg-muted);
  }
</style>
