import { describe, expect, it } from "vitest";
import {
  PhyrexianLifePerSymbol,
  clampPhyrexianLife,
  maxPhyrexianLife,
  phyrexianLifeCost,
  phyrexianSymbolsForAbility,
  phyrexianSymbolsForCast,
  shouldAskPhyrexianLife,
} from "./phyrexianLife";
import { applyCastChoices } from "./targeting";
import type { ActivatedAbilityView, CardView } from "./protocol";

// phyrexianLife.test.ts — #916. The board had no way to pay a
// Phyrexian mana symbol with life, so Gitaxian Probe cost {U} from
// hand however much the engine was willing to accept. These pin the
// arithmetic the stepper is built on: what its ceiling is, what the
// claim costs, when it should not open at all, and what goes on the
// wire.

function card(over: Partial<CardView> = {}): CardView {
  return {
    instance_id: "card-1",
    name: "Gitaxian Probe",
    owner: "me",
    controller: "me",
    ...over,
  } as CardView;
}

describe("phyrexianSymbolsForCast", () => {
  it("reads the printed cost's count off the server's field", () => {
    expect(phyrexianSymbolsForCast(card({ phyrexian_symbols: 1 }), undefined)).toBe(1);
    expect(phyrexianSymbolsForCast(card({ phyrexian_symbols: 2 }), undefined)).toBe(2);
  });

  it("is zero for a card with no Phyrexian symbol", () => {
    expect(phyrexianSymbolsForCast(card(), undefined)).toBe(0);
  });

  it("takes the CHOSEN alternative cost's count instead — the offer replaces the mana cost", () => {
    const c = card({
      phyrexian_symbols: 1,
      alternative_costs: [
        { key: "free", label: "Cast without paying its mana cost" },
        { key: "compleated", label: "Pay {W/P}{W/P}", phyrexian_symbols: 2 },
      ],
    });
    // Force of Will's free cast prints no symbol however many the
    // printed cost has.
    expect(phyrexianSymbolsForCast(c, "free")).toBe(0);
    expect(phyrexianSymbolsForCast(c, "compleated")).toBe(2);
    // No offer claimed is the printed cost.
    expect(phyrexianSymbolsForCast(c, undefined)).toBe(1);
  });

  it("is zero for an offer key the card does not carry", () => {
    expect(phyrexianSymbolsForCast(card({ phyrexian_symbols: 2 }), "nonesuch")).toBe(0);
  });
});

describe("phyrexianSymbolsForAbility", () => {
  it("reads the ability's own count", () => {
    const pod = { index: 0, phyrexian_symbols: 1 } as ActivatedAbilityView;
    const solphim = { index: 1, phyrexian_symbols: 2 } as ActivatedAbilityView;
    const plain = { index: 2 } as ActivatedAbilityView;
    expect(phyrexianSymbolsForAbility(pod)).toBe(1);
    expect(phyrexianSymbolsForAbility(solphim)).toBe(2);
    expect(phyrexianSymbolsForAbility(plain)).toBe(0);
  });
});

describe("maxPhyrexianLife", () => {
  it("is the symbol count when life is plentiful", () => {
    expect(maxPhyrexianLife(2, 40)).toBe(2);
  });

  it("is capped by CR 119.4 — a payment no larger than the life total", () => {
    // Two symbols cost 4; at 3 life only one is affordable.
    expect(maxPhyrexianLife(2, 3)).toBe(1);
    expect(maxPhyrexianLife(2, 4)).toBe(2);
  });

  it("allows paying down to exactly 0, which the engine allows", () => {
    expect(maxPhyrexianLife(1, PhyrexianLifePerSymbol)).toBe(1);
  });

  it("is zero when even one symbol is unaffordable", () => {
    expect(maxPhyrexianLife(1, PhyrexianLifePerSymbol - 1)).toBe(0);
    expect(maxPhyrexianLife(1, 0)).toBe(0);
  });

  it("is zero when the cost prints no Phyrexian symbol at all", () => {
    expect(maxPhyrexianLife(0, 40)).toBe(0);
  });

  it("reads a missing life total as nothing to spend, not as no limit", () => {
    expect(maxPhyrexianLife(2, undefined)).toBe(0);
  });
});

describe("clampPhyrexianLife", () => {
  it("holds the value inside [0, max]", () => {
    expect(clampPhyrexianLife(-1, 2)).toBe(0);
    expect(clampPhyrexianLife(0, 2)).toBe(0);
    expect(clampPhyrexianLife(3, 2)).toBe(2);
    expect(clampPhyrexianLife(1, 2)).toBe(1);
  });

  it("floors a fraction and rejects a non-number", () => {
    expect(clampPhyrexianLife(1.9, 2)).toBe(1);
    expect(clampPhyrexianLife(Number.NaN, 2)).toBe(0);
  });

  it("collapses to 0 when the cap drops to 0 under an open prompt", () => {
    // A life loss while the modal is up: the claim has to shrink
    // with the ceiling rather than sit there unpayable.
    expect(clampPhyrexianLife(2, 0)).toBe(0);
  });
});

describe("phyrexianLifeCost", () => {
  it("is CR 107.4f's 2 life per symbol", () => {
    expect(phyrexianLifeCost(0)).toBe(0);
    expect(phyrexianLifeCost(1)).toBe(2);
    expect(phyrexianLifeCost(3)).toBe(6);
  });
});

describe("shouldAskPhyrexianLife", () => {
  it("does not open for a cost with no Phyrexian symbol", () => {
    expect(shouldAskPhyrexianLife(0, 40)).toBe(false);
  });

  it("does not open when CR 119.4 leaves 0 as the only answer", () => {
    expect(shouldAskPhyrexianLife(1, 1)).toBe(false);
  });

  it("opens when at least one symbol can be bought", () => {
    expect(shouldAskPhyrexianLife(1, 2)).toBe(true);
    expect(shouldAskPhyrexianLife(2, 40)).toBe(true);
  });
});

describe("applyCastChoices", () => {
  it("sends the claim as phyrexian_life", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { phyrexianLife: 2 });
    expect(params.phyrexian_life).toBe(2);
  });

  it("omits a zero claim — the server default, and what pre-#916 clients send", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { phyrexianLife: 0 });
    expect(params).not.toHaveProperty("phyrexian_life");
  });

  it("omits the field entirely when nothing was announced", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { xValue: 3 });
    expect(params).not.toHaveProperty("phyrexian_life");
    expect(params.x_value).toBe(3);
  });
});
