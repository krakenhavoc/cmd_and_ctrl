// tombstoneAnchor — where a tombstone sits on its seat's panel (#4552).
//
// One rule for every seat: the stone is centred across the panel, and
// vertically centred in the part of the panel the seat's own identity (its
// name, avatar, life, counters and badges) leaves free. That part is the
// battlefield, which is empty for a player who has left the game.
//
// The identity is not at the same edge on every panel: a top-row opponent's
// is against the top edge (the hand is up there), a seat in the bottom row
// has it against the bottom edge, and the viewer's stands above their hand.
// A stone centred on the whole panel is clear of it on a tall panel and
// lands on it on a short one, which is what #4552 saw. So the stone reads
// where the identity is and keeps to the side of it that has the room.
//
// Tombstone.svelte measures the boxes and hands them here; this is the
// decision, pinned by tombstoneAnchor.test.ts.

/** A vertical span, in any one coordinate space (the viewport's, say). */
export interface Span {
  top: number;
  bottom: number;
}

/** The clear space between the identity and the stone. */
export const TOMBSTONE_GAP = 10;

/** How far the stone's area stops short of the panel's far edge. */
export const TOMBSTONE_MARGIN = 10;

/** The insets, from the panel's top and bottom edges, of the stone's area. */
export interface Band {
  top: number;
  bottom: number;
}

/**
 * tombstoneBand returns the insets of the area the stone is centred in.
 * With nothing to keep clear the area is the whole panel (both zero). With
 * an identity in the panel's top half it starts below the identity; in the
 * bottom half it ends above it. The side with the room is the side that the
 * identity's middle is not on, so a panel is never split in the middle of
 * the avatar. The insets are never negative, and never leave less than a
 * zero-height line inside the panel.
 */
export function tombstoneBand(panel: Span, blocked: readonly Span[]): Band {
  const spans = blocked.filter(
    (s) => Number.isFinite(s.top) && Number.isFinite(s.bottom) && s.bottom > s.top,
  );
  const height = panel.bottom - panel.top;
  if (spans.length === 0 || !(height > 0)) return { top: 0, bottom: 0 };
  const top = Math.min(...spans.map((s) => s.top));
  const bottom = Math.max(...spans.map((s) => s.bottom));
  const nearTop = (top + bottom) / 2 < (panel.top + panel.bottom) / 2;
  if (nearTop) {
    const inset = Math.min(height, Math.max(0, Math.round(bottom + TOMBSTONE_GAP - panel.top)));
    return { top: inset, bottom: Math.min(TOMBSTONE_MARGIN, height - inset) };
  }
  const inset = Math.min(height, Math.max(0, Math.round(panel.bottom - (top - TOMBSTONE_GAP))));
  return { top: Math.min(TOMBSTONE_MARGIN, height - inset), bottom: inset };
}

/** The height the full stone needs: cross, epitaph, name, commander and its padding. */
export const TOMBSTONE_FULL_HEIGHT = 124;

/**
 * tombstoneFits is whether the full stone has room in the band: a short
 * panel (a top-row seat on a small screen) leaves less under the avatar
 * than the stone is tall, and takes the snug stone, as a collapsed seat's
 * card does, rather than have it cover the avatar's badges.
 */
export function tombstoneFits(panel: Span, band: Band): boolean {
  const height = panel.bottom - panel.top;
  // A panel that is not laid out (no height yet) says nothing about room.
  if (!(height > 0)) return true;
  return height - band.top - band.bottom >= TOMBSTONE_FULL_HEIGHT;
}
