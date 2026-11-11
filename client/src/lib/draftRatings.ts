// draftRatings.ts — the rating shown beside a draft card, and the viewer's
// choice to hide it (ADR 0222 §7, amendment 2026-11-11, #5703).
//
// Mirrors the `rating` object of server/internal/draft/ratings.go
// (CardRating). Hand-maintained: update both sides in lockstep.
//
// Two sources, and they are not on one scale:
//   - "17lands": the card's games-in-hand win rate from 17Lands' public
//     data, in percent, with a letter tier read off the set's own spread.
//     Credit is a condition of using it (the owner's decision of
//     2026-11-11): every screen that shows one carries the credit line.
//   - "c2": this server's own rating of the card from its printed text,
//     in points on the draft bots' scale. It is not a win rate.
//
// This file is pure. The viewer's choice lives in draftRatings.svelte.ts.

export const RATINGS_CREDIT_URL = "https://www.17lands.com/";

export type RatingSource = "17lands" | "c2";

export interface DraftRating {
  /** "17lands" or "c2"; any other value is shown as the in-house kind. */
  source: string;
  /** 17Lands: GIH win rate in percent, one decimal. c2: points. */
  value: number;
  /** 17Lands only: a letter grade from the set's own spread ("B+"). */
  tier?: string;
  /** 17Lands only: the games behind the win rate. */
  games?: number;
  /** The pick number the card is taken at on average, and last seen at. */
  avg_pick?: number;
  avg_seen?: number;
  /** The set whose 17Lands numbers these are, lowercase. */
  set?: string;
  /** The 17Lands format ("PremierDraft") and the date fetched. */
  format?: string;
  fetched?: string;
}

/** True for a 17Lands rating. */
export function isLandsRating(r: DraftRating | undefined | null): boolean {
  return !!r && r.source === "17lands";
}

/** True when any of these cards carries a 17Lands rating: the credit is owed. */
export function anyLandsRating(cards: readonly { rating?: DraftRating | null }[]): boolean {
  return cards.some((c) => isLandsRating(c.rating));
}

/** True when any of these cards carries a rating at all. */
export function anyRating(cards: readonly { rating?: DraftRating | null }[]): boolean {
  return cards.some((c) => !!c.rating);
}

/** "58.3%" for 17Lands, "3.4" for the in-house rating. */
export function ratingValueText(r: DraftRating): string {
  return isLandsRating(r) ? `${r.value.toFixed(1)}%` : r.value.toFixed(1);
}

/** The letter tier of a 17Lands rating, "" for any other. */
export function ratingTierText(r: DraftRating): string {
  return isLandsRating(r) ? (r.tier ?? "") : "";
}

/** The format as 17Lands names it, with spaces: "PremierDraft" → "Premier Draft". */
function formatName(format: string): string {
  return format.replace(/([a-z])([A-Z])/g, "$1 $2");
}

const FORMAT_FIELDS = new Intl.NumberFormat("en-US");

/**
 * The tooltip: what the number is and where it comes from, in words.
 * cardSet is the set the card is drafted from; a 17Lands number from
 * another set (a cube card, or a reprint) says so.
 */
export function ratingTooltip(r: DraftRating, cardSet?: string): string {
  if (!isLandsRating(r)) {
    return (
      `C2 rating ${r.value.toFixed(1)}: this site's own score on the draft bots' scale ` +
      "(a two-mana 2/2 is about 2.7, a bomb is above 5). It is worked out from the card's " +
      "printed text, not from games played. 17Lands has no win rate for this card."
    );
  }
  const parts = [
    `17Lands: ${r.value.toFixed(1)}% win rate in games where this card was in hand (the opening hand or drawn).`,
  ];
  if (r.tier) {
    parts.push(`Grade ${r.tier}: where it ranks among the rated cards of its set.`);
  }
  const from =
    r.set && cardSet && r.set.toLowerCase() !== cardSet.toLowerCase() ? r.set.toUpperCase() : "";
  const sample = [
    r.games ? `${FORMAT_FIELDS.format(r.games)} games` : "",
    r.format ? formatName(r.format) : "",
    from ? `from ${from}` : "",
    r.fetched ? `fetched ${r.fetched}` : "",
  ].filter(Boolean);
  if (sample.length > 0) parts.push(`${sample.join(", ")}.`);
  if (r.avg_pick) parts.push(`Taken at pick ${r.avg_pick.toFixed(1)} on average.`);
  return parts.join(" ");
}

/** The words a screen reader adds to a card's name: short, with the source. */
export function ratingAria(r: DraftRating): string {
  if (!isLandsRating(r)) return `C2 rating ${r.value.toFixed(1)}`;
  return r.tier
    ? `17Lands win rate ${r.value.toFixed(1)} percent, grade ${r.tier}`
    : `17Lands win rate ${r.value.toFixed(1)} percent`;
}

// ---- The viewer's choice ----
//
// Hidden or shown, kept in localStorage: a per-viewer convenience, not
// state anyone else needs. Every read and write is guarded, because the
// storage can be missing or refuse (a private window, blocked site data),
// and the page must read the same without it: shown.

export const RATINGS_PREF_KEY = "c2.draft.ratings";

/** Whether the viewer wants ratings shown. Shown unless they said no. */
export function loadShowRatings(): boolean {
  try {
    return localStorage.getItem(RATINGS_PREF_KEY) !== "hide";
  } catch {
    return true;
  }
}

/** Remembers the choice; a failure is silent, the choice lasts the page. */
export function saveShowRatings(show: boolean): void {
  try {
    if (show) localStorage.removeItem(RATINGS_PREF_KEY);
    else localStorage.setItem(RATINGS_PREF_KEY, "hide");
  } catch {
    // storage unavailable: the choice holds until the page is reloaded
  }
}
