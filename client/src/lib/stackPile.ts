// stackPile — where the pile sits and how big it is (ADR 0119 §1).
//
// The pile is the default stack style: the stack as a pile of large,
// readable cards. On a two-seat table it sits on the LEFT, so the hover
// zoom (pinned top-right) is never covered by it; with two or more seats
// in the top row it docks in the middle of the table, on the seam
// between them (#3380, the ADR 0176 amendment). StackLaneHost measures the board
// and hands the numbers here; everything that is a decision — the
// card's width, how many lower items peek out, where the pile's top
// edge goes, and when the board is too small for it and the compact
// card is drawn instead — is a pure function pinned by
// stackPile.test.ts. The two DOM readers at the bottom only collect
// rects for them.

import { L } from "./labels";

/** A Scryfall `normal` image is 488×680. */
export const CARD_RATIO = 680 / 488;

/** The pile's left edge sits this far in from the board's, like the attention strip. */
export const PILE_INSET = 12;
/** The gap kept to the board's edges, the strip and a bottom-left fixture. */
export const PILE_MARGIN = 8;

/** The top card's width: clamp(200px, 30% of the board's height, 280px)… */
export const PILE_CARD_MIN_W = 200;
export const PILE_CARD_MAX_W = 280;
export const PILE_CARD_HEIGHT_SHARE = 0.3;
/** …and never more than 24% of the board's width. */
export const PILE_CARD_WIDTH_SHARE = 0.24;
/** The top card while the viewer chooses a target (§1, "shrinks"). */
export const PILE_SHRUNK_W = 160;

/** Each lower item peeks out this far above the one in front: a `normal` image's name bar at 280px. */
export const PILE_PEEK_OFFSET = 34;
/** At most this many lower items peek out; deeper ones are behind "+N more". */
export const PILE_PEEK_MAX = 4;

/**
 * Everything in the pile that is not a card image: the host's header
 * and summary line, the caption under the top card (caster, title,
 * target line, chips, Counter) and the panel's padding. An estimate:
 * it only decides how far the card shrinks on a short board, so being
 * a few px out moves the card by a few px.
 */
export const PILE_CHROME_H = 168;
/** One pending-trigger row under the pile, and the group's heading. */
export const PILE_PENDING_ROW_H = 24;
export const PILE_PENDING_HEAD_H = 22;

/** How many lower items peek out above the top card, and how many are hidden. */
export interface PileDepth {
  peeking: number;
  more: number;
}

export function pileDepth(stackCount: number): PileDepth {
  const lower = Math.max(0, stackCount - 1);
  const peeking = Math.min(lower, PILE_PEEK_MAX);
  return { peeking, more: lower - peeking };
}

/** The height the pending-trigger group adds under the pile. */
export function pendingHeight(pendingCount: number): number {
  return pendingCount > 0 ? PILE_PENDING_HEAD_H + pendingCount * PILE_PENDING_ROW_H : 0;
}

/** The vertical band the pile may occupy, in board pixels. */
export interface PileBounds {
  minTop: number;
  maxBottom: number;
  /** maxBottom − minTop. */
  room: number;
}

export interface PileBoundsInput {
  boardH: number;
  /**
   * The bottom of the attention strip's content (the bot feed, a
   * reveal, a toast), in board pixels, or null when it has none.
   */
  stripBottom: number | null;
  /**
   * The top of a bottom-left fixture over the pile's column (the
   * tutorial's coach card, the dev dock), in board pixels, or null.
   */
  fixtureTop: number | null;
}

export function pileBounds({ boardH, stripBottom, fixtureTop }: PileBoundsInput): PileBounds {
  const minTop = stripBottom !== null ? stripBottom + PILE_MARGIN : PILE_MARGIN;
  let maxBottom = boardH - PILE_MARGIN;
  if (fixtureTop !== null) maxBottom = Math.min(maxBottom, fixtureTop - PILE_MARGIN);
  return { minTop, maxBottom, room: Math.max(0, maxBottom - minTop) };
}

export interface PileCardWidthInput {
  boardW: number;
  boardH: number;
  /** PileBounds.room. */
  room: number;
  /** Lower items peeking out (pileDepth().peeking). */
  peeking: number;
  /** Pending triggers listed under the pile. */
  pendingCount: number;
}

/**
 * The top card's width. The ADR's clamp, capped by the board's width,
 * and on a short board shrunk until the whole pile fits between the
 * strip and the bottom — but not below the targeting size, under which
 * the text is not worth drawing; the panel scrolls instead.
 */
