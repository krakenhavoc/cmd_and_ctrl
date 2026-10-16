// ADR 0119 §1 — the pile's placement and size, as pure functions.

import { describe, expect, it } from "vitest";

import {
  CARD_RATIO,
  PILE_CARD_MAX_W,
  PILE_CARD_MIN_W,
  PILE_CHROME_H,
  PILE_INSET,
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
  pileLeft,
  pileTop,
  tableSeams,
  type SeatSlot,
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

  it("ignores a fixture left of a pile docked in the middle (#3380)", () => {
    const coach = { left: 115, top: 600, width: 300, height: 200 };
    // The coach card ends 315px into the board; a pile from 810 to 1110 clears it.
    expect(fixtureTopFrom(board, [coach], 1110, 810)).toBe(null);
    expect(fixtureTopFrom(board, [coach], 1110, 315)).toBe(null);
    expect(fixtureTopFrom(board, [coach], 1110, 300)).toBe(550);
  });

  it("sums the strip's blocks rather than reading its last bottom", () => {
    expect(columnBottom(10, [], 6)).toBeNull();
    expect(columnBottom(10, [0, 0], 6)).toBeNull();
    expect(columnBottom(10, [40], 6)).toBe(50);
    expect(columnBottom(10, [40, 0, 30], 6)).toBe(86);
  });
});

