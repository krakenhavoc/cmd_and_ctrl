// ADR 0198 — the logic behind the outside_collection picker.

import { describe, it, expect } from "vitest";

import {
  canTakeCollection,
  collectionRows,
  takeCollectionLabel,
  toggleCollectionPick,
} from "./outsideCollection";
import type { CardView } from "./protocol";

const suggestion = (name: string, type_line: string, mana_cost: string): CardView =>
  ({ instance_id: "", name, type_line, mana_cost }) as CardView;

describe("outside_collection picker", () => {
  it("shows the suggestions until the query is long enough to search", () => {
    const rows = collectionRows(
      [suggestion("Wrath of God", "Sorcery", "{2}{W}{W}")],
      [{ id: "x", name: "Cultivate", type_line: "Sorcery", mana_cost: "{2}{G}", set: "m21" }],
      "c",
    );
    expect(rows).toEqual([{ name: "Wrath of God", typeLine: "Sorcery", manaCost: "{2}{W}{W}" }]);
  });

  it("shows the search results once the player searches", () => {
    const rows = collectionRows(
      [suggestion("Wrath of God", "Sorcery", "{2}{W}{W}")],
      [{ id: "x", name: "Cultivate", type_line: "Sorcery", mana_cost: "{2}{G}", set: "m21" }],
      "cu",
    );
    expect(rows).toEqual([{ name: "Cultivate", typeLine: "Sorcery", manaCost: "{2}{G}" }]);
  });

  it("replaces the pick when one card is wanted", () => {
    expect(toggleCollectionPick(["Sol Ring"], "Ponder", 1)).toEqual(["Ponder"]);
    expect(toggleCollectionPick(["Sol Ring"], "Sol Ring", 1)).toEqual([]);
  });

  it("adds up to the ceiling and no further when more are wanted", () => {
    let picked = toggleCollectionPick([], "A", 2);
    picked = toggleCollectionPick(picked, "B", 2);
    expect(picked).toEqual(["A", "B"]);
    expect(toggleCollectionPick(picked, "C", 2)).toEqual(["A", "B"]);
    expect(toggleCollectionPick(picked, "A", 2)).toEqual(["B"]);
  });

  it("takes only a pick within the bounds, and never an empty one", () => {
    expect(canTakeCollection([], 0, 1)).toBe(false);
    expect(canTakeCollection(["A"], 0, 1)).toBe(true);
    expect(canTakeCollection(["A"], 2, 2)).toBe(false);
    expect(canTakeCollection(["A", "B"], 0, 2)).toBe(true);
  });

  it("names what the take button takes", () => {
    expect(takeCollectionLabel([])).toBe("Take…");
    expect(takeCollectionLabel(["Sol Ring", "Ponder"])).toBe("Take Sol Ring and Ponder");
  });
});