export function pileCardWidth(i: PileCardWidthInput): number {
  const want = Math.min(
    Math.max(PILE_CARD_MIN_W, PILE_CARD_HEIGHT_SHARE * i.boardH),
    PILE_CARD_MAX_W,
  );
  const budget =
    i.room - PILE_CHROME_H - i.peeking * PILE_PEEK_OFFSET - pendingHeight(i.pendingCount);
  const byRoom = Math.max(PILE_SHRUNK_W, budget / CARD_RATIO);
  const byWidth = PILE_CARD_WIDTH_SHARE * i.boardW;
  return Math.floor(Math.min(want, byRoom, byWidth));
}

export interface PileTopInput extends PileBoundsInput {
  /** The seam between the opponents' row and the viewer's, or null for none. */
  seamY: number | null;
  /** The pile's measured height. */
  pileH: number;
}

/**
 * The pile's top edge: centred on the seam, then kept inside the band.
 * The strip wins when the pile is taller than the band, so the pile
 * never slides under a prompt; its panel scrolls instead.
 */
export function pileTop(i: PileTopInput): number {
  const { minTop, maxBottom } = pileBounds(i);
  const anchor = i.seamY ?? i.boardH / 2;
  const centred = anchor - i.pileH / 2;
  return Math.max(minTop, Math.min(centred, maxBottom - i.pileH));
}

export interface PileLeftInput {
  boardW: number;
  /** The pile's measured width. */
  pileW: number;
  /** tableSeams().seamX: the seam between two top-row seats, or null. */
  seamX: number | null;
}

/**
 * #3380, ADR 0176 amendment (option a, the fixed gutter): the pile's left
 * edge, centred on the seam between the top row's seats so it docks in
 * the middle of the table instead of over a corner seat, and kept 12px
 * inside the board. Null when the top row has one seat (a two-seat
 * table): the pile keeps ADR 0119's spot, 12px from the left edge.
 */
export function pileLeft({ boardW, pileW, seamX }: PileLeftInput): number | null {
  if (seamX === null || boardW <= 0 || pileW <= 0) return null;
  const max = Math.max(PILE_INSET, boardW - pileW - PILE_INSET);
  return Math.round(Math.min(Math.max(seamX - pileW / 2, PILE_INSET), max));
}

/** The grid's gap between seats (Board.svelte's `.board { gap }`). */
export const SEAT_GAP = 8;

/** One seat's slot on the board, in board pixels. */
export interface SeatSlot {
  left: number;
  top: number;
  width: number;
  height: number;
  /** The viewer's own slot (`data-pos="self"`). */
  self: boolean;
}

export interface TableSeams {
  /** The seam between the opponents' row and the viewer's, or null. */
  seamY: number | null;
  /**
   * The seam between two seats side by side in the top row, the one
   * nearest the board's centre (the left one on a tie), or null when the
   * top row has a single seat.
   */
  seamX: number | null;
}

/**
 * Where the table's seams are, from its seat slots. Every seated layout
 * puts the viewer in the bottom row and every opponent above, so the row
 * seam is half a gap above the viewer's slot; a spectator's uniform grid
 * has no viewer and no row seam. The column seam is read off the top row
 * (the seats whose top is the highest), which is the opponents' row when
 * seated and the first row of a spectator's grid.
 */
export function tableSeams(boardW: number, slots: readonly SeatSlot[]): TableSeams {
  const self = slots.find((s) => s.self);
  const others = slots.filter((s) => !s.self);
  const seamY = self && others.length > 0 ? self.top - SEAT_GAP / 2 : null;

  // An unlaid-out slot (no width) has no seam beside it.
  const drawn = others.filter((s) => s.width > 0 && s.height > 0);
  let seamX: number | null = null;
  if (drawn.length > 1) {
    const rowTop = Math.min(...drawn.map((s) => s.top));
    const row = drawn.filter((s) => Math.abs(s.top - rowTop) < 1).sort((a, b) => a.left - b.left);
    const centre = boardW / 2;
    for (let i = 1; i < row.length; i++) {
      const x = (row[i - 1].left + row[i - 1].width + row[i].left) / 2;
      // Half a pixel of slack, so a symmetric row's rounding cannot pick
      // the right-hand seam over the left.
      if (seamX === null || Math.abs(x - centre) < Math.abs(seamX - centre) - 0.5) seamX = x;
    }
  }
  return { seamY, seamX };
}

