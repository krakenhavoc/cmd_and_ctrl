// ADR 0119 §1 — the pile's placement and size, as pure functions.

import { describe, expect, it } from "vitest";

import {
  CARD_RATIO,
  PILE_CARD_MAX_W,
  PILE_CARD_MIN_W,
  PILE_CHROME_H,
  PILE_MARGIN,
  PILE_PEEK_OFFSET,
  PILE_SHRUNK_W,
  columnBottom,
  fixtureTopFrom,
  pendingHeight,
  pileBounds,
  pileCardWidth,
  pileDepth,
  pileFallsBack,
  pileTop,
} from "./stackPile";

describe("the pile's depth", () => {
  it("peeks up to four lower items and counts the rest as more", () => {
    expect(pileDepth(0)).toEqual({ peeking: 0, more: 0 });
    expect(pileDepth(1)).toEqual({ peeking: 0, more: 0 });
    expect(pileDepth(3)).toEqual({ peeking: 2, more: 0 });
    expect(pileDepth(5)).toEqual({ peeking: 4, more: 0 });
    // A fifth lower item is the first behind "+N more".
    expect(pileDepth(6)).toEqual({ peeking: 4, more: 1 });
    expect(pileDepth(11)).toEqual({ peeking: 4, more: 6 });
  });
});

describe("the pile's band", () => {
  it("keeps 8px from the board's edges with nothing in the way", () => {
    expect(pileBounds({ boardH: 900, stripBottom: null, fixtureTop: null })).toEqual({
      minTop: PILE_MARGIN,
      maxBottom: 900 - PILE_MARGIN,
      room: 900 - 2 * PILE_MARGIN,
    });
  });

  it("sits below the strip's content and above a bottom-left fixture", () => {
    const b = pileBounds({ boardH: 900, stripBottom: 120, fixtureTop: 700 });
    expect(b.minTop).toBe(128);
    expect(b.maxBottom).toBe(692);
    expect(b.room).toBe(564);
  });

  it("ignores a fixture below the board", () => {
    expect(pileBounds({ boardH: 900, stripBottom: null, fixtureTop: 950 }).maxBottom).toBe(892);
  });
});

describe("the pile's top edge", () => {
  const base = { boardH: 900, stripBottom: null, fixtureTop: null };

  it("centres on the seam", () => {
    expect(pileTop({ ...base, seamY: 400, pileH: 300 })).toBe(250);
  });

  it("centres on the board's middle with no seam", () => {
    expect(pileTop({ ...base, seamY: null, pileH: 300 })).toBe(300);
  });

  it("stays below the strip's content", () => {
    expect(pileTop({ ...base, stripBottom: 200, seamY: 300, pileH: 300 })).toBe(208);
  });

  it("stays above the coach card", () => {
    expect(pileTop({ ...base, fixtureTop: 600, seamY: 500, pileH: 300 })).toBe(292);
  });

  it("lets the strip win when the pile is taller than the band", () => {
    expect(pileTop({ ...base, stripBottom: 100, fixtureTop: 400, seamY: 300, pileH: 600 })).toBe(
      108,
    );
  });
});

describe("the top card's width", () => {
  const roomy = { boardW: 1600, room: 2000, peeking: 0, pendingCount: 0 };

  it("is 30% of the board's height between 200 and 280px", () => {
    expect(pileCardWidth({ ...roomy, boardH: 900 })).toBe(270);
    expect(pileCardWidth({ ...roomy, boardH: 500 })).toBe(PILE_CARD_MIN_W);
    expect(pileCardWidth({ ...roomy, boardH: 1400 })).toBe(PILE_CARD_MAX_W);
  });

  it("is never more than 24% of the board's width", () => {
    expect(pileCardWidth({ ...roomy, boardH: 900, boardW: 1000 })).toBe(240);
  });

  it("shrinks on a short band so the pile fits, down to the targeting size", () => {
    const peeking = 4;
    const room = 600;
    const fits = Math.floor((room - PILE_CHROME_H - peeking * PILE_PEEK_OFFSET) / CARD_RATIO);
    expect(pileCardWidth({ boardW: 1600, boardH: 900, room, peeking, pendingCount: 0 })).toBe(fits);
    expect(pileCardWidth({ boardW: 1600, boardH: 900, room: 300, peeking, pendingCount: 0 })).toBe(
      PILE_SHRUNK_W,
    );
  });

  it("makes room for the pending triggers listed under it", () => {
    const room = 600;
    const a = pileCardWidth({ boardW: 1600, boardH: 900, room, peeking: 0, pendingCount: 0 });
    const b = pileCardWidth({ boardW: 1600, boardH: 900, room, peeking: 0, pendingCount: 3 });
    expect(b).toBeLessThan(a);
    expect(pendingHeight(0)).toBe(0);
    expect(pendingHeight(3)).toBeGreaterThan(pendingHeight(1));
  });
});

describe("the compact fallback", () => {
  it("is taken on a phone, whatever the board", () => {
    expect(pileFallsBack({ phone: true, boardH: 900, stripBottom: null })).toBe(true);
  });

  it("is not taken on a desktop board with room", () => {
    expect(pileFallsBack({ phone: false, boardH: 900, stripBottom: 200 })).toBe(false);
  });

  it("is taken on a board too short for a 200px card between the strip and the bottom", () => {
    const card = PILE_CARD_MIN_W * CARD_RATIO; // ≈ 279px
    const fitsExactly = card + 2 * PILE_MARGIN;
    expect(pileFallsBack({ phone: false, boardH: fitsExactly, stripBottom: null })).toBe(false);
    expect(pileFallsBack({ phone: false, boardH: fitsExactly - 1, stripBottom: null })).toBe(true);
    expect(pileFallsBack({ phone: false, boardH: 500, stripBottom: 250 })).toBe(true);
  });

  it("is not taken on an unmeasured board", () => {
    expect(pileFallsBack({ phone: false, boardH: 0, stripBottom: null })).toBe(false);
  });
});

describe("geometry from rects", () => {
  const board = { left: 100, top: 50, width: 1200, height: 800 };

  it("finds the highest fixture over the pile's column", () => {
    const coach = { left: 115, top: 600, width: 300, height: 200 };
    const dev = { left: 0, top: 780, width: 200, height: 70 };
    expect(fixtureTopFrom(board, [coach, dev], 300)).toBe(550);
  });

  it("ignores a fixture right of the pile, off the board, or not drawn", () => {
    expect(fixtureTopFrom(board, [{ left: 700, top: 600, width: 100, height: 100 }], 300)).toBe(
      null,
    );
    expect(fixtureTopFrom(board, [{ left: 115, top: 900, width: 100, height: 100 }], 300)).toBe(
      null,
    );
    expect(fixtureTopFrom(board, [{ left: 115, top: 600, width: 0, height: 0 }], 300)).toBe(null);
  });

  it("sums the strip's blocks rather than reading its last bottom", () => {
    expect(columnBottom(10, [], 6)).toBeNull();
    expect(columnBottom(10, [0, 0], 6)).toBeNull();
    expect(columnBottom(10, [40], 6)).toBe(50);
    expect(columnBottom(10, [40, 0, 30], 6)).toBe(86);
  });
});
