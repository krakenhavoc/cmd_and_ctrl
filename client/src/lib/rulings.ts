/**
 * rulings.ts — where "Rulings" on a card's right-click menu goes (#3831).
 *
 * Wizards' Gatherer is the official source of a card's rulings, and it
 * is linked whenever the printing has a multiverse id. Universes Beyond
 * cards and newer printings often have none, and then the Scryfall card
 * page is linked at its rulings section (Scryfall mirrors Wizards'
 * rulings). Nothing is fetched from either site: the link is built from
 * what GET /cards/{id} already says about the printing (cardMetaCache).
 *
 * Gatherer's current address for a card is /<SET>/en-us/<number>/<name>,
 * which needs the name's slug, so a set and collector number are not
 * enough to build it. The old Details.aspx?multiverseid= address still
 * redirects to it, for both faces of a double-faced card, so that is the
 * one used.
 */

import type { CardView } from "./protocol";
import type { CardMeta } from "./cardMetaCache";

type RulingsCard = Pick<CardView, "scryfall_id" | "is_token" | "face_down" | "active_face">;

/**
 * offersRulings says whether a card's menu may carry "Rulings". Not for
 * a card the viewer cannot see (a redacted card arrives with no
 * scryfall_id, and a link would name it), not for a face-down object
 * (a morph, a manifested or cloaked card, a foretold card: even to the
 * one who may look, the link would put its name in the address), and
 * not for a token, whose printing is only an art choice (ADR 0078).
 */
export function offersRulings(card: RulingsCard | null | undefined): boolean {
  return !!card && !!card.scryfall_id && !card.is_token && !card.face_down;
}

/**
 * rulingsURL is the page that lists the card's rulings, or null when the
 * card does not offer them. `meta` is the printing's GET /cards/{id}
 * answer, or null while it is not here yet (or never comes): the link
 * then names the Scryfall id, which Scryfall redirects to the card.
 *
 * `face` picks a double-faced card's Gatherer entry (each side has its
 * own multiverse id); it defaults to the face that is up.
 */
export function rulingsURL(
  card: RulingsCard | null | undefined,
  meta: Pick<CardMeta, "set" | "collector_number" | "multiverse_ids"> | null | undefined,
  face?: number,
): string | null {
  if (!card || !offersRulings(card)) return null;
  const ids = (meta?.multiverse_ids ?? []).filter((n) => Number.isInteger(n) && n > 0);
  if (ids.length > 0) {
    const id = ids[face ?? card.active_face ?? 0] ?? ids[0];
    return `https://gatherer.wizards.com/Pages/Card/Details.aspx?multiverseid=${id}`;
  }
  const rest =
    meta?.set && meta.collector_number
      ? `${encodeURIComponent(meta.set)}/${encodeURIComponent(meta.collector_number)}`
      : encodeURIComponent(card.scryfall_id as string);
  return `https://scryfall.com/card/${rest}#rulings`;
}