export interface PileFallbackInput {
  /** The viewport is at ADR 0111 §8's phone break (max-width: 599px). */
  phone: boolean;
  boardH: number;
  stripBottom: number | null;
}

/**
 * True when the board draws the compact card instead of the pile, the
 * stored setting untouched: on a phone (owner answer 3), and on a board
 * too short for a 200px top card between the strip and the bottom. An
 * unmeasured board (height 0, the first frame) is not short.
 */
export function pileFallsBack({ phone, boardH, stripBottom }: PileFallbackInput): boolean {
  if (phone) return true;
  if (boardH <= 0) return false;
  const { room } = pileBounds({ boardH, stripBottom, fixtureTop: null });
  return room < PILE_CARD_MIN_W * CARD_RATIO;
}

// ---------------------------------------------------------------- //
// Geometry from rects

interface Rect {
  left: number;
  top: number;
  width: number;
  height: number;
}

/**
 * The top of the highest fixture that covers the pile's column, in
 * board pixels, or null. A fixture counts when it overlaps the board
 * vertically and overlaps the pile's column, from `pileLeft` (the
 * board's left edge unless given) to `pileRight`.
 */
export function fixtureTopFrom(
  board: Rect,
  fixtures: readonly Rect[],
  pileRight: number,
  pileLeft = 0,
): number | null {
  let top: number | null = null;
  for (const r of fixtures) {
    if (r.width <= 0 || r.height <= 0) continue;
    if (r.left - board.left > pileRight || r.left + r.width - board.left <= pileLeft) continue;
    if (r.top >= board.top + board.height || r.top + r.height <= board.top) continue;
    const t = r.top - board.top;
    top = top === null ? t : Math.min(top, t);
  }
  return top;
}

/**
 * The bottom of a column of blocks laid out from `top` with `gap`
 * between them, or null when none has height. The strip is such a
 * column; this sums its blocks rather than reading its last block's
 * bottom, so the docked stack card (left out by the caller) cannot push
 * the answer down while the board draws it.
 */
export function columnBottom(top: number, heights: readonly number[], gap: number): number | null {
  const hs = heights.filter((h) => h > 0);
  if (hs.length === 0) return null;
  return top + hs.reduce((a, b) => a + b, 0) + gap * (hs.length - 1);
}

// ---------------------------------------------------------------- //
// DOM readers

/** Marks an element that sits over the board's bottom-left corner. */
export const BOTTOM_LEFT_FIXTURE_ATTR = "data-bottom-left-fixture";
/** Marks the docked stack card, which the strip measure leaves out. */
export const STACK_OVERLAY_ATTR = "data-stack-overlay";

/** The attention strip, a direct child of the board. */
export function attentionStrip(board: HTMLElement): HTMLElement | null {
  return board.querySelector<HTMLElement>(`:scope > [role="region"][aria-label="${L.attention}"]`);
}

/** The strip's content bottom in board pixels, the docked stack card left out. */
export function stripContentBottom(board: HTMLElement): number | null {
  const strip = attentionStrip(board);
  if (!strip) return null;
  const heights: number[] = [];
  for (const el of Array.from(strip.children)) {
    if (el.hasAttribute(STACK_OVERLAY_ATTR)) continue;
    heights.push(el.getBoundingClientRect().height);
  }
  const style = typeof getComputedStyle === "function" ? getComputedStyle(strip) : null;
  const gap = parseFloat(style?.rowGap ?? "") || 6;
  const top = strip.getBoundingClientRect().top - board.getBoundingClientRect().top;
  return columnBottom(top, heights, gap);
}

/** fixtureTopFrom over every marked fixture in the document. */
export function bottomLeftFixtureTop(
  board: HTMLElement,
  pileRight: number,
  pileLeft = 0,
): number | null {
  const rects = Array.from(
    board.ownerDocument.querySelectorAll<HTMLElement>(`[${BOTTOM_LEFT_FIXTURE_ATTR}]`),
  ).map((el) => el.getBoundingClientRect());
  return fixtureTopFrom(board.getBoundingClientRect(), rects, pileRight, pileLeft);
}

/** The seat slots directly under the board, in board pixels. */
export function seatSlots(board: HTMLElement): SeatSlot[] {
  const b = board.getBoundingClientRect();
  return Array.from(board.querySelectorAll<HTMLElement>(":scope > .slot")).map((el) => {
    const r = el.getBoundingClientRect();
    return {
      left: r.left - b.left,
      top: r.top - b.top,
      width: r.width,
      height: r.height,
      self: el.dataset.pos === "self",
    };
  });
}
