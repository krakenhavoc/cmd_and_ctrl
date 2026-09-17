import { describe, expect, it } from "vitest";
import { colorButtons, manaAbilityEntries } from "./manaPick";

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

// Owner decision (2026-09-17): a manual tap of a source with exactly
// one identity colour on offer makes that colour in one click, and the
// menu keeps every other printed colour one row away.
describe("manaAbilityEntries", () => {
  it("leaves a single-colour ability with one bare row", () => {
    expect(manaAbilityEntries({ index: 0, label: "Add {G}", produced: "{G}" })).toEqual([
      { key: "0", label: "Add {G}" },
    ]);
  });

  it("names the one-click colour and offers the rest, identity first", () => {
    const rows = manaAbilityEntries({
      index: 1,
      label: "Add one mana of any color",
      produced: "{W|U|B|R|G}",
      color_options: ["G", "W", "U", "B", "R"],
      one_click_color: "G",
    });
    expect(rows.map((r) => r.color)).toEqual([undefined, "W", "U", "B", "R"]);
    expect(rows[0].label).toBe("Add one mana of any color → {G}");
    expect(rows[1].label).toBe("Add one mana of any color → {W}");
    expect(rows.map((r) => r.key)).toEqual(["1", "1-W", "1-U", "1-B", "1-R"]);
  });

  it("says 'first' when the source has more than one colour slot", () => {
    const rows = manaAbilityEntries({
      index: 0,
      label: "Add {W}{W}, {W}{U}, or {U}{U}",
      produced: "{W|U}{W|U}",
      color_options: ["W", "U"],
      one_click_color: "W",
    });
    expect(rows[1].label).toBe("Add {W}{W}, {W}{U}, or {U}{U} → {U} first");
  });

  it("keeps one row for a source that still prompts", () => {
    // Two identity colours on offer: the server sends neither field,
    // and the pick modal lists every colour as before.
    const rows = manaAbilityEntries({
      index: 2,
      label: "Add one mana of any color",
      produced: "{W|U|B|R|G}",
    });
    expect(rows).toEqual([{ key: "2", label: "Add one mana of any color" }]);
  });
});
