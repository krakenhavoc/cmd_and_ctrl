// ADR 0119 §1 — the pile's placement and size, as pure functions.

import { describe, expect, it } from "vitest";

import {
  CARD_RATIO,
  PILE_CARD_MAX_W,
  PILE_CARD_MIN_W,
  PILE_CHROME_H,
  PILE_INSET,
  PILE_MARGIN,
  PILE_MIN_ROOM,
  PILE_LANE_EXTRA_W,
  PILE_SHRUNK_W,
  avatarClampsFrom,
  cascadeDim,
  cascadeInward,
  cascadeLayout,
  cascadeStep,
  columnBottom,
  fixtureTopFrom,
  pendingHeight,
  pileBounds,
  pileColumn,
  pileDepth,
  pileFallsBack,
  pileFit,
  pileHeight,
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
  const roomy = { boardW: 1600, room: 2000, lower: 0, pendingCount: 0 };

  it("is 30% of the board's height between 200 and 280px", () => {
    expect(pileFit({ ...roomy, boardH: 900 }).cardWidth).toBe(270);
    expect(pileFit({ ...roomy, boardH: 500 }).cardWidth).toBe(PILE_CARD_MIN_W);
    expect(pileFit({ ...roomy, boardH: 1400 }).cardWidth).toBe(PILE_CARD_MAX_W);
  });

  it("is never more than 24% of the board's width", () => {
    expect(pileFit({ ...roomy, boardH: 900, boardW: 1000 }).cardWidth).toBe(240);
  });

  it("shrinks only when one cascade step does not fit at full size, down to the targeting size", () => {
    const room = 500;
    const fits = Math.floor((room - PILE_CHROME_H) / (CARD_RATIO * 1.12));
    const f = pileFit({ boardW: 1600, boardH: 900, room, lower: 4, pendingCount: 0 });
    expect(f).toEqual({ cardWidth: fits, shown: 1, more: 3 });
    expect(pileFit({ boardW: 1600, boardH: 900, room: 300, lower: 4, pendingCount: 0 })).toEqual({
      cardWidth: PILE_SHRUNK_W,
      shown: 1,
      more: 3,
    });
    // A lone top card has no step to fit.
    const alone = pileFit({ boardW: 1600, boardH: 900, room, lower: 0, pendingCount: 0 });
    expect(alone.cardWidth).toBe(Math.floor((room - PILE_CHROME_H) / CARD_RATIO));
  });

  it("makes room for the pending triggers listed under it", () => {
    const room = 600;
    const a = pileFit({ boardW: 1600, boardH: 900, room, lower: 0, pendingCount: 0 });
    const b = pileFit({ boardW: 1600, boardH: 900, room, lower: 0, pendingCount: 3 });
    expect(b.cardWidth).toBeLessThan(a.cardWidth);
    expect(pendingHeight(0)).toBe(0);
    expect(pendingHeight(3)).toBeGreaterThan(pendingHeight(1));
  });
});

