<script lang="ts">
  // DraftRatingBadge — the rating beside one draft card (ADR 0222 §7,
  // #5703): a small chip with the number, and the source named in words
  // so it is never colour alone.
  //
  // 17Lands: its letter grade and the games-in-hand win rate. The
  // in-house rating: "C2" and its points, on a plainer chip so the two
  // do not read as one scale. The title is the explanation; the card's
  // own control carries the same in its accessible name (ratingAria), so
  // the chip is decoration for a screen reader.
  //
  // It draws nothing when the card has no rating or the viewer hid them
  // (the switch is DraftRatingsBar's). `overlay` pins it to the foot of
  // the nearest positioned parent, a card image; without it the chip
  // sits in the flow, for a list row.
  import {
    isLandsRating,
    ratingTierText,
    ratingTooltip,
    ratingValueText,
    type DraftRating,
  } from "../../draftRatings";
  import { ratingsPref } from "../../draftRatings.svelte";

  interface Props {
    rating?: DraftRating | null;
    /** The set the card is drafted from, so a number from another set says so. */
    cardSet?: string;
    overlay?: boolean;
  }
  const { rating, cardSet, overlay = false }: Props = $props();

  const lands = $derived(isLandsRating(rating));
  const tier = $derived(rating ? ratingTierText(rating) : "");
</script>

{#if rating && ratingsPref.show}
  <span
    class="rating"
    class:overlay
    class:lands
    data-testid="rating-badge"
    data-source={lands ? "17lands" : "c2"}
    title={ratingTooltip(rating, cardSet)}
    aria-hidden="true"
  >
    {#if tier}<b class="tier">{tier}</b>{/if}
    <span class="value">{ratingValueText(rating)}</span>
    {#if !lands}<span class="src">C2</span>{/if}
  </span>
{/if}

<style>
  .rating {
    flex: none;
    display: inline-flex;
    align-items: baseline;
    gap: 5px;
    padding: 2px 7px;
    border-radius: 999px;
    background: var(--surface-raised);
    border: 1px solid var(--border-strong);
    color: var(--fg);
    font-family: var(--font-mono);
    font-size: 11px;
    line-height: 1.4;
    font-variant-numeric: tabular-nums;
    box-shadow: var(--shadow-sm);
    white-space: nowrap;
  }
  .rating.overlay {
    position: absolute;
    left: 8px;
    bottom: 8px;
    z-index: 2;
  }
  /* 17Lands' numbers take the accent; ours stay neutral and carry the
     word, so the two scales are told apart without colour. */
  .rating.lands {
    border-color: var(--accent-strong);
  }
  .tier {
    font-weight: 800;
    color: var(--fg);
  }
  .value {
    color: var(--fg);
  }
  .src {
    color: var(--fg-muted);
    font-weight: 700;
    letter-spacing: 0.04em;
  }
</style>
