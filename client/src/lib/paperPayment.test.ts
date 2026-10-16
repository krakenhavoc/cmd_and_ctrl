import { describe, expect, it } from "vitest";

import {
  FREE_PAYMENT,
  declaredPayment,
  paperPaymentSymbols,
  paperPaymentValid,
  prefillPaperPayment,
  slotCount,
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

  // #3904: a run longer than the server lists slot by slot (an X-sized
  // shortfall) is one slot with a count, answered by one counted entry.
  it("answers a counted slot with one counted entry", () => {
    const drain: PaperPaymentPrompt = {
      slots: [
        { symbol: "{B}", options: ["B"], count: 999991 },
        { symbol: "{1}", options: ANY },
      ],
      artifact_sources: true,
    };
    const choices = prefillPaperPayment(drain);
    expect(choices).toEqual([
      { color: "B", artifact: false, count: 999991 },
      { color: "C", artifact: false },
    ]);
    expect(paperPaymentValid(drain, choices)).toBe(true);
    expect(paperPaymentValid(drain, [{ color: "B", artifact: false }, choices[1]!])).toBe(false);
    choices[0] = { ...choices[0]!, artifact: true };
    expect(declaredPayment(drain, choices)).toEqual({
      mana: [{ color: "B", artifact: true, count: 999991 }, { color: "C" }],
    });
    expect(paperPaymentSymbols(choices)).toBe("{B}×999991{C}");
    expect(slotCount({})).toBe(1);
    expect(slotCount({ count: 1 })).toBe(1);
  });

  // #3941: Soul Burn's "spend only black and/or red mana on X" — the X
  // slots arrive marked, and the mana declared for them says so, so the
  // record keeps which mana paid for X.
  it("marks the mana declared for a slot of the X part", () => {
    const soulBurn: PaperPaymentPrompt = {
      slots: [
        { symbol: "{B}", options: ["B"] },
        { symbol: "{B/R}", options: ["B", "R"], x: true },
        { symbol: "{B/R}", options: ["B", "R"], x: true },
        { symbol: "{1}", options: ANY },
      ],
    };
    const choices = prefillPaperPayment(soulBurn);
    choices[2] = { ...choices[2]!, color: "R" };
    expect(declaredPayment(soulBurn, choices)).toEqual({
      mana: [{ color: "B" }, { color: "B", x: true }, { color: "R", x: true }, { color: "C" }],
    });
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