describe("the cascade (ADR 0119 amendment §A)", () => {
  it("steps 12% of the top card's height a step", () => {
    expect(cascadeStep(270)).toBe(Math.round(0.12 * 270 * CARD_RATIO)); // 45px
    expect(cascadeStep(PILE_CARD_MIN_W)).toBe(33);
  });

  it("dims 28% a step, capped at 60%", () => {
    expect([0, 1, 2, 3, 4].map(cascadeDim)).toEqual([0, 0.28, 0.56, 0.6, 0.6]);
  });

  it("puts each lower card up, inward to the right and 4% smaller, on the left half", () => {
    const c = cascadeLayout(250, 4, "right");
    const step = cascadeStep(250);
    expect(c.step).toBe(step);
    expect(c.rise).toBe(4 * step);
    expect(c.topLeft).toBe(0);
    // The deepest card overhangs the top card's inward edge by 2% a step.
    expect(c.width).toBe(250 + 20);
    expect(c.cards).toEqual([
      { depth: 1, left: 15, top: 3 * step, width: 240, dim: 0.28 },
      { depth: 2, left: 30, top: 2 * step, width: 230, dim: 0.56 },
      { depth: 3, left: 45, top: step, width: 220, dim: 0.6 },
      { depth: 4, left: 60, top: 0, width: 210, dim: 0.6 },
    ]);
    // Each card's strip shows exactly one step above the card in front.
    for (let i = 1; i < c.cards.length; i++) {
      expect(c.cards[i - 1].top - c.cards[i].top).toBe(step);
    }
    expect(c.rise - c.cards[0].top).toBe(step);
  });

  it("mirrors to the left on the right half: the outer (right) edges step in", () => {
    const c = cascadeLayout(250, 4, "left");
    expect(c.width).toBe(270);
    expect(c.topLeft).toBe(20);
    const topRight = c.topLeft + 250;
    expect(c.cards.map((k) => topRight - (k.left + k.width))).toEqual([15, 30, 45, 60]);
    expect(c.cards.map((k) => k.left)).toEqual([15, 10, 5, 0]);
  });

  it("is just the top card with nothing under it", () => {
    expect(cascadeLayout(250, 0, "right")).toEqual({
      width: 250,
      rise: 0,
      step: 0,
      topLeft: 0,
      cards: [],
    });
  });

  it("steps toward the board's centre", () => {
    expect(cascadeInward({ left: 12, width: 290, boardW: 1400 })).toBe("right");
    expect(cascadeInward({ left: 1000, width: 290, boardW: 1400 })).toBe("left");
    // Docked on the middle seam: the default.
    expect(cascadeInward({ left: 555, width: 290, boardW: 1400 })).toBe("right");
    expect(cascadeInward({ left: 0, width: 0, boardW: 0 })).toBe("right");
  });
});

describe("the cascade's depth on a board", () => {
  // The band a board of this height gives with nothing in the strip.
  const room = (boardH: number) => pileBounds({ boardH, stripBottom: null, fixtureTop: null }).room;
  const fit = (boardH: number, lower: number, pendingCount = 0) =>
    pileFit({ boardW: 1600, boardH, room: room(boardH), lower, pendingCount });

  it("shows all four lower cards on a 1440×900 table's board", () => {
    expect(fit(820, 6)).toEqual({ cardWidth: 246, shown: 4, more: 2 });
    expect(fit(820, 2)).toEqual({ cardWidth: 246, shown: 2, more: 0 });
  });

  it("sheds lower cards before it shrinks the top card, down to one", () => {
    // Boards this short want the 200px card; each shorter one drops a step.
    const want = PILE_CARD_MIN_W;
    const at = (n: number) => Math.ceil(pileHeight(want, n, 0)) + 2 * PILE_MARGIN;
    for (const n of [4, 3, 2, 1]) {
      expect(fit(at(n), 4)).toMatchObject({ cardWidth: want, shown: n, more: 4 - n });
      // A pixel short of n steps shows n − 1 (still at full size), or at one, shrinks.
      const short = fit(at(n) - 1, 4);
      if (n > 1) expect(short).toMatchObject({ cardWidth: want, shown: n - 1 });
      else expect(short.cardWidth).toBeLessThan(want);
    }
  });

  it("counts the pending triggers against the cascade, not just the card", () => {
    expect(fit(820, 4, 6).shown).toBeLessThan(fit(820, 4, 0).shown);
  });
});

