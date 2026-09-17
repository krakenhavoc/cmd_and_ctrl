import { describe, expect, it } from "vitest";
import { colorButtons } from "./manaPick";

// Owner decision (2026-09-17): "any color" mana offers all five colours
// with the commander's identity listed first. The picker renders what
// the server sends, in the server's order.

describe("colorButtons", () => {
  it("renders all five colours of a mono-green Birds pick, green first", () => {
    const buttons = colorButtons(["G", "W", "U", "B", "R"]);
    expect(buttons.map((b) => b.color)).toEqual(["G", "W", "U", "B", "R"]);
    expect(buttons.map((b) => b.label)).toEqual(["Green", "White", "Blue", "Black", "Red"]);
    expect(buttons.every((b) => b.amount === 1)).toBe(true);
  });

  it("renders a narrowed Command Tower pick as sent", () => {
    expect(colorButtons(["G"]).map((b) => b.color)).toEqual(["G"]);
  });

  it("carries per-colour amounts for a one-colour-N-mana pick", () => {
    const buttons = colorButtons(["G", "U"], { G: 4 });
    expect(buttons.map((b) => [b.color, b.amount])).toEqual([
      ["G", 4],
      ["U", 1],
    ]);
  });

  it("renders nothing without options", () => {
    expect(colorButtons(undefined)).toEqual([]);
  });
});
