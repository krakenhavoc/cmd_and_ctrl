import { describe, it, expect } from "vitest";
import type { GameView } from "./protocol";
import {
  laneSpot,
  sameStackPosition,
  savedStackPosition,
  stackFractions,
  stackPixels,
} from "./stackPosition";

const board = { width: 1000, height: 600 };
const panel = { width: 300, height: 200 };

describe("savedStackPosition", () => {
  const view = { stack_positions: { me: { x: 0.5, y: 0.25 } } } as unknown as GameView;

  it("reads the viewer's own entry", () => {
    expect(savedStackPosition(view, "me")).toEqual({ x: 0.5, y: 0.25 });
  });

  it("is null for a spectator, a seat with none, and a view with none", () => {
    expect(savedStackPosition(view, null)).toBeNull();
    expect(savedStackPosition(view, "bob")).toBeNull();
    expect(savedStackPosition({} as GameView, "me")).toBeNull();
    expect(savedStackPosition(undefined, "me")).toBeNull();
  });

  it("ignores a value that is not a number", () => {
    const bad = { stack_positions: { me: { x: Number.NaN, y: 0.1 } } } as unknown as GameView;
    expect(savedStackPosition(bad, "me")).toBeNull();
  });
});

describe("stackPixels", () => {
  it("scales fractions by the board", () => {
    expect(stackPixels({ x: 0.5, y: 0.25 }, board, panel)).toEqual({ left: 500, top: 150 });
  });

  it("keeps the whole card on the board", () => {
    expect(stackPixels({ x: 1, y: 1 }, board, panel)).toEqual({ left: 700, top: 400 });
    expect(stackPixels({ x: 0, y: 0 }, board, panel)).toEqual({ left: 0, top: 0 });
  });

  it("clamps again after the board shrinks, without losing the saved spot", () => {
    const saved = { x: 0.7, y: 0.5 };
    expect(stackPixels(saved, board, panel)).toEqual({ left: 700, top: 300 });
    const small = { width: 800, height: 400 };
    expect(stackPixels(saved, small, panel)).toEqual({ left: 500, top: 200 });
    // The board grows back and the card returns to where it was put.
    expect(stackPixels(saved, board, panel)).toEqual({ left: 700, top: 300 });
  });

  it("puts a card larger than the board at the corner", () => {
    expect(stackPixels({ x: 0.9, y: 0.9 }, { width: 200, height: 100 }, panel)).toEqual({
      left: 0,
      top: 0,
    });
  });
});

describe("stackFractions", () => {
  it("turns a dropped corner into fractions of the board", () => {
    expect(stackFractions(250, 120, board, panel)).toEqual({ x: 0.25, y: 0.2 });
  });

  it("clamps a drop past an edge", () => {
    expect(stackFractions(-40, -5, board, panel)).toEqual({ x: 0, y: 0 });
    expect(stackFractions(950, 590, board, panel)).toEqual({ x: 0.7, y: 400 / 600 });
  });

  it("round-trips with stackPixels", () => {
    const f = stackFractions(420, 77, board, panel);
    expect(stackPixels(f, board, panel)).toEqual({ left: 420, top: 77 });
  });

  it("is the origin on a board with no size", () => {
    expect(stackFractions(10, 10, { width: 0, height: 0 }, panel)).toEqual({ x: 0, y: 0 });
  });
});

describe("laneSpot", () => {
  // #3380: the pile's default on a four-seat table, docked on the seam.
  const anchor = { left: 350, top: 180 };

  it("a saved spot wins over the default", () => {
    expect(laneSpot({ x: 0.1, y: 0.1 }, board, panel, anchor)).toEqual({
      left: 100,
      top: 60,
      placed: true,
    });
    // Clamped on the board like any saved spot.
    expect(laneSpot({ x: 0.9, y: 0.9 }, board, panel, anchor)).toEqual({
      left: 700,
      top: 400,
      placed: true,
    });
  });

  it("is the default with nothing saved", () => {
    expect(laneSpot(null, board, panel, anchor)).toEqual({ ...anchor, placed: false });
    expect(laneSpot(null, board, panel, { left: null, top: null })).toEqual({
      left: null,
      top: null,
      placed: false,
    });
  });

  it("waits for a measured board and panel before drawing a saved spot", () => {
    const saved = { x: 0.1, y: 0.1 };
    expect(laneSpot(saved, { width: 0, height: 0 }, panel, anchor).placed).toBe(false);
    expect(laneSpot(saved, board, { width: 0, height: 0 }, anchor).placed).toBe(false);
  });
});

describe("sameStackPosition", () => {
  it("compares both fractions, and null with null", () => {
    expect(sameStackPosition({ x: 1, y: 2 }, { x: 1, y: 2 })).toBe(true);
    expect(sameStackPosition({ x: 1, y: 2 }, { x: 1, y: 3 })).toBe(false);
    expect(sameStackPosition(null, null)).toBe(true);
    expect(sameStackPosition({ x: 1, y: 2 }, null)).toBe(false);
  });
});
