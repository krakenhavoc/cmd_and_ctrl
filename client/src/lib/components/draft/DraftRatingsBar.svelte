<script lang="ts">
  // DraftRatingsBar — the credit for 17Lands' numbers, what the numbers
  // are, and the viewer's switch to hide them (ADR 0222 §7, #5703).
  //
  // The credit is the owner's condition for using 17Lands' data
  // (2026-11-11): it shows wherever one of its numbers shows, and links
  // to 17lands.com. It goes with the cards, so a screen that shows
  // ratings mounts this bar with the same cards. The bar draws nothing
  // when none of them has a rating.
  //
  // The explanation is on the page, not only in a tooltip, because a
  // touch screen has no hover.
  import { anyLandsRating, anyRating, RATINGS_CREDIT_URL } from "../../draftRatings";
  import { ratingsPref } from "../../draftRatings.svelte";
  import type { DraftCard } from "../../draft";
  import { L } from "../../labels";

  interface Props {
    /** The cards the screen shows ratings beside. */
    cards: readonly DraftCard[];
  }
  const { cards }: Props = $props();

  const rated = $derived(anyRating(cards));
  const lands = $derived(anyLandsRating(cards));
</script>

{#if rated}
  <div class="ratings-bar" data-testid="ratings-bar">
    <label class="switch">
      <input
        type="checkbox"
        checked={ratingsPref.show}
        onchange={(e) => ratingsPref.set(e.currentTarget.checked)}
      />
      <span>{L.draftRatingsShow}</span>
    </label>
    {#if ratingsPref.show}
      <p class="what">
        {#if lands}
          <a href={RATINGS_CREDIT_URL} target="_blank" rel="noopener noreferrer"
            >{L.draftRatingsCredit}</a
          >: the win rate of games with the card in hand, and its grade among the set's rated cards.
          A badge marked C2 is this site's own score, for a card 17Lands has no win rate for.
        {:else}
          A badge marked C2 is this site's own score from the card's text, for a card that has no
          win rate from outside data.
        {/if}
      </p>
    {/if}
  </div>
{/if}

<style>
  .ratings-bar {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 6px 16px;
    margin: 0 0 12px;
    color: var(--fg-muted);
    font-size: 0.8125rem;
    line-height: 1.5;
  }
  .switch {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    color: var(--fg);
    font-weight: 600;
    cursor: pointer;
  }
  .switch input {
    margin: 0;
    accent-color: var(--accent-strong);
  }
  .switch input:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  .what {
    flex: 1 1 28ch;
    min-width: 0;
    margin: 0;
  }
  .what a {
    color: var(--fg);
    font-weight: 600;
  }
</style>
