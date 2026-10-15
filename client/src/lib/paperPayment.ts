// paperPayment.ts — "What did you pay?" (#3855, ADR 0068 amendment
// 2026-10-15).
//
// When a payment is waived (strict mana off with too little in the pool,
// or Cast anyway) and something on the table reads the mana spent — a
// converge or sunburst spell, Vexing Bauble, Coin of Mastery — the server
// refuses the cast with `paper_payment_required`, naming what the pool
// left unpaid one slot at a time. Nothing has been paid. The player says
// what they paid outside the pool, or that an effect let them cast it
// without paying its mana cost, and the client re-sends the SAME action
// with `paper_payment`. A cast nothing reads never gets here and stays
// one click.
//
// The server is the judge of a declaration (exactly the shortfall, each
// coloured slot paid by one of its options). The picker only ever offers
// a slot its own options, so what it builds is always legal; the
// validity check here is for the confirm button, not for the rules.

import type { PaperPaymentPrompt } from "./protocol";

export interface PaperMana {
  color: string;
  artifact?: boolean;
}

export interface PaperPayment {
  free?: boolean;
  mana?: PaperMana[];
}

// The picker's state for one slot: the symbol chosen and, when the
// prompt asks about sources, whether an artifact made it.
export interface PaperSlotChoice {
  color: string;
  artifact: boolean;
}

// prefillPaperPayment is the cost minus what the pool paid: each slot's
// first option — the colour a coloured slot names, the first half of a
// hybrid — and {C} for a generic slot, whose options end in C.
export function prefillPaperPayment(prompt: PaperPaymentPrompt): PaperSlotChoice[] {
  return prompt.slots.map((s) => ({
    color: s.symbol === "{1}" ? "C" : (s.options[0] ?? "C"),
    artifact: false,
  }));
}

// paperPaymentValid: one choice per slot, each one of its slot's options.
export function paperPaymentValid(prompt: PaperPaymentPrompt, choices: PaperSlotChoice[]): boolean {
  return (
    choices.length === prompt.slots.length &&
    choices.every((c, i) => prompt.slots[i]!.options.includes(c.color))
  );
}

// declaredPayment is the `paper_payment` a set of choices sends. The
// artifact flag rides only when the prompt asked about sources.
export function declaredPayment(
  prompt: PaperPaymentPrompt,
  choices: PaperSlotChoice[],
): PaperPayment {
  return {
    mana: choices.map((c) =>
      prompt.artifact_sources && c.artifact
        ? { color: c.color, artifact: true }
        : { color: c.color },
    ),
  };
}

// FREE_PAYMENT is "an effect on paper let me cast it without paying its
// mana cost".
export const FREE_PAYMENT: PaperPayment = { free: true };

// paperPaymentSymbols is a declaration as the log will print it,
// "{C}{G}", for the confirm button.
export function paperPaymentSymbols(choices: PaperSlotChoice[]): string {
  return choices.map((c) => `{${c.color}}`).join("");
}

// withPaperPayment is the refused action's own payload with the answer
// added, so its targets, modes, X and every other choice survive the
// re-send — the same replay the insufficient-mana override does.
export function withPaperPayment(
  prev: Record<string, unknown>,
  payment: PaperPayment,
): Record<string, unknown> {
  return { ...prev, paper_payment: payment };
}
