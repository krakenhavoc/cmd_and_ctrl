import { describe, it, expect } from "vitest";
import { chargedManaCostLabel, chargedManaCostNote } from "./contextMenu.logic";

// chargedManaCost.test.ts — the client half of #1190. The server
// stamps charged_mana_cost alongside the printed mana_cost on both
// ability rows (ActivatedAbilityView and ManaAbilityView) — what the
// CR 601.2f cost-modifier pass will actually charge right now, equal
// to mana_cost when no discount reaches the ability. chargedManaCostNote
// is the one predicate that decides whether there is anything to tell
// the player about the difference; ManaAbilityMenu's chip and
// contextMenu.logic's row hint both call it so the two surfaces never
// disagree about when a discount is worth mentioning.

describe("chargedManaCostNote", () => {
  it("is silent when there is no mana component at all", () => {
    expect(chargedManaCostNote({})).toBe("");
  });

  it("is silent when the server did not price the ability (pre-#1190 or a pricing error)", () => {
    expect(chargedManaCostNote({ mana_cost: "{3}{R}" })).toBe("");
  });

  it("is silent when the charged cost equals the printed one — nearly every ability in the game", () => {
    expect(chargedManaCostNote({ mana_cost: "{3}{R}", charged_mana_cost: "{3}{R}" })).toBe("");
  });

  it("names the printed cost when a discount made the two differ", () => {
    expect(chargedManaCostNote({ mana_cost: "{3}{R}", charged_mana_cost: "{1}{R}" })).toBe(
      "printed cost {3}{R}",
    );
  });

  it("names the printed cost even when the discount empties it out completely", () => {
    expect(chargedManaCostNote({ mana_cost: "{3}", charged_mana_cost: "" })).toBe(
      "printed cost {3}",
    );
  });
});

describe("chargedManaCostLabel", () => {
  it("falls back to the printed cost when the server did not price the ability", () => {
    expect(chargedManaCostLabel({ mana_cost: "{3}{R}" })).toBe("{3}{R}");
  });

  it("shows the charged cost when it is present and non-empty", () => {
    expect(chargedManaCostLabel({ mana_cost: "{3}{R}", charged_mana_cost: "{1}{R}" })).toBe(
      "{1}{R}",
    );
  });

  it("shows the charged cost even when it equals the printed one", () => {
    expect(chargedManaCostLabel({ mana_cost: "{3}{R}", charged_mana_cost: "{3}{R}" })).toBe(
      "{3}{R}",
    );
  });

  // The property this function exists for: an empty charged_mana_cost
  // is a real, priced answer (the discount emptied the component out
  // completely) and must not be treated as falsy and fall back to the
  // stale printed cost the way `a.charged_mana_cost || a.mana_cost`
  // would.
  it('shows "free" rather than the stale printed cost when the charge is a real empty string', () => {
    expect(chargedManaCostLabel({ mana_cost: "{2}", charged_mana_cost: "" })).toBe("free");
  });
});
