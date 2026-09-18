import { describe, expect, it } from "vitest";
import { colorButtons, colorPromptAnswerable } from "./manaPick";

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

// #844, CR 903.4f: "any color in your commander's color identity" adds
// no mana for a player with no commander, or a colourless one. The
// server queues no pick at all in that case, so an empty option list
// should never reach the modal — and if one ever did, it must not open
// a picker nobody can answer.
describe("colorPromptAnswerable", () => {
  it("accepts a pick with colours on offer", () => {
    expect(colorPromptAnswerable({ kind: "mana_pick", color_options: ["G"] })).toBe(true);
  });

  it("refuses a mana pick with no colours", () => {
    expect(colorPromptAnswerable({ kind: "mana_pick", color_options: [] })).toBe(false);
    expect(colorPromptAnswerable({ kind: "mana_pick" })).toBe(false);
  });

  it("refuses a choose_color with no colours", () => {
    expect(colorPromptAnswerable({ kind: "choose_color", color_options: [] })).toBe(false);
  });

  it("leaves every other prompt kind alone", () => {
    expect(colorPromptAnswerable({ kind: "discard_from_hand" })).toBe(true);
  });
});
