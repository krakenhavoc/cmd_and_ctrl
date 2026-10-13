import { describe, expect, it } from "vitest";

import type { CardView } from "./protocol";
import { applyCastChoices, assistOf, assistPlayerLabel } from "./targeting";
import { payAmountHint, payAmountResource } from "./payEnergy";

// assist.test.ts — ADR 0150 §5 (CR 702.132a), the client half of the
// caster's side: the offer the cast modal's Assist row reads, how a
// bot seat is named in it, and the wire (assist_player rides only when
// somebody was asked).

function card(over: Partial<CardView> = {}): CardView {
  return { instance_id: "fan", name: "Fan Favorite", owner: "p0", controller: "p0", ...over };
}

describe("assist — cast_spell params", () => {
  it("writes assist_player only when a player was asked", () => {
    const alone: Record<string, unknown> = {};
    applyCastChoices(alone, {});
    expect(alone).not.toHaveProperty("assist_player");

    const empty: Record<string, unknown> = {};
    applyCastChoices(empty, { assistPlayer: "" });
    expect(empty).not.toHaveProperty("assist_player");

    const asked: Record<string, unknown> = {};
    applyCastChoices(asked, { assistPlayer: "p1" });
    expect(asked).toEqual({ assist_player: "p1" });
  });
});

describe("assist — the offer", () => {
  const offer = { max: 3, players: ["p1", "p2"] };

  it("is the card's offer, and nothing for a forced cast or a card without one", () => {
    expect(assistOf(card({ assist: offer }))).toEqual(offer);
    expect(assistOf(card({ assist: offer }), true)).toBeUndefined();
    expect(assistOf(card())).toBeUndefined();
    expect(assistOf(card({ assist: { max: 0, players: ["p1"] } }))).toBeUndefined();
  });

  it("marks a bot seat as always declining", () => {
    expect(assistPlayerLabel({ name: "Bob" })).toBe("Bob");
    expect(assistPlayerLabel({ name: "Botty", is_bot: true })).toBe("Botty (always declines)");
    expect(assistPlayerLabel(undefined)).toBe("A player");
  });
});

describe("assist — the helper's stepper", () => {
  it("reads as generic mana paid from the pool, then untapped sources", () => {
    const pa = { min: 0, max: 2, unit: "assist" as const, resource: "mana" as const };
    expect(payAmountResource(pa)).toBe("mana");
    const hint = payAmountHint(pa, "Fan Favorite");
    expect(hint).toContain("Pay up to {2} of its generic mana, or nothing.");
    expect(hint).toContain("mana pool first");
  });
});
