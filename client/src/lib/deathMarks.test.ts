import { describe, expect, it } from "vitest";
import { deathMarkBadge } from "./deathMarks";

describe("deathMarkBadge", () => {
  it("is null for a permanent with no marks", () => {
    expect(deathMarkBadge({})).toBeNull();
    expect(deathMarkBadge({ exiled_if_it_dies: [], cant_be_regenerated: false })).toBeNull();
  });

  it("names the sources of an exile mark once each", () => {
    expect(deathMarkBadge({ exiled_if_it_dies: ["Lava Coil", "Lava Coil", "Magma Spray"] })).toEqual({
      text: "EXILED IF IT DIES",
      title: "Exiled instead if it dies this turn — Lava Coil, Magma Spray",
    });
  });

  it("shows both marks together", () => {
    expect(deathMarkBadge({ exiled_if_it_dies: ["Disintegrate"], cant_be_regenerated: true })).toEqual({
      text: "NO REGEN · EXILED IF IT DIES",
      title: "Can't be regenerated this turn\nExiled instead if it dies this turn — Disintegrate",
    });
  });

  it("shows can't be regenerated alone", () => {
    expect(deathMarkBadge({ cant_be_regenerated: true })).toEqual({
      text: "NO REGEN",
      title: "Can't be regenerated this turn",
    });
  });
});
