import { describe, expect, it } from "vitest";
import { discardCostOf, discardXCount, hasXCost, xCeilingFor } from "./targeting";
import type { CardView } from "./protocol";

// #6145: "As an additional cost to cast this spell, discard X cards"
// (Channeled Force) — the X prompt opens for a card whose mana cost has no
// {X}, is capped at what the hand pays, and the discard picker that
// follows it asks for exactly X cards, never before X is known.

function force(max: number): CardView {
  return {
    instance_id: "spell",
    name: "Channeled Force",
    owner: "me",
    controller: "me",
    mana_cost: "{2}{U}{R}",
    additional_cost: {
      label: "Discard X cards",
      demands_x: true,
      discard_x: true,
      discard_x_max: max,
    },
  };
}

describe("discard X cards", () => {
  it("asks for X although the mana cost has none, capped at the hand", () => {
    const card = force(3);
    expect(hasXCost(card)).toBe(true);
    expect(xCeilingFor(card, undefined)).toBe(3);
    expect(xCeilingFor({ ...card, x_max: 2 }, undefined)).toBe(2);
  });

  it("opens no discard picker before X, and sizes the one after X to X", () => {
    const card = force(3);
    expect(discardCostOf(card, {})).toBe(0);
    expect(discardXCount(card, { xValue: 2 })).toBe(2);
    expect(discardXCount(card, { xValue: 0 })).toBe(0);
    expect(discardXCount(card, {})).toBe(0);
  });

  it("is zero for a fixed discard", () => {
    const card: CardView = { ...force(3), additional_cost: { discard_cards: 1 } };
    expect(discardXCount(card, { xValue: 2 })).toBe(0);
    expect(discardCostOf(card, {})).toBe(1);
  });
});
