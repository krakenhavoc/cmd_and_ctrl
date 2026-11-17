import { describe, expect, it } from "vitest";

import type { DraftCard } from "./draft";
import { byManaValue, columnLabel, columnOf, MV_COLUMNS } from "./draftColumns";

const card = (id: string, name: string, mana_value?: number): DraftCard => ({
  id,
  name,
  mana_value,
});

describe("columnOf", () => {
  it("is the mana value, 0 for none, and the last column for 6 and more", () => {
    expect(columnOf({ mana_value: 3 })).toBe(3);
    expect(columnOf({})).toBe(0);
    expect(columnOf({ mana_value: 6 })).toBe(6);
    expect(columnOf({ mana_value: 11 })).toBe(MV_COLUMNS - 1);
    expect(columnOf({ mana_value: -1 })).toBe(0);
    expect(columnOf({ mana_value: Number.NaN })).toBe(0);
  });
});

describe("columnLabel", () => {
  it("is the number, and 6+ for the last", () => {
    expect(columnLabel(0)).toBe("0");
    expect(columnLabel(5)).toBe("5");
    expect(columnLabel(6)).toBe("6+");
  });
});

describe("byManaValue", () => {
  it("always has seven columns, each sorted by name", () => {
    const cols = byManaValue([
      card("a", "Zephyr", 2),
      card("b", "Aether", 2),
      card("c", "Bog", 8),
      card("d", "Land"),
    ]);
    expect(cols).toHaveLength(MV_COLUMNS);
    expect(cols[0].map((c) => c.id)).toEqual(["d"]);
    expect(cols[2].map((c) => c.id)).toEqual(["b", "a"]);
    expect(cols[6].map((c) => c.id)).toEqual(["c"]);
    expect(cols[1]).toEqual([]);
  });

  it("is empty columns for no cards", () => {
    expect(byManaValue([]).every((c) => c.length === 0)).toBe(true);
  });
});
