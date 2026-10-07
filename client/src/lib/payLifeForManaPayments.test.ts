import { describe, expect, it, vi } from "vitest";
import { choiceRequest } from "./choiceDock";
import type { PendingChoiceView } from "./protocol";
import { payUnlessAnswer } from "./waterbend";

// payLifeForManaPayments.test.ts — ADR 0131 §2 (#2531), PR 2: a mana
// pay_unless (ward {B}, Rhystic Study's tax) offers "Pay with N life" for
// the symbols the chooser could pay 2 life each for, and the answer carries
// `phyrexian_life`.

function ward(extra: Partial<PendingChoiceView> = {}): PendingChoiceView {
  return {
    id: "c1",
    kind: "pay_unless",
    chooser: "me",
    pay_cost: "{B}",
    phyrexian_symbols: 1,
    phyrexian_granted: 1,
    ...extra,
  } as PendingChoiceView;
}

function handlers() {
  return { onAnswer: vi.fn(), onCoin: vi.fn(), onOption: vi.fn(), onLoop: vi.fn() };
}

describe("pay_unless in the dock — paying with life", () => {
  it("offers a pay-with-life answer between Pay and Don't pay", () => {
    const h = handlers();
    const r = choiceRequest(ward(), { sourceName: "Ward", life: 20 }, h);
    expect(r.primary?.label).toBe("Pay {B}");
    expect(r.secondary?.map((a) => a.label)).toEqual(["Pay with 2 life", "Don't pay"]);
    r.secondary![0].onPress();
    expect(h.onAnswer).toHaveBeenCalledWith(true, 1);
  });

  it("offers one answer per count, bounded by the symbols and by CR 119.4", () => {
    const two = ward({ pay_cost: "{B}{B}", phyrexian_symbols: 2, phyrexian_granted: 2 });
    expect(
      choiceRequest(two, { sourceName: "Ward", life: 20 }, handlers()).secondary?.map(
        (a) => a.label,
      ),
    ).toEqual(["Pay with 2 life", "Pay with 4 life", "Don't pay"]);
    // 3 life buys one symbol, not two.
    expect(
      choiceRequest(two, { sourceName: "Ward", life: 3 }, handlers()).secondary?.map(
        (a) => a.label,
      ),
    ).toEqual(["Pay with 2 life", "Don't pay"]);
  });

  it("offers none when life cannot buy a symbol, or none is payable with life", () => {
    for (const [c, life] of [
      [ward(), 1],
      [ward(), undefined],
      [ward({ phyrexian_symbols: 0, phyrexian_granted: 0 }), 20],
      [ward({ phyrexian_symbols: undefined, phyrexian_granted: undefined }), 20],
    ] as const) {
      const r = choiceRequest(c, { sourceName: "Ward", life }, handlers());
      expect(r.secondary?.map((a) => a.label)).toEqual(["Don't pay"]);
    }
  });

  it("names what the life pays for in the hint", () => {
    const granted = choiceRequest(ward(), { sourceName: "Ward", life: 20 }, handlers());
    expect(granted.hint).toContain("The {B} can instead be paid with 2 life each.");
    const printed = choiceRequest(
      ward({ pay_cost: "{B/P}", phyrexian_granted: 0 }),
      { sourceName: "Ward", life: 20 },
      handlers(),
    );
    expect(printed.hint).toContain("The Phyrexian symbol can instead be paid with 2 life each.");
  });
});

describe("payUnlessAnswer — phyrexian_life", () => {
  it("rides a pay, never a decline", () => {
    expect(payUnlessAnswer(ward(), true, [], [], 1)).toEqual({
      choice_id: "c1",
      apply: true,
      phyrexian_life: 1,
    });
    expect(payUnlessAnswer(ward(), false, [], [], 1)).toEqual({ choice_id: "c1", apply: false });
    expect(payUnlessAnswer(ward(), true)).toEqual({ choice_id: "c1", apply: true });
  });
});
