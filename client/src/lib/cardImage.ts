/**
 * cardImage.ts — the single place the client builds a card-image URL.
 *
 * Before ADR 0034 this string was assembled inline in five components
 * (Card.svelte, HoverZoomOverlay, StackOverlay, PlayerIdentity, and
 * Game.svelte's mulligan grid), each with its own template literal.
 * That was survivable while the URL had one parameter. Multi-face
 * cards add a second — `?face=N` picks which printed side to serve —
 * and five hand-rolled URL builders is five places to forget it.
 *
 * The service worker caches on the full query string
 * (service-worker.js CARD_IMAGE_PATH), so a face-qualified URL is
 * cached independently and correctly with no change there.
 */

import type { CardView } from "./protocol";
import { LAYOUT_FLIP, LAYOUT_SPLIT } from "./faces";

/** Scryfall image sizes the server's `?size=` accepts. */
export type CardImageSize = "small" | "normal" | "large" | "png" | "art_crop" | "border_crop";

/**
 * cardImageURL returns the path to a card's art, or null when the
 * card has no Scryfall ID to serve one from (fixtures, the demo
 * seed, and every card redacted by the per-viewer filter — a card
 * the viewer is not a knower of arrives with scryfall_id cleared,
 * which is exactly when the caller should be rendering a card back).
 *
 * The face defaults to the card's ACTIVE face rather than to zero.
 * That is what makes a transformed permanent, or an MDFC played as
 * its land half, show the right side everywhere without any caller
 * opting in: `active_face` is already on the wire and already says
 * which side is up. Pass an explicit face only to show a side that
 * is NOT currently up — the picker showing both halves, or the hover
 * overlay's back-face panel.
 *
 * A FLIP card (ADR 0151) always asks for face 0, whichever half is up
 * or asked for: it has one printed image, and its flipped half is that
 * image turned upside down (Card.svelte and the hover panel turn it).
 * One URL per printing keeps the browser and service-worker caches to
 * one copy of it.
 *
 * A SPLIT card (Fire // Ice, aftermath, every Room) is the same: Scryfall
 * prints ONE image of the whole card, at the top level, and its
 * card_faces carry none, so `?face=1` is a 404 (#3518). Every half
 * therefore asks for face 0, the whole card; the picker turns it
 * (see isSidewaysArt).
 */
export function cardImageURL(
  card:
    | (Pick<CardView, "scryfall_id" | "active_face"> & Partial<Pick<CardView, "layout">>)
    | null
    | undefined,
  size: CardImageSize = "normal",
  face?: number,
): string | null {
  if (!card?.scryfall_id) return null;
  if (card.layout === LAYOUT_FLIP || card.layout === LAYOUT_SPLIT) {
    return scryfallImageURL(card.scryfall_id, size, 0);
  }
  return scryfallImageURL(card.scryfall_id, size, face ?? card.active_face ?? 0);
}

/**
 * scryfallImageURL is cardImageURL for callers holding a bare
 * Scryfall ID rather than a CardView — the seat header's commander
 * art crop, which is handed an ID and nothing else.
 *
 * Face 0 deliberately emits NO `?face=` parameter. Every image
 * already cached by a browser, a service worker or the server's
 * on-disk store was fetched under the unqualified URL; adding a
 * redundant `face=0` would miss all of them and re-download the
 * entire cache.
 */
export function scryfallImageURL(
  scryfallID: string,
  size: CardImageSize = "normal",
  face = 0,
): string {
  const base = `/cards/${scryfallID}/image?size=${size}`;
  return face > 0 ? `${base}&face=${face}` : base;
}

/**
 * tableImageSize picks the size a board or hand card requests (#1954).
 * `artOnly` is the experimental "card art only" setting; it applies to
 * a face-up card and nothing else, so a face-down card, a card back and
 * every caller that does not opt in keep `size`. The server answers an
 * art_crop request for a card with no crop URI with the "normal" image
 * (cards.ImageURIForFace), so there is no client-side 404 to handle:
 * the full card appears where the art does not exist.
 */
export function tableImageSize(
  card: Pick<CardView, "face_down"> | null | undefined,
  size: CardImageSize,
  artOnly: boolean,
): CardImageSize {
  if (!artOnly || card?.face_down) return size;
  return "art_crop";
}
