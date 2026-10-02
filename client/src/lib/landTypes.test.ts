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

describe("landTypeEffectLine", () => {
  it("joins three types the way a card prints them", () => {
    expect(landTypeEffectLine({ types: ["Mountain", "Forest", "Plains"] })).toBe(
      "Mountain, Forest and Plains",
    );
  });
});
