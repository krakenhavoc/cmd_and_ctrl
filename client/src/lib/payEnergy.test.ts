// payEnergy.test.ts — ADR 0129 §3: the two prompts that pay energy while
// a spell or ability resolves, as the dock draws them. The pure half;
// the stepper itself lives in ChoicePromptModal.

import { describe, it, expect, vi } from "vitest";

import { choiceRequest, isInlineChoice } from "./choiceDock";
import { L } from "./labels";
import {
  energyShortBy,
  payAmountAnswerable,
  payAmountCanSubmit,
  payAmountClamp,
  payAmountStart,
} from "./payEnergy";
import type { PayAmountView, PendingChoiceView } from "./protocol";

const choice = (over: Partial<PendingChoiceView>): PendingChoiceView =>
  ({
    id: "c1",
    kind: "pay_unless",
    chooser: "me",
    from_player: "me",
    count: 1,
    ...over,
  }) as PendingChoiceView;

const handlers = () => ({
  onAnswer: vi.fn(),
  onCoin: vi.fn(),
  onOption: vi.fn(),
  onLoop: vi.fn(),
  onAmount: vi.fn(),
});

describe("an energy pay_unless", () => {
  const rhino = choice({
    pay_cost: "{E}{E}",
    pay_energy: 2,
    reason: "Thriving Rhino — pay {E}{E}?",
  });

  it("is inline, and Pay is live when the seat has the energy", () => {
    expect(isInlineChoice(rhino)).toBe(true);
    const h = handlers();
    const req = choiceRequest(rhino, { sourceName: "Thriving Rhino", energy: 3 }, h);
    expect(req.primary?.label).toBe("Pay {E}{E}");
    expect(req.primary?.disabled).toBeFalsy();
    expect(req.tag).toBe("pay energy");
    req.primary?.onPress();
    expect(h.onAnswer).toHaveBeenCalledWith(true);
  });

  it("greys Pay with the server's reason when the seat is short", () => {
    const req = choiceRequest(rhino, { sourceName: "Thriving Rhino", energy: 1 }, handlers());
    expect(req.primary?.disabled).toBe(true);
    expect(req.primary?.title).toBe("Not enough energy (have 1, need 2)");
    expect(req.hintWarn).toBe(true);
    expect(req.secondary?.[0]?.label).toBe("Don't pay");
  });

  it("is short only for an energy payment", () => {
    expect(energyShortBy(choice({ pay_cost: "{2}" }), 0)).toBe(0);
    expect(energyShortBy(rhino, 0)).toBe(2);
    expect(energyShortBy(choice({ pay_energy: 0 }), 0)).toBe(0);
  });
});

describe("a pay_amount", () => {
  const pa: PayAmountView = { min: 0, max: 5, goal: 3, unit: "damage" };
  const lightning = choice({ kind: "pay_amount", pay_amount: pa, reason: "Harnessed Lightning" });

  it("opens on the card's threshold, else the smallest payment", () => {
    expect(payAmountStart(pa)).toBe(3);
    expect(payAmountStart({ ...pa, goal: 0 })).toBe(1);
    expect(payAmountStart({ min: 1, max: 2, unit: "counters" })).toBe(1);
  });

  it("takes 0 or min..max, whole numbers only", () => {
    const one: PayAmountView = { min: 1, max: 4, unit: "counters" };
    expect([0, 1, 4].map((n) => payAmountAnswerable(one, n))).toEqual([true, true, true]);
    expect([-1, 5, 1.5].map((n) => payAmountAnswerable(one, n))).toEqual([false, false, false]);
    expect(payAmountClamp(one, 9)).toBe(4);
    expect(payAmountClamp(one, -3)).toBe(1);
  });

  it("is inline: Pay sends the stepper's amount, the decline sends 0", () => {
    expect(isInlineChoice(lightning)).toBe(true);
    const h = handlers();
    const req = choiceRequest(
      lightning,
      { sourceName: "Harnessed Lightning", payAmount: 3, payAmountAnswerable: true },
      h,
    );
    expect(req.primary?.label).toBe(L.payEnergy(3));
    req.primary?.onPress();
    expect(h.onAmount).toHaveBeenCalledWith(3);
    expect(req.secondary?.[0]?.label).toBe(L.payNothing);
    req.secondary?.[0]?.onPress();
    expect(h.onAmount).toHaveBeenLastCalledWith(0);
  });

  it("reads Don't pay for one or more, and greys Pay on an amount the server refuses", () => {
    const pia = choice({ kind: "pay_amount", pay_amount: { min: 1, max: 4, unit: "power" } });
    const req = choiceRequest(
      pia,
      { sourceName: "Pia Nalaar", payAmount: 9, payAmountAnswerable: false },
      handlers(),
    );
    expect(req.secondary?.[0]?.label).toBe("Don't pay");
    expect(req.primary?.disabled).toBe(true);
  });
});

// #1941, ADR 0129's amendment of 2026-10-09: the same prompt asks for
// life, and for a number that is chosen and not paid.
describe("a pay_amount in life", () => {
  const pa: PayAmountView = { min: 0, max: 30, goal: 3, unit: "cards", resource: "life" };
  const necro = choice({ kind: "pay_amount", pay_amount: pa, reason: "Necrodominance" });

  it("reads Pay N life with a decline, and names life in the hint", () => {
    const h = handlers();
    const req = choiceRequest(
      necro,
      { sourceName: "Necrodominance", payAmount: 3, payAmountAnswerable: true },
      h,
    );
    expect(req.primary?.label).toBe(L.payLife(3));
    expect(req.tag).toBe("pay life");
    expect(req.hint).toContain("Pay any amount of life, up to 30, or nothing.");
    req.primary?.onPress();
    expect(h.onAmount).toHaveBeenCalledWith(3);
    expect(req.secondary?.[0]?.label).toBe(L.payNothing);
  });
});

describe("a pay_amount that is only chosen", () => {
  const pa: PayAmountView = {
    min: 0,
    max: 1_000_000,
    goal: 4,
    unit: "damage",
    resource: "none",
    no_max: true,
    marks: [4, 40],
    self_damage: true,
  };
  const hellion = choice({ kind: "pay_amount", pay_amount: pa, reason: "Volcano Hellion" });

  it("starts at 0 when there is no goal, takes 0, and has no decline", () => {
    expect(payAmountStart({ ...pa, goal: 0 })).toBe(0);
    expect(payAmountAnswerable(pa, 0)).toBe(true);
    expect(payAmountCanSubmit(pa, 0)).toBe(true);
    expect(payAmountAnswerable({ ...pa, min: 1 }, 0)).toBe(false);
    const h = handlers();
    const req = choiceRequest(
      hellion,
      { sourceName: "Volcano Hellion", payAmount: 0, payAmountAnswerable: true },
      h,
    );
    expect(req.primary?.label).toBe(L.chooseNumber(0));
    expect(req.primary?.disabled).toBeFalsy();
    expect(req.secondary ?? []).toEqual([]);
    expect(req.tag).toBe("choose a number");
    expect(req.hint).toContain("Choose any number from 0 up.");
    expect(req.hint).toContain("you are dealt the same amount");
    req.primary?.onPress();
    expect(h.onAmount).toHaveBeenCalledWith(0);
  });

  it("takes a number past any goal", () => {
    expect(payAmountAnswerable(pa, 500)).toBe(true);
    expect(payAmountClamp(pa, 500)).toBe(500);
  });
});
