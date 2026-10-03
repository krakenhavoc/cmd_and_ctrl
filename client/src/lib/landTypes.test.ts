import { describe, expect, it } from "vitest";
import { landTypeBadge, landTypeEffectLine } from "./landTypes";

describe("landTypeBadge", () => {
  it("is null for a land no resolved effect has changed", () => {
    expect(landTypeBadge({})).toBeNull();
    expect(landTypeBadge({ land_type_effects: [] })).toBeNull();
    expect(landTypeBadge({ land_type_effects: [{ types: [] }] })).toBeNull();
  });

  it("names the type, the duration and the source", () => {
    expect(
      landTypeBadge({
        land_type_effects: [
          { types: ["Island"], until: "until end of turn", source: "Tidal Warrior" },
        ],
      }),
    ).toEqual({ text: "ISLAND", title: "Island until end of turn — Tidal Warrior" });
  });

  it("marks a type added in addition to the land's own", () => {
    expect(
      landTypeBadge({
        land_type_effects: [
          {
            types: ["Forest"],
            in_addition: true,
            until: "until end of turn",
            source: "Navigator's Compass",
          },
        ],
      }),
    ).toEqual({
      text: "+FOREST",
      title: "Forest in addition to its other types until end of turn — Navigator's Compass",
    });
  });

  it("shows the newest effect and lists every one, oldest first", () => {
    const badge = landTypeBadge({
      land_type_effects: [
        { types: ["Swamp"], source: "Thelonite Monk" },
        { types: ["Plains", "Island"], until: "until Bob's next turn", source: "Orcish Farmer" },
      ],
    });
    expect(badge).toEqual({
      text: "PLAINS ISLAND",
      title: "Swamp — Thelonite Monk\nPlains and Island until Bob's next turn — Orcish Farmer",
    });
  });
});

describe("a land that lost all its land types (ADR 0109 §2)", () => {
  it("says what the effect took and gave, and for how long", () => {
    expect(
      landTypeBadge({
        land_type_effects: [
          {
            types: [],
            loses_all: true,
            loses_abilities: true,
            gains: ["{T}: Add {C}."],
            until: "for as long as it has a blight counter on it",
            source: "Ultima, Origin of Oblivion",
          },
        ],
      }),
    ).toEqual({
      text: "NO LAND TYPES",
      title:
        'No land types, no abilities, "{T}: Add {C}." for as long as it has a blight counter on it — Ultima, Origin of Oblivion',
    });
  });

  it("names only the land types when the effect takes nothing else", () => {
    expect(landTypeEffectLine({ types: [], loses_all: true, source: "Test" })).toBe(
      "No land types — Test",
    );
  });
});

describe("landTypeEffectLine", () => {
  it("joins three types the way a card prints them", () => {
    expect(landTypeEffectLine({ types: ["Mountain", "Forest", "Plains"] })).toBe(
      "Mountain, Forest and Plains",
    );
  });
});
