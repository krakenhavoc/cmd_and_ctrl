import { describe, expect, it } from "vitest";
import {
  freeCastOffer,
  handFreeCastTarget,
  isFreeCastPick,
  GRANTED_COST_KEY,
} from "./freeCastRequest";
import type { AlternativeCostView, CardView, PendingChoiceView } from "./protocol";

// #3270 (CR 608.2g): picking a card for "cast a spell from your hand
// without paying its mana cost" starts that card's cast at once. Board
// reads these helpers; the chain after them is the ordinary one.

const offer = (over: Partial<AlternativeCostView> = {}): AlternativeCostView => ({
  key: GRANTED_COST_KEY,
  mana_cost: "{0}",
  ...over,
});

const card = (id: string, offers?: AlternativeCostView[]): CardView =>
  ({ instance_id: id, alternative_costs: offers }) as unknown as CardView;

describe("isFreeCastPick", () => {
  it("is a choose_cards prompt that declares cast_free", () => {
    const pick = { kind: "choose_cards", choose_destination: "cast_free" } as PendingChoiceView;
    expect(isFreeCastPick(pick)).toBe(true);
  });

  it("is not a plain pick, a put onto the battlefield or another kind", () => {
    expect(isFreeCastPick({ kind: "choose_cards" } as PendingChoiceView)).toBe(false);
    expect(
      isFreeCastPick({
        kind: "choose_cards",
        choose_destination: "battlefield",
      } as PendingChoiceView),
    ).toBe(false);
    expect(
      isFreeCastPick({ kind: "may_cast", choose_destination: "cast_free" } as PendingChoiceView),
    ).toBe(false);
    expect(isFreeCastPick(undefined)).toBe(false);
  });
});

describe("freeCastOffer", () => {
  it("finds the granted free offer", () => {
    expect(freeCastOffer(card("a", [offer()]))?.key).toBe(GRANTED_COST_KEY);
  });

  it("stays open for a sorcery in combat: the server judges timing per offer", () => {
    expect(freeCastOffer(card("a", [offer({ timing_closed: false })]))).toBeDefined();
  });

  it("ignores a closed offer, another key and a card with none", () => {
    expect(freeCastOffer(card("a", [offer({ timing_closed: true })]))).toBeUndefined();
    expect(freeCastOffer(card("a", [offer({ key: "miracle" })]))).toBeUndefined();
    expect(freeCastOffer(card("a"))).toBeUndefined();
  });
});

describe("handFreeCastTarget", () => {
  it("starts the cast once the hand card shows the free offer", () => {
    const t = handFreeCastTarget([card("a", [offer()])], "a", false);
    expect(t?.card.instance_id).toBe("a");
    expect(t?.key).toBe(GRANTED_COST_KEY);
  });

  it("waits while the pick is still open", () => {
    expect(handFreeCastTarget([card("a")], "a", true)).toBeUndefined();
    expect(handFreeCastTarget([], "a", true)).toBeUndefined();
  });

  it("drops a pick that never got an offer (no grant) once the prompt is gone", () => {
    expect(handFreeCastTarget([card("a")], "a", false)).toBeNull();
  });

  it("drops a pick whose card left the hand", () => {
    expect(handFreeCastTarget([], "a", false)).toBeNull();
  });

  it("has nothing to do without a request", () => {
    expect(handFreeCastTarget([card("a", [offer()])], null, false)).toBeNull();
  });
});
