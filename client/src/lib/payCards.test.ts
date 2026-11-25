import { describe, expect, it } from "vitest";

import {
  canPayCards,
  payCardsOptions,
  payCardsProgress,
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

  it("reads an exile's options from the chooser's own graveyard", () => {
    const card = (id: string, name: string) => ({ instance_id: id, name }) as CardView;
    const me = {
      id: "p1",
      hand: { cards: [card("h1", "In hand")] },
      graveyard: { cards: [card("g2", "Two"), card("g1", "One")] },
    } as PlayerView;
    const exile: PayCardsView = { action: "exile_graveyard", count: 1, options: ["g1", "g2"] };
    const bf = [card("g1", "On the battlefield")];
    expect(payCardsOptions(exile, bf, me).map((c) => c.name)).toEqual(["One", "Two"]);
    expect(payCardsVerb(exile)).toBe("Choose one card from your graveyard to exile.");
  });

  it("says what to do", () => {
    expect(payCardsVerb(imp)).toBe("Choose one card to discard.");
    expect(payCardsVerb(surger)).toBe("Choose 2 cards to sacrifice.");
  });

  it("reads a tap's options from the battlefield (#1997, Koskun Falls)", () => {
    const card = (id: string, name: string) => ({ instance_id: id, name }) as CardView;
    const me = { id: "p1", hand: { cards: [card("h1", "In hand")] } } as PlayerView;
    const tap: PayCardsView = { action: "tap", count: 1, options: ["b2", "b1"] };
    const bf = [card("b1", "Bear One"), card("b2", "Bear Two")];
    expect(payCardsOptions(tap, bf, me).map((c) => c.name)).toEqual(["Bear Two", "Bear One"]);
    expect(payCardsVerb(tap)).toBe("Choose one card to tap.");
  });

  it("reads a -1/-1 counter's options from the battlefield (#5746, Thelon's Chant)", () => {
    const card = (id: string, name: string) => ({ instance_id: id, name }) as CardView;
    const me = { id: "p1", hand: { cards: [card("h1", "In hand")] } } as PlayerView;
    const counter: PayCardsView = { action: "minus_one_counter", count: 1, options: ["b1", "b2"] };
    const bf = [card("b2", "Bear Two"), card("b1", "Bear One")];
    expect(payCardsOptions(counter, bf, me).map((c) => c.name)).toEqual(["Bear One", "Bear Two"]);
    expect(payCardsVerb(counter)).toBe("Choose one card to put a -1/-1 counter on.");
    expect(canPayCards({ ...counter, options: [] })).toBe(false);
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

describe("#6187 payments", () => {
  const dreadnought: PayCardsView = {
    action: "sacrifice_total_power",
    count: 0,
    options: ["a", "b", "c"],
    total_power: 12,
    powers: { a: 7, b: 4, c: 1 },
  };
  it("a total-power sacrifice takes any number and is ready at the total", () => {
    expect(canPayCards(dreadnought)).toBe(true);
    expect(payCardsReady(dreadnought, [])).toBe(false);
    expect(payCardsReady(dreadnought, ["a", "b"])).toBe(false);
    expect(payCardsReady(dreadnought, ["a", "b", "c"])).toBe(true);
    expect(togglePayCard(dreadnought, ["a", "b", "c"], "a")).toEqual(["b", "c"]);
    expect(payCardsProgress(dreadnought, ["a", "b"])).toBe("power 11 / 12");
    expect(payCardsVerb(dreadnought)).toContain("total power 12");
  });
  it("a total-power sacrifice short of the total cannot pay", () => {
    expect(canPayCards({ ...dreadnought, powers: { a: 7, b: 4, c: 0 } })).toBe(false);
  });
  it("a hand-to-library payment picks from the hand", () => {
    const specter: PayCardsView = { action: "hand_to_library_top", count: 1, options: ["h1"] };
    const chooser = {
      hand: { cards: [{ instance_id: "h1", name: "Island" }] },
    } as unknown as PlayerView;
    expect(payCardsOptions(specter, [], chooser).map((c) => c.name)).toEqual(["Island"]);
    expect(payCardsVerb(specter)).toBe("Choose one card to put on top of your library.");
  });
});
