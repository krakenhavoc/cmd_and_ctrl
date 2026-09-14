// prebuiltDecks.ts — the pre-built deck picker's types, fetch, and the
// sentence it puts under each deck's name.
//
// Mirrors server/internal/lobby/prebuilt.go and the decks.Coverage it
// embeds. Hand-maintained, like protocol.ts and catalog.ts — update
// both sides in lockstep.
//
// The summarising lives here rather than in the component for the same
// reason catalog.ts's filter does: this project has no jsdom, so a
// `.svelte` file cannot be unit-tested, and "how honest is this
// sentence" is exactly the kind of decision that deserves a test. The
// sentence is the feature. A picker that says "all cards implemented"
// over a deck with twenty-two simplifications would be worse than no
// picker at all, because the player would find out one surprise at a
// time instead of once, up front.

/** One card the engine does not carry out in full. */
export interface ImperfectCard {
  name: string;
  /** Player-facing descriptions of the clauses the engine skips. */
  caveats?: string[];
  /**
   * The card is registered but nobody has graded it: we do not know
   * of a gap, and we have not checked. Distinct from a caveat card,
   * and said differently.
   */
  unreviewed?: boolean;
}

/**
 * A deck's engine-coverage profile. Counts are per distinct card, not
 * per copy — which is why they do not add up to `card_count`: a
 * deck's ten basic lands are one or two rows in `basics`.
 */
export interface DeckCoverage {
  /** Non-basic cards graded = full + caveats + unreviewed. */
  cards: number;
  full: number;
  caveats: number;
  unreviewed: number;
  /** Basic-land rows, which are exempt — the engine always plays them. */
  basics: number;
  /** Cards with no implementation at all. Zero; a build test enforces it. */
  unregistered: number;
  imperfect?: ImperfectCard[];
}

/** One pre-built deck, as GET /decks reports it. */
export interface PrebuiltDeck {
  id: string;
  name: string;
  archetype?: string;
  summary?: string;
  commander?: string;
  /** WUBRG letters. */
  colors?: string[];
  /** Physical size, counting every copy of a basic land. 100. */
  card_count: number;
  coverage: DeckCoverage;
}

export interface PrebuiltDecksResponse {
  decks: PrebuiltDeck[];
}

/**
 * How well this build implements a deck.
 *
 * - `full` — every card plays exactly as printed.
 * - `caveats` — every card is implemented; some are simplified, or
 *   have not been reviewed.
 * - `gap` — at least one card has no implementation at all. Should be
 *   unreachable (server-side build test), and is rendered rather than
 *   assumed away, because a wrong claim here is the one failure this
 *   feature cannot afford.
 */
export type CoverageTone = "full" | "caveats" | "gap";

export interface CoverageSummary {
  tone: CoverageTone;
  /** The one line under the deck name. */
  headline: string;
  /** The breakdown, or "" when there is nothing more to say. */
  detail: string;
}

/**
 * summariseCoverage turns the counts into the two lines the picker
 * shows.
 *
 * The rule it enforces: the phrase "exactly as printed" is only ever
 * applied to the `full` count, and "every card is implemented" is only
 * said when `unregistered` is zero. Everything else is a number the
 * player can weigh.
 */
export function summariseCoverage(cov: DeckCoverage): CoverageSummary {
  const total = cov.cards;
  if (cov.unregistered > 0) {
    return {
      tone: "gap",
      headline: `${cov.unregistered} of ${total} cards have no implementation in this build`,
      detail: "Those cards behave as manual sandbox cards — you move the pieces yourself.",
    };
  }
  if (cov.caveats === 0 && cov.unreviewed === 0) {
    return {
      tone: "full",
      headline: `All ${total} nonbasic cards play exactly as printed`,
      detail: basicsNote(cov),
    };
  }
  const parts: string[] = [];
  if (cov.caveats > 0) parts.push(`${cov.caveats} simplified`);
  if (cov.unreviewed > 0) parts.push(`${cov.unreviewed} not yet reviewed`);
  const basics = basicsNote(cov);
  return {
    tone: "caveats",
    headline: `Every card is implemented · ${cov.full} of ${total} nonbasic cards as printed`,
    detail: [parts.join(", "), basics].filter(Boolean).join(". "),
  };
}

function basicsNote(cov: DeckCoverage): string {
  if (cov.basics <= 0) return "";
  return "Basic lands are always played in full";
}

/**
 * imperfectCardLine renders one entry of the disclosure list. An
 * unreviewed card has nothing to disclose but its own status, and
 * saying so beats an empty bullet that reads like a missing string.
 */
export function imperfectCardLine(card: ImperfectCard): string {
  const text = (card.caveats ?? []).filter((c) => c.trim() !== "");
  if (text.length > 0) return text.join(" ");
  if (card.unreviewed) return "Not yet reviewed — it may or may not play exactly as printed.";
  return "";
}

/** Colour pips, deduped and in WUBRG order, for the picker rows. */
export const WUBRG = ["W", "U", "B", "R", "G"] as const;

export function orderedColors(colors: string[] | undefined): string[] {
  if (!colors || colors.length === 0) return [];
  const present = new Set(colors.map((c) => c.toUpperCase()));
  return WUBRG.filter((c) => present.has(c));
}

/**
 * deckSubtitle is the "archetype — commander" line. Both halves are
 * optional on the wire; the join has to survive either being absent
 * rather than rendering a stray dash.
 */
export function deckSubtitle(deck: PrebuiltDeck): string {
  return [deck.archetype, deck.commander].filter((s) => (s ?? "").trim() !== "").join(" · ");
}
