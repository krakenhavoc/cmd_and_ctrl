/**
 * rulings.svelte.ts — the live "Rulings" link for a card's menu (#3831).
 *
 * `useRulings` follows the card's printing in the metadata cache, so the
 * link is Scryfall's the moment a menu opens and becomes Gatherer's if
 * the printing has a multiverse id once GET /cards/{id} lands (it is
 * normally cached already: the hover that precedes a right-click fetches
 * it). The subscription lives in an effect for the reason useArtist's
 * does: metaFor may start a fetch (see the #740 note in cardMetaCache.ts).
 */

import { metaFor, type CardMeta } from "./cardMetaCache";
import { offersRulings, rulingsURL } from "./rulings";
import type { CardView } from "./protocol";

export function useRulings(source: () => CardView | null | undefined) {
  let meta = $state<CardMeta | null>(null);
  $effect(() => {
    const card = source();
    if (!card?.scryfall_id || !offersRulings(card)) {
      meta = null;
      return;
    }
    return metaFor(card.scryfall_id).subscribe((m) => {
      meta = m;
    });
  });
  return {
    /** The page to open, or null when this card offers no rulings. */
    get url() {
      return rulingsURL(source(), meta);
    },
  };
}