describe("the compact fallback", () => {
  it("is taken on a phone, whatever the board", () => {
    expect(pileFallsBack({ phone: true, boardH: 900, stripBottom: null })).toBe(true);
  });

  it("is not taken on a desktop board with room", () => {
    expect(pileFallsBack({ phone: false, boardH: 900, stripBottom: 200 })).toBe(false);
  });

  it("is taken on a board too short for a 200px card and one cascade step", () => {
    const card = PILE_CARD_MIN_W * CARD_RATIO; // ≈ 279px
    expect(PILE_MIN_ROOM).toBe(card + cascadeStep(PILE_CARD_MIN_W)); // ≈ 312px
    const fitsExactly = PILE_MIN_ROOM + 2 * PILE_MARGIN;
    // A board that fits the card alone but not the step falls back.
    expect(pileFallsBack({ phone: false, boardH: card + 2 * PILE_MARGIN, stripBottom: null })).toBe(
      true,
    );
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

  // ADR 0204 §4: the default spot covers no avatar. A top-row avatar
  // over the pile's column pushes its top down; yours keeps its bottom up.
  it("clamps the pile between the avatars over its column", () => {
    const seamY = 400;
    const top = { left: 300, top: 60, width: 120, height: 110 }; // board y 10..120
    const mine = { left: 380, top: 700, width: 160, height: 110 }; // board y 650..760
    const clamps = avatarClampsFrom(board, [top, mine], seamY, 150, 450);
    expect(clamps).toEqual({ above: 120, below: 650 });
  });

  it("ignores avatars beside the pile's column and ones not drawn", () => {
    const right = { left: 700, top: 60, width: 120, height: 110 };
    const none = { left: 300, top: 700, width: 0, height: 0 };
    expect(avatarClampsFrom(board, [right, none], 400, 150, 450)).toEqual({
      above: null,
      below: null,
    });
  });

  it("splits a seamless board's avatars at its middle", () => {
    const high = { left: 300, top: 100, width: 100, height: 100 }; // centre 100
    const low = { left: 300, top: 600, width: 100, height: 100 }; // centre 600
    expect(avatarClampsFrom(board, [high, low], null, 150, 450)).toEqual({
      above: 150,
      below: 550,
    });
  });

  it("reads the clamps over the pile's widest column, not its drawn width", () => {
    // 1400×900: the card wants min(max(200, 270), 280, 336) = 270.
    const c = pileColumn(1400, 900, 700);
    expect(c.right - c.left).toBe(270 + PILE_LANE_EXTRA_W);
    expect((c.left + c.right) / 2).toBe(700);
    expect(pileColumn(1400, 900, null).left).toBe(PILE_INSET);
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

  // ADR 0077 as amended 2026-10-17 (#4102): the avatar cards stack in
  // one 176px column at the row's left, each half the row's height, and
  // the pile's default spot sits right of the column.
  describe("with the reflow's avatar cards stacked in one column", () => {
    const COL = 176;
    const rowH = innerH / 2;
    const cardH = (rowH - GAP) / 2;
    const card = (i: number): SeatSlot => ({
      left: PAD,
      top: PAD + i * (cardH + GAP),
      width: COL,
      height: cardH,
      self: false,
      avatar: true,
    });
    const focusLeft = PAD + COL + GAP;
    const focusW = innerW - COL - GAP;
    const focus: SeatSlot = { left: focusLeft, top: PAD, width: focusW, height: rowH, self: false };
    const self: SeatSlot = {
      left: PAD,
      top: PAD + rowH + GAP,
      width: innerW,
      height: rowH,
      self: true,
    };

    it("docks right of the column with two cards, never over either", () => {
      const s = spot([card(0), card(1), focus, self]);
      expect(s.dock).toEqual({ left: focusLeft, right: focusLeft + focusW });
      expect(s.left).toBe(Math.round(focusLeft + PILE_INSET));
      expect(s.left).toBeGreaterThan(PAD + COL);
    });

    it("docks right of the column with one card", () => {
      const one: SeatSlot = { ...card(0), height: rowH };
      const s = spot([one, focus, self]);
      expect(s.dock).toEqual({ left: focusLeft, right: focusLeft + focusW });
      expect(s.left).toBeGreaterThan(PAD + COL);
    });

    it("stays over the full boards when a prompt keeps a second board beside the focus", () => {
      const half = (focusW - GAP) / 3;
      const needed: SeatSlot = { ...focus, width: half };
      const wide: SeatSlot = { ...focus, left: focusLeft + half + GAP, width: focusW - half - GAP };
      const s = spot([{ ...card(0), height: rowH }, needed, wide, self]);
      expect(s.dock).toEqual({ left: focusLeft, right: focusLeft + focusW });
      expect(s.left).toBeGreaterThanOrEqual(focusLeft + PILE_INSET);
    });
  });
});
