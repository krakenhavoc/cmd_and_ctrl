import { describe, expect, it } from "vitest";

import {
  canPayCards,
  payCardsOptions,
  payCardsReady,
  payCardsVerb,
  togglePayCard,
} from "./payCards";
import type { CardView, PayCardsView, PendingChoiceView, PlayerView } from "./protocol";
import { payUnlessAnswer } from "./waterbend";

// payCards.test.ts — ADR 0108 §5: a pay-unless whose payment discards or
// sacrifices.

const surger: PayCardsView = { action: "sacrifice", count: 2, options: ["a", "b", "c"] };
const imp: PayCardsView = { action: "discard", count: 1, options: ["h1", "h2"] };

describe("payCards", () => {
  it("can pay only with enough options (CR 118.3)", () => {
    expect(canPayCards(surger)).toBe(true);
    expect(canPayCards({ ...surger, options: ["a"] })).toBe(false);
  });

  it("is ready only at exactly the count", () => {
    expect(payCardsReady(surger, ["a"])).toBe(false);
    expect(payCardsReady(surger, ["a", "b"])).toBe(true);
  });

  it("toggles without passing the count or naming a non-option", () => {
    let picks = togglePayCard(surger, [], "a");
    picks = togglePayCard(surger, picks, "b");
    expect(togglePayCard(surger, picks, "c")).toEqual(["a", "b"]);
    expect(togglePayCard(surger, picks, "a")).toEqual(["b"]);
    expect(togglePayCard(surger, [], "zz")).toEqual([]);
  });

  it("reads a discard's options from the chooser's hand, a sacrifice's from the battlefield", () => {
    const card = (id: string, name: string) => ({ instance_id: id, name }) as CardView;
    const me = { id: "p1", hand: { cards: [card("h2", "Two"), card("h1", "One")] } } as PlayerView;
    expect(payCardsOptions(imp, [], me).map((c) => c.name)).toEqual(["One", "Two"]);
    const bf = [card("b", "Island B"), card("a", "Island A")];
    expect(payCardsOptions(surger, bf, me).map((c) => c.name)).toEqual(["Island A", "Island B"]);
  });

  it("says what to do", () => {
    expect(payCardsVerb(imp)).toBe("Choose one card to discard.");
    expect(payCardsVerb(surger)).toBe("Choose 2 cards to sacrifice.");
  });

  it("sends card_ids beside apply on a Pay, never on a decline", () => {
    const choice = {
      id: "c1",
      kind: "pay_unless",
      chooser: "p1",
      count: 1,
      pay_cost: "Sacrifice two lands",
      pay_cards: surger,
    } as PendingChoiceView;
    expect(payUnlessAnswer(choice, true, [], ["a", "b"])).toEqual({
      choice_id: "c1",
      apply: true,
      card_ids: ["a", "b"],
    });
    expect(payUnlessAnswer(choice, false, [], ["a"])).toEqual({ choice_id: "c1", apply: false });
  });
});