// #3380, the ADR 0176 amendment (option a): with nothing saved, the pile
// docks on the seam between the top row's seats instead of over a corner
// seat. The slots below are Board.svelte's grids at a 1920×1000 board:
// 6px padding, 8px gaps, the quadrant's 0.7fr / 1.3fr rows, and the row
// layout's even split (#4066).
describe("the pile's default spot", () => {
  const W = 1920;
  const H = 1000;
  const PAD = 6;
  const GAP = 8;
  const innerW = W - 2 * PAD;
  const innerH = H - 2 * PAD - GAP;

  type Row = { cols: number; spans?: number[]; self?: number };
  /** The slots of a grid whose rows are `rows`, `shares` of the height each. */
  function grid(rows: Row[], shares: number[]): SeatSlot[] {
    const out: SeatSlot[] = [];
    let top = PAD;
    const total = shares.reduce((a, b) => a + b, 0);
    rows.forEach((row, ri) => {
      const h = (innerH * shares[ri]) / total;
      const colW = (innerW - GAP * (row.cols - 1)) / row.cols;
      let col = 0;
      (row.spans ?? Array<number>(row.cols).fill(1)).forEach((span, i) => {
        out.push({
          left: PAD + col * (colW + GAP),
          top,
          width: colW * span + GAP * (span - 1),
          height: h,
          self: row.self === i,
        });
        col += span;
      });
      top += h + GAP;
    });
    return out;
  }
  const QUADRANT = [0.7, 1.3];
  const EVEN = [1, 1];
  const PILE_W = 300;

  const spot = (slots: SeatSlot[]) => {
    const seams = tableSeams(W, slots);
    return {
      ...seams,
      left: pileLeft({ boardW: W, pileW: PILE_W, seamX: seams.seamX, dock: seams.dock }),
    };
  };

  it("keeps the left edge on a two-seat table, centred on the row seam", () => {
    // "across" over "self", one column: no seam between seats in a row.
    const s = spot(grid([{ cols: 1 }, { cols: 1, self: 0 }], EVEN));
    expect(s.seamX).toBeNull();
    expect(s.left).toBeNull();
    expect(s.seamY).toBeCloseTo(PAD + innerH / 2 + GAP / 2);
  });

  it("docks in the middle on a three-seat table", () => {
    // "next across" over "self self".
    const s = spot(grid([{ cols: 2 }, { cols: 2, spans: [2], self: 0 }], QUADRANT));
    expect(s.seamX).toBe(W / 2);
    expect(s.left).toBe(W / 2 - PILE_W / 2);
    expect(s.seamY).toBeCloseTo(PAD + innerH * 0.35 + GAP / 2);
  });

  it("docks where the four seats meet on a four-seat quadrant table", () => {
    // "across across_next" over "next self".
    const s = spot(grid([{ cols: 2 }, { cols: 2, self: 1 }], QUADRANT));
    expect(s.seamX).toBe(W / 2);
    expect(s.left).toBe(810);
    expect(s.seamY).toBeCloseTo(PAD + innerH * 0.35 + GAP / 2);
  });

  it("takes the left of the two seams on a four-seat row table", () => {
    // "next across across_next" over the full-width self, split evenly
    // since #4066.
    const s = spot(grid([{ cols: 3 }, { cols: 3, spans: [3], self: 0 }], EVEN));
    const colW = (innerW - 2 * GAP) / 3;
    expect(s.seamX).toBeCloseTo(PAD + colW + GAP / 2);
    expect(s.left).toBe(Math.round(PAD + colW + GAP / 2 - PILE_W / 2));
  });

  it("docks in the middle of a spectator's 2×2 grid, which has no row seam", () => {
    const slots = grid([{ cols: 2 }, { cols: 2 }], EVEN);
    const s = spot(slots);
    expect(s.seamY).toBeNull();
    expect(s.seamX).toBe(W / 2);
    // Three seats leave the fourth cell empty; the top row still has two.
    expect(spot(slots.slice(0, 3)).seamX).toBe(W / 2);
  });

  it("stays 12px inside a board too narrow to centre it", () => {
    expect(pileLeft({ boardW: 400, pileW: 300, seamX: 200 })).toBe(50);
    expect(pileLeft({ boardW: 400, pileW: 300, seamX: 60 })).toBe(PILE_INSET);
    expect(pileLeft({ boardW: 400, pileW: 300, seamX: 380 })).toBe(400 - 300 - PILE_INSET);
    expect(pileLeft({ boardW: 300, pileW: 400, seamX: 150 })).toBe(PILE_INSET);
  });

  it("is the CSS fallback until the board and the pile are measured", () => {
    expect(pileLeft({ boardW: 0, pileW: 300, seamX: 960 })).toBeNull();
    expect(pileLeft({ boardW: W, pileW: 0, seamX: 960 })).toBeNull();
    // jsdom's unlaid-out slots have no size and no seam.
    const flat = [0, 1, 2].map(() => ({ left: 0, top: 0, width: 0, height: 0, self: false }));
    expect(tableSeams(W, flat).seamX).toBeNull();
  });

  // #4065, the reflow: the active opponent's board takes the row less
  // two 132px avatar cards, and the pile stays over the full board.
  describe("with the reflow's avatar cards in the row", () => {
    const AV = 132;
    const rowH = innerH / 2;
    const top = (left: number, width: number, avatar = false): SeatSlot => ({
      left,
      top: PAD,
      width,
      height: rowH,
      self: false,
      avatar,
    });
    const self: SeatSlot = {
      left: PAD,
      top: PAD + rowH + GAP,
      width: innerW,
      height: rowH,
      self: true,
    };
    const focusW = innerW - 2 * AV - 2 * GAP;

    it("has no dock without an avatar card", () => {
      const s = spot(grid([{ cols: 3 }, { cols: 3, spans: [3], self: 0 }], EVEN));
      expect(s.dock).toBeNull();
    });

    it("docks at the focus board's left edge when the avatar cards are to its left", () => {
      const focusLeft = PAD + 2 * (AV + GAP);
      const s = spot([
        top(PAD, AV, true),
        top(PAD + AV + GAP, AV, true),
        top(focusLeft, focusW),
        self,
      ]);
      expect(s.dock).toEqual({ left: focusLeft, right: focusLeft + focusW });
      expect(s.left).toBe(Math.round(focusLeft + PILE_INSET));
    });

    it("stays left of the avatar cards when they are to the focus board's right", () => {
      const s = spot([
        top(PAD, focusW),
        top(PAD + focusW + GAP, AV, true),
        top(PAD + focusW + GAP + AV + GAP, AV, true),
        self,
      ]);
      expect(s.left! + PILE_W).toBeLessThanOrEqual(PAD + focusW - PILE_INSET + 0.5);
      expect(s.left).toBeGreaterThanOrEqual(PAD + PILE_INSET);
    });

    it("never covers an avatar card when the focus board is between them", () => {
      const focusLeft = PAD + AV + GAP;
      const s = spot([
        top(PAD, AV, true),
        top(focusLeft, focusW),
        top(focusLeft + focusW + GAP, AV, true),
        self,
      ]);
      expect(s.left).toBeGreaterThanOrEqual(focusLeft + PILE_INSET);
      expect(s.left! + PILE_W).toBeLessThanOrEqual(focusLeft + focusW - PILE_INSET + 0.5);
    });
  });
});
