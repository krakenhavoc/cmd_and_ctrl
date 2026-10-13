/**
 * artistCredit.svelte.ts — "Illus. <artist>" for a card shown as an art crop.
 *
 * Scryfall's art_crop image carries neither the artist line nor the
 * copyright line, and their image guidelines ask that neither be
 * cropped or covered out of view (#2982). Where the art crop is the
 * main view of a card we therefore print the artist ourselves. The name
 * comes from GET /cards/{id} (the metadata the hover panel already
 * fetches and caches), not from the per-frame card view: it is the same
 * for every copy of a printing and would cost every snapshot a string
 * per card.
 *
 * Full card images (the hover zoom, the catalogue grid, the mulligan
 * grid) print the artist line themselves and are never given an overlay
 * here.
 */

import { metaFor, type CardMeta } from "./cardMetaCache";

/** artistOf reads the illustrator of the printed face the viewer sees. */
export function artistOf(meta: CardMeta | null | undefined, face = 0): string | null {
  if (!meta) return null;
  const name = meta.card_faces?.[face]?.artist || meta.artist;
  return name ? name : null;
}

/** creditText is the one string every art-crop surface prints. */
export function creditText(artist: string): string {
  return `Illus. ${artist}`;
}

/**
 * useArtist returns a reactive `.value` holding the artist of the card
 * `source()` names, or null until the metadata lands (and for ever when
 * it never does: the credit is a nicety, not a gate). Call it during
 * component initialisation.
 *
 * The subscription lives in an effect rather than a $derived because
 * `metaFor` may start a fetch (see the #740 note in cardMetaCache.ts).
 */
export function useArtist(source: () => { id: string | null | undefined; face?: number }) {
  let artist = $state<string | null>(null);
  $effect(() => {
    const { id, face } = source();
    if (!id) {
      artist = null;
      return;
    }
    const unsub = metaFor(id).subscribe((meta) => {
      artist = artistOf(meta, face ?? 0);
    });
    return unsub;
  });
  return {
    get value() {
      return artist;
    },
  };
}
