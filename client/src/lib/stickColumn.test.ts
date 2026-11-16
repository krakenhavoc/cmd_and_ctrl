import { describe, expect, it } from "vitest";

import { STICK_GAP, stickTop } from "./stickColumn";

// #5848 §8: a column shorter than the window sticks under the header and
// the fuse; a taller one sticks by its end.

describe("stickTop", () => {
  it("is the offset for a column that fits the window", () => {
    expect(stickTop(85, 400, 900)).toBe(85);
    // It just fits: the offset and the gap leave room.
    expect(stickTop(85, 900 - STICK_GAP - 85, 900)).toBe(85);
  });

  it("is negative for a taller column, so its end shows at the foot of the window", () => {
    // 1200 tall in a 900 window: sticks 300 + the gap above the top edge.
    expect(stickTop(85, 1200, 900)).toBe(900 - 1200 - STICK_GAP);
    expect(stickTop(85, 1200, 900)).toBeLessThan(0);
  });

  it("follows the window as it is resized", () => {
    expect(stickTop(85, 700, 600)).toBe(600 - 700 - STICK_GAP);
    expect(stickTop(85, 700, 1000)).toBe(85);
  });
});
