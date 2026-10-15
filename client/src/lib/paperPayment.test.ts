import { describe, expect, it } from "vitest";

import {
  FREE_PAYMENT,
  declaredPayment,
  paperPaymentSymbols,
  paperPaymentValid,
  prefillPaperPayment,
  withPaperPayment,
} from "./paperPayment";
import type { PaperPaymentPrompt } from "./protocol";

const ANY = ["W", "U", "B", "R", "G", "C"];

// Painful Truths {2}{B} with nothing in the pool: the server names the
// {B} first, then the generic.
const truths: PaperPaymentPrompt = {
  slots: [
    { symbol: "{B}", options: ["B"] },
    { symbol: "{1}", options: ANY },
    { symbol: "{1}", options: ANY },
  ],
};

describe("paperPayment (#3855)", () => {
  it("prefills the cost minus what the pool paid: the coloured slot's colour, {C} for generic", () => {
    expect(prefillPaperPayment(truths)).toEqual([
      { color: "B", artifact: false },
      { color: "C", artifact: false },
      { color: "C", artifact: false },
    ]);
  });

  it("prefills a hybrid slot with its first half", () => {
    const p: PaperPaymentPrompt = { slots: [{ symbol: "{W/U}", options: ["W", "U"] }] };
    expect(prefillPaperPayment(p)).toEqual([{ color: "W", artifact: false }]);
  });

  it("is valid only with one of each slot's options per slot", () => {
    const ok = prefillPaperPayment(truths);
    expect(paperPaymentValid(truths, ok)).toBe(true);
    expect(paperPaymentValid(truths, [{ color: "R", artifact: false }, ...ok.slice(1)])).toBe(
      false,
    );
    expect(paperPaymentValid(truths, ok.slice(1))).toBe(false);
  });

  it("sends the declared mana, with the artifact flag only when the prompt asked", () => {
    const choices = [
      { color: "B", artifact: true },
      { color: "W", artifact: false },
      { color: "C", artifact: true },
    ];
    expect(declaredPayment(truths, choices)).toEqual({
      mana: [{ color: "B" }, { color: "W" }, { color: "C" }],
    });
    expect(declaredPayment({ ...truths, artifact_sources: true }, choices)).toEqual({
      mana: [{ color: "B", artifact: true }, { color: "W" }, { color: "C", artifact: true }],
    });
    expect(paperPaymentSymbols(choices)).toBe("{B}{W}{C}");
  });

  it("re-sends the refused action's own payload with the answer", () => {
    const prev = { instance_id: "abc", targets: [{ kind: "player", id: "p2" }], force_cast: true };
    expect(withPaperPayment(prev, FREE_PAYMENT)).toEqual({
      ...prev,
      paper_payment: { free: true },
    });
    expect(prev).not.toHaveProperty("paper_payment");
  });
});
