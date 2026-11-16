// draftReasons.ts — why a card is rated what it is (#5848 §3): the box
// under the large preview. A heading ("Rating 58.1 (B+)"), up to four
// short lines, and a footer of the numbers and where they came from.
//
// The lines are built from the numbers and the card's own data, never
// written per card:
//   - its win rate when drawn (GIH WR) against the set's average,
//   - what it does (removal, a flier, card draw, a counterspell, or a
//     plain creature), from draftDeckStats' roles, and
//   - a card of two or more colours.
//
// One line is left out on purpose: how early the card goes (ALSA, from
// `avg_seen`). It waits on #5726, which decides whether the number can be
// trusted; neither the lines nor the footer mention it until then.

import { isLand, type DraftCard } from "./draft";
import { cardRoles } from "./draftDeckStats";
import { isLandsRating, ratingTierText } from "./draftRatings";
import { cardTypes } from "./draftStats";
import { L } from "./labels";

/** The most lines the box shows. */
export const MAX_REASONS = 4;

/** "+3.2" or "-1.4": the sign is always said. A tie is "+0.0". */
function signed(n: number): string {
  const s = n.toFixed(1);
  return n < 0 && Number(s) !== 0 ? s : `+${Math.abs(Number(s)).toFixed(1)}`;
}

/** "Rating 58.1 (B+)" for a 17Lands card, "Rating 3.4" for the in-house kind. */
export function ratingHeading(card: DraftCard): string {
  const r = card.rating;
  if (!r) return "";
  return L.draftRatingHeading(r.value.toFixed(1), ratingTierText(r));
}

/**
 * The lines under the heading, in order: the win rate against the set's,
 * what the card does, then its colours. `setAvg` is the draft view's
 * `set_avg_win_rate`, in percent; without it the win rate stands alone.
 * A card with no 17Lands number says so instead of a win rate.
 */
export function ratingReasons(card: DraftCard, setAvg?: number): string[] {
  const out: string[] = [];
  const r = card.rating;
  if (r && isLandsRating(r)) {
    out.push(
      setAvg !== undefined && Number.isFinite(setAvg) && setAvg > 0
        ? L.draftReasonWinRate(r.value.toFixed(1), signed(r.value - setAvg))
        : L.draftReasonWinRateOnly(r.value.toFixed(1)),
    );
  } else if (r) {
    out.push(L.draftReasonInHouse);
  }

  if (!isLand(card)) {
    const roles = cardRoles(card);
    const keywords = card.keywords ?? [];
    if (roles.includes("removal")) out.push(L.draftReasonRemoval);
    else if (roles.includes("fliers")) out.push(L.draftReasonFlier);
    else if (roles.includes("counters")) out.push(L.draftReasonCounter);
    else if (roles.includes("draw")) out.push(L.draftReasonDraw);
    else if (
      cardTypes(card).includes("creature") &&
      keywords.length === 0 &&
      !(card.oracle_text ?? "").trim()
    ) {
      out.push(L.draftReasonPlain);
    }
  }

  const colours = card.colors?.length ?? 0;
  if (colours === 2) out.push(L.draftReasonTwoColours);
  else if (colours > 2) out.push(L.draftReasonManyColours);
  return out.slice(0, MAX_REASONS);
}

/** The footer's parts: the win rate when there is one, the mana value and the source. */
export function ratingFooter(card: DraftCard): string[] {
  const r = card.rating;
  const parts: string[] = [];
  if (r && isLandsRating(r)) parts.push(L.draftFooterWinRate(r.value.toFixed(1)));
  parts.push(L.draftFooterManaValue(card.mana_value ?? 0));
  if (r) parts.push(L.draftFooterSource(isLandsRating(r) ? (r.fetched ?? "") : null));
  return parts;
}
