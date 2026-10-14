import { describe, expect, it } from "vitest";
import { overlaps, sheetCoversBoardAction, type Box } from "./sheetCover";

// The damage sheet at 1280x720 (#3095): x 695-1255, y 212-524.
const SHEET: Box = { left: 695, top: 212, right: 1255, bottom: 524 };
// A blocker under it, and one clear of it to the left.
const UNDER: Box = { left: 703, top: 265, right: 775, bottom: 365 };
const CLEAR: Box = { left: 120, top: 265, right: 192, bottom: 365 };

describe("overlaps", () => {
  it("is true for boxes sharing area and false for touching or empty ones", () => {
    expect(overlaps(SHEET, UNDER)).toBe(true);
    expect(overlaps(SHEET, CLEAR)).toBe(false);
    expect(overlaps(SHEET, { left: 600, top: 300, right: 695, bottom: 400 })).toBe(false);
    expect(overlaps(SHEET, { left: 800, top: 300, right: 800, bottom: 300 })).toBe(false);
  });
});

describe("sheetCoversBoardAction (#2959, #3095)", () => {
  it("is true when something the player can act on is under the sheet", () => {
    expect(sheetCoversBoardAction(SHEET, [CLEAR, UNDER])).toBe(true);
  });

  it("is false when everything actionable is clear of the sheet", () => {
    expect(sheetCoversBoardAction(SHEET, [CLEAR])).toBe(false);
  });

  it("is false for a sheet with nothing to click on the board, or no sheet", () => {
    expect(sheetCoversBoardAction(SHEET, [])).toBe(false);
    expect(sheetCoversBoardAction(null, [UNDER])).toBe(false);
  });
});
