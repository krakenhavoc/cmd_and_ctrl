import { describe, expect, it } from "vitest";
import { stampManaEnforcement } from "./manaEnforcement";

// #1296: "Equip costs were not paid when equipping to Vivi Ornitier".
// The strictMana setting reached cast_spell and nothing else, so every
// activated ability took the server's paper path — and an empty pool
// made it free.
describe("stampManaEnforcement", () => {
  const equip = {
    source_card_id: "blade",
    ability_index: 0,
    targets: [{ kind: "card", id: "vivi" }],
  };

  it("charges a catalog activation when the player enforces mana", () => {
    expect(stampManaEnforcement("activate_ability", equip, true)).toEqual({
      ...equip,
      strict: true,
      auto_tap: true,
    });
  });

  it("leaves an activation on paper when the player does not", () => {
    expect(stampManaEnforcement("activate_ability", equip, false)).toEqual(equip);
  });

  it("leaves the free-form announce alone — it has no cost to charge", () => {
    const freeForm = { source_card_id: "x", label: "{T}: something" };
    expect(stampManaEnforcement("activate_ability", freeForm, true)).toEqual(freeForm);
  });

  it("stamps a cast with the setting either way, as it always has", () => {
    expect(stampManaEnforcement("cast_spell", { instance_id: "c" }, true)).toEqual({
      instance_id: "c",
      strict: true,
    });
    expect(stampManaEnforcement("cast_spell", { instance_id: "c" }, false)).toEqual({
      instance_id: "c",
      strict: false,
    });
  });

  it("respects a payload that already says strict (the override retries)", () => {
    const forced = { instance_id: "c", strict: true, force_cast: true };
    expect(stampManaEnforcement("cast_spell", forced, false)).toBe(forced);
    const paper = { ...equip, strict: false };
    expect(stampManaEnforcement("activate_ability", paper, true)).toBe(paper);
  });

  it("does not touch other actions", () => {
    const tap = { instance_id: "land" };
    expect(stampManaEnforcement("tap", tap, true)).toBe(tap);
  });
});
